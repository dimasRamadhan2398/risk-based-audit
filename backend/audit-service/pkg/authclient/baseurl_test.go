package authclient

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"audit-service/pkg/serviceurl"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestNewUsesStackURL(t *testing.T) {
	cases := []struct{ name, explicit, service, want string }{
		{"tenant derived", "", "audit-service-bai", "http://auth-service-bai:8001"},
		{"prod default", "", "", "http://auth-service:8001"},
		{"explicit wins", " http://localhost:8001/ ", "audit-service-bai", "http://localhost:8001"},
		{"malformed tenant name", "", "audit-service-", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(serviceurl.AuthServiceURLEnv, tc.explicit)
			t.Setenv(serviceurl.ServiceNameEnv, tc.service)
			if got := New().BaseURL; got != tc.want {
				t.Fatalf("BaseURL = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestTenantClientCallsTenantAuthService(t *testing.T) {
	t.Setenv(serviceurl.AuthServiceURLEnv, "")
	t.Setenv(serviceurl.ServiceNameEnv, "audit-service-bai")

	var gotHost, gotAuth string
	c := New()
	c.HTTPClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotHost, gotAuth = r.URL.Host, r.Header.Get("Authorization")
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{},
			Body: io.NopCloser(strings.NewReader(`{"data":{"users":[{"id":"u1","full_name":"Budi"}]}}`))}, nil
	})}

	u, err := c.AssignableUser(context.Background(), "Bearer t", "u1")
	if err != nil || u.ID != "u1" {
		t.Fatalf("user=%v err=%v", u, err)
	}
	if gotHost != "auth-service-bai:8001" {
		t.Fatalf("request went to %q, want auth-service-bai:8001", gotHost)
	}
	if gotAuth != "Bearer t" {
		t.Fatalf("Authorization = %q", gotAuth)
	}
}

func TestEmptyBaseURLNeverCalls(t *testing.T) {
	called := false
	c := &Client{BaseURL: "", HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		called = true
		return nil, io.EOF
	})}}
	if _, err := c.AssignableUser(context.Background(), "Bearer t", "u1"); err == nil {
		t.Fatal("want error without a base URL")
	}
	if _, err := c.Contacts(context.Background(), "Bearer t", []string{"u1"}, ""); err == nil {
		t.Fatal("want error without a base URL")
	}
	if called {
		t.Fatal("a request was sent without a base URL")
	}
}
