package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditResultReport struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ActivityPlanID     *uuid.UUID     `gorm:"type:uuid;index" json:"activity_plan_id"`
	AssignmentLetterID string         `gorm:"type:varchar(100)" json:"assignmentLetterId"`
	ReportTitle        string         `gorm:"type:varchar(255)" json:"reportTitle"`
	OverallRating      string         `gorm:"type:varchar(100)" json:"overallRating"`
	FindingsCount      int            `gorm:"type:int" json:"findingsCount"`
	ReportNumber       string         `gorm:"type:varchar(100);index" json:"report_number"`
	Title              string         `gorm:"type:varchar(255)" json:"title"`
	AuditObject        string         `gorm:"type:varchar(255)" json:"audit_object"`
	Department      string         `gorm:"type:varchar(100)" json:"department"`
	AuditPeriod     string         `gorm:"type:varchar(100)" json:"audit_period"`
	ExecutiveSummary string        `gorm:"type:text" json:"executive_summary"`
	Scope           string         `gorm:"type:text" json:"scope"`
	Methodology     string         `gorm:"type:text" json:"methodology"`
	FindingSummary  string         `gorm:"type:text" json:"finding_summary"`
	Recommendation  string         `gorm:"type:text" json:"recommendation"`
	Conclusion      string         `gorm:"type:text" json:"conclusion"`
	PreparedBy      string         `gorm:"type:varchar(200)" json:"prepared_by"`
	ReviewedBy      string         `gorm:"type:varchar(200)" json:"reviewed_by"`
	ApprovedBy      string         `gorm:"type:varchar(200)" json:"approved_by"`
	ReportDate      *time.Time     `json:"report_date"`
	Status          string         `gorm:"type:varchar(50);default:'DRAFT'" json:"status"`
	Attachment      string         `gorm:"type:varchar(500)" json:"attachment"`
	Findings        []AuditReportFinding `gorm:"serializer:json" json:"findings"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

type AuditReportFinding struct {
	Title    string `json:"title"`
	Severity string `json:"severity"`
}
