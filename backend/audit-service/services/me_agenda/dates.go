package me_agenda

import (
	"strings"
	"time"
)

// dayFormats are the layouts date strings are stored in (letter periods and
// fieldwork dates are free varchar columns).
var dayFormats = []string{
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02",
	"02/01/2006",
}

// ParseDay reads a stored date string as a calendar day (UTC midnight). An
// RFC3339 value keeps its written date, with no time-zone shift, which is how
// the frontend's toCalendarDay reads it too.
func ParseDay(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	for _, f := range dayFormats {
		if t, err := time.Parse(f, s); err == nil {
			y, m, d := t.Date()
			return time.Date(y, m, d, 0, 0, 0, 0, time.UTC), true
		}
	}
	if len(s) >= 10 {
		if t, err := time.Parse("2006-01-02", s[:10]); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// FormatDay renders a day as YYYY-MM-DD.
func FormatDay(t time.Time) string { return t.Format("2006-01-02") }

func dayPtr(t time.Time, ok bool) *time.Time {
	if !ok {
		return nil
	}
	return &t
}

// daysBetween is the whole number of days from a to b (b - a).
func daysBetween(a, b time.Time) int { return int(b.Sub(a).Hours() / 24) }
