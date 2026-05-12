package models

import "time"

// DataPatternReport represents the historical report for data patterns, anomalies, findings, etc.
type DataPatternReport struct {
	TotalFindings   int `json:"total_findings"`
	Resolved        int `json:"resolved"`
	Open            int `json:"open"`
	OverdueFollowUp int `json:"overdue_follow_up"`

	// Trends could be represented as an array of data points for a chart
	FindingTrends []TrendPoint `json:"finding_trends"`

	// Anomalies could be string descriptions or objects
	Anomalies []Anomaly `json:"anomalies"`
}

type TrendPoint struct {
	Month string `json:"month"`
	Count int    `json:"count"`
}

type Anomaly struct {
	Description string `json:"description"`
	Severity    string `json:"severity"` // e.g., "High", "Medium"
	Date        string `json:"date"`
}

// PredictiveAnalysis represents future risk trends
type PredictiveAnalysis struct {
	// The forecasted future points
	Forecast []ForecastPoint `json:"forecast"`

	// Overall risk trend direction
	TrendDirection string `json:"trend_direction"` // "Up", "Down", "Stable"

	ModelAccuracy float64 `json:"model_accuracy"`
}

type ForecastPoint struct {
	Date          time.Time `json:"date"`
	PredictedRisk float64   `json:"predicted_risk"`
	LowerBound    float64   `json:"lower_bound"`
	UpperBound    float64   `json:"upper_bound"`
}
