package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ControlType string
type ControlEvaluationResult string

const (
	ControlTypePreventive ControlType = "PREVENTIVE"
	ControlTypeDetective  ControlType = "DETECTIVE"
	ControlTypeCorrective ControlType = "CORRECTIVE"

	ControlEvaluationEffective   ControlEvaluationResult = "EFFECTIVE"
	ControlEvaluationIneffective ControlEvaluationResult = "INEFFECTIVE"
)

type RiskControl struct {
	ID                 uuid.UUID               `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RiskID             uuid.UUID               `gorm:"type:uuid;not null;index" json:"risk_id"`
	Risk               RiskRegister            `gorm:"foreignKey:RiskID" json:"risk"`
	ControlDescription string                  `gorm:"type:text;not null" json:"control_description"`
	ControlType        ControlType             `gorm:"type:varchar(20);not null" json:"control_type"`
	EvaluationResult   ControlEvaluationResult `gorm:"type:varchar(20);not null" json:"evaluation_result"`
	CreatedAt          time.Time               `json:"created_at"`
	UpdatedAt          time.Time               `json:"updated_at"`
	DeletedAt          gorm.DeletedAt          `gorm:"index" json:"-"`
}

func (RiskControl) TableName() string {
	return "risk_control"
}
