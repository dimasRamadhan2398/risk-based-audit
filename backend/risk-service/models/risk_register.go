package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RiskSource string
type RiskFinalLevel string
type RiskRegisterStatus string

const (
	RiskSourceAudit      RiskSource = "AUDIT"
	RiskSourceCompliance RiskSource = "COMPLIANCE"
	RiskSourceAdvisory   RiskSource = "ADVISORY"
	RiskSourceDirect     RiskSource = "DIRECT"

	RiskFinalLevelLow      RiskFinalLevel = "LOW"
	RiskFinalLevelMedium   RiskFinalLevel = "MEDIUM"
	RiskFinalLevelHigh     RiskFinalLevel = "HIGH"
	RiskFinalLevelCritical RiskFinalLevel = "CRITICAL"

	RiskRegisterStatusPending  RiskRegisterStatus = "PENDING"
	RiskRegisterStatusApproved RiskRegisterStatus = "APPROVED"
	RiskRegisterStatusRejected RiskRegisterStatus = "REJECTED"
)

type RiskRegister struct {
	ID                    uuid.UUID          `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProfileID             uuid.UUID          `gorm:"type:uuid;not null;index" json:"profile_id"`
	Profile               RiskProfile        `gorm:"foreignKey:ProfileID" json:"profile"`
	RiskSource            RiskSource         `gorm:"type:varchar(20);not null" json:"risk_source"`
	RiskEvent             string             `gorm:"type:text;not null" json:"risk_event"`
	InherentLikelihood    int                `gorm:"type:int;not null" json:"inherent_likelihood"`
	InherentImpact        int                `gorm:"type:int;not null" json:"inherent_impact"`
	InherentScore         int                `gorm:"type:int;not null" json:"inherent_score"`
	ControlEffectiveness  int                `gorm:"type:int;not null" json:"control_effectiveness"`
	ResidualScore         int                `gorm:"type:int;not null" json:"residual_score"`
	IsAuditorOverride     bool               `gorm:"default:false" json:"is_auditor_override"`
	OverrideScore         *int               `gorm:"type:int" json:"override_score,omitempty"`
	OverrideJustification string             `gorm:"type:text" json:"override_justification"`
	FinalRiskLevel        RiskFinalLevel     `gorm:"type:varchar(20);not null" json:"final_risk_level"`
	Status                RiskRegisterStatus `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	CreatedAt             time.Time          `json:"created_at"`
	UpdatedAt             time.Time          `json:"updated_at"`
	DeletedAt             gorm.DeletedAt     `gorm:"index" json:"-"`
}

func (RiskRegister) TableName() string {
	return "risk_register"
}
