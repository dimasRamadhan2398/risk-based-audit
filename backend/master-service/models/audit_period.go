package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditPeriodStatus string

const (
	AuditPeriodStatusDraft     AuditPeriodStatus = "DRAFT"
	AuditPeriodStatusActive    AuditPeriodStatus = "ACTIVE"
	AuditPeriodStatusClosed    AuditPeriodStatus = "CLOSED"
)

type AuditPeriod struct {
	ID              uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	PeriodCode      string              `gorm:"type:varchar(50);uniqueIndex;not null" json:"period_code"` // e.g., "2024", "Q1-2024"
	PeriodName      string              `gorm:"type:varchar(100);not null" json:"period_name"` // e.g., "Annual 2024", "Q1 2024"
	Year            int                 `gorm:"type:int;not null;index" json:"year"`
	Quarter         *int                `gorm:"type:int" json:"quarter,omitempty"` // 1, 2, 3, 4 (nullable for annual periods)
	StartDate       time.Time           `gorm:"not null" json:"start_date"`
	EndDate         time.Time           `gorm:"not null" json:"end_date"`
	Status          AuditPeriodStatus   `gorm:"type:varchar(20);default:'DRAFT'" json:"status"`

	// Company context
	CompanyID       uuid.UUID           `gorm:"type:uuid;index" json:"company_id"`
	Company         Company             `gorm:"foreignKey:CompanyID" json:"company"`

	IsActive        bool                `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	DeletedAt       gorm.DeletedAt      `gorm:"index" json:"-"`
}
