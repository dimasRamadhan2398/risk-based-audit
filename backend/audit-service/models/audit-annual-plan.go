package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RevisionHistoryEntry struct {
	Date    string `json:"date"`
	Version string `json:"version"`
	Changes string `json:"changes"`
	User    string `json:"user"`
}

// AnnualAuditActivity is one row of an annual plan. It is stored as JSON inside
// audit_annuals.activities, so new fields need no SQL migration.
//
// AuditUniverseID and DepartmentID are the stable links to risk-service's
// corporate_audit_universe.id and master-service's departments.id. RiskName and
// Department hold the names as last synced from those services (see
// services/annual_plan_sync). Plans saved before the links existed have only the
// names; the sync links them by name the first time a matching event arrives.
type AnnualAuditActivity struct {
	Name            string `json:"name"`
	Category        string `json:"category"`
	Department      string `json:"department"`
	DepartmentID    string `json:"departmentId,omitempty"`
	RiskName        string `json:"riskName"`
	RiskLevel       string `json:"riskLevel"`
	AuditUniverseID string `json:"auditUniverseId,omitempty"`
}

type AnnualAuditAttachment struct {
	Name string `json:"name"`
	Size string `json:"size"`
	URL  string `json:"url"`
}

type AuditAnnual struct {
	ID                   uuid.UUID                `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	// Code is the plan code, PKAT-{year}-{seq}, assigned by the server on create
	// (see pkg/activitycode NextPlan). Unique among live plans: soft-deleted
	// plans do not hold a code, but their numbers are never reissued.
	Code                 string                   `gorm:"type:varchar(50);uniqueIndex:idx_audit_annuals_code_live,where:deleted_at IS NULL AND code <> ''" json:"code"`
	Version              string                   `gorm:"type:varchar(50);default:'v1.0'" json:"version"`
	RevisionHistory      []RevisionHistoryEntry   `gorm:"serializer:json" json:"revisionHistory"`
	Activities           []AnnualAuditActivity    `gorm:"serializer:json" json:"activities"`
	Status               string                   `gorm:"type:varchar(50);default:'DRAFT'" json:"status"` // e.g., DRAFT, PENDING_APPROVAL, APPROVED, CLOSED
	SelectedMonths       []int                    `gorm:"serializer:json" json:"selectedMonths"`
	Quarters             []string                 `gorm:"serializer:json" json:"quarters"`
	AuditorCount         int                      `gorm:"type:int" json:"auditorCount"`
	DaysPerAuditor       int                      `gorm:"type:int" json:"daysPerAuditor"`
	TotalMandays         int                      `gorm:"type:int" json:"totalMandays"`
	SupervisorID         string                   `gorm:"type:varchar(50)" json:"supervisorId"`
	SupervisorName       string                   `gorm:"type:varchar(255)" json:"supervisorName"`
	Year                 int                      `gorm:"type:int;not null" json:"year"`
	Notes                string                   `gorm:"type:text" json:"notes"`
	AttachmentCategory   string                   `gorm:"type:varchar(100)" json:"attachmentCategory"`
	Attachments          []AnnualAuditAttachment  `gorm:"serializer:json" json:"attachments"`
	AttachmentUploadedBy string                   `gorm:"type:varchar(255)" json:"attachmentUploadedBy"`
	AttachmentUploadDate string                   `gorm:"type:varchar(100)" json:"attachmentUploadDate"`
	IsActive             bool                     `gorm:"type:boolean;default:true" json:"isActive"`

	StaffApprovalNote    string                   `gorm:"type:text;column:staff_approval_note" json:"staffApprovalNote"`
	ManagerApprovalNote  string                   `gorm:"type:text;column:manager_approval_note" json:"managerApprovalNote"`
	ChiefApprovalNote    string                   `gorm:"type:text;column:chief_approval_note" json:"chiefApprovalNote"`

	// Activity Plans linked to this Annual Plan
	ActivityPlans []AuditActivity `gorm:"foreignKey:AnnualPlanID" json:"activity_plans,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}