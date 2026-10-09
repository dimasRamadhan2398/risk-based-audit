package routes

import (
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

// PrivateUploadDirs are subdirectories of ./uploads that must never be served
// by the public static /uploads route. Files there are only reachable through
// authenticated endpoints (e.g. ATR evidence via
// GET /api/v1/action-taken-reports/:id/evidence/:evidenceId).
var PrivateUploadDirs = []string{"action-taken-reports"}

// RegisterUploads serves root at /uploads like engine.Static did, except for
// PrivateUploadDirs, which answer 404. Directory listings are disabled.
func RegisterUploads(engine *gin.Engine, root string) {
	fs := gin.Dir(root, false)
	fileServer := http.StripPrefix("/uploads", http.FileServer(fs))
	handler := func(c *gin.Context) {
		if IsPrivateUploadPath(c.Param("filepath")) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
	engine.GET("/uploads/*filepath", handler)
	engine.HEAD("/uploads/*filepath", handler)
}

// IsPrivateUploadPath reports whether a /uploads sub-path falls in one of
// PrivateUploadDirs (case-insensitive, after cleaning "..", "//" and the like)
func IsPrivateUploadPath(p string) bool {
	cleaned := strings.ToLower(path.Clean("/" + strings.ReplaceAll(p, `\`, "/")))
	for _, dir := range PrivateUploadDirs {
		d := "/" + strings.ToLower(dir)
		if cleaned == d || strings.HasPrefix(cleaned, d+"/") {
			return true
		}
	}
	return false
}
