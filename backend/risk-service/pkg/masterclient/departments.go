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

	"github.com/google/uuid"
)

// ErrDepartmentUnresolved means the caller's department could not be
// determined (no employee link in the token, employee unknown, or
// master-service unreachable). Callers must fail closed.
var ErrDepartmentUnresolved = errors.New("caller department unresolved")

// DepartmentSource resolves departments through master-service.
type DepartmentSource interface {
	// CallerDepartment returns the department of the employee behind the
	// token's employee_code (or employee_id when there is no code).
	CallerDepartment(ctx context.Context, employeeCode, employeeID string) (uuid.UUID, error)
	// DepartmentName returns a department's name; ok is false when unknown or
	// master-service is unavailable. Best-effort, never an error.
	DepartmentName(ctx context.Context, id uuid.UUID) (name string, ok bool)
}

// DepartmentClient is a DepartmentSource with a short in-memory cache. The
// cache is keyed by the employee identity taken from a validated JWT, and a
// department move takes effect after at most the TTL.
type DepartmentClient struct {
	baseURL    string
	httpClient *http.Client
	ttl        time.Duration
	now        func() time.Time

	mu          sync.Mutex
	callers     map[string]cachedID
	names       map[uuid.UUID]cachedName
	nameFailure time.Time // last failed name fetch; names are skipped briefly after
}

type cachedID struct {
	id uuid.UUID
	at time.Time
}

type cachedName struct {
	name string
	at   time.Time
}

// NewDepartmentClient builds a client for the master-service at baseURL. Zero
// timeout or ttl fall back to 2s and 60s.
func NewDepartmentClient(baseURL string, timeout, ttl time.Duration) *DepartmentClient {
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	if ttl <= 0 {
		ttl = defaultTTL
	}
	return &DepartmentClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: timeout},
		ttl:        ttl,
		now:        time.Now,
		callers:    map[string]cachedID{},
		names:      map[uuid.UUID]cachedName{},
	}
}

func (c *DepartmentClient) CallerDepartment(ctx context.Context, employeeCode, employeeID string) (uuid.UUID, error) {
	employeeCode, employeeID = strings.TrimSpace(employeeCode), strings.TrimSpace(employeeID)
	var key, path string
	switch {
	case employeeCode != "":
		key, path = "code:"+strings.ToLower(employeeCode), "/api/v1/employees/by-code/"+url.PathEscape(employeeCode)
	case employeeID != "":
		if _, err := uuid.Parse(employeeID); err != nil {
			return uuid.Nil, ErrDepartmentUnresolved
		}
		key, path = "id:"+strings.ToLower(employeeID), "/api/v1/employees/"+url.PathEscape(employeeID)
	default:
		return uuid.Nil, ErrDepartmentUnresolved
	}

	c.mu.Lock()
	if e, ok := c.callers[key]; ok && c.now().Sub(e.at) < c.ttl {
		c.mu.Unlock()
		return e.id, nil
	}
	c.mu.Unlock()

	var body struct {
		Data struct {
			DepartmentID string `json:"department_id"`
		} `json:"data"`
	}
	if err := c.get(ctx, path, &body); err != nil {
		return uuid.Nil, fmt.Errorf("%w: %v", ErrDepartmentUnresolved, err)
	}
	id, err := uuid.Parse(body.Data.DepartmentID)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, fmt.Errorf("%w: employee has no department", ErrDepartmentUnresolved)
	}
	c.mu.Lock()
	c.callers[key] = cachedID{id: id, at: c.now()}
	c.mu.Unlock()
	return id, nil
}

func (c *DepartmentClient) DepartmentName(ctx context.Context, id uuid.UUID) (string, bool) {
	c.mu.Lock()
	if e, ok := c.names[id]; ok && c.now().Sub(e.at) < c.ttl {
		c.mu.Unlock()
		return e.name, true
	}
	if !c.nameFailure.IsZero() && c.now().Sub(c.nameFailure) < retryAfter {
		c.mu.Unlock()
		return "", false
	}
	c.mu.Unlock()

	var body struct {
		Data struct {
			Name string `json:"department_name"`
		} `json:"data"`
	}
	if err := c.get(ctx, "/api/v1/departments/"+id.String(), &body); err != nil || strings.TrimSpace(body.Data.Name) == "" {
		c.mu.Lock()
		c.nameFailure = c.now()
		c.mu.Unlock()
		return "", false
	}
	c.mu.Lock()
	c.names[id] = cachedName{name: body.Data.Name, at: c.now()}
	c.mu.Unlock()
	return body.Data.Name, true
}

func (c *DepartmentClient) get(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if auth, ok := ctx.Value(authKey{}).(string); ok && auth != "" {
		req.Header.Set("Authorization", auth)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: status %d", path, resp.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, maxBody)).Decode(out)
}
