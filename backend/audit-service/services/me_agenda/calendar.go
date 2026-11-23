package me_agenda

import (
	"sort"
	"strings"
	"time"

	"audit-service/models"
)

// Calendar event kinds. They match the frontend's CalendarEventKind.
const (
	KindEngagement     = "engagement"
	KindInterview      = "interview"
	KindObservation    = "observation"
	KindSample         = "sample"
	KindPBCDue         = "pbcDue"
	KindTestControlDue = "testControlDue"
	KindATRDue         = "atrDue"
)

// MyRolePIC is the role of a caller who is the person in charge of an ATR.
const MyRolePIC = "PIC"

// CalendarEvent is one entry of GET /me/calendar.
type CalendarEvent struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Start string `json:"start"`
	End   string `json:"end,omitempty"`
	Title string `json:"title"`
	// Subtitle is the letter number (or the LHA number for an ATR).
	Subtitle string `json:"subtitle,omitempty"`
	// IsOverdue is only set for items with a deadline (PBC, test control, ATR).
	// Planned fieldwork dates are never overdue on the calendar.
	IsOverdue bool `json:"is_overdue"`
	// Done: the item has its evidence (file attached; test result filled).
	Done bool `json:"done"`
	// Mine: the caller is on the audit team (or the ATR's PIC).
	Mine               bool   `json:"mine"`
	MyRole             string `json:"my_role,omitempty"` // LEADER | MEMBER | PIC
	To                 string `json:"to,omitempty"`
	LetterNumber       string `json:"letter_number,omitempty"`
	AssignmentLetterID string `json:"assignment_letter_id,omitempty"`
	EntityType         string `json:"entity_type"`
	EntityID           string `json:"entity_id"`
	Status             string `json:"status,omitempty"`
}

// CalendarMeta describes the caller's link and what was left out.
type CalendarMeta struct {
	Linked     bool       `json:"linked"`
	LinkStatus LinkStatus `json:"link_status"` // employee | name_only | none
	// MineLetters is how many letters the caller was matched to.
	MineLetters int `json:"mine_letters"`
	// UnparsedDates counts events dropped because their date is not a date.
	UnparsedDates int `json:"unparsed_dates"`
	// Scope: "full" (audit staff), "engagements_only" (VIEWER) or "own_atr"
	// (anyone else, for instance AUDITEE).
	Scope string `json:"scope"`
}

// Calendar roles scope.
const (
	ScopeFull            = "full"
	ScopeEngagementsOnly = "engagements_only"
	ScopeOwnATR          = "own_atr"
)

// CalendarResult is the data of GET /me/calendar.
type CalendarResult struct {
	Range  CalendarRange   `json:"range"`
	Events []CalendarEvent `json:"events"`
	Meta   CalendarMeta    `json:"meta"`
}

// CalendarRange echoes the requested window.
type CalendarRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type calendarBuilder struct {
	events   []CalendarEvent
	from, to time.Time
	today    time.Time
	unparsed int
}

func (b *calendarBuilder) inWindow(start, end time.Time) bool {
	return !start.After(b.to) && !end.Before(b.from)
}

// point adds a single-day event when its day is inside the window.
func (b *calendarBuilder) point(e CalendarEvent, raw string) {
	day, ok := ParseDay(raw)
	if !ok {
		if strings.TrimSpace(raw) != "" {
			b.unparsed++
		}
		return
	}
	if !b.inWindow(day, day) {
		return
	}
	e.Start = FormatDay(day)
	e.IsOverdue = e.IsOverdue && !e.Done && day.Before(b.today)
	b.events = append(b.events, e)
}

// BuildCalendar lists, for the window [from, to] (UTC-midnight days,
// inclusive), the ongoing or upcoming audits of everybody, each flagged mine
// when the caller is on the team. scope decides how much of it the caller may
// see (see the Scope constants).
func BuildCalendar(d *Data, c Caller, from, to time.Time, scope string, now time.Time) CalendarResult {
	b := &calendarBuilder{from: from, to: to, today: Today(now)}
	ix := newIndex(d, c)
	link := c.Link()

	mineLetters := 0
	for _, lc := range ix.letters {
		if lc.membership.Mine() {
			mineLetters++
		}
	}

	if scope == ScopeFull || scope == ScopeEngagementsOnly {
		b.engagements(ix)
	}
	if scope == ScopeFull {
		b.fieldwork(d, ix)
	}
	if scope == ScopeFull || scope == ScopeOwnATR {
		b.atrs(d, ix, c, scope == ScopeOwnATR)
	}

	sort.SliceStable(b.events, func(i, j int) bool {
		a, e := b.events[i], b.events[j]
		if a.Start != e.Start {
			return a.Start < e.Start
		}
		return a.ID < e.ID
	})
	if b.events == nil {
		b.events = []CalendarEvent{}
	}
	return CalendarResult{
		Range:  CalendarRange{From: FormatDay(from), To: FormatDay(to)},
		Events: b.events,
		Meta: CalendarMeta{
			Linked: link != LinkNone, LinkStatus: link, MineLetters: mineLetters,
			UnparsedDates: b.unparsed, Scope: scope,
		},
	}
}

