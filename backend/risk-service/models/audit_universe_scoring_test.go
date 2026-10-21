package models

import (
	"math"
	"testing"
)

func TestClassifyRiskIndexRoundsBeforeClassifying(t *testing.T) {
	cases := []struct {
		in       float64
		rounded  float64
		level    string
		priority bool
	}{
		{0, 0, RiskLevelLow, false},
		{19.94, 19.9, RiskLevelLow, false},
		{19.95, 20, RiskLevelLowToMedium, false},
		{39.95, 40, RiskLevelMedium, false},
		{59.94, 59.9, RiskLevelMedium, false},
		{59.95, 60, RiskLevelMediumToHigh, true}, // stored as 60.0, so never "Medium"
		{60, 60, RiskLevelMediumToHigh, true},
		{79.94, 79.9, RiskLevelMediumToHigh, true},
		{79.95, 80, RiskLevelHigh, true},
		{80, 80, RiskLevelHigh, true},
		{100, 100, RiskLevelHigh, true},
	}
	for _, c := range cases {
		rounded, level, priority := ClassifyRiskIndex(c.in)
		if math.Abs(rounded-c.rounded) > 1e-9 || level != c.level || priority != c.priority {
			t.Errorf("ClassifyRiskIndex(%v) = (%v, %q, %v), want (%v, %q, %v)",
				c.in, rounded, level, priority, c.rounded, c.level, c.priority)
		}
	}
}

func TestComputeRiskIndexUsesOnlyScoredFactors(t *testing.T) {
	// All factors scored: weights sum to 1, same as sum(w*s)/5*100.
	idx, ok := ComputeRiskIndex([]ScoredFactor{{Weight: 0.6, Score: 5}, {Weight: 0.4, Score: 3}})
	if !ok || math.Abs(idx-84) > 1e-9 {
		t.Fatalf("full set: got (%v, %v), want (84, true)", idx, ok)
	}

	// A factor without a score is left out, not counted as 3 or as 0.
	idx, ok = ComputeRiskIndex([]ScoredFactor{{Weight: 0.6, Score: 5}})
	if !ok || math.Abs(idx-100) > 1e-9 {
		t.Fatalf("partial set: got (%v, %v), want (100, true)", idx, ok)
	}

	if _, ok := ComputeRiskIndex(nil); ok {
		t.Fatal("no scores must report not scored")
	}
	if _, ok := ComputeRiskIndex([]ScoredFactor{{Weight: 0, Score: 5}}); ok {
		t.Fatal("only zero-weight scores must report not scored")
	}
}
