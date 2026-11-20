package me_agenda

import (
	"sort"
	"strings"
	"time"

	"audit-service/models"
)

// Todo kinds.
const (
	TodoInterview   = "INTERVIEW"     // interview scheduled, no record/evidence attached
	TodoObservation = "OBSERVATION"   // observation scheduled, no record/evidence attached
	TodoPBCDocument = "PBC_DOCUMENT"  // requested document without an uploaded file
	TodoTestControl = "TEST_CONTROL"  // control test without a test result
	TodoKKAForm     = "KKA_FORM"      // working paper form F01..F05 with no rows
	TodoLHADraft    = "LHA_DRAFT"     // LHA still a draft (not submitted for review)
	TodoLHAReview   = "LHA_REVIEW"    // LHA submitted, waiting for review/CAE approval
	TodoATRReview   = "ATR_REVIEW"    // ATR PENDING_REVIEW: auditor must verify
	TodoATRFollowUp = "ATR_FOLLOW_UP" // open ATR overdue or due soon: chase the PIC
)

// ATRDueSoonDays: an open (PLANNED/IN_PROGRESS) ATR becomes a follow-up to-do
// when it is overdue or due within this many days. PENDING_REVIEW is always one.
const ATRDueSoonDays = 14

// Todo is one entry of GET /me/todos.
type Todo struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Title string `json:"title"`
	// Form is F01..F05 for KKA_FORM, otherwise empty.
	Form string `json:"form,omitempty"`
	// DueDate is YYYY-MM-DD, null when the item has no date (KKA forms, LHA).
	DueDate            *string `json:"due_date"`
	IsOverdue          bool    `json:"is_overdue"`
	OverdueDays        int     `json:"overdue_days"`
	AssignmentLetterID string  `json:"assignment_letter_id"`
	LetterNumber       string  `json:"letter_number"`
	AuditTitle         string  `json:"audit_title,omitempty"`
	MyRole             string  `json:"my_role"` // LEADER | MEMBER
	EntityType         string  `json:"entity_type"`
	EntityID           string  `json:"entity_id"`
	Status             string  `json:"status,omitempty"`
	To                 string  `json:"to"`
}

// TodosMeta describes the caller's link and the list.
type TodosMeta struct {
	Linked      bool       `json:"linked"`
	LinkStatus  LinkStatus `json:"link_status"`
	MineLetters int        `json:"mine_letters"`
	Total       int        `json:"total"`
	Overdue     int        `json:"overdue"`
	// Scope: "full" for audit staff; "none" when the role gets no to-do list.
	Scope string `json:"scope"`
}

// TodosResult is the data of GET /me/todos.
type TodosResult struct {
	Todos []Todo    `json:"todos"`
	Meta  TodosMeta `json:"meta"`
}

// ScopeNone: the role has no audit to-do list (VIEWER, AUDITEE, ...).
const ScopeNone = "none"

type todoBuilder struct {
	todos []Todo
	today time.Time
}

func (b *todoBuilder) add(lc *letterCtx, t Todo, due *time.Time) {
	l := lc.letter
	t.AssignmentLetterID = l.ID.String()
	t.LetterNumber = l.LetterNumber
	t.AuditTitle = l.AuditTitle
	t.MyRole = lc.membership.Role
	if due != nil {
		s := FormatDay(*due)
		t.DueDate = &s
		if due.Before(b.today) {
			t.IsOverdue = true
			t.OverdueDays = daysBetween(*due, b.today)
		}
	}
	b.todos = append(b.todos, t)
}

