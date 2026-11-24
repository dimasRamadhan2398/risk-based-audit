package provisioner

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const testToken = "test-token-0123456789abcdef0123456789"

func newTestServer(t *testing.T, f *fixture) *httptest.Server {
	t.Helper()
	s := &Server{
		Manager: f.m, Token: testToken, Logger: quietLogger(), Heartbeat: 50 * time.Millisecond,
		Ports: func(context.Context) (NextPorts, error) { return NextPorts{FEPort: 3015, APIPort: 8095}, nil },
	}
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	return ts
}

func do(t *testing.T, method, url, token, user, body string) *http.Response {
	t.Helper()
	req, _ := http.NewRequest(method, url, strings.NewReader(body))
	if token != "" {
		req.Header.Set(TokenHeader, token)
	}
	if user != "" {
		req.Header.Set(UserHeader, user)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func TestTokenRequired(t *testing.T) {
	ts := newTestServer(t, newFixture(t, ""))
	for _, path := range []string{"/jobs", "/jobs/x", "/jobs/x/events", "/ports/next"} {
		if r := do(t, "GET", ts.URL+path, "", "", ""); r.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s without token: %d", path, r.StatusCode)
		}
		if r := do(t, "GET", ts.URL+path, testToken+"x", "", ""); r.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s with wrong token: %d", path, r.StatusCode)
		}
	}
	if r := do(t, "GET", ts.URL+"/healthz", "", "", ""); r.StatusCode != http.StatusOK {
		t.Errorf("healthz: %d", r.StatusCode)
	}
	if r := do(t, "GET", ts.URL+"/ports/next", testToken, "", ""); r.StatusCode != http.StatusOK {
		t.Errorf("ports/next: %d", r.StatusCode)
	}
}

func TestCreateValidationOverHTTP(t *testing.T) {
	ts := newTestServer(t, newFixture(t, ""))
	r := do(t, "POST", ts.URL+"/jobs", testToken, "", `{"slug":"Bad Slug","client_name":"x","mode":"local","admin_email":"no","created_by":"u"}`)
	if r.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", r.StatusCode)
	}
	var body struct {
		Error struct{ Code string }
		Data  struct{ Fields map[string]string }
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Error.Code != "VALIDATION_FAILED" || body.Data.Fields["slug"] == "" || body.Data.Fields["client_name"] == "" || body.Data.Fields["admin_email"] == "" {
		t.Fatalf("body = %+v", body)
	}
	if r := do(t, "POST", ts.URL+"/jobs", testToken, "", `{bad`); r.StatusCode != http.StatusBadRequest {
		t.Errorf("bad json: %d", r.StatusCode)
	}
	if r := do(t, "GET", ts.URL+"/jobs/00000000-0000-4000-8000-000000000000", testToken, "", ""); r.StatusCode != http.StatusNotFound {
		t.Errorf("unknown job: %d", r.StatusCode)
	}
}

type sseEvent struct {
	id, event, data string
}

func readSSE(t *testing.T, resp *http.Response) (events []sseEvent, heartbeats int) {
	t.Helper()
	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 64*1024), 1<<20)
	var cur sseEvent
	for sc.Scan() {
		line := sc.Text()
		switch {
		case line == "":
			if cur.event != "" {
				events = append(events, cur)
			}
			cur = sseEvent{}
		case strings.HasPrefix(line, ":"):
			heartbeats++
		case strings.HasPrefix(line, "id: "):
			cur.id = strings.TrimPrefix(line, "id: ")
		case strings.HasPrefix(line, "event: "):
			cur.event = strings.TrimPrefix(line, "event: ")
		case strings.HasPrefix(line, "data: "):
			cur.data = strings.TrimPrefix(line, "data: ")
		}
	}
	return events, heartbeats
}

