package models

import (
	"time"

	"github.com/google/uuid"
)

type CorporateAuditUniverse struct {
	ID                      uuid.UUID                `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	StandardAuditUniverseID *uuid.UUID               `gorm:"type:uuid;index" json:"standard_audit_universe_id,omitempty"` // nullable for custom entities
	StandardAuditUniverse   *StandardAuditUniverse   `gorm:"foreignKey:StandardAuditUniverseID" json:"standard_audit_universe,omitempty"`
	Name                    string                   `gorm:"type:varchar(200);not null" json:"name"`
	ParentID                *uuid.UUID               `gorm:"type:uuid;index" json:"parent_id,omitempty"`
	Parent                  *CorporateAuditUniverse  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children                []CorporateAuditUniverse `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	// DepartmentID is the master-service department that owns this entry. NULL
	// for entries created before department scoping and for sync-master rows;
	// only full audit-universe writers may edit those.
	DepartmentID *uuid.UUID `gorm:"type:uuid;index" json:"department_id"`
	// DepartmentName is resolved from master-service on read; never stored.
	DepartmentName *string   `gorm:"-" json:"department_name"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (CorporateAuditUniverse) TableName() string {
	return "corporate_audit_universe"
}
