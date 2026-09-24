package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// ─── Sync Models (matches Data Hub API sync.py) ──────────────────────────────

type CompanySync struct {
	ID          string `json:"id,omitempty"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	LegalName   string `json:"legal_name,omitempty"`
	CompanyType string `json:"company_type,omitempty"`
	IsActive    bool   `json:"is_active"`
}

type DepartmentSync struct {
	ID             string `json:"id,omitempty"`
	DepartmentCode string `json:"department_code"`
	DepartmentName string `json:"department_name"`
	BusinessUnitID string `json:"business_unit_id,omitempty"`
	IsActive       bool   `json:"is_active"`
}

type EmployeeSync struct {
	ID               string `json:"id,omitempty"`
	EmployeeIDNumber string `json:"employee_id_number"`
	FullName         string `json:"full_name"`
	Email            string `json:"email,omitempty"`
	DepartmentID     string `json:"department_id,omitempty"`
	IsActive         bool   `json:"is_active"`
}

type RiskCategorySync struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type RiskRegisterSync struct {
	ID             string `json:"id,omitempty"`
	Code           string `json:"code"`
	Name           string `json:"name"`
	DepartmentID   string `json:"department_id,omitempty"`
	RiskCategoryID string `json:"risk_category_id,omitempty"`
	InherentScore  int    `json:"inherent_score"`
	ResidualScore  int    `json:"residual_score"`
	IsActive       bool   `json:"is_active"`
}

type AuditFindingSync struct {
	ID           string `json:"id,omitempty"`
	FindingCode  string `json:"finding_code"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	Severity     string `json:"severity,omitempty"`
	Status       string `json:"status,omitempty"`
	DepartmentID string `json:"department_id,omitempty"`
}

type MasterDataSyncPayload struct {
	Companies      []CompanySync      `json:"companies,omitempty"`
	Departments    []DepartmentSync    `json:"departments,omitempty"`
	Employees      []EmployeeSync      `json:"employees,omitempty"`
	RiskCategories []RiskCategorySync  `json:"risk_categories,omitempty"`
	RiskRegisters  []RiskRegisterSync  `json:"risk_registers,omitempty"`
	AuditFindings  []AuditFindingSync  `json:"audit_findings,omitempty"`
}

// ─── DataHubSyncClient ───────────────────────────────────────────────────────

type DataHubSyncClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewDataHubSyncClient() *DataHubSyncClient {
	baseURL := os.Getenv("DATA_HUB_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8100"
	}
	apiKey := os.Getenv("DATA_HUB_API_KEY")
	if apiKey == "" {
		apiKey = "dev-api-key"
	}
	return &DataHubSyncClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 8 * time.Second,
		},
	}
}

// SyncMasterData pushes master data changes to Data Hub Bronze & Silver zones
func (c *DataHubSyncClient) SyncMasterData(payload MasterDataSyncPayload) error {
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal sync payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/sync/master-data", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create sync request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("Data Hub sync connection error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Data Hub sync responded with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// SyncMasterDataAsync runs sync asynchronously in background without blocking CRUD
func (c *DataHubSyncClient) SyncMasterDataAsync(payload MasterDataSyncPayload) {
	go func() {
		if err := c.SyncMasterData(payload); err != nil {
			log.Printf("[DataHubSync] Background sync notice (non-fatal): %v", err)
		} else {
			log.Println("[DataHubSync] Successfully synced master data batch to Data Hub")
		}
	}()
}
