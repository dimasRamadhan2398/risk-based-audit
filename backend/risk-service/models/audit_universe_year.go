package models

import (
	"time"

	"github.com/google/uuid"
)

type AuditUniverseYear struct {
	ID                       uuid.UUID               `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CorporateAuditUniverseID uuid.UUID               `gorm:"type:uuid;not null;index" json:"corporate_audit_universe_id"`
	CorporateAuditUniverse   *CorporateAuditUniverse `gorm:"foreignKey:CorporateAuditUniverseID" json:"corporate_audit_universe,omitempty"`
	Year                     int                     `gorm:"not null" json:"year"` // e.g. 2026
	RiskIndex                float64                 `gorm:"type:decimal(5,2);default:0" json:"risk_index"` // e.g. 83.6
	RiskLevel                string                  `gorm:"type:varchar(50);default:''" json:"risk_level"` // High, Medium to High, Medium, Low to Medium, Low
	AuditPriority            bool                    `gorm:"default:false" json:"audit_priority"` // true if High or Medium to High
	RiskScores               []AuditUniverseRiskScore `gorm:"foreignKey:AuditUniverseYearID;constraint:OnDelete:CASCADE" json:"risk_scores,omitempty"`
	CreatedAt                time.Time               `json:"created_at"`
	UpdatedAt                time.Time               `json:"updated_at"`
}

func (AuditUniverseYear) TableName() string {
	return "audit_universe_years"
}
