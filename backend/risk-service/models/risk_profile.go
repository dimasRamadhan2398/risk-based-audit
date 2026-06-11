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
}

func (RiskProfile) TableName() string {
	return "risk_profile"
}
