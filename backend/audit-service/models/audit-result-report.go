package models

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditResultReport struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ActivityPlanID     *uuid.UUID     `gorm:"type:uuid;index" json:"activity_plan_id"`
	AssignmentLetterID string         `gorm:"type:varchar(100)" json:"assignmentLetterId"`
	ReportTitle        string         `gorm:"type:varchar(255)" json:"reportTitle"`
	FindingsCount      int            `gorm:"type:int" json:"findingsCount"`
	ReportNumber       string         `gorm:"type:varchar(100);index" json:"reportNumber"`
	Title              string         `gorm:"type:varchar(255)" json:"title"`
	AuditObject        string         `gorm:"type:varchar(255)" json:"audit_object"`
	Department         string         `gorm:"type:varchar(100)" json:"department"`
	AuditPeriod        string         `gorm:"type:varchar(100)" json:"audit_period"`
	ExecutiveSummary   string         `gorm:"type:text" json:"executive_summary"`
	Scope              string         `gorm:"type:text" json:"scope"`
	Methodology        string         `gorm:"type:text" json:"methodology"`
	FindingSummary     string         `gorm:"type:text" json:"finding_summary"`
	Recommendation     string         `gorm:"type:text" json:"recommendation"`
	Conclusion         string         `gorm:"type:text" json:"conclusion"`
	PreparedBy         string         `gorm:"type:varchar(200)" json:"prepared_by"`
	ReviewedBy         string         `gorm:"type:varchar(200)" json:"reviewed_by"`
	ApprovedBy         string         `gorm:"type:varchar(200)" json:"approved_by"`
	ReportDate         *time.Time     `json:"report_date"`
	Status             string         `gorm:"type:varchar(50);default:'DRAFT'" json:"status"`
	Attachment         string         `gorm:"type:varchar(500)" json:"attachment"`
	Findings           []AuditReportFinding `gorm:"serializer:json" json:"findings"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

type AuditReportFinding struct {
	Title    string `json:"title"`
	Category string `json:"category"`
	Action   string `json:"action"`
}

func (r *AuditResultReport) UnmarshalJSON(data []byte) error {
	type Alias AuditResultReport
	aux := struct {
		ReportDate    *string `json:"report_date"`
		ReportDateAlt *string `json:"reportDate"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	targetDate := aux.ReportDate
	if targetDate == nil {
		targetDate = aux.ReportDateAlt
	}

	if targetDate != nil {
		dateStr := strings.TrimSpace(*targetDate)
		if dateStr == "" || dateStr == "null" {
			r.ReportDate = nil
		} else {
			formats := []string{
				time.RFC3339,
				"2006-01-02T15:04:05Z07:00",
				"2006-01-02T15:04:05",
				"2006-01-02 15:04:05",
				"2006-01-02",
			}
			var parsed bool
			for _, f := range formats {
				if t, err := time.Parse(f, dateStr); err == nil {
					r.ReportDate = &t
					parsed = true
					break
				}
			}
			if !parsed && len(dateStr) >= 10 {
				if t, err := time.Parse("2006-01-02", dateStr[:10]); err == nil {
					r.ReportDate = &t
					parsed = true
				}
			}
			if !parsed {
				r.ReportDate = nil
			}
		}
	} else {
		r.ReportDate = nil
	}

	return nil
}

