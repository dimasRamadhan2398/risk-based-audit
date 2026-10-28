package middleware

import (
	"strings"

	"auth-service/pkg/response"

	"github.com/gin-gonic/gin"
)

// ViewerRole is the seeded read-only role.
const ViewerRole = "VIEWER"

// IsViewerOnly reports whether every role on the caller's token is VIEWER
// (case-insensitive). A caller who also holds another role keeps that role's
// rights; an empty role list is not a viewer. Authenticate must have run first.
func IsViewerOnly(c *gin.Context) bool {
	v, _ := c.Get("roles")
	roles, _ := v.([]string)
	if len(roles) == 0 {
		return false
	}
	for _, r := range roles {
		if !strings.EqualFold(strings.TrimSpace(r), ViewerRole) {
			return false
		}
	}
	return true
}

// ViewerRoute is one route a VIEWER-only caller may use. Path is the gin route
// pattern (FullPath). SelfParam, when set, also requires that path parameter to
// equal the caller's own user ID.
type ViewerRoute struct {
	Method    string
	Path      string
	SelfParam string
}

// ViewerAllowlist is deny-by-default for a VIEWER-only caller: any route not in
// allow answers 403, so an auth-service route added later is closed to VIEWER
// until it is listed here. Other callers pass through. Authenticate must have
// run first.
func ViewerAllowlist(allow ...ViewerRoute) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsViewerOnly(c) {
			c.Next()
			return
		}
		for _, a := range allow {
			if a.Method != c.Request.Method || a.Path != c.FullPath() {
				continue
			}
			if a.SelfParam != "" {
				if uid, _ := c.Get("user_id"); uid == nil || uid != c.Param(a.SelfParam) {
					continue
				}
			}
			c.Next()
			return
		}
		response.Forbidden(c, "Viewer accounts cannot use this endpoint")
		c.Abort()
	}
}
