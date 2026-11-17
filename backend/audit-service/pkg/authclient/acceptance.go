package authclient

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

// MaxAcceptanceCodes is the most employee codes auth-service accepts in one
// GET /api/v1/confidentiality/acceptance-summary call
const MaxAcceptanceCodes = 500

// Acceptance is one row of the acceptance summary: whether the employee has a
// login account (has_user) and whether that account accepted the current
// SYSTEM integrity pact, unexpired (accepted). Counts only, no content.
type Acceptance struct {
	EmployeeCode string     `json:"employee_code"`
	HasUser      bool       `json:"has_user"`
	Accepted     bool       `json:"accepted"`
	AcceptedAt   *time.Time `json:"accepted_at"`
	Version      string     `json:"version"`
}

// AcceptanceSummary calls GET /api/v1/confidentiality/acceptance-summary
// ?employee_codes=a,b,c with the caller's token, in chunks of
// MaxAcceptanceCodes. Any failure is returned as an error ("could not check").
func (c *Client) AcceptanceSummary(ctx context.Context, authorization string, codes []string) ([]Acceptance, error) {
	if strings.TrimSpace(authorization) == "" {
		return nil, errors.New("no authorization to forward")
	}
	if c.BaseURL == "" {
		return nil, errNoBaseURL
	}
	var out []Acceptance
	for start := 0; start < len(codes); start += MaxAcceptanceCodes {
		end := start + MaxAcceptanceCodes
		if end > len(codes) {
			end = len(codes)
		}
		rows, err := c.acceptanceChunk(ctx, authorization, codes[start:end])
		if err != nil {
			return nil, err
		}
		out = append(out, rows...)
	}
	return out, nil
}

func (c *Client) acceptanceChunk(ctx context.Context, authorization string, codes []string) ([]Acceptance, error) {
	q := url.Values{}
	q.Set("employee_codes", strings.Join(codes, ","))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.BaseURL+"/api/v1/confidentiality/acceptance-summary?"+q.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authorization)
	req.Header.Set("Accept", "application/json")
	hc := c.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: 5 * time.Second}
	}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
		return nil, fmt.Errorf("auth-service answered %d", resp.StatusCode)
	}
	var body struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode auth-service response: %w", err)
	}
	// data is the bare array; an object wrapping it under "items" or
	// "results" is accepted too
	var rows []Acceptance
	if err := json.Unmarshal(body.Data, &rows); err == nil {
		return rows, nil
	}
	var wrapped struct {
		Items   []Acceptance `json:"items"`
		Results []Acceptance `json:"results"`
	}
	if err := json.Unmarshal(body.Data, &wrapped); err != nil {
		return nil, fmt.Errorf("decode acceptance summary: %w", err)
	}
	if wrapped.Items != nil {
		return wrapped.Items, nil
	}
	return wrapped.Results, nil
}
