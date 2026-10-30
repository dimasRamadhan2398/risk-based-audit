package middleware

import (
	"net/http"
	"strings"

	"audit-service/pkg/response"

	"github.com/gin-gonic/gin"
)

// ViewerRole is the seeded read-only role. Policy: a VIEWER may look at
// business data but not change it, not touch files and not export anything.
const ViewerRole = "VIEWER"

// IsViewerOnly reports whether every role on the caller's token is VIEWER
// (case-insensitive). A caller who also holds another role keeps that role's
// rights; an empty role list is not a viewer (RequireRoles already refuses it
// where it matters). Authenticate must have run first.
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

func isSafeMethod(m string) bool {
	return m == http.MethodGet || m == http.MethodHead || m == http.MethodOptions
}

func viewerForbidden(c *gin.Context, message string) {
	response.Forbidden(c, message)
	c.Abort()
}

// DenyViewerWrites refuses every method other than GET/HEAD/OPTIONS for a
// VIEWER-only caller. It is deny-by-default: a route added later is covered
// without being listed. allow holds "METHOD /full/route/path" entries (gin
// FullPath, e.g. "POST /api/v1/x/search") for POSTs that only read.
func DenyViewerWrites(allow ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allow))
	for _, a := range allow {
		allowed[a] = struct{}{}
	}
	return func(c *gin.Context) {
		if isSafeMethod(c.Request.Method) || !IsViewerOnly(c) {
			c.Next()
			return
		}
		if _, ok := allowed[c.Request.Method+" "+c.FullPath()]; ok {
			c.Next()
			return
		}
		viewerForbidden(c, "Viewer accounts are read-only")
	}
}

// DenyViewer refuses a VIEWER-only caller on every method.
func DenyViewer() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsViewerOnly(c) {
			viewerForbidden(c, "Viewer accounts cannot access this resource")
			return
		}
		c.Next()
	}
}

// IsFileOrExportRoute reports whether a gin route pattern serves a file or an
// exported artifact: /uploads, /api/v1/media, ATR evidence, and any path
// segment that is "view"/"evidence"/"print" or contains "download"/"export".
func IsFileOrExportRoute(fullPath string) bool {
	p := strings.ToLower(fullPath)
	if p == "/uploads" || strings.HasPrefix(p, "/uploads/") || strings.HasPrefix(p, "/api/v1/media") {
		return true
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == "view" || seg == "evidence" || seg == "print" ||
			strings.Contains(seg, "download") || strings.Contains(seg, "export") {
			return true
		}
	}
	return false
}

// DenyViewerFiles refuses a VIEWER-only caller on file, download, inline view,
// evidence, media and export routes (matched on the route pattern, so it must
// run as route/group middleware, not as engine.Use).
func DenyViewerFiles() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsViewerOnly(c) && IsFileOrExportRoute(c.FullPath()) {
			viewerForbidden(c, "Viewer accounts cannot access files or exports")
			return
		}
		c.Next()
	}
}
