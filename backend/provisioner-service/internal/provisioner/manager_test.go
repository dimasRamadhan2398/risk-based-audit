package provisioner

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSuccessfulJobAndPasswordOnce(t *testing.T) {
	f := newFixture(t, "")
	job, err := f.m.Create(validRequest("bai"))
	if err != nil {
		t.Fatal(err)
	}
	if job.Status != JobQueued || len(job.Steps) != 9 || job.Steps[0].Status != StepPending {
		t.Fatalf("new job = %+v", job)
	}
	done := waitFor(t, f.m, job.ID, terminal)
	if done.Status != JobSucceeded {
		t.Fatalf("status = %s, error = %v", done.Status, done.Error)
	}
	for _, s := range done.Steps {
		want := StepDone
		if s.Key == "ingress" {
			want = StepSkipped
		}
		if s.Status != want || s.StartedAt == nil || s.FinishedAt == nil {
			t.Errorf("step %s = %+v", s.Key, s)
		}
	}
	if done.Result == nil || done.Result.FrontendURL != "http://bai.localhost:3015" || done.Result.AdminInitialPassword != nil {
		t.Fatalf("result for anonymous read = %+v", done.Result)
	}

	// Someone other than the creator never sees it.
	other, _ := f.m.Get(job.ID, "user-2")
	if other.Result.AdminInitialPassword != nil {
		t.Fatal("password returned to a non-creator")
	}
	// List never carries it.
	for _, j := range f.m.List(10) {
		if j.Result != nil && j.Result.AdminInitialPassword != nil {
			t.Fatal("password in list")
		}
	}
	// The creator gets it once.
	first, _ := f.m.Get(job.ID, "user-1")
	if first.Result.AdminInitialPassword == nil || *first.Result.AdminInitialPassword != fakePassword {
		t.Fatalf("creator did not get the password: %+v", first.Result)
	}
	second, _ := f.m.Get(job.ID, "user-1")
	if second.Result.AdminInitialPassword != nil {
		t.Fatal("password returned twice")
	}

	// Never on disk, never in an event.
	for _, name := range []string{"job.json", "events.jsonl"} {
		b, err := os.ReadFile(filepath.Join(f.dataDir, "jobs", job.ID, name))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(b), fakePassword) {
			t.Errorf("%s contains the password", name)
		}
	}
	evs, finished, _, err := f.m.EventsAfter(job.ID, 0)
	if err != nil || !finished {
		t.Fatalf("events: finished=%v err=%v", finished, err)
	}
	var sawResult, sawStderr bool
	for i, ev := range evs {
		if ev.Seq != int64(i+1) {
			t.Fatalf("seq %d at position %d", ev.Seq, i)
		}
		if strings.Contains(string(ev.Data), fakePassword) {
			t.Errorf("event %d leaks the password: %s", ev.Seq, ev.Data)
		}
		if ev.Type == EventResult {
			sawResult = true
		}
		if ev.Type == EventLog && strings.Contains(string(ev.Data), `"stream":"stderr"`) {
			sawStderr = true
		}
	}
	last := evs[len(evs)-1]
	if last.Type != EventStatus || !strings.Contains(string(last.Data), JobSucceeded) {
		t.Errorf("last event = %s %s", last.Type, last.Data)
	}
	if !sawResult || !sawStderr {
		t.Errorf("result event %v, stderr log %v", sawResult, sawStderr)
	}
}

