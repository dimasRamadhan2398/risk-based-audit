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

// AnnualAuditActivity is one row of the Rencana Kegiatan Audit (the activity
// table of an annual plan). It is stored as JSON inside audit_annuals.activities,
// so new fields need no SQL migration.
//
// Every field is entered once. The template columns map as follows:
//
//	No                          itemNumber
//	Kode Penugasan              activityCode
//	Objek Penugasan Audit       name
//	Jenis Audit                 category
//	Audit Universe              auditUniverseId (picked; corporate id)
//	Justifikasi Pemilihan Objek justification
//	Tujuan Audit                auditObjective
//	Fokus Audit                 auditFocus
//	Jadwal Penugasan            timelineText
//	Jumlah Tim (Orang)          auditorCount
//	Total Mandays               totalMandays
//	Ketua Tim                   supervisorName
//
// ID is a server-assigned UUID that stays with the row across updates (see
// controllers/annual_plan). RiskName, RiskLevel and Department are display
// snapshots of AuditUniverseID, the yearly score and DepartmentID. They are the
// only fields that change without a user edit: services/annual_plan_sync
// refreshes them from the Audit Universe and Department events.
type AnnualAuditActivity struct {
	ID              string `json:"id"`
	ItemNumber      string `json:"itemNumber"`
	ActivityCode    string `json:"activityCode"`
	Name            string `json:"name"`
	Category        string `json:"category"`
	AuditUniverseID string `json:"auditUniverseId"`
	RiskName        string `json:"riskName"`
	RiskLevel       string `json:"riskLevel"`
	DepartmentID    string `json:"departmentId"`
	Department      string `json:"department"`
	Justification   string `json:"justification"`
	AuditObjective  string `json:"auditObjective"`
	AuditFocus      string `json:"auditFocus"`
	TimelineText    string `json:"timelineText"`
	AuditorCount    int    `json:"auditorCount"`
	TotalMandays    int    `json:"totalMandays"`
	SupervisorName  string `json:"supervisorName"`
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
	// Status is one of the AnnualPlanStatus* values (annual_plan_status.go). Changes
	// are checked by the update hook in controllers/annual_plan.
	Status               string                   `gorm:"type:varchar(50);default:'Draft'" json:"status"`
	SelectedMonths       []int                    `gorm:"serializer:json" json:"selectedMonths"`
	Quarters             []string                 `gorm:"serializer:json" json:"quarters"`
	AuditorCount         int                      `gorm:"type:int" json:"auditorCount"`
	DaysPerAuditor       int                      `gorm:"type:int" json:"daysPerAuditor"`
	TotalMandays         int                      `gorm:"type:int" json:"totalMandays"`
	SupervisorID         string                   `gorm:"type:varchar(50)" json:"supervisorId"`
	SupervisorName       string                   `gorm:"type:varchar(255)" json:"supervisorName"`
	Year                 int                      `gorm:"type:int;not null" json:"year"`
	// Title is the Judul Kegiatan Audit (plan header).
	Title                string                   `gorm:"type:varchar(255)" json:"title"`
	// ActivityType is the Jenis Kegiatan Audit (plan header). When set it must be
	// one of AnnualPlanActivityTypes (see controllers/annual_plan).
	ActivityType         string                   `gorm:"type:varchar(100)" json:"activityType"`
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