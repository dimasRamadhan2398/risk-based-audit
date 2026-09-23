// Package audit_completion provides a continuous background analyzer for the
// Audit Completion Rate KPI. It runs on a configurable ticker interval,
// queries the database for the current status distribution of all audit
// activities and executions, computes two distinct completion rates, caches
// the result in Redis for near-instant dashboard reads, and persists a
// point-in-time snapshot row for historical trend charts.
//
// # Formula — Plan-Based Rate (headline KPI)
//
//	PlanBasedRate = COMPLETED / (TOTAL - CANCELLED) × 100
//
// This answers: "Of everything we committed to this year, how much did we
// deliver?" It is the standard IIA work-plan realization metric.
//
// # Formula — Operational Rate (secondary / drill-down)
//
//	OperationalRate = completed_executions / total_executions × 100
//
// This answers: "Of audits that actually started, how many finished?" Useful
// for the operations team but intentionally NOT used as the headline KPI
// because it hides audits that were planned but never started.
package audit_completion

import (
	"audit-service/models"
	"audit-service/pkg/logger"
	"audit-service/pkg/redis"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	redisgo "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// CacheKeyFmt is the Redis key pattern for the completion result.
// Formatted as: audit:completion_rate:<year>
const CacheKeyFmt = "audit:completion_rate:%d"

// CacheTTL is how long the Redis result lives before the next analyzer run
// overwrites it. The analyzer fires every AnalyzerInterval, so TTL should be
// slightly longer to avoid a brief cache-miss window between runs.
const CacheTTL = 6 * time.Minute

// AnalyzerInterval controls how often the background goroutine re-computes.
const AnalyzerInterval = 5 * time.Minute

// statusActivity maps the status strings used in audit_activities.
var (
	statusCompleted  = []string{"COMPLETED", "Completed", "completed"}
	statusInProgress = []string{"IN_PROGRESS", "In Progress", "in_progress"}
	statusReporting  = []string{"REPORTING", "Reporting", "reporting"}
	statusCancelled  = []string{"CANCELLED", "Cancelled", "cancelled"}
)

// AuditCompletionAnalyzer is the main background component.
// Create it with NewAuditCompletionAnalyzer and launch Start in a goroutine.
type AuditCompletionAnalyzer struct {
	db          *gorm.DB
	redisClient *redis.Client // may be nil when Redis is not configured
	interval    time.Duration
}

// NewAuditCompletionAnalyzer creates a new analyzer. redisClient may be nil
// (the analyzer will still persist snapshots but skip caching).
func NewAuditCompletionAnalyzer(db *gorm.DB, redisClient *redis.Client) *AuditCompletionAnalyzer {
	return &AuditCompletionAnalyzer{
		db:          db,
		redisClient: redisClient,
		interval:    AnalyzerInterval,
	}
}

// WithInterval overrides the default 5-minute tick (useful for testing).
func (a *AuditCompletionAnalyzer) WithInterval(d time.Duration) *AuditCompletionAnalyzer {
	a.interval = d
	return a
}

// Start runs the analyzer loop. It fires once immediately, then on every tick.
// Cancel the context to stop cleanly.
func (a *AuditCompletionAnalyzer) Start(ctx context.Context) {
	logger.Info("[AuditCompletionAnalyzer] Starting — interval: " + a.interval.String())

	// Run immediately on startup so the cache is warm before the first request.
	a.runAnalysis(ctx)

	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("[AuditCompletionAnalyzer] Stopping — context cancelled")
			return
		case <-ticker.C:
			a.runAnalysis(ctx)
		}
	}
}

// runAnalysis performs the full analysis cycle for the current year.
func (a *AuditCompletionAnalyzer) runAnalysis(ctx context.Context) {
	year := time.Now().Year()
	result, err := a.compute(ctx, year)
	if err != nil {
		logger.Error("[AuditCompletionAnalyzer] Compute failed",
			logger.LogField("error", err),
			logger.LogField("year", year),
		)
		return
	}

	// 1. Cache result in Redis
	if err := a.cacheResult(ctx, year, result); err != nil {
		logger.Warn("[AuditCompletionAnalyzer] Redis cache write failed — continuing",
			logger.LogField("error", err),
		)
	}

	// 2. Persist snapshot for historical trending
	if err := a.persistSnapshot(result); err != nil {
		logger.Warn("[AuditCompletionAnalyzer] Snapshot persist failed — continuing",
			logger.LogField("error", err),
		)
	}

	logger.Info("[AuditCompletionAnalyzer] Analysis complete",
		logger.LogField("year", year),
		logger.LogField("plan_based_rate", fmt.Sprintf("%.1f%%", result.PlanBasedRate)),
		logger.LogField("operational_rate", fmt.Sprintf("%.1f%%", result.OperationalRate)),
		logger.LogField("total_activities", result.TotalActivities),
		logger.LogField("completed_activities", result.CompletedActivities),
	)
}

