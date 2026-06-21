package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RiskMitigation struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RiskID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"riskId"`
	RiskEvent      string         `gorm:"type:text;not null" json:"riskEvent"`
	MitigationPlan string         `gorm:"type:text;not null" json:"mitigationPlan"`
	Supervisor     string         `gorm:"type:varchar(200)" json:"supervisor"`
	PIC            string         `gorm:"type:varchar(200)" json:"pic"`
	UnitInCharge   string         `gorm:"type:varchar(200)" json:"unitInCharge"`
	StartDate      time.Time      `json:"start_date"`
	EndDate        time.Time      `json:"end_date"`
	Notes          string         `gorm:"type:text" json:"notes"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RiskMitigation) TableName() string {
	return "risk_mitigation"
}
