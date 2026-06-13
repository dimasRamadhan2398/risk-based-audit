package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkpaperStatus string
type WorkpaperType string

const (
	WorkpaperStatusDraft       WorkpaperStatus = "DRAFT"
	WorkpaperStatusInReview    WorkpaperStatus = "IN_REVIEW"
	WorkpaperStatusApproved    WorkpaperStatus = "APPROVED"
	WorkpaperStatusArchived    WorkpaperStatus = "ARCHIVED"

	WorkpaperTypePlanning      WorkpaperType = "PLANNING"
	WorkpaperTypeRiskAssessment WorkpaperType = "RISK_ASSESSMENT"
	WorkpaperTypeControlTesting WorkpaperType = "CONTROL_TESTING"
	WorkpaperTypeSubstantive   WorkpaperType = "SUBSTANTIVE"
	WorkpaperTypeSampling      WorkpaperType = "SAMPLING"
	WorkpaperTypeInterview     WorkpaperType = "INTERVIEW"
	WorkpaperTypeObservation   WorkpaperType = "OBSERVATION"
	WorkpaperTypeInspection    WorkpaperType = "INSPECTION"
	WorkpaperTypeConclusion    WorkpaperType = "CONCLUSION"
)

type AuditWorkpaper struct {
	ID              uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	WorkpaperCode   string              `gorm:"type:varchar(50);uniqueIndex;not null" json:"workpaper_code"`
	Title           string              `gorm:"type:varchar(200);not null" json:"title"`
	WorkpaperType   WorkpaperType       `gorm:"type:varchar(30);not null" json:"workpaper_type"`
	Description     string              `gorm:"type:text" json:"description"`

	// Link to audit
	AuditPlanID     uuid.UUID           `gorm:"type:uuid;not null;index" json:"audit_plan_id"`
	AuditPlan       AnnualAuditPlan     `gorm:"foreignKey:AuditPlanID" json:"audit_plan"`

	AuditScopeID    *uuid.UUID          `gorm:"type:uuid;index" json:"audit_scope_id,omitempty"`
	AuditScope      *AuditScope         `gorm:"foreignKey:AuditScopeID" json:"audit_scope,omitempty"`

	ControlAssessmentID *uuid.UUID      `gorm:"type:uuid;index" json:"control_assessment_id,omitempty"`
	ControlAssessment   *ControlAssessment `gorm:"foreignKey:ControlAssessmentID" json:"control_assessment,omitempty"`

	AuditFindingID  *uuid.UUID          `gorm:"type:uuid;index" json:"audit_finding_id,omitempty"`
	AuditFinding     *AuditFinding       `gorm:"foreignKey:AuditFindingID" json:"audit_finding,omitempty"`

	// Content
	Content         string              `gorm:"type:text" json:"content"`
	AttachmentURL   string              `gorm:"type:varchar(500)" json:"attachment_url"`

	// Status & versioning
	Status          WorkpaperStatus    `gorm:"type:varchar(20);default:'DRAFT'" json:"status"`
	Version         int                `gorm:"type:int;default:1" json:"version"`

	// Auditor info
	AuditorID       uuid.UUID           `gorm:"type:uuid;index" json:"auditor_id"`
	Auditor         Employee            `gorm:"foreignKey:AuditorID" json:"auditor"`

	// Reviewer
	ReviewerID      *uuid.UUID          `gorm:"type:uuid;index" json:"reviewer_id,omitempty"`
	Reviewer        *Employee           `gorm:"foreignKey:ReviewerID" json:"reviewer,omitempty"`
	ReviewedAt      *time.Time          `json:"reviewed_at,omitempty"`
	ReviewNotes     string              `gorm:"type:text" json:"review_notes,omitempty"`

	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	DeletedAt       gorm.DeletedAt      `gorm:"index" json:"-"`
}
