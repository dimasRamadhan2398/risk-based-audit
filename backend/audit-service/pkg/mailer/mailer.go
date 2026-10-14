// Package mailer sends transactional email for audit-service (ATR reminders)
// through the Resend HTTP API, the provider auth-service already uses.
package mailer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// DefaultFrom is used when RESEND_FROM is unset (same as auth-service)
const DefaultFrom = "AuditSphere <no-reply@mail.auditsphere.app>"

// ErrNotConfigured means RESEND_API_KEY is unset: nothing can be sent
var ErrNotConfigured = errors.New("email is not configured (RESEND_API_KEY is unset)")

const resendEndpoint = "https://api.resend.com/emails"

// Sender sends one plain-text email to one or more recipients
type Sender interface {
	Send(ctx context.Context, to []string, subject, text string) error
}

// Resend sends through the Resend REST API
type Resend struct {
	APIKey     string
	From       string
	Endpoint   string
	HTTPClient *http.Client
}

// NewFromEnv returns a Resend sender configured from RESEND_API_KEY and
// RESEND_FROM, or ErrNotConfigured when there is no API key
func NewFromEnv() (*Resend, error) {
	key := strings.TrimSpace(os.Getenv("RESEND_API_KEY"))
	if key == "" {
		return nil, ErrNotConfigured
	}
	from := strings.TrimSpace(os.Getenv("RESEND_FROM"))
	if from == "" {
		from = DefaultFrom
	}
	return &Resend{APIKey: key, From: from, Endpoint: resendEndpoint, HTTPClient: &http.Client{Timeout: 12 * time.Second}}, nil
}

type resendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
}

// Send posts the email to Resend. Subject line breaks are removed so a
// finding title cannot inject headers.
func (r *Resend) Send(ctx context.Context, to []string, subject, text string) error {
	if len(to) == 0 {
		return errors.New("no recipients")
	}
	subject = strings.Join(strings.Fields(subject), " ")
	payload, err := json.Marshal(resendRequest{From: r.From, To: to, Subject: subject, Text: text})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.Endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+r.APIKey)
	req.Header.Set("Content-Type", "application/json")

	hc := r.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: 12 * time.Second}
	}
	resp, err := hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("resend answered %d: %s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}
	return nil
}
