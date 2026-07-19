package models

import (
	"time"

	"github.com/google/uuid"
)

type CorporateAuditUniverse struct {
	ID                      uuid.UUID               `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	StandardAuditUniverseID *uuid.UUID              `gorm:"type:uuid;index" json:"standard_audit_universe_id,omitempty"` // nullable for custom entities
	StandardAuditUniverse   *StandardAuditUniverse  `gorm:"foreignKey:StandardAuditUniverseID" json:"standard_audit_universe,omitempty"`
	Name                    string                  `gorm:"type:varchar(200);not null" json:"name"`
	ParentID                *uuid.UUID              `gorm:"type:uuid;index" json:"parent_id,omitempty"`
	Parent                  *CorporateAuditUniverse `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children                []CorporateAuditUniverse `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	CreatedAt               time.Time               `json:"created_at"`
	UpdatedAt               time.Time               `json:"updated_at"`
}

func (CorporateAuditUniverse) TableName() string {
	return "corporate_audit_universe"
}
