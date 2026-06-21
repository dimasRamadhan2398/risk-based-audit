package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QAReportAttachment struct {
	Name       string `json:"name"`
	Size       string `json:"size"`
	UploadedAt string `json:"uploadedAt"`
}

type QAReport struct {
	ID                uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Type              string              `gorm:"type:varchar(100);not null" json:"type"`
	Period            string              `gorm:"type:varchar(50);not null" json:"period"`
	ReportName        string              `gorm:"type:varchar(255);not null" json:"reportName"`
	Result            string              `gorm:"type:varchar(100);not null" json:"result"`
	Status            string              `gorm:"type:varchar(50);default:'Planned'" json:"status"`
	AssessmentTitle   string              `gorm:"type:varchar(255)" json:"assessmentTitle"`
	Validator         string              `gorm:"type:varchar(255)" json:"validator,omitempty"`
	InternalEvaluator string              `gorm:"type:varchar(255)" json:"internalEvaluator,omitempty"`
	Attachment        *QAReportAttachment `gorm:"serializer:json" json:"attachment,omitempty"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
	DeletedAt         gorm.DeletedAt      `gorm:"index" json:"-"`
}

func (QAReport) TableName() string {
	return "qa_reports"
}
