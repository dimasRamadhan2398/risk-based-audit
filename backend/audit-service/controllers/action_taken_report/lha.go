package actiontakenreport

import (
	"time"

	"audit-service/models"

	"github.com/google/uuid"
)

// LHAResponse is the LHA (audit result report) an ATR follows, merged into the
// ATR read endpoints. It is read live from audit_result_reports, so it always
// shows the current LHA; the ATR's own copies (report number/title, finding
// title/category/recommendation) stay as they are for list filters and history.
type LHAResponse struct {
	ID               string                   `json:"id"`
	ReportNumber     string                   `json:"report_number"`
	Title            string                   `json:"title"`
	ReportTitle      string                   `json:"report_title"`
	AuditObject      string                   `json:"audit_object"`
	Department       string                   `json:"department"`
	CompanyName      string                   `json:"company_name"`
	AuditPeriod      string                   `json:"audit_period"`
	ExecutiveSummary string                   `json:"executive_summary"`
	Scope            string                   `json:"scope"`
	Methodology      string                   `json:"methodology"`
	FindingSummary   string                   `json:"finding_summary"`
	Recommendation   string                   `json:"recommendation"`
	Conclusion       string                   `json:"conclusion"`
	PreparedBy       string                   `json:"prepared_by"`
	ReviewedBy       string                   `json:"reviewed_by"`
	ApprovedBy       string                   `json:"approved_by"`
	ReportDate       *string                  `json:"report_date"`
	Status           string                   `json:"status"`
	Attachment       string                   `json:"attachment"`
	FindingsCount    int                      `json:"findings_count"`
	Findings         []LHAFindingResponse     `json:"findings"`
	Finding          *LHAFindingResponse      `json:"finding"`
	SignaturePlace   string                   `json:"signature_place"`
	SignatureDate    *string                  `json:"signature_date"`
	Signatures       []models.ReportSignature `json:"signatures"`
	CreatedAt        time.Time                `json:"created_at"`
	UpdatedAt        time.Time                `json:"updated_at"`
}

// LHAFindingResponse is one finding of the LHA. Finding is the one this ATR
// follows (matched by finding_id); Findings lists all of them.
type LHAFindingResponse struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Category string `json:"category"`
	Action   string `json:"action"`
}

func newLHAFindings(in []models.AuditReportFinding) []LHAFindingResponse {
	out := make([]LHAFindingResponse, 0, len(in))
	for _, f := range in {
		out = append(out, LHAFindingResponse{ID: f.ID, Title: f.Title, Category: f.Category, Action: f.Action})
	}
	return out
}

// newLHAResponse builds the LHA block; nil when the LHA is missing
func newLHAResponse(l *models.AuditResultReport, findingID string) *LHAResponse {
	if l == nil {
		return nil
	}
	out := &LHAResponse{
		ID:               l.ID.String(),
		ReportNumber:     l.ReportNumber,
		Title:            l.Title,
		ReportTitle:      l.ReportTitle,
		AuditObject:      l.AuditObject,
		Department:       l.Department,
		CompanyName:      l.CompanyName,
		AuditPeriod:      l.AuditPeriod,
		ExecutiveSummary: l.ExecutiveSummary,
		Scope:            l.Scope,
		Methodology:      l.Methodology,
		FindingSummary:   l.FindingSummary,
		Recommendation:   l.Recommendation,
		Conclusion:       l.Conclusion,
		PreparedBy:       l.PreparedBy,
		ReviewedBy:       l.ReviewedBy,
		ApprovedBy:       l.ApprovedBy,
		Status:           l.Status,
		Attachment:       l.Attachment,
		FindingsCount:    len(l.Findings),
		Findings:         newLHAFindings(l.Findings),
		SignaturePlace:   l.SignaturePlace,
		Signatures:       l.Signatures,
		CreatedAt:        l.CreatedAt,
		UpdatedAt:        l.UpdatedAt,
	}
	if out.Signatures == nil {
		out.Signatures = []models.ReportSignature{}
	}
	if l.ReportDate != nil && !l.ReportDate.IsZero() {
		s := l.ReportDate.UTC().Format(time.RFC3339)
		out.ReportDate = &s
	}
	if l.SignatureDate != nil && !l.SignatureDate.IsZero() {
		s := l.SignatureDate.UTC().Format(time.RFC3339)
		out.SignatureDate = &s
	}
	for i := range out.Findings {
		if out.Findings[i].ID == findingID {
			f := out.Findings[i]
			out.Finding = &f
			break
		}
	}
	return out
}

// lhasFor loads the live LHAs the given ATRs follow, in one query. Soft-deleted
// LHAs are not returned (their ATRs get a null lha).
func (ctl *Controller) lhasFor(rows []models.ActionTakenReport) (map[uuid.UUID]*models.AuditResultReport, error) {
	ids := make([]uuid.UUID, 0, len(rows))
	seen := map[uuid.UUID]bool{}
	for i := range rows {
		if rows[i].AuditResultReportID == nil || *rows[i].AuditResultReportID == uuid.Nil {
			continue
		}
		id := *rows[i].AuditResultReportID
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	out := map[uuid.UUID]*models.AuditResultReport{}
	if len(ids) == 0 {
		return out, nil
	}
	var reports []models.AuditResultReport
	if err := ctl.DB.Where("id IN ?", ids).Find(&reports).Error; err != nil {
		return nil, err
	}
	for i := range reports {
		out[reports[i].ID] = &reports[i]
	}
	return out, nil
}

// lhaOf picks the LHA of r from a lhasFor result
func lhaOf(lhas map[uuid.UUID]*models.AuditResultReport, r *models.ActionTakenReport) *models.AuditResultReport {
	if r.AuditResultReportID == nil {
		return nil
	}
	return lhas[*r.AuditResultReportID]
}
