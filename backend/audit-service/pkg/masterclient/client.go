// Package masterclient reads companies and departments from this stack's
// master-service on behalf of the caller.
//
// master-service requires a JWT on every /api/v1 route and audit-service has
// no service account, so each request forwards the caller's own Authorization
// header unchanged (same approach as pkg/authclient). Without a token no
// request is made.
package masterclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"audit-service/pkg/logger"
	"audit-service/pkg/serviceurl"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Company struct {
	ID          string `json:"id"`
	CompanyCode string `json:"code"`
	CompanyName string `json:"name"`
	LegalName   string `json:"legal_name"`
	CompanyType string `json:"company_type"`
}

// displayName prefers the registered legal name
func (c *Company) displayName() string {
	if c == nil {
		return ""
	}
	if n := strings.TrimSpace(c.LegalName); n != "" {
		return n
	}
	return strings.TrimSpace(c.CompanyName)
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

var (
	// ErrNoAuthorization means there was no caller token to forward, so
	// master-service was not called
	ErrNoAuthorization = errors.New("no authorization to forward")
	// ErrNotConfigured means no master-service URL could be worked out for
	// this stack (see serviceurl.Resolve)
	ErrNotConfigured = errors.New("master-service URL not configured for this stack")
	// ErrNotFound means master-service answered but nothing matched
	ErrNotFound = errors.New("not found in master-service")
)

const (
	// cacheTTL bounds how long a renamed company keeps its old name
	cacheTTL = 5 * time.Minute
	// departmentPageSize: GET /departments pages at 10 by default, which
	// would hide most departments from the working-unit match
	departmentPageSize = "500"
	maxBodyBytes       = 4 << 20
)

type cachedCompany struct {
	comp *Company
	at   time.Time
}

type cachedDept struct {
	dept *Department
	at   time.Time
}

// MasterClient talks to one master-service base URL. The caches are per
// process: each tenant runs its own audit-service container with its own
// MasterClient, so entries never cross tenants. Only answers that came back
// 200 with a usable body are cached; errors (401, 5xx, unreachable) never are.
// Any authenticated user may read companies and departments in master-service,
// so a cached entry does not widen what another caller of the same stack sees.
type MasterClient struct {
	baseURL    string
	httpClient *http.Client
	now        func() time.Time
	cacheMu    sync.RWMutex
	compCache  map[string]cachedCompany
	deptCache  map[string]cachedDept
}

var (
	defaultClient *MasterClient
	once          sync.Once
)

// New returns a client for baseURL ("" makes every lookup fail with
// ErrNotConfigured). hc nil uses a 2s timeout client.
func New(baseURL string, hc *http.Client) *MasterClient {
	if hc == nil {
		hc = &http.Client{Timeout: 2 * time.Second}
	}
	return &MasterClient{
		baseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		httpClient: hc,
		now:        time.Now,
		compCache:  make(map[string]cachedCompany),
		deptCache:  make(map[string]cachedDept),
	}
}

// GetClient returns the process-wide client for this stack's master-service
// (MASTER_SERVICE_URL, else derived from KAFKA_SERVICE_NAME, else
// http://master-service:8002; see serviceurl.Resolve).
func GetClient() *MasterClient {
	once.Do(func() {
		defaultClient = New(serviceurl.Master(), nil)
	})
	return defaultClient
}

// BaseURL is the master-service base URL this client calls
func (c *MasterClient) BaseURL() string { return c.baseURL }

// getData performs GET <base><path>?<query> with the caller's token and
// returns the envelope's data. Any non-200 answer is an error.
func (c *MasterClient) getData(ctx context.Context, authorization, path string, query url.Values) (json.RawMessage, error) {
	if strings.TrimSpace(authorization) == "" {
		return nil, ErrNoAuthorization
	}
	if c.baseURL == "" {
		return nil, ErrNotConfigured
	}
	reqURL := c.baseURL + path
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authorization)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxBodyBytes))
		return nil, fmt.Errorf("master-service answered %d for %s", resp.StatusCode, path)
	}

	var apiResp APIResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxBodyBytes)).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode master-service response: %w", err)
	}
	return apiResp.Data, nil
}

