package mailer

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewFromEnvNeedsAPIKey(t *testing.T) {
	t.Setenv("RESEND_API_KEY", "  ")
	if _, err := NewFromEnv(); err != ErrNotConfigured {
		t.Fatalf("err = %v, want ErrNotConfigured", err)
	}
	t.Setenv("RESEND_API_KEY", "re_test")
	t.Setenv("RESEND_FROM", "")
	m, err := NewFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if m.From != DefaultFrom {
		t.Fatalf("From = %q, want default", m.From)
	}
}

func TestResendSendPostsOneRecipient(t *testing.T) {
	var gotAuth string
	var got resendRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"id":"email_1"}`)
	}))
	defer srv.Close()

	m := &Resend{APIKey: "re_test", From: "AuditSphere <no-reply@mail.auditsphere.app>", Endpoint: srv.URL, HTTPClient: srv.Client()}
	err := m.Send(context.Background(), []string{"budi@example.id"}, "Pengingat\r\nBcc: evil@x.id", "isi")
	if err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer re_test" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
	if len(got.To) != 1 || got.To[0] != "budi@example.id" {
		t.Fatalf("to = %v", got.To)
	}
	if strings.ContainsAny(got.Subject, "\r\n") || got.Subject != "Pengingat Bcc: evil@x.id" {
		t.Fatalf("subject = %q", got.Subject)
	}
	if got.Text != "isi" || got.From != m.From {
		t.Fatalf("payload = %+v", got)
	}
}

func TestResendSendReportsProviderError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, "invalid api key")
	}))
	defer srv.Close()
	m := &Resend{APIKey: "bad", From: DefaultFrom, Endpoint: srv.URL, HTTPClient: srv.Client()}
	if err := m.Send(context.Background(), []string{"a@x.id"}, "s", "t"); err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("err = %v, want a 401 error", err)
	}
	if err := m.Send(context.Background(), nil, "s", "t"); err == nil {
		t.Fatal("no recipients must be an error")
	}
}
