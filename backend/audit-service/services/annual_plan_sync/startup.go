package annual_plan_sync

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"audit-service/pkg/logger"

	"gorm.io/gorm"
)

// reconcileLockKey is the Postgres advisory lock that serialises reconciles
// within one database (advisory locks are per database, so each tenant's
// database has its own). The value is "AnnualSn" in ASCII.
const reconcileLockKey int64 = 0x416e6e75616c536e

// ErrReconcileBusy means another instance holds the reconcile lock.
var ErrReconcileBusy = errors.New("annual_plan_sync: another instance is reconciling")

// TryReconcileLock takes the transaction-scoped reconcile lock on Postgres
// (released at commit / rollback). It reports false when another instance
// holds it. Other dialects (the sqlite tests) have no advisory locks and
// always get true; the upserts and the optimistic plan writes are safe to run
// concurrently anyway, the lock only avoids duplicate work and lock waits.
func TryReconcileLock(tx *gorm.DB) (bool, error) {
	if tx.Dialector == nil || tx.Dialector.Name() != "postgres" {
		return true, nil
	}
	var got bool
	if err := tx.Raw("SELECT pg_try_advisory_xact_lock(?)", reconcileLockKey).Scan(&got).Error; err != nil {
		return false, err
	}
	return got, nil
}

// ReconcileExclusive runs Reconcile in its own transaction under the reconcile
// lock. When another instance holds the lock it writes nothing and returns
// ErrReconcileBusy: that instance is doing the same work.
func ReconcileExclusive(ctx context.Context, db *gorm.DB, st SourceState) (ReconcileResult, error) {
	var res ReconcileResult
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		got, err := TryReconcileLock(tx)
		if err != nil {
			return fmt.Errorf("take reconcile lock: %w", err)
		}
		if !got {
			return ErrReconcileBusy
		}
		res, err = Reconcile(ctx, tx, st)
		return err
	})
	return res, err
}

// LoadFunc reads the sources. It is called once per attempt, so each attempt
// can present a freshly signed token.
type LoadFunc func(ctx context.Context) (SourceState, error)

// StartupOptions shape the start-up reconcile's retries.
type StartupOptions struct {
	// InitialDelay lets master-service and risk-service finish starting.
	InitialDelay time.Duration
	// Attempts is the total number of tries (at least 1).
	Attempts int
	// Backoff is the wait after the first failure; it doubles up to MaxBackoff.
	Backoff    time.Duration
	MaxBackoff time.Duration

	// sleep waits d or until ctx ends (tests replace it).
	sleep func(ctx context.Context, d time.Duration) error
}

// DefaultStartupOptions: first try 20s after boot, then 15s, 30s, 1m, 2m, 4m
// (about 8 minutes in all) before giving up until the next start.
func DefaultStartupOptions() StartupOptions {
	return StartupOptions{
		InitialDelay: 20 * time.Second,
		Attempts:     6,
		Backoff:      15 * time.Second,
		MaxBackoff:   5 * time.Minute,
	}
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// RunStartupReconcile refreshes the lookup from the sources and fills empty
// plan snapshots (ReconcileExclusive), retrying with backoff. It blocks, so
// call it in a goroutine; it never panics (a panic in an attempt is recovered,
// logged and counted as a failed attempt) and only logs failures. It returns
// the last error for tests; nil also when another instance did the work.
func RunStartupReconcile(ctx context.Context, db *gorm.DB, load LoadFunc, opt StartupOptions) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("annual snapshot reconcile panicked: %v", r)
			logger.Error("Annual snapshot startup reconcile panicked", logger.LogField("panic", fmt.Sprint(r)))
		}
	}()

	sleep := opt.sleep
	if sleep == nil {
		sleep = sleepCtx
	}
	attempts := opt.Attempts
	if attempts < 1 {
		attempts = 1
	}
	if err := sleep(ctx, opt.InitialDelay); err != nil {
		return err
	}

	backoff := opt.Backoff
	for attempt := 1; attempt <= attempts; attempt++ {
		var res ReconcileResult
		res, err = startupAttempt(ctx, db, load)
		switch {
		case err == nil:
			logger.Info("Annual snapshot startup reconcile completed",
				logger.LogField("attempt", attempt),
				logger.LogField("refs_departments_upserted", res.Refs.Departments),
				logger.LogField("refs_entities_upserted", res.Refs.Entities),
				logger.LogField("refs_year_scores_upserted", res.Refs.YearScores),
				logger.LogField("refs_source_records_skipped", res.Refs.Skipped),
				logger.LogField("plans_changed", res.Backfill.PlansChanged),
				logger.LogField("fields_filled", res.Backfill.FieldsFilled))
			return nil
		case errors.Is(err, ErrReconcileBusy):
			logger.Info("Annual snapshot startup reconcile skipped: another instance is running it")
			return nil
		case ctx.Err() != nil:
			return ctx.Err()
		}
		if attempt == attempts {
			break
		}
		logger.Warn("Annual snapshot startup reconcile failed, retrying",
			logger.LogField("attempt", attempt),
			logger.LogField("retry_in", backoff.String()),
			logger.LogField("error", err.Error()))
		if serr := sleep(ctx, backoff); serr != nil {
			return serr
		}
		backoff *= 2
		if opt.MaxBackoff > 0 && backoff > opt.MaxBackoff {
			backoff = opt.MaxBackoff
		}
	}
	logger.Error("Annual snapshot startup reconcile gave up; the service keeps running. "+
		"It runs again on the next start, or run: ./audit migrate backfill-annual-snapshots --from-source",
		logger.LogField("attempts", attempts),
		logger.LogField("error", err.Error()))
	return err
}

// startupAttempt is one load + reconcile. A panic becomes an error so the
// next attempt still runs.
func startupAttempt(ctx context.Context, db *gorm.DB, load LoadFunc) (res ReconcileResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
			logger.Error("Annual snapshot startup reconcile attempt panicked",
				logger.LogField("panic", fmt.Sprint(r)),
				logger.LogField("stack", string(debug.Stack())))
		}
	}()
	st, err := load(ctx)
	if err != nil {
		return ReconcileResult{}, err
	}
	return ReconcileExclusive(ctx, db, st)
}
