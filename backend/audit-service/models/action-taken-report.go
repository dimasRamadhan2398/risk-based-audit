package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ActionTakenReport struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AuditFindingID      *uuid.UUID     `gorm:"type:uuid;index" json:"audit_finding_id"`
	AuditFinding        *AuditFinding  `gorm:"foreignKey:AuditFindingID" json:"audit_finding"`
	AuditRef            string         `gorm:"type:varchar(100);not null" json:"auditRef"`
	Title               string         `gorm:"type:varchar(255);not null" json:"title"`
	Department          string         `gorm:"type:varchar(100)" json:"department"`
	AuditObject         string         `gorm:"type:varchar(255)" json:"auditObject"`
	FindingCategory     string         `gorm:"type:varchar(100)" json:"findingCategory"`
	Condition           string         `gorm:"type:text" json:"condition"`
	Criteria            string         `gorm:"type:text" json:"criteria"`
	Recommendation      string         `gorm:"type:text" json:"recommendation"`
	PIC                 string         `gorm:"type:varchar(200)" json:"pic"`
	Deadline            string         `gorm:"type:varchar(100)" json:"deadline"`
	Status              string         `gorm:"type:varchar(50);default:'PLANNED'" json:"status"`
	Attachment          string         `gorm:"type:varchar(500)" json:"attachment"`
	ProgressDescription string         `gorm:"type:text" json:"progressDescription"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}
