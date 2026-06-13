package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MitigationStatus string
type MitigationProgress string

const (
	MitigationStatusPlanned     MitigationStatus = "PLANNED"
	MitigationStatusInProgress  MitigationStatus = "IN_PROGRESS"
	MitigationStatusOnHold     MitigationStatus = "ON_HOLD"
	MitigationStatusCompleted  MitigationStatus = "COMPLETED"
	MitigationStatusCancelled  MitigationStatus = "CANCELLED"
	MitigationStatusOverdue    MitigationStatus = "OVERDUE"

	MitigationProgressNotStarted MitigationProgress = "NOT_STARTED"
	MitigationProgressInitiated  MitigationProgress = "INITIATED"
	MitigationProgressInTesting MitigationProgress = "IN_TESTING"
	MitigationProgressDeployed   MitigationProgress = "DEPLOYED"
)

type MitigationAction struct {
	ID              uuid.UUID             `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ActionCode      string                `gorm:"type:varchar(50);uniqueIndex;not null" json:"action_code"`

	// Linkages
	RiskRegisterID  uuid.UUID             `gorm:"type:uuid;not null;index" json:"risk_register_id"`
	RiskRegister    RiskRegister          `gorm:"foreignKey:RiskRegisterID" json:"risk_register"`

	ControlID       *uuid.UUID            `gorm:"type:uuid;index" json:"control_id,omitempty"`
	Control         *Control              `gorm:"foreignKey:ControlID" json:"control,omitempty"`

	IssueID         *uuid.UUID            `gorm:"type:uuid;index" json:"issue_id,omitempty"`
	Issue           *AuditIssue           `gorm:"foreignKey:IssueID" json:"issue,omitempty"`

	// Action details
	ActionTitle     string                `gorm:"type:varchar(200);not null" json:"action_title"`
	Description     string                `gorm:"type:text" json:"description"`
	ActionType      string                `gorm:"type:varchar(50)" json:"action_type"` // new_control, enhance_control, process_change, etc.

	// Status & progress
	Status          MitigationStatus      `gorm:"type:varchar(20);default:'PLANNED'" json:"status"`
	Progress        MitigationProgress    `gorm:"type:varchar(20);default:'NOT_STARTED'" json:"progress"`
	CompletionPercent int                 `gorm:"type:int;default:0" json:"completion_percent"`

	// Owner
	OwnerID         uuid.UUID             `gorm:"type:uuid;index" json:"owner_id"`
	Owner           Employee              `gorm:"foreignKey:OwnerID" json:"owner"`

	DepartmentID    uuid.UUID             `gorm:"type:uuid;index" json:"department_id"`
	Department      Department            `gorm:"foreignKey:DepartmentID" json:"department"`

	// Resource
	BudgetEstimate  *float64              `gorm:"type:decimal(15,2)" json:"budget_estimate,omitempty"`
	ActualCost      *float64              `gorm:"type:decimal(15,2)" json:"actual_cost,omitempty"`

	// Dates
	PlannedStartDate time.Time            `json:"planned_start_date"`
	PlannedEndDate   time.Time            `json:"planned_end_date"`
	ActualStartDate *time.Time           `json:"actual_start_date,omitempty"`
	ActualEndDate   *time.Time           `json:"actual_end_date,omitempty"`

	// Verification
	VerifiedByID    *uuid.UUID            `gorm:"type:uuid;index" json:"verified_by_id,omitempty"`
	VerifiedBy      *Employee             `gorm:"foreignKey:VerifiedByID" json:"verified_by,omitempty"`
	VerifiedAt      *time.Time           `json:"verified_at,omitempty"`
	VerificationResult string             `gorm:"type:text" json:"verification_result,omitempty"`

	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
	DeletedAt       gorm.DeletedAt        `gorm:"index" json:"-"`
}
