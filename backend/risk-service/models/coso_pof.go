package models

import (
	"time"

	"github.com/google/uuid"
)

// COSO 2013 Point of Focus (PoF) catalogue with "unsur pemenuhan" (fulfilment
// elements). Reference data for the weighted, quantitative assessment method
// (docs/coso-pof-assessment-api.md).
//
// The built-in catalogue is an ORIGINAL AuditSphere catalogue: the wording is
// written from the publicly known COSO 2013 points of focus, in English and
// Indonesian, and is not copied from any third-party working paper.
//
// Rows are immutable per catalogue version. A new version (new weights or
// wording) is a new set of rows; `risk up` inserts what is missing and never
// updates or deletes a row (cmd/coso_catalogue.go), so assessments that
// recorded their catalog_version keep their meaning.

// CosoPofFramework is the framework of the PoF catalogue rows.
const CosoPofFramework = CosoFramework2013

// Catalogue weight levels (coso_catalog_weights.level).
const (
	CosoWeightLevelComponent = "COMPONENT"
	CosoWeightLevelPrinciple = "PRINCIPLE"
)

// Element types.
const (
	CosoElementPresent  = "PRESENT"  // the thing exists / is documented
	CosoElementFunction = "FUNCTION" // the thing operates in practice
)

// Element scale types. DISCRETE accepts 0, 0.25, 0.5, 0.75, 1. QUESTIONNAIRE
// accepts any value in [0, 1] (a questionnaire average).
const (
	CosoScaleDiscrete      = "DISCRETE"
	CosoScaleQuestionnaire = "QUESTIONNAIRE"
)

// CosoCatalogVersion is one published version of the PoF catalogue.
type CosoCatalogVersion struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Framework string    `gorm:"type:varchar(20);not null;default:'COSO_2013';uniqueIndex:uq_coso_catalog_versions,priority:1" json:"framework"`
	Version   string    `gorm:"type:varchar(20);not null;uniqueIndex:uq_coso_catalog_versions,priority:2" json:"version"`
	Name      string    `gorm:"type:varchar(200)" json:"name"`
	// IsCurrent marks the version new assessments use. Exactly one per
	// framework; the seed sets it only when no current version exists.
	IsCurrent bool      `gorm:"not null;default:false" json:"is_current"`
	CreatedAt time.Time `json:"created_at"`
}

func (CosoCatalogVersion) TableName() string { return "coso_catalog_versions" }

// CosoCatalogWeight is the weight of one component or principle in one
// catalogue version (component weights 25/25/20/17.5/12.5, principle weights
// that add up to their component's).
type CosoCatalogWeight struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Framework string    `gorm:"type:varchar(20);not null;uniqueIndex:uq_coso_catalog_weights,priority:1" json:"framework"`
	Version   string    `gorm:"type:varchar(20);not null;uniqueIndex:uq_coso_catalog_weights,priority:2" json:"version"`
	Level     string    `gorm:"type:varchar(12);not null;uniqueIndex:uq_coso_catalog_weights,priority:3" json:"level"`
	Code      string    `gorm:"type:varchar(10);not null;uniqueIndex:uq_coso_catalog_weights,priority:4" json:"code"`
	Weight    float64   `gorm:"type:numeric(9,4);not null" json:"weight"`
	CreatedAt time.Time `json:"created_at"`
}

func (CosoCatalogWeight) TableName() string { return "coso_catalog_weights" }

// CosoPointOfFocus is one point of focus of a principle. Code is
// "component.principle.pof" (for example 1.1.1); principle_code is P01..P17.
type CosoPointOfFocus struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Framework      string    `gorm:"type:varchar(20);not null;uniqueIndex:uq_coso_pofs,priority:1" json:"framework"`
	CatalogVersion string    `gorm:"type:varchar(20);not null;uniqueIndex:uq_coso_pofs,priority:2" json:"catalog_version"`
	Code           string    `gorm:"type:varchar(20);not null;uniqueIndex:uq_coso_pofs,priority:3" json:"code"`
	PrincipleCode  string    `gorm:"type:varchar(10);not null;index" json:"principle_code"`
	ComponentCode  string    `gorm:"type:varchar(10);not null" json:"component_code"`
	NameEn         string    `gorm:"type:varchar(255);not null" json:"name_en"`
	NameID         string    `gorm:"type:varchar(255);not null" json:"name_id"`
	Weight         float64   `gorm:"type:numeric(9,4);not null" json:"weight"`
	SortOrder      int       `gorm:"type:int" json:"sort_order"`
	// DefaultOwnerRole is a comma separated list of owner role labels (for
	// example "RISK_MANAGEMENT,DEPARTMENT_HEAD"). A default only: tenants
	// override it per principle and no permission is derived from it.
	DefaultOwnerRole string    `gorm:"type:varchar(200)" json:"default_owner_role"`
	IsActive         bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
}

func (CosoPointOfFocus) TableName() string { return "coso_points_of_focus" }

// CosoPofElement is one fulfilment element ("unsur") of a point of focus.
// Code is "component.principle.pof.element" (for example 1.1.1.1).
type CosoPofElement struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Framework      string    `gorm:"type:varchar(20);not null;uniqueIndex:uq_coso_pof_elements,priority:1" json:"framework"`
	CatalogVersion string    `gorm:"type:varchar(20);not null;uniqueIndex:uq_coso_pof_elements,priority:2" json:"catalog_version"`
	Code           string    `gorm:"type:varchar(30);not null;uniqueIndex:uq_coso_pof_elements,priority:3" json:"code"`
	PofCode        string    `gorm:"type:varchar(20);not null;index" json:"pof_code"`
	Type           string    `gorm:"type:varchar(10);not null" json:"type"`
	DescriptionEn  string    `gorm:"type:text;not null" json:"description_en"`
	DescriptionID  string    `gorm:"type:text;not null" json:"description_id"`
	Weight         float64   `gorm:"type:numeric(9,4);not null" json:"weight"`
	// DefaultMethods is a comma separated subset of D (document), K
	// (questionnaire), W (interview), O (observation).
	DefaultMethods string `gorm:"type:varchar(16)" json:"default_methods"`
	ScaleType      string `gorm:"type:varchar(14);not null;default:'DISCRETE'" json:"scale_type"`
	// RequiresIndependentAssessor: the element evaluates the internal audit
	// function itself, so the CAE must not be the one who fills it in.
	RequiresIndependentAssessor bool      `gorm:"not null;default:false" json:"requires_independent_assessor"`
	SortOrder                   int       `gorm:"type:int" json:"sort_order"`
	CreatedAt                   time.Time `json:"created_at"`
}

func (CosoPofElement) TableName() string { return "coso_pof_elements" }

// CosoPofModels lists the PoF catalogue models for AutoMigrate.
func CosoPofModels() []interface{} {
	return []interface{}{
		&CosoCatalogVersion{},
		&CosoCatalogWeight{},
		&CosoPointOfFocus{},
		&CosoPofElement{},
	}
}
