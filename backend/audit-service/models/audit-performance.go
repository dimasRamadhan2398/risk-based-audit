package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// KPIAchievement represents Feature 12: Internal Audit Performance - KPI Achievement
type KPIAchievement struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Year            int            `gorm:"type:int;not null" json:"year"`
	KPIName         string         `gorm:"type:varchar(255);not null" json:"kpi_name"`
	Target          float64        `gorm:"type:decimal(15,2)" json:"target"`
	Actual          float64        `gorm:"type:decimal(15,2)" json:"actual"`
	AchievementRate float64        `gorm:"type:decimal(5,2)" json:"achievement_rate"` // Percentage
	Notes           string         `gorm:"type:text" json:"notes"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (KPIAchievement) TableName() string {
	return "kpi_achievements"
}

// WorkPlanRealization represents Feature 12: Internal Audit Performance - Work Plan Realization
type WorkPlanRealization struct {
	ID                uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Year              int            `gorm:"type:int;not null" json:"year"`
	AuditAnnualPlanID uuid.UUID      `gorm:"type:uuid;not null;index" json:"audit_annual_plan_id"`
	AnnualPlan        AuditAnnual    `gorm:"foreignKey:AuditAnnualPlanID" json:"annual_plan"`
	PlannedActivities int            `gorm:"type:int" json:"planned_activities"`
	ExecutedActivities int            `gorm:"type:int" json:"executed_activities"`
	RealizationRate   float64        `gorm:"type:decimal(5,2)" json:"realization_rate"` // Percentage
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WorkPlanRealization) TableName() string {
	return "work_plan_realizations"
}
