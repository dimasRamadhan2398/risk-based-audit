package models

import "math"

// Risk levels of an audit universe year row. An unscored row has the empty
// level "" (see AuditUniverseYear.ScoredAt).
const (
	RiskLevelHigh         = "High"
	RiskLevelMediumToHigh = "Medium to High"
	RiskLevelMedium       = "Medium"
	RiskLevelLowToMedium  = "Low to Medium"
	RiskLevelLow          = "Low"
)

// ClassifyRiskIndex is the single place for the risk-level thresholds. It
// rounds the index to one decimal first and classifies the rounded value, so
// the index that is stored and shown always agrees with its level: 59.95 is
// stored as 60.0 and is therefore "Medium to High", never "Medium".
// Priority is true for "High" and "Medium to High".
func ClassifyRiskIndex(index float64) (rounded float64, level string, priority bool) {
	rounded = math.Round(index*10) / 10
	switch {
	case rounded >= 80:
		return rounded, RiskLevelHigh, true
	case rounded >= 60:
		return rounded, RiskLevelMediumToHigh, true
	case rounded >= 40:
		return rounded, RiskLevelMedium, false
	case rounded >= 20:
		return rounded, RiskLevelLowToMedium, false
	default:
		return rounded, RiskLevelLow, false
	}
}

// ScoredFactor is one risk factor score (1-5) with the factor's current
// corporate weight as a fraction (0.15 = 15%).
type ScoredFactor struct {
	Weight float64
	Score  int
}

// ComputeRiskIndex returns the unrounded risk index (0-100) of a year row:
//
//	index = (sum(weight*score) / sum(weight)) / 5 * 100
//
// over the factors that HAVE a score. A corporate factor without a score row
// (for example one added after the entity was scored) is "not scored" for that
// entity and is left out of both sums; it is never assumed to be a 3. When
// every factor is scored the weights sum to 1 and this is the original
// sum(weight*score)/5*100.
//
// ok is false when no factor with a positive weight has a score: the row then
// has nothing to rank on and must be treated as not scored.
func ComputeRiskIndex(factors []ScoredFactor) (index float64, ok bool) {
	var weighted, weights float64
	for _, f := range factors {
		if f.Weight <= 0 {
			continue
		}
		weighted += f.Weight * float64(f.Score)
		weights += f.Weight
	}
	if weights <= 0 {
		return 0, false
	}
	return weighted / weights / 5.0 * 100.0, true
}
