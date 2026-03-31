package middleware

import (
	"runtime/debug"

	"auth-service/pkg/logger"
	"auth-service/pkg/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("stack", string(debug.Stack())),
				)

				// Return internal server error
				response.InternalServerError(c, "Internal server error")
				c.Abort()
			}
		}()

		c.Next()
	}
}
