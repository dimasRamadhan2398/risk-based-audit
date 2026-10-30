package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Department is an org unit. In legacy tenants it is flat; in flexible tenants
// (ORG_STRUCTURE_MODE=flexible) it also has a parent, read and written through
// DepartmentOrg.
//
// Do NOT add persisted fields here: this struct is auto-migrated on every
// tenant, so a new column would alter legacy tenants on their next migrate.
// The fields tagged gorm:"-" are response-only and filled in flexible mode.
//
// DepartmentCode and DepartmentName carry no uniqueIndex tag on purpose. The
// legacy unique indexes (idx_departments_department_code/_name) are created by
// pkg/database/orgstructure only when they are missing on a legacy database;
// flexible tenants replace them with partial indexes.
type Department struct {
	ID                    uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DepartmentCode        string         `gorm:"type:varchar(50);not null" json:"department_code"`
	DepartmentName        string         `gorm:"type:varchar(100);not null" json:"department_name"`
	DepartmentDescription string         `gorm:"type:text" json:"department_description"`
	PicID                 *uuid.UUID     `gorm:"type:uuid;index" json:"pic_id"`
	Level                 int            `gorm:"type:int;not null" json:"level"`
	IsActive              bool           `gorm:"default:true" json:"is_active"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`

	CompanyID uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	Company   Company   `gorm:"foreignKey:CompanyID" json:"company"`

	BusinessUnitID *uuid.UUID    `gorm:"type:uuid;index" json:"business_unit_id"`
	BusinessUnit   *BusinessUnit `gorm:"foreignKey:BusinessUnitID" json:"business_unit,omitempty"`

	// Response-only, flexible mode (never persisted through this struct).
	ParentDepartmentID *uuid.UUID          `gorm:"-" json:"parent_department_id,omitempty"`
	LevelName          string              `gorm:"-" json:"level_name,omitempty"`
	NameKey            *string             `gorm:"-" json:"name_key,omitempty"`
	Path               []DepartmentPathSeg `gorm:"-" json:"path,omitempty"`
}
