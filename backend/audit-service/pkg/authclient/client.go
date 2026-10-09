// Package authclient reads users from auth-service on behalf of the caller.
//
// audit-service has no service account for auth-service, so a request is made
// with the caller's own bearer token (forwarded unchanged): auth-service then
// applies its own role check, and the lookup can never see more than the
// caller could fetch directly.
package authclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// DefaultBaseURL is used when AUTH_SERVICE_URL is unset: the auth-service
// container of the main stack. Tenant stacks must set AUTH_SERVICE_URL to
// their own auth-service (http://auth-service-<slug>:8001), because all
// stacks share the rb_audit_network docker network.
const DefaultBaseURL = "http://auth-service:8001"

// ErrNotFound means auth-service answered and has no active user with the id
var ErrNotFound = errors.New("no active user with this id")

// User is an assignable user as returned by GET /api/v1/users/assignable
type User struct {
	ID         string `json:"id"`
	FullName   string `json:"full_name"`
	Department string `json:"department"`
	Position   string `json:"position"`
}

// Lookup resolves an assignable user. authorization is the caller's
// Authorization header value ("Bearer ...").
type Lookup interface {
	AssignableUser(ctx context.Context, authorization, id string) (*User, error)
}

// Client calls auth-service over HTTP
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// New returns a client for AUTH_SERVICE_URL (or DefaultBaseURL)
func New() *Client {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("AUTH_SERVICE_URL")), "/")
	if base == "" {
		base = DefaultBaseURL
	}
	return &Client{BaseURL: base, HTTPClient: &http.Client{Timeout: 3 * time.Second}}
}

// AssignableUser calls GET /api/v1/users/assignable?id=<id>. It returns
// ErrNotFound when auth-service answers 200 without that user (unknown,
// inactive or deleted); any other failure (unreachable, 401/403, 5xx, bad
// body) is returned as a plain error, meaning "could not check".
func (c *Client) AssignableUser(ctx context.Context, authorization, id string) (*User, error) {
	if strings.TrimSpace(authorization) == "" {
		return nil, errors.New("no authorization to forward")
	}
	q := url.Values{}
	q.Set("id", id)
	q.Set("page_size", "1")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/api/v1/users/assignable?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authorization)
	req.Header.Set("Accept", "application/json")

	hc := c.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: 3 * time.Second}
	}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth-service answered %d", resp.StatusCode)
	}

	var body struct {
		Data struct {
			Users []User `json:"users"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode auth-service response: %w", err)
	}
	for i := range body.Data.Users {
		if strings.EqualFold(body.Data.Users[i].ID, id) {
			return &body.Data.Users[i], nil
		}
	}
	return nil, ErrNotFound
}
