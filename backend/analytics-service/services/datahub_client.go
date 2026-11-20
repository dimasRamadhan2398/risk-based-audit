package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type DataHubClient struct {
	baseURL     string
	initBaseURL string
	apiKey      string
	client      *http.Client
	mu          sync.RWMutex
	candidates  []string
}

func buildAnalyticsDataHubCandidates(primary string) []string {
	primary = strings.TrimRight(strings.TrimSpace(primary), "/")
	seen := map[string]bool{}
	var out []string
	add := func(u string) {
		u = strings.TrimRight(strings.TrimSpace(u), "/")
		if u != "" && !seen[u] {
			seen[u] = true
			out = append(out, u)
		}
	}
	if primary != "" {
		add(primary)
	}
	if primary == "" || strings.HasSuffix(primary, ":8100") {
		add("http://host.docker.internal:8100")
		add("http://localhost:8100")
		add("http://127.0.0.1:8100")
	}
	return out
}

func NewDataHubClient() *DataHubClient {
	baseURL := os.Getenv("DATA_HUB_URL")
	if baseURL == "" {
		baseURL = os.Getenv("DATA_HUB_BASE_URL")
	}
	if baseURL == "" {
		baseURL = "http://localhost:8100" // Default Data Hub API port
	}
	apiKey := os.Getenv("DATA_HUB_API_KEY")
	if apiKey == "" {
		apiKey = "dev-api-key"
	}
	candidates := buildAnalyticsDataHubCandidates(baseURL)
	return &DataHubClient{
		baseURL:     candidates[0],
		initBaseURL: candidates[0],
		apiKey:      apiKey,
		candidates:  candidates,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *DataHubClient) candidateURLs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.baseURL != "" && c.baseURL != c.initBaseURL {
		return []string{strings.TrimRight(c.baseURL, "/")}
	}
	if len(c.candidates) > 0 {
		out := make([]string, len(c.candidates))
		copy(out, c.candidates)
		return out
	}
	if c.baseURL != "" {
		return []string{strings.TrimRight(c.baseURL, "/")}
	}
	return []string{"http://localhost:8100"}
}

func (c *DataHubClient) promoteCandidate(winner string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.baseURL = winner
	c.initBaseURL = winner
	if len(c.candidates) <= 1 {
		return
	}
	reordered := []string{winner}
	for _, u := range c.candidates {
		if u != winner {
			reordered = append(reordered, u)
		}
	}
	c.candidates = reordered
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
	Data     []map[string]interface{} `json:"data"`
	Findings []map[string]interface{} `json:"findings,omitempty"`
	Summary  map[string]interface{}   `json:"summary"`
}

type StratificationResponse struct {
	Data    []map[string]interface{} `json:"data"`
	Summary map[string]interface{}   `json:"summary"`
}

type ReconciliationResponse struct {
	Data    []map[string]interface{} `json:"data"`
	Summary map[string]interface{}   `json:"summary"`
}

type PolicyViolationsResponse struct {
	Data    []map[string]interface{} `json:"data"`
	Summary map[string]interface{}   `json:"summary"`
}

type DataQualityResponse struct {
	Data                []map[string]interface{} `json:"data"`
	Summary             map[string]interface{}   `json:"summary"`
	OverallQualityScore float64                  `json:"overall_quality_score"`
}

// ─── Client Methods ─────────────────────────────────────────────────────────

func buildCAATTEndpoint(basePath string, params map[string]string) string {
	v := url.Values{}
	for k, val := range params {
		if val != "" {
			v.Set(k, val)
		}
	}
	if encoded := v.Encode(); encoded != "" {
		return basePath + "?" + encoded
	}
	return basePath
}

