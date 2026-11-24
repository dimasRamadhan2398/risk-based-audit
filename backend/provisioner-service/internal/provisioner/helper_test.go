package provisioner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

// The fake onboard-tenant.sh is this test binary re-executed with
// GO_PROVISIONER_HELPER=1, so the tests need neither bash nor docker. The
// behaviour is chosen by the slug prefix (the first positional after "--"):
//
//	echo...   prints the received argv as JSON, then succeeds
//	fail...   fails at step 2
//	block...  reports step 1 running, then waits for $FAKE_RELEASE to exist
//	other     reports all 9 steps and a @@RESULT
const fakePassword = "S3cret-Initial-pw"

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_PROVISIONER_HELPER") != "1" {
		return
	}
	args := os.Args
	for i, a := range args {
		if a == "--" {
			args = args[i+1:]
			break
		}
	}
	os.Exit(fakeScript(args))
}

func progress(key string, idx int, status, msg string) {
	b, _ := json.Marshal(Progress{Key: key, Index: idx, Total: 9, Status: status, Message: msg})
	fmt.Println("@@PROGRESS " + string(b))
}

func fakeScript(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "no args")
		return 2
	}
	// The positionals follow the "--" that ends the options.
	var pos []string
	for i, a := range args {
		if a == "--" {
			pos = args[i+1:]
			break
		}
	}
	if len(pos) != 2 {
		fmt.Fprintln(os.Stderr, "want -- <slug> <client_name>")
		return 2
	}
	slug := pos[0]
	switch {
	case strings.HasPrefix(slug, "echo"):
		b, _ := json.Marshal(args)
		fmt.Println("ARGS:" + string(b))
		return 0
	case strings.HasPrefix(slug, "env"):
		b, _ := json.Marshal(os.Environ())
		fmt.Println("ENV:" + string(b))
		return 0
	case strings.HasPrefix(slug, "prod"):
		// Production: step 9 is skipped under --allow-no-tls. The fake still
		// reports https URLs, to check that the provisioner corrects them.
		noTLS := false
		for _, a := range args {
			noTLS = noTLS || a == "--allow-no-tls"
		}
		for _, s := range Steps {
			progress(s.Key, s.Index, StepRunning, "doing "+s.Key)
			if s.Key == "ingress" && noTLS {
				progress(s.Key, s.Index, StepSkipped, "TLS NOT CONFIGURED, plain HTTP only")
				continue
			}
			progress(s.Key, s.Index, StepDone, s.Key+" done")
		}
		b, _ := json.Marshal(ScriptResult{Slug: slug, FrontendURL: "https://" + slug + ".example.com", APIURL: "https://api-" + slug + ".example.com",
			AdminUsername: "admin", AdminEmail: "cae@example.com", AdminInitialPassword: fakePassword})
		fmt.Println("@@RESULT " + string(b))
		return 0
	case strings.HasPrefix(slug, "fail"):
		progress("generate", 1, StepRunning, "generating")
		progress("generate", 1, StepDone, "generated")
		progress("databases", 2, StepRunning, "creating databases")
		fmt.Fprintln(os.Stderr, "psql: connection refused")
		progress("databases", 2, StepFailed, "psql: connection refused")
		return 3
	case strings.HasPrefix(slug, "block"):
		progress("generate", 1, StepRunning, "waiting")
		deadline := time.Now().Add(20 * time.Second)
		for time.Now().Before(deadline) {
			if _, err := os.Stat(os.Getenv("FAKE_RELEASE")); err == nil {
				break
			}
			time.Sleep(20 * time.Millisecond)
		}
		progress("generate", 1, StepDone, "released")
		return 0
	}
	for _, s := range Steps {
		progress(s.Key, s.Index, StepRunning, "doing "+s.Key)
		if s.Key == "ingress" {
			progress(s.Key, s.Index, StepSkipped, "local mode")
			continue
		}
		progress(s.Key, s.Index, StepDone, s.Key+" done")
	}
	fmt.Println("plain log line ADMIN_PASSWORD=" + fakePassword)
	fmt.Fprintln(os.Stderr, "a warning on stderr")
	b, _ := json.Marshal(ScriptResult{Slug: slug, FrontendURL: "http://" + slug + ".localhost:3015", APIURL: "http://localhost:8095",
		AdminUsername: "admin", AdminEmail: "cae@example.com", AdminInitialPassword: fakePassword})
	fmt.Println("@@RESULT " + string(b))
	// After @@RESULT the provisioner knows the value and masks it anywhere.
	fmt.Println("echo of the password itself: " + fakePassword)
	return 0
}

func fakeRunner(t *testing.T, dir string, release string) *ExecRunner {
	t.Helper()
	return &ExecRunner{
		Command: []string{os.Args[0], "-test.run=^TestHelperProcess$", "--"},
		Dir:     dir,
		Env:     []string{"GO_PROVISIONER_HELPER=1", "FAKE_RELEASE=" + release},
	}
}

func quietLogger() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

type fixture struct {
	m       *Manager
	store   *Store
	dataDir string
	workDir string
	release string
	stop    func()
}

func newFixture(t *testing.T, registry string) *fixture {
	t.Helper()
	dataDir, workDir := t.TempDir(), t.TempDir()
	store, err := NewStore(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	regPath := workDir + "/registry.tsv"
	if registry != "" {
		if err := os.WriteFile(regPath, []byte(registry), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	release := workDir + "/release"
	m, err := NewManager(ManagerOptions{Store: store, Runner: fakeRunner(t, workDir, release), Logger: quietLogger(), RegistryPath: regPath})
	if err != nil {
		t.Fatal(err)
	}
	f := &fixture{m: m, store: store, dataDir: dataDir, workDir: workDir, release: release}
	f.stop = f.start()
	t.Cleanup(func() { f.stop() })
	return f
}

func (f *fixture) start() func() {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { defer close(done); f.m.Run(ctx) }()
	return func() { cancel(); <-done }
}

func (f *fixture) releaseBlocked(t *testing.T) {
	if err := os.WriteFile(f.release, nil, 0o600); err != nil {
		t.Fatal(err)
	}
}

func validRequest(slug string) CreateRequest {
	return CreateRequest{ClientName: "PT Bank Arta Indonesia", Slug: slug, Mode: "local", AdminName: "Siti CAE",
		AdminEmail: "cae@bai.co.id", CreatedBy: "user-1"}
}

func waitFor(t *testing.T, m *Manager, id string, cond func(Job) bool) Job {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		j, err := m.Get(id, "")
		if err != nil {
			t.Fatal(err)
		}
		if cond(j) {
			return j
		}
		time.Sleep(10 * time.Millisecond)
	}
	j, _ := m.Get(id, "")
	t.Fatalf("timed out waiting for job %s (status %s)", id, j.Status)
	return Job{}
}

func terminal(j Job) bool { return j.terminal() }
