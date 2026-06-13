package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ApprovalAction string

const (
	ApprovalActionSubmitted ApprovalAction = "SUBMITTED"
	ApprovalActionApproved  ApprovalAction = "APPROVED"
	ApprovalActionRejected  ApprovalAction = "REJECTED"
)

type RiskApprovalLog struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RiskID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"risk_id"`
	Risk       RiskRegister   `gorm:"foreignKey:RiskID" json:"risk"`
	ApproverID uuid.UUID      `gorm:"type:uuid;not null;index" json:"approver_id"`
	Action     ApprovalAction `gorm:"type:varchar(20);not null" json:"action"`
	Notes      string         `gorm:"type:text" json:"notes"`
	ActionDate time.Time      `json:"action_date"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RiskApprovalLog) TableName() string {
	return "risk_approval_log"
}