// GetCompanyByID fetches GET /api/v1/companies/:id with the caller's token
func (c *MasterClient) GetCompanyByID(ctx context.Context, authorization, id string) (*Company, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, fmt.Errorf("empty company id")
	}

	c.cacheMu.RLock()
	entry, ok := c.compCache[id]
	c.cacheMu.RUnlock()
	if ok && c.now().Sub(entry.at) < cacheTTL {
		return entry.comp, nil
	}

	data, err := c.getData(ctx, authorization, "/api/v1/companies/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, err
	}
	var comp Company
	if err := json.Unmarshal(data, &comp); err != nil {
		return nil, fmt.Errorf("decode company: %w", err)
	}
	if comp.displayName() == "" {
		return nil, ErrNotFound
	}

	c.cacheMu.Lock()
	c.compCache[id] = cachedCompany{comp: &comp, at: c.now()}
	c.cacheMu.Unlock()
	return &comp, nil
}

// companyOf returns d's company: the embedded one when it carries a name
// (master-service's list does not preload it, so it usually arrives as a zero
// value), else a lookup by company_id.
func (c *MasterClient) companyOf(ctx context.Context, authorization string, d *Department) (*Company, error) {
	if d.Company != nil && d.Company.displayName() != "" {
		return d.Company, nil
	}
	if id, err := uuid.Parse(strings.TrimSpace(d.CompanyID)); err == nil && id != uuid.Nil {
		return c.GetCompanyByID(ctx, authorization, id.String())
	}
	return nil, ErrNotFound
}

// GetCompanyByWorkingUnit finds the department whose name or code matches
// unitName (exact first, then substring either way) and returns its company.
func (c *MasterClient) GetCompanyByWorkingUnit(ctx context.Context, authorization, unitName string) (*Company, error) {
	unitLower := strings.ToLower(strings.TrimSpace(unitName))
	if unitLower == "" {
		return nil, fmt.Errorf("empty working unit")
	}

	c.cacheMu.RLock()
	entry, ok := c.deptCache[unitLower]
	c.cacheMu.RUnlock()
	if ok && c.now().Sub(entry.at) < cacheTTL {
		return c.companyOf(ctx, authorization, entry.dept)
	}

	q := url.Values{}
	q.Set("page_size", departmentPageSize)
	data, err := c.getData(ctx, authorization, "/api/v1/departments", q)
	if err != nil {
		return nil, err
	}

	var depts []Department
	if err := json.Unmarshal(data, &depts); err != nil {
		var wrapper struct {
			Departments []Department `json:"departments"`
		}
		if err2 := json.Unmarshal(data, &wrapper); err2 != nil {
			return nil, fmt.Errorf("decode departments: %w", err2)
		}
		depts = wrapper.Departments
	}

	now := c.now()
	var exact, partial *Department
	c.cacheMu.Lock()
	for i := range depts {
		d := &depts[i]
		name := strings.ToLower(strings.TrimSpace(d.DepartmentName))
		code := strings.ToLower(strings.TrimSpace(d.DepartmentCode))
		if name != "" {
			c.deptCache[name] = cachedDept{dept: d, at: now}
		}
		if code != "" {
			c.deptCache[code] = cachedDept{dept: d, at: now}
		}
		if exact == nil && (name == unitLower || code == unitLower) {
			exact = d
		}
		// An empty name would "contain" every unit, so it never matches
		if partial == nil && name != "" && (strings.Contains(name, unitLower) || strings.Contains(unitLower, name)) {
			partial = d
		}
	}
	c.cacheMu.Unlock()

	match := exact
	if match == nil {
		match = partial
	}
	if match == nil {
		return nil, ErrNotFound
	}
	return c.companyOf(ctx, authorization, match)
}

// maxDepartmentPages stops ListDepartments from following a pagination that
// never ends (500 x 200 = 100k departments is far beyond any stack).
const maxDepartmentPages = 200

