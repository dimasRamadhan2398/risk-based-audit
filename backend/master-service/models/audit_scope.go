package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditScopeType string

const (
	AuditScopeTypeProcess        AuditScopeType = "PROCESS"
	AuditScopeTypeDepartment     AuditScopeType = "DEPARTMENT"
	AuditScopeTypeSystem         AuditScopeType = "SYSTEM"
	AuditScopeTypeProject        AuditScopeType = "PROJECT"
	AuditScopeTypeCompliance     AuditScopeType = "COMPLIANCE"
)

type AuditScope struct {
	ID              uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ScopeCode       string              `gorm:"type:varchar(50);uniqueIndex;not null" json:"scope_code"`
	ScopeName       string              `gorm:"type:varchar(200);not null" json:"scope_name"`
	ScopeType       AuditScopeType      `gorm:"type:varchar(30);not null" json:"scope_type"`
	Description     string              `gorm:"type:text" json:"description"`

	// Link to audit plan
	AuditPlanID     uuid.UUID           `gorm:"type:uuid;not null;index" json:"audit_plan_id"`
	AuditPlan       AnnualAuditPlan     `gorm:"foreignKey:AuditPlanID" json:"audit_plan"`

	// Coverage
	InScope         string              `gorm:"type:text" json:"in_scope"`         // what will be covered
	OutOfScope      string              `gorm:"type:text" json:"out_of_scope"`      // what will NOT be covered

	// Reference entities
	DepartmentID    *uuid.UUID          `gorm:"type:uuid;index" json:"department_id,omitempty"`
	Department      *Department         `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	RiskRegisterID  *uuid.UUID          `gorm:"type:uuid;index" json:"risk_register_id,omitempty"`
	RiskRegister    *RiskRegister       `gorm:"foreignKey:RiskRegisterID" json:"risk_register,omitempty"`

	// Objectives
	Objectives      string              `gorm:"type:text" json:"objectives"`

	IsActive        bool                `gorm:"default:true" json:"is_active"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	DeletedAt       gorm.DeletedAt      `gorm:"index" json:"-"`
}
