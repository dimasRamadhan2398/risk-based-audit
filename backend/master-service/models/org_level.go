package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Preset name keys of org levels (docs/org-structure-model-plan.md 4.1). The
// frontend translates a level by its key; a NULL key means "show the stored
// name as typed". Must match the CHECK constraint in pkg/database/orgstructure.
const (
	OrgLevelKeyDirectorate = "directorate"
	OrgLevelKeyDivision    = "division"
	OrgLevelKeyDepartment  = "department"
	OrgLevelKeySection     = "section"
)

// OrgLevelNameKeys lists every allowed name_key.
var OrgLevelNameKeys = []string{OrgLevelKeyDirectorate, OrgLevelKeyDivision, OrgLevelKeyDepartment, OrgLevelKeySection}

// IsOrgLevelNameKey reports whether k is an allowed name_key.
func IsOrgLevelNameKey(k string) bool {
	for _, v := range OrgLevelNameKeys {
		if v == k {
			return true
		}
	}
	return false
}

// OrgLevel is one level of a flexible tenant's organisation (Direktorat,
// Divisi, ...). Levels are gapped numbers (10, 20, 30); departments.level holds
// the number. The table is created by the gated DDL only and is NOT in the
// AutoMigrate list.
type OrgLevel struct {
	ID        uuid.UUID      `gorm:"column:id;type:uuid;primaryKey" json:"id"`
	Level     int            `gorm:"column:level" json:"level"`
	Name      string         `gorm:"column:name" json:"name"`
	NameKey   *string        `gorm:"column:name_key" json:"name_key"`
	IsActive  bool           `gorm:"column:is_active" json:"is_active"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at" json:"-"`
}

func (OrgLevel) TableName() string { return "org_levels" }
