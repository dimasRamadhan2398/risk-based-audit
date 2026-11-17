package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Employee struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	EmployeeCode string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"employee_code"` // e.g. "EMP-0001"
	FullName     string    `gorm:"type:varchar(100);not null" json:"full_name"`
	Email        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Phone        string    `gorm:"type:varchar(20)" json:"phone"`

	CompanyID uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	Company   Company   `gorm:"foreignKey:CompanyID" json:"company"`
	// FK to master data
	DepartmentID uuid.UUID  `gorm:"type:uuid;not null;index" json:"department_id"`
	Department   Department `gorm:"foreignKey:DepartmentID" json:"department"`

	JobRoleID uuid.UUID `gorm:"type:uuid;not null;index" json:"job_role_id"`
	JobRole   JobRole   `gorm:"foreignKey:JobRoleID" json:"job_role"`

	// LevelGrade is no longer part of the API: never sent, and level_grade in a
	// request body is ignored. The column stays (INT NOT NULL); the service
	// stores 1 when it is unset, so inserts keep working.
	LevelGrade int `gorm:"type:int;not null" json:"-"`

	WorkLocationID *uuid.UUID `gorm:"type:uuid;index" json:"work_location_id"`
	WorkLocation   *Location  `gorm:"foreignKey:WorkLocationID" json:"work_location,omitempty"`

	ResidenceAddress  string `gorm:"type:text" json:"residence_address"`
	ResidenceCity     string `gorm:"type:varchar(100)" json:"residence_city"`
	ResidenceProvince string `gorm:"type:varchar(100)" json:"residence_province"`
	ResidencePostal   string `gorm:"type:varchar(20)" json:"residence_postal_code"`
	// Direct manager reference (self-referencing)
	ManagerID *uuid.UUID `gorm:"type:uuid;index" json:"manager_id"`
	Manager   *Employee  `gorm:"foreignKey:ManagerID" json:"manager,omitempty"`

	IsActive  bool           `gorm:"default:true" json:"is_active"`
	JoinDate  time.Time      `json:"join_date"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Account describes the login account provisioned while this employee was
	// created. It is set only on the create response and is never stored.
	Account *EmployeeAccount `gorm:"-" json:"account,omitempty"`

	// AccessRole is an optional request-only override of the login role given to
	// the new account on create. Empty = use the job role's default_access_role.
	// Never stored.
	AccessRole string `gorm:"-" json:"access_role,omitempty"`
}

// EmployeeAccount is the login account linked to an employee at creation time.
// Status is "created" (new account, TemporaryPassword set once) or "linked"
// (an account with the same email already existed; its password is unchanged).
//
// SeatDowngraded is true when auth-service created the account with VIEWER
// only because every full seat was taken; RequestedRoles and AssignedRoles
// then differ. Both role lists are set only for status "created".
type EmployeeAccount struct {
	Status            string   `json:"status"`
	Username          string   `json:"username"`
	TemporaryPassword string   `json:"temporary_password,omitempty"`
	SeatDowngraded    bool     `json:"seat_downgraded"`
	RequestedRoles    []string `json:"requested_roles,omitempty"`
	AssignedRoles     []string `json:"assigned_roles,omitempty"`
}

// EmployeeReference is one place that still points at an employee, returned in
// data.references of a 409 EMPLOYEE_IN_USE. Count includes soft-deleted rows:
// the foreign keys still hold them.
type EmployeeReference struct {
	Table  string `json:"table"`
	Column string `json:"column"`
	Label  string `json:"label"`
	Count  int64  `json:"count"`
}
