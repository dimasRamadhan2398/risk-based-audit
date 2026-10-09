package models

import (
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RiskControlMatrix struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RiskID             *uuid.UUID     `gorm:"type:uuid;index" json:"risk_id,omitempty"`
	RiskCode           string         `gorm:"type:varchar(50);not null" json:"risk_code"`
	RiskEvent          string         `gorm:"type:text;not null" json:"risk_event"`
	ControlCode        string         `gorm:"type:varchar(50);not null" json:"control_code"`
	ControlDescription string         `gorm:"type:text;not null" json:"control_description"`
	ControlType        string         `gorm:"type:varchar(30);not null" json:"control_type"` // Preventive, Detective, Corrective
	ControlOwner       string         `gorm:"type:varchar(100);not null" json:"control_owner"`
	Department         string         `gorm:"type:varchar(100);not null" json:"department"`
	Year               int            `gorm:"type:int;not null" json:"year"`

	// COSO 2013 5 Dimensions (Weight % & Rating 1-5)
	DesignEffectivenessWeight    float64 `gorm:"type:decimal(5,2);default:20.0" json:"design_effectiveness_weight"`
	DesignEffectivenessRating    int     `gorm:"type:int;default:3" json:"design_effectiveness_rating"`

	OperatingEffectivenessWeight float64 `gorm:"type:decimal(5,2);default:20.0" json:"operating_effectiveness_weight"`
	OperatingEffectivenessRating int     `gorm:"type:int;default:3" json:"operating_effectiveness_rating"`

	CoverageCompletenessWeight   float64 `gorm:"type:decimal(5,2);default:20.0" json:"coverage_completeness_weight"`
	CoverageCompletenessRating   int     `gorm:"type:int;default:3" json:"coverage_completeness_rating"`

	TimelinessWeight             float64 `gorm:"type:decimal(5,2);default:20.0" json:"timeliness_weight"`
	TimelinessRating             int     `gorm:"type:int;default:3" json:"timeliness_rating"`

	AutomationMonitoringWeight   float64 `gorm:"type:decimal(5,2);default:20.0" json:"automation_monitoring_weight"`
	AutomationMonitoringRating   int     `gorm:"type:int;default:3" json:"automation_monitoring_rating"`

	// Total Weighted Score (sum of Rating * Weight / 100)
	TotalWeightedScore float64 `gorm:"type:decimal(5,2);default:3.0" json:"total_weighted_score"`

	// Risk Profile Post Year-End snapshot values for effectiveness calculation
	InherentRisk  int `gorm:"type:int;default:20" json:"inherent_risk"`
	ResidualRisk  int `gorm:"type:int;default:8" json:"residual_risk"`

	Notes string `gorm:"type:text" json:"notes"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RiskControlMatrix) TableName() string {
	return "risk_control_matrix"
}

func (rcm *RiskControlMatrix) CalculateWeightedScore() float64 {
	desWeight := rcm.DesignEffectivenessWeight
	if desWeight <= 0 {
		desWeight = 20.0
	}
	opWeight := rcm.OperatingEffectivenessWeight
	if opWeight <= 0 {
		opWeight = 20.0
	}
	covWeight := rcm.CoverageCompletenessWeight
	if covWeight <= 0 {
		covWeight = 20.0
	}
	timeWeight := rcm.TimelinessWeight
	if timeWeight <= 0 {
		timeWeight = 20.0
	}
	autoWeight := rcm.AutomationMonitoringWeight
	if autoWeight <= 0 {
		autoWeight = 20.0
	}

	score := (float64(rcm.DesignEffectivenessRating) / 5.0 * desWeight) +
		(float64(rcm.OperatingEffectivenessRating) / 5.0 * opWeight) +
		(float64(rcm.CoverageCompletenessRating) / 5.0 * covWeight) +
		(float64(rcm.TimelinessRating) / 5.0 * timeWeight) +
		(float64(rcm.AutomationMonitoringRating) / 5.0 * autoWeight)
	rcm.TotalWeightedScore = math.Round(score*100) / 100
	return rcm.TotalWeightedScore
}
