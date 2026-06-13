package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IssueStatus string
type IssuePriority string
type IssueCategory string

const (
	IssueStatusOpen           IssueStatus = "OPEN"
	IssueStatusAccepted       IssueStatus = "ACCEPTED"
	IssueStatusInRemediation  IssueStatus = "IN_REMEDIATION"
	IssueStatusRemediated     IssueStatus = "REMEDIATED"
	IssueStatusValidated      IssueStatus = "VALIDATED"
	IssueStatusClosed         IssueStatus = "CLOSED"
	IssueStatusRiskAccepted   IssueStatus = "RISK_ACCEPTED"

	IssuePriorityCritical     IssuePriority = "CRITICAL"
	IssuePriorityHigh         IssuePriority = "HIGH"
	IssuePriorityMedium       IssuePriority = "MEDIUM"
	IssuePriorityLow          IssuePriority = "LOW"

	IssueCategoryControlDeficiency IssueCategory = "CONTROL_DEFICIENCY"
	IssueCategoryProcessGap        IssueCategory = "PROCESS_GAP"
	IssueCategoryComplianceGap     IssueCategory = "COMPLIANCE_GAP"
	IssueCategoryPolicyViolation    IssueCategory = "POLICY_VIOLATION"
	IssueCategorySystemIssue        IssueCategory = "SYSTEM_ISSUE"
	IssueCategoryOperationalRisk    IssueCategory = "OPERATIONAL_RISK"
)

type AuditIssue struct {
	ID              uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	IssueCode       string              `gorm:"type:varchar(50);uniqueIndex;not null" json:"issue_code"`

	// Link to finding
	FindingID       *uuid.UUID          `gorm:"type:uuid;index" json:"finding_id,omitempty"`
	Finding         *AuditFinding       `gorm:"foreignKey:FindingID" json:"finding,omitempty"`

	AuditPlanID     uuid.UUID           `gorm:"type:uuid;not null;index" json:"audit_plan_id"`
	AuditPlan       AnnualAuditPlan     `gorm:"foreignKey:AuditPlanID" json:"audit_plan"`

	// Classification
	Category        IssueCategory      `gorm:"type:varchar(30);not null" json:"category"`
	Priority        IssuePriority      `gorm:"type:varchar(20);not null" json:"priority"`

	// Issue details
	Title           string              `gorm:"type:varchar(200);not null" json:"title"`
	Description     string              `gorm:"type:text;not null" json:"description"`
	CurrentState    string              `gorm:"type:text" json:"current_state"`
	ExpectedState   string              `gorm:"type:text" json:"expected_state"`
	Impact          string              `gorm:"type:text" json:"impact"`

	// Risk linkage
	RiskRegisterID  *uuid.UUID          `gorm:"type:uuid;index" json:"risk_register_id,omitempty"`
	RiskRegister    *RiskRegister       `gorm:"foreignKey:RiskRegisterID" json:"risk_register,omitempty"`

	ControlID       *uuid.UUID          `gorm:"type:uuid;index" json:"control_id,omitempty"`
	Control         *Control            `gorm:"foreignKey:ControlID" json:"control,omitempty"`

	// Department & owner
	DepartmentID    uuid.UUID           `gorm:"type:uuid;index" json:"department_id"`
	Department      Department          `gorm:"foreignKey:DepartmentID" json:"department"`

	OwnerID         uuid.UUID           `gorm:"type:uuid;index" json:"owner_id"`
	Owner           Employee            `gorm:"foreignKey:OwnerID" json:"owner"`

	// Auditor
	AuditorID       uuid.UUID           `gorm:"type:uuid;index" json:"auditor_id"`
	Auditor         Employee            `gorm:"foreignKey:AuditorID" json:"auditor"`

	// Status tracking
	Status          IssueStatus         `gorm:"type:varchar(25);default:'OPEN'" json:"status"`
	RootCause       string              `gorm:"type:text" json:"root_cause"`

	// Dates
	IssueDate       time.Time           `json:"issue_date"`
	DueDate         *time.Time          `json:"due_date,omitempty"`
	RemediatedDate  *time.Time          `json:"remediated_date,omitempty"`
	ValidatedDate   *time.Time          `json:"validated_date,omitempty"`
	ClosedAt        *time.Time          `json:"closed_at,omitempty"`

	// Remediation
	RemediationPlan string              `gorm:"type:text" json:"remediation_plan"`
	RemediationEvidence string         `gorm:"type:text" json:"remediation_evidence,omitempty"`

	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	DeletedAt       gorm.DeletedAt      `gorm:"index" json:"-"`
}
