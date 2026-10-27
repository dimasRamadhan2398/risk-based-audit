// Package riskclient reads the corporate Audit Universe and its yearly scores
// from this stack's risk-service on behalf of the caller.
//
// risk-service requires a JWT on every /api/v1 route and audit-service has no
// service account, so each request forwards the caller's own Authorization
// header unchanged (same approach as pkg/masterclient and pkg/authclient).
// Without a token no request is made.
package riskclient

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
	"time"

	"audit-service/pkg/serviceurl"
)

var (
	// ErrNoAuthorization means there was no caller token to forward, so
	// risk-service was not called
	ErrNoAuthorization = errors.New("no authorization to forward")
	// ErrNotConfigured means no risk-service URL could be worked out for this
	// stack (see serviceurl.Resolve)
	ErrNotConfigured = errors.New("risk-service URL not configured for this stack")
)

// maxBodyBytes bounds one answer. The yearly list preloads every score row of
// every entity, so it is larger than master-service's answers.
const maxBodyBytes = 32 << 20

// Entity is a corporate Audit Universe node
// (risk-service models.CorporateAuditUniverse).
type Entity struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	ParentID *string  `json:"parent_id,omitempty"`
	Children []Entity `json:"children,omitempty"`
}

// YearRow is an entity established for one year, with its score
// (risk-service models.AuditUniverseYear). RiskLevel is "" while the row has
// never been scored.
type YearRow struct {
	CorporateAuditUniverseID string     `json:"corporate_audit_universe_id"`
	Year                     int        `json:"year"`
	RiskIndex                float64    `json:"risk_index"`
	RiskLevel                string     `json:"risk_level"`
	AuditPriority            bool       `json:"audit_priority"`
	ScoredAt                 *time.Time `json:"scored_at"`
	CorporateAuditUniverse   *Entity    `json:"corporate_audit_universe,omitempty"`
}

type apiResponse struct {
	Success bool            `json:"success"`
	Error   string          `json:"error"`
	Data    json.RawMessage `json:"data"`
}

// RiskClient talks to one risk-service base URL. It caches nothing.
type RiskClient struct {
	baseURL    string
	httpClient *http.Client
}

// New returns a client for baseURL ("" makes every call fail with
// ErrNotConfigured). hc nil uses a 10s timeout client.
func New(baseURL string, hc *http.Client) *RiskClient {
	if hc == nil {
		hc = &http.Client{Timeout: 10 * time.Second}
	}
	return &RiskClient{
		baseURL:    strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		httpClient: hc,
	}
}

// NewFromEnv returns a client for this stack's risk-service
// (RISK_SERVICE_URL, else derived from KAFKA_SERVICE_NAME, else
// http://risk-service:8002; see serviceurl.Resolve).
func NewFromEnv(hc *http.Client) *RiskClient {
	return New(serviceurl.Risk(), hc)
}

// BaseURL is the risk-service base URL this client calls
func (c *RiskClient) BaseURL() string { return c.baseURL }

// getData performs GET <base><path> with the caller's token and returns the
// envelope's data. Any non-200 answer is an error.
func (c *RiskClient) getData(ctx context.Context, authorization, path string) (json.RawMessage, error) {
	if strings.TrimSpace(authorization) == "" {
		return nil, ErrNoAuthorization
	}
	if c.baseURL == "" {
		return nil, ErrNotConfigured
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
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
		return nil, fmt.Errorf("risk-service answered %d for %s", resp.StatusCode, path)
	}

	var body apiResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxBodyBytes)).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode risk-service response: %w", err)
	}
	if !body.Success {
		return nil, fmt.Errorf("risk-service reported failure for %s: %s", path, body.Error)
	}
	return body.Data, nil
}

// ListCorporateEntities returns every corporate Audit Universe node that
// GET /api/v1/audit-universe/corporate returns, flattened (roots and their
// children, at whatever depth risk-service preloads; it preloads two levels
// below the roots today).
func (c *RiskClient) ListCorporateEntities(ctx context.Context, authorization string) ([]Entity, error) {
	data, err := c.getData(ctx, authorization, "/api/v1/audit-universe/corporate")
	if err != nil {
		return nil, err
	}
	var roots []Entity
	if err := json.Unmarshal(data, &roots); err != nil {
		return nil, fmt.Errorf("decode corporate audit universe: %w", err)
	}
	return Flatten(roots), nil
}

// ListYearUniverse returns the entities established for year with their
// scores (GET /api/v1/audit-universe/year/:year).
func (c *RiskClient) ListYearUniverse(ctx context.Context, authorization string, year int) ([]YearRow, error) {
	if year <= 0 {
		return nil, fmt.Errorf("invalid year %d", year)
	}
	data, err := c.getData(ctx, authorization, "/api/v1/audit-universe/year/"+url.PathEscape(strconv.Itoa(year)))
	if err != nil {
		return nil, err
	}
	var rows []YearRow
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, fmt.Errorf("decode yearly audit universe %d: %w", year, err)
	}
	return rows, nil
}

// Flatten returns every node of the trees, parents before children, with
// Children cleared.
func Flatten(roots []Entity) []Entity {
	var out []Entity
	var walk func(nodes []Entity)
	walk = func(nodes []Entity) {
		for _, n := range nodes {
			children := n.Children
			n.Children = nil
			out = append(out, n)
			walk(children)
		}
	}
	walk(roots)
	return out
}
