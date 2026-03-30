package middleware

import (
	"auth-service/pkg/errors"
	"auth-service/pkg/response"
	"net/http"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gin-gonic/gin"
)

func RateLimiter(lmt *limiter.Limiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := tollbooth.LimitByRequest(lmt, c.Writer, c.Request)
		if err != nil {
			c.JSON(http.StatusTooManyRequests, response.Response{
				Success: false,
				Message: errors.ErrInternalServer.Error(),
			})
			c.Abort()
		}
		c.Next()
	}
}
