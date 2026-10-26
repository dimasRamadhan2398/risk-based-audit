package controllers

import (
	"fmt"
	"math"
	"strings"
	"time"

	"audit-service/models"

	"gorm.io/gorm"
)

// TrendWindowMonths is how many months the KPI trend charts show.
const TrendWindowMonths = 6

// ReportTimelinessSLADays is how many days after the assignment letter's
// finish period a report may be dated and still count as on time.
const ReportTimelinessSLADays = 14.0

// trendMonth is one calendar month of the trend window.
type trendMonth struct {
	Year  int
	Month time.Month
}

func (m trendMonth) start() time.Time {
	return time.Date(m.Year, m.Month, 1, 0, 0, 0, 0, time.UTC)
}

func (m trendMonth) key() int { return m.Year*100 + int(m.Month) }

// trendWindow is the TrendWindowMonths months, oldest first, ending at the
// current month or, for a past year, at December of that year. A future year
// has nothing stored yet, so it also ends at the current month.
func trendWindow(year int, now time.Time) []trendMonth {
	end := trendMonth{now.Year(), now.Month()}
	if year < now.Year() {
		end = trendMonth{year, time.December}
	}
	months := make([]trendMonth, TrendWindowMonths)
	for i := range months {
		t := end.start().AddDate(0, i-(TrendWindowMonths-1), 0)
		months[i] = trendMonth{t.Year(), t.Month()}
	}
	return months
}

// reportDaysAfterFinish is how many days after the finish period (YYYY-MM-DD)
// a report was dated. ok is false when the finish period does not parse or the
// report is dated before it; such reports are not counted.
func reportDaysAfterFinish(reportDate time.Time, finishPeriod string) (days float64, ok bool) {
	finish, err := time.Parse("2006-01-02", strings.TrimSpace(finishPeriod))
	if err != nil {
		return 0, false
	}
	days = reportDate.Sub(finish).Hours() / 24.0
	return days, days >= 0
}

func roundTo(v float64, decimals int) *float64 {
	p := math.Pow(10, float64(decimals))
	r := math.Round(v*p) / p
	return &r
}

// buildMonthlyTrends computes each month of the trend window from stored data:
//
//   - completion: plan-based completion rate of the month's last
//     audit_completion_snapshots row (the analyzer writes one every few minutes)
//   - timeliness: share of audit result reports dated in the month that were
//     at most ReportTimelinessSLADays after their letter's finish period
//   - csat: average overall_score of the month's auditee surveys
func buildMonthlyTrends(db *gorm.DB, year int, now time.Time) (MonthlyTrendResponse, error) {
	months := trendWindow(year, now)
	first, last := months[0], months[len(months)-1]
	from, to := first.start(), last.start().AddDate(0, 1, 0)
	index := make(map[int]int, len(months))
	resp := MonthlyTrendResponse{
		Labels:               make([]string, len(months)),
		Periods:              make([]string, len(months)),
		CompletionRateSeries: make([]*float64, len(months)),
		TimelinessSeries:     make([]*float64, len(months)),
		CsatSeries:           make([]*float64, len(months)),
	}
	for i, m := range months {
		index[m.key()] = i
		resp.Labels[i] = m.Month.String()[:3]
		resp.Periods[i] = fmt.Sprintf("%04d-%02d", m.Year, int(m.Month))
	}
	slot := func(t time.Time) (int, bool) {
		t = t.UTC()
		i, ok := index[t.Year()*100+int(t.Month())]
		return i, ok
	}

	// Completion: the last snapshot of each month.
	var snaps []models.AuditCompletionSnapshot
	if err := db.Where("snapshot_at >= ? AND snapshot_at < ?", from, to).
		Order("snapshot_at ASC").Find(&snaps).Error; err != nil {
		return resp, fmt.Errorf("read completion snapshots: %w", err)
	}
	for _, s := range snaps {
		if i, ok := slot(s.SnapshotAt); ok {
			resp.CompletionRateSeries[i] = roundTo(s.PlanBasedRate, 1)
		}
	}

	// Timeliness: reports by the month of their report date.
	var reports []struct {
		ReportDate   *time.Time
		FinishPeriod string
	}
	if err := db.Model(&models.AuditResultReport{}).
		Select("audit_result_reports.report_date, assignment_letters.finish_period").
		Joins("JOIN assignment_letters ON assignment_letters.letter_number = audit_result_reports.assignment_letter_id").
		Where("audit_result_reports.report_date >= ? AND audit_result_reports.report_date < ?", from, to).
		Scan(&reports).Error; err != nil {
		return resp, fmt.Errorf("read report dates: %w", err)
	}
	onTime := make([]int, len(months))
	counted := make([]int, len(months))
	for _, r := range reports {
		if r.ReportDate == nil {
			continue
		}
		i, ok := slot(*r.ReportDate)
		if !ok {
			continue
		}
		days, ok := reportDaysAfterFinish(*r.ReportDate, r.FinishPeriod)
		if !ok {
			continue
		}
		counted[i]++
		if days <= ReportTimelinessSLADays {
			onTime[i]++
		}
	}
	for i := range months {
		if counted[i] > 0 {
			resp.TimelinessSeries[i] = roundTo(float64(onTime[i])/float64(counted[i])*100, 1)
		}
	}

	// CSAT: surveys are filed under a year and month.
	var csat []struct {
		Year  int
		Month int
		Avg   float64
	}
	if err := db.Model(&models.AuditeeSurvey{}).
		Select("year, month, AVG(overall_score) AS avg").
		Where("year * 100 + month BETWEEN ? AND ?", first.key(), last.key()).
		Group("year, month").
		Scan(&csat).Error; err != nil {
		return resp, fmt.Errorf("read survey scores: %w", err)
	}
	for _, c := range csat {
		if i, ok := index[c.Year*100+c.Month]; ok && c.Avg > 0 {
			resp.CsatSeries[i] = roundTo(c.Avg, 1)
		}
	}

	return resp, nil
}
