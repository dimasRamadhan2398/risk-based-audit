package middleware

import (
	"strings"

	"master-service/pkg/response"

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

// DenyViewer refuses a VIEWER-only caller on every method. Policy: master data
// (companies, departments, employees, QA, consulting, data sources ...) is not
// visible to VIEWER at all, so it is mounted on the whole /api/v1 group.
func DenyViewer() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsViewerOnly(c) {
			response.Forbidden(c, "Viewer accounts cannot access master data")
			c.Abort()
			return
		}
		c.Next()
	}
}
