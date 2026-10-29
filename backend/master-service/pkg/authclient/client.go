// Package authclient calls auth-service's user administration endpoints.
//
// master-service creates the login account for an employee right after the
// employee row is saved. The admin's bearer token is forwarded, so auth-service
// applies its own ADMIN check. The caller's Authorization header travels in the
// request context (see WithAuthorization) so the service layer does not depend
// on gin.
package authclient

import (
	"bytes"
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

// maxBody bounds how much of an auth-service response is read.
const maxBody = 1 << 20

type authorizationKey struct{}

// WithAuthorization returns a context that carries the caller's Authorization
// header value ("Bearer <jwt>") for the outgoing auth-service calls.
func WithAuthorization(ctx context.Context, header string) context.Context {
	return context.WithValue(ctx, authorizationKey{}, header)
}

// AuthorizationFrom returns the Authorization header stored by WithAuthorization,
// or "" when there is none.
func AuthorizationFrom(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(authorizationKey{}).(string)
	return v
}

// APIError is a non-2xx answer from auth-service. Message is the human-readable
// message auth-service returned. It never contains a password.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("auth-service responded %d (%s): %s", e.StatusCode, e.Code, e.Message)
}

// User is the subset of an auth-service account that master-service reads back.
type User struct {
	ID         string `json:"id"`
	EmployeeID string `json:"employee_id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
}

// CreateUserInput is the body of POST /api/v1/users.
type CreateUserInput struct {
	Username         string   `json:"username"`
	Email            string   `json:"email"`
	FullName         string   `json:"full_name,omitempty"`
	Roles            []string `json:"roles"`
	EmployeeID       string   `json:"employee_id"`
	GeneratePassword bool     `json:"generate_password"`
}

// CreatedUser is the answer to a successful account creation. TemporaryPassword
// is set only when GeneratePassword was requested.
type CreatedUser struct {
	ID                string `json:"id"`
	Username          string `json:"username"`
	TemporaryPassword string `json:"temporary_password"`
}

// Client talks to auth-service over HTTP.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient builds a client for the auth-service at baseURL.
func NewClient(baseURL string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: timeout},
	}
}

// FindUserByEmail returns the account with that email, or (nil, nil) when there
// is none. authHeader is the caller's full Authorization header value.
func (c *Client) FindUserByEmail(ctx context.Context, email, authHeader string) (*User, error) {
	var out struct {
		Data User `json:"data"`
	}
	query := url.Values{"email": {email}}
	err := c.do(ctx, http.MethodGet, "/api/v1/users/lookup?"+query.Encode(), authHeader, nil, &out)
	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if out.Data.ID == "" {
		return nil, nil
	}
	return &out.Data, nil
}

// CreateUser creates a login account. With GeneratePassword set, the answer
// carries the one-time temporary password.
func (c *Client) CreateUser(ctx context.Context, in CreateUserInput, authHeader string) (*CreatedUser, error) {
	var out struct {
		Data CreatedUser `json:"data"`
	}
	if err := c.do(ctx, http.MethodPost, "/api/v1/users", authHeader, in, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// LinkEmployee sets employee_id on an existing account.
func (c *Client) LinkEmployee(ctx context.Context, userID, employeeCode, authHeader string) error {
	body := map[string]string{"employee_code": employeeCode}
	return c.do(ctx, http.MethodPost, "/api/v1/users/"+url.PathEscape(userID)+"/link-employee", authHeader, body, nil)
}

// do sends one request. A non-2xx answer becomes *APIError. A 2xx answer is
// decoded into out when out is non-nil. Transport failures are returned as-is.
func (c *Client) do(ctx context.Context, method, path, authHeader string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		var envelope struct {
			Message string `json:"message"`
			Error   *struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal(payload, &envelope) == nil {
			apiErr.Message = envelope.Message
			if envelope.Error != nil {
				apiErr.Code = envelope.Error.Code
				if envelope.Error.Message != "" {
					apiErr.Message = envelope.Error.Message
				}
			}
		}
		return apiErr
	}

	if out != nil && len(payload) > 0 {
		if err := json.Unmarshal(payload, out); err != nil {
			return fmt.Errorf("decode auth-service response: %w", err)
		}
	}
	return nil
}
