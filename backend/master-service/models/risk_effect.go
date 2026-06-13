package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RiskEffectCategory string

const (
	RiskEffectCategoryFinancial       RiskEffectCategory = "FINANCIAL"
	RiskEffectCategoryOperational    RiskEffectCategory = "OPERATIONAL"
	RiskEffectCategoryCompliance      RiskEffectCategory = "COMPLIANCE"
	RiskEffectCategoryReputation      RiskEffectCategory = "REPUTATION"
	RiskEffectCategoryStrategic       RiskEffectCategory = "STRATEGIC"
	RiskEffectCategoryInformation     RiskEffectCategory = "INFORMATION"
)

type RiskEffect struct {
	ID              uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EffectCode     string              `gorm:"type:varchar(50);uniqueIndex;not null" json:"effect_code"`
	EffectName     string              `gorm:"type:varchar(200);not null" json:"effect_name"`
	Description    string              `gorm:"type:text" json:"description"`
	Category       RiskEffectCategory  `gorm:"type:varchar(30);not null" json:"category"`

	// Link to risk register
	RiskRegisterID uuid.UUID           `gorm:"type:uuid;index" json:"risk_register_id"`
	RiskRegister   RiskRegister       `gorm:"foreignKey:RiskRegisterID" json:"risk_register"`

	IsActive       bool                `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
	DeletedAt      gorm.DeletedAt      `gorm:"index" json:"-"`
}
