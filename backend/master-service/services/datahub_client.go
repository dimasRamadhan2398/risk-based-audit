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

// DataHubClient communicates with Data Hub API on Data Server via VPN / Local
type DataHubClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

func NewDataHubClient() *DataHubClient {
	baseURL := os.Getenv("DATA_HUB_BASE_URL")
	if baseURL == "" {
		baseURL = os.Getenv("DATA_HUB_URL")
	}
	if baseURL == "" {
		baseURL = "http://localhost:8100"
	}

	apiKey := os.Getenv("DATA_HUB_API_KEY")
	if apiKey == "" {
		apiKey = "dev-api-key"
	}

	return &DataHubClient{
		BaseURL:    baseURL,
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// RegisterSource registers or updates connection config in Data Hub
func (c *DataHubClient) RegisterSource(payload map[string]interface{}) error {
	return c.post("/api/v1/sources/register", payload)
}

// TriggerIngest triggers ingestion for specified source IDs
func (c *DataHubClient) TriggerIngest(sourceIDs []string, mode string) (map[string]interface{}, error) {
	body := map[string]interface{}{
		"source_ids":    sourceIDs,
		"mode":          mode,
		"run_transform": true,
	}
	return c.postWithResponse("/api/v1/sources/ingest", body)
}

// TriggerBatchIngest triggers ingestion for ALL registered sources
func (c *DataHubClient) TriggerBatchIngest() (map[string]interface{}, error) {
	body := map[string]interface{}{
		"source_ids":    []string{},
		"mode":          "all",
		"run_transform": true,
	}
	return c.postWithResponse("/api/v1/sources/ingest", body)
}

// DeregisterSource removes a source from Data Hub
func (c *DataHubClient) DeregisterSource(sourceID string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/sources/%s", c.BaseURL, sourceID), nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", c.APIKey)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// GetIngestStatus checks progress of a multi-source ingest job
func (c *DataHubClient) GetIngestStatus(jobID string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/sources/ingest/%s/status", c.BaseURL, jobID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", c.APIKey)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Data Hub returned %d: %s", resp.StatusCode, string(b))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetSourceSchema connects to source database catalog via Data Hub and returns real tables/columns
func (c *DataHubClient) GetSourceSchema(sourceID string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/sources/%s/schema", c.BaseURL, sourceID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", c.APIKey)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Data Hub returned %d: %s", resp.StatusCode, string(b))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// PreviewSourceTable fetches sample rows from source DB via Data Hub
func (c *DataHubClient) PreviewSourceTable(sourceID string, tableName string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/sources/%s/preview/%s", c.BaseURL, sourceID, tableName), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", c.APIKey)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Data Hub returned %d: %s", resp.StatusCode, string(b))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *DataHubClient) post(path string, body interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", c.BaseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.APIKey)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("DataHub API error %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

func (c *DataHubClient) postWithResponse(path string, body interface{}) (map[string]interface{}, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", c.BaseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.APIKey)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("DataHub API error %d: %s", resp.StatusCode, string(b))
	}
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}
