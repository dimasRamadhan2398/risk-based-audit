package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type PythonAIClient struct {
	baseURL string
	client  *http.Client
}

func NewPythonAIClient() *PythonAIClient {
	baseURL := os.Getenv("PYTHON_AI_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000" // Default fallback for local dev
	}
	return &PythonAIClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// --- Requests ---
type DepartmentRiskRequest struct {
	Entity               string  `json:"entity"`
	RiskCategory         string  `json:"risk_category"`
	InherentLikelihood   float64 `json:"inherent_likelihood"`
	InherentImpact       float64 `json:"inherent_impact"`
	AuditFindingsCount   float64 `json:"audit_findings_count"`
	KPIBelowTargetCount  float64 `json:"kpi_below_target_count"`
	KPIVolatility        float64 `json:"kpi_volatility"`
	PreviousRiskScore    float64 `json:"previous_risk_score"`
	AssessmentMonth      int     `json:"assessment_month"`
}

type AnomalyRequest struct {
	Entity           string  `json:"entity"`
	Description      string  `json:"description"`
	Amount           float64 `json:"amount"`
	HourOfDay        int     `json:"hour_of_day"`
	DayOfWeek        int     `json:"day_of_week"`
	IsNewBeneficiary int     `json:"is_new_beneficiary"`
	IsRoundAmount    int     `json:"is_round_amount"`
}

type TextRequest struct {
	Text string `json:"text"`
}

type PerformanceTrendRequest struct {
	KPIName        string    `json:"kpi_name"`
	HistoricalData []float64 `json:"historical_data"`
}

// --- Responses ---
type DepartmentRiskResponse struct {
	Entity              string             `json:"entity"`
	Type                string             `json:"type"`
	PredictedImpact     int                `json:"predicted_impact"`
	PredictedLikelihood int                `json:"predicted_likelihood"`
	PredictedScore      float64            `json:"predicted_score"`
	ActualScore         float64            `json:"actual_score"`
	RiskLevel           string             `json:"risk_level"`
	ActualRiskLevel     string             `json:"actual_risk_level"`
	Confidence          float64            `json:"confidence"`
	Delta               float64            `json:"delta"`
	Trend               string             `json:"trend"`
	FeatureImportance   map[string]float64 `json:"feature_importance"`
}

type AnomalyResponse struct {
	ID                  string  `json:"id"`
	Entity              string  `json:"entity"`
	Type                string  `json:"type"`
	AnomalyScore        float64 `json:"anomaly_score"`
	Description         string  `json:"description"`
	Severity            string  `json:"severity"`
	Date                string  `json:"date"`
	Amount              float64 `json:"amount"`
	IsAnomaly           bool    `json:"is_anomaly"`
	PredictedImpact     int     `json:"predicted_impact"`
	PredictedLikelihood int     `json:"predicted_likelihood"`
	RiskLevel           string  `json:"risk_level"`
}

type IndoBERTResponse struct {
	DocID          string  `json:"docId"`
	Title          string  `json:"title"`
	Source         string  `json:"source"`
	RiskCategory   string  `json:"risk_category"`
	Sentiment      string  `json:"sentiment"`
	Impact         int     `json:"impact"`
	Likelihood     int     `json:"likelihood"`
	SeverityScore  int     `json:"severityScore"`
	Confidence     float64 `json:"confidence"`
	Excerpt        string  `json:"excerpt"`
	RiskLevel      string  `json:"risk_level"`
}

type LSTMResponse struct {
	KPIName              string    `json:"kpi_name"`
	PredictedPerformance float64   `json:"predicted_performance"`
	ForecastSeries       []float64 `json:"forecast_series"`
	Trend                string    `json:"trend"`
	Impact               int       `json:"impact"`
	Likelihood           int       `json:"likelihood"`
	AlertLevel           string    `json:"alert_level"`
	RiskLevel            string    `json:"risk_level"`
}

// --- Methods ---

func (c *PythonAIClient) PredictRiskScore(reqData DepartmentRiskRequest) (*DepartmentRiskResponse, error) {
	respData, err := c.postJSON("/predict/risk-score", reqData)
	if err != nil {
		return nil, err
	}
	var res DepartmentRiskResponse
	if err := json.Unmarshal(respData, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *PythonAIClient) GetRiskScoreBatch() ([]byte, error) {
	return c.getJSON("/predict/risk-score/batch")
}

func (c *PythonAIClient) PredictAnomaly(reqData AnomalyRequest) (*AnomalyResponse, error) {
	respData, err := c.postJSON("/predict/anomaly", reqData)
	if err != nil {
		return nil, err
	}
	var res AnomalyResponse
	if err := json.Unmarshal(respData, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *PythonAIClient) GetAnomalyBatch() ([]byte, error) {
	return c.getJSON("/predict/anomaly/batch")
}

func (c *PythonAIClient) PredictTextAnalysis(reqData TextRequest) (*IndoBERTResponse, error) {
	respData, err := c.postJSON("/predict/text-analysis", reqData)
	if err != nil {
		return nil, err
	}
	var res IndoBERTResponse
	if err := json.Unmarshal(respData, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *PythonAIClient) GetTextAnalysisBatch() ([]byte, error) {
	return c.getJSON("/predict/text-analysis/batch")
}

func (c *PythonAIClient) PredictPerformanceTrend(reqData PerformanceTrendRequest) (*LSTMResponse, error) {
	respData, err := c.postJSON("/predict/performance-trend", reqData)
	if err != nil {
		return nil, err
	}
	var res LSTMResponse
	if err := json.Unmarshal(respData, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *PythonAIClient) GetPerformanceTrendBatch() ([]byte, error) {
	return c.getJSON("/predict/performance-trend/batch")
}

func (c *PythonAIClient) TriggerAutoRetrain() ([]byte, error) {
	return c.postJSON("/retrain/auto", map[string]string{})
}

func (c *PythonAIClient) getJSON(endpoint string) ([]byte, error) {
	req, err := http.NewRequest("GET", c.baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (c *PythonAIClient) postJSON(endpoint string, body interface{}) ([]byte, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
