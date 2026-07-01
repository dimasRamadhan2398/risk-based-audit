package handlers

import (
	"net/http"
	"time"

	"kong-gateway/pkg/gateway"
	"kong-gateway/pkg/response"

	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	proxy *gateway.Proxy
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(proxy *gateway.Proxy) *HealthHandler {
	return &HealthHandler{
		proxy: proxy,
	}
}

// Health returns basic health status
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "kong-gateway",
		"version": "1.0.0",
		"time":    time.Now().UTC(),
	})
}

// Ready checks if the gateway is ready to serve requests
func (h *HealthHandler) Ready(c *gin.Context) {
	health := h.proxy.GetHealthStatus()

	// Check if all services are healthy
	allHealthy := true
	for _, svc := range health.Services {
		if !svc.Healthy {
			allHealthy = false
			break
		}
	}

	status := "ready"
	statusCode := http.StatusOK
	if !allHealthy {
		status = "degraded"
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"status":    status,
		"services":  health.Services,
		"version":   "1.0.0",
		"time":      time.Now().UTC(),
	})
}

// Live checks if the gateway is alive
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}

// Stats returns gateway statistics
func (h *HealthHandler) Stats(c *gin.Context) {
	stats := h.proxy.GetStats()
	response.Success(c, stats)
}

// Services returns list of configured services
func (h *HealthHandler) Services(c *gin.Context) {
	services := h.proxy.GetServices()
	response.Success(c, services)
}

// HealthDetailed returns detailed health status
func (h *HealthHandler) HealthDetailed(c *gin.Context) {
	stats := h.proxy.GetStats()
	health := h.proxy.GetHealthStatus()
	services := h.proxy.GetServices()

	c.JSON(http.StatusOK, gin.H{
		"gateway": gin.H{
			"status":         health.Status,
			"uptime":         health.Uptime.String(),
			"total_requests": stats.TotalRequests,
			"total_errors":   stats.TotalErrors,
			"version":        "1.0.0",
		},
		"services": services,
		"stats":    stats,
		"time":     time.Now().UTC(),
	})
}
