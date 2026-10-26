package controllers

import (
	"reflect"
	"testing"
	"time"

	"audit-service/models"
	"audit-service/pkg/activitycode/sqlitetest"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestTrendWindowEndsAtCurrentMonth(t *testing.T) {
	now := time.Date(2026, time.February, 15, 10, 0, 0, 0, time.UTC)
	periods := func(year int) []string {
		var out []string
		for _, m := range trendWindow(year, now) {
			out = append(out, m.start().Format("2006-01"))
		}
		return out
	}
	cases := map[int][]string{
		2026: {"2025-09", "2025-10", "2025-11", "2025-12", "2026-01", "2026-02"}, // current year: reaches into the previous one
		2024: {"2024-07", "2024-08", "2024-09", "2024-10", "2024-11", "2024-12"}, // past year: ends at December
		2027: {"2025-09", "2025-10", "2025-11", "2025-12", "2026-01", "2026-02"}, // future year: nothing stored, ends now
	}
	for year, want := range cases {
		if got := periods(year); !reflect.DeepEqual(got, want) {
			t.Errorf("year %d: %v, want %v", year, got, want)
		}
	}
}

func trendsDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := sqlitetest.Open(t, &models.AuditCompletionSnapshot{}, &models.AuditResultReport{}, &models.AssignmentLetter{}, &models.AuditeeSurvey{})
	return db.Session(&gorm.Session{SkipHooks: true})
}

func mustCreate(t *testing.T, db *gorm.DB, v interface{}) {
	t.Helper()
	if err := db.Create(v).Error; err != nil {
		t.Fatal(err)
	}
}

func TestBuildMonthlyTrendsUsesOnlyStoredData(t *testing.T) {
	db := trendsDB(t)
	now := time.Date(2026, time.October, 21, 12, 0, 0, 0, time.UTC)
	at := func(m time.Month, d int) time.Time { return time.Date(2026, m, d, 9, 0, 0, 0, time.UTC) }

	// Completion: the last snapshot of a month wins; March is outside the window (May..Oct).
	for _, s := range []models.AuditCompletionSnapshot{
		{ID: uuid.New(), Year: 2026, Month: 3, SnapshotAt: at(time.March, 1), PlanBasedRate: 10},
		{ID: uuid.New(), Year: 2026, Month: 8, SnapshotAt: at(time.August, 1), PlanBasedRate: 40},
		{ID: uuid.New(), Year: 2026, Month: 8, SnapshotAt: at(time.August, 30), PlanBasedRate: 45.26},
		{ID: uuid.New(), Year: 2026, Month: 10, SnapshotAt: at(time.October, 20), PlanBasedRate: 60},
	} {
		mustCreate(t, db, &s)
	}

	// Timeliness in September: one report 5 days after finish (on time), one 20 days after (late).
	mustCreate(t, db, &models.AssignmentLetter{ID: uuid.New(), LetterNumber: "L-1", FinishPeriod: "2026-09-01"})
	mustCreate(t, db, &models.AssignmentLetter{ID: uuid.New(), LetterNumber: "L-2", FinishPeriod: "2026-09-01"})
	for letter, day := range map[string]int{"L-1": 6, "L-2": 21} {
		d := at(time.September, day)
		mustCreate(t, db, &models.AuditResultReport{ID: uuid.New(), AssignmentLetterID: letter, ReportDate: &d})
	}

	// CSAT in June: average of 4 and 5; and one survey of January 2026, outside the window.
	for _, s := range []struct {
		month int
		score float64
	}{{6, 4}, {6, 5}, {1, 1}} {
		mustCreate(t, db, &models.AuditeeSurvey{ID: uuid.New(), AuditeeName: "x", Year: 2026, Month: s.month, OverallScore: s.score})
	}

	resp, err := buildMonthlyTrends(db, 2026, now)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"May", "Jun", "Jul", "Aug", "Sep", "Oct"}; !reflect.DeepEqual(resp.Labels, want) {
		t.Fatalf("labels %v, want %v", resp.Labels, want)
	}
	if resp.Periods[0] != "2026-05" || resp.Periods[5] != "2026-10" {
		t.Fatalf("periods %v", resp.Periods)
	}
	check := func(name string, series []*float64, want []interface{}) {
		t.Helper()
		for i, w := range want {
			got := series[i]
			switch {
			case w == nil && got != nil:
				t.Errorf("%s[%s] = %v, want null", name, resp.Labels[i], *got)
			case w != nil && (got == nil || *got != w.(float64)):
				t.Errorf("%s[%s] = %v, want %v", name, resp.Labels[i], got, w)
			}
		}
	}
	check("completion", resp.CompletionRateSeries, []interface{}{nil, nil, nil, 45.3, nil, 60.0})
	check("timeliness", resp.TimelinessSeries, []interface{}{nil, nil, nil, nil, 50.0, nil})
	check("csat", resp.CsatSeries, []interface{}{nil, 4.5, nil, nil, nil, nil})
}

func TestBuildMonthlyTrendsEmptyIsAllNull(t *testing.T) {
	resp, err := buildMonthlyTrends(trendsDB(t), 2026, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for _, series := range [][]*float64{resp.CompletionRateSeries, resp.TimelinessSeries, resp.CsatSeries} {
		if len(series) != TrendWindowMonths {
			t.Fatalf("len %d, want %d", len(series), TrendWindowMonths)
		}
		for _, v := range series {
			if v != nil {
				t.Fatalf("value %v with no stored data", *v)
			}
		}
	}
}
