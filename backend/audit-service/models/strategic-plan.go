package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StrategicPlan struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code               string         `gorm:"type:varchar(50)" json:"code"`
	StrategicObjective string         `gorm:"type:text" json:"strategicObjective"`
	KPI                string         `gorm:"type:text" json:"kpi"`
	Unit               string         `gorm:"type:varchar(50)" json:"unit"`
	HibHig             string         `gorm:"type:varchar(50)" json:"hibHig"`
	PeriodType         string         `gorm:"type:varchar(50)" json:"periodType"`
	SelectedPeriod     string         `gorm:"type:varchar(50)" json:"selectedPeriod"`
	YearStart          int            `gorm:"type:int" json:"yearStart"`
	YearEnd            int            `gorm:"type:int" json:"yearEnd"`
	KPITargets         map[int]string `gorm:"serializer:json" json:"kpiTargets"`
	InternalAuditSO    string         `gorm:"type:text" json:"internalAuditSO"`
	Actual             string         `gorm:"type:varchar(100)" json:"actual"`
	Target             string         `gorm:"type:varchar(100)" json:"target"`
	Calculation        string         `gorm:"type:varchar(100)" json:"calculation"`
	Status             string         `gorm:"type:varchar(50)" json:"status"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (StrategicPlan) TableName() string {
	return "strategic_plans"
}
