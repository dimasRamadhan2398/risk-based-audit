package provisioner

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestTLSAndHealthTimeoutValidation(t *testing.T) {
	n := func(v int) *int { return &v }
	cases := []struct {
		name  string
		mut   func(*CreateRequest)
		field string // "" = valid
	}{
		{"allow_no_tls in local", func(r *CreateRequest) { r.AllowNoTLS = true }, "allow_no_tls"},
		{"allow_no_tls false in local", func(r *CreateRequest) { r.AllowNoTLS = false }, ""},
		{"allow_no_tls in production", func(r *CreateRequest) { r.Mode = "production"; r.AllowNoTLS = true }, ""},
		{"timeout too low", func(r *CreateRequest) { r.HealthTimeout = n(29) }, "health_timeout_seconds"},
		{"timeout too high", func(r *CreateRequest) { r.HealthTimeout = n(901) }, "health_timeout_seconds"},
		{"timeout min", func(r *CreateRequest) { r.HealthTimeout = n(30) }, ""},
		{"timeout max", func(r *CreateRequest) { r.HealthTimeout = n(900) }, ""},
	}
	for _, c := range cases {
		r := validRequest("bai")
		c.mut(&r)
		r.Normalize()
		verr := ValidateFormat(&r)
		switch {
		case c.field == "" && verr != nil:
			t.Errorf("%s: unexpected error %v", c.name, verr.Fields)
		case c.field != "" && (verr == nil || verr.Fields[c.field] == ""):
			t.Errorf("%s: want error on %s, got %v", c.name, c.field, verr)
		}
	}
}

func TestScriptArgsTLSAndHealthTimeout(t *testing.T) {
	n := 300
	j := &Job{Slug: "bai", ClientName: "PT Bai", Mode: "production", Domain: "bai.auditsphere.app",
		AdminEmail: "cae@bai.co.id", ProvisionEmail: true, AllowNoTLS: true, HealthTimeout: &n}
	want := []string{"--progress", "--domain", "bai.auditsphere.app", "--admin-email", "cae@bai.co.id", "--empty-data",
		"--health-timeout", "300", "--allow-no-tls", "--", "bai", "PT Bai"}
	if got := ScriptArgs(j); !reflect.DeepEqual(got, want) {
		t.Fatalf("args =\n%q\nwant\n%q", got, want)
	}
	// Local mode never passes --allow-no-tls, even for an old stored job.
	j.Mode, j.HealthTimeout = "local", nil
	for _, a := range ScriptArgs(j) {
		if a == "--allow-no-tls" || a == "--health-timeout" {
			t.Fatalf("local args contain %s", a)
		}
	}
}

func TestJobStoresTLSFieldsAndHTTPResult(t *testing.T) {
	f := newFixture(t, "")
	n := 600
	req := validRequest("prodnotls")
	req.Mode, req.AllowNoTLS, req.HealthTimeout = "production", true, &n
	job, err := f.m.Create(req)
	if err != nil {
		t.Fatal(err)
	}
	if !job.AllowNoTLS || job.HealthTimeout == nil || *job.HealthTimeout != 600 {
		t.Fatalf("job fields = allow_no_tls %v, health_timeout %v", job.AllowNoTLS, job.HealthTimeout)
	}
	done := waitFor(t, f.m, job.ID, terminal)
	if done.Status != JobSucceeded {
		t.Fatalf("status = %s, error = %v", done.Status, done.Error)
	}
	if done.Steps[8].Status != StepSkipped {
		t.Fatalf("ingress = %s", done.Steps[8].Status)
	}
	r := done.Result
	if r.FrontendURL != "http://prodnotls.example.com" || r.APIURL != "http://api-prodnotls.example.com" || r.TLS {
		t.Fatalf("result = %+v", r)
	}
	b, _ := json.Marshal(done)
	for _, want := range []string{`"allow_no_tls":true`, `"health_timeout_seconds":600`, `"tls":false`} {
		if !strings.Contains(string(b), want) {
			t.Errorf("job JSON lacks %s", want)
		}
	}
	// Persisted and reloaded.
	reloaded, errs := f.store.LoadJobs()
	if len(errs) > 0 {
		t.Fatal(errs)
	}
	for _, rj := range reloaded {
		if rj.ID == job.ID && (!rj.AllowNoTLS || rj.HealthTimeout == nil || rj.Result == nil || rj.Result.TLS) {
			t.Fatalf("stored job = %+v", rj)
		}
	}

	// With TLS: step 9 done, https kept, tls true.
	req = validRequest("prodtls")
	req.Mode = "production"
	job, err = f.m.Create(req)
	if err != nil {
		t.Fatal(err)
	}
	done = waitFor(t, f.m, job.ID, terminal)
	if done.Result == nil || !done.Result.TLS || done.Result.FrontendURL != "https://prodtls.example.com" {
		t.Fatalf("tls result = %+v", done.Result)
	}
}

func TestScriptDoesNotInheritProvisionerToken(t *testing.T) {
	t.Setenv("PROVISIONER_TOKEN", "super-secret-token-0123456789abcdef")
	t.Setenv("ADMIN_PASSWORD", "leak-me")
	t.Setenv("DOCKER_HOST", "unix:///run/docker.sock")
	t.Setenv("BASE_DOMAIN", "auditsphere.app")
	r := fakeRunner(t, t.TempDir(), "")
	var envLine string
	code, err := r.Run(context.Background(), []string{"--", "envcheck", "X"}, func(stream, line string) {
		if strings.HasPrefix(line, "ENV:") {
			envLine = strings.TrimPrefix(line, "ENV:")
		}
	})
	if err != nil || code != 0 {
		t.Fatalf("run: code %d err %v", code, err)
	}
	var env []string
	if err := json.Unmarshal([]byte(envLine), &env); err != nil {
		t.Fatalf("parse env %q: %v", envLine, err)
	}
	seen := map[string]bool{}
	for _, kv := range env {
		name, _, _ := strings.Cut(kv, "=")
		seen[strings.ToUpper(name)] = true
	}
	for _, bad := range []string{"PROVISIONER_TOKEN", "ADMIN_PASSWORD"} {
		if seen[bad] {
			t.Errorf("%s reached the script", bad)
		}
	}
	for _, good := range []string{"PATH", "DOCKER_HOST", "BASE_DOMAIN", "GO_PROVISIONER_HELPER"} {
		if !seen[good] {
			t.Errorf("%s did not reach the script", good)
		}
	}
	if os.Getenv("PROVISIONER_TOKEN") == "" {
		t.Fatal("setup: token not set")
	}
}

func TestFilterEnvLinux(t *testing.T) {
	in := []string{"PATH=/usr/bin", "PROVISIONER_TOKEN=x", "PROVISIONER_DATA_DIR=/var/lib/p", "https_proxy=http://p:3128",
		"LC_ALL=C.UTF-8", "COMPOSE_PROJECT_NAME=x", "RESEND_MASTER_API_KEY=re_x", `SystemRoot=C:Windows`, "path=lower", "noequals"}
	got := filterEnv(in, false)
	want := []string{"PATH=/usr/bin", "https_proxy=http://p:3128", "LC_ALL=C.UTF-8", "COMPOSE_PROJECT_NAME=x"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("filterEnv = %q, want %q", got, want)
	}
	if got := filterEnv([]string{`SystemRoot=C:Windows`, `Path=C:bin`, "PROVISIONER_TOKEN=x"}, true); len(got) != 2 {
		t.Fatalf("windows filterEnv = %q", got)
	}
}