func (c *DataHubClient) TriggerCAATTRun(sourceID string) (map[string]interface{}, error) {
	endpoint := buildCAATTEndpoint("/api/v1/caatt/run", map[string]string{
		"source_id": sourceID,
	})
	raw, err := c.postJSON(endpoint, map[string]interface{}{"source_id": sourceID})
	if err != nil {
		return nil, err
	}
	var res map[string]interface{}
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c *DataHubClient) GetFullPopulation(sourceID string, limit int) (*FullPopulationResponse, error) {
	endpoint := buildCAATTEndpoint("/api/v1/caatt/full-population", map[string]string{
		"limit":     fmt.Sprintf("%d", limit),
		"source_id": sourceID,
	})
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

func (c *DataHubClient) GetDuplicateGap(sourceID, resultType string, limit int) (*DuplicateGapResponse, error) {
	endpoint := buildCAATTEndpoint("/api/v1/caatt/duplicate-gap", map[string]string{
		"limit":       fmt.Sprintf("%d", limit),
		"result_type": resultType,
		"source_id":   sourceID,
	})
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

func (c *DataHubClient) GetBenfordAnalysis(sourceID string) (*BenfordResponse, error) {
	return c.GetBenfordAnalysisWithParams(sourceID, "", "")
}

// GetBenfordAnalysisWithParams fetches Benford digit analysis and real offending transactions filtered by source, digit (1-9), and findings_limit.
func (c *DataHubClient) GetBenfordAnalysisWithParams(sourceID, digit, findingsLimit string) (*BenfordResponse, error) {
	params := map[string]string{
		"source_id": sourceID,
	}
	if len(digit) == 1 && digit[0] >= '1' && digit[0] <= '9' {
		params["digit"] = digit
	}
	if findingsLimit != "" {
		isNum := true
		for i := 0; i < len(findingsLimit); i++ {
			if findingsLimit[i] < '0' || findingsLimit[i] > '9' {
				isNum = false
				break
			}
		}
		if isNum {
			params["findings_limit"] = findingsLimit
		}
	}
	endpoint := buildCAATTEndpoint("/api/v1/caatt/benford-analysis", params)
	raw, err := c.getJSON(endpoint)
	if err != nil {
		return nil, err
	}
	var res BenfordResponse
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *DataHubClient) GetStratification(sourceID, category string) (*StratificationResponse, error) {
	endpoint := buildCAATTEndpoint("/api/v1/caatt/stratification", map[string]string{
		"category":  category,
		"source_id": sourceID,
	})
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

func (c *DataHubClient) GetReconciliation(sourceID string) (*ReconciliationResponse, error) {
	endpoint := buildCAATTEndpoint("/api/v1/caatt/reconciliation", map[string]string{
		"source_id": sourceID,
	})
	raw, err := c.getJSON(endpoint)
	if err != nil {
		return nil, err
	}
	var res ReconciliationResponse
	if err := json.Unmarshal(raw, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *DataHubClient) GetPolicyViolations(sourceID, severity string, limit int) (*PolicyViolationsResponse, error) {
	endpoint := buildCAATTEndpoint("/api/v1/caatt/policy-violations", map[string]string{
		"limit":     fmt.Sprintf("%d", limit),
		"severity":  severity,
		"source_id": sourceID,
	})
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

func (c *DataHubClient) GetDataQualityMetrics(sourceID string) (*DataQualityResponse, error) {
	endpoint := buildCAATTEndpoint("/api/v1/caatt/data-quality", map[string]string{
		"source_id": sourceID,
	})
	raw, err := c.getJSON(endpoint)
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

func (c *DataHubClient) doRequest(method, endpoint string, body []byte, contentType string) (*http.Response, error) {
	urls := c.candidateURLs()
	var lastErr error
	for _, base := range urls {
		var reader io.Reader
		if body != nil {
			reader = bytes.NewReader(body)
		}
		req, err := http.NewRequest(method, base+endpoint, reader)
		if err != nil {
			return nil, err
		}
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		req.Header.Set("X-API-Key", c.apiKey)
		req.Header.Set("Accept", "application/json")

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		if base != urls[0] {
			c.promoteCandidate(base)
		}
		return resp, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("no Data Hub URL candidates available")
}

func (c *DataHubClient) getJSON(endpoint string) ([]byte, error) {
	resp, err := c.doRequest("GET", endpoint, nil, "")
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

	resp, err := c.doRequest("POST", endpoint, jsonBody, "application/json")
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
