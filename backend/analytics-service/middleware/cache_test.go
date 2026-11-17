package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCAATTRoutesDisableStaleHTTPCache(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ResponseCache())
	r.GET("/api/analytics/caatt/benford", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})
	r.GET("/api/analytics/caatt/full-population", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	for _, path := range []string{
		"/api/analytics/caatt/benford?source_id=src-1",
		"/api/analytics/caatt/full-population?source_id=src-1",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s returned %d", path, rec.Code)
		}
		cc := rec.Header().Get("Cache-Control")
		if !strings.Contains(cc, "no-store") {
			t.Fatalf("%s expected Cache-Control to contain no-store, got %q", path, cc)
		}
		if rec.Header().Get("ETag") != "" {
			t.Fatalf("%s should not emit ETag when MaxAge <= 0, got %q", path, rec.Header().Get("ETag"))
		}
	}
}

