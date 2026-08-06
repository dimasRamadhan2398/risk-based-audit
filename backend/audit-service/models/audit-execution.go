package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
