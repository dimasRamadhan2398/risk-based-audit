package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Office functions (docs/office-structure-plan.md 2.2, D5). The function of an
// office is independent of its level.
const (
	OfficeFunctionHeadOffice  = "head_office"
	OfficeFunctionRnD         = "rnd"
	OfficeFunctionOperational = "operational"
	OfficeFunctionWarehouse   = "warehouse"
	OfficeFunctionDataCenter  = "data_center"
	OfficeFunctionOther       = "other"
)

// OfficeFunctions lists every allowed office_function.
var OfficeFunctions = []string{
	OfficeFunctionHeadOffice, OfficeFunctionRnD, OfficeFunctionOperational,
	OfficeFunctionWarehouse, OfficeFunctionDataCenter, OfficeFunctionOther,
}

// IsOfficeFunction reports whether f is an allowed office_function.
func IsOfficeFunction(f string) bool {
	for _, v := range OfficeFunctions {
		if v == f {
			return true
		}
	}
	return false
}

// Location is an office (docs/office-structure-plan.md D1/D2). The table and
// its id stay; risk-service, audit-service and employees keep referencing it.
// The office columns are additive and nullable: company_id is NULL only until
// `master backfill-offices` (or an admin) assigns an owner company.
type Location struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name       string         `gorm:"type:varchar(100);not null" json:"name"` // e.g. "HQ Jakarta", "Surabaya Branch"
	Address    string         `gorm:"type:text;not null" json:"address"`      // Street address
	City       string         `gorm:"type:varchar(100);not null" json:"city"`
	Province   string         `gorm:"type:varchar(100)" json:"province"`
	PostalCode string         `gorm:"type:varchar(20)" json:"postal_code"`
	Country    string         `gorm:"type:varchar(100);default:'Indonesia'" json:"country"`
	Latitude   *float64       `gorm:"type:decimal(10,8)" json:"latitude,omitempty"`  // optional GPS
	Longitude  *float64       `gorm:"type:decimal(11,8)" json:"longitude,omitempty"` // optional GPS
	IsActive   bool           `gorm:"default:true" json:"is_active"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Office structure. Foreign keys and partial unique indexes are created by
	// pkg/database/officestructure (no GORM relations, to keep the
	// Company <-> Location migrate order free of cycles).
	CompanyID        *uuid.UUID `gorm:"type:uuid;index" json:"company_id"`
	ParentLocationID *uuid.UUID `gorm:"type:uuid;index" json:"parent_location_id"`
	Level            *int       `gorm:"type:int" json:"level"`
	OfficeCode       *string    `gorm:"type:varchar(30)" json:"office_code"`
	OfficeFunction   string     `gorm:"type:varchar(30);not null;default:'operational'" json:"office_function"`
	HeadEmployeeID   *uuid.UUID `gorm:"type:uuid" json:"head_employee_id"`
}
