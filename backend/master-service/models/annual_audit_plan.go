package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditPlanStatus string
type AuditPlanPriority string

const (
	AuditPlanStatusDraft      AuditPlanStatus = "DRAFT"
	AuditPlanStatusSubmitted  AuditPlanStatus = "SUBMITTED"
	AuditPlanStatusApproved  AuditPlanStatus = "APPROVED"
	AuditPlanStatusRejected  AuditPlanStatus = "REJECTED"
	AuditPlanStatusRevised   AuditPlanStatus = "REVISED"

	AuditPlanPriorityCritical AuditPlanPriority = "CRITICAL"
	AuditPlanPriorityHigh     AuditPlanPriority = "HIGH"
	AuditPlanPriorityMedium   AuditPlanPriority = "MEDIUM"
	AuditPlanPriorityLow      AuditPlanPriority = "LOW"
)

type AnnualAuditPlan struct {
	ID              uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PlanCode        string              `gorm:"type:varchar(50);uniqueIndex;not null" json:"plan_code"` // "AAP-2024-001"
	PlanTitle       string              `gorm:"type:varchar(200);not null" json:"plan_title"`
	Description     string              `gorm:"type:text" json:"description"`

	// Period reference
	AuditPeriodID   uuid.UUID           `gorm:"type:uuid;not null;index" json:"audit_period_id"`
	AuditPeriod     AuditPeriod         `gorm:"foreignKey:AuditPeriodID" json:"audit_period"`

	// Risk-based selection
	RiskRegisterID  *uuid.UUID          `gorm:"type:uuid;index" json:"risk_register_id,omitempty"`
	RiskRegister    *RiskRegister       `gorm:"foreignKey:RiskRegisterID" json:"risk_register,omitempty"`

	// Audit scope
	DepartmentID    uuid.UUID           `gorm:"type:uuid;index" json:"department_id"`
	Department      Department         `gorm:"foreignKey:DepartmentID" json:"department"`

	// Child Activities (PKPT line items)
	Activities      []AnnualAuditPlanActivity `gorm:"foreignKey:AnnualAuditPlanID" json:"activities"`

	// Planning details
	Priority        AuditPlanPriority   `gorm:"type:varchar(20);not null" json:"priority"`
	Status          AuditPlanStatus     `gorm:"type:varchar(20);default:'DRAFT'" json:"status"`
	EstimatedDays   int                 `gorm:"type:int;default:0" json:"estimated_days"`

	PlannedStartDate time.Time          `json:"planned_start_date"`
	PlannedEndDate   time.Time          `json:"planned_end_date"`

	// Approval workflow
	ApprovedByID    *uuid.UUID          `gorm:"type:uuid;index" json:"approved_by_id,omitempty"`
	ApprovedBy      *Employee           `gorm:"foreignKey:ApprovedByID" json:"approved_by,omitempty"`
	ApprovedAt      *time.Time          `json:"approved_at,omitempty"`
	ApprovalNotes   string              `gorm:"type:text" json:"approval_notes,omitempty"`

	// Who created/requested
	RequestedByID   uuid.UUID           `gorm:"type:uuid;index" json:"requested_by_id"`
	RequestedBy     Employee            `gorm:"foreignKey:RequestedByID" json:"requested_by"`

	RevisionNumber   int                 `gorm:"type:int;default:1" json:"revision_number"`
	ParentPlanID    *uuid.UUID          `gorm:"type:uuid;index" json:"parent_plan_id,omitempty"` // for revised plans

	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	DeletedAt       gorm.DeletedAt      `gorm:"index" json:"-"`
}
