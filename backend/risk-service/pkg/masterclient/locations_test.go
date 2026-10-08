package masterclient

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

const locationsBody = `{"success":true,"message":"Locations fetched successfully","data":[
 {"id":"7d1000c0-0b66-468b-978d-09e32c8d9df6","name":"Head Office","is_active":true,"city":"Jakarta Selatan"},
 {"id":"4c875bfc-3cfb-4066-ada2-757af0ef46a2","name":"Bali Branch","is_active":false},
 {"id":"00000000-0000-0000-0000-000000000000","name":"broken row"},
 {"id":"3430b58b-2377-4992-a32d-48914a7eb470","name":"  "}
]}`

type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time          { return c.t }
func (c *fakeClock) advance(d time.Duration) { c.t = c.t.Add(d) }

func newTestClient(url string) (*Client, *fakeClock) {
	c := NewClient(url, time.Second, time.Minute)
	clock := &fakeClock{t: time.Date(2026, 10, 8, 9, 0, 0, 0, time.UTC)}
	c.now = clock.now
	return c, clock
}

func TestListLocationsParsesMasterResponse(t *testing.T) {
	var gotAuth, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth, gotPath = r.Header.Get("Authorization"), r.URL.Path
		_, _ = w.Write([]byte(locationsBody))
	}))
	defer srv.Close()

	c, _ := newTestClient(srv.URL + "/")
	ctx := WithAuthorization(context.Background(), "Bearer abc")
	locs, err := c.ListLocations(ctx)
	if err != nil {
		t.Fatalf("ListLocations: %v", err)
	}
	if gotPath != "/api/v1/locations" {
		t.Errorf("path = %q", gotPath)
	}
	if gotAuth != "Bearer abc" {
		t.Errorf("Authorization not forwarded: %q", gotAuth)
	}
	// Rows with a nil ID or blank name are dropped; inactive rows are kept.
	if len(locs) != 2 || locs[0].Name != "Head Office" || locs[1].Name != "Bali Branch" || locs[1].IsActive {
		t.Errorf("locs = %+v", locs)
	}
}

func TestListLocationsCachesWithinTTL(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		_, _ = w.Write([]byte(locationsBody))
	}))
	defer srv.Close()

	c, clock := newTestClient(srv.URL)
	for i := 0; i < 3; i++ {
		if _, err := c.ListLocations(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if hits != 1 {
		t.Errorf("hits = %d within TTL, want 1", hits)
	}
	clock.advance(61 * time.Second)
	if _, err := c.ListLocations(context.Background()); err != nil {
		t.Fatal(err)
	}
	if hits != 2 {
		t.Errorf("hits = %d after TTL, want 2", hits)
	}
}

func TestListLocationsUnreachableWithoutCache(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close() // connection refused from now on

	c, _ := newTestClient(url)
	locs, err := c.ListLocations(context.Background())
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
	if locs != nil {
		t.Errorf("locs = %+v, want nil", locs)
	}
}

func TestListLocationsNon200IsUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"success":false}`, http.StatusUnauthorized)
	}))
	defer srv.Close()

	c, _ := newTestClient(srv.URL)
	if _, err := c.ListLocations(context.Background()); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable", err)
	}
}

func TestListLocationsServesStaleThenGivesUp(t *testing.T) {
	var fail atomic.Bool
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		if fail.Load() {
			http.Error(w, "down", http.StatusBadGateway)
			return
		}
		_, _ = w.Write([]byte(locationsBody))
	}))
	defer srv.Close()

	c, clock := newTestClient(srv.URL)
	if _, err := c.ListLocations(context.Background()); err != nil {
		t.Fatal(err)
	}

	fail.Store(true)
	clock.advance(2 * time.Minute) // past TTL, within the stale window
	locs, err := c.ListLocations(context.Background())
	if err != nil || len(locs) != 2 {
		t.Fatalf("stale serve: locs=%d err=%v", len(locs), err)
	}

	// Within retryAfter of the failure, master-service is not hit again.
	before := atomic.LoadInt32(&hits)
	clock.advance(time.Second)
	if _, err := c.ListLocations(context.Background()); err != nil {
		t.Fatal(err)
	}
	if atomic.LoadInt32(&hits) != before {
		t.Error("retried master-service inside the back-off window")
	}

	clock.advance(staleFor) // last good copy is now too old
	if _, err := c.ListLocations(context.Background()); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("err = %v, want ErrUnavailable once the cache is too stale", err)
	}

	fail.Store(false)
	clock.advance(retryAfter)
	if _, err := c.ListLocations(context.Background()); err != nil {
		t.Fatalf("recovery: %v", err)
	}
}
