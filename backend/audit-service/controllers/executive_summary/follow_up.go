package executivesummary

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"audit-service/models"
	apperrors "audit-service/pkg/errors"
	"audit-service/pkg/logger"
	"audit-service/pkg/response"
	atrsvc "audit-service/services/action_taken_report"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// FollowUpNoAction is the label of the fourth follow-up table row and of the
// statusTL / status of a frozen finding row in the NO_ACTION bucket
const FollowUpNoAction = "No Action"

// followUpLabels maps a bucket onto the label the executive summary JSON uses
var followUpLabels = map[string]string{
	models.FollowUpBucketClosed:     FollowUpClosed,
	models.FollowUpBucketInProgress: FollowUpInProgress,
	models.FollowUpBucketOverdue:    FollowUpOverdue,
	models.FollowUpBucketNoAction:   FollowUpNoAction,
}

// FollowUpBuckets counts findings per follow-up bucket
type FollowUpBuckets struct {
	Closed     int64 `json:"closed"`
	InProgress int64 `json:"inProgress"`
	Overdue    int64 `json:"overdue"`
	NoAction   int64 `json:"noAction"`
}

// Total is the number of counted findings (cancelled ones are not counted)
func (b FollowUpBuckets) Total() int64 {
	return b.Closed + b.InProgress + b.Overdue + b.NoAction
}

func (b *FollowUpBuckets) add(bucket string) {
	switch bucket {
	case models.FollowUpBucketClosed:
		b.Closed++
	case models.FollowUpBucketInProgress:
		b.InProgress++
	case models.FollowUpBucketOverdue:
		b.Overdue++
	case models.FollowUpBucketNoAction:
		b.NoAction++
	}
}

// FollowUpEvidence is one evidence file (bukti TL) of a finding's ATR
type FollowUpEvidence struct {
	ID         string    `json:"id"`
	FileName   string    `json:"fileName"`
	UploadedAt time.Time `json:"uploadedAt"`
}

// FollowUpFinding is one finding in the follow-up status. Pointer fields are
// null when the finding has no ATR row (or the ATR has no value).
type FollowUpFinding struct {
	FindingID           string             `json:"findingId"`
	AuditResultReportID string             `json:"auditResultReportId"`
	ReportNumber        string             `json:"reportNumber"`
	ATRID               *string            `json:"atrId"`
	Title               string             `json:"title"`
	Divisi              string             `json:"divisi"`
	ATRStatus           *string            `json:"atrStatus"`
	Bucket              string             `json:"bucket"`
	NoActionReason      *string            `json:"noActionReason"`
	DueDate             *string            `json:"dueDate"`
	Progress            *int               `json:"progress"`
	PICName             *string            `json:"picName"`
	OverdueDays         int                `json:"overdueDays"`
	Evidence            []FollowUpEvidence `json:"evidence"`
}

// FollowUpStatus is the data payload of GET /executive-summaries/follow-up-status
type FollowUpStatus struct {
	AsOf    time.Time       `json:"asOf"`
	Buckets FollowUpBuckets `json:"buckets"`
	Total   int64           `json:"total"`
	// Findings is empty (never null) for callers without ATR view_all;
	// FindingsRestricted is then true
	Findings           []FollowUpFinding `json:"findings"`
	FindingsRestricted bool              `json:"findingsRestricted"`
	// Cancelled counts findings left out because their ATR is CANCELLED
	Cancelled int64 `json:"-"`
}