func TestClientNameReachesScriptVerbatim(t *testing.T) {
	f := newFixture(t, "")
	name := "PT \"Quote\" 'single' $(touch pwned1) `touch pwned2`; touch pwned3 && echo ${HOME}"
	req := validRequest("echoargs")
	req.ClientName = name
	req.AdminName = "Siti $(touch pwned4)"
	job, err := f.m.Create(req)
	if err != nil {
		t.Fatal(err)
	}
	if done := waitFor(t, f.m, job.ID, terminal); done.Status != JobSucceeded {
		t.Fatalf("status = %s %v", done.Status, done.Error)
	}
	evs, _, _, _ := f.m.EventsAfter(job.ID, 0)
	var argv []string
	for _, ev := range evs {
		var ld LogData
		if ev.Type == EventLog && json.Unmarshal(ev.Data, &ld) == nil && strings.HasPrefix(ld.Line, "ARGS:") {
			if err := json.Unmarshal([]byte(strings.TrimPrefix(ld.Line, "ARGS:")), &argv); err != nil {
				t.Fatal(err)
			}
		}
	}
	if n := len(argv); n < 3 || argv[n-3] != "--" || argv[n-2] != "echoargs" || argv[n-1] != name {
		t.Fatalf("script argv = %q", argv)
	}
	if !contains(argv, "--progress") || !contains(argv, "Siti $(touch pwned4)") {
		t.Fatalf("script argv = %q", argv)
	}
	for _, p := range []string{"pwned1", "pwned2", "pwned3", "pwned4"} {
		if _, err := os.Stat(filepath.Join(f.workDir, p)); err == nil {
			t.Errorf("%s exists: an argument was interpreted by a shell", p)
		}
	}
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}

func TestFailedJob(t *testing.T) {
	f := newFixture(t, "")
	job, err := f.m.Create(validRequest("failco"))
	if err != nil {
		t.Fatal(err)
	}
	done := waitFor(t, f.m, job.ID, terminal)
	if done.Status != JobFailed || done.Error == nil || !strings.Contains(*done.Error, "psql: connection refused") {
		t.Fatalf("job = %s %v", done.Status, done.Error)
	}
	if done.ExitCode == nil || *done.ExitCode != 3 {
		t.Errorf("exit code = %v", done.ExitCode)
	}
	if done.Steps[0].Status != StepDone || done.Steps[1].Status != StepFailed || done.Steps[2].Status != StepPending {
		t.Errorf("steps = %+v", done.Steps)
	}
	if done.CurrentStep != "databases" || done.Result != nil {
		t.Errorf("current_step = %q result = %+v", done.CurrentStep, done.Result)
	}
}

func TestQueueRunsOneJobAtATime(t *testing.T) {
	f := newFixture(t, "")
	a, err := f.m.Create(validRequest("blocka"))
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, f.m, a.ID, func(j Job) bool { return j.Status == JobRunning && j.Steps[0].Status == StepRunning })

	b, err := f.m.Create(validRequest("bai"))
	if err != nil {
		t.Fatal(err)
	}
	// The same slug cannot be queued twice while a job holds it.
	if _, err := f.m.Create(validRequest("blocka")); err == nil {
		t.Fatal("duplicate active slug accepted")
	} else {
		var verr *ValidationError
		if !errors.As(err, &verr) || !verr.Conflict {
			t.Fatalf("duplicate slug error = %v", err)
		}
	}
	time.Sleep(200 * time.Millisecond)
	if j, _ := f.m.Get(b.ID, ""); j.Status != JobQueued || j.StartedAt != nil {
		t.Fatalf("second job started while the first was running: %s", j.Status)
	}

	f.releaseBlocked(t)
	ad := waitFor(t, f.m, a.ID, terminal)
	bd := waitFor(t, f.m, b.ID, terminal)
	if ad.Status != JobSucceeded || bd.Status != JobSucceeded {
		t.Fatalf("statuses = %s, %s", ad.Status, bd.Status)
	}
	if bd.StartedAt.Before(*ad.FinishedAt) {
		t.Errorf("second job started (%v) before the first finished (%v)", bd.StartedAt, ad.FinishedAt)
	}
	list := f.m.List(10)
	if len(list) != 2 || list[0].ID != b.ID {
		t.Errorf("list is not newest first: %v", list)
	}
}

