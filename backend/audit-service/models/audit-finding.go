package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FindingSeverity string

const (
	SeverityLow      FindingSeverity = "LOW"
	SeverityMedium   FindingSeverity = "MEDIUM"
	SeverityHigh     FindingSeverity = "HIGH"
	SeverityCritical FindingSeverity = "CRITICAL"
)

type AuditFinding struct {
	ID               uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ActivityPlanID   uuid.UUID       `gorm:"type:uuid;not null;index" json:"activity_plan_id"`
	
	// Optional: A finding is usually derived from a specific working paper
	WorkingPaperID   *uuid.UUID      `gorm:"type:uuid;index" json:"working_paper_id"`
	WorkingPaper     *WorkingPaper   `gorm:"foreignKey:WorkingPaperID" json:"working_paper,omitempty"`

	FindingCode      string          `gorm:"type:varchar(50);uniqueIndex;not null" json:"finding_code"` // e.g., "FND-2024-001"
	Title            string          `gorm:"type:varchar(255);not null" json:"title"`
	
	// The 5 C's of internal audit findings
	Condition        string          `gorm:"type:text;not null" json:"condition"`     // What is?
	Criteria         string          `gorm:"type:text;not null" json:"criteria"`      // What should be?
	Cause            string          `gorm:"type:text;not null" json:"cause"`         // Why did it happen?
	Effect           string          `gorm:"type:text;not null" json:"effect"`        // What is the impact?
	Recommendation   string          `gorm:"type:text;not null" json:"recommendation"`// How to fix it?

	ManagementAction string          `gorm:"type:text" json:"management_action"`      // How the business unit responded
	TargetDate       *time.Time      `json:"target_date"`                             // When management promises to fix it

	Severity         FindingSeverity `gorm:"type:varchar(20);not null" json:"severity"`
	Status           string          `gorm:"type:varchar(50);default:'OPEN'" json:"status"` // OPEN, RESOLVED, CLOSED, OVERDUE

	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	DeletedAt        gorm.DeletedAt  `gorm:"index" json:"-"`
}