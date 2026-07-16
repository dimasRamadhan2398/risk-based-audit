package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Department struct {
	ID                    uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	DepartmentCode        string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"department_code"`
	DepartmentName        string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"department_name"`
	DepartmentDescription string         `gorm:"type:text" json:"department_description"`
	PicID                 uuid.UUID      `gorm:"type:uuid;not null;index" json:"pic_id"`
	Level                 int            `gorm:"type:int;not null" json:"level"`
	IsActive              bool           `gorm:"default:true" json:"is_active"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`

	CompanyID uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	Company   Company   `gorm:"foreignKey:CompanyID" json:"company"`

	BusinessUnitID uuid.UUID    `gorm:"type:uuid;not null;index" json:"business_unit_id"`
	BusinessUnit   BusinessUnit `gorm:"foreignKey:BusinessUnitID" json:"business_unit,omitempty"`
}