// BuildTodos derives the caller's to-do list from the audits they are on.
// A caller who cannot be matched to any letter gets an empty list; the
// meta.linked / link_status fields say whether that is because the account
// cannot be matched at all. "Not done" is derived from fields that exist:
//   - interview, observation: no file attached (there is no status column)
//   - PBC document: no file uploaded
//   - test control: empty test result
//   - KKA form: the form has no rows for the letter
//   - LHA: status Draft (submit it) or in review (waiting)
//   - ATR: PENDING_REVIEW, or open and overdue / due soon
func BuildTodos(d *Data, c Caller, now time.Time) TodosResult {
	b := &todoBuilder{today: Today(now)}
	ix := newIndex(d, c)
	link := c.Link()

	mine := 0
	for _, lc := range ix.letters {
		if lc.membership.Mine() {
			mine++
		}
	}
	mineActive := func(key string) *letterCtx {
		if lc := ix.byKey[strings.TrimSpace(key)]; lc != nil && lc.active() && lc.membership.Mine() {
			return lc
		}
		return nil
	}
	due := func(raw string) *time.Time { return dayPtr(ParseDay(raw)) }

	for i := range d.Interviews {
		r := &d.Interviews[i]
		if lc := mineActive(r.AssignmentLetterID); lc != nil && !hasFile(r.FileName, r.FilePath, r.FileUrl) {
			b.add(lc, Todo{ID: "interview:" + r.ID.String(), Kind: TodoInterview,
				Title:      firstNonEmpty(r.Topic, r.Interviewee, lc.letter.LetterNumber),
				EntityType: "fieldwork_interview", EntityID: r.ID.String(),
				To: routeWithLetter(RouteFieldwork, lc.letter.LetterNumber)}, due(r.Date))
		}
	}
	for i := range d.Observations {
		r := &d.Observations[i]
		if lc := mineActive(r.AssignmentLetterID); lc != nil && !hasFile(r.FileName, r.FilePath, r.FileUrl) {
			b.add(lc, Todo{ID: "observation:" + r.ID.String(), Kind: TodoObservation,
				Title:      firstNonEmpty(r.Activity, lc.letter.LetterNumber),
				EntityType: "fieldwork_observation", EntityID: r.ID.String(),
				To: routeWithLetter(RouteFieldwork, lc.letter.LetterNumber)}, due(r.Date))
		}
	}
	for i := range d.Documents {
		r := &d.Documents[i]
		if lc := mineActive(r.AssignmentLetterID); lc != nil && !hasFile(r.FileName, r.FilePath, r.FileUrl) {
			b.add(lc, Todo{ID: "pbc:" + r.ID.String(), Kind: TodoPBCDocument,
				Title:      firstNonEmpty(r.DocumentName, lc.letter.LetterNumber),
				EntityType: "fieldwork_document", EntityID: r.ID.String(),
				To: routeWithLetter(RouteFieldwork, lc.letter.LetterNumber)}, due(r.RequiredDate))
		}
	}
	for i := range d.TestControls {
		r := &d.TestControls[i]
		if lc := mineActive(r.AssignmentLetterID); lc != nil && !testControlDone(r) {
			b.add(lc, Todo{ID: "testcontrol:" + r.ID.String(), Kind: TodoTestControl,
				Title:      firstNonEmpty(r.ControlName, lc.letter.LetterNumber),
				EntityType: "fieldwork_test_control", EntityID: r.ID.String(),
				To: routeWithLetter(RouteFieldwork, lc.letter.LetterNumber)}, due(r.DueDate))
		}
	}

	for _, lc := range ix.letters {
		if !lc.membership.Mine() || !lc.active() {
			continue
		}
		l := lc.letter
		present := map[string]bool{}
		for _, k := range letterKeys(l) {
			for code, ok := range d.KKAForms[k] {
				if ok {
					present[code] = true
				}
			}
		}
		for _, code := range KKAFormCodes {
			if present[code] {
				continue
			}
			b.add(lc, Todo{ID: "kka:" + l.ID.String() + ":" + code, Kind: TodoKKAForm, Form: code,
				Title: code, EntityType: "working_paper_form", EntityID: l.ID.String(),
				To: routeWithLetter(RouteWorking, l.LetterNumber, "step", strings.ToLower(code))}, nil)
		}
		for _, r := range lc.lhaDraft {
			b.add(lc, Todo{ID: "lha:" + r.ID.String(), Kind: TodoLHADraft,
				Title:      firstNonEmpty(r.ReportTitle, r.Title, r.ReportNumber, l.LetterNumber),
				EntityType: "audit_result_report", EntityID: r.ID.String(), Status: r.Status,
				To: routeWithLetter(RouteLHA, l.LetterNumber)}, nil)
		}
		for _, r := range lc.lhaInReview {
			b.add(lc, Todo{ID: "lha:" + r.ID.String(), Kind: TodoLHAReview,
				Title:      firstNonEmpty(r.ReportTitle, r.Title, r.ReportNumber, l.LetterNumber),
				EntityType: "audit_result_report", EntityID: r.ID.String(), Status: r.Status,
				To: routeWithLetter(RouteLHA, l.LetterNumber)}, nil)
		}
	}

	for i := range d.ATRs {
		r := &d.ATRs[i]
		if r.AuditResultReportID == nil || models.IsATRClosed(r.Status) {
			continue
		}
		lc := ix.reportLetter[*r.AuditResultReportID]
		if lc == nil || !lc.membership.Mine() {
			continue
		}
		st, _ := models.NormalizeATRStatus(r.Status)
		var dueDay *time.Time
		if r.DueDate != nil {
			dueDay = dayPtr(ParseDay(r.DueDate.UTC().Format("2006-01-02")))
		}
		kind := ""
		switch {
		case st == models.ATRStatusPendingReview:
			kind = TodoATRReview
		case st == models.ATRStatusPlanned || st == models.ATRStatusInProgress:
			if dueDay != nil && !dueDay.After(b.today.AddDate(0, 0, ATRDueSoonDays)) {
				kind = TodoATRFollowUp
			}
		}
		if kind == "" {
			continue
		}
		b.add(lc, Todo{ID: "atr:" + r.ID.String(), Kind: kind,
			Title:      firstNonEmpty(r.FindingTitle, r.ReportNumber, "-"),
			EntityType: "action_taken_report", EntityID: r.ID.String(), Status: r.Status,
			To: RouteATR + "?lha=" + r.AuditResultReportID.String()}, dueDay)
	}

	sort.SliceStable(b.todos, func(i, j int) bool {
		x, y := b.todos[i], b.todos[j]
		if x.IsOverdue != y.IsOverdue {
			return x.IsOverdue
		}
		if (x.DueDate == nil) != (y.DueDate == nil) {
			return x.DueDate != nil
		}
		if x.DueDate != nil && *x.DueDate != *y.DueDate {
			return *x.DueDate < *y.DueDate
		}
		if x.LetterNumber != y.LetterNumber {
			return x.LetterNumber < y.LetterNumber
		}
		return x.ID < y.ID
	})
	overdue := 0
	for _, t := range b.todos {
		if t.IsOverdue {
			overdue++
		}
	}
	if b.todos == nil {
		b.todos = []Todo{}
	}
	return TodosResult{Todos: b.todos, Meta: TodosMeta{
		Linked: link != LinkNone, LinkStatus: link, MineLetters: mine,
		Total: len(b.todos), Overdue: overdue, Scope: ScopeFull,
	}}
}

// EmptyTodos is the answer for a role that gets no to-do list.
func EmptyTodos(c Caller) TodosResult {
	link := c.Link()
	return TodosResult{Todos: []Todo{}, Meta: TodosMeta{Linked: link != LinkNone, LinkStatus: link, Scope: ScopeNone}}
}
