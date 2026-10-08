package models

import (
	"encoding/json"
	"testing"
)

// Rows like these exist in production (seed data from 2026-07-23); reading them
// into a strict float64 made GET /activity-plans return 500.
func TestPlannedActivityReadsLegacyNumbers(t *testing.T) {
	raw := `[
		{"id":"a","budgetEstimation":"10,000,000","duration":15,"numberOfAuditors":2},
		{"id":"b","budgetEstimation":"","duration":"","numberOfAuditors":null},
		{"id":"c","budgetEstimation":2500000.5,"duration":"10","numberOfAuditors":"3"},
		{"id":"d","budgetEstimation":"n/a"}
	]`
	var got []PlannedActivity
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	want := []struct{ budget, duration, auditors LenientNumber }{
		{10000000, 15, 2},
		{0, 0, 0},
		{2500000.5, 10, 3},
		{0, 0, 0},
	}
	for i, w := range want {
		g := got[i]
		if g.BudgetEstimation != w.budget || g.Duration != w.duration || g.NumberOfAuditors != w.auditors {
			t.Errorf("row %s: got budget=%v duration=%v auditors=%v, want %v %v %v",
				g.ID, g.BudgetEstimation, g.Duration, g.NumberOfAuditors, w.budget, w.duration, w.auditors)
		}
	}
}

func TestLenientNumberWritesPlainNumber(t *testing.T) {
	out, err := json.Marshal(PlannedActivity{BudgetEstimation: 10000000, Duration: 15})
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatal(err)
	}
	if v, ok := m["budgetEstimation"].(float64); !ok || v != 10000000 {
		t.Errorf("budgetEstimation = %#v, want number 10000000", m["budgetEstimation"])
	}
	if v, ok := m["duration"].(float64); !ok || v != 15 {
		t.Errorf("duration = %#v, want number 15", m["duration"])
	}
}

func TestLenientNumberRejectsInvalidJSON(t *testing.T) {
	var n LenientNumber
	if err := json.Unmarshal([]byte(`{}`), &n); err == nil {
		t.Error("expected an error for a JSON object")
	}
}
