package services

import (
	"testing"
)

func TestPredictFutureTrends(t *testing.T) {
	service := NewAnalyticsService()
	result := service.PredictFutureTrends()

	if result.ModelAccuracy != 0.87 {
		t.Errorf("Expected model accuracy 0.87, got %v", result.ModelAccuracy)
	}

	if len(result.Forecast) != 3 {
		t.Errorf("Expected 3 forecast points, got %d", len(result.Forecast))
	}

	if result.TrendDirection != "Up" && result.TrendDirection != "Down" && result.TrendDirection != "Stable" {
		t.Errorf("Expected valid TrendDirection, got %v", result.TrendDirection)
	}
}

func TestGenerateReport(t *testing.T) {
	service := NewAnalyticsService()
	report := service.GenerateReport()

	if report.TotalFindings != 142 {
		t.Errorf("Expected total findings 142, got %d", report.TotalFindings)
	}

	if len(report.FindingTrends) != 6 {
		t.Errorf("Expected 6 trend points, got %d", len(report.FindingTrends))
	}
}