// ListDepartments returns every live department of this stack's
// master-service (GET /api/v1/departments, all pages) using the caller's
// Authorization header. Nothing is cached: it is for bulk reconciliation
// (services/annual_plan_sync), not for per-request lookups.
func (c *MasterClient) ListDepartments(ctx context.Context, authorization string) ([]Department, error) {
	var all []Department
	for page := 1; page <= maxDepartmentPages; page++ {
		q := url.Values{}
		q.Set("page", fmt.Sprint(page))
		q.Set("page_size", departmentPageSize)
		data, err := c.getData(ctx, authorization, "/api/v1/departments", q)
		if err != nil {
			return nil, err
		}

		var wrapper struct {
			Departments []Department `json:"departments"`
			Pagination  struct {
				TotalPages int `json:"total_pages"`
			} `json:"pagination"`
		}
		if err := json.Unmarshal(data, &wrapper); err != nil {
			// An unpaginated master-service answers with the bare array.
			var depts []Department
			if err2 := json.Unmarshal(data, &depts); err2 != nil {
				return nil, fmt.Errorf("decode departments: %w", err)
			}
			return append(all, depts...), nil
		}
		all = append(all, wrapper.Departments...)
		if len(wrapper.Departments) == 0 || page >= wrapper.Pagination.TotalPages {
			return all, nil
		}
	}
	return nil, fmt.Errorf("departments: more than %d pages", maxDepartmentPages)
}

// ResolveCompanyName resolves the company name for a report: the report's own
// name, then the assignment letter's, then master-service (by report company
// id, letter company id, then working unit / department) using the caller's
// Authorization header. It returns "" when nothing resolves — never a
// hardcoded company — and logs one warning saying why.
func ResolveCompanyName(
	ctx context.Context,
	authorization string,
	repCompanyID *uuid.UUID,
	repCompanyName string,
	stCompanyID *uuid.UUID,
	stCompanyName string,
	workingUnit string,
	department string,
) string {
	return resolveCompanyName(ctx, GetClient(), authorization, repCompanyID, repCompanyName, stCompanyID, stCompanyName, workingUnit, department)
}

func resolveCompanyName(
	ctx context.Context,
	client *MasterClient,
	authorization string,
	repCompanyID *uuid.UUID,
	repCompanyName string,
	stCompanyID *uuid.UUID,
	stCompanyName string,
	workingUnit string,
	department string,
) string {
	// 1. Direct explicit name from report
	if n := strings.TrimSpace(repCompanyName); n != "" {
		return n
	}
	// 2. Direct explicit name from assignment letter
	if n := strings.TrimSpace(stCompanyName); n != "" {
		return n
	}

	targetUnit := strings.TrimSpace(workingUnit)
	if targetUnit == "" {
		targetUnit = strings.TrimSpace(department)
	}

	var lastErr error
	if strings.TrimSpace(authorization) == "" {
		lastErr = ErrNoAuthorization
	} else {
		// 3./4. Lookup by report.CompanyID, then st.CompanyID
		for _, id := range []*uuid.UUID{repCompanyID, stCompanyID} {
			if id == nil || *id == uuid.Nil {
				continue
			}
			comp, err := client.GetCompanyByID(ctx, authorization, id.String())
			if err == nil {
				if n := comp.displayName(); n != "" {
					return n
				}
				continue
			}
			lastErr = err
		}
		// 5. Lookup by working unit / department
		if targetUnit != "" {
			comp, err := client.GetCompanyByWorkingUnit(ctx, authorization, targetUnit)
			if err == nil {
				if n := comp.displayName(); n != "" {
					return n
				}
			} else {
				lastErr = err
			}
		}
		if lastErr == nil {
			lastErr = ErrNotFound
		}
	}

	// 6. Nothing resolved: no hardcoded fallback
	if logger.Log != nil {
		logger.Warn("company name not resolved from report, assignment letter or master-service",
			zap.String("master_service_url", client.BaseURL()),
			zap.Bool("has_authorization", strings.TrimSpace(authorization) != ""),
			zap.Bool("has_report_company_id", repCompanyID != nil && *repCompanyID != uuid.Nil),
			zap.Bool("has_letter_company_id", stCompanyID != nil && *stCompanyID != uuid.Nil),
			zap.Bool("has_working_unit", targetUnit != ""),
			zap.String("reason", lastErr.Error()),
		)
	}
	return ""
}
