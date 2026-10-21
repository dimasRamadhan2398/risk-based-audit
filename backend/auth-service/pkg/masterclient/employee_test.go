package masterclient

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFindByCode_ReturnsEmployeeAndForwardsAuthorization(t *testing.T) {
	var gotPath, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"message":"ok","data":{"id":"6f1c1b1e-0000-4000-8000-000000000001","employee_code":"EMP001","department_id":"6f1c1b1e-0000-4000-8000-000000000002"}}`))
	}))
	defer srv.Close()

	c := NewEmployeeClient(srv.URL+"/", time.Second)
	emp, err := c.FindByCode(context.Background(), "EMP001", "Bearer tok")
	if err != nil {
		t.Fatalf("FindByCode returned error: %v", err)
	}
	if emp.ID != "6f1c1b1e-0000-4000-8000-000000000001" || emp.DepartmentID != "6f1c1b1e-0000-4000-8000-000000000002" {
		t.Fatalf("unexpected employee: %+v", emp)
	}
	if gotPath != "/api/v1/employees/by-code/EMP001" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotAuth != "Bearer tok" {
		t.Fatalf("Authorization = %q, want forwarded bearer", gotAuth)
	}
}

func TestFindByCode_NotFoundAndFailures(t *testing.T) {
	status := http.StatusNotFound
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}))
	defer srv.Close()
	c := NewEmployeeClient(srv.URL, time.Second)

	if _, err := c.FindByCode(context.Background(), "EMP404", ""); !errors.Is(err, ErrEmployeeNotFound) {
		t.Fatalf("404 error = %v, want ErrEmployeeNotFound", err)
	}

	status = http.StatusInternalServerError
	if _, err := c.FindByCode(context.Background(), "EMP001", ""); err == nil || errors.Is(err, ErrEmployeeNotFound) {
		t.Fatalf("500 error = %v, want a non-not-found error", err)
	}

	if _, err := c.FindByCode(context.Background(), "  ", ""); !errors.Is(err, ErrEmployeeNotFound) {
		t.Fatalf("blank code error = %v, want ErrEmployeeNotFound", err)
	}
}

func TestBaseURL_DerivedPerTenant(t *testing.T) {
	// Covered in pkg/config; this only checks the client trims the trailing slash.
	if got := NewEmployeeClient("http://master-service:8002/", 0).BaseURL(); got != "http://master-service:8002" {
		t.Fatalf("BaseURL = %q", got)
	}
}