// LoadFollowUp buckets every finding of the given approved LHAs at asOf (see
// models.ClassifyFollowUp). The LHAs must carry id, report_number, department
// and findings.
//
//   - A finding is matched to its ATR by (LHA id, finding id).
//   - A finding whose ATR was soft-deleted by an admin is left out, like a
//     cancelled one (SyncFromReport never recreates it), but not counted as
//     cancelled.
//   - A live ATR of the LHA whose finding is no longer in the LHA (removed
//     after work started, see SyncFromReport) is still followed up, so it is
//     counted too, with the finding copy on the ATR.
//   - The LHA approval time for NO_PLAN_14D is the creation time of the LHA's
//     first ATR: ATRs are created when the LHA is approved.
//
// Evidence is loaded only when withEvidence is set. Findings are in LHA order,
// then finding order, then orphan ATRs.
func LoadFollowUp(db *gorm.DB, lhas []models.AuditResultReport, asOf time.Time, withEvidence bool) (FollowUpStatus, error) {
	out := FollowUpStatus{AsOf: asOf, Findings: []FollowUpFinding{}}
	if len(lhas) == 0 {
		return out, nil
	}

	ids := make([]uuid.UUID, 0, len(lhas))
	for _, l := range lhas {
		ids = append(ids, l.ID)
	}
	var atrs []models.ActionTakenReport
	if err := db.Unscoped().
		Where("audit_result_report_id IN ?", ids).
		Order("created_at ASC").Order("id ASC").
		Find(&atrs).Error; err != nil {
		return out, err
	}

	type atrKey struct {
		report  uuid.UUID
		finding string
	}
	byKey := make(map[atrKey]*models.ActionTakenReport, len(atrs))
	approvedAt := map[uuid.UUID]time.Time{}
	for i := range atrs {
		a := &atrs[i]
		if a.AuditResultReportID == nil {
			continue
		}
		rid := *a.AuditResultReportID
		if t, ok := approvedAt[rid]; !ok || a.CreatedAt.Before(t) {
			approvedAt[rid] = a.CreatedAt
		}
		k := atrKey{rid, findingKey(a.FindingID)}
		// A live row wins over a soft-deleted one in the same slot
		if prev, ok := byKey[k]; !ok || (prev.DeletedAt.Valid && !a.DeletedAt.Valid) {
			byKey[k] = a
		}
	}

	used := map[*models.ActionTakenReport]bool{}
	listedATRs := []*models.ActionTakenReport{}
	add := func(lha *models.AuditResultReport, findingID, title string, atr *models.ActionTakenReport) {
		in := models.FollowUpInput{}
		if t, ok := approvedAt[lha.ID]; ok {
			in.LHAApprovedAt = &t
		}
		if atr != nil {
			in.HasATR = true
			in.Status = atr.Status
			in.DueDate = atr.DueDate
			in.PICUserID = atr.PICUserID
			in.PICName = atr.PICName
		}
		res := models.ClassifyFollowUp(in, asOf)
		if res.Excluded {
			out.Cancelled++
			return
		}
		out.Buckets.add(res.Bucket)
		f := FollowUpFinding{
			FindingID:           findingID,
			AuditResultReportID: lha.ID.String(),
			ReportNumber:        strings.TrimSpace(lha.ReportNumber),
			Title:               title,
			Divisi:              strings.TrimSpace(lha.Department),
			Bucket:              res.Bucket,
			OverdueDays:         res.OverdueDays,
			Evidence:            []FollowUpEvidence{},
		}
		if res.NoActionReason != "" {
			reason := res.NoActionReason
			f.NoActionReason = &reason
		}
		if atr != nil {
			id := atr.ID.String()
			f.ATRID = &id
			status := displayATRStatus(atr.Status)
			f.ATRStatus = &status
			if atr.DueDate != nil && !atr.DueDate.IsZero() {
				d := atr.DueDate.UTC().Format(dateLayout)
				f.DueDate = &d
			}
			progress := atr.Progress
			f.Progress = &progress
			if name := strings.TrimSpace(atr.PICName); name != "" {
				f.PICName = &name
			}
			if f.Title == "" {
				f.Title = strings.TrimSpace(atr.FindingTitle)
			}
			listedATRs = append(listedATRs, atr)
		}
		out.Findings = append(out.Findings, f)
	}

	lhaByID := make(map[uuid.UUID]*models.AuditResultReport, len(lhas))
	for i := range lhas {
		lha := &lhas[i]
		lhaByID[lha.ID] = lha
		for _, finding := range lha.Findings {
			fid := findingKey(finding.ID)
			atr := byKey[atrKey{lha.ID, fid}]
			if fid == "" {
				atr = nil
			}
			if atr != nil {
				used[atr] = true
				if atr.DeletedAt.Valid {
					continue
				}
			}
			add(lha, fid, strings.TrimSpace(finding.Title), atr)
		}
	}
	// Live ATRs of these LHAs whose finding is not (or no longer) in the LHA
	for i := range atrs {
		a := &atrs[i]
		if used[a] || a.DeletedAt.Valid || a.AuditResultReportID == nil {
			continue
		}
		lha := lhaByID[*a.AuditResultReportID]
		if lha == nil {
			continue
		}
		add(lha, findingKey(a.FindingID), strings.TrimSpace(a.FindingTitle), a)
	}
	out.Total = out.Buckets.Total()

	if withEvidence && len(listedATRs) > 0 {
		atrIDs := make([]uuid.UUID, 0, len(listedATRs))
		for _, a := range listedATRs {
			atrIDs = append(atrIDs, a.ID)
		}
		var evidence []models.ActionTakenReportEvidence
		if err := db.Where("action_taken_report_id IN ?", atrIDs).
			Order("uploaded_at ASC").Order("id ASC").
			Find(&evidence).Error; err != nil {
			return out, err
		}
		byATR := map[string][]FollowUpEvidence{}
		for _, e := range evidence {
			k := e.ActionTakenReportID.String()
			byATR[k] = append(byATR[k], FollowUpEvidence{ID: e.ID.String(), FileName: e.FileName, UploadedAt: e.UploadedAt})
		}
		for i := range out.Findings {
			if id := out.Findings[i].ATRID; id != nil && len(byATR[*id]) > 0 {
				out.Findings[i].Evidence = byATR[*id]
			}
		}
	}
	return out, nil
}

