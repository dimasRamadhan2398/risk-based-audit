package models

import (
	"time"

	"github.com/google/uuid"
)

type RiskAssessment struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RiskRegisterID uuid.UUID `gorm:"type:uuid;not null;index" json:"risk_register_id"`
	Year           int       `gorm:"type:int;not null" json:"year"`
	ImpactQ1       int       `gorm:"type:int;default:0" json:"impact_q1"`
	ImpactQ2       int       `gorm:"type:int;default:0" json:"impact_q2"`
	ImpactQ3       int       `gorm:"type:int;default:0" json:"impact_q3"`
	ImpactQ4       int       `gorm:"type:int;default:0" json:"impact_q4"`
	LikelihoodQ1   int       `gorm:"type:int;default:0" json:"likelihood_q1"`
	LikelihoodQ2   int       `gorm:"type:int;default:0" json:"likelihood_q2"`
	LikelihoodQ3   int       `gorm:"type:int;default:0" json:"likelihood_q3"`
	LikelihoodQ4   int       `gorm:"type:int;default:0" json:"likelihood_q4"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (RiskAssessment) TableName() string {
	return "risk_assessment"
}
