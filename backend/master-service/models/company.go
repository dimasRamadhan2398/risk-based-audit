package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CompanyType string

const (
	CompanyTypeHolding    CompanyType = "HOLDING"
	CompanyTypeSubsidiary CompanyType = "SUBSIDIARY"
	CompanyTypeBranch     CompanyType = "BRANCH"
)

type Company struct {
	ID          uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CompanyCode string      `gorm:"type:varchar(20);uniqueIndex;not null" json:"code"` // e.g. "HLD-001", "SUB-001"
	CompanyName string      `gorm:"type:varchar(150);not null" json:"name"`
	LegalName   string      `gorm:"type:varchar(200)" json:"legal_name"`        // official registered name
	TaxID       string      `gorm:"type:varchar(50);uniqueIndex" json:"tax_id"` // NPWP
	CompanyType CompanyType `gorm:"type:varchar(20);not null" json:"company_type"`

	// Self-referencing for holding → subsidiary → branch hierarchy
	ParentID *uuid.UUID `gorm:"type:uuid;index" json:"parent_id"`
	Parent   *Company   `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Company  `gorm:"foreignKey:ParentID" json:"children,omitempty"`

	// FK to Location (HQ / registered office)
	LocationID *uuid.UUID `gorm:"type:uuid;index" json:"location_id"`
	Location   *Location  `gorm:"foreignKey:LocationID" json:"location,omitempty"`

	Phone         string         `gorm:"type:varchar(20)" json:"phone"`
	Email         string         `gorm:"type:varchar(100)" json:"email"`
	Website       string         `gorm:"type:varchar(200)" json:"website"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	EstablishedAt *time.Time     `json:"established_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