// findingKey is the canonical (lower-case, trimmed) form of a finding id
func findingKey(id string) string {
	return strings.ToLower(strings.TrimSpace(id))
}

// displayATRStatus is the stored status in its canonical form; a blank status
// is PLANNED (the column default) and an unknown legacy value is kept as is
func displayATRStatus(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return models.ATRStatusPlanned
	}
	if s, ok := models.NormalizeATRStatus(raw); ok {
		return s
	}
	return strings.TrimSpace(raw)
}

// approvedLHAsByID returns the LHA with id when it is approved; an empty list
// when it exists but is not approved. gorm.ErrRecordNotFound when it does not
// exist (or is deleted).
func approvedLHAsByID(db *gorm.DB, id uuid.UUID) ([]models.AuditResultReport, error) {
	var lha models.AuditResultReport
	if err := db.Select(lhaColumns).Where("id = ?", id).Take(&lha).Error; err != nil {
		return nil, err
	}
	if !models.IsLHAApproved(lha.Status) {
		return []models.AuditResultReport{}, nil
	}
	return []models.AuditResultReport{lha}, nil
}

// approvedLHAsByNumber returns the approved LHAs with the report number
// (trimmed, exact match). gorm.ErrRecordNotFound when no LHA has the number.
func approvedLHAsByNumber(db *gorm.DB, number string) ([]models.AuditResultReport, error) {
	number = strings.TrimSpace(number)
	var rows []models.AuditResultReport
	if err := db.Select(lhaColumns).
		Where("TRIM(report_number) = ?", number).
		Order("id ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	approved := make([]models.AuditResultReport, 0, len(rows))
	for _, r := range rows {
		if models.IsLHAApproved(r.Status) {
			approved = append(approved, r)
		}
	}
	return approved, nil
}

// errNoFollowUpScope: none of audit_result_report_id, report_number, year given
var errNoFollowUpScope = apperrors.New("FOLLOW_UP_SCOPE_REQUIRED",
	"one of audit_result_report_id, report_number or year is required", http.StatusBadRequest)

// followUpScopeFromQuery resolves the LHAs of a follow-up status request:
// audit_result_report_id (Individual), else report_number (Individual,
// fallback), else year with an optional quarter (Compilation, the same LHAs as
// compilation-stats).
func followUpScopeFromQuery(db *gorm.DB, c *gin.Context) ([]models.AuditResultReport, *apperrors.AppError) {
	if raw := strings.TrimSpace(c.Query("audit_result_report_id")); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil, apperrors.New("INVALID_AUDIT_RESULT_REPORT_ID", "audit_result_report_id must be a UUID", http.StatusBadRequest)
		}
		lhas, err := approvedLHAsByID(db, id)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New("AUDIT_RESULT_REPORT_NOT_FOUND", "Audit result report not found", http.StatusNotFound)
		}
		if err != nil {
			return nil, apperrors.Wrap("INTERNAL_ERROR", "Failed to load the audit result report", http.StatusInternalServerError, err)
		}
		return lhas, nil
	}
	if raw := strings.TrimSpace(c.Query("report_number")); raw != "" {
		lhas, err := approvedLHAsByNumber(db, raw)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.New("AUDIT_RESULT_REPORT_NOT_FOUND", "Audit result report not found", http.StatusNotFound)
		}
		if err != nil {
			return nil, apperrors.Wrap("INTERNAL_ERROR", "Failed to load the audit result report", http.StatusInternalServerError, err)
		}
		return lhas, nil
	}
	if raw := strings.TrimSpace(c.Query("year")); raw != "" {
		year, err := strconv.Atoi(raw)
		if err != nil || year < 1 {
			return nil, apperrors.New("INVALID_YEAR", "year must be a year (YYYY)", http.StatusBadRequest)
		}
		quarter := 0
		if rawQ := strings.TrimSpace(c.Query("quarter")); rawQ != "" {
			quarter, err = strconv.Atoi(rawQ)
			if err != nil || quarter < 1 || quarter > 4 {
				return nil, apperrors.New("INVALID_QUARTER", ErrInvalidQuarter.Error(), http.StatusBadRequest)
			}
		}
		lhas, err := approvedLHAs(db, year, quarter)
		if err != nil {
			return nil, apperrors.Wrap("INTERNAL_ERROR", "Failed to load the audit result reports", http.StatusInternalServerError, err)
		}
		return lhas, nil
	}
	return nil, errNoFollowUpScope
}

