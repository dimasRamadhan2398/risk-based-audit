package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Preset name keys of office levels (docs/office-structure-plan.md 2.1). The
// frontend translates a level by its key; a NULL key means "show name".
const (
	OfficeLevelKeyHeadOffice     = "head_office"
	OfficeLevelKeyRegionalOffice = "regional_office"
	OfficeLevelKeyAreaOffice     = "area_office"
	OfficeLevelKeyBranch         = "branch"
	OfficeLevelKeySubBranch      = "sub_branch"
	OfficeLevelKeyUnit           = "unit"
)

// OfficeLevelNameKeys lists every allowed office level name_key.
var OfficeLevelNameKeys = []string{
	OfficeLevelKeyHeadOffice, OfficeLevelKeyRegionalOffice, OfficeLevelKeyAreaOffice,
	OfficeLevelKeyBranch, OfficeLevelKeySubBranch, OfficeLevelKeyUnit,
}

// IsOfficeLevelNameKey reports whether k is an allowed name_key.
func IsOfficeLevelNameKey(k string) bool {
	for _, v := range OfficeLevelNameKeys {
		if v == k {
			return true
		}
	}
	return false
}

// OfficeLevel is one level of an office scheme. CompanyID NULL is the tenant
// default scheme; otherwise the scheme of that company (inherited by its
// descendants that have none). Levels are gapped numbers (10, 20, 30), lower =
// higher in the tree; locations.level holds the number.
//
// Create rows with Select("*") (or set IsActive after create): with the
// default:true tag GORM would otherwise drop is_active=false on insert.
type OfficeLevel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyID *uuid.UUID     `gorm:"type:uuid;index" json:"company_id"`
	Level     int            `gorm:"type:int;not null" json:"level"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	NameKey   *string        `gorm:"type:varchar(50)" json:"name_key"`
	IsActive  bool           `gorm:"not null;default:true" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (OfficeLevel) TableName() string { return "office_levels" }

// LocationCompany shares an office with a company other than its owner
// (docs/office-structure-plan.md D2, 2.3). It does not place the office in
// that company's tree.
type LocationCompany struct {
	LocationID uuid.UUID `gorm:"type:uuid;primaryKey" json:"location_id"`
	CompanyID  uuid.UUID `gorm:"type:uuid;primaryKey;index" json:"company_id"`
	CreatedAt  time.Time `json:"created_at"`
}

func (LocationCompany) TableName() string { return "location_companies" }

// DefaultOfficeLevels is the tenant default scheme seeded on every tenant.
func DefaultOfficeLevels() []OfficeLevel {
	key := func(s string) *string { return &s }
	return []OfficeLevel{
		{Level: 10, Name: "Kantor Pusat", NameKey: key(OfficeLevelKeyHeadOffice), IsActive: true},
		{Level: 20, Name: "Kantor Wilayah", NameKey: key(OfficeLevelKeyRegionalOffice), IsActive: true},
		{Level: 30, Name: "Kantor Cabang", NameKey: key(OfficeLevelKeyBranch), IsActive: true},
	}
}