// compute queries the DB and returns a populated AuditCompletionResult.
func (a *AuditCompletionAnalyzer) compute(ctx context.Context, year int) (*models.AuditCompletionResult, error) {
	result := &models.AuditCompletionResult{
		Year:        year,
		ComputedAt:  time.Now().UTC(),
		CacheSource: "live_query",
	}

	// 0. Synchronize AuditExecution status to AuditActivity status for the year
	if err := SyncAllExecutionsForYear(ctx, a.db, year); err != nil {
		logger.Warn("[AuditCompletionAnalyzer] Pre-compute execution sync warning", logger.LogField("error", err))
	}

	// ----------------------------------------------------------------
	// Query 1: audit_activities — plan-based rate
	// Join through audit_annuals to scope by year.
	// ----------------------------------------------------------------
	type statusCount struct {
		Status string
		Count  int
	}
	var statusCounts []statusCount

	err := a.db.WithContext(ctx).
		Model(&models.AuditActivity{}).
		Select("audit_activities.status as status, COUNT(*) as count").
		Joins("JOIN audit_annuals ON audit_annuals.id = audit_activities.annual_plan_id").
		Where("audit_annuals.year = ? AND audit_annuals.deleted_at IS NULL AND audit_activities.deleted_at IS NULL", year).
		Group("audit_activities.status").
		Scan(&statusCounts).Error
	if err != nil {
		return nil, fmt.Errorf("query audit_activities by status: %w", err)
	}

	for _, sc := range statusCounts {
		result.TotalActivities += sc.Count
		switch {
		case containsCI(statusCompleted, sc.Status):
			result.CompletedActivities += sc.Count
		case containsCI(statusInProgress, sc.Status):
			result.InProgressActivities += sc.Count
		case containsCI(statusReporting, sc.Status):
			result.ReportingActivities += sc.Count
		case containsCI(statusCancelled, sc.Status):
			result.CancelledActivities += sc.Count
		default:
			// Treat everything else (PLANNED, DRAFT, etc.) as planned
			result.PlannedActivities += sc.Count
		}
	}

	// Plan-Based Rate: exclude CANCELLED from denominator
	denominator := result.TotalActivities - result.CancelledActivities
	if denominator > 0 {
		result.PlanBasedRate = math.Round(
			float64(result.CompletedActivities)/float64(denominator)*1000,
		) / 10
	}

	startOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endOfYear := time.Date(year, 12, 31, 23, 59, 59, 999999999, time.UTC)

	var totalExec int64
	a.db.WithContext(ctx).Model(&models.AuditExecution{}).
		Where("created_at >= ? AND created_at <= ?", startOfYear, endOfYear).
		Count(&totalExec)
	result.TotalExecutions = int(totalExec)

	var completedExec int64
	a.db.WithContext(ctx).
		Model(&models.AuditExecution{}).
		Where("created_at >= ? AND created_at <= ? AND (status IN ? OR progress = ?)",
			startOfYear, endOfYear, []string{"Completed", "COMPLETED", "completed"}, 100,
		).
		Count(&completedExec)
	result.CompletedExecutions = int(completedExec)

	if result.TotalExecutions > 0 {
		result.OperationalRate = math.Round(
			float64(result.CompletedExecutions)/float64(result.TotalExecutions)*1000,
		) / 10
	}

	return result, nil
}

func (a *AuditCompletionAnalyzer) cacheResult(ctx context.Context, year int, result *models.AuditCompletionResult) error {
	if a.redisClient == nil {
		return nil
	}

	key := fmt.Sprintf(CacheKeyFmt, year)
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}

	return a.redisClient.Set(ctx, key, data, CacheTTL).Err()
}

// persistSnapshot writes a snapshot row to the database.
func (a *AuditCompletionAnalyzer) persistSnapshot(result *models.AuditCompletionResult) error {
	snap := &models.AuditCompletionSnapshot{
		Year:                 result.Year,
		Month:                int(result.ComputedAt.Month()),
		SnapshotAt:           result.ComputedAt,
		TotalActivities:      result.TotalActivities,
		CompletedActivities:  result.CompletedActivities,
		InProgressActivities: result.InProgressActivities,
		ReportingActivities:  result.ReportingActivities,
		PlannedActivities:    result.PlannedActivities,
		CancelledActivities:  result.CancelledActivities,
		TotalExecutions:      result.TotalExecutions,
		CompletedExecutions:  result.CompletedExecutions,
		PlanBasedRate:        result.PlanBasedRate,
		OperationalRate:      result.OperationalRate,
	}
	return a.db.Create(snap).Error
}

// GetCachedResult retrieves the cached result from Redis for the given year.
// Returns (nil, nil) when Redis is not configured. Returns ErrCacheMiss when
// the key is absent so the caller can fall back to a live query.
var ErrCacheMiss = fmt.Errorf("cache miss")

// GetCachedResult attempts to read the current year's completion result from Redis.
func GetCachedResult(ctx context.Context, redisClient *redis.Client, year int) (*models.AuditCompletionResult, error) {
	if redisClient == nil {
		return nil, ErrCacheMiss
	}

	key := fmt.Sprintf(CacheKeyFmt, year)
	data, err := redisClient.Get(ctx, key).Bytes()
	if err != nil {
		if err == redisgo.Nil {
			return nil, ErrCacheMiss
		}
		return nil, fmt.Errorf("redis get %s: %w", key, err)
	}

	var result models.AuditCompletionResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal cached result: %w", err)
	}

	result.CacheSource = "redis_cache"
	ttl, _ := redisClient.TTL(ctx, key).Result()
	result.NextRefreshIn = ttl.Round(time.Second).String()

	return &result, nil
}

// ComputePublic is the exported wrapper around compute, allowing the controller
// to trigger an on-demand analysis on a cache miss without starting the full
// background goroutine.
func (a *AuditCompletionAnalyzer) ComputePublic(ctx context.Context, year int) (*models.AuditCompletionResult, error) {
	return a.compute(ctx, year)
}

// containsCI checks if a string is in a slice (case-insensitive membership).
func containsCI(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
