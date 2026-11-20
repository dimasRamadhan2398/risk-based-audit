package me_agenda

import (
	"net/url"
	"strings"
	"time"

	"audit-service/models"

	"github.com/google/uuid"
)

// Data is everything the builders read. The loader fills it from the
// database; tests build it by hand.
type Data struct {
	// Letters are the Published assignment letters (Draft letters never reach
	// the calendar or the to-do list).
	Letters      []models.AssignmentLetter
	Reports      []models.AuditResultReport
	Interviews   []models.FieldworkInterview
	Observations []models.FieldworkObservation
	Documents    []models.FieldworkDocument
	Samples      []models.FieldworkSample
	TestControls []models.FieldworkTestControl
	ATRs         []models.ActionTakenReport
	// KKAForms[letterKey] holds the form codes (F01..F05) with at least one
	// row. letterKey is the letter number or id, as stored on the form rows.
	KKAForms map[string]map[string]bool
}

// Form codes of the working paper (KKA) forms and the table behind each.
const (
	FormF01 = "F01" // working_paper_headers
	FormF02 = "F02" // working_paper_risks
	FormF03 = "F03" // working_paper_samples
	FormF04 = "F04" // working_paper_causes
	FormF05 = "F05" // working_paper_plans
)

// KKAFormCodes in display order.
var KKAFormCodes = []string{FormF01, FormF02, FormF03, FormF04, FormF05}

// letterCtx is a letter with what is derived about it for this caller.
type letterCtx struct {
	letter     *models.AssignmentLetter
	membership Membership
	// completed: an approved (Final/Approved/Published) LHA exists, so the
	// audit is past fieldwork and reporting. Its fieldwork and forms are no
	// longer "ongoing or upcoming"; ATRs follow-ups still are.
	completed bool
	// lhaDraft / lhaInReview: the letter has an LHA in that state.
	lhaDraft    []*models.AuditResultReport
	lhaInReview []*models.AuditResultReport
}

func (l *letterCtx) active() bool { return !l.completed }

type index struct {
	byKey        map[string]*letterCtx // letter number and id, trimmed
	letters      []*letterCtx
	reportLetter map[uuid.UUID]*letterCtx
}

func letterKeys(l *models.AssignmentLetter) []string {
	keys := make([]string, 0, 2)
	if n := strings.TrimSpace(l.LetterNumber); n != "" {
		keys = append(keys, n)
	}
	keys = append(keys, l.ID.String())
	return keys
}

// LetterKeys are the values form rows may use to point at the letter.
func LetterKeys(l *models.AssignmentLetter) []string { return letterKeys(l) }

func newIndex(d *Data, c Caller) *index {
	ix := &index{byKey: map[string]*letterCtx{}, reportLetter: map[uuid.UUID]*letterCtx{}}
	for i := range d.Letters {
		l := &d.Letters[i]
		lc := &letterCtx{letter: l, membership: MatchLetter(c, l)}
		ix.letters = append(ix.letters, lc)
		for _, k := range letterKeys(l) {
			ix.byKey[k] = lc
		}
	}
	for i := range d.Reports {
		r := &d.Reports[i]
		lc := ix.byKey[strings.TrimSpace(r.AssignmentLetterID)]
		if lc == nil {
			continue
		}
		ix.reportLetter[r.ID] = lc
		switch {
		case models.IsLHARejectedOrCancelled(r.Status):
		case models.IsLHAApproved(r.Status):
			lc.completed = true
		case models.IsLHAInReview(r.Status):
			lc.lhaInReview = append(lc.lhaInReview, r)
		default:
			lc.lhaDraft = append(lc.lhaDraft, r)
		}
	}
	return ix
}

// Routes the frontend opens for each source. Query values are escaped here.
const (
	RouteLetters    = "/assignment-letter"
	RouteFieldwork  = "/audit-fieldwork"
	RouteWorking    = "/working-paper"
	RouteLHA        = "/audit-result-report"
	RouteATR        = "/action-taken-report"
	routeLetterParm = "letter"
)

func routeWithLetter(path, letterNumber string, extra ...string) string {
	q := url.Values{}
	for i := 0; i+1 < len(extra); i += 2 {
		q.Set(extra[i], extra[i+1])
	}
	if letterNumber != "" {
		q.Set(routeLetterParm, letterNumber)
	}
	if len(q) == 0 {
		return path
	}
	return path + "?" + q.Encode()
}

func hasFile(name, path, fileURL string) bool {
	return strings.TrimSpace(name) != "" || strings.TrimSpace(path) != "" || strings.TrimSpace(fileURL) != ""
}

// Today is the calendar day "now" falls on in Asia/Jakarta, as UTC midnight
// (the same rule ATR overdue uses).
func Today(now time.Time) time.Time { return models.ATRToday(now) }

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v = strings.TrimSpace(v); v != "" {
			return v
		}
	}
	return ""
}
