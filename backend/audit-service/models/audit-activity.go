package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ActivityPlan struct {
	ID           uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AnnualPlanID uuid.UUID   `gorm:"type:uuid;not null;index" json:"annual_plan_id"`
	AnnualPlan   AuditAnnual `gorm:"foreignKey:AnnualPlanID" json:"annual_plan"`

	// Link to Master Data: Which unit is being audited?
	TargetUnitID uuid.UUID `gorm:"type:uuid;not null;index" json:"target_unit_id"`

	ProjectCode string `gorm:"type:varchar(50);uniqueIndex;not null" json:"project_code"`
	Title       string `gorm:"type:varchar(255);not null" json:"title"`
	Objective   string `gorm:"type:text" json:"objective"`
	Scope       string `gorm:"type:text" json:"scope"`

	PlannedStart time.Time `json:"planned_start"`
	PlannedEnd   time.Time `json:"planned_end"`
	Status       string    `gorm:"type:varchar(50);default:'PLANNED'" json:"status"` // PLANNED, IN_PROGRESS, REPORTING, COMPLETED

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
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
