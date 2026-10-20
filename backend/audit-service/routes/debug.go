package routes

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// envFlagEnabled reports whether an environment flag is set to "true". Anything
// else, including unset, means off.
func envFlagEnabled(name string) bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv(name)), "true")
}

// DebugRoutesEnv gates GET /debug/routes, which lists every registered route.
const DebugRoutesEnv = "ENABLE_DEBUG_ROUTES"

// SwaggerEnv gates the Swagger UI under /swagger.
const SwaggerEnv = "ENABLE_SWAGGER"

// registerDebugRoutes exposes the route table only when ENABLE_DEBUG_ROUTES=true.
// It is off by default and must stay off in production.
func registerDebugRoutes(engine *gin.Engine) {
	if !envFlagEnabled(DebugRoutesEnv) {
		return
	}
	engine.GET("/debug/routes", func(c *gin.Context) {
		routes := []gin.H{}
		for _, route := range engine.Routes() {
			routes = append(routes, gin.H{
				"method": route.Method,
				"path":   route.Path,
			})
		}
		c.JSON(http.StatusOK, gin.H{
			"routes": routes,
		})
	})
}

// SwaggerEnabled reports whether the Swagger UI should be mounted (ENABLE_SWAGGER=true).
func SwaggerEnabled() bool {
	return envFlagEnabled(SwaggerEnv)
}
