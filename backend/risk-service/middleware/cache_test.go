package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRiskFactorAndUniverseListsAreNotCached(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ResponseCache())
	paths := []string{
		"/api/v1/risk-factors/standard",
		"/api/v1/risk-factors/corporate",
		"/api/v1/audit-universe/standard",
		"/api/v1/audit-universe/corporate",
		"/api/v1/mitigations",
	}
	for _, p := range paths {
		r.GET(p, func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"success": true}) })
	}

	for _, p := range paths {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, p, nil))
		// Result() returns the headers as sent with the status line.
		got := w.Result().Header.Get("Cache-Control")
		if got != "private, no-cache" {
			t.Errorf("%s: Cache-Control = %q, want %q", p, got, "private, no-cache")
		}
	}
}
