package models

import (
	"time"

	"github.com/google/uuid"
)

type RiskLevel struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RiskCode        string    `gorm:"type:varchar(20);uniqueIndex;not null" json:"code"` // "LOW", "LOW_MODERATE", "MODERATE", "MODERATE_HIGH", "HIGH"
	RiskName        string    `gorm:"type:varchar(50);not null" json:"name"`
	RiskDescription string    `gorm:"type:text" json:"description"`
	MinScore        int       `gorm:"not null" json:"min_score"`
	MaxScore        int       `gorm:"not null" json:"max_score"`
	Color           string    `gorm:"type:varchar(20)" json:"color"` // UI hint: "green", "yellow", "red"
	IsActive        bool      `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (RiskLevel) TableName() string {
	return "risk_level"
}
