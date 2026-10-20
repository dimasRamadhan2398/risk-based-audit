package services

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeStore struct {
	data    map[string][]byte
	failAll bool
	sets    int
}

func newFakeStore() *fakeStore { return &fakeStore{data: map[string][]byte{}} }

func (f *fakeStore) Get(_ context.Context, key string) ([]byte, error) {
	if f.failAll {
		return nil, errors.New("redis down")
	}
	b, ok := f.data[key]
	if !ok {
		return nil, errCacheMiss
	}
	return b, nil
}

func (f *fakeStore) Set(_ context.Context, key string, val []byte, _ time.Duration) error {
	if f.failAll {
		return errors.New("redis down")
	}
	f.sets++
	f.data[key] = val
	return nil
}

func (f *fakeStore) SetNX(_ context.Context, key string, _ time.Duration) (bool, error) {
	if f.failAll {
		return false, errors.New("redis down")
	}
	if _, ok := f.data[key]; ok {
		return false, nil
	}
	f.data[key] = []byte("1")
	return true, nil
}

func countingLoader(body string, err error) (func() ([]byte, error), *int) {
	calls := 0
	return func() ([]byte, error) {
		calls++
		return []byte(body), err
	}, &calls
}

func TestGetOrLoadCachesSuccessfulJSON(t *testing.T) {
	c := &AnalyticsCache{store: newFakeStore(), batchTTL: time.Minute}
	load, calls := countingLoader(`{"data":[1,2]}`, nil)

	for i := 0; i < 3; i++ {
		b, err := c.GetOrLoad("risk-score", load)
		if err != nil || string(b) != `{"data":[1,2]}` {
			t.Fatalf("call %d: got %q, %v", i, b, err)
		}
	}
	if *calls != 1 {
		t.Fatalf("python-ai called %d times, want 1", *calls)
	}
}

func TestGetOrLoadKeepsEndpointsApart(t *testing.T) {
	c := &AnalyticsCache{store: newFakeStore(), batchTTL: time.Minute}
	riskLoad, _ := countingLoader(`{"k":"risk"}`, nil)
	anomLoad, _ := countingLoader(`{"k":"anomaly"}`, nil)

	r, _ := c.GetOrLoad("risk-score", riskLoad)
	a, _ := c.GetOrLoad("anomaly", anomLoad)
	if string(r) == string(a) {
		t.Fatalf("endpoints share a cache entry: %q", r)
	}
}

func TestGetOrLoadDoesNotCacheFailuresOrInvalidJSON(t *testing.T) {
	store := newFakeStore()
	c := &AnalyticsCache{store: store, batchTTL: time.Minute}

	failing, failCalls := countingLoader("", errors.New("python-ai down"))
	if _, err := c.GetOrLoad("anomaly", failing); err == nil {
		t.Fatal("want the loader error")
	}
	_, _ = c.GetOrLoad("anomaly", failing)
	if *failCalls != 2 {
		t.Fatalf("failed load called %d times, want 2 (not cached)", *failCalls)
	}

	garbage, garbageCalls := countingLoader("<html>502</html>", nil)
	_, _ = c.GetOrLoad("text-analysis", garbage)
	_, _ = c.GetOrLoad("text-analysis", garbage)
	if *garbageCalls != 2 || store.sets != 0 {
		t.Fatalf("invalid JSON cached: calls=%d sets=%d", *garbageCalls, store.sets)
	}
}

func TestGetOrLoadFallsBackWhenRedisFails(t *testing.T) {
	c := &AnalyticsCache{store: &fakeStore{data: map[string][]byte{}, failAll: true}, batchTTL: time.Minute}
	load, calls := countingLoader(`{"ok":true}`, nil)

	b, err := c.GetOrLoad("performance-trend", load)
	if err != nil || string(b) != `{"ok":true}` || *calls != 1 {
		t.Fatalf("got %q, %v, calls=%d", b, err, *calls)
	}
}

func TestGetOrLoadWithoutRedisAlwaysLoads(t *testing.T) {
	var nilCache *AnalyticsCache
	for _, c := range []*AnalyticsCache{nilCache, {batchTTL: time.Minute}} {
		load, calls := countingLoader(`{}`, nil)
		_, _ = c.GetOrLoad("risk-score", load)
		_, _ = c.GetOrLoad("risk-score", load)
		if *calls != 2 {
			t.Fatalf("calls=%d, want 2", *calls)
		}
	}
}

func TestAllowRetrainThrottles(t *testing.T) {
	c := &AnalyticsCache{store: newFakeStore(), retrainInterval: time.Hour}
	if !c.AllowRetrain() {
		t.Fatal("first retrain should be allowed")
	}
	if c.AllowRetrain() {
		t.Fatal("second retrain inside the window should be skipped")
	}
}

func TestAllowRetrainFailsOpen(t *testing.T) {
	down := &AnalyticsCache{store: &fakeStore{failAll: true}, retrainInterval: time.Hour}
	var disabled *AnalyticsCache
	if !down.AllowRetrain() || !disabled.AllowRetrain() {
		t.Fatal("without a working Redis every retrain must be allowed (old behaviour)")
	}
}

func TestTriggerAutoRetrainSkipsWithoutCallingPythonAI(t *testing.T) {
	store := newFakeStore()
	store.data[cacheKeyPrefix+"retrain:auto:last"] = []byte("1")
	// baseURL points nowhere: reaching python-ai would return an error.
	s := &AnalyticsService{
		aiClient: &PythonAIClient{baseURL: "http://127.0.0.1:1", client: NewPythonAIClient().client},
		cache:    &AnalyticsCache{store: store, retrainInterval: time.Hour},
	}
	res, err := s.TriggerAutoRetrain()
	if err != nil {
		t.Fatalf("throttled retrain should not call python-ai: %v", err)
	}
	if m, ok := res.(map[string]string); !ok || m["status"] != "skipped" {
		t.Fatalf("got %#v, want status=skipped", res)
	}
}
