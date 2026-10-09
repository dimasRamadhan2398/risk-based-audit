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
	ID       string `json:"id"`
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
// whatever the client sent, and gives every finding a stable id (existing ids
// are preserved, see AssignFindingIDs). For map updates (crud.Update) the
// count and ids follow the "findings" value being written; an update that
// only sends findings_count is reset to the stored findings' length.
func (r *AuditResultReport) BeforeSave(tx *gorm.DB) error {
	if dest, ok := tx.Statement.Dest.(map[string]interface{}); ok {
		if v, present := dest["findings"]; present {
			if v == nil {
				tx.Statement.SetColumn("findings_count", 0)
			} else {
				incoming, err := decodeFindings(v)
				if err != nil {
					return err
				}
				stored := r.Findings
				if r.ID != uuid.Nil {
					if s, ok := loadStoredFindings(tx, r.ID); ok {
						stored = s
					}
				}
				merged := AssignFindingIDs(incoming, stored)
				encoded, err := json.Marshal(merged)
				if err != nil {
					return err
				}
				tx.Statement.SetColumn("findings", string(encoded))
				tx.Statement.SetColumn("findings_count", len(merged))
			}
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
	var stored []AuditReportFinding
	if r.ID != uuid.Nil {
		stored, _ = loadStoredFindings(tx, r.ID)
	}
	if r.Findings != nil {
		r.Findings = AssignFindingIDs(r.Findings, stored)
	}
	r.FindingsCount = len(r.Findings)
	if strings.Contains(r.ReportNumber, "/LHA/") {
		r.ReportNumber = normalizeLHAReportNumber(r.ReportNumber, "SKAI", time.Now().Year())
	}
	return nil
}

// LHAApprovedStatuses are the AuditResultReport.Status values (compared
// case-insensitively) that mean the report is approved. The LHA form sends
// "Final"; seeded and older reports use "APPROVED" or "Published".
var LHAApprovedStatuses = []string{"FINAL", "APPROVED", "PUBLISHED"}

// IsLHAApproved reports whether an AuditResultReport status means approved
func IsLHAApproved(status string) bool {
	s := strings.ToUpper(strings.TrimSpace(status))
	for _, a := range LHAApprovedStatuses {
		if s == a {
			return true
		}
	}
	return false
}

// AssignFindingIDs returns incoming with a stable id on every finding:
//   - a finding that carries a valid uuid keeps it (first occurrence wins when
//     the same id is sent twice);
//   - a finding without one takes the id of an unclaimed stored finding with
//     the same title (case/space-insensitive), so clients that drop the id
//     on save do not orphan the finding's action taken report;
//   - anything else gets a new uuid.
//
// Running it again on its own output changes nothing.
func AssignFindingIDs(incoming, stored []AuditReportFinding) []AuditReportFinding {
	if incoming == nil {
		return nil
	}
	out := make([]AuditReportFinding, len(incoming))
	copy(out, incoming)

	used := make(map[string]bool, len(out))
	pending := make([]int, 0)
	for i := range out {
		id, ok := canonicalFindingID(out[i].ID)
		if ok && !used[id] {
			out[i].ID = id
			used[id] = true
			continue
		}
		out[i].ID = ""
		pending = append(pending, i)
	}

	byTitle := make(map[string][]string)
	for _, f := range stored {
		id, ok := canonicalFindingID(f.ID)
		if !ok {
			continue
		}
		key := findingTitleKey(f.Title)
		byTitle[key] = append(byTitle[key], id)
	}

	for _, i := range pending {
		key := findingTitleKey(out[i].Title)
		for len(byTitle[key]) > 0 {
			candidate := byTitle[key][0]
			byTitle[key] = byTitle[key][1:]
			if !used[candidate] {
				out[i].ID = candidate
				break
			}
		}
		if out[i].ID == "" {
			out[i].ID = uuid.NewString()
		}
		used[out[i].ID] = true
	}
	return out
}

// FindingsNeedIDs reports whether any finding lacks a valid, unique id
func FindingsNeedIDs(findings []AuditReportFinding) bool {
	seen := make(map[string]bool, len(findings))
	for _, f := range findings {
		id, ok := canonicalFindingID(f.ID)
		if !ok || seen[id] || id != f.ID {
			return true
		}
		seen[id] = true
	}
	return false
}

func canonicalFindingID(raw string) (string, bool) {
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil || id == uuid.Nil {
		return "", false
	}
	return id.String(), true
}

func findingTitleKey(title string) string {
	return strings.ToLower(strings.Join(strings.Fields(title), " "))
}

// loadStoredFindings reads the findings currently stored for report id, in a
// fresh session so the statement being saved is not affected
func loadStoredFindings(tx *gorm.DB, id uuid.UUID) ([]AuditReportFinding, bool) {
	var stored struct {
		Findings []AuditReportFinding `gorm:"serializer:json"`
	}
	err := tx.Session(&gorm.Session{NewDB: true}).
		Table("audit_result_reports").
		Select("findings").
		Where("id = ?", id).
		Take(&stored).Error
	if err != nil {
		return nil, false
	}
	return stored.Findings, true
}

// decodeFindings reads a map-update "findings" value: a JSON array as
// []byte/string (what crud.Update produces), a decoded slice, or a typed slice
func decodeFindings(v interface{}) ([]AuditReportFinding, error) {
	var raw []byte
	switch val := v.(type) {
	case []AuditReportFinding:
		return val, nil
	case []byte:
		raw = val
	case string:
		raw = []byte(val)
	case json.RawMessage:
		raw = val
	default:
		b, err := json.Marshal(val)
		if err != nil {
			return nil, err
		}
		raw = b
	}
	if len(raw) == 0 || string(raw) == "null" {
		return []AuditReportFinding{}, nil
	}
	var out []AuditReportFinding
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("invalid findings: %w", err)
	}
	if out == nil {
		out = []AuditReportFinding{}
	}
	return out, nil
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
