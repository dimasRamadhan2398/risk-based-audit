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
		baseURL = "http://python-ai:8000" // Default for docker-compose
	}
	return &PythonAIClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// --- Requests ---
type XGBoostRequest struct {
	KPIData          float64 `json:"kpi_data"`
	PreviousFindings float64 `json:"previous_findings"`
	MasterData       float64 `json:"master_data"`
}

type IsolationForestRequest struct {
	Feature1 float64 `json:"feature1"`
	Feature2 float64 `json:"feature2"`
}

type TextRequest struct {
	Text string `json:"text"`
}

type LSTMRequest struct {
	HistoricalData []float64 `json:"historical_data"`
}

// --- Responses ---
type XGBoostResponse struct {
	RiskScore         float64            `json:"risk_score"`
	FeatureImportance map[string]float64 `json:"feature_importance"`
}

type IsolationForestResponse struct {
	IsAnomaly    bool    `json:"is_anomaly"`
	AnomalyScore float64 `json:"anomaly_score"`
}

type IndoBERTResponse struct {
	RiskCategory string  `json:"risk_category"`
	Confidence   float64 `json:"confidence"`
	Sentiment    string  `json:"sentiment"`
}

type LSTMResponse struct {
	PredictedPerformance float64 `json:"predicted_performance"`
	Trend                string  `json:"trend"`
}

// --- Methods ---

func (c *PythonAIClient) PredictRiskScore(reqData XGBoostRequest) (*XGBoostResponse, error) {
	respData, err := c.postJSON("/predict/risk-score", reqData)
	if err != nil {
		return nil, err
	}
	var res XGBoostResponse
	if err := json.Unmarshal(respData, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *PythonAIClient) PredictAnomaly(reqData IsolationForestRequest) (*IsolationForestResponse, error) {
	respData, err := c.postJSON("/predict/anomaly", reqData)
	if err != nil {
		return nil, err
	}
	var res IsolationForestResponse
	if err := json.Unmarshal(respData, &res); err != nil {
		return nil, err
	}
	return &res, nil
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

func (c *PythonAIClient) PredictPerformanceTrend(reqData LSTMRequest) (*LSTMResponse, error) {
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
