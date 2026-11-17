package masterclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// ErrDepartmentNotFound means master-service answered and has no live
// department with that id. ErrDepartmentsUnavailable means it could not be
// asked (network, 5xx, 401, bad body); writes must refuse rather than guess.
var (
	ErrDepartmentNotFound     = errors.New("department not found in master-service")
	ErrDepartmentsUnavailable = errors.New("master-service departments unavailable")
)

// Department is the subset of a master-service department risk-service uses.
type Department struct {
	ID       uuid.UUID `json:"id"`
	Code     string    `json:"department_code"`
	Name     string    `json:"department_name"`
	IsActive bool      `json:"is_active"`
}

// DepartmentLookup validates a department on write. Unlike DepartmentName it
// tells "unknown" (ErrDepartmentNotFound) from "cannot ask"
// (ErrDepartmentsUnavailable).
type DepartmentLookup interface {
	LookupDepartment(ctx context.Context, id uuid.UUID) (Department, error)
}

// LookupDepartment reads GET /api/v1/departments/:id. A successful answer
// refreshes the name cache that DepartmentName serves.
func (c *DepartmentClient) LookupDepartment(ctx context.Context, id uuid.UUID) (Department, error) {
	if id == uuid.Nil {
		return Department{}, ErrDepartmentNotFound
	}
	status, body, err := c.getRaw(ctx, "/api/v1/departments/"+id.String())
	if err != nil {
		return Department{}, fmt.Errorf("%w: %v", ErrDepartmentsUnavailable, err)
	}
	switch {
	case status == http.StatusNotFound:
		return Department{}, ErrDepartmentNotFound
	case status != http.StatusOK:
		return Department{}, fmt.Errorf("%w: GET /api/v1/departments/%s: status %d", ErrDepartmentsUnavailable, id, status)
	}
	var env struct {
		Data Department `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return Department{}, fmt.Errorf("%w: decode department: %v", ErrDepartmentsUnavailable, err)
	}
	if env.Data.ID != id || strings.TrimSpace(env.Data.Name) == "" {
		return Department{}, ErrDepartmentNotFound
	}
	c.mu.Lock()
	c.names[id] = cachedName{name: env.Data.Name, at: c.now()}
	c.mu.Unlock()
	return env.Data, nil
}

// listPageSize is the page size ListDepartments asks for.
const listPageSize = 100

// maxListPages bounds ListDepartments (100 x 100 = 10,000 departments).
const maxListPages = 100

// ListDepartments returns every live department (active and inactive) from
// GET /api/v1/departments, following its pagination. Any failure is
// ErrDepartmentsUnavailable; a partial list is never returned.
func (c *DepartmentClient) ListDepartments(ctx context.Context) ([]Department, error) {
	var out []Department
	for page := 1; page <= maxListPages; page++ {
		q := url.Values{}
		q.Set("page", strconv.Itoa(page))
		q.Set("page_size", strconv.Itoa(listPageSize))
		status, body, err := c.getRaw(ctx, "/api/v1/departments?"+q.Encode())
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrDepartmentsUnavailable, err)
		}
		if status != http.StatusOK {
			return nil, fmt.Errorf("%w: GET /api/v1/departments: status %d", ErrDepartmentsUnavailable, status)
		}
		var env struct {
			Success bool `json:"success"`
			Data    struct {
				Departments []Department `json:"departments"`
				Pagination  struct {
					Total      int64 `json:"total"`
					TotalPages int64 `json:"total_pages"`
				} `json:"pagination"`
			} `json:"data"`
		}
		if err := json.Unmarshal(body, &env); err != nil {
			return nil, fmt.Errorf("%w: decode departments: %v", ErrDepartmentsUnavailable, err)
		}
		if !env.Success {
			return nil, fmt.Errorf("%w: master-service reported success=false", ErrDepartmentsUnavailable)
		}
		for _, d := range env.Data.Departments {
			if d.ID != uuid.Nil {
				out = append(out, d)
			}
		}
		if int64(page) >= env.Data.Pagination.TotalPages || len(env.Data.Departments) == 0 {
			return out, nil
		}
	}
	return nil, fmt.Errorf("%w: more than %d pages of departments", ErrDepartmentsUnavailable, maxListPages)
}

// getRaw performs a GET and returns the status and (bounded) body.
func (c *DepartmentClient) getRaw(ctx context.Context, path string) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Accept", "application/json")
	if auth, ok := ctx.Value(authKey{}).(string); ok && auth != "" {
		req.Header.Set("Authorization", auth)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, body, nil
}

// BaseURL reports the master-service URL this client reads from.
func (c *DepartmentClient) BaseURL() string { return c.baseURL }
