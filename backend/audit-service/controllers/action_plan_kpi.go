package controllers

import (
	"fmt"
	"math"

	"audit-service/models"

	"gorm.io/gorm"
)

// ATRClosureStats are the action taken report counts behind the "Action
// Plan Closed" KPI card, bucketed by the stored status (any casing):
// COMPLETED is closed; PLANNED, IN_PROGRESS, PENDING_REVIEW (and any
// unrecognised legacy value) are open; CANCELLED counts as neither.
type ATRClosureStats struct {
	Completed int64
	Open      int64
	Cancelled int64
}

// Countable is the KPI denominator: every ATR except cancelled ones
func (s ATRClosureStats) Countable() int64 { return s.Completed + s.Open }

// ClosedPercent is Completed / Countable as a percentage with one decimal.
// ok is false when there is nothing to count (no data, not 0%).
func (s ATRClosureStats) ClosedPercent() (pct float64, ok bool) {
	if s.Countable() == 0 {
		return 0, false
	}
	return math.Round(float64(s.Completed)/float64(s.Countable())*1000) / 10, true
}

// LoadATRClosureStats counts the (non-deleted) ATRs per status
func LoadATRClosureStats(db *gorm.DB) (ATRClosureStats, error) {
	var rows []struct {
		Status string
		N      int64
	}
	if err := db.Model(&models.ActionTakenReport{}).
		Select("status, COUNT(*) AS n").
		Group("status").
		Scan(&rows).Error; err != nil {
		return ATRClosureStats{}, err
	}
	var s ATRClosureStats
	for _, r := range rows {
		status, _ := models.NormalizeATRStatus(r.Status)
		switch status {
		case models.ATRStatusCompleted:
			s.Completed += r.N
		case models.ATRStatusCancelled:
			s.Cancelled += r.N
		default:
			s.Open += r.N
		}
	}
	return s, nil
}

// actionPlanClosedCard builds the "Action Plan Closed" card. Without
// countable ATRs, value, gap and trend are empty (the dashboard shows "-")
// and the actual is 0. Sub-metric texts keep the formats the frontend
// translates ("%d open", "%d total recommendations registered").
func actionPlanClosedCard(s ATRClosureStats, target float64, unit string) SummaryCardResponse {
	card := SummaryCardResponse{
		Title:        "Action Plan Closed",
		Key:          "action_plan_closed",
		Target:       formatWithUnit("%.0f", target, unit),
		TargetNumber: target,
		Unit:         unit,
		SubMetrics: []SubMetric{
			{
				Title:  "Open / Pending Action Plans",
				Value:  fmt.Sprintf("%d open", s.Open),
				Target: "0 overdue",
				Trend:  fmt.Sprintf("%d total recommendations registered", s.Countable()),
			},
		},
	}
	actual, ok := s.ClosedPercent()
	if !ok {
		return card
	}
	gap := actual - target
	card.Value = formatWithUnit("%.0f", actual, unit)
	card.ActualNumber = actual
	card.Gap = fmt.Sprintf("%+.1f%%", gap)
	card.Trend = fmt.Sprintf("%+.1f%% vs target", gap)
	card.TrendUp = actual >= target
	return card
}