func TestRestartFailsRunningAndRequeuesQueued(t *testing.T) {
	dataDir := t.TempDir()
	store, _ := NewStore(dataDir)
	now := time.Now().UTC()
	running := &Job{ID: newID(), Status: JobRunning, Slug: "aaa", Mode: "local", CreatedBy: "u", CreatedAt: now.Add(-2 * time.Minute), StartedAt: &now, Steps: initialSteps()}
	running.Steps[0].Status, running.Steps[0].StartedAt = StepRunning, &now
	queued := &Job{ID: newID(), Status: JobQueued, Slug: "bbb", Mode: "local", AdminEmail: "a@b.co", CreatedBy: "u", CreatedAt: now.Add(-time.Minute), Steps: initialSteps()}
	for _, j := range []*Job{running, queued} {
		if err := store.SaveJob(j); err != nil {
			t.Fatal(err)
		}
	}
	_ = store.AppendEvent(running.ID, Event{Seq: 1, Type: EventStatus, Data: json.RawMessage(`{"status":"running","error":null}`), TS: now})

	workDir := t.TempDir()
	m, err := NewManager(ManagerOptions{Store: store, Runner: fakeRunner(t, workDir, ""), Logger: quietLogger(), RegistryPath: filepath.Join(workDir, "r.tsv")})
	if err != nil {
		t.Fatal(err)
	}
	j, _ := m.Get(running.ID, "")
	if j.Status != JobFailed || j.Error == nil || *j.Error != ErrRestarted || j.FinishedAt == nil {
		t.Fatalf("restarted job = %s %v", j.Status, j.Error)
	}
	if j.Steps[0].Status != StepFailed {
		t.Errorf("running step = %+v", j.Steps[0])
	}
	evs, finished, _, _ := m.EventsAfter(running.ID, 0)
	if !finished || evs[len(evs)-1].Seq <= 1 || !strings.Contains(string(evs[len(evs)-1].Data), ErrRestarted) {
		t.Errorf("events after restart = %+v", evs)
	}
	// Persisted, so a second restart sees it as failed too.
	b, _ := os.ReadFile(filepath.Join(dataDir, "jobs", running.ID, "job.json"))
	if !strings.Contains(string(b), ErrRestarted) {
		t.Error("restart failure not persisted")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go m.Run(ctx)
	if q := waitFor(t, m, queued.ID, terminal); q.Status != JobSucceeded {
		t.Errorf("re-queued job = %s %v", q.Status, q.Error)
	}
}

func TestRetryAfterFailedJob(t *testing.T) {
	f := newFixture(t, "")
	first, err := f.m.Create(validRequest("failco"))
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, f.m, first.ID, terminal)
	ok, err := f.m.Create(validRequest("bai"))
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, f.m, ok.ID, terminal)

	// The failed run (and the successful one) registered their slugs.
	reg := "slug\tclient_name\tdomain\tapi_domain\tfe_port\tkong_port\tkong_admin_port\tredis_db\tcreated_at\n" +
		"failco\tF\tfailco.localhost\tlocalhost\t3010\t8090\t8019\t1\tx\n" +
		"bai\tB\tbai.localhost\tlocalhost\t3011\t8091\t8020\t2\tx\n" +
		"clionly\tC\tclionly.localhost\tlocalhost\t3012\t8092\t8021\t3\tx\n"
	if err := os.WriteFile(f.workDir+"/registry.tsv", []byte(reg), 0o600); err != nil {
		t.Fatal(err)
	}

	p := func(v int) *int { return &v }
	retry := validRequest("failco")
	retry.FEPort, retry.APIPort = p(3010), p(8090) // its own registered ports
	f.m.portFree = func(int) bool { return false } // even if the failed stack still holds them
	again, err := f.m.Create(retry)
	if err != nil {
		t.Fatalf("retry of a failed slug rejected: %v", err)
	}
	if again.RetryOf != first.ID {
		t.Errorf("retry_of = %q, want %q", again.RetryOf, first.ID)
	}
	f.m.portFree = nil
	// While the retry is queued/running, the slug is held again.
	if _, err := f.m.Create(validRequest("failco")); err == nil {
		t.Error("second concurrent retry accepted")
	}
	// A slug whose latest job succeeded, or that was never provisioned here, stays taken.
	for _, slug := range []string{"bai", "clionly"} {
		if _, err := f.m.Create(validRequest(slug)); err == nil {
			t.Errorf("%s accepted", slug)
		}
	}
	// Another slug may not take the failed tenant's ports.
	other := validRequest("newco")
	other.FEPort, other.APIPort = p(3010), p(8090)
	if _, err := f.m.Create(other); err == nil {
		t.Error("another slug took a registered port")
	}
	waitFor(t, f.m, again.ID, terminal)
}
