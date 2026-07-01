package middleware

import (
	"time"

	"kong-gateway/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-Id")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-Id", requestID)
		c.Next()
	}
}

// LoggerMiddleware logs request information
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log after request is processed
		latency := time.Since(startTime)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		requestID, _ := c.Get("request_id")

		fields := []interface{}{
			logger.LogField("method", method),
			logger.LogField("path", path),
			logger.LogField("status", statusCode),
			logger.LogField("latency_ms", latency.Milliseconds()),
			logger.LogField("client_ip", clientIP),
			logger.LogField("request_id", requestID),
		}

		if query != "" {
			fields = append(fields, logger.LogField("query", query))
		}

		// Log based on status code
		switch {
		case statusCode >= 500:
			logger.Error("Server error", fields...)
		case statusCode >= 400:
			logger.Warn("Client error", fields...)
		default:
			logger.Info("Request completed", fields...)
		}
	}
}

// RecoveryMiddleware recovers from panics and logs the error
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get("request_id")
				logger.Error("Panic recovered",
					logger.LogField("error", err),
					logger.LogField("path", c.Request.URL.Path),
					logger.LogField("method", c.Request.Method),
					logger.LogField("request_id", requestID),
				)
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}
