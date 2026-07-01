package models

import (
	"net/http"
	"time"
)

// Service represents an upstream service
type Service struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	URL         string            `json:"url"`
	Prefix      string            `json:"prefix"`
	Methods     []string          `json:"methods"`
	Headers     map[string]string `json:"headers"`
	Timeout     int               `json:"timeout"`
	RetryCount  int               `json:"retry_count"`
	HealthCheck HealthCheck       `json:"health_check"`
	IsHealthy   bool              `json:"is_healthy"`
	LastChecked time.Time         `json:"last_checked"`
}

// HealthCheck represents health check configuration
type HealthCheck struct {
	Enabled  bool `json:"enabled"`
	Interval int  `json:"interval"`
	Timeout  int  `json:"timeout"`
}

// Route represents a gateway route
type Route struct {
	ID        string      `json:"id"`
	Service   string      `json:"service"`    // service name
	Path      string      `json:"path"`       // path pattern
	Methods   []string    `json:"methods"`    // allowed methods
	Headers   []HeaderRule `json:"headers"`   // header rules
	Plugins   []Plugin    `json:"plugins"`     // attached plugins
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// HeaderRule represents a header matching rule
type HeaderRule struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	Operator  string `json:"operator"` // eq, neq, regex
}

// Plugin represents a gateway plugin
type Plugin struct {
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
}

// ProxyRequest represents a request being proxied
type ProxyRequest struct {
	Method     string            `json:"method"`
	Path       string            `json:"path"`
	Query      string            `json:"query"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
	Service    string            `json:"service"`
	UpstreamURL string          `json:"upstream_url"`
	StartTime  time.Time        `json:"start_time"`
}

// ProxyResponse represents a proxied response
type ProxyResponse struct {
	StatusCode  int               `json:"status_code"`
	Headers     map[string]string `json:"headers"`
	Body        []byte            `json:"body"`
	Latency     time.Duration     `json:"latency"`
	Upstream   string            `json:"upstream"`
	Error       string            `json:"error,omitempty"`
}

// GatewayStats represents gateway statistics
type GatewayStats struct {
	TotalRequests    int64            `json:"total_requests"`
	ActiveRequests   int64            `json:"active_requests"`
	TotalErrors      int64            `json:"total_errors"`
	ServiceStats     map[string]*ServiceStats `json:"service_stats"`
	Uptime           time.Duration    `json:"uptime"`
	StartedAt        time.Time        `json:"started_at"`
}

// ServiceStats represents statistics for a single service
type ServiceStats struct {
	Name            string `json:"name"`
	Requests       int64  `json:"requests"`
	Errors         int64  `json:"errors"`
	AvgLatency     int64  `json:"avg_latency_ms"`
	MinLatency     int64  `json:"min_latency_ms"`
	MaxLatency     int64  `json:"max_latency_ms"`
	LastRequestAt  time.Time `json:"last_request_at"`
}

// HealthStatus represents the health status of the gateway
type HealthStatus struct {
	Status     string                 `json:"status"` // healthy, degraded, unhealthy
	Services   map[string]ServiceHealth `json:"services"`
	Version    string                 `json:"version"`
	Uptime     time.Duration          `json:"uptime"`
}

// ServiceHealth represents the health of a service
type ServiceHealth struct {
	Name      string `json:"name"`
	Healthy   bool   `json:"healthy"`
	Latency   int64  `json:"latency_ms"`
	Message   string `json:"message,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(message string, code int) ErrorResponse {
	return ErrorResponse{
		Error:   http.StatusText(code),
		Message: message,
		Code:    code,
	}
}

// RateLimitInfo represents rate limit information
type RateLimitInfo struct {
	Remaining int   `json:"remaining"`
	Limit     int   `json:"limit"`
	Reset     int64 `json:"reset"` // Unix timestamp
}
