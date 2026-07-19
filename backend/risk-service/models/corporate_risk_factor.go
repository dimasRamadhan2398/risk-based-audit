package models

import (
	"time"

	"github.com/google/uuid"
)

type CorporateRiskFactor struct {
	ID                   uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	StandardRiskFactorID uuid.UUID           `gorm:"type:uuid;not null;index" json:"standard_risk_factor_id"`
	StandardRiskFactor   *StandardRiskFactor `gorm:"foreignKey:StandardRiskFactorID" json:"standard_risk_factor,omitempty"`
	Weight               float64             `gorm:"type:decimal(5,4);not null;default:0" json:"weight"` // weight in percent, e.g., 0.15 for 15%
	CreatedAt            time.Time           `json:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at"`
}

func (CorporateRiskFactor) TableName() string {
	return "corporate_risk_factors"
}
