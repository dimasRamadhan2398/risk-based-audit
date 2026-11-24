package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StrategicPlan struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code               string         `gorm:"type:varchar(50)" json:"code"`
	GoalID             string         `gorm:"type:varchar(255)" json:"goalId"`
	StrategicObjective string         `gorm:"type:text" json:"strategicObjective"`
	KPI                string         `gorm:"type:text" json:"kpi"`
	Unit               string         `gorm:"type:varchar(50)" json:"unit"`
	HibHig             string         `gorm:"type:varchar(50)" json:"hibHig"`
	PeriodType         string         `gorm:"type:varchar(50)" json:"periodType"`
	SelectedPeriod     string         `gorm:"type:varchar(50)" json:"selectedPeriod"`
	YearStart          int            `gorm:"type:int" json:"yearStart"`
	YearEnd            int            `gorm:"type:int" json:"yearEnd"`
	KPITargets         map[string]string `gorm:"serializer:json" json:"kpiTargets"`
	KPIActuals         map[string]string `gorm:"serializer:json" json:"kpiActuals"`
	InternalAuditSO    string         `gorm:"type:text" json:"internalAuditSO"`
	Actual             string         `gorm:"type:varchar(100)" json:"actual"`
	Target             string         `gorm:"type:varchar(100)" json:"target"`
	Calculation        string         `gorm:"type:varchar(100)" json:"calculation"`
	Status             string         `gorm:"type:varchar(50)" json:"status"`
	// Category is one of StrategicPlanCategories, or "" for uncategorised
	// (existing rows). Validated on POST/PUT by controllers/strategic_plan.
	Category           string         `gorm:"type:varchar(50);default:''" json:"category"`
	// Initiative is the main program/activity (Program/Kegiatan Utama) that
	// supports this KPI. InitiativeYears says, per year key (same keys as
	// KPITargets), whether it is carried out that year: "√" yes, "-" no.
	Initiative         string            `gorm:"type:text" json:"initiative"`
	InitiativeYears    map[string]string `gorm:"serializer:json" json:"initiativeYears"`
	// TargetNotes holds a free-text remark per year key, e.g. "PC" for a target
	// that is only assessed periodically (external assessment).
	TargetNotes        map[string]string `gorm:"serializer:json" json:"targetNotes"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (StrategicPlan) TableName() string {
	return "strategic_plans"
}
