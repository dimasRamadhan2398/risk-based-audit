package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type JobRole struct {
	ID                 uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	JobRoleCode        string          `gorm:"type:varchar(50);uniqueIndex;not null" json:"job_role_code"`
	JobRoleName        string          `gorm:"type:varchar(100);uniqueIndex;not null" json:"job_role_name"`
	JobRoleDescription string          `gorm:"type:text" json:"job_role_description"`
	JobPositionType    JobPositionType `gorm:"type:varchar(100);not null" json:"job_position_type"`
	IsActive           bool            `gorm:"default:true" json:"is_active"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	DeletedAt          gorm.DeletedAt  `gorm:"index" json:"-"`
}

type JobPositionType string

const (
	PositionTypeStaff          JobPositionType = "STAFF"
	PositionTypeSupervisor     JobPositionType = "SUPERVISOR"
	PositionTypeManager        JobPositionType = "MANAGER"
	PositionTypeSeniorManager  JobPositionType = "SENIOR_MANAGER"
	PositionTypeGeneralManager JobPositionType = "GENERAL_MANAGER"
	PositionTypeDirector       JobPositionType = "DIRECTOR"
	PositionTypeCommissioner   JobPositionType = "COMMISSIONER"

	// Audit-specific structural positions
	PositionTypeAuditCommittee  JobPositionType = "AUDIT_COMMITTEE"
	PositionTypeChiefAuditExec  JobPositionType = "CHIEF_AUDIT_EXECUTIVE" // CAE
	PositionTypeExternalAuditor JobPositionType = "EXTERNAL_AUDITOR"
)
