package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditCompletionSnapshot records a point-in-time snapshot of the audit completion
// analysis result. These rows accumulate over time and are used for monthly trend charts,
// replacing the hardcoded baseline arrays in the performance controller.
type AuditCompletionSnapshot struct {
	ID   uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Year int       `gorm:"type:int;not null;index"                         json:"year"`
	// Month is 1-12; 0 means "whole year" (used for the latest live snapshot)
	Month int `gorm:"type:int;not null" json:"month"`

	// SnapshotAt is the exact time the analyzer ran and computed these numbers
	SnapshotAt time.Time `gorm:"type:timestamptz;not null;index" json:"snapshot_at"`

	// Status bucket counts from audit_activities for the scoped year
	TotalActivities      int `gorm:"type:int" json:"total_activities"`
	CompletedActivities  int `gorm:"type:int" json:"completed_activities"`
	InProgressActivities int `gorm:"type:int" json:"in_progress_activities"`
	ReportingActivities  int `gorm:"type:int" json:"reporting_activities"`
	PlannedActivities    int `gorm:"type:int" json:"planned_activities"`
	CancelledActivities  int `gorm:"type:int" json:"cancelled_activities"`

	// Operational counts from audit_executions (the finer-grained tracking layer)
	TotalExecutions     int `gorm:"type:int" json:"total_executions"`
	CompletedExecutions int `gorm:"type:int" json:"completed_executions"`

	// PlanBasedRate = CompletedActivities / (TotalActivities - CancelledActivities) * 100
	// This is the headline KPI that matches IIA "work plan realization" definition.
	PlanBasedRate float64 `gorm:"type:decimal(5,2)" json:"plan_based_rate"`

	// OperationalRate = CompletedExecutions / TotalExecutions * 100
	// Answers: "of audits that started, how many actually finished?"
	OperationalRate float64 `gorm:"type:decimal(5,2)" json:"operational_rate"`

	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AuditCompletionSnapshot) TableName() string {
	return "audit_completion_snapshots"
}

// AuditCompletionResult is the cached/computed result stored in Redis.
// It is a superset of what the dashboard card needs, plus drill-down data.
type AuditCompletionResult struct {
	Year int `json:"year"`

	// Headline KPI — use this for the dashboard card
	PlanBasedRate   float64 `json:"plan_based_rate"`
	OperationalRate float64 `json:"operational_rate"`

	// Breakdown counters
	TotalActivities      int `json:"total_activities"`
	CompletedActivities  int `json:"completed_activities"`
	InProgressActivities int `json:"in_progress_activities"`
	ReportingActivities  int `json:"reporting_activities"`
	PlannedActivities    int `json:"planned_activities"`
	CancelledActivities  int `json:"cancelled_activities"`
	TotalExecutions      int `json:"total_executions"`
	CompletedExecutions  int `json:"completed_executions"`

	// Metadata
	ComputedAt    time.Time `json:"computed_at"`
	CacheSource   string    `json:"cache_source"`   // "redis_cache" | "live_query"
	NextRefreshIn string    `json:"next_refresh_in"` // human-readable TTL hint
}
