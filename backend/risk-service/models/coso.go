package models

import (
	"time"

	"github.com/google/uuid"
)

// COSO 2013 Internal Control - Integrated Framework catalogue and the per-RCM
// principle tags (docs/coso-2013-component-assessment-plan.md, 4.1).
//
// The catalogue is reference data: `risk up` inserts missing rows and never
// updates or deletes one (cmd/coso_catalogue.go). Display names come from the
// frontend i18n (coso.components.*, coso.principles.*); NameEn is for exports
// and API readers. audit-service keeps the same codes as a constant list.

// CosoFramework2013 is the framework value of the COSO 2013 rows. The column
// leaves room for COSO ERM 2017 (plan section 11) without new tables.
const CosoFramework2013 = "COSO_2013"

// CosoComponent is one of the five COSO components (CE, RA, CA, IC, MA).
type CosoComponent struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Framework string    `gorm:"type:varchar(20);not null;default:'COSO_2013';uniqueIndex:uq_coso_components_framework_code,priority:1" json:"framework"`
	Code      string    `gorm:"type:varchar(10);not null;uniqueIndex:uq_coso_components_framework_code,priority:2" json:"code"`
	SortOrder int       `gorm:"type:int" json:"sort_order"`
	NameEn    string    `gorm:"type:varchar(100)" json:"name_en"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (CosoComponent) TableName() string { return "coso_components" }

// CosoPrinciple is one of the 17 COSO 2013 principles (P01..P17).
type CosoPrinciple struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Framework     string    `gorm:"type:varchar(20);not null;default:'COSO_2013';uniqueIndex:uq_coso_principles_framework_code,priority:1" json:"framework"`
	Code          string    `gorm:"type:varchar(10);not null;uniqueIndex:uq_coso_principles_framework_code,priority:2" json:"code"`
	Number        int       `gorm:"type:int;not null" json:"number"`
	ComponentCode string    `gorm:"type:varchar(10);not null;index" json:"component_code"`
	NameEn        string    `gorm:"type:varchar(255)" json:"name_en"`
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (CosoPrinciple) TableName() string { return "coso_principles" }

// RcmCosoPrinciple tags an RCM row with a principle it supports.
//
// Deliberately an explicit model and not a many2many association on
// RiskControlMatrix: UpdateRCM saves the row with db.Save, which would upsert
// associations. The tags are written in the same transaction as the row
// (delete the row's tags, insert the new set). RCM rows are soft-deleted, so
// the tags of a deleted row stay but are never read.
//
// The two belongs-to fields exist only so AutoMigrate creates the foreign keys
// (rcm_id ON DELETE CASCADE, principle_id ON DELETE RESTRICT); they are never
// loaded or written (writes use Omit(clause.Associations)).
type RcmCosoPrinciple struct {
	RcmID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"rcm_id"`
	PrincipleID uuid.UUID `gorm:"type:uuid;primaryKey;index" json:"principle_id"`
	CreatedAt   time.Time `json:"created_at"`

	Rcm       *RiskControlMatrix `gorm:"foreignKey:RcmID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	Principle *CosoPrinciple     `gorm:"foreignKey:PrincipleID;references:ID;constraint:OnDelete:RESTRICT" json:"-"`
}

func (RcmCosoPrinciple) TableName() string { return "rcm_coso_principles" }

// CosoPrincipleSeed is one principle of the fixed COSO 2013 catalogue.
type CosoPrincipleSeed struct {
	Code   string
	Number int
	NameEn string
}

// CosoComponentSeed is one component of the fixed COSO 2013 catalogue.
type CosoComponentSeed struct {
	Code       string
	SortOrder  int
	NameEn     string
	Principles []CosoPrincipleSeed
}

// Coso2013Catalogue is the COSO 2013 catalogue (plan 4.1.1). The 17
// principles are fixed by the standard; audit-service mirrors these codes.
var Coso2013Catalogue = []CosoComponentSeed{
	{Code: "CE", SortOrder: 1, NameEn: "Control Environment", Principles: []CosoPrincipleSeed{
		{"P01", 1, "Demonstrates commitment to integrity and ethical values"},
		{"P02", 2, "Board of directors exercises oversight responsibility"},
		{"P03", 3, "Establishes structure, authority and responsibility"},
		{"P04", 4, "Demonstrates commitment to competence"},
		{"P05", 5, "Enforces accountability"},
	}},
	{Code: "RA", SortOrder: 2, NameEn: "Risk Assessment", Principles: []CosoPrincipleSeed{
		{"P06", 6, "Specifies suitable objectives"},
		{"P07", 7, "Identifies and analyzes risk"},
		{"P08", 8, "Assesses fraud risk"},
		{"P09", 9, "Identifies and analyzes significant change"},
	}},
	{Code: "CA", SortOrder: 3, NameEn: "Control Activities", Principles: []CosoPrincipleSeed{
		{"P10", 10, "Selects and develops control activities"},
		{"P11", 11, "Selects and develops general controls over technology"},
		{"P12", 12, "Deploys through policies and procedures"},
	}},
	{Code: "IC", SortOrder: 4, NameEn: "Information & Communication", Principles: []CosoPrincipleSeed{
		{"P13", 13, "Uses relevant information"},
		{"P14", 14, "Communicates internally"},
		{"P15", 15, "Communicates externally"},
	}},
	{Code: "MA", SortOrder: 5, NameEn: "Monitoring Activities", Principles: []CosoPrincipleSeed{
		{"P16", 16, "Conducts ongoing and/or separate evaluations"},
		{"P17", 17, "Evaluates and communicates deficiencies"},
	}},
}
