package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AssignmentStatus string

const (
	AssignmentStatusPending   AssignmentStatus = "pending"
	AssignmentStatusAssigned  AssignmentStatus = "assigned"
	AssignmentStatusAccepted  AssignmentStatus = "accepted"
	AssignmentStatusCompleted AssignmentStatus = "completed"
	AssignmentStatusRejected  AssignmentStatus = "rejected"
)

type AuditAssignment struct {
	ID              uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AssignmentTitle  string           `gorm:"type:varchar(200);not null" json:"assignment_title"`
	Description     string           `gorm:"type:text" json:"description"`
	AuditorID       uuid.UUID        `gorm:"type:uuid;not null" json:"auditor_id"`
	AuditPlanID     uuid.UUID        `gorm:"type:uuid;not null" json:"audit_plan_id"`
	Status          AssignmentStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	StartDate       *time.Time       `json:"start_date,omitempty"`
	EndDate         *time.Time       `json:"end_date,omitempty"`
	Notes           string           `gorm:"type:text" json:"notes"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `gorm:"index" json:"-"`
}

// Request DTOs
type CreateAuditAssignmentRequest struct {
	AssignmentTitle string     `json:"assignment_title" binding:"required" validate:"required,max=200"`
	Description    string     `json:"description"`
	AuditorID      uuid.UUID  `json:"auditor_id" binding:"required"`
	AuditPlanID    uuid.UUID  `json:"audit_plan_id" binding:"required"`
	StartDate      *time.Time `json:"start_date"`
	EndDate        *time.Time `json:"end_date"`
	Notes          string     `json:"notes"`
}

type UpdateAuditAssignmentRequest struct {
	AssignmentTitle *string    `json:"assignment_title" validate:"omitempty,max=200"`
	Description     *string    `json:"description"`
	StartDate       *time.Time `json:"start_date"`
	EndDate         *time.Time `json:"end_date"`
	Notes           *string    `json:"notes"`
}

type UpdateAssignmentStatusRequest struct {
	Status AssignmentStatus `json:"status" binding:"required"`
}

type ListAuditAssignmentsRequest struct {
	Page       int        `form:"page" validate:"min=1"`
	PageSize   int        `form:"page_size" validate:"min=1,max=100"`
	Search     string     `form:"search"`
	AuditorID  *uuid.UUID `form:"auditor_id"`
	AuditPlanID *uuid.UUID `form:"audit_plan_id"`
	Status     *string    `form:"status"`
}

// Response DTOs
type AuditAssignmentResponse struct {
	ID              string `json:"id"`
	AssignmentTitle  string `json:"assignment_title"`
	Description     string `json:"description"`
	AuditorID       string `json:"auditor_id"`
	AuditPlanID     string `json:"audit_plan_id"`
	Status          string `json:"status"`
	StartDate       *string `json:"start_date,omitempty"`
	EndDate         *string `json:"end_date,omitempty"`
	Notes           string `json:"notes"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}