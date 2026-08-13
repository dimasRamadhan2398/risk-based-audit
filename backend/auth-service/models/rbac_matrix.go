package models

import (
	"time"

	"github.com/google/uuid"
)

// RbacMatrixFeature represents an individual module feature item in the
// AuditSphere Role-Based Access Control (RBAC) Matrix.
//
// This model maps to the database table `rbac_matrix_features` and establishes
// the explicit permission levels for each user role (Admin/CAE, Audit Manager,
// Auditor, Auditee/Dept Head, Viewer) across all system modules.
type RbacMatrixFeature struct {
	ID                 uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FeatureNumber      int       `gorm:"not null" json:"feature_number"`                          // Module sequence number (#1 - #15)
	Module             string    `gorm:"type:varchar(100);not null" json:"module"`                 // Main module name (e.g. "Risk Profile")
	Submodule          string    `gorm:"type:varchar(150)" json:"submodule"`                       // Submodule title (e.g. "Corporate Risk Profile")
	FeatureCode        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"feature_code"` // Unique key identifier (e.g. "risk_profile_corporate")
	Description        string    `gorm:"type:text" json:"description"`                             // Operational description of feature
	
	// Access rights per role according to AuditSphere RBAC Document Matrix
	// Access levels: FULL, READ, NONE, LIMITED, DRAFT, EDIT_OWN_DEPT, RESPOND, REVIEW, CREATE, UPLOAD, FILL_SURVEY, UPDATE_CAPA, REQUEST
	AdminAccess        string    `gorm:"type:varchar(50);not null;default:'FULL'" json:"admin_access"`
	AuditManagerAccess string    `gorm:"type:varchar(50);not null;default:'READ'" json:"audit_manager_access"`
	AuditorAccess      string    `gorm:"type:varchar(50);not null;default:'READ'" json:"auditor_access"`
	AuditeeAccess      string    `gorm:"type:varchar(50);not null;default:'NONE'" json:"auditee_access"`
	ViewerAccess       string    `gorm:"type:varchar(50);not null;default:'READ'" json:"viewer_access"`

	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// TableName specifies the database table name for GORM
func (RbacMatrixFeature) TableName() string {
	return "rbac_matrix_features"
}