func TestSSEStreamsLiveAndEnds(t *testing.T) {
	f := newFixture(t, "")
	ts := newTestServer(t, f)
	r := do(t, "POST", ts.URL+"/jobs", testToken, "", `{"slug":"blockb","client_name":"PT \"B\" $(id)","mode":"local","admin_email":"cae@b.co.id","created_by":"user-1"}`)
	if r.StatusCode != http.StatusAccepted {
		t.Fatalf("create: %d", r.StatusCode)
	}
	var job Job
	_ = json.NewDecoder(r.Body).Decode(&job)
	if job.ClientName != `PT "B" $(id)` {
		t.Errorf("client_name = %q", job.ClientName)
	}

	resp := do(t, "GET", ts.URL+"/jobs/"+job.ID+"/events", testToken, "", "")
	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content type %q", ct)
	}
	// Let a few heartbeats through while the job is blocked, then release it.
	go func() { time.Sleep(300 * time.Millisecond); f.releaseBlocked(t) }()
	events, heartbeats := readSSE(t, resp) // returns when the server ends the stream
	if heartbeats == 0 {
		t.Error("no heartbeats while the job was blocked")
	}
	if len(events) < 3 {
		t.Fatalf("events = %+v", events)
	}
	last := events[len(events)-1]
	if last.event != EventStatus || !strings.Contains(last.data, JobSucceeded) {
		t.Fatalf("last event = %+v", last)
	}
	if events[0].id != "1" || events[0].event != EventStatus {
		t.Errorf("first event = %+v", events[0])
	}

	// Resume after event 2: only the later ones are replayed, then the stream ends.
	resp = do(t, "GET", ts.URL+"/jobs/"+job.ID+"/events?since=2", testToken, "", "")
	replay, _ := readSSE(t, resp)
	if len(replay) != len(events)-2 || replay[0].id != "3" {
		t.Errorf("replay since=2: %d events (want %d), first %+v", len(replay), len(events)-2, replay)
	}
	if r := do(t, "GET", ts.URL+"/jobs/"+job.ID+"/events?since=-1", testToken, "", ""); r.StatusCode != http.StatusBadRequest {
		t.Errorf("since=-1: %d", r.StatusCode)
	}
}

func TestPasswordOnceOverHTTP(t *testing.T) {
	f := newFixture(t, "")
	ts := newTestServer(t, f)
	r := do(t, "POST", ts.URL+"/jobs", testToken, "", `{"slug":"bai","client_name":"PT Bai","mode":"local","admin_email":"cae@b.co.id","created_by":"user-1"}`)
	var job Job
	_ = json.NewDecoder(r.Body).Decode(&job)
	waitFor(t, f.m, job.ID, terminal)

	get := func(user string) *string {
		var j Job
		_ = json.NewDecoder(do(t, "GET", ts.URL+"/jobs/"+job.ID, testToken, user, "").Body).Decode(&j)
		if j.Result == nil {
			t.Fatal("no result")
		}
		return j.Result.AdminInitialPassword
	}
	if get("") != nil || get("user-2") != nil {
		t.Fatal("password returned to a non-creator")
	}
	if p := get("user-1"); p == nil || *p != fakePassword {
		t.Fatalf("creator got %v", p)
	}
	if get("user-1") != nil {
		t.Fatal("password returned twice")
	}
	resp := do(t, "GET", ts.URL+"/jobs/"+job.ID+"/events", testToken, "user-1", "")
	events, _ := readSSE(t, resp)
	for _, ev := range events {
		if strings.Contains(ev.data, fakePassword) {
			t.Fatalf("SSE leaks the password: %+v", ev)
		}
	}
}

func TestNextPortsError(t *testing.T) {
	f := newFixture(t, "")
	s := &Server{Manager: f.m, Token: testToken, Logger: quietLogger(),
		Ports: func(context.Context) (NextPorts, error) { return NextPorts{}, errors.New("boom") }}
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()
	if r := do(t, "GET", ts.URL+"/ports/next", testToken, "", ""); r.StatusCode != http.StatusBadGateway {
		t.Errorf("status = %d", r.StatusCode)
	}
}
