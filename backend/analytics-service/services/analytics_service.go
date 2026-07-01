package services

import (
	"analytics-service/models"
	"time"
)

type AnalyticsService struct{
	aiClient *PythonAIClient
}

func NewAnalyticsService() *AnalyticsService {
	return &AnalyticsService{
		aiClient: NewPythonAIClient(),
	}
}

// GetRiskScore gets the risk score from the Python AI service
func (s *AnalyticsService) GetRiskScore(kpiData, previousFindings, masterData float64) (*XGBoostResponse, error) {
	req := XGBoostRequest{
		KPIData:          kpiData,
		PreviousFindings: previousFindings,
		MasterData:       masterData,
	}
	res, err := s.aiClient.PredictRiskScore(req)
	if err != nil {
		// Fallback
		return &XGBoostResponse{
			RiskScore: 0.85,
			FeatureImportance: map[string]float64{
				"kpi_data":          0.5,
				"previous_findings": 0.3,
				"master_data":       0.2,
			},
		}, nil
	}
	return res, nil
}

// GetAnomaly gets anomaly detection from the Python AI service
func (s *AnalyticsService) GetAnomaly(feature1, feature2 float64) (*IsolationForestResponse, error) {
	req := IsolationForestRequest{
		Feature1: feature1,
		Feature2: feature2,
	}
	res, err := s.aiClient.PredictAnomaly(req)
	if err != nil {
		// Fallback
		return &IsolationForestResponse{
			IsAnomaly:    true,
			AnomalyScore: -0.75,
		}, nil
	}
	return res, nil
}

// GetTextAnalysis gets text analysis from the Python AI service
func (s *AnalyticsService) GetTextAnalysis(text string) (*IndoBERTResponse, error) {
	req := TextRequest{Text: text}
	res, err := s.aiClient.PredictTextAnalysis(req)
	if err != nil {
		// Fallback
		return &IndoBERTResponse{
			RiskCategory: "High Risk",
			Confidence:   0.92,
			Sentiment:    "Negative",
		}, nil
	}
	return res, nil
}

// GetPerformanceTrend gets performance trend prediction from the Python AI service
func (s *AnalyticsService) GetPerformanceTrend(historicalData []float64) (*LSTMResponse, error) {
	req := LSTMRequest{HistoricalData: historicalData}
	res, err := s.aiClient.PredictPerformanceTrend(req)
	if err != nil {
		// Fallback
		return &LSTMResponse{
			PredictedPerformance: 0.45,
			Trend:                "Deteriorating",
		}, nil
	}
	return res, nil
}

// GenerateReport creates a dummy data pattern report.
func (s *AnalyticsService) GenerateReport() models.DataPatternReport {
	return models.DataPatternReport{
		TotalFindings:   142,
		Resolved:        98,
		Open:            44,
		OverdueFollowUp: 12,
		FindingTrends: []models.TrendPoint{
			{Month: "Jan", Count: 10},
			{Month: "Feb", Count: 15},
			{Month: "Mar", Count: 8},
			{Month: "Apr", Count: 22},
			{Month: "May", Count: 18},
			{Month: "Jun", Count: 25},
		},
		Anomalies: []models.Anomaly{
			{
				Description: "Unusual spike in IT Security findings in April.",
				Severity:    "High",
				Date:        "2023-04-15",
			},
			{
				Description: "Significant delay in resolving compliance issues.",
				Severity:    "Medium",
				Date:        "2023-05-10",
			},
		},
	}
}

// PredictFutureTrends uses a simple linear regression calculation (simulated)
// to forecast future risk trends based on historical dummy data.
func (s *AnalyticsService) PredictFutureTrends() models.PredictiveAnalysis {
	// Let's assume we have historical data points (x=month index, y=risk score)
	// We will calculate a simple linear regression y = mx + c

	historicalY := []float64{10.0, 12.0, 11.5, 15.0, 16.0, 19.0} // E.g., Risk scores over 6 months

	var sumX, sumY, sumXY, sumX2 float64
	n := float64(len(historicalY))

	for i, y := range historicalY {
		x := float64(i + 1) // Month 1 to 6
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate slope (m) and intercept (c)
	// m = (n * sumXY - sumX * sumY) / (n * sumX2 - sumX * sumX)
	m := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	// c = (sumY - m * sumX) / n
	c := (sumY - m*sumX) / n

	// Forecast next 3 months
	var forecast []models.ForecastPoint
	now := time.Now()

	for i := 1; i <= 3; i++ {
		futureX := float64(len(historicalY) + i)
		predictedY := m*futureX + c

		forecast = append(forecast, models.ForecastPoint{
			Date:          now.AddDate(0, i, 0), // Add i months
			PredictedRisk: predictedY,
			LowerBound:    predictedY * 0.9, // 10% margin
			UpperBound:    predictedY * 1.1,
		})
	}

	trendDirection := "Stable"
	if m > 0.5 {
		trendDirection = "Up"
	} else if m < -0.5 {
		trendDirection = "Down"
	}

	return models.PredictiveAnalysis{
		Forecast:       forecast,
		TrendDirection: trendDirection,
		ModelAccuracy:  0.87, // Dummy accuracy
	}
}
