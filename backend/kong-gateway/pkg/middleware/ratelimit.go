package middleware

import (
	"net/http"
	"sync"
	"time"

	"kong-gateway/pkg/config"
	"kong-gateway/pkg/response"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	requestsPerMin int
	burstSize      int
	cleanupInterval time.Duration
	buckets        map[string]*bucket
	mu             sync.RWMutex
	stopCleanup    chan struct{}
}

type bucket struct {
	tokens     float64
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(cfg *config.RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		requestsPerMin: cfg.RequestsPerMin,
		burstSize:      cfg.BurstSize,
		cleanupInterval: time.Duration(cfg.CleanupInterval) * time.Second,
		buckets:        make(map[string]*bucket),
		stopCleanup:    make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// cleanup removes stale buckets
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			for key, b := range rl.buckets {
				b.mu.Lock()
				if time.Since(b.lastRefill) > rl.cleanupInterval*2 {
					delete(rl.buckets, key)
				}
				b.mu.Unlock()
			}
			rl.mu.Unlock()
		case <-rl.stopCleanup:
			return
		}
	}
}

// Stop stops the cleanup goroutine
func (rl *RateLimiter) Stop() {
	close(rl.stopCleanup)
}

// getBucket gets or creates a bucket for a key
func (rl *RateLimiter) getBucket(key string) *bucket {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if b, ok := rl.buckets[key]; ok {
		return b
	}

	b := &bucket{
		tokens:     float64(rl.burstSize),
		lastRefill: time.Now(),
	}
	rl.buckets[key] = b
	return b
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow(key string) (bool, int, int64) {
	b := rl.getBucket(key)
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()

	// Refill tokens based on time elapsed
	refillRate := float64(rl.requestsPerMin) / 60.0
	b.tokens += elapsed * refillRate
	if b.tokens > float64(rl.burstSize) {
		b.tokens = float64(rl.burstSize)
	}
	b.lastRefill = now

	// Check if request is allowed
	if b.tokens >= 1 {
		b.tokens--
		remaining := int(b.tokens)
		if remaining < 0 {
			remaining = 0
		}
		return true, remaining, now.Add(time.Duration(1/refillRate) * time.Second).Unix()
	}

	return false, 0, now.Add(time.Duration(1/refillRate) * time.Second).Unix()
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client identifier (IP or user ID)
		key := c.ClientIP()

		// Check rate limit
		allowed, remaining, reset := limiter.Allow(key)

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", string(rune(limiter.requestsPerMin)))
		c.Header("X-RateLimit-Remaining", string(rune(remaining)))
		c.Header("X-RateLimit-Reset", string(rune(reset)))

		if !allowed {
			c.Header("Retry-After", "60")
			response.TooManyRequests(c, "Rate limit exceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}
