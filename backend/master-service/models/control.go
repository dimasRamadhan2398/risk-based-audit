package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ControlType string
type ControlCategory string
type ControlEffectiveness string

const (
	ControlTypePreventive  ControlType = "PREVENTIVE"
	ControlTypeDetective   ControlType = "DETECTIVE"
	ControlTypeCorrective  ControlType = "CORRECTIVE"

	ControlCategoryManual      ControlCategory = "MANUAL"
	ControlCategoryAutomated   ControlCategory = "AUTOMATED"
	ControlCategoryITGeneral   ControlCategory = "IT_GENERAL"
	ControlCategoryITApplication ControlCategory = "IT_APPLICATION"

	ControlEffectivenessEffective      ControlEffectiveness = "EFFECTIVE"
	ControlEffectivenessIneffective   ControlEffectiveness = "INEFFECTIVE"
	ControlEffectivenessNeedsImprovement ControlEffectiveness = "NEEDS_IMPROVEMENT"
	ControlEffectivenessNotTested     ControlEffectiveness = "NOT_TESTED"
)

type Control struct {
	ID              uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ControlCode     string              `gorm:"type:varchar(50);uniqueIndex;not null" json:"control_code"`
	ControlName     string              `gorm:"type:varchar(200);not null" json:"control_name"`
	Description     string              `gorm:"type:text" json:"description"`
	ControlType     ControlType         `gorm:"type:varchar(20);not null" json:"control_type"`
	ControlCategory ControlCategory     `gorm:"type:varchar(30);not null" json:"control_category"`

	// Link to risk it mitigates
	RiskRegisterID  uuid.UUID           `gorm:"type:uuid;index" json:"risk_register_id"`
	RiskRegister    RiskRegister        `gorm:"foreignKey:RiskRegisterID" json:"risk_register"`

	// Control owner
	OwnerID         uuid.UUID           `gorm:"type:uuid;index" json:"owner_id"`
	Owner           Employee            `gorm:"foreignKey:OwnerID" json:"owner"`

	// Department responsible
	DepartmentID    uuid.UUID           `gorm:"type:uuid;index" json:"department_id"`
	Department      Department          `gorm:"foreignKey:DepartmentID" json:"department"`

	Frequency       string              `gorm:"type:varchar(50)" json:"frequency"`      // daily, weekly, monthly, quarterly, annually
	Documentation   string              `gorm:"type:text" json:"documentation"`         // control documentation/policy reference
	IsKeyControl    bool                `gorm:"default:false" json:"is_key_control"`
	IsActive        bool                `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	DeletedAt       gorm.DeletedAt      `gorm:"index" json:"-"`
}
