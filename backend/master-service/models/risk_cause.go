package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RiskCauseCategory string

const (
	RiskCauseCategoryInternal     RiskCauseCategory = "INTERNAL"
	RiskCauseCategoryExternal     RiskCauseCategory = "EXTERNAL"
	RiskCauseCategoryProcess      RiskCauseCategory = "PROCESS"
	RiskCauseCategoryPeople       RiskCauseCategory = "PEOPLE"
	RiskCauseCategoryTechnology    RiskCauseCategory = "TECHNOLOGY"
	RiskCauseCategoryExternalEvent RiskCauseCategory = "EXTERNAL_EVENT"
)

type RiskCause struct {
	ID              uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CauseCode       string              `gorm:"type:varchar(50);uniqueIndex;not null" json:"cause_code"`
	CauseName       string              `gorm:"type:varchar(200);not null" json:"cause_name"`
	Description     string              `gorm:"type:text" json:"description"`
	Category        RiskCauseCategory   `gorm:"type:varchar(30);not null" json:"category"`

	// Link to risk register
	RiskRegisterID  uuid.UUID           `gorm:"type:uuid;index" json:"risk_register_id"`
	RiskRegister    RiskRegister        `gorm:"foreignKey:RiskRegisterID" json:"risk_register"`

	IsActive        bool                `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	DeletedAt       gorm.DeletedAt      `gorm:"index" json:"-"`
}
