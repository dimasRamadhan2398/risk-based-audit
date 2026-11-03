package models

import (
	"strings"
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
	// DefaultAccessRole is the auth-service role a new login account gets for an
	// employee holding this job role. Empty = no mapping (provisioning falls
	// back to DefaultLoginAccessRole). Only values from AllowedAccessRoles.
	DefaultAccessRole string         `gorm:"type:varchar(50)" json:"default_access_role"`
	IsActive          bool           `gorm:"default:true" json:"is_active"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
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

// Access roles a job role may map to: names of roles seeded in auth-service.
// ADMIN is deliberately absent: a job title must never grant administrator
// rights. AUDIT_MANAGER / CHIEF_AUDIT_EXECUTIVE only exist if an admin creates
// them in auth-service, so they cannot be offered here.
const (
	AccessRoleAuditor        = "AUDITOR"
	AccessRoleExecutive      = "EXECUTIVE"
	AccessRoleAuditee        = "AUDITEE"
	AccessRoleViewer         = "VIEWER"
	AccessRoleDepartmentHead = "DEPARTMENT_HEAD"

	// DefaultLoginAccessRole is used when a job role has no mapping.
	DefaultLoginAccessRole = AccessRoleAuditee
)

// AllowedAccessRoles returns the valid DefaultAccessRole values (a fresh slice).
func AllowedAccessRoles() []string {
	return []string{AccessRoleAuditor, AccessRoleExecutive, AccessRoleAuditee, AccessRoleViewer, AccessRoleDepartmentHead}
}

// NormalizeAccessRole trims and upper-cases v and reports whether it is allowed.
// An empty value is valid and means "no mapping".
func NormalizeAccessRole(v string) (string, bool) {
	n := strings.ToUpper(strings.TrimSpace(v))
	if n == "" {
		return "", true
	}
	for _, r := range AllowedAccessRoles() {
		if r == n {
			return n, true
		}
	}
	return n, false
}

// AccessRoleForPosition is the default mapping used to backfill job roles that
// have none. Ambiguous positions (manager, supervisor, committee, external)
// return "" and stay unmapped.
func AccessRoleForPosition(p JobPositionType) string {
	switch p {
	case PositionTypeSeniorManager, PositionTypeGeneralManager, PositionTypeDirector, PositionTypeCommissioner:
		return AccessRoleExecutive
	case PositionTypeChiefAuditExec:
		return AccessRoleAuditor
	case PositionTypeStaff:
		return AccessRoleAuditee
	}
	return ""
}
