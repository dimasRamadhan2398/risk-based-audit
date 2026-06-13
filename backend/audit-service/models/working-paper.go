package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WorkingPaper struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ActivityPlanID uuid.UUID      `gorm:"type:uuid;not null;index" json:"activity_plan_id"`
	
	// Link to the Assignment to know exactly who prepared it
	PreparedByID   uuid.UUID      `gorm:"type:uuid;not null" json:"prepared_by_id"`
	ReviewedByID   *uuid.UUID     `gorm:"type:uuid" json:"reviewed_by_id"` // Nullable until reviewed

	PaperCode      string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"paper_code"` // e.g., "WP-IT-001"
	Title          string         `gorm:"type:varchar(255);not null" json:"title"`
	Methodology    string         `gorm:"type:text" json:"methodology"`
	TestResults    string         `gorm:"type:text" json:"test_results"`
	Conclusion     string         `gorm:"type:text" json:"conclusion"`

	// Link to Master Data: Which Risk/Control was tested?
	RiskID         *uuid.UUID     `gorm:"type:uuid" json:"risk_id"`
	ControlID      *uuid.UUID     `gorm:"type:uuid" json:"control_id"`

	Status         string         `gorm:"type:varchar(50);default:'DRAFT'" json:"status"` // DRAFT, SUBMITTED, REVIEWED, REJECTED

	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}