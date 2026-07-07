package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// VisionMissionGoalsStatus enum
type VisionMissionGoalsStatus string

const (
	VmgStatusDraft     VisionMissionGoalsStatus = "Draft"
	VmgStatusInReview VisionMissionGoalsStatus = "In Review"
	VmgStatusApproved VisionMissionGoalsStatus = "Approved"
	VmgStatusPublished VisionMissionGoalsStatus = "Published"
)

// VisionMissionGoals represents the main entity for Vision, Mission & Goals
type VisionMissionGoals struct {
	ID             uuid.UUID               `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID      uuid.UUID               `gorm:"type:uuid;not null;index" json:"company_id"`
	Company        *Company                `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Period         string                  `gorm:"type:varchar(50);not null" json:"period"` // e.g., "2026 - 2031"
	EffectiveDate  *time.Time              `gorm:"type:date" json:"effective_date"`
	Vision         string                  `gorm:"type:text" json:"vision"`
	Mission        string                  `gorm:"type:text" json:"mission"`
	Version        string                  `gorm:"type:varchar(20)" json:"version"`
	Status         VisionMissionGoalsStatus `gorm:"type:varchar(20);default:'Draft'" json:"status"`
	Notes          string                  `gorm:"type:text" json:"notes"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
	CreatedBy      string                  `gorm:"type:varchar(100)" json:"created_by"`
	ModifiedBy     string                  `gorm:"type:varchar(100)" json:"modified_by"`
	DeletedAt      gorm.DeletedAt          `gorm:"index" json:"-"`
	Goals          []VmgGoal               `gorm:"foreignKey:VmgID" json:"goals,omitempty"`
}

func (VisionMissionGoals) TableName() string {
	return "vision_mission_goals"
}

// VmgGoal represents individual goal within VisionMissionGoals
type VmgGoal struct {
	ID                  uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	VmgID               uuid.UUID `gorm:"type:uuid;not null;index" json:"vmg_id"`
	GoalCode            string    `gorm:"type:varchar(20);not null" json:"goal_code"` // e.g., "G-001"
	GoalName            string    `gorm:"type:varchar(255);not null" json:"goal_name"`
	GoalDescription     string    `gorm:"type:text" json:"goal_description"`
	StrategicObjective  string    `gorm:"type:varchar(255)" json:"strategic_objective"`
	KPI                 string    `gorm:"type:varchar(255)" json:"kpi"`
	Target              string    `gorm:"type:varchar(100)" json:"target"`
	Unit                string    `gorm:"type:varchar(50)" json:"unit"` // %, Rp, Amount, dll
	BaselineYear        string    `gorm:"type:varchar(10)" json:"baseline_year"`
	BaselineValue       string    `gorm:"type:varchar(100)" json:"baseline_value"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func (VmgGoal) TableName() string {
	return "vmg_goals"
}