func (b *calendarBuilder) engagements(ix *index) {
	for _, lc := range ix.letters {
		if !lc.active() {
			continue
		}
		l := lc.letter
		start, okS := ParseDay(l.StartPeriod)
		end, okE := ParseDay(l.FinishPeriod)
		if !okS && !okE {
			if strings.TrimSpace(l.StartPeriod) != "" || strings.TrimSpace(l.FinishPeriod) != "" {
				b.unparsed++
			}
			continue
		}
		if !okS {
			start = end
		}
		if !okE || end.Before(start) {
			end = start
		}
		if !b.inWindow(start, end) {
			continue
		}
		e := CalendarEvent{
			ID: "letter:" + l.ID.String(), Kind: KindEngagement, Start: FormatDay(start),
			Title: firstNonEmpty(l.AuditTitle, l.LetterNumber), Subtitle: l.LetterNumber,
			Mine: lc.membership.Mine(), MyRole: lc.membership.Role, To: RouteLetters,
			LetterNumber: l.LetterNumber, AssignmentLetterID: l.ID.String(),
			EntityType: "assignment_letter", EntityID: l.ID.String(), Status: l.Status,
		}
		if !end.Equal(start) {
			e.End = FormatDay(end)
		}
		b.events = append(b.events, e)
	}
}

func (b *calendarBuilder) fieldwork(d *Data, ix *index) {
	base := func(lc *letterCtx, kind, entityType, id, title string) CalendarEvent {
		l := lc.letter
		return CalendarEvent{
			ID: kind + ":" + id, Kind: kind, Title: firstNonEmpty(title, l.LetterNumber), Subtitle: l.LetterNumber,
			Mine: lc.membership.Mine(), MyRole: lc.membership.Role,
			To:           routeWithLetter(RouteFieldwork, l.LetterNumber),
			LetterNumber: l.LetterNumber, AssignmentLetterID: l.ID.String(),
			EntityType: entityType, EntityID: id,
		}
	}
	live := func(key string) *letterCtx {
		if lc := ix.byKey[strings.TrimSpace(key)]; lc != nil && lc.active() {
			return lc
		}
		return nil
	}
	for i := range d.Interviews {
		r := &d.Interviews[i]
		if lc := live(r.AssignmentLetterID); lc != nil {
			e := base(lc, KindInterview, "fieldwork_interview", r.ID.String(), firstNonEmpty(r.Topic, r.Interviewee))
			e.Done = hasFile(r.FileName, r.FilePath, r.FileUrl)
			b.point(e, r.Date)
		}
	}
	for i := range d.Observations {
		r := &d.Observations[i]
		if lc := live(r.AssignmentLetterID); lc != nil {
			e := base(lc, KindObservation, "fieldwork_observation", r.ID.String(), r.Activity)
			e.Done = hasFile(r.FileName, r.FilePath, r.FileUrl)
			b.point(e, r.Date)
		}
	}
	for i := range d.Samples {
		r := &d.Samples[i]
		if lc := live(r.AssignmentLetterID); lc != nil {
			e := base(lc, KindSample, "fieldwork_sample", r.ID.String(), r.DocumentName)
			e.Done = hasFile(r.FileName, r.FilePath, r.FileUrl)
			b.point(e, r.Date)
		}
	}
	for i := range d.Documents {
		r := &d.Documents[i]
		if lc := live(r.AssignmentLetterID); lc != nil {
			e := base(lc, KindPBCDue, "fieldwork_document", r.ID.String(), r.DocumentName)
			e.Done = hasFile(r.FileName, r.FilePath, r.FileUrl)
			e.IsOverdue = !e.Done
			b.point(e, r.RequiredDate)
		}
	}
	for i := range d.TestControls {
		r := &d.TestControls[i]
		if lc := live(r.AssignmentLetterID); lc != nil {
			e := base(lc, KindTestControlDue, "fieldwork_test_control", r.ID.String(), r.ControlName)
			e.Done = testControlDone(r)
			e.IsOverdue = !e.Done
			b.point(e, r.DueDate)
		}
	}
}

func testControlDone(r *models.FieldworkTestControl) bool {
	return strings.TrimSpace(r.TestResult) != ""
}

// atrs adds open ATR due dates. ownOnly restricts them to ATRs whose PIC is
// the caller (the ATR scope of AUDITEE and DEPARTMENT_HEAD).
func (b *calendarBuilder) atrs(d *Data, ix *index, c Caller, ownOnly bool) {
	for i := range d.ATRs {
		r := &d.ATRs[i]
		if models.IsATRClosed(r.Status) || r.DueDate == nil {
			continue
		}
		isPIC := strings.TrimSpace(c.UserID) != "" && strings.EqualFold(strings.TrimSpace(r.PICUserID), strings.TrimSpace(c.UserID))
		if ownOnly && !isPIC {
			continue
		}
		var lc *letterCtx
		if r.AuditResultReportID != nil {
			lc = ix.reportLetter[*r.AuditResultReportID]
		}
		e := CalendarEvent{
			ID: "atr:" + r.ID.String(), Kind: KindATRDue,
			Title: firstNonEmpty(r.FindingTitle, r.ReportNumber, "-"), Subtitle: r.ReportNumber,
			To: RouteATR, EntityType: "action_taken_report", EntityID: r.ID.String(), Status: r.Status,
			IsOverdue: r.IsOverdue, // derived by the model on load
		}
		if r.AuditResultReportID != nil {
			e.To = RouteATR + "?lha=" + r.AuditResultReportID.String()
		}
		if lc != nil {
			e.LetterNumber = lc.letter.LetterNumber
			e.AssignmentLetterID = lc.letter.ID.String()
			if lc.membership.Mine() {
				e.Mine, e.MyRole = true, lc.membership.Role
			}
		}
		if isPIC && !e.Mine {
			e.Mine, e.MyRole = true, MyRolePIC
		}
		// Derive overdue here as well so a row that skipped AfterFind is right.
		ov, _ := models.ComputeATROverdue(r.DueDate, r.Status, b.today)
		e.IsOverdue = e.IsOverdue || ov
		day, _ := ParseDay(r.DueDate.UTC().Format("2006-01-02"))
		if !b.inWindow(day, day) {
			continue
		}
		e.Start = FormatDay(day)
		b.events = append(b.events, e)
	}
}
