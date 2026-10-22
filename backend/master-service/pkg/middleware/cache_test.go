package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func serveCached(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ResponseCache())
	ok := func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"success": true}) }
	r.GET("/api/v1/employees", ok)
	r.GET("/api/v1/companies", ok)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("GET %s status = %d, want 200", path, w.Code)
	}
	return w
}

func TestResponseCache_EmployeesTTLIs15Minutes(t *testing.T) {
	w := serveCached(t, "/api/v1/employees")
	if got := w.Header().Get("Cache-Control"); got != "private, max-age=900" {
		t.Fatalf("Cache-Control = %q, want %q", got, "private, max-age=900")
	}
}

// Other master-data routes share the middleware but keep their own TTL.
func TestResponseCache_OtherRoutesKeepOneHour(t *testing.T) {
	w := serveCached(t, "/api/v1/companies")
	if got := w.Header().Get("Cache-Control"); got != "private, max-age=3600" {
		t.Fatalf("Cache-Control = %q, want %q", got, "private, max-age=3600")
	}
}
