package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IndicatorStatus string
type IndicatorDirection string

const (
	IndicatorStatusActive    IndicatorStatus = "ACTIVE"
	IndicatorStatusInactive  IndicatorStatus = "INACTIVE"
	IndicatorStatusAlert     IndicatorStatus = "ALERT"

	IndicatorDirectionIncreasing  IndicatorDirection = "INCREASING"
	IndicatorDirectionDecreasing  IndicatorDirection = "DECREASING"
	IndicatorDirectionStable      IndicatorDirection = "STABLE"
)

type RiskIndicator struct {
	ID              uuid.UUID            `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	IndicatorCode   string              `gorm:"type:varchar(50);uniqueIndex;not null" json:"indicator_code"`
	IndicatorName   string              `gorm:"type:varchar(200);not null" json:"indicator_name"`
	Description     string              `gorm:"type:text" json:"description"`

	// Link to risk
	RiskRegisterID  uuid.UUID            `gorm:"type:uuid;not null;index" json:"risk_register_id"`
	RiskRegister    RiskRegister         `gorm:"foreignKey:RiskRegisterID" json:"risk_register"`

	// Measurement
	Metric          string              `gorm:"type:varchar(100)" json:"metric"`       // what is measured
	Unit            string              `gorm:"type:varchar(20)" json:"unit"`           // %, count, currency, etc.
	Frequency       string              `gorm:"type:varchar(20)" json:"frequency"`       // daily, weekly, monthly

	// Thresholds
	ThresholdMin    *float64            `gorm:"type:decimal(15,2)" json:"threshold_min,omitempty"`
	ThresholdMax    *float64            `gorm:"type:decimal(15,2)" json:"threshold_max,omitempty"`
	ToleranceLevel  *float64            `gorm:"type:decimal(15,2)" json:"tolerance_level,omitempty"`

	// Current value
	CurrentValue    *float64            `gorm:"type:decimal(15,2)" json:"current_value,omitempty"`
	LastUpdatedAt   *time.Time          `json:"last_updated_at,omitempty"`

	// Trend
	Trend           IndicatorDirection  `gorm:"type:varchar(20)" json:"trend"`
	TrendComment    string              `gorm:"type:text" json:"trend_comment"`

	// Status
	Status          IndicatorStatus     `gorm:"type:varchar(20);default:'ACTIVE'" json:"status"`

	// Data source
	DataSource      string              `gorm:"type:varchar(100)" json:"data_source"`    // system, manual entry, etc.
	DataSourceURL   string              `gorm:"type:varchar(500)" json:"data_source_url,omitempty"`

	// Owner
	OwnerID         uuid.UUID           `gorm:"type:uuid;index" json:"owner_id"`
	Owner           Employee            `gorm:"foreignKey:OwnerID" json:"owner"`

	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	DeletedAt       gorm.DeletedAt      `gorm:"index" json:"-"`
}

// RiskIndicatorLog - historical readings of the indicator
type RiskIndicatorLog struct {
	ID              uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	IndicatorID     uuid.UUID           `gorm:"type:uuid;not null;index" json:"indicator_id"`
	Indicator       RiskIndicator       `gorm:"foreignKey:IndicatorID" json:"indicator"`

	Value           float64             `gorm:"type:decimal(15,2);not null" json:"value"`
	RecordedAt      time.Time           `gorm:"not null" json:"recorded_at"`

	Note            string              `gorm:"type:text" json:"note,omitempty"`

	CreatedAt       time.Time           `json:"created_at"`
}
