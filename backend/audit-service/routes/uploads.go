package routes

import (
	"net/http"
	"path/filepath"

	"audit-service/pkg/safefile"

	"github.com/gin-gonic/gin"
)

// RegisterUploads serves files under root at /uploads. Every request must pass
// authenticate, so the files are no longer public. Private subdirectories
// answer 404, directory listings are not served, and files go through
// safefile: confined paths, an allowlisted Content-Type, and an attachment for
// anything that is not safe to show inline.
func RegisterUploads(engine *gin.Engine, root string, authenticate gin.HandlerFunc) {
	handler := func(c *gin.Context) {
		rel := c.Param("filepath")
		if safefile.IsPrivatePath(rel) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		full, err := safefile.ResolveUnder(root, rel)
		if err != nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		download := c.Query("download") == "true"
		if err := safefile.ServeFile(c, full, filepath.Base(full), download); err != nil {
			c.AbortWithStatus(http.StatusNotFound)
		}
	}
	engine.GET("/uploads/*filepath", authenticate, handler)
	engine.HEAD("/uploads/*filepath", authenticate, handler)
}
