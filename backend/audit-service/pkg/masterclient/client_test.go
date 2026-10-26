package masterclient

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"audit-service/pkg/serviceurl"

	"github.com/google/uuid"
)

const token = "Bearer caller-token"

// fakeMaster serves /api/v1/companies/:id and /api/v1/departments and records
// what it was asked
type fakeMaster struct {
	mu        sync.Mutex
	hits      int32
	auth      []string
	queries   []string
	status    int
	companies map[string]string // id -> JSON body of data
	depts     string            // JSON body of data
}

func (f *fakeMaster) handler(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt32(&f.hits, 1)
	f.mu.Lock()
	f.auth = append(f.auth, r.Header.Get("Authorization"))
	f.queries = append(f.queries, r.URL.RawQuery)
	f.mu.Unlock()
	if f.status != 0 && f.status != http.StatusOK {
		w.WriteHeader(f.status)
		_, _ = io.WriteString(w, `{"success":false,"message":"unauthorized"}`)
		return
	}
	switch {
	case strings.HasPrefix(r.URL.Path, "/api/v1/companies/"):
		body, ok := f.companies[strings.TrimPrefix(r.URL.Path, "/api/v1/companies/")]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = io.WriteString(w, `{"success":true,"data":`+body+`}`)
	case r.URL.Path == "/api/v1/departments":
		_, _ = io.WriteString(w, `{"success":true,"data":`+f.depts+`}`)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func newFake(t *testing.T, f *fakeMaster) *MasterClient {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(f.handler))
	t.Cleanup(srv.Close)
	return New(srv.URL, srv.Client())
}

func TestResolveForwardsAuthorizationToMaster(t *testing.T) {
	id := uuid.New()
	f := &fakeMaster{companies: map[string]string{
		id.String(): `{"id":"` + id.String() + `","name":"Acme","legal_name":"PT Acme Tbk"}`,
	}}
	c := newFake(t, f)

	got := resolveCompanyName(context.Background(), c, token, &id, "", nil, "", "", "")
	if got != "PT Acme Tbk" {
		t.Fatalf("got %q, want legal name", got)
	}
	if len(f.auth) != 1 || f.auth[0] != token {
		t.Fatalf("master saw Authorization %q, want %q", f.auth, token)
	}

	// cached: a second resolve does not call master again
	if got := resolveCompanyName(context.Background(), c, token, &id, "", nil, "", "", ""); got != "PT Acme Tbk" {
		t.Fatalf("second resolve got %q", got)
	}
	if n := atomic.LoadInt32(&f.hits); n != 1 {
		t.Fatalf("master hit %d times, want 1 (cached)", n)
	}
}

func TestResolveUnauthorizedIsEmptyAndNotCached(t *testing.T) {
	id := uuid.New()
	f := &fakeMaster{status: http.StatusUnauthorized}
	c := newFake(t, f)

	for i := 0; i < 2; i++ {
		got := resolveCompanyName(context.Background(), c, token, &id, "", &id, "", "Finance", "")
		if got != "" {
			t.Fatalf("call %d: got %q, want empty", i, got)
		}
		if strings.Contains(strings.ToUpper(got), "AIFL") {
			t.Fatalf("hardcoded fallback leaked: %q", got)
		}
	}
	// each call asks master again (report id, letter id, departments) — no
	// negative or 401 result was cached
	if n := atomic.LoadInt32(&f.hits); n != 6 {
		t.Fatalf("master hit %d times, want 6", n)
	}
	c.cacheMu.RLock()
	defer c.cacheMu.RUnlock()
	if len(c.compCache) != 0 || len(c.deptCache) != 0 {
		t.Fatalf("401 answers were cached: comp=%d dept=%d", len(c.compCache), len(c.deptCache))
	}
}

func TestResolveWithoutTokenDoesNotCallMaster(t *testing.T) {
	id := uuid.New()
	f := &fakeMaster{}
	c := newFake(t, f)

	if got := resolveCompanyName(context.Background(), c, "  ", &id, "", nil, "", "Finance", ""); got != "" {
		t.Fatalf("got %q, want empty", got)
	}
	if n := atomic.LoadInt32(&f.hits); n != 0 {
		t.Fatalf("master was called %d times without a token", n)
	}
	if _, err := c.GetCompanyByID(context.Background(), "", id.String()); !errors.Is(err, ErrNoAuthorization) {
		t.Fatalf("GetCompanyByID without token: err = %v", err)
	}
}

func TestResolveExplicitNamesSkipMaster(t *testing.T) {
	f := &fakeMaster{}
	c := newFake(t, f)
	if got := resolveCompanyName(context.Background(), c, token, nil, " PT Report ", nil, "PT Letter", "", ""); got != "PT Report" {
		t.Fatalf("got %q", got)
	}
	if got := resolveCompanyName(context.Background(), c, token, nil, "", nil, "PT Letter", "", ""); got != "PT Letter" {
		t.Fatalf("got %q", got)
	}
	if n := atomic.LoadInt32(&f.hits); n != 0 {
		t.Fatalf("master called %d times", n)
	}
}

func TestResolveByWorkingUnitFollowsCompanyID(t *testing.T) {
	compID := uuid.New().String()
	// master-service's department list does not preload company, so it arrives
	// as a zero value; the client must follow company_id instead
	f := &fakeMaster{
		depts: `{"departments":[
			{"id":"d0","department_code":"","department_name":"","company_id":"` + uuid.New().String() + `"},
			{"id":"d1","department_code":"FIN","department_name":"Finance","company_id":"` + compID + `",
			 "company":{"id":"00000000-0000-0000-0000-000000000000","name":"","legal_name":""}}
		],"pagination":{"page":1}}`,
		companies: map[string]string{compID: `{"id":"` + compID + `","name":"Beta"}`},
	}
	c := newFake(t, f)

	// substring match; the nameless department d0 must not match everything
	if got := resolveCompanyName(context.Background(), c, token, nil, "", nil, "", "", "Finance Dept"); got != "Beta" {
		t.Fatalf("got %q, want Beta", got)
	}
	if !strings.Contains(f.queries[0], "page_size=") {
		t.Fatalf("departments fetched without page_size: %q", f.queries[0])
	}
	for _, a := range f.auth {
		if a != token {
			t.Fatalf("a request went out with Authorization %q", a)
		}
	}
	hits := atomic.LoadInt32(&f.hits)
	if got := resolveCompanyName(context.Background(), c, token, nil, "", nil, "", "FIN", ""); got != "Beta" {
		t.Fatalf("by code got %q", got)
	}
	if n := atomic.LoadInt32(&f.hits); n != hits {
		t.Fatalf("cached department/company lookup called master again (%d -> %d)", hits, n)
	}
}

func TestEmptyBaseURLNeverCalls(t *testing.T) {
	c := New("", nil)
	if _, err := c.GetCompanyByID(context.Background(), token, uuid.New().String()); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("err = %v, want ErrNotConfigured", err)
	}
	id := uuid.New()
	if got := resolveCompanyName(context.Background(), c, token, &id, "", nil, "", "", ""); got != "" {
		t.Fatalf("got %q", got)
	}
}

// roundTripFunc lets a test see where a request was sent without DNS
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestTenantClientCallsTenantMaster(t *testing.T) {
	t.Setenv(serviceurl.MasterServiceURLEnv, "")
	t.Setenv(serviceurl.ServiceNameEnv, "audit-service-bai")

	var gotHost, gotAuth string
	hc := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotHost, gotAuth = r.URL.Host, r.Header.Get("Authorization")
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{},
			Body: io.NopCloser(strings.NewReader(`{"success":true,"data":{"name":"PT BAI"}}`))}, nil
	})}
	c := New(serviceurl.Master(), hc)

	comp, err := c.GetCompanyByID(context.Background(), token, uuid.New().String())
	if err != nil || comp.displayName() != "PT BAI" {
		t.Fatalf("comp=%v err=%v", comp, err)
	}
	if gotHost != "master-service-bai:8002" {
		t.Fatalf("request went to %q, want master-service-bai:8002", gotHost)
	}
	if gotAuth != token {
		t.Fatalf("Authorization = %q", gotAuth)
	}
}
