package middleware

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// CacheConfig defines cache parameters for master service endpoints
type CacheConfig struct {
	MaxAge   int    // Time-to-live in seconds
	// IsPublic controls the Cache-Control visibility token. Keep this false for
	// anything behind authentication: "public" lets Kong proxy-cache, a CDN or
	// any shared proxy store the response and replay it to a caller who never
	// presented a token. Every route below is authenticated, so all are private.
	IsPublic bool
	VaryBy   string // Vary header value (e.g., "Accept-Encoding")
}

// DefaultCacheConfigs maps master service routes to their cache configurations
var DefaultCacheConfigs = map[string]CacheConfig{
	// Organizational Master Data - very stable, infrequently updated
	"GET /api/v1/companies": {
		MaxAge:   3600, // 1 hour
		IsPublic: false,
		VaryBy:   "Accept-Encoding, Authorization",
	},
	"GET /api/v1/business-units": {
		MaxAge:   3600, // 1 hour
		IsPublic: false,
		VaryBy:   "Accept-Encoding, Authorization",
	},
	"GET /api/v1/departments": {
		MaxAge:   3600, // 1 hour
		IsPublic: false,
		VaryBy:   "Accept-Encoding, Authorization",
	},
	"GET /api/v1/employees": {
		MaxAge:   3600, // 1 hour
		IsPublic: false,
		VaryBy:   "Accept-Encoding, Authorization",
	},
	"GET /api/v1/job-roles": {
		MaxAge:   3600, // 1 hour
		IsPublic: false,
		VaryBy:   "Accept-Encoding, Authorization",
	},
	"GET /api/v1/locations": {
		MaxAge:   3600, // 1 hour
		IsPublic: false,
		VaryBy:   "Accept-Encoding, Authorization",
	},
	// Quality Assurance Reports
	"GET /api/v1/quality-assurance": {
		MaxAge:   300, // 5 minutes
		IsPublic: false,
		VaryBy:   "Accept-Encoding, Authorization",
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

// ResponseCache returns a middleware that sets cache headers on master service responses
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
// Callers behind authentication must pass isPublic=false; see CacheConfig.IsPublic.
func SetCacheControl(c *gin.Context, maxAge int, isPublic bool) {
	visibility := "public"
	if !isPublic {
		visibility = "private"
	}

	c.Header("Cache-Control", fmt.Sprintf("%s, max-age=%d", visibility, maxAge))
	c.Header("Vary", "Accept-Encoding, Authorization")
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
