package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FindingSeverity string
type FindingStatus string
type FindingType string

const (
	FindingSeverityCritical FindingSeverity = "CRITICAL"
	FindingSeverityHigh     FindingSeverity = "HIGH"
	FindingSeverityMedium   FindingSeverity = "MEDIUM"
	FindingSeverityLow      FindingSeverity = "LOW"
	FindingSeverityInfo     FindingSeverity = "INFO"

	FindingStatusOpen       FindingStatus = "OPEN"
	FindingStatusInProgress FindingStatus = "IN_PROGRESS"
	FindingStatusVerified   FindingStatus = "VERIFIED"
	FindingStatusClosed     FindingStatus = "CLOSED"
	FindingStatusValidated  FindingStatus = "VALIDATED"

	FindingTypeDeficiency  FindingType = "DEFICIENCY"
	FindingTypeObservations FindingType = "OBSERVATION"
	FindingTypeOpportunity FindingType = "OPPORTUNITY"
	FindingTypeNonCompliance FindingType = "NON_COMPLIANCE"
	FindingTypeControlFailure FindingType = "CONTROL_FAILURE"
)

type AuditFinding struct {
	ID              uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FindingCode     string              `gorm:"type:varchar(50);uniqueIndex;not null" json:"finding_code"` // "FND-2024-001"

	// Links
	AuditPlanID     uuid.UUID           `gorm:"type:uuid;not null;index" json:"audit_plan_id"`
	AuditPlan       AnnualAuditPlan     `gorm:"foreignKey:AuditPlanID" json:"audit_plan"`

	AuditScopeID    *uuid.UUID          `gorm:"type:uuid;index" json:"audit_scope_id,omitempty"`
	AuditScope      *AuditScope         `gorm:"foreignKey:AuditScopeID" json:"audit_scope,omitempty"`

	ControlAssessmentID *uuid.UUID      `gorm:"type:uuid;index" json:"control_assessment_id,omitempty"`
	ControlAssessment   *ControlAssessment `gorm:"foreignKey:ControlAssessmentID" json:"control_assessment,omitempty"`

	// Classification
	FindingType     FindingType         `gorm:"type:varchar(30);not null" json:"finding_type"`
	Severity        FindingSeverity     `gorm:"type:varchar(20);not null" json:"severity"`

	// Finding details
	Title           string              `gorm:"type:varchar(200);not null" json:"title"`
	Description     string              `gorm:"type:text;not null" json:"description"`
	Facts           string              `gorm:"type:text" json:"facts"`
	Criterion       string              `gorm:"type:text" json:"criterion"`       // what should be
	Condition       string              `gorm:"type:text" json:"condition"`        // what was found
	Cause           string              `gorm:"type:text" json:"cause"`          // why it happened
	Effect          string              `gorm:"type:text" json:"effect"`          // impact of the finding

	// Risk linkage
	RiskRegisterID  *uuid.UUID          `gorm:"type:uuid;index" json:"risk_register_id,omitempty"`
	RiskRegister    *RiskRegister       `gorm:"foreignKey:RiskRegisterID" json:"risk_register,omitempty"`

	// Department
	DepartmentID    uuid.UUID           `gorm:"type:uuid;index" json:"department_id"`
	Department      Department          `gorm:"foreignKey:DepartmentID" json:"department"`

	// Owner (responsible for remediation)
	OwnerID         uuid.UUID           `gorm:"type:uuid;index" json:"owner_id"`
	Owner           Employee            `gorm:"foreignKey:OwnerID" json:"owner"`

	// Auditor
	AuditorID       uuid.UUID           `gorm:"type:uuid;index" json:"auditor_id"`
	Auditor         Employee            `gorm:"foreignKey:AuditorID" json:"auditor"`

	// Status
	Status          FindingStatus       `gorm:"type:varchar(20);default:'OPEN'" json:"status"`
	AgreedWithManagement bool            `gorm:"default:false" json:"agreed_with_management"`

	// Dates
	IssueDate       time.Time           `json:"issue_date"`
	DueDate         *time.Time          `json:"due_date,omitempty"`
	ClosedAt        *time.Time          `json:"closed_at,omitempty"`

	// Remediation evidence
	Evidence        string              `gorm:"type:text" json:"evidence,omitempty"`

	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	DeletedAt       gorm.DeletedAt      `gorm:"index" json:"-"`
}
