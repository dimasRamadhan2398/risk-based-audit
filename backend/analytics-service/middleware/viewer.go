package middleware

import (
	"net/http"
	"strings"

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

// DenyViewerWrites refuses every method other than GET/HEAD/OPTIONS for a
// VIEWER-only caller. It is deny-by-default: a route added later is covered
// without being listed. allow holds "METHOD /full/route/path" entries (gin
// FullPath) for POSTs that only read.
func DenyViewerWrites(allow ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allow))
	for _, a := range allow {
		allowed[a] = struct{}{}
	}
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}
		if !IsViewerOnly(c) {
			c.Next()
			return
		}
		if _, ok := allowed[c.Request.Method+" "+c.FullPath()]; ok {
			c.Next()
			return
		}
		forbidden(c, "Viewer accounts are read-only")
	}
}
