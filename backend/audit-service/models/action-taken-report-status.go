package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// ATR statuses as stored in action_taken_reports.status
//
//	PLANNED         created when the LHA is approved
//	IN_PROGRESS     the PIC saved an action plan (or the reviewer rejected)
//	PENDING_REVIEW  the PIC submitted; waiting for an auditor's review
//	COMPLETED       an auditor verified the action
//	CANCELLED       cancelled by a manager/CAE/admin
const (
	ATRStatusPlanned       = "PLANNED"
	ATRStatusInProgress    = "IN_PROGRESS"
	ATRStatusPendingReview = "PENDING_REVIEW"
	ATRStatusCompleted     = "COMPLETED"
	ATRStatusCancelled     = "CANCELLED"
)

// ATRStatuses is the set of statuses an ATR may have
var ATRStatuses = []string{ATRStatusPlanned, ATRStatusInProgress, ATRStatusPendingReview, ATRStatusCompleted, ATRStatusCancelled}

// ATROpenStatuses are the statuses of an ATR whose follow-up is not finished
var ATROpenStatuses = []string{ATRStatusPlanned, ATRStatusInProgress, ATRStatusPendingReview}

// NormalizeATRStatus maps any casing/spelling of a known status ("In Progress",
// "pending-review", "completed", ...) to its stored upper snake form. ok is
// false for unknown values.
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

// IsATRClosed reports whether status is COMPLETED or CANCELLED (any casing)
func IsATRClosed(status string) bool {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case ATRStatusCompleted, ATRStatusCancelled:
		return true
	}
	return false
}

// atrLocation is the timezone whose calendar day decides whether an ATR due
// date has passed. Jakarta has no DST, so the fixed +07:00 fallback is exact
// when tzdata is missing.
var atrLocation = func() *time.Location {
	if loc, err := time.LoadLocation("Asia/Jakarta"); err == nil {
		return loc
	}
	return time.FixedZone("WIB", 7*60*60)
}()

// atrNow is the clock used for the overdue computation; tests replace it
var atrNow = time.Now

// ATRToday returns today's date in Asia/Jakarta as UTC midnight, the form due
// dates are stored in. An open ATR is overdue when due_date < ATRToday(now).
func ATRToday(now time.Time) time.Time {
	y, m, d := now.In(atrLocation).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// ATRDueDate normalises a due date to the stored form: the calendar date as
// UTC midnight. A date-only value is taken as is; an instant (RFC3339) is
// read in Asia/Jakarta, so "2026-04-14T17:00:00Z" (Jakarta midnight) is 15 Apr.
func ATRDueDate(t time.Time) time.Time {
	if t.Location() == time.UTC && t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 && t.Nanosecond() == 0 {
		return t
	}
	return ATRToday(t)
}

// ParseATRDueDate parses a due date sent by a client: "YYYY-MM-DD" or RFC3339
func ParseATRDueDate(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if d, err := time.Parse("2006-01-02", s); err == nil {
		return d, true
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return ATRDueDate(t), true
	}
	return time.Time{}, false
}

// ParseLegacyATRDeadline extracts the calendar date of an old free-text
// deadline stored as "YYYY-MM-DD" or "YYYY-MM-DDTHH:MM:SS..." (date part
// used). Used when migrating the old deadline column into due_date.
func ParseLegacyATRDeadline(deadline string) (time.Time, bool) {
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

// ComputeATROverdue reports whether an ATR is overdue at now: it has a due
// date strictly before today's date in Asia/Jakarta and its status is not
// COMPLETED or CANCELLED. days is the number of whole days past the due date
// (0 when not overdue).
func ComputeATROverdue(due *time.Time, status string, now time.Time) (overdue bool, days int) {
	if due == nil || due.IsZero() || IsATRClosed(status) {
		return false, 0
	}
	y, m, d := due.UTC().Date()
	dueDay := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	today := ATRToday(now)
	if !dueDay.Before(today) {
		return false, 0
	}
	return true, int(today.Sub(dueDay).Hours() / 24)
}

// FillOverdue sets the derived, non-persisted overdue fields
func (r *ActionTakenReport) FillOverdue() {
	r.IsOverdue, r.OverdueDays = ComputeATROverdue(r.DueDate, r.Status, atrNow())
}

// AfterFind fills the derived overdue fields on every load (incl. preloads)
func (r *ActionTakenReport) AfterFind(tx *gorm.DB) error {
	r.FillOverdue()
	return nil
}

// AfterSave fills the derived overdue fields so write responses carry them
func (r *ActionTakenReport) AfterSave(tx *gorm.DB) error {
	r.FillOverdue()
	return nil
}
