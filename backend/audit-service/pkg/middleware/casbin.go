package middleware

import (
	"net/http"

	"audit-service/pkg/permissions"

	"github.com/gin-gonic/gin"
)

func CasbinMiddleware(enforcer *permissions.CasbinEnforcer) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// get userID from JWT context (set by auth middleware)
		userID, exists := ctx.Get("user_id")
		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			})
			ctx.Abort()
			return
		}

		path := ctx.FullPath()   // e.g. /audits/:id → /audits/*
		method := ctx.Request.Method

		allowed, err := enforcer.Enforce(userID.(string), path, method)
		if err != nil || !allowed {
			ctx.JSON(http.StatusForbidden, gin.H{
				"code":    "FORBIDDEN",
				"message": "You do not have permission to perform this action",
			})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}