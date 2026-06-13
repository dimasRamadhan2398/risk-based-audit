package routes

import (
	"net/http"

	"master-service/controllers"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires all HTTP routes for the master-service.
func RegisterRoutes(router *gin.Engine, controller controllers.IControllerRegistry) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "master-service", "status": "ok"})
	})

	api := router.Group("/api/v1")
	departments := api.Group("/departments")
	{
		departments.GET("", controller.GetDepartment().FindAll)
		departments.GET("/:id", controller.GetDepartment().FindById)
		departments.POST("", controller.GetDepartment().Create)
		departments.PUT("/:id", controller.GetDepartment().Update)
		departments.DELETE("/:id", controller.GetDepartment().Delete)
	}
}
