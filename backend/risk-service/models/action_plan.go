package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ActionPlanStatus string

const (
	ActionPlanStatusOpen       ActionPlanStatus = "OPEN"
	ActionPlanStatusInProgress ActionPlanStatus = "IN_PROGRESS"
	ActionPlanStatusClosed     ActionPlanStatus = "CLOSED"
)

type ActionPlan struct {
	ID                uuid.UUID        `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RiskID            uuid.UUID        `gorm:"type:uuid;not null;index" json:"risk_id"`
	Risk              RiskRegister     `gorm:"foreignKey:RiskID" json:"risk"`
	ActionDescription string           `gorm:"type:text;not null" json:"action_description"`
	PICID             uuid.UUID        `gorm:"type:uuid;not null;index" json:"pic_id"`
	DueDate           time.Time        `gorm:"type:date;not null" json:"due_date"`
	Status            ActionPlanStatus `gorm:"type:varchar(20);default:'OPEN'" json:"status"`
	AgingDays         int              `gorm:"type:int;default:0" json:"aging_days"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
	DeletedAt         gorm.DeletedAt   `gorm:"index" json:"-"`
}

func (ActionPlan) TableName() string {
	return "action_plan"
}
