package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RiskProfile struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DepartmentID uuid.UUID      `gorm:"type:uuid;not null;index" json:"department_id"`
	OwnerID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"owner_id"`
	Category     string         `gorm:"type:varchar(100);not null" json:"category"`
	Description  string         `gorm:"type:text" json:"description"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Risk scoring fields
	Impact         int          `gorm:"not null" json:"impact"`
	Likelihood     int          `gorm:"not null" json:"likelihood"`
	RiskScore      int          `gorm:"not null" json:"risk_score"`   // Impact × Likelihood
	SeverityWeight float64      `gorm:"type:decimal(5,4)" json:"severity_weight"`

	// Relation to RiskLevel
	RiskLevelID *uuid.UUID `gorm:"type:uuid;index" json:"risk_level_id"`
	RiskLevel   *RiskLevel `gorm:"foreignKey:RiskLevelID" json:"risk_level,omitempty"`
}

func (RiskProfile) TableName() string {
	return "risk_profile"
}