// FollowUpStatusHandler handles GET /api/v1/executive-summaries/follow-up-status:
//
//	?audit_result_report_id=<uuid>   one LHA (Individual)
//	?report_number=<nomor>           LHA(s) by report number (Individual, fallback)
//	?year=YYYY[&quarter=N]           the LHAs compilation-stats counts (Compilation)
//
// Counts are live (as of now) and open to any authenticated caller, like
// compilation-stats. The per-finding list is returned only to callers with
// the ATR view_all permission; others get an empty list and
// findingsRestricted = true. An LHA that is not approved yields zero counts.
// 400 for a missing or invalid parameter, 404 for an unknown LHA.
func FollowUpStatusHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		now := time.Now().UTC()
		tx := db.WithContext(c.Request.Context())

		lhas, appErr := followUpScopeFromQuery(tx, c)
		if appErr != nil {
			if appErr.StatusCode >= http.StatusInternalServerError && logger.Log != nil {
				logger.Error("failed to resolve executive summary follow-up scope", zap.Error(appErr))
			}
			writeAppError(c, appErr)
			return
		}

		rawRoles, _ := c.Get("roles")
		roles, _ := rawRoles.([]string)
		canSeeFindings := atrsvc.PermissionsFor(roles)[atrsvc.PermViewAll]

		status, err := LoadFollowUp(tx, lhas, now, canSeeFindings)
		if err != nil {
			if logger.Log != nil {
				logger.Error("failed to load executive summary follow-up status", zap.Error(err))
			}
			response.InternalServerError(c, "Failed to fetch follow-up status")
			return
		}
		if !canSeeFindings {
			status.Findings = []FollowUpFinding{}
			status.FindingsRestricted = true
		}
		response.OK(c, "Follow-up status retrieved successfully", status)
	}
}

// compilationPrefixes identify a compilation executive summary by its document
// number: "ESC-" today, the other two in older documents
var compilationPrefixes = []string{"ESC-", "EXSUM-KOMP-", "KOMPILASI-"}

// IsCompilationSummary reports whether the document number is a compilation's
func IsCompilationSummary(nomorDokumen string) bool {
	n := strings.ToUpper(strings.TrimSpace(nomorDokumen))
	for _, p := range compilationPrefixes {
		if strings.HasPrefix(n, p) {
			return true
		}
	}
	return false
}

// summaryLHAs returns the approved LHAs an executive summary reports on: a
// compilation covers its year (and quarter, 1-4), an individual summary the
// LHA whose report number is its nomor dokumen. ok is false when the scope
// cannot be resolved (no year, no matching approved LHA).
func summaryLHAs(db *gorm.DB, es *models.ExecutiveSummary) (lhas []models.AuditResultReport, ok bool, err error) {
	if IsCompilationSummary(es.NomorDokumen) {
		if es.Tahun < 1 {
			return nil, false, nil
		}
		quarter := es.Quarter
		if quarter < 1 || quarter > 4 {
			quarter = 0
		}
		lhas, err = approvedLHAs(db, es.Tahun, quarter)
		return lhas, err == nil, err
	}
	if strings.TrimSpace(es.NomorDokumen) == "" {
		return nil, false, nil
	}
	lhas, err = approvedLHAsByNumber(db, es.NomorDokumen)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return lhas, len(lhas) > 0, nil
}

