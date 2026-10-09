package models

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm/schema"
)

func jakarta(t *testing.T) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.FixedZone("WIB", 7*60*60)
	}
	return loc
}

// day returns the stored form of a due date: the calendar date at UTC midnight
func day(y int, m time.Month, d int) *time.Time {
	t := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return &t
}

func TestComputeATROverdue(t *testing.T) {
	wib := jakarta(t)
	// 1 Oct 2026, 10:00 in Jakarta
	now := time.Date(2026, 10, 1, 10, 0, 0, 0, wib)

	cases := []struct {
		name    string
		due     *time.Time
		status  string
		now     time.Time
		overdue bool
		days    int
	}{
		{"past due planned", day(2026, 9, 30), ATRStatusPlanned, now, true, 1},
		{"past due in progress", day(2026, 9, 1), ATRStatusInProgress, now, true, 30},
		{"past due pending review is still open", day(2026, 9, 1), ATRStatusPendingReview, now, true, 30},
		{"past due lowercase in progress", day(2026, 9, 1), "in_progress", now, true, 30},
		{"past due empty status", day(2026, 9, 1), "", now, true, 30},
		{"past due unknown legacy status", day(2026, 9, 1), "Something", now, true, 30},
		{"past due completed", day(2026, 9, 1), ATRStatusCompleted, now, false, 0},
		{"past due lowercase completed", day(2026, 9, 1), "completed", now, false, 0},
		{"past due cancelled", day(2026, 9, 1), ATRStatusCancelled, now, false, 0},
		{"past due mixed case cancelled", day(2026, 9, 1), " Cancelled ", now, false, 0},
		{"due today", day(2026, 10, 1), ATRStatusPlanned, now, false, 0},
		{"future due", day(2026, 10, 2), ATRStatusInProgress, now, false, 0},
		{"far past across year", day(2025, 10, 1), ATRStatusPlanned, now, true, 365},
		{"no due date", nil, ATRStatusPlanned, now, false, 0},
		{"zero due date", &time.Time{}, ATRStatusPlanned, now, false, 0},

		// Postgres may return the stored UTC-midnight instant in another zone;
		// the calendar date of the instant in UTC is what counts
		{"due read back in jakarta zone", func() *time.Time {
			d := day(2026, 9, 30).In(wib)
			return &d
		}(), ATRStatusPlanned, now, true, 1},

		// Jakarta midnight boundary: 30 Sep 23:59:59 WIB is 16:59:59 UTC; 1 Oct
		// 00:00 WIB is 17:00 UTC on 30 Sep. Due 2026-09-30 only becomes overdue
		// at Jakarta midnight, regardless of the clock's zone.
		{"jakarta 23:59:59 on due day", day(2026, 9, 30), ATRStatusPlanned, time.Date(2026, 9, 30, 16, 59, 59, 0, time.UTC), false, 0},
		{"jakarta 00:00 day after due", day(2026, 9, 30), ATRStatusPlanned, time.Date(2026, 9, 30, 17, 0, 0, 0, time.UTC), true, 1},
		{"jakarta midnight given in WIB", day(2026, 9, 30), ATRStatusPlanned, time.Date(2026, 10, 1, 0, 0, 0, 0, wib), true, 1},
		{"utc still previous day but jakarta not", day(2026, 9, 30), ATRStatusPlanned, time.Date(2026, 9, 30, 20, 0, 0, 0, time.UTC), true, 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			overdue, days := ComputeATROverdue(tc.due, tc.status, tc.now)
			if overdue != tc.overdue || days != tc.days {
				t.Fatalf("ComputeATROverdue(%v, %q, %s) = (%v, %d), want (%v, %d)",
					tc.due, tc.status, tc.now.Format(time.RFC3339), overdue, days, tc.overdue, tc.days)
			}
		})
	}
}

