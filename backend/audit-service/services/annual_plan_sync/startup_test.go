package annual_plan_sync

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"audit-service/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// recordSleeps replaces the real wait: it records each duration and returns
// at once (or the context's error).
func recordSleeps(opt StartupOptions, slept *[]time.Duration) StartupOptions {
	var mu sync.Mutex
	opt.sleep = func(ctx context.Context, d time.Duration) error {
		mu.Lock()
		*slept = append(*slept, d)
		mu.Unlock()
		return ctx.Err()
	}
	return opt
}

var testOpts = StartupOptions{InitialDelay: 20 * time.Second, Attempts: 5, Backoff: 10 * time.Second, MaxBackoff: 30 * time.Second}

func TestStartupReconcileFillsPlans(t *testing.T) {
	db := testDB(t)
	plan := seedPlan(t, db, 2026, models.AnnualAuditActivity{DepartmentID: "dep-1", AuditUniverseID: "uni-1"})
	load := func(context.Context) (SourceState, error) {
		return SourceState{
			Departments: []DepartmentData{{ID: "dep-1", DepartmentName: "Finance"}},
			Entities:    []EntityData{{ID: "uni-1", Name: "Treasury"}},
			YearScores:  []YearScoreData{{CorporateAuditUniverseID: "uni-1", Year: 2026, RiskLevel: "High"}},
		}, nil
	}
	var slept []time.Duration
	if err := RunStartupReconcile(context.Background(), db, load, recordSleeps(testOpts, &slept)); err != nil {
		t.Fatal(err)
	}
	if len(slept) != 1 || slept[0] != testOpts.InitialDelay {
		t.Fatalf("waits = %v, want only the initial delay", slept)
	}
	row := reload(t, db, plan.ID.String()).Activities[0]
	if row.Department != "Finance" || row.RiskName != "Treasury" || row.RiskLevel != "High" {
		t.Fatalf("row not filled: %+v", row)
	}
}

// Sources down for the whole window: every attempt fails, the waits back off
// (doubling, capped), the error is returned, nothing panics, nothing is written.
func TestStartupReconcileGivesUpWithBackoffWhenSourcesFail(t *testing.T) {
	db := testDB(t)
	plan := seedPlan(t, db, 2026, models.AnnualAuditActivity{DepartmentID: "dep-1"})
	boom := errors.New("dial tcp: connection refused")
	calls := 0
	load := func(context.Context) (SourceState, error) { calls++; return SourceState{}, boom }

	var slept []time.Duration
	err := RunStartupReconcile(context.Background(), db, load, recordSleeps(testOpts, &slept))
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want the source error", err)
	}
	if calls != testOpts.Attempts {
		t.Fatalf("attempts = %d, want %d", calls, testOpts.Attempts)
	}
	want := []time.Duration{20 * time.Second, 10 * time.Second, 20 * time.Second, 30 * time.Second, 30 * time.Second}
	if len(slept) != len(want) {
		t.Fatalf("waits = %v, want %v", slept, want)
	}
	for i := range want {
		if slept[i] != want[i] {
			t.Fatalf("waits = %v, want %v", slept, want)
		}
	}
	if got := reload(t, db, plan.ID.String()).Activities[0].Department; got != "" {
		t.Fatalf("failed run wrote Department = %q", got)
	}
}

// A panicking source is recovered per attempt; the next attempt still runs.
func TestStartupReconcileRecoversFromPanicAndRetries(t *testing.T) {
	db := testDB(t)
	plan := seedPlan(t, db, 2026, models.AnnualAuditActivity{DepartmentID: "dep-1"})
	calls := 0
	load := func(context.Context) (SourceState, error) {
		calls++
		if calls == 1 {
			var m map[string]int
			m["boom"] = 1 // nil map write: a real panic
		}
		return SourceState{Departments: []DepartmentData{{ID: "dep-1", DepartmentName: "Finance"}}}, nil
	}
	var slept []time.Duration
	if err := RunStartupReconcile(context.Background(), db, load, recordSleeps(testOpts, &slept)); err != nil {
		t.Fatalf("err = %v, want success on the second attempt", err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
	if got := reload(t, db, plan.ID.String()).Activities[0].Department; got != "Finance" {
		t.Fatalf("Department = %q", got)
	}
}

// The caller runs it in a goroutine; it must also stop promptly when its
// context ends, even in the initial delay (real timer here).
func TestStartupReconcileDoesNotBlockAndStopsOnCancel(t *testing.T) {
	db := testDB(t)
	called := false
	load := func(context.Context) (SourceState, error) { called = true; return SourceState{}, nil }
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	start := time.Now()
	go func() {
		done <- RunStartupReconcile(ctx, db, load, StartupOptions{InitialDelay: time.Hour, Attempts: 3, Backoff: time.Hour})
	}()
	if time.Since(start) > time.Second {
		t.Fatal("starting the reconcile blocked the caller")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("err = %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("reconcile did not stop after cancel")
	}
	if called {
		t.Fatal("sources were read before the initial delay ended")
	}
}

// Two instances at once: both succeed (sqlite has no advisory lock, so this
// exercises the unlocked path: upserts and optimistic plan writes converge),
// the refs are not duplicated and the plan is filled once.
func TestStartupReconcileTwoInstancesConverge(t *testing.T) {
	db := testDB(t)
	plan := seedPlan(t, db, 2026, models.AnnualAuditActivity{DepartmentID: "dep-1"})
	load := func(context.Context) (SourceState, error) {
		return SourceState{Departments: []DepartmentData{{ID: "dep-1", DepartmentName: "Finance"}}}, nil
	}
	opt := StartupOptions{Attempts: 3}
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() { errs <- RunStartupReconcile(context.Background(), db, load, opt) }()
	}
	for i := 0; i < 2; i++ {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}
	counts, err := RefCounts(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if counts[models.SnapshotRefDepartment] != 1 {
		t.Fatalf("department refs = %d, want 1", counts[models.SnapshotRefDepartment])
	}
	if got := reload(t, db, plan.ID.String()).Activities[0].Department; got != "Finance" {
		t.Fatalf("Department = %q", got)
	}
}

// On Postgres the lock is a non-blocking, transaction-scoped advisory lock.
func TestTryReconcileLockSQLOnPostgres(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=127.0.0.1 dbname=none"}), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var sqls []string
	_ = db.Callback().Row().After("gorm:row").Register("test:capture_row", func(tx *gorm.DB) {
		sqls = append(sqls, tx.Statement.SQL.String())
	})
	// DryRun cannot execute a row query, so it reports ErrDryRunModeUnsupported
	// after building the statement; the statement is what is checked here.
	got, err := TryReconcileLock(db.Session(&gorm.Session{Logger: db.Logger.LogMode(1)}))
	if err != nil && !errors.Is(err, gorm.ErrDryRunModeUnsupported) {
		t.Fatal(err)
	}
	if got {
		t.Fatal("dry run reported the lock as taken")
	}
	if len(sqls) != 1 || !strings.Contains(sqls[0], "pg_try_advisory_xact_lock($1)") {
		t.Fatalf("sql = %v", sqls)
	}
}

func TestTryReconcileLockOtherDialectsAlwaysTrue(t *testing.T) {
	got, err := TryReconcileLock(testDB(t))
	if err != nil || !got {
		t.Fatalf("sqlite: got %v, err %v", got, err)
	}
}
