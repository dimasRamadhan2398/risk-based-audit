package masterclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCallerDepartmentResolvesForwardsTokenAndCaches(t *testing.T) {
	dept := uuid.New()
	var hits int32
	var gotAuth atomic.Value
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		gotAuth.Store(r.Header.Get("Authorization"))
		if r.URL.Path != "/api/v1/employees/by-code/EMP-1" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"success":true,"data":{"id":"` + uuid.NewString() + `","department_id":"` + dept.String() + `"}}`))
	}))
	defer srv.Close()

	c := NewDepartmentClient(srv.URL, time.Second, time.Minute)
	now := time.Now()
	c.now = func() time.Time { return now }
	ctx := WithAuthorization(context.Background(), "Bearer tok")

	for i := 0; i < 3; i++ {
		got, err := c.CallerDepartment(ctx, "EMP-1", "")
		if err != nil || got != dept {
			t.Fatalf("got %v, %v", got, err)
		}
	}
	if h := atomic.LoadInt32(&hits); h != 1 || gotAuth.Load() != "Bearer tok" {
		t.Fatalf("hits=%d auth=%v, want 1 request with the caller token", h, gotAuth.Load())
	}
	now = now.Add(2 * time.Minute) // past the TTL: looked up again
	if _, err := c.CallerDepartment(ctx, "EMP-1", ""); err != nil || atomic.LoadInt32(&hits) != 2 {
		t.Fatalf("after TTL: hits=%d err=%v", hits, err)
	}
}

func TestCallerDepartmentFailsClosed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/employees/by-code/NODEPT" {
			_, _ = w.Write([]byte(`{"data":{"id":"x","department_id":""}}`))
			return
		}
		http.NotFound(w, r)
	}))
	c := NewDepartmentClient(srv.URL, time.Second, time.Minute)

	for name, args := range map[string][2]string{
		"no link": {"", ""}, "unknown employee": {"GHOST", ""}, "no department": {"NODEPT", ""}, "bad employee id": {"", "not-a-uuid"},
	} {
		if _, err := c.CallerDepartment(context.Background(), args[0], args[1]); err == nil {
			t.Errorf("%s: expected ErrDepartmentUnresolved", name)
		}
	}
	srv.Close()
	if _, err := c.CallerDepartment(context.Background(), "EMP-1", ""); err == nil {
		t.Error("master-service down: expected an error, not a department")
	}
}

func TestDepartmentNameCachedAndBestEffort(t *testing.T) {
	id := uuid.New()
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		if r.URL.Path != "/api/v1/departments/"+id.String() {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"data":{"department_name":"Finance"}}`))
	}))
	defer srv.Close()
	c := NewDepartmentClient(srv.URL, time.Second, time.Minute)
	for i := 0; i < 3; i++ {
		if n, ok := c.DepartmentName(context.Background(), id); !ok || n != "Finance" {
			t.Fatalf("name = %q, %v", n, ok)
		}
	}
	if h := atomic.LoadInt32(&hits); h != 1 {
		t.Fatalf("hits = %d, want 1", h)
	}
	if _, ok := c.DepartmentName(context.Background(), uuid.New()); ok {
		t.Error("unknown department must report ok=false")
	}
	// After a failure, further lookups are skipped briefly instead of each
	// paying the HTTP timeout.
	before := atomic.LoadInt32(&hits)
	if _, ok := c.DepartmentName(context.Background(), uuid.New()); ok || atomic.LoadInt32(&hits) != before {
		t.Errorf("lookup right after a failure should be skipped: ok=%v", ok)
	}
}