func TestParseATRDueDate(t *testing.T) {
	cases := []struct {
		in   string
		want string // YYYY-MM-DD, "" = invalid
	}{
		{"2026-04-15", "2026-04-15"},
		{" 2026-04-15 ", "2026-04-15"},
		// Jakarta midnight of 15 Apr sent as UTC by a browser date picker
		{"2026-04-14T17:00:00Z", "2026-04-15"},
		{"2026-04-15T00:00:00+07:00", "2026-04-15"},
		// UTC midnight is taken as the date itself
		{"2026-04-15T00:00:00Z", "2026-04-15"},
		{"2026-04-15T23:30:00+07:00", "2026-04-15"},
		{"", ""},
		{"15/04/2026", ""},
		{"2026-02-30", ""},
		{"tomorrow", ""},
	}
	for _, tc := range cases {
		got, ok := ParseATRDueDate(tc.in)
		if tc.want == "" {
			if ok {
				t.Errorf("ParseATRDueDate(%q) = %v, want invalid", tc.in, got)
			}
			continue
		}
		if !ok {
			t.Errorf("ParseATRDueDate(%q) invalid, want %s", tc.in, tc.want)
			continue
		}
		if got.Location() != time.UTC || got.Hour() != 0 || got.Minute() != 0 {
			t.Errorf("ParseATRDueDate(%q) = %v, want UTC midnight", tc.in, got)
		}
		if s := got.Format("2006-01-02"); s != tc.want {
			t.Errorf("ParseATRDueDate(%q) = %s, want %s", tc.in, s, tc.want)
		}
	}
}

// The old free-text deadline column is converted into due_date by the
// migration; only "YYYY-MM-DD" (optionally followed by a time) is accepted
func TestParseLegacyATRDeadline(t *testing.T) {
	cases := []struct {
		in   string
		want string // "" = not parsed
	}{
		{"2026-09-30", "2026-09-30"},
		{"2026-09-28T23:00:00Z", "2026-09-28"},
		{"2026-09-30T23:59:59+07:00", "2026-09-30"},
		{"2026-09-29 08:00:00", "2026-09-29"},
		{"  2026-09-01  ", "2026-09-01"},
		{"", ""},
		{"   ", ""},
		{"not a date", ""},
		{"15 Apr 2026", ""},
		{"15/04/2026", ""},
		{"2026-02-30", ""},
		{"2026-09-01x", ""},
	}
	for _, tc := range cases {
		got, ok := ParseLegacyATRDeadline(tc.in)
		if tc.want == "" {
			if ok {
				t.Errorf("ParseLegacyATRDeadline(%q) = %v, want not parsed", tc.in, got)
			}
			continue
		}
		if !ok || got.Format("2006-01-02") != tc.want || got.Location() != time.UTC {
			t.Errorf("ParseLegacyATRDeadline(%q) = (%v, %v), want %s UTC", tc.in, got, ok, tc.want)
		}
	}
}

