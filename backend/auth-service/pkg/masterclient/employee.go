// Package masterclient reads employee master data owned by master-service.
//
// auth-service has no access to master-service's database, so the employee
// behind a login account (users.employee_id = employees.employee_code) is read
// over HTTP from GET /api/v1/employees/by-code/:code. The caller's bearer token is
// forwarded: master-service validates it with the same jwt.secret.
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
	"time"
)

// ErrEmployeeNotFound means master-service has no employee with that code.
var ErrEmployeeNotFound = errors.New("master-service employee not found")

// maxBody guards against a misrouted URL streaming an unexpectedly large body.
const maxBody = 1 << 20

// Employee is the subset of master-service's employee record auth-service needs.
type Employee struct {
	ID           string `json:"id"`
	EmployeeCode string `json:"employee_code"`
	DepartmentID string `json:"department_id"`
}

// EmployeeClient looks up employees by code on master-service.
type EmployeeClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewEmployeeClient builds a client for the master-service at baseURL.
func NewEmployeeClient(baseURL string, timeout time.Duration) *EmployeeClient {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	return &EmployeeClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: timeout},
	}
}

// BaseURL returns the master-service base URL this client calls.
func (c *EmployeeClient) BaseURL() string {
	return c.baseURL
}

// FindByCode returns the employee with the given employee_code. authHeader is
// the caller's full Authorization header value ("Bearer <jwt>"); it may be empty.
// ErrEmployeeNotFound is returned on 404. Any other failure is returned as an
// error and callers should treat the link as unavailable, not fail the request.
func (c *EmployeeClient) FindByCode(ctx context.Context, code, authHeader string) (*Employee, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, ErrEmployeeNotFound
	}

	reqURL := fmt.Sprintf("%s/api/v1/employees/by-code/%s", c.baseURL, url.PathEscape(code))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrEmployeeNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("master-service employee lookup returned status %d", resp.StatusCode)
	}

	var envelope struct {
		Data Employee `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxBody)).Decode(&envelope); err != nil {
		return nil, err
	}
	if envelope.Data.ID == "" {
		return nil, ErrEmployeeNotFound
	}
	return &envelope.Data, nil
}
