package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UniverseEntityType string
type UniverseStatus string

const (
	UniverseEntityTypeProcess      UniverseEntityType = "PROCESS"
	UniverseEntityTypeDepartment   UniverseEntityType = "DEPARTMENT"
	UniverseEntityTypeSystem        UniverseEntityType = "SYSTEM"
	UniverseEntityTypeVendor       UniverseEntityType = "VENDOR"
	UniverseEntityTypeProject      UniverseEntityType = "PROJECT"
	UniverseEntityTypeLocation     UniverseEntityType = "LOCATION"
	UniverseEntityTypeProduct      UniverseEntityType = "PRODUCT"

	UniverseStatusActive          UniverseStatus = "ACTIVE"
	UniverseStatusInactive        UniverseStatus = "INACTIVE"
	UniverseStatusUnderAudit      UniverseStatus = "UNDER_AUDIT"
	UniverseStatusAudited         UniverseStatus = "AUDITED"
)

type AuditUniverse struct {
	ID              uuid.UUID            `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EntityCode      string              `gorm:"type:varchar(50);uniqueIndex;not null" json:"entity_code"`
	EntityName      string              `gorm:"type:varchar(200);not null" json:"entity_name"`
	EntityType      UniverseEntityType  `gorm:"type:varchar(30);not null" json:"entity_type"`

	// Description
	Description     string              `gorm:"type:text" json:"description"`
	BusinessOwner   string              `gorm:"type:varchar(100)" json:"business_owner"`

	// Hierarchy
	ParentID        *uuid.UUID          `gorm:"type:uuid;index" json:"parent_id,omitempty"`
	Parent          *AuditUniverse       `gorm:"foreignKey:ParentID" json:"parent,omitempty"`

	// Risk assessment
	RiskRating      string              `gorm:"type:varchar(20)" json:"risk_rating"`   // based on inherent risk assessment
	LastAuditYear   *int                `gorm:"type:int" json:"last_audit_year,omitempty"`
	LastAuditDate   *time.Time          `json:"last_audit_date,omitempty"`
	LastAuditResult string              `gorm:"type:varchar(50)" json:"last_audit_result,omitempty"`

	// Coverage
	AuditFrequency   int                 `gorm:"type:int;default:1" json:"audit_frequency"` // years between audits
	IsMandatory     bool                `gorm:"default:false" json:"is_mandatory"`   // regulatory required
	IsHighPriority  bool                `gorm:"default:false" json:"is_high_priority"`

	// Department link
	DepartmentID    *uuid.UUID          `gorm:"type:uuid;index" json:"department_id,omitempty"`
	Department      *Department         `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`

	// Status
	Status          UniverseStatus      `gorm:"type:varchar(20);default:'ACTIVE'" json:"status"`

	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	DeletedAt       gorm.DeletedAt      `gorm:"index" json:"-"`
}
