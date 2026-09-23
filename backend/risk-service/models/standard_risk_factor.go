package models

import (
	"time"

	"github.com/google/uuid"
)

type StandardRiskFactor struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name            string    `gorm:"type:varchar(200);not null" json:"name"`
	Description     string    `gorm:"type:text" json:"description"`
	ScoreGuidelines string    `gorm:"type:text" json:"score_guidelines"` // Store JSON string of scores 1-5 characteristics
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (StandardRiskFactor) TableName() string {
	return "standard_risk_factors"
}
