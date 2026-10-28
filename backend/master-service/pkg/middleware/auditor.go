package middleware

import (
	"net/http"
	"strings"

	"master-service/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuditorRole is the seeded auditor role. It reads master data (lookups,
// dropdowns) but never writes it.
const AuditorRole = "AUDITOR"

// IsAuditorOnly reports whether the caller holds AUDITOR (case-insensitive) and
// no other role. A caller who also holds another role keeps that role's rights,
// same rule as IsViewerOnly. Authenticate must have run first.
func IsAuditorOnly(c *gin.Context) bool {
	v, _ := c.Get("roles")
	roles, _ := v.([]string)
	if len(roles) == 0 {
		return false
	}
	for _, r := range roles {
		if !strings.EqualFold(strings.TrimSpace(r), AuditorRole) {
			return false
		}
	}
	return true
}

// DenyAuditorWrites refuses an AUDITOR-only caller on every method except the
// safe read methods. Mount it on master-data route groups only.
func DenyAuditorWrites() gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}
		if IsAuditorOnly(c) {
			response.Forbidden(c, "Auditor accounts cannot modify master data")
			c.Abort()
			return
		}
		c.Next()
	}
}
