package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditResultReport struct {
	ID                 uuid.UUID            `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ActivityPlanID     *uuid.UUID           `gorm:"type:uuid;index" json:"activity_plan_id"`
	AssignmentLetterID string               `gorm:"type:varchar(100)" json:"assignmentLetterId"`
	ReportTitle        string               `gorm:"type:varchar(255)" json:"reportTitle"`
	FindingsCount      int                  `gorm:"type:int" json:"findingsCount"`
	ReportNumber       string               `gorm:"type:varchar(100);index" json:"reportNumber"`
	Title              string               `gorm:"type:varchar(255)" json:"title"`
	AuditObject        string               `gorm:"type:varchar(255)" json:"audit_object"`
	Department         string               `gorm:"type:varchar(100)" json:"department"`
	CompanyID          *uuid.UUID           `gorm:"type:uuid;index" json:"company_id,omitempty"`
	CompanyName        string               `gorm:"type:varchar(255)" json:"company_name,omitempty"`
	AuditPeriod        string               `gorm:"type:varchar(100)" json:"audit_period"`
	ExecutiveSummary   string               `gorm:"type:text" json:"executive_summary"`
	Scope              string               `gorm:"type:text" json:"scope"`
	Methodology        string               `gorm:"type:text" json:"methodology"`
	FindingSummary     string               `gorm:"type:text" json:"finding_summary"`
	Recommendation     string               `gorm:"type:text" json:"recommendation"`
	Conclusion         string               `gorm:"type:text" json:"conclusion"`
	PreparedBy         string               `gorm:"type:varchar(200)" json:"prepared_by"`
	ReviewedBy         string               `gorm:"type:varchar(200)" json:"reviewed_by"`
	ApprovedBy         string               `gorm:"type:varchar(200)" json:"approved_by"`
	ReportDate         *time.Time           `json:"report_date"`
	Status             string               `gorm:"type:varchar(50);default:'DRAFT'" json:"status"`
	Attachment         string               `gorm:"type:varchar(500)" json:"attachment"`
	Findings           []AuditReportFinding `gorm:"serializer:json" json:"findings"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
	DeletedAt          gorm.DeletedAt       `gorm:"index" json:"-"`
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
		CompanyID     *string `json:"company_id"`
		CompanyIDAlt  *string `json:"companyId"`
		*Alias
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	targetCompID := aux.CompanyID
	if targetCompID == nil {
		targetCompID = aux.CompanyIDAlt
	}
	if targetCompID != nil {
		cidStr := strings.TrimSpace(*targetCompID)
		if cidStr == "" || cidStr == "null" {
			r.CompanyID = nil
		} else if parsedID, err := uuid.Parse(cidStr); err == nil {
			r.CompanyID = &parsedID
		} else {
			r.CompanyID = nil
		}
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

func (r *AuditResultReport) BeforeCreate(tx *gorm.DB) error {
	if strings.TrimSpace(r.ReportNumber) == "" {
		year := time.Now().Year()
		month := int(time.Now().Month())
		if r.ReportDate != nil {
			year = r.ReportDate.Year()
			month = int(r.ReportDate.Month())
		}

		var reports []AuditResultReport
		tx.Model(&AuditResultReport{}).Unscoped().Select("report_number").Find(&reports)

		maxSeq := 20
		prefix := "/LHA/"
		for _, rep := range reports {
			numStr := strings.TrimSpace(rep.ReportNumber)
			if strings.Contains(numStr, prefix) {
				parts := strings.Split(numStr, "/")
				if len(parts) > 0 {
					var seq int
					if _, err := fmt.Sscanf(parts[0], "%d", &seq); err == nil && seq > maxSeq {
						maxSeq = seq
					}
				}
			}
		}
		r.ReportNumber = fmt.Sprintf("%03d/LHA/%02d/KS IAD/%d", maxSeq+1, month, year)
	}
	return nil
}

// BeforeSave keeps FindingsCount equal to len(Findings) on every create/save,
// whatever the client sent. For map updates (crud.Update) the count follows
// the "findings" value being written; an update that only sends
// findings_count is reset to the stored findings' length.
func (r *AuditResultReport) BeforeSave(tx *gorm.DB) error {
	if dest, ok := tx.Statement.Dest.(map[string]interface{}); ok {
		if v, present := dest["findings"]; present {
			tx.Statement.SetColumn("findings_count", countFindings(v))
		} else if _, present := dest["findings_count"]; present {
			tx.Statement.SetColumn("findings_count", len(r.Findings))
		}
		return nil
	}
	r.FindingsCount = len(r.Findings)
	return nil
}

// countFindings counts the findings in a map-update value: a JSON array as
// []byte/string (what crud.Update produces), a decoded slice, or nil.
func countFindings(v interface{}) int {
	switch val := v.(type) {
	case nil:
		return 0
	case []AuditReportFinding:
		return len(val)
	case []interface{}:
		return len(val)
	case []byte:
		return countJSONArray(val)
	case string:
		return countJSONArray([]byte(val))
	case json.RawMessage:
		return countJSONArray(val)
	default:
		return 0
	}
}

func countJSONArray(b []byte) int {
	var items []json.RawMessage
	if err := json.Unmarshal(b, &items); err != nil {
		return 0
	}
	return len(items)
}
