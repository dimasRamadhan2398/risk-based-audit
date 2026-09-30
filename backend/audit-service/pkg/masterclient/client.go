package masterclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID          string `json:"id"`
	CompanyCode string `json:"code"`
	CompanyName string `json:"name"`
	LegalName   string `json:"legal_name"`
	CompanyType string `json:"company_type"`
}

type Department struct {
	ID             string   `json:"id"`
	DepartmentCode string   `json:"department_code"`
	DepartmentName string   `json:"department_name"`
	CompanyID      string   `json:"company_id"`
	Company        *Company `json:"company,omitempty"`
}

type APIResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type MasterClient struct {
	baseURLs   []string
	httpClient *http.Client
	cacheMu    sync.RWMutex
	compCache  map[string]*Company
	deptCache  map[string]*Department
}

var (
	defaultClient *MasterClient
	once          sync.Once
)

func GetClient() *MasterClient {
	once.Do(func() {
		var urls []string
		if envURL := os.Getenv("MASTER_SERVICE_URL"); envURL != "" {
			urls = append(urls, strings.TrimRight(envURL, "/"))
		}
		// Candidate URLs: internal docker network, then host network fallback
		urls = append(urls, "http://master-service:8002", "http://localhost:8003", "http://127.0.0.1:8003")

		defaultClient = &MasterClient{
			baseURLs: urls,
			httpClient: &http.Client{
				Timeout: 2 * time.Second,
			},
			compCache: make(map[string]*Company),
			deptCache: make(map[string]*Department),
		}
	})
	return defaultClient
}

func (c *MasterClient) GetCompanyByID(ctx context.Context, id string) (*Company, error) {
	if id == "" {
		return nil, fmt.Errorf("empty company id")
	}

	c.cacheMu.RLock()
	if comp, ok := c.compCache[id]; ok {
		c.cacheMu.RUnlock()
		return comp, nil
	}
	c.cacheMu.RUnlock()

	var lastErr error
	for _, base := range c.baseURLs {
		reqURL := fmt.Sprintf("%s/api/v1/companies/%s", base, id)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			lastErr = err
			continue
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
			continue
		}

		var apiResp APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
			lastErr = err
			continue
		}

		var comp Company
		if err := json.Unmarshal(apiResp.Data, &comp); err != nil {
			lastErr = err
			continue
		}

		c.cacheMu.Lock()
		c.compCache[id] = &comp
		c.cacheMu.Unlock()
		return &comp, nil
	}

	return nil, lastErr
}

func (c *MasterClient) GetCompanyByWorkingUnit(ctx context.Context, unitName string) (*Company, error) {
	if unitName == "" {
		return nil, fmt.Errorf("empty working unit")
	}

	unitLower := strings.ToLower(strings.TrimSpace(unitName))

	c.cacheMu.RLock()
	if dept, ok := c.deptCache[unitLower]; ok && dept.Company != nil {
		c.cacheMu.RUnlock()
		return dept.Company, nil
	}
	c.cacheMu.RUnlock()

	var lastErr error
	for _, base := range c.baseURLs {
		reqURL := fmt.Sprintf("%s/api/v1/departments", base)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			lastErr = err
			continue
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
			continue
		}

		var apiResp APIResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
			lastErr = err
			continue
		}

		var depts []Department
		if err := json.Unmarshal(apiResp.Data, &depts); err != nil {
			var wrapper struct {
				Departments []Department `json:"departments"`
			}
			if err2 := json.Unmarshal(apiResp.Data, &wrapper); err2 == nil {
				depts = wrapper.Departments
			}
		}

		for _, d := range depts {
			dLower := strings.ToLower(d.DepartmentName)
			cLower := strings.ToLower(d.DepartmentCode)
			c.cacheMu.Lock()
			c.deptCache[dLower] = &d
			c.deptCache[cLower] = &d
			c.cacheMu.Unlock()

			if strings.Contains(dLower, unitLower) || strings.Contains(unitLower, dLower) {
				if d.Company != nil {
					return d.Company, nil
				}
				if d.CompanyID != "" {
					return c.GetCompanyByID(ctx, d.CompanyID)
				}
			}
		}
	}

	return nil, lastErr
}

// ResolveCompanyName resolves the dynamic company name from report, assignment letter, or master-service.
func ResolveCompanyName(
	ctx context.Context,
	repCompanyID *uuid.UUID,
	repCompanyName string,
	stCompanyID *uuid.UUID,
	stCompanyName string,
	workingUnit string,
	department string,
) string {
	// 1. Direct explicit name from report
	if strings.TrimSpace(repCompanyName) != "" {
		return strings.TrimSpace(repCompanyName)
	}

	// 2. Direct explicit name from assignment letter
	if strings.TrimSpace(stCompanyName) != "" {
		return strings.TrimSpace(stCompanyName)
	}

	client := GetClient()

	// 3. Lookup by report.CompanyID
	if repCompanyID != nil && *repCompanyID != uuid.Nil {
		if comp, err := client.GetCompanyByID(ctx, repCompanyID.String()); err == nil && comp != nil {
			if comp.LegalName != "" {
				return comp.LegalName
			}
			if comp.CompanyName != "" {
				return comp.CompanyName
			}
		}
	}

	// 4. Lookup by st.CompanyID
	if stCompanyID != nil && *stCompanyID != uuid.Nil {
		if comp, err := client.GetCompanyByID(ctx, stCompanyID.String()); err == nil && comp != nil {
			if comp.LegalName != "" {
				return comp.LegalName
			}
			if comp.CompanyName != "" {
				return comp.CompanyName
			}
		}
	}

	// 5. Lookup by working unit
	targetUnit := strings.TrimSpace(workingUnit)
	if targetUnit == "" {
		targetUnit = strings.TrimSpace(department)
	}
	if targetUnit != "" {
		if comp, err := client.GetCompanyByWorkingUnit(ctx, targetUnit); err == nil && comp != nil {
			if comp.LegalName != "" {
				return comp.LegalName
			}
			if comp.CompanyName != "" {
				return comp.CompanyName
			}
		}
	}

	// 6. Default fallback
	return "PT AIFL Indonesia"
}
