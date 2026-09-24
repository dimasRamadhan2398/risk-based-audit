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

type DataHubClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewDataHubClient() *DataHubClient {
	baseURL := os.Getenv("DATA_HUB_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8100" // Default Data Hub API port
	}
	apiKey := os.Getenv("DATA_HUB_API_KEY")
	if apiKey == "" {
		apiKey = "dev-api-key"
	}
	return &DataHubClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ─── CAATT Data Models ──────────────────────────────────────────────────────

type FullPopulationResponse struct {
	Data    []map[string]interface{} `json:"data"`
	Summary map[string]interface{}    `json:"summary"`
}

type DuplicateGapResponse struct {
	Data    []map[string]interface{} `json:"data"`
	Summary map[string]interface{}    `json:"summary"`
}

type BenfordResponse struct {
	Data    []map[string]interface{} `json:"data"`
	Summary map[string]interface{}    `json:"summary"`
}

type StratificationResponse struct {
	Data    []map[string]interface{} `json:"data"`
	Summary map[string]interface{}    `json:"summary"`
}

type ReconciliationResponse struct {
	Data    []map[string]interface{} `json:"data"`
	Summary map[string]interface{}    `json:"summary"`
}

type PolicyViolationsResponse struct {
	Data    []map[string]interface{} `json:"data"`
	Summary map[string]interface{}    `json:"summary"`
}

type DataQualityResponse struct {
	Data                []map[string]interface{} `json:"data"`
	Summary             map[string]interface{}    `json:"summary"`
	OverallQualityScore float64                  `json:"overall_quality_score"`
}

// ─── Client Methods ─────────────────────────────────────────────────────────

func (c *DataHubClient) GetFullPopulation(limit int) (*FullPopulationResponse, error) {
	endpoint := fmt.Sprintf("/api/v1/caatt/full-population?limit=%d", limit)
	raw, err := c.getJSON(endpoint)
	if err != nil {
		return nil, err
	}
	var res FullPopulationResponse
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *DataHubClient) GetDuplicateGap(resultType string, limit int) (*DuplicateGapResponse, error) {
	endpoint := fmt.Sprintf("/api/v1/caatt/duplicate-gap?limit=%d", limit)
	if resultType != "" {
		endpoint += fmt.Sprintf("&result_type=%s", resultType)
	}
	raw, err := c.getJSON(endpoint)
	if err != nil {
		return nil, err
	}
	var res DuplicateGapResponse
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *DataHubClient) GetBenfordAnalysis() (*BenfordResponse, error) {
	raw, err := c.getJSON("/api/v1/caatt/benford-analysis")
	if err != nil {
		return nil, err
	}
	var res BenfordResponse
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *DataHubClient) GetStratification(category string) (*StratificationResponse, error) {
	endpoint := "/api/v1/caatt/stratification"
	if category != "" {
		endpoint += fmt.Sprintf("?category=%s", category)
	}
	raw, err := c.getJSON(endpoint)
	if err != nil {
		return nil, err
	}
	var res StratificationResponse
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *DataHubClient) GetReconciliation() (*ReconciliationResponse, error) {
	raw, err := c.getJSON("/api/v1/caatt/reconciliation")
	if err != nil {
		return nil, err
	}
	var res ReconciliationResponse
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *DataHubClient) GetPolicyViolations(severity string, limit int) (*PolicyViolationsResponse, error) {
	endpoint := fmt.Sprintf("/api/v1/caatt/policy-violations?limit=%d", limit)
	if severity != "" {
		endpoint += fmt.Sprintf("&severity=%s", severity)
	}
	raw, err := c.getJSON(endpoint)
	if err != nil {
		return nil, err
	}
	var res PolicyViolationsResponse
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *DataHubClient) GetDataQualityMetrics() (*DataQualityResponse, error) {
	raw, err := c.getJSON("/api/v1/caatt/data-quality")
	if err != nil {
		return nil, err
	}
	var res DataQualityResponse
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// ─── HTTP Utilities ─────────────────────────────────────────────────────────

func (c *DataHubClient) getJSON(endpoint string) ([]byte, error) {
	req, err := http.NewRequest("GET", c.baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Data Hub API returned status %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

func (c *DataHubClient) postJSON(endpoint string, body interface{}) ([]byte, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Data Hub API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return io.ReadAll(resp.Body)
}
