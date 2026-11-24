package provisioner

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSlugValidation(t *testing.T) {
	cases := map[string]bool{
		"bai": true, "ab": true, "a1": true, "bank2024": true, "a" + strings.Repeat("b", 30): true,
		"a": false, "1abc": false, "Bai": false, "ba-i": false, "ba_i": false, "ba i": false,
		"a" + strings.Repeat("b", 31): false, "": false, "bai;rm": false, "bai$(x)": false, "bai\n": false,
	}
	for slug, ok := range cases {
		r := validRequest(slug)
		r.Normalize()
		r.Slug = slug // Normalize trims; test the raw value
		err := ValidateFormat(&r)
		if ok && err != nil {
			t.Errorf("slug %q rejected: %v", slug, err)
		}
		if !ok && (err == nil || err.Fields["slug"] == "") {
			t.Errorf("slug %q accepted", slug)
		}
	}
}

func TestClientNameIsData(t *testing.T) {
	accepted := []string{
		`PT "Bank" Arta`,
		`O'Brien & Sons`,
		`$(rm -rf /)`,
		"`id`",
		`Bank; echo pwned | sh`,
		`${HOME} > /etc/passwd`,
		`PT Bank Arta Indonesia (Persero) Tbk.`,
		"Bank Ünïcødé 银行",
	}
	for _, name := range accepted {
		r := validRequest("bai")
		r.ClientName = name
		r.Normalize()
		if err := ValidateFormat(&r); err != nil {
			t.Errorf("client_name %q rejected: %v", name, err)
		}
	}
	rejected := []string{"", "a", strings.Repeat("x", 121), "line\nbreak", "tab\there", "nul\x00", "--local", "-x"}
	for _, name := range rejected {
		r := validRequest("bai")
		r.ClientName = name
		r.Normalize()
		if err := ValidateFormat(&r); err == nil || err.Fields["client_name"] == "" {
			t.Errorf("client_name %q accepted", name)
		}
	}
}

func TestOtherFieldValidation(t *testing.T) {
	p := func(v int) *int { return &v }
	cases := []struct {
		name  string
		mod   func(*CreateRequest)
		field string
	}{
		{"bad mode", func(r *CreateRequest) { r.Mode = "staging" }, "mode"},
		{"bad email", func(r *CreateRequest) { r.AdminEmail = "not-an-email" }, "admin_email"},
		{"email with name part", func(r *CreateRequest) { r.AdminEmail = "Siti <cae@bai.co.id>" }, "admin_email"},
		{"email with shell chars", func(r *CreateRequest) { r.AdminEmail = "a$(id)@bai.co.id" }, "admin_email"},
		{"missing email", func(r *CreateRequest) { r.AdminEmail = "" }, "admin_email"},
		{"only one port", func(r *CreateRequest) { r.FEPort = p(3015) }, "fe_port"},
		{"low port", func(r *CreateRequest) { r.FEPort, r.APIPort = p(80), p(8095) }, "fe_port"},
		{"same ports", func(r *CreateRequest) { r.FEPort, r.APIPort = p(3015), p(3015) }, "api_port"},
		{"bad domain", func(r *CreateRequest) { r.Mode, r.Domain = "production", "bai example.com" }, "domain"},
		{"domain not for slug", func(r *CreateRequest) { r.Mode, r.Domain = "production", "other.example.com" }, "domain"},
		{"no created_by", func(r *CreateRequest) { r.CreatedBy = "" }, "created_by"},
		{"admin name flag", func(r *CreateRequest) { r.AdminName = "--skip-email" }, "admin_name"},
	}
	for _, tc := range cases {
		r := validRequest("bai")
		tc.mod(&r)
		r.Normalize()
		err := ValidateFormat(&r)
		if err == nil || err.Fields[tc.field] == "" {
			t.Errorf("%s: want error on %s, got %v", tc.name, tc.field, err)
		}
	}
	r := validRequest("bai")
	r.Mode, r.Domain, r.FEPort, r.APIPort = "production", "bai.example.com", p(3015), p(8095)
	r.Normalize()
	if err := ValidateFormat(&r); err != nil {
		t.Errorf("valid production request rejected: %v", err)
	}
	if r.ProvisionEmail == nil || !*r.ProvisionEmail {
		t.Error("provision_email should default to true")
	}
}

func TestAvailability(t *testing.T) {
	dir := t.TempDir()
	reg := filepath.Join(dir, "registry.tsv")
	content := "slug\tclient_name\tdomain\tapi_domain\tfe_port\tkong_port\tkong_admin_port\tredis_db\tcreated_at\n" +
		"bai\tPT Bank\tbai.auditsphere.app\tapi-bai.auditsphere.app\t3010\t8090\t8019\t1\t2026-01-01T00:00:00Z\n"
	if err := os.WriteFile(reg, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	taken, err := ReadRegistry(reg)
	if err != nil {
		t.Fatal(err)
	}
	p := func(v int) *int { return &v }

	r := validRequest("bai")
	if err := ValidateAvailability(&r, taken, nil); err == nil || !err.Conflict || err.Fields["slug"] == "" {
		t.Errorf("existing slug accepted: %v", err)
	}
	r = validRequest("newco")
	r.FEPort, r.APIPort = p(3010), p(8091)
	if err := ValidateAvailability(&r, taken, nil); err == nil || err.Fields["fe_port"] == "" {
		t.Errorf("registry port accepted: %v", err)
	}
	r.FEPort = p(3011)
	busy := func(port int) bool { return port != 8091 }
	if err := ValidateAvailability(&r, taken, busy); err == nil || err.Fields["api_port"] == "" {
		t.Errorf("OS-busy port accepted: %v", err)
	}
	if err := ValidateAvailability(&r, taken, func(int) bool { return true }); err != nil {
		t.Errorf("free ports rejected: %v", err)
	}
	if _, err := ReadRegistry(filepath.Join(dir, "missing.tsv")); err != nil {
		t.Errorf("missing registry should be empty, got %v", err)
	}
}

func TestScriptArgs(t *testing.T) {
	p := func(v int) *int { return &v }
	j := &Job{Slug: "bai", ClientName: `PT "Bai" $(x)`, Mode: "production", Domain: "bai.example.com",
		FEPort: p(3015), APIPort: p(8095), AdminName: "Siti CAE", AdminEmail: "cae@bai.co.id", DryRun: true}
	got := ScriptArgs(j)
	want := []string{"--progress", "--dry-run", "--domain", "bai.example.com", "--fe-port", "3015", "--api-port", "8095",
		"--admin-name", "Siti CAE", "--admin-email", "cae@bai.co.id", "--empty-data", "--skip-email", "--", "bai", `PT "Bai" $(x)`}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args =\n%q\nwant\n%q", got, want)
	}
	j = &Job{Slug: "bai", ClientName: "Bai", Mode: "local", Domain: "ignored.example.com", AdminEmail: "a@b.co", WithDemoData: true, ProvisionEmail: true}
	got = ScriptArgs(j)
	want = []string{"--progress", "--local", "--admin-email", "a@b.co", "--with-demo-data", "--", "bai", "Bai"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("local args = %q, want %q", got, want)
	}
}
