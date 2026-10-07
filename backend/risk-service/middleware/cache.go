package middleware

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// CacheConfig defines cache parameters for risk service endpoints
type CacheConfig struct {
	MaxAge   int    // Time-to-live in seconds
	IsPublic bool   // Whether cache is public or private
	VaryBy   string // Vary header value (e.g., "Accept-Encoding")
}

// DefaultCacheConfigs maps risk service routes to their cache configurations
var DefaultCacheConfigs = map[string]CacheConfig{
	// Risk Endpoints - dashboard data
	"GET /api/v1/risks": {
		MaxAge:   300, // 5 minutes
		IsPublic: true,
		VaryBy:   "Accept-Encoding",
	},
	// Mitigation Endpoints
	"GET /api/v1/mitigations": {
		MaxAge:   300, // 5 minutes
		IsPublic: true,
		VaryBy:   "Accept-Encoding",
	},
	// Risk Factors
	"GET /api/v1/risk-factors/standard": {
		MaxAge:   1800, // 30 minutes - stable data
		IsPublic: true,
		VaryBy:   "Accept-Encoding",
	},
	"GET /api/v1/risk-factors/corporate": {
		MaxAge:   600, // 10 minutes - may change
		IsPublic: true,
		VaryBy:   "Accept-Encoding",
	},
	// Audit Universe
	"GET /api/v1/audit-universe/standard": {
		MaxAge:   1800, // 30 minutes - stable data
		IsPublic: true,
		VaryBy:   "Accept-Encoding",
	},
	"GET /api/v1/audit-universe/corporate": {
		MaxAge:   600, // 10 minutes
		IsPublic: true,
		VaryBy:   "Accept-Encoding",
	},
	// RCM - Risk Control Matrix (Dashboard KPI data)
	"GET /api/v1/rcm": {
		MaxAge:   600, // 10 minutes
		IsPublic: true,
		VaryBy:   "Accept-Encoding",
	},
	"GET /api/v1/rcm/summary": {
		MaxAge:   300, // 5 minutes - dashboard refresh
		IsPublic: true,
		VaryBy:   "Accept-Encoding",
	},
}

// responseWriter captures response data for cache header generation
type responseWriter struct {
	gin.ResponseWriter
	body []byte
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}

// ResponseCache returns a middleware that sets cache headers on risk service responses
func ResponseCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only cache GET and HEAD requests
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.Next()
			return
		}

		// Get cache config for this route, if any
		routeKey := fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)
		config, exists := DefaultCacheConfigs[routeKey]
		if !exists {
			// No cache config for this route
			c.Next()
			return
		}

		// Wrap response writer to capture body
		wrappedWriter := &responseWriter{ResponseWriter: c.Writer}
		c.Writer = wrappedWriter

		// Process request
		c.Next()

		// Only cache successful responses (2xx)
		if c.Writer.Status() < 200 || c.Writer.Status() >= 300 {
			return
		}

		// Generate ETag from response body
		hash := md5.New()
		io.WriteString(hash, fmt.Sprintf("%s-%d", string(wrappedWriter.body), time.Now().Unix()))
		etag := fmt.Sprintf(`"%x"`, hash.Sum(nil))

		// Set cache headers
		visibility := "public"
		if !config.IsPublic {
			visibility = "private"
		}

		c.Header("Cache-Control", fmt.Sprintf("%s, max-age=%d", visibility, config.MaxAge))
		c.Header("ETag", etag)
		c.Header("Vary", config.VaryBy)
		c.Header("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
		c.Header("Expires", time.Now().Add(time.Duration(config.MaxAge)*time.Second).UTC().Format(http.TimeFormat))
	}
}

// SetCacheControl is a helper to set cache headers for a specific response
func SetCacheControl(c *gin.Context, maxAge int, isPublic bool) {
	visibility := "public"
	if !isPublic {
		visibility = "private"
	}

	c.Header("Cache-Control", fmt.Sprintf("%s, max-age=%d", visibility, maxAge))
	c.Header("Vary", "Accept-Encoding")
	c.Header("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
	c.Header("Expires", time.Now().Add(time.Duration(maxAge)*time.Second).UTC().Format(http.TimeFormat))
}

// SetETag generates and sets an ETag for the response
func SetETag(c *gin.Context, data string) {
	hash := md5.New()
	io.WriteString(hash, data)
	etag := fmt.Sprintf(`"%x"`, hash.Sum(nil))
	c.Header("ETag", etag)
}
