package middleware

import (
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"

	"github.com/gin-gonic/gin"
)

// RequireFlexible answers 404 ORG_STRUCTURE_NOT_ENABLED unless the tenant runs
// the flexible organisation structure. Legacy tenants never reach the handler,
// so the org tables (absent there) are never queried.
func RequireFlexible(flexible bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !flexible {
			base.RespondError(c, apperrors.ErrOrgStructureNotEnabled)
			c.Abort()
			return
		}
		c.Next()
	}
}