func TestNormalizeATRStatus(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"PLANNED", "PLANNED", true},
		{"planned", "PLANNED", true},
		{"Planned", "PLANNED", true},
		{"IN_PROGRESS", "IN_PROGRESS", true},
		{"In Progress", "IN_PROGRESS", true},
		{"in progress", "IN_PROGRESS", true},
		{"in_progress", "IN_PROGRESS", true},
		{"in-progress", "IN_PROGRESS", true},
		{"  In   Progress ", "IN_PROGRESS", true},
		{"PENDING_REVIEW", "PENDING_REVIEW", true},
		{"Pending Review", "PENDING_REVIEW", true},
		{"pending-review", "PENDING_REVIEW", true},
		{"Completed", "COMPLETED", true},
		{"completed", "COMPLETED", true},
		{"CANCELLED", "CANCELLED", true},
		{"Cancelled", "CANCELLED", true},
		{"canceled", "CANCELLED", true},
		{"", "", false},
		{"Closed", "", false},
		{"OVERDUE", "", false},
		{"Fieldwork", "", false},
		{"INPROGRESS", "", false},
		{"PLANNED; DROP TABLE x", "", false},
	}
	for _, tc := range cases {
		got, ok := NormalizeATRStatus(tc.in)
		if got != tc.want || ok != tc.ok {
			t.Errorf("NormalizeATRStatus(%q) = (%q, %v), want (%q, %v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestATRHooksUseInjectedClock(t *testing.T) {
	orig := atrNow
	defer func() { atrNow = orig }()
	atrNow = func() time.Time { return time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC) }

	r := &ActionTakenReport{DueDate: day(2026, 9, 21), Status: ATRStatusInProgress, IsOverdue: false, OverdueDays: 99}
	if err := r.AfterFind(nil); err != nil {
		t.Fatal(err)
	}
	if !r.IsOverdue || r.OverdueDays != 10 {
		t.Fatalf("AfterFind: got (%v, %d), want (true, 10)", r.IsOverdue, r.OverdueDays)
	}

	// Stale derived values are overwritten after save
	r = &ActionTakenReport{DueDate: day(2026, 12, 31), Status: ATRStatusPlanned, IsOverdue: true, OverdueDays: 5}
	if err := r.AfterSave(nil); err != nil {
		t.Fatal(err)
	}
	if r.IsOverdue || r.OverdueDays != 0 {
		t.Fatalf("AfterSave: got (%v, %d), want (false, 0)", r.IsOverdue, r.OverdueDays)
	}

	// A completed ATR is never overdue
	r = &ActionTakenReport{DueDate: day(2026, 9, 1), Status: ATRStatusCompleted}
	_ = r.AfterFind(nil)
	if r.IsOverdue || r.OverdueDays != 0 {
		t.Fatalf("completed: got (%v, %d), want (false, 0)", r.IsOverdue, r.OverdueDays)
	}
}

func TestATRDerivedFieldsAreNotColumns(t *testing.T) {
	s, err := schema.Parse(&ActionTakenReport{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range s.Fields {
		if f.Name == "IsOverdue" || f.Name == "OverdueDays" {
			if f.DBName != "" {
				t.Fatalf("%s maps to column %q; it must not be persisted", f.Name, f.DBName)
			}
		}
	}
	for _, col := range []string{"is_overdue", "overdue_days", "days_overdue"} {
		if _, ok := s.FieldsByDBName[col]; ok {
			t.Fatalf("schema has column %q", col)
		}
	}
	// The assignment-letter binding is gone
	for _, col := range []string{"assignment_letter_id", "audit_ref", "audit_finding_id", "deadline"} {
		if _, ok := s.FieldsByDBName[col]; ok {
			t.Fatalf("schema still has legacy column %q", col)
		}
	}
	// New columns exist
	for _, col := range []string{"audit_result_report_id", "finding_id", "pic_user_id", "pic_name", "due_date", "progress", "review_note", "reviewed_by", "reviewed_at"} {
		if _, ok := s.FieldsByDBName[col]; !ok {
			t.Fatalf("schema is missing column %q", col)
		}
	}
}

func TestATRJSONShape(t *testing.T) {
	b, err := json.Marshal(ActionTakenReport{DueDate: day(2026, 9, 1), Status: ATRStatusPlanned, IsOverdue: true, OverdueDays: 30})
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	if m["is_overdue"] != true || m["overdue_days"] != float64(30) || m["status"] != "PLANNED" {
		t.Fatalf("unexpected JSON: %s", b)
	}
	if m["due_date"] != "2026-09-01T00:00:00Z" {
		t.Fatalf("due_date = %v, want RFC3339 UTC midnight", m["due_date"])
	}
	for _, legacy := range []string{"isOverdue", "daysOverdue", "deadline", "assignment_letter_id", "auditRef"} {
		if _, ok := m[legacy]; ok {
			t.Fatalf("JSON still has legacy key %q: %s", legacy, b)
		}
	}
}
