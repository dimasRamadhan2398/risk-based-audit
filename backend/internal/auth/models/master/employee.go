package master

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Employee struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
    EmployeeCode string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"employee_code"` // e.g. "EMP-0001"
    FullName     string         `gorm:"type:varchar(100);not null" json:"full_name"`
    Email        string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
    Phone        string         `gorm:"type:varchar(20)" json:"phone"`

	CompanyID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"company_id"`
    Company        Company        `gorm:"foreignKey:CompanyID" json:"company"`
    // FK to master data
    DepartmentID uuid.UUID      `gorm:"type:uuid;not null;index" json:"department_id"`
    Department   Department     `gorm:"foreignKey:DepartmentID" json:"department"`

    JobRoleID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"job_role_id"`
    JobRole      JobRole        `gorm:"foreignKey:JobRoleID" json:"job_role"`

	LevelGrade   int            `gorm:"type:int;not null" json:"level_grade"`

	WorkLocationID *uuid.UUID     `gorm:"type:uuid;index" json:"work_location_id"`
    WorkLocation   *Location      `gorm:"foreignKey:WorkLocationID" json:"work_location,omitempty"`

	ResidenceAddress  string      `gorm:"type:text" json:"residence_address"`
    ResidenceCity     string      `gorm:"type:varchar(100)" json:"residence_city"`
    ResidenceProvince string      `gorm:"type:varchar(100)" json:"residence_province"`
    ResidencePostal   string      `gorm:"type:varchar(20)" json:"residence_postal_code"`
    // Direct manager reference (self-referencing)
    ManagerID    *uuid.UUID     `gorm:"type:uuid;index" json:"manager_id"`
    Manager      *Employee      `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`



    IsActive     bool           `gorm:"default:true" json:"is_active"`
    JoinDate     time.Time      `json:"join_date"`
    CreatedAt    time.Time      `json:"created_at"`
    UpdatedAt    time.Time      `json:"updated_at"`
    DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}