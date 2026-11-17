package services

// The 5x5 risk level matrix, ported from getRiskLevel in
// frontend/stores/risk-profile.ts (the only other copy). Keep the two in step;
// TestRiskLevelMatrixParity compares them cell by cell.
//
// Levels use the backend codes of risk_level.risk_code (LOW, LOW_MODERATE,
// MODERATE, MODERATE_HIGH, HIGH); the frontend enum values are their
// kebab-case equivalents.

const (
	RiskLevelCodeLow          = "LOW"
	RiskLevelCodeLowModerate  = "LOW_MODERATE"
	RiskLevelCodeModerate     = "MODERATE"
	RiskLevelCodeModerateHigh = "MODERATE_HIGH"
	RiskLevelCodeHigh         = "HIGH"
)

// riskLevelMatrix[likelihood][impact], both 1..5.
var riskLevelMatrix = map[int]map[int]string{
	5: {1: RiskLevelCodeLowModerate, 2: RiskLevelCodeModerate, 3: RiskLevelCodeModerateHigh, 4: RiskLevelCodeHigh, 5: RiskLevelCodeHigh},
	4: {1: RiskLevelCodeLow, 2: RiskLevelCodeLowModerate, 3: RiskLevelCodeModerate, 4: RiskLevelCodeModerateHigh, 5: RiskLevelCodeHigh},
	3: {1: RiskLevelCodeLow, 2: RiskLevelCodeLowModerate, 3: RiskLevelCodeModerate, 4: RiskLevelCodeModerateHigh, 5: RiskLevelCodeHigh},
	2: {1: RiskLevelCodeLow, 2: RiskLevelCodeLowModerate, 3: RiskLevelCodeLowModerate, 4: RiskLevelCodeModerateHigh, 5: RiskLevelCodeHigh},
	1: {1: RiskLevelCodeLow, 2: RiskLevelCodeLow, 3: RiskLevelCodeLowModerate, 4: RiskLevelCodeModerate, 5: RiskLevelCodeHigh},
}

// RiskLevelFor returns the matrix level. Like the frontend, a missing or out
// of range likelihood/impact reads as LOW.
func RiskLevelFor(likelihood, impact int) string {
	if lvl, ok := riskLevelMatrix[likelihood][impact]; ok {
		return lvl
	}
	return RiskLevelCodeLow
}

// WithinAppetite is the risk appetite rule the frontend applies
// (pages/risk-appetite): LOW and LOW_MODERATE are accepted without
// mitigation; anything higher is above appetite.
func WithinAppetite(level string) bool {
	return level == RiskLevelCodeLow || level == RiskLevelCodeLowModerate
}
