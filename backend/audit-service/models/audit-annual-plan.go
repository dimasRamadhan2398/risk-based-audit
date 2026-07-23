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

type AnnualAuditActivity struct {
	Name       string `json:"name"`
	Category   string `json:"category"`
	Department string `json:"department"`
	RiskName   string `json:"riskName"`
	RiskLevel  string `json:"riskLevel"`
}

type AnnualAuditAttachment struct {
	Name string `json:"name"`
	Size string `json:"size"`
	URL  string `json:"url"`
}

type AuditAnnual struct {
	ID                   uuid.UUID                `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Code                 string                   `gorm:"type:varchar(50)" json:"code"`
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