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
	RiskLevelQ1    string    `gorm:"type:varchar(50)" json:"risk_level_q1"`
	RiskLevelQ2    string    `gorm:"type:varchar(50)" json:"risk_level_q2"`
	RiskLevelQ3    string    `gorm:"type:varchar(50)" json:"risk_level_q3"`
	RiskLevelQ4    string    `gorm:"type:varchar(50)" json:"risk_level_q4"`

	// How the risk read in this year's CRP. risk_register/risk_profile hold only
	// the latest version, so an edit made while viewing one year is written
	// here and must not rewrite how the risk read in another year. SnapshotAt is
	// nil only on rows written before these columns existed; `risk up`
	// backfills them from the register/profile.
	RiskEvent    string     `gorm:"type:text" json:"risk_event"`
	Category     string     `gorm:"type:varchar(100)" json:"category"`
	Description  string     `gorm:"type:text" json:"description"`
	LocationID   *uuid.UUID `gorm:"type:uuid" json:"location_id,omitempty"`
	LocationName string     `gorm:"type:varchar(100)" json:"location_name"`
	SnapshotAt   *time.Time `json:"snapshot_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (RiskAssessment) TableName() string {
	return "risk_assessment"
}
