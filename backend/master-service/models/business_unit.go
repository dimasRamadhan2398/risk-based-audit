package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BusinessUnit struct {
	ID 		uuid.UUID 		`gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	BusinessUnitCode 	string 	`gorm:"type:varchar(50);uniqueIndex;not null" json:"business_unit_code"`
	BusinessUnitName 	string 	`gorm:"type:varchar(100);uniqueIndex;not null" json:"business_unit_name"`
	BusinessUnitDescription string 	`gorm:"type:text" json:"business_unit_description"`
	CostCenterCode	string 	`gorm:"type:varchar(50);uniqueIndex;not null" json:"cost_center_code"`
	ParentBusinessUnitID *uuid.UUID 	`gorm:"type:uuid;index" json:"parent_business_unit_id"`
	ParentBusinessUnit      *BusinessUnit  `gorm:"foreignKey:ParentBusinessUnitID" json:"parent_business_unit,omitempty"`
	IsActive 	bool 		`gorm:"default:true" json:"is_active"`
	CreatedAt time.Time 	`json:"created_at"`
	UpdatedAt time.Time 	`json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	CompanyID uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	Company Company `gorm:"foreignKey:CompanyID" json:"company"`

	Departments 	[]Department `gorm:"foreignKey:BusinessUnitID" json:"departments,omitempty"`
}