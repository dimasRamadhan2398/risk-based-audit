package controllers

import (
	"testing"
	"time"

	"audit-service/models"
	"audit-service/pkg/testdb"

	"github.com/google/uuid"
)

func TestLoadATRClosureStats(t *testing.T) {
	db := testdb.New(t, &models.ActionTakenReport{})
	add := func(status string, deleted bool) {
		t.Helper()
		r := models.ActionTakenReport{ID: uuid.New(), Status: status}
		if err := db.Create(&r).Error; err != nil {
			t.Fatal(err)
		}
		if deleted {
			db.Delete(&r)
		}
	}
	for _, s := range []string{"COMPLETED", "COMPLETED", "completed", "PLANNED", "IN_PROGRESS", "PENDING_REVIEW", "In Progress", "CANCELLED", "Cancelled"} {
		add(s, false)
	}
	add("COMPLETED", true) // soft-deleted rows do not count

	s, err := LoadATRClosureStats(db)
	if err != nil {
		t.Fatal(err)
	}
	if s.Completed != 3 || s.Open != 4 || s.Cancelled != 2 || s.Countable() != 7 {
		t.Fatalf("stats %+v", s)
	}
	pct, ok := s.ClosedPercent()
	if !ok || pct != 42.9 {
		t.Fatalf("closed %% = %v (%v), want 42.9", pct, ok)
	}
}

func TestActionPlanClosedCard(t *testing.T) {
	// No ATRs at all, or only cancelled ones: no data, not 87% and not 0%
	for _, s := range []ATRClosureStats{{}, {Cancelled: 4}} {
		card := actionPlanClosedCard(s, 90, "%")
		if card.Value != "" || card.ActualNumber != 0 || card.Gap != "" || card.Trend != "" || card.TrendUp {
			t.Fatalf("no-data card for %+v: %+v", s, card)
		}
		if card.Key != "action_plan_closed" || card.Target != "90%" || card.SubMetrics[0].Value != "0 open" ||
			card.SubMetrics[0].Trend != "0 total recommendations registered" {
			t.Fatalf("no-data card for %+v: %+v", s, card)
		}
	}

	// PENDING_REVIEW counts as open; cancelled out of the denominator
	card := actionPlanClosedCard(ATRClosureStats{Completed: 9, Open: 1, Cancelled: 5}, 90, "%")
	if card.ActualNumber != 90 || card.Value != "90%" || card.Gap != "+0.0%" || !card.TrendUp ||
		card.SubMetrics[0].Value != "1 open" || card.SubMetrics[0].Trend != "10 total recommendations registered" {
		t.Fatalf("card %+v", card)
	}
	card = actionPlanClosedCard(ATRClosureStats{Completed: 1, Open: 3}, 90, "%")
	if card.ActualNumber != 25 || card.TrendUp || card.Gap != "-65.0%" {
		t.Fatalf("card %+v", card)
	}
	_ = time.Now
}
