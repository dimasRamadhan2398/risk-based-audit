package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	AuditStatusPlanning      = "planning"
	AuditStatusEntryMeeting  = "entry meeting"
	AuditStatusFieldwork     = "fieldwork"
	AuditStatusDraftFindings = "draft findings"
	AuditStatusReporting     = "reporting"
	AuditStatusCompleted     = "completed"
)

// MapProgressToStatus maps progress percentage (0-100) to its corresponding status and status detail
func MapProgressToStatus(progress int) (string, string) {
	switch {
	case progress >= 100:
		return AuditStatusCompleted, "Audit Completed"
	case progress >= 76:
		return AuditStatusReporting, "Reporting & Exit Meeting"
	case progress >= 51:
		return AuditStatusDraftFindings, "Draft Findings & Recommendations"
	case progress >= 26:
		return AuditStatusFieldwork, "Fieldwork & Control Testing"
	case progress >= 1:
		return AuditStatusEntryMeeting, "Entry Meeting & Scope Alignment"
	default:
		return AuditStatusPlanning, "Planning & Preparation"
	}
}

// NormalizeExecutionStatus normalizes any status string into one of the 6 canonical stages
func NormalizeExecutionStatus(status string, progress int) (string, string) {
	s := strings.ToLower(strings.TrimSpace(status))
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")

	switch s {
	case "planning", "planned", "perencanaan", "preparation":
		return AuditStatusPlanning, "Planning & Preparation"
	case "entry meeting", "entry", "scope alignment":
		return AuditStatusEntryMeeting, "Entry Meeting & Scope Alignment"
	case "fieldwork", "testing", "field work":
		return AuditStatusFieldwork, "Fieldwork & Control Testing"
	case "draft findings", "draft finding", "draft temuan", "findings":
		return AuditStatusDraftFindings, "Draft Findings & Recommendations"
	case "reporting", "report", "pelaporan", "exit meeting":
		return AuditStatusReporting, "Reporting & Exit Meeting"
	case "completed", "complete", "selesai", "done", "finished":
		return AuditStatusCompleted, "Audit Completed"
	default:
		return MapProgressToStatus(progress)
	}
}

type TestControlsSub struct {
	Progress    int    `json:"progress"`
	Description string `json:"description"`
}

type WorkingPapersSub struct {
	Condition string `json:"condition"`
	Criteria  string `json:"criteria"`
}

type ImprovementsSub struct {
	Recommendation string `json:"recommendation"`
	Deadline       string `json:"deadline"`
	PIC            string `json:"pic"`
}

type LatestUpdateSub struct {
	Attachment  string `json:"attachment"`
	Description string `json:"description"`
}

type AuditExecution struct {
	ID                     uuid.UUID         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ActivityID             *uuid.UUID        `gorm:"type:uuid;index" json:"activity_id,omitempty"`
	Ref                    string            `gorm:"type:varchar(100)" json:"ref"`
	Name                   string            `gorm:"type:varchar(255)" json:"name"`
	Category               string            `gorm:"type:varchar(100)" json:"category"`
	Progress               int               `gorm:"type:int" json:"progress"`
	LeadAuditor            string            `gorm:"type:varchar(200)" json:"lead_auditor"`
	Status                 string            `gorm:"type:varchar(50)" json:"status"`
	StatusDetail           string            `gorm:"type:varchar(100)" json:"status_detail"`
	SampleDataTestControls *TestControlsSub  `gorm:"serializer:json" json:"sample_data_test_controls,omitempty"`
	WorkingPapers          *WorkingPapersSub `gorm:"serializer:json" json:"working_papers,omitempty"`
	ActionPlanImprovements *ImprovementsSub  `gorm:"serializer:json" json:"action_plan_improvements,omitempty"`
	LatestUpdateProgress   *LatestUpdateSub  `gorm:"serializer:json" json:"latest_update_progress,omitempty"`
	CreatedAt              time.Time         `json:"created_at"`
	UpdatedAt              time.Time         `json:"updated_at"`
	DeletedAt              gorm.DeletedAt    `gorm:"index" json:"-"`
}

func (AuditExecution) TableName() string {
	return "audit_executions"
}

func (a *AuditExecution) BeforeSave(tx *gorm.DB) error {
	normalizedStatus, defaultDetail := NormalizeExecutionStatus(a.Status, a.Progress)
	a.Status = normalizedStatus
	if a.StatusDetail == "" || strings.EqualFold(a.StatusDetail, "in progress") || strings.EqualFold(a.StatusDetail, "on time") {
		a.StatusDetail = defaultDetail
	}
	return nil
}
