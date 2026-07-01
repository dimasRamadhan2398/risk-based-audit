package gateway

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"kong-gateway/pkg/config"
	"kong-gateway/pkg/logger"
	"kong-gateway/pkg/models"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Proxy handles the reverse proxy logic
type Proxy struct {
	config    *config.Config
	services  map[string]*models.Service
	stats     *models.GatewayStats
	mu        sync.RWMutex
	client    *http.Client
	startTime time.Time
}

// NewProxy creates a new proxy instance
func NewProxy(cfg *config.Config) *Proxy {
	// Create HTTP client with timeouts
	client := &http.Client{
		Timeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// Initialize services from config
	services := make(map[string]*models.Service)
	for _, svc := range cfg.Upstream.Services {
		services[svc.Name] = &models.Service{
			ID:          uuid.New().String(),
			Name:        svc.Name,
			URL:         svc.URL,
			Prefix:      svc.Prefix,
			Methods:     svc.Methods,
			Headers:     svc.Headers,
			Timeout:     svc.Timeout,
			RetryCount:  svc.RetryCount,
			HealthCheck: svc.HealthCheck,
			IsHealthy:   true,
			LastChecked: time.Now(),
		}
	}

	stats := &models.GatewayStats{
		ServiceStats: make(map[string]*models.ServiceStats),
		StartedAt:    time.Now(),
	}
	for name := range services {
		stats.ServiceStats[name] = &models.ServiceStats{
			Name: name,
		}
	}

	proxy := &Proxy{
		config:    cfg,
		services:  services,
		stats:     stats,
		client:    client,
		startTime: time.Now(),
	}

	// Start health checkers
	go proxy.startHealthChecks()

	return proxy
}

// ServeHTTP handles incoming HTTP requests
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	requestID := uuid.New().String()

	// Log request
	logger.Info("Incoming request",
		logger.LogField("method", r.Method),
		logger.LogField("path", r.PathValue),
		logger.LogField("request_id", requestID),
	)

	// Find matching service
	service, path, err := p.findService(r.URL.Path)
	if err != nil {
		logger.Warn("Service not found", logger.LogField("path", r.URL.Path))
		http.Error(w, "Service not found", http.StatusNotFound)
		p.incrementError("unknown")
		return
	}

	// Check if service is healthy
	if !service.IsHealthy {
		logger.Warn("Service unhealthy", logger.LogField("service", service.Name))
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		p.incrementError(service.Name)
		return
	}

	// Check method
	if !p.isMethodAllowed(service, r.Method) {
		logger.Warn("Method not allowed",
			logger.LogField("method", r.Method),
			logger.LogField("service", service.Name),
		)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Build upstream URL
	upstreamURL, err := p.buildUpstreamURL(service, path, r.URL.RawQuery)
	if err != nil {
		logger.Error("Failed to build upstream URL", logger.LogField("error", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create proxy request
	req, err := p.createProxyRequest(r, upstreamURL, service, requestID)
	if err != nil {
		logger.Error("Failed to create proxy request", logger.LogField("error", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Execute request with retries
	resp, err := p.executeWithRetry(req, service)
	if err != nil {
		logger.Error("Proxy request failed",
			logger.LogField("error", err),
			logger.LogField("service", service.Name),
		)
		http.Error(w, "Bad gateway", http.StatusBadGateway)
		p.incrementError(service.Name)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Add custom headers
	w.Header().Set("X-Request-Id", requestID)
	w.Header().Set("X-Gateway", "kong-gateway")
	w.Header().Set("X-Upstream", service.Name)

	// Copy response body
	w.WriteHeader(resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response body", logger.LogField("error", err))
		return
	}
	w.Write(body)

	// Update stats
	latency := time.Since(startTime)
	p.updateStats(service.Name, resp.StatusCode, latency)

	logger.Info("Request completed",
		logger.LogField("method", r.Method),
		logger.LogField("path", r.URL.Path),
		logger.LogField("upstream", service.Name),
		logger.LogField("status", resp.StatusCode),
		logger.LogField("latency_ms", latency.Milliseconds()),
		logger.LogField("request_id", requestID),
	)
}

// findService finds the appropriate service for the given path
func (p *Proxy) findService(path string) (*models.Service, string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var bestMatch *models.Service
	var bestMatchLen int

	for _, svc := range p.services {
		if strings.HasPrefix(path, svc.Prefix) {
			if len(svc.Prefix) > bestMatchLen {
				bestMatch = svc
				bestMatchLen = len(svc.Prefix)
			}
		}
	}

	if bestMatch == nil {
		return nil, "", fmt.Errorf("no service found for path: %s", path)
	}

	// Extract the remaining path
	remainingPath := strings.TrimPrefix(path, bestMatch.Prefix)
	if !strings.HasPrefix(remainingPath, "/") {
		remainingPath = "/" + remainingPath
	}

	return bestMatch, remainingPath, nil
}

// isMethodAllowed checks if the method is allowed for the service
func (p *Proxy) isMethodAllowed(service *models.Service, method string) bool {
	if len(service.Methods) == 0 {
		return true
	}

	for _, m := range service.Methods {
		if strings.EqualFold(m, method) {
			return true
		}
	}
	return false
}

// buildUpstreamURL builds the upstream URL for the request
func (p *Proxy) buildUpstreamURL(service *models.Service, path, query string) (string, error) {
	upstreamURL := service.URL

	// Parse the base URL
	baseURL, err := url.Parse(upstreamURL)
	if err != nil {
		return "", fmt.Errorf("invalid upstream URL: %w", err)
	}

	// Set path
	baseURL.Path = path

	// Add query string if present
	if query != "" {
		baseURL.RawQuery = query
	}

	return baseURL.String(), nil
}

// createProxyRequest creates a new HTTP request to the upstream service
func (p *Proxy) createProxyRequest(r *http.Request, upstreamURL string, service *models.Service, requestID string) (*http.Request, error) {
	// Read body
	var body []byte
	if r.Body != nil {
		body, _ = io.ReadAll(r.Body)
		r.Body.Close()
		r.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	// Create new request
	req, err := http.NewRequest(r.Method, upstreamURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// Copy headers
	for key, values := range r.Header {
		for _, value := range values {
			// Skip hop-by-hop headers
			if !isHopByHopHeader(key) {
				req.Header.Add(key, value)
			}
		}
	}

	// Add service-specific headers
	for key, value := range service.Headers {
		req.Header.Set(key, value)
	}

	// Add gateway headers
	req.Header.Set("X-Request-Id", requestID)
	req.Header.Set("X-Forwarded-For", r.RemoteAddr)
	req.Header.Set("X-Forwarded-Host", r.Host)

	return req, nil
}

// executeWithRetry executes the request with retries
func (p *Proxy) executeWithRetry(req *http.Request, service *models.Service) (*http.Response, error) {
	var lastErr error
	maxRetries := service.RetryCount

	if maxRetries < 1 {
		maxRetries = 1
	}

	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			logger.Warn("Retrying request",
				logger.LogField("attempt", i+1),
				logger.LogField("max_retries", maxRetries),
			)
			time.Sleep(time.Duration(i*100) * time.Millisecond)
		}

		ctx, cancel := context.WithTimeout(req.Context(), time.Duration(service.Timeout)*time.Second)
		defer cancel()

		req = req.WithContext(ctx)

		resp, err := p.client.Do(req)
		if err == nil {
			return resp, nil
		}

		lastErr = err
	}

	return nil, lastErr
}

// startHealthChecks starts health check routines for all services
func (p *Proxy) startHealthChecks() {
	for _, svc := range p.services {
		if svc.HealthCheck.Enabled {
			go p.healthCheckLoop(svc)
		}
	}
}

// healthCheckLoop runs periodic health checks for a service
func (p *Proxy) healthCheckLoop(service *models.Service) {
	ticker := time.NewTicker(time.Duration(service.HealthCheck.Interval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		p.checkServiceHealth(service)
	}
}

// checkServiceHealth checks if a service is healthy
func (p *Proxy) checkServiceHealth(service *models.Service) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(service.HealthCheck.Timeout)*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", service.URL+"/health", nil)
	if err != nil {
		p.setServiceHealth(service.Name, false, "Failed to create request")
		return
	}

	resp, err := p.client.Do(req)
	if err != nil {
		p.setServiceHealth(service.Name, false, err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		p.setServiceHealth(service.Name, true, "")
	} else {
		p.setServiceHealth(service.Name, false, fmt.Sprintf("Status: %d", resp.StatusCode))
	}
}

// setServiceHealth updates the health status of a service
func (p *Proxy) setServiceHealth(name string, healthy bool, message string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if svc, ok := p.services[name]; ok {
		svc.IsHealthy = healthy
		svc.LastChecked = time.Now()

		status := "healthy"
		if !healthy {
			status = "unhealthy"
		}

		logger.Debug("Service health updated",
			logger.LogField("service", name),
			logger.LogField("status", status),
			logger.LogField("message", message),
		)
	}
}

// updateStats updates gateway statistics
func (p *Proxy) updateStats(serviceName string, statusCode int, latency time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.stats.TotalRequests++
	p.stats.Uptime = time.Since(p.startTime)

	if svcStats, ok := p.stats.ServiceStats[serviceName]; ok {
		svcStats.Requests++
		svcStats.LastRequestAt = time.Now()

		// Update latency stats
		latencyMs := latency.Milliseconds()
		if svcStats.MinLatency == 0 || latencyMs < svcStats.MinLatency {
			svcStats.MinLatency = latencyMs
		}
		if latencyMs > svcStats.MaxLatency {
			svcStats.MaxLatency = latencyMs
		}
		// Running average
		svcStats.AvgLatency = (svcStats.AvgLatency*(svcStats.Requests-1) + latencyMs) / svcStats.Requests
	}

	if statusCode >= 400 {
		p.stats.TotalErrors++
		if svcStats, ok := p.stats.ServiceStats[serviceName]; ok {
			svcStats.Errors++
		}
	}
}

// incrementError increments error count for a service
func (p *Proxy) incrementError(serviceName string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.stats.TotalErrors++
	if svcStats, ok := p.stats.ServiceStats[serviceName]; ok {
		svcStats.Errors++
	}
}

// GetStats returns current gateway statistics
func (p *Proxy) GetStats() *models.GatewayStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return &models.GatewayStats{
		TotalRequests:  p.stats.TotalRequests,
		ActiveRequests: p.stats.ActiveRequests,
		TotalErrors:    p.stats.TotalErrors,
		ServiceStats:   p.stats.ServiceStats,
		Uptime:         time.Since(p.startTime),
		StartedAt:      p.stats.StartedAt,
	}
}

// GetHealthStatus returns the health status of the gateway
func (p *Proxy) GetHealthStatus() *models.HealthStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()

	status := "healthy"
	services := make(map[string]models.ServiceHealth)

	for name, svc := range p.services {
		health := models.ServiceHealth{
			Name:    name,
			Healthy: svc.IsHealthy,
		}

		if !svc.IsHealthy {
			status = "degraded"
		}

		services[name] = health
	}

	return &models.HealthStatus{
		Status:   status,
		Services: services,
		Uptime:   time.Since(p.startTime),
	}
}

// GetServices returns all configured services
func (p *Proxy) GetServices() map[string]*models.Service {
	p.mu.RLock()
	defer p.mu.RUnlock()

	services := make(map[string]*models.Service)
	for name, svc := range p.services {
		services[name] = svc
	}
	return services
}

// isHopByHopHeader checks if a header is hop-by-hop
func isHopByHopHeader(name string) bool {
	hopByHopHeaders := map[string]bool{
		"Connection":          true,
		"Keep-Alive":           true,
		"Proxy-Authenticate":   true,
		"Proxy-Authorization":  true,
		"Te":                   true,
		"Trailers":             true,
		"Transfer-Encoding":    true,
		"Upgrade":              true,
		"X-Accel-Buffering":    true,
		"X-Request-Id":         true,
	}
	return hopByHopHeaders[name]
}
