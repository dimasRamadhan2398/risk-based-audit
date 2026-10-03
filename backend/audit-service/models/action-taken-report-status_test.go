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

func TestComputeATROverdue(t *testing.T) {
	wib := jakarta(t)
	// 1 Oct 2026, 10:00 in Jakarta
	now := time.Date(2026, 10, 1, 10, 0, 0, 0, wib)

	cases := []struct {
		name     string
		deadline string
		status   string
		now      time.Time
		overdue  bool
		days     int
	}{
		{"past deadline planned", "2026-09-30", "PLANNED", now, true, 1},
		{"past deadline in progress", "2026-09-01", "IN_PROGRESS", now, true, 30},
		{"past deadline lowercase in progress", "2026-09-01", "in_progress", now, true, 30},
		{"past deadline padded planned", "2026-09-01", "  planned ", now, true, 30},
		{"past deadline empty status", "2026-09-01", "", now, true, 30},
		{"past deadline unknown status", "2026-09-01", "Something", now, true, 30},
		{"past deadline completed", "2026-09-01", "COMPLETED", now, false, 0},
		{"past deadline lowercase completed", "2026-09-01", "completed", now, false, 0},
		{"past deadline cancelled", "2026-09-01", "CANCELLED", now, false, 0},
		{"past deadline mixed case cancelled", "2026-09-01", " Cancelled ", now, false, 0},
		{"deadline today", "2026-10-01", "PLANNED", now, false, 0},
		{"future deadline", "2026-10-02", "IN_PROGRESS", now, false, 0},
		{"far past across year", "2025-10-01", "PLANNED", now, true, 365},
		{"rfc3339 deadline", "2026-09-28T23:00:00Z", "PLANNED", now, true, 3},
		{"rfc3339 deadline with offset uses date part", "2026-09-30T23:59:59+07:00", "PLANNED", now, true, 1},
		{"datetime with space", "2026-09-29 08:00:00", "PLANNED", now, true, 2},
		{"rfc3339 deadline today", "2026-10-01T00:00:00Z", "PLANNED", now, false, 0},
		{"empty deadline", "", "PLANNED", now, false, 0},
		{"blank deadline", "   ", "PLANNED", now, false, 0},
		{"invalid deadline", "not a date", "PLANNED", now, false, 0},
		{"display format deadline ignored", "15 Apr 2026", "PLANNED", now, false, 0},
		{"dd/mm/yyyy ignored", "15/04/2026", "PLANNED", now, false, 0},
		{"impossible date", "2026-02-30", "PLANNED", now, false, 0},
		{"date with trailing junk", "2026-09-01x", "PLANNED", now, false, 0},

		// Jakarta midnight boundary: 30 Sep 23:59:59 WIB is 16:59:59 UTC; 1 Oct
		// 00:00 WIB is 17:00 UTC on 30 Sep. The deadline 2026-09-30 only becomes
		// overdue at Jakarta midnight, regardless of the clock's zone.
		{"jakarta 23:59:59 on deadline day", "2026-09-30", "PLANNED", time.Date(2026, 9, 30, 16, 59, 59, 0, time.UTC), false, 0},
		{"jakarta 00:00 day after deadline", "2026-09-30", "PLANNED", time.Date(2026, 9, 30, 17, 0, 0, 0, time.UTC), true, 1},
		{"jakarta midnight given in WIB", "2026-09-30", "PLANNED", time.Date(2026, 10, 1, 0, 0, 0, 0, wib), true, 1},
		{"utc still previous day but jakarta not", "2026-09-30", "PLANNED", time.Date(2026, 9, 30, 20, 0, 0, 0, time.UTC), true, 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			overdue, days := ComputeATROverdue(tc.deadline, tc.status, tc.now)
			if overdue != tc.overdue || days != tc.days {
				t.Fatalf("ComputeATROverdue(%q, %q, %s) = (%v, %d), want (%v, %d)",
					tc.deadline, tc.status, tc.now.Format(time.RFC3339), overdue, days, tc.overdue, tc.days)
			}
		})
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

	r := &ActionTakenReport{Deadline: "2026-09-21", Status: "IN_PROGRESS", IsOverdue: false, DaysOverdue: 99}
	if err := r.AfterFind(nil); err != nil {
		t.Fatal(err)
	}
	if !r.IsOverdue || r.DaysOverdue != 10 {
		t.Fatalf("AfterFind: got (%v, %d), want (true, 10)", r.IsOverdue, r.DaysOverdue)
	}

	// A client-supplied value is overwritten after save
	r = &ActionTakenReport{Deadline: "2026-12-31", Status: "PLANNED", IsOverdue: true, DaysOverdue: 5}
	if err := r.AfterSave(nil); err != nil {
		t.Fatal(err)
	}
	if r.IsOverdue || r.DaysOverdue != 0 {
		t.Fatalf("AfterSave: got (%v, %d), want (false, 0)", r.IsOverdue, r.DaysOverdue)
	}
}

func TestATRDerivedFieldsAreNotColumns(t *testing.T) {
	s, err := schema.Parse(&ActionTakenReport{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range s.Fields {
		if f.Name == "IsOverdue" || f.Name == "DaysOverdue" {
			if f.DBName != "" {
				t.Fatalf("%s maps to column %q; it must not be persisted", f.Name, f.DBName)
			}
		}
	}
	for _, col := range []string{"is_overdue", "days_overdue"} {
		if _, ok := s.FieldsByDBName[col]; ok {
			t.Fatalf("schema has column %q", col)
		}
	}
}

func TestATRJSONShape(t *testing.T) {
	b, err := json.Marshal(ActionTakenReport{Deadline: "2026-09-01", Status: "PLANNED", IsOverdue: true, DaysOverdue: 30})
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	if m["isOverdue"] != true || m["daysOverdue"] != float64(30) || m["status"] != "PLANNED" {
		t.Fatalf("unexpected JSON: %s", b)
	}
}
