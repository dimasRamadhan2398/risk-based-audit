package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RecommendationPriority string
type RecommendationStatus string
type RecommendationResponse string

const (
	RecommendationPriorityImmediate RecommendationPriority = "IMMEDIATE"
	RecommendationPriorityShortTerm RecommendationPriority = "SHORT_TERM"
	RecommendationPriorityMediumTerm RecommendationPriority = "MEDIUM_TERM"
	RecommendationPriorityLongTerm  RecommendationPriority = "LONG_TERM"

	RecommendationStatusOpen       RecommendationStatus = "OPEN"
	RecommendationStatusInProgress RecommendationStatus = "IN_PROGRESS"
	RecommendationStatusCompleted RecommendationStatus = "COMPLETED"
	RecommendationStatusRejected  RecommendationStatus = "REJECTED"
	RecommendationStatusClosed    RecommendationStatus = "CLOSED"

	RecommendationResponseAgree    RecommendationResponse = "AGREE"
	RecommendationResponsePartial RecommendationResponse = "PARTIAL"
	RecommendationResponseDisagree RecommendationResponse = "DISAGREE"
)

type AuditRecommendation struct {
	ID              uuid.UUID                `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RecommendationCode string                 `gorm:"type:varchar(50);uniqueIndex;not null" json:"recommendation_code"`

	// Link to finding
	FindingID       uuid.UUID                `gorm:"type:uuid;not null;index" json:"finding_id"`
	Finding         AuditFinding             `gorm:"foreignKey:FindingID" json:"finding"`

	// Recommendation details
	Recommendation  string                   `gorm:"type:text;not null" json:"recommendation"`
	Priority        RecommendationPriority   `gorm:"type:varchar(20);not null" json:"priority"`
	Benefits        string                   `gorm:"type:text" json:"benefits"`          // benefit of implementing

	// Implementation
	ManagementResponse RecommendationResponse `gorm:"type:varchar(20)" json:"management_response"`
	ManagementComment string                 `gorm:"type:text" json:"management_comment"`
	ActionPlan       string                  `gorm:"type:text" json:"action_plan"`        // planned action

	// Responsible person
	ResponsibleID   uuid.UUID                `gorm:"type:uuid;index" json:"responsible_id"`
	Responsible     Employee                 `gorm:"foreignKey:ResponsibleID" json:"responsible"`

	// Dates
	DueDate         *time.Time               `json:"due_date,omitempty"`
	CompletedDate   *time.Time               `json:"completed_date,omitempty"`
	ClosedAt        *time.Time               `json:"closed_at,omitempty"`

	// Status tracking
	Status          RecommendationStatus     `gorm:"type:varchar(20);default:'OPEN'" json:"status"`
	Progress        int                      `gorm:"type:int;default:0" json:"progress"` // 0-100

	// Verification
	VerifiedByID    *uuid.UUID               `gorm:"type:uuid;index" json:"verified_by_id,omitempty"`
	VerifiedBy      *Employee                `gorm:"foreignKey:VerifiedByID" json:"verified_by,omitempty"`
	VerifiedAt      *time.Time               `json:"verified_at,omitempty"`
	VerificationNote string                   `gorm:"type:text" json:"verification_note,omitempty"`

	CostEstimate    *float64                 `gorm:"type:decimal(15,2)" json:"cost_estimate,omitempty"`
	ActualCost      *float64                 `gorm:"type:decimal(15,2)" json:"actual_cost,omitempty"`

	CreatedAt       time.Time                `json:"created_at"`
	UpdatedAt       time.Time                `json:"updated_at"`
	DeletedAt       gorm.DeletedAt           `gorm:"index" json:"-"`
}