// FreezeFollowUp recomputes the follow-up status of the summary at asOf (the
// approval time) and returns the executive_summaries columns to store with
// the approval:
//
//   - follow_up_table: Closed, In Progress, Overdue, plus a fourth "No Action"
//     row only when the stored table already has one (older frontends accept
//     exactly three rows). Each row keeps its keterangan (and any other key);
//     persentase is jumlah / total of the rows written.
//   - top_findings[].statusTL and matriks_kompilasi[].status of rows that
//     carry a findingId found in the follow-up. Rows without a findingId, or
//     whose finding is excluded (cancelled ATR) or out of scope, are untouched.
//
// Nothing is returned when the scope cannot be resolved, so a summary that
// matches no approved LHA keeps what was typed in.
func FreezeFollowUp(db *gorm.DB, es *models.ExecutiveSummary, asOf time.Time) (map[string]interface{}, error) {
	lhas, ok, err := summaryLHAs(db, es)
	if err != nil || !ok {
		return nil, err
	}
	status, err := LoadFollowUp(db, lhas, asOf, false)
	if err != nil {
		return nil, err
	}

	cols := map[string]interface{}{}
	if table, err := frozenFollowUpTable(es.FollowUpTable, status.Buckets); err == nil {
		cols["follow_up_table"] = table
	} else {
		return nil, err
	}

	labels := make(map[string]string, len(status.Findings))
	for _, f := range status.Findings {
		if f.FindingID != "" {
			labels[f.FindingID] = followUpLabels[f.Bucket]
		}
	}
	if s, changed := relabelFindingRows(es.TopFindings, "statusTL", labels); changed {
		cols["top_findings"] = s
	}
	if s, changed := relabelFindingRows(es.MatriksKompilasi, "status", labels); changed {
		cols["matriks_kompilasi"] = s
	}
	return cols, nil
}

// frozenFollowUpTable builds the follow-up table JSON from the buckets,
// keeping the notes of the stored table (see FreezeFollowUp). A stored value
// that is not a JSON array is replaced.
func frozenFollowUpTable(stored string, b FollowUpBuckets) (string, error) {
	var rows []map[string]interface{}
	if strings.TrimSpace(stored) != "" {
		if err := json.Unmarshal([]byte(stored), &rows); err != nil {
			rows = nil
		}
	}
	previous := map[string]map[string]interface{}{}
	withNoAction := false
	for _, r := range rows {
		label, _ := r["status"].(string)
		key := strings.ToLower(strings.Join(strings.Fields(label), " "))
		if key == strings.ToLower(FollowUpNoAction) {
			withNoAction = true
		}
		if _, seen := previous[key]; !seen && r != nil {
			previous[key] = r
		}
	}

	counts := []struct {
		label string
		n     int64
	}{
		{FollowUpClosed, b.Closed},
		{FollowUpInProgress, b.InProgress},
		{FollowUpOverdue, b.Overdue},
	}
	if withNoAction {
		counts = append(counts, struct {
			label string
			n     int64
		}{FollowUpNoAction, b.NoAction})
	}
	var total int64
	for _, c := range counts {
		total += c.n
	}

	out := make([]map[string]interface{}, 0, len(counts))
	for _, c := range counts {
		row := map[string]interface{}{}
		for k, v := range previous[strings.ToLower(c.label)] {
			row[k] = v
		}
		if _, ok := row["keterangan"]; !ok {
			row["keterangan"] = ""
		}
		row["status"] = c.label
		row["jumlah"] = c.n
		pct := 0.0
		if total > 0 {
			pct = float64(c.n) / float64(total) * 100
		}
		row["persentase"] = pct
		out = append(out, row)
	}
	encoded, err := json.Marshal(out)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

// relabelFindingRows sets field on every row of the stored JSON array that
// has a findingId present in labels. changed is false (and the stored value
// should be kept verbatim) when nothing changed or the value is not an array
// of objects.
func relabelFindingRows(stored, field string, labels map[string]string) (string, bool) {
	if strings.TrimSpace(stored) == "" || len(labels) == 0 {
		return stored, false
	}
	var rows []map[string]interface{}
	if err := json.Unmarshal([]byte(stored), &rows); err != nil {
		return stored, false
	}
	changed := false
	for _, r := range rows {
		if r == nil {
			continue
		}
		id, _ := r["findingId"].(string)
		label, ok := labels[findingKey(id)]
		if !ok || label == "" {
			continue
		}
		if cur, _ := r[field].(string); cur != label {
			r[field] = label
			changed = true
		}
	}
	if !changed {
		return stored, false
	}
	encoded, err := json.Marshal(rows)
	if err != nil {
		return stored, false
	}
	return string(encoded), true
}
