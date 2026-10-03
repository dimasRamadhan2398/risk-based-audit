package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// ATR statuses as stored in action_taken_reports.status
const (
	ATRStatusPlanned    = "PLANNED"
	ATRStatusInProgress = "IN_PROGRESS"
	ATRStatusCompleted  = "COMPLETED"
	ATRStatusCancelled  = "CANCELLED"
)

// ATRStatuses is the set of statuses an ATR may be saved with
var ATRStatuses = []string{ATRStatusPlanned, ATRStatusInProgress, ATRStatusCompleted, ATRStatusCancelled}

// NormalizeATRStatus maps any casing/spelling of a known status ("In Progress",
// "in-progress", "completed", ...) to its stored upper snake form. ok is false
// for unknown values.
func NormalizeATRStatus(raw string) (status string, ok bool) {
	s := strings.ToUpper(strings.TrimSpace(raw))
	s = strings.Join(strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == '_' || r == '-' || r == '\t'
	}), "_")
	if s == "CANCELED" {
		s = ATRStatusCancelled
	}
	for _, known := range ATRStatuses {
		if s == known {
			return s, true
		}
	}
	return "", false
}

// atrLocation is the timezone whose calendar day decides whether an ATR deadline
// has passed. Jakarta has no DST, so the fixed +07:00 fallback is exact when
// tzdata is missing.
var atrLocation = func() *time.Location {
	if loc, err := time.LoadLocation("Asia/Jakarta"); err == nil {
		return loc
	}
	return time.FixedZone("WIB", 7*60*60)
}()

// atrNow is the clock used for the overdue computation; tests replace it
var atrNow = time.Now

// parseATRDeadline extracts the calendar date of a deadline stored as
// "YYYY-MM-DD" or an RFC3339-like "YYYY-MM-DDTHH:MM:SS..." (date part used).
func parseATRDeadline(deadline string) (time.Time, bool) {
	s := strings.TrimSpace(deadline)
	if len(s) < 10 {
		return time.Time{}, false
	}
	if len(s) > 10 && s[10] != 'T' && s[10] != 't' && s[10] != ' ' {
		return time.Time{}, false
	}
	d, err := time.Parse("2006-01-02", s[:10])
	if err != nil {
		return time.Time{}, false
	}
	return d, true
}

// ComputeATROverdue reports whether an ATR is overdue at now: its deadline is a
// valid date strictly before today's date in Asia/Jakarta and its status is not
// COMPLETED or CANCELLED. days is the number of whole days past the deadline
// (0 when not overdue).
func ComputeATROverdue(deadline, status string, now time.Time) (overdue bool, days int) {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case ATRStatusCompleted, ATRStatusCancelled:
		return false, 0
	}
	due, ok := parseATRDeadline(deadline)
	if !ok {
		return false, 0
	}
	y, m, d := now.In(atrLocation).Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	if !due.Before(today) {
		return false, 0
	}
	return true, int(today.Sub(due).Hours() / 24)
}

// fillOverdue sets the derived, non-persisted overdue fields
func (r *ActionTakenReport) fillOverdue() {
	r.IsOverdue, r.DaysOverdue = ComputeATROverdue(r.Deadline, r.Status, atrNow())
}

// AfterFind fills the derived overdue fields on every load (incl. preloads)
func (r *ActionTakenReport) AfterFind(tx *gorm.DB) error {
	r.fillOverdue()
	return nil
}

// AfterSave fills the derived overdue fields so create responses carry them and
// any client-sent value is overwritten
func (r *ActionTakenReport) AfterSave(tx *gorm.DB) error {
	r.fillOverdue()
	return nil
}
