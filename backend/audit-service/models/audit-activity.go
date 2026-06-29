package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlannedActivity struct {
	ID                string `json:"id"`
	AuditName         string `json:"auditName"`
	Auditee           string `json:"auditee"`
	Category          string `json:"category"`
	RiskLevel         string `json:"riskLevel"`
	Duration          int    `json:"duration"`
	Priority          string `json:"priority"`
	NumberOfAuditors  int    `json:"numberOfAuditors"`
	EstimatedSchedule string `json:"estimatedSchedule"`
	BudgetEstimation  string `json:"budgetEstimation"`
}

type ResourceAuditor struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Position     string `json:"position"`
	Competence   string `json:"competence"`
	Availability string `json:"availability"`
}

type PlanBudget struct {
	TotalEstimatedCost   float64 `json:"totalEstimatedCost"`
	TotalAllocatedBudget float64 `json:"totalAllocatedBudget"`
	BudgetNotes          string  `json:"budgetNotes"`
}

type PlanReview struct {
	CreatorName      string `json:"creatorName"`
	CreatorPosition  string `json:"creatorPosition"`
	ApproverName     string `json:"approverName"`
	ApproverPosition string `json:"approverPosition"`
	ApprovalDate     string `json:"approvalDate"`
	AdditionalNotes  string `json:"additionalNotes"`
}

type ActivityPlan struct {
	ID                uuid.UUID         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PlanTitle         string            `gorm:"type:varchar(255)" json:"planTitle"`
	PlanYear          string            `gorm:"type:varchar(50)" json:"planYear"`
	PlanPeriodStart   string            `gorm:"type:varchar(100)" json:"planPeriodStart"`
	PlanPeriodEnd     string            `gorm:"type:varchar(100)" json:"planPeriodEnd"`
	Department        string            `gorm:"type:varchar(255)" json:"department"`
	CreatedBy         string            `gorm:"type:varchar(255)" json:"createdBy"`
	CreationDate      string            `gorm:"type:varchar(100)" json:"creationDate"`
	PlannedActivities []PlannedActivity `gorm:"serializer:json" json:"plannedActivities"`
	ResourceAuditors  []ResourceAuditor `gorm:"serializer:json" json:"resourceAuditors"`
	Budget            PlanBudget        `gorm:"serializer:json" json:"budget"`
	Review            PlanReview        `gorm:"serializer:json" json:"review"`
	Status            string            `gorm:"type:varchar(50);default:'PLANNED'" json:"status"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	DeletedAt         gorm.DeletedAt    `gorm:"index" json:"-"`
}

func (ActivityPlan) TableName() string {
	return "activity_plans"
}

type AuditActivity struct {
	ID           uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AnnualPlanID uuid.UUID   `gorm:"type:uuid;not null;index" json:"annual_plan_id"`
	AnnualPlan   AuditAnnual `gorm:"foreignKey:AnnualPlanID" json:"annual_plan"`

	// Link to Master Data: Which unit is being audited?
	TargetUnitID uuid.UUID `gorm:"type:uuid;not null;index" json:"target_unit_id"`

	ProjectCode     string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"project_code"`
	Title           string    `gorm:"type:varchar(255);not null" json:"title"`
	AuditType       string    `gorm:"type:varchar(100)" json:"audit_type"` // e.g., Assurance, Special, Investigation
	AuditUniverseID uuid.UUID `gorm:"type:uuid" json:"audit_universe_id"`
	Justification   string    `gorm:"type:text" json:"justification"`
	AuditPurpose    string    `gorm:"type:text" json:"audit_purpose"`
	AuditFocus      string    `gorm:"type:text" json:"audit_focus"`
	TeamSize        int       `gorm:"type:int" json:"team_size"`
	TotalMandays    int       `gorm:"type:int" json:"total_mandays"`
	TeamLeaderID    uuid.UUID `gorm:"type:uuid" json:"team_leader_id"`

	Objective string `gorm:"type:text" json:"objective"`
	Scope     string `gorm:"type:text" json:"scope"`

	PlannedStart time.Time `json:"planned_start"`
	PlannedEnd   time.Time `json:"planned_end"`
	Status       string    `gorm:"type:varchar(50);default:'PLANNED'" json:"status"` // PLANNED, IN_PROGRESS, REPORTING, COMPLETED

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AuditActivity) TableName() string {
	return "audit_activities"
}

// Request DTOs
type CreateActivityPlanRequest struct {
	AnnualPlanID uuid.UUID `json:"annual_plan_id" binding:"required"`
	TargetUnitID uuid.UUID `json:"target_unit_id" binding:"required"`
	ProjectCode  string    `json:"project_code" binding:"required" validate:"required,max=50"`
	Title        string    `json:"title" binding:"required" validate:"required,max=255"`
	Objective    string    `json:"objective"`
	Scope        string    `json:"scope"`
	PlannedStart time.Time `json:"planned_start" binding:"required"`
	PlannedEnd   time.Time `json:"planned_end" binding:"required"`
	Status       string    `json:"status" validate:"omitempty,max=50"`
}

type UpdateActivityPlanRequest struct {
	TargetUnitID *uuid.UUID `json:"target_unit_id"`
	Title        *string    `json:"title" validate:"omitempty,max=255"`
	Objective    *string    `json:"objective"`
	Scope        *string    `json:"scope"`
	PlannedStart *time.Time `json:"planned_start"`
	PlannedEnd   *time.Time `json:"planned_end"`
	Status       *string    `json:"status" validate:"omitempty,max=50"`
}

type ListActivityPlansRequest struct {
	Page         int        `form:"page" validate:"min=1"`
	PageSize     int        `form:"page_size" validate:"min=1,max=100"`
	Search       string     `form:"search"`
	AnnualPlanID *uuid.UUID `form:"annual_plan_id"`
	TargetUnitID *uuid.UUID `form:"target_unit_id"`
	Status       *string    `form:"status"`
}

// Response DTOs
type ActivityPlanResponse struct {
	ID           string `json:"id"`
	AnnualPlanID string `json:"annual_plan_id"`
	TargetUnitID string `json:"target_unit_id"`
	ProjectCode  string `json:"project_code"`
	Title        string `json:"title"`
	Objective    string `json:"objective"`
	Scope        string `json:"scope"`
	PlannedStart string `json:"planned_start"`
	PlannedEnd   string `json:"planned_end"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}
