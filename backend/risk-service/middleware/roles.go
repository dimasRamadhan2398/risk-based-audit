package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Role names exactly as seeded by auth-service and as they appear in the JWT
// "roles" claim. AUDIT_MANAGER and CHIEF_AUDIT_EXECUTIVE are not seeded; they
// only exist once an admin creates them. Matching is case-insensitive on the
// trimmed name, nothing fuzzier: an unrecognised spelling is refused.
const (
	RoleAdmin               = "ADMIN"
	RoleExecutive           = "EXECUTIVE"
	RoleChiefAuditExecutive = "CHIEF_AUDIT_EXECUTIVE"
	RoleDepartmentHead      = "DEPARTMENT_HEAD"
)

// UniverseFullWriters may write the whole audit universe and the standard
// libraries without department scoping. Controllers use it to tell a full
// writer from a department-scoped caller.
var UniverseFullWriters = []string{RoleAdmin, RoleExecutive, RoleChiefAuditExecutive}

// HasAnyRole reports whether any role on the token equals one of allowed
// (case-insensitive). A caller's rights are the union of their roles' rights,
// and VIEWER is never in an allow-list, so it contributes nothing:
// AUDITOR+VIEWER is exactly AUDITOR.
func HasAnyRole(c *gin.Context, allowed ...string) bool {
	v, _ := c.Get("roles")
	roles, _ := v.([]string)
	for _, r := range roles {
		for _, a := range allowed {
			if strings.EqualFold(strings.TrimSpace(r), a) {
				return true
			}
		}
	}
	return false
}

// RequireAnyRole refuses callers holding none of allowed. Authenticate must
// have run first.
func RequireAnyRole(allowed ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !HasAnyRole(c, allowed...) {
			forbidden(c, "Your role is not permitted to perform this action")
			return
		}
		c.Next()
	}
}

// RequireAnyRoleOnWrites applies RequireAnyRole to every method except
// GET/HEAD/OPTIONS. Mounted on a route group it covers routes added later by
// default, and leaves reads to the other guards.
func RequireAnyRoleOnWrites(allowed ...string) gin.HandlerFunc {
	check := RequireAnyRole(allowed...)
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
		default:
			check(c)
		}
	}
}
