package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ActionTakenReport (ATR) tracks the follow-up of ONE finding of an approved
// audit result report (LHA). Rows are created by the server when the LHA is
// approved (services/action_taken_report.SyncFromReport), one per finding,
// unique on (audit_result_report_id, finding_id).
//
// The report and finding fields are copies taken when the ATR is created
// (report number/title are refreshed when the LHA is saved), so an ATR stays
// readable if its LHA or finding is later removed. Rows migrated from the old
// assignment-letter based table may have no LHA link (both ids NULL).
type ActionTakenReport struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`

	// Link to the LHA and the finding inside its findings JSON array
	AuditResultReportID *uuid.UUID         `gorm:"type:uuid;index;uniqueIndex:idx_atr_report_finding,priority:1" json:"audit_result_report_id"`
	AuditResultReport   *AuditResultReport `gorm:"foreignKey:AuditResultReportID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"-"`
	FindingID           string             `gorm:"type:varchar(64);uniqueIndex:idx_atr_report_finding,priority:2" json:"finding_id"`

	// Copies of the LHA / finding
	ReportNumber    string `gorm:"type:varchar(100)" json:"report_number"`
	ReportTitle     string `gorm:"type:varchar(255)" json:"report_title"`
	FindingTitle    string `gorm:"type:text" json:"finding_title"`
	FindingCategory string `gorm:"type:varchar(100)" json:"finding_category"`
	Recommendation  string `gorm:"type:text" json:"recommendation"`

	// Action plan, owned by the PIC (an auth-service user id)
	ActionPlan string     `gorm:"type:text" json:"action_plan"`
	PICUserID  string     `gorm:"column:pic_user_id;type:varchar(64);index" json:"pic_user_id"`
	PICName    string     `gorm:"column:pic_name;type:varchar(200)" json:"pic_name"`
	DueDate    *time.Time `json:"due_date"`
	Progress   int        `gorm:"not null;default:0" json:"progress"`
	Status     string     `gorm:"type:varchar(50);default:'PLANNED';index" json:"status"`
	// PlannedActions are the agreed action plans (Rencana Aksi) copied from the
	// LHA finding's actionPlans: the LHA is the source of truth, the ATR
	// monitors their execution. FinalApprover is the first plan's final
	// approver. Both are refreshed from the LHA while the ATR is still PLANNED
	// and the PIC has not submitted a plan.
	PlannedActions []FindingActionPlan `gorm:"type:text;serializer:json" json:"planned_actions"`
	FinalApprover  string              `gorm:"type:varchar(200)" json:"final_approver"`
	// ActionPlanSubmittedAt is set once, the first time the PIC saves a
	// non-empty action plan (PUT /action-taken-reports/:id/action-plan).
	// NULL on history; the COSO management-response signal skips such rows.
	ActionPlanSubmittedAt *time.Time `json:"action_plan_submitted_at"`

	// Auditor review (also holds the reason of a cancellation)
	ReviewNote string     `gorm:"type:text" json:"review_note"`
	ReviewedBy string     `gorm:"type:varchar(200)" json:"reviewed_by"`
	ReviewedAt *time.Time `json:"reviewed_at"`

	Evidence []ActionTakenReportEvidence `gorm:"foreignKey:ActionTakenReportID;constraint:OnDelete:CASCADE" json:"evidence"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Derived on load/save from DueDate and Status (see action-taken-report-status.go);
	// never stored
	IsOverdue   bool `gorm:"-" json:"is_overdue"`
	OverdueDays int  `gorm:"-" json:"overdue_days"`
}

// BeforeCreate assigns the id in Go so rows can be created without the
// Postgres gen_random_uuid() default (and the id is known before INSERT)
func (r *ActionTakenReport) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// ActionTakenReportEvidence is a file the PIC uploaded for an ATR. The file is
// stored under the evidence directory as StoredName ("<uuid>-<safe name>") and
// is only served through the authenticated download endpoint.
type ActionTakenReportEvidence struct {
	ID                  uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ActionTakenReportID uuid.UUID `gorm:"type:uuid;not null;index" json:"action_taken_report_id"`
	FileName            string    `gorm:"type:varchar(255);not null" json:"file_name"`
	StoredName          string    `gorm:"type:varchar(320);not null" json:"-"`
	ContentType         string    `gorm:"type:varchar(150)" json:"-"`
	FileSize            int64     `json:"file_size"`
	UploadedBy          string    `gorm:"type:varchar(200)" json:"uploaded_by"`
	UploadedByUserID    string    `gorm:"type:varchar(64)" json:"-"`
	UploadedAt          time.Time `json:"uploaded_at"`
}

// TableName keeps the table name readable (GORM would pluralise "evidence")
func (ActionTakenReportEvidence) TableName() string {
	return "action_taken_report_evidences"
}

func (e *ActionTakenReportEvidence) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}
