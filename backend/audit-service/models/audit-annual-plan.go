package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditAnnual struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Year        int            `gorm:"type:int;not null;uniqueIndex" json:"year"`
	Title       string         `gorm:"type:varchar(255);not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Status      string         `gorm:"type:varchar(50);default:'DRAFT'" json:"status"` // e.g., DRAFT, PENDING_APPROVAL, APPROVED, CLOSED

	// Activity Plans linked to this Annual Plan
	ActivityPlans []ActivityPlan `gorm:"foreignKey:AnnualPlanID" json:"activity_plans,omitempty"`

	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}