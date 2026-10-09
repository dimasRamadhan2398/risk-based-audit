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
	SignaturePlace     string               `gorm:"type:varchar(100)" json:"signaturePlace"`
	SignatureDate      *time.Time           `json:"signatureDate"`
	Signatures         []ReportSignature    `gorm:"serializer:json" json:"signatures"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
	DeletedAt          gorm.DeletedAt       `gorm:"index" json:"-"`
}

type ReportSignature struct {
	Name      string `json:"name"`
	Role      string `json:"role"`
	Signature string `json:"signature"`
	SignedAt  string `json:"signedAt,omitempty"`
}

type AuditReportFinding struct {
	Title    string `json:"title"`
	Category string `json:"category"`
	Action   string `json:"action"`
}

func (r *AuditResultReport) UnmarshalJSON(data []byte) error {
	type Alias AuditResultReport
	aux := struct {
		ReportDate        *string `json:"report_date"`
		ReportDateAlt     *string `json:"reportDate"`
		CompanyID         *string `json:"company_id"`
		CompanyIDAlt      *string `json:"companyId"`
		SignaturePlace    *string `json:"signaturePlace"`
		SignaturePlaceAlt *string `json:"signature_place"`
		SignatureDate     *string `json:"signatureDate"`
		SignatureDateAlt  *string `json:"signature_date"`
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

	targetSigPlace := aux.SignaturePlace
	if targetSigPlace == nil {
		targetSigPlace = aux.SignaturePlaceAlt
	}
	if targetSigPlace != nil {
		r.SignaturePlace = strings.TrimSpace(*targetSigPlace)
	}

	targetSigDate := aux.SignatureDate
	if targetSigDate == nil {
		targetSigDate = aux.SignatureDateAlt
	}
	if targetSigDate != nil {
		dateStr := strings.TrimSpace(*targetSigDate)
		if dateStr == "" || dateStr == "null" {
			r.SignatureDate = nil
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
					r.SignatureDate = &t
					parsed = true
					break
				}
			}
			if !parsed && len(dateStr) >= 10 {
				if t, err := time.Parse("2006-01-02", dateStr[:10]); err == nil {
					r.SignatureDate = &t
					parsed = true
				}
			}
			if !parsed {
				r.SignatureDate = nil
			}
		}
	}

	return nil
}

func normalizeLHAReportNumber(raw string, team string, defaultYear int) string {
	raw = strings.TrimSpace(raw)
	if strings.Contains(raw, "/LHA/") {
		parts := strings.Split(raw, "/")
		if len(parts) >= 2 {
			seq := strings.TrimSpace(parts[0])
			year := fmt.Sprintf("%d", defaultYear)
			if len(parts) >= 5 && strings.TrimSpace(parts[4]) != "" {
				year = strings.TrimSpace(parts[4])
			}
			if team == "" {
				team = "SKAI"
			}
			return fmt.Sprintf("LHA-%s/%s/%s", seq, strings.ToUpper(team), year)
		}
	}
	return raw
}

func (r *AuditResultReport) BeforeCreate(tx *gorm.DB) error {
	if strings.Contains(r.ReportNumber, "/LHA/") {
		year := time.Now().Year()
		if r.ReportDate != nil {
			year = r.ReportDate.Year()
		}
		team := "SKAI"
		if r.AssignmentLetterID != "" {
			var al AssignmentLetter
			if err := tx.Model(&AssignmentLetter{}).Where("letter_number = ? OR id::text = ?", r.AssignmentLetterID, r.AssignmentLetterID).First(&al).Error; err == nil && strings.TrimSpace(al.AuditTeam) != "" {
				team = strings.TrimSpace(al.AuditTeam)
			}
		}
		r.ReportNumber = normalizeLHAReportNumber(r.ReportNumber, team, year)
		return nil
	}

	if strings.TrimSpace(r.ReportNumber) == "" {
		year := time.Now().Year()
		if r.ReportDate != nil {
			year = r.ReportDate.Year()
		}

		team := "SKAI"
		if r.AssignmentLetterID != "" {
			var al AssignmentLetter
			if err := tx.Model(&AssignmentLetter{}).Where("letter_number = ? OR id::text = ?", r.AssignmentLetterID, r.AssignmentLetterID).First(&al).Error; err == nil && strings.TrimSpace(al.AuditTeam) != "" {
				team = strings.TrimSpace(al.AuditTeam)
			} else {
				parts := strings.Split(r.AssignmentLetterID, "/")
				if len(parts) >= 2 && strings.TrimSpace(parts[1]) != "" {
					team = strings.TrimSpace(parts[1])
				}
			}
		}

		var reports []AuditResultReport
		tx.Model(&AuditResultReport{}).Unscoped().Select("report_number").Find(&reports)

		maxSeq := 20
		for _, rep := range reports {
			numStr := strings.TrimSpace(rep.ReportNumber)
			if strings.HasPrefix(strings.ToUpper(numStr), "LHA-") {
				parts := strings.Split(numStr, "/")
				if len(parts) > 0 {
					var seq int
					numPart := strings.TrimPrefix(strings.ToUpper(parts[0]), "LHA-")
					if _, err := fmt.Sscanf(numPart, "%d", &seq); err == nil && seq > maxSeq {
						maxSeq = seq
					}
				}
			} else if strings.Contains(numStr, "/LHA/") {
				parts := strings.Split(numStr, "/")
				if len(parts) > 0 {
					var seq int
					if _, err := fmt.Sscanf(parts[0], "%d", &seq); err == nil && seq > maxSeq {
						maxSeq = seq
					}
				}
			}
		}
		r.ReportNumber = fmt.Sprintf("LHA-%03d/%s/%d", maxSeq+1, strings.ToUpper(team), year)
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
		if repNum, present := dest["report_number"]; present {
			if strVal, ok := repNum.(string); ok && strings.Contains(strVal, "/LHA/") {
				tx.Statement.SetColumn("report_number", normalizeLHAReportNumber(strVal, "SKAI", time.Now().Year()))
			}
		}
		return nil
	}
	r.FindingsCount = len(r.Findings)
	if strings.Contains(r.ReportNumber, "/LHA/") {
		r.ReportNumber = normalizeLHAReportNumber(r.ReportNumber, "SKAI", time.Now().Year())
	}
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
