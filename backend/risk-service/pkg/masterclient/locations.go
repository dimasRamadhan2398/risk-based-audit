// Package masterclient reads master data owned by master-service.
//
// risk-service has no access to master-service's database (each service owns
// its own), so Location master rows are read over HTTP from GET
// /api/v1/locations on the master-service of the same stack.
package masterclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ErrUnavailable means the Location master could not be read and no usable
// cached copy exists. Callers must not invent branch names in that case.
var ErrUnavailable = errors.New("master-service locations unavailable")

// Location is the subset of master-service's Location the risk service needs.
type Location struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	IsActive bool      `json:"is_active"`
}

// LocationSource lists the locations registered in the Location master.
type LocationSource interface {
	ListLocations(ctx context.Context) ([]Location, error)
}

type authKey struct{}

// WithAuthorization carries the caller's Authorization header so it can be
// forwarded to master-service. master-service does not check JWTs today; if it
// starts to, lookups keep working instead of silently degrading to null.
func WithAuthorization(ctx context.Context, header string) context.Context {
	if header == "" {
		return ctx
	}
	return context.WithValue(ctx, authKey{}, header)
}

const (
	defaultTimeout = 2 * time.Second
	defaultTTL     = 60 * time.Second
	// staleFor is how long the last good list is still served while
	// master-service is unreachable. Names are master data, just slightly old.
	staleFor = 10 * time.Minute
	// retryAfter stops every request paying the HTTP timeout during an outage.
	retryAfter = 5 * time.Second
	// maxBody guards against a misrouted URL streaming something huge.
	maxBody = 4 << 20
)

// Client is a LocationSource backed by master-service with an in-memory cache.
// Locations change rarely and the list is small, so the whole list is cached.
type Client struct {
	baseURL    string
	httpClient *http.Client
	ttl        time.Duration
	now        func() time.Time

	mu          sync.Mutex
	cached      []Location
	fetchedAt   time.Time
	lastFailure time.Time
}

// NewClient builds a client for the master-service at baseURL. Zero timeout or
// ttl fall back to 2s and 60s.
func NewClient(baseURL string, timeout, ttl time.Duration) *Client {
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	if ttl <= 0 {
		ttl = defaultTTL
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: timeout},
		ttl:        ttl,
		now:        time.Now,
	}
}

// BaseURL reports the master-service URL this client reads from.
func (c *Client) BaseURL() string { return c.baseURL }

// ListLocations returns all registered locations (active and inactive).
//
// Fresh cache is served without a request. On failure the last good list is
// served for up to staleFor; past that ErrUnavailable is returned.
func (c *Client) ListLocations(ctx context.Context) ([]Location, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	if c.cached != nil && now.Sub(c.fetchedAt) < c.ttl {
		return c.cached, nil
	}
	if !c.lastFailure.IsZero() && now.Sub(c.lastFailure) < retryAfter {
		return c.staleOrErr(now, errors.New("recent fetch failed"))
	}

	locs, err := c.fetch(ctx)
	if err != nil {
		c.lastFailure = now
		return c.staleOrErr(now, err)
	}

	c.cached = locs
	c.fetchedAt = now
	c.lastFailure = time.Time{}
	return locs, nil
}

func (c *Client) staleOrErr(now time.Time, cause error) ([]Location, error) {
	if c.cached != nil && now.Sub(c.fetchedAt) < staleFor {
		return c.cached, nil
	}
	return nil, fmt.Errorf("%w: %v", ErrUnavailable, cause)
}

func (c *Client) fetch(ctx context.Context) ([]Location, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/locations", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if auth, ok := ctx.Value(authKey{}).(string); ok && auth != "" {
		req.Header.Set("Authorization", auth)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s/api/v1/locations: status %d", c.baseURL, resp.StatusCode)
	}

	var body struct {
		Success bool       `json:"success"`
		Data    []Location `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxBody)).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode locations: %w", err)
	}
	if !body.Success {
		return nil, errors.New("master-service reported success=false for locations")
	}

	locs := make([]Location, 0, len(body.Data))
	for _, l := range body.Data {
		if l.ID == uuid.Nil || strings.TrimSpace(l.Name) == "" {
			continue
		}
		locs = append(locs, l)
	}
	return locs, nil
}
