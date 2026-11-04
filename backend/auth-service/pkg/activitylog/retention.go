package activitylog

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Purger deletes activity rows that occurred before cutoff.
type Purger interface {
	DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error)
}

// Retention runs the daily purge of rows older than the retention window.
type Retention struct {
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// StartRetention purges once shortly after start and then every interval.
// A failed purge is logged and retried at the next tick.
func StartRetention(p Purger, days int, interval time.Duration, log *zap.Logger) *Retention {
	if log == nil {
		log = zap.NewNop()
	}
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	ctx, cancel := context.WithCancel(context.Background())
	r := &Retention{cancel: cancel}
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		// Let the service finish booting before the first purge.
		if !sleep(ctx, time.Minute) {
			return
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			PurgeOnce(ctx, p, days, time.Now(), log)
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
	return r
}

// PurgeOnce deletes rows older than now minus days.
func PurgeOnce(ctx context.Context, p Purger, days int, now time.Time, log *zap.Logger) (int64, error) {
	if log == nil {
		log = zap.NewNop()
	}
	cutoff := now.UTC().AddDate(0, 0, -days)
	pctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	n, err := p.DeleteOlderThan(pctx, cutoff)
	if err != nil {
		log.Error("activity log retention: purge failed", zap.Time("cutoff", cutoff), zap.Error(err))
		return 0, err
	}
	log.Info("activity log retention: purged", zap.Int64("rows", n), zap.Time("cutoff", cutoff), zap.Int("days", days))
	return n, nil
}

// Stop ends the purge loop.
func (r *Retention) Stop() {
	if r == nil {
		return
	}
	r.cancel()
	r.wg.Wait()
}
