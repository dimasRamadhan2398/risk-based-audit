package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RiskAppetiteStatus string

const (
	RiskAppetiteStatusDraft     RiskAppetiteStatus = "DRAFT"
	RiskAppetiteStatusSubmitted RiskAppetiteStatus = "SUBMITTED"
	RiskAppetiteStatusApproved  RiskAppetiteStatus = "APPROVED"
)

type RiskAppetite struct {
	ID             uuid.UUID          `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Statement      string             `gorm:"type:text;not null" json:"statement"`
	ThresholdLimit float64            `gorm:"type:decimal(15,2);not null" json:"threshold_limit"`
	Status         RiskAppetiteStatus `gorm:"type:varchar(20);default:'DRAFT'" json:"status"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
	DeletedAt      gorm.DeletedAt     `gorm:"index" json:"-"`
}

func (RiskAppetite) TableName() string {
	return "risk_appetite"
}
