package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AssessmentStatus string
type TestingStatus string

const (
	AssessmentStatusPlanned    AssessmentStatus = "PLANNED"
	AssessmentStatusInProgress  AssessmentStatus = "IN_PROGRESS"
	AssessmentStatusCompleted   AssessmentStatus = "COMPLETED"
	AssessmentStatusDeferred    AssessmentStatus = "DEFERRED"

	TestingStatusNotTested     TestingStatus = "NOT_TESTED"
	TestingStatusPassed        TestingStatus = "PASSED"
	TestingStatusFailed        TestingStatus = "FAILED"
	TestingStatusPartial       TestingStatus = "PARTIAL"
	TestingStatusNotApplicable TestingStatus = "NOT_APPLICABLE"
)

type ControlAssessment struct {
	ID              uuid.UUID             `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AssessmentCode  string                `gorm:"type:varchar(50);uniqueIndex;not null" json:"assessment_code"`

	// Link to control and audit scope
	ControlID       uuid.UUID             `gorm:"type:uuid;not null;index" json:"control_id"`
	Control         Control               `gorm:"foreignKey:ControlID" json:"control"`

	AuditPlanID     uuid.UUID             `gorm:"type:uuid;not null;index" json:"audit_plan_id"`
	AuditPlan       AnnualAuditPlan       `gorm:"foreignKey:AuditPlanID" json:"audit_plan"`

	AuditScopeID    *uuid.UUID            `gorm:"type:uuid;index" json:"audit_scope_id,omitempty"`
	AuditScope      *AuditScope           `gorm:"foreignKey:AuditScopeID" json:"audit_scope,omitempty"`

	// Assessment details
	AssessmentStatus AssessmentStatus      `gorm:"type:varchar(20);default:'PLANNED'" json:"assessment_status"`
	TestingStatus   TestingStatus         `gorm:"type:varchar(20);default:'NOT_TESTED'" json:"testing_status"`
	Effectiveness   ControlEffectiveness  `gorm:"type:varchar(30)" json:"effectiveness"`

	// Testing information
	TestMethod      string                `gorm:"type:varchar(100)" json:"test_method"`      // inquiry, observation, inspection, re-performance
	TestPeriodStart time.Time             `json:"test_period_start"`
	TestPeriodEnd   time.Time             `json:"test_period_end"`
	SampleSize      int                   `gorm:"type:int" json:"sample_size"`
	ExceptionsFound int                   `gorm:"type:int;default:0" json:"exceptions_found"`

	// Findings
	Finding          string                `gorm:"type:text" json:"finding"`
	RootCause        string                `gorm:"type:text" json:"root_cause"`
	Recommendation   string                `gorm:"type:text" json:"recommendation"`

	// Auditor
	AuditorID       uuid.UUID             `gorm:"type:uuid;index" json:"auditor_id"`
	Auditor         Employee              `gorm:"foreignKey:AuditorID" json:"auditor"`

	AssessmentDate  *time.Time            `json:"assessment_date,omitempty"`
	CompletedAt     *time.Time            `json:"completed_at,omitempty"`

	IsActive        bool                  `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
	DeletedAt       gorm.DeletedAt        `gorm:"index" json:"-"`
}
