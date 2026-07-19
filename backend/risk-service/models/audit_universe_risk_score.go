package models

import (
	"time"

	"github.com/google/uuid"
)

type AuditUniverseRiskScore struct {
	ID                    uuid.UUID            `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AuditUniverseYearID   uuid.UUID            `gorm:"type:uuid;not null;index" json:"audit_universe_year_id"`
	CorporateRiskFactorID uuid.UUID            `gorm:"type:uuid;not null;index" json:"corporate_risk_factor_id"`
	CorporateRiskFactor   *CorporateRiskFactor `gorm:"foreignKey:CorporateRiskFactorID" json:"corporate_risk_factor,omitempty"`
	Score                 int                  `gorm:"not null;default:1" json:"score"` // 1 to 5
	WeightedScore         float64              `gorm:"type:decimal(5,4);not null;default:0" json:"weighted_score"` // weight * score
	CreatedAt             time.Time            `json:"created_at"`
	UpdatedAt             time.Time            `json:"updated_at"`
}

func (AuditUniverseRiskScore) TableName() string {
	return "audit_universe_risk_scores"
}
