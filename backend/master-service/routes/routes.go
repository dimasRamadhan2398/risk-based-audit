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

	// Company routes
	companies := api.Group("/companies")
	{
		companies.GET("", controller.GetCompany().List)
		companies.GET("/:id", controller.GetCompany().FindById)
		companies.POST("", controller.GetCompany().Create)
		companies.PUT("/:id", controller.GetCompany().Update)
		companies.DELETE("/:id", controller.GetCompany().Delete)
	}

	// Department routes
	departments := api.Group("/departments")
	{
		departments.GET("", controller.GetDepartment().List)
		departments.GET("/:id", controller.GetDepartment().FindById)
		departments.POST("", controller.GetDepartment().Create)
		departments.PUT("/:id", controller.GetDepartment().Update)
		departments.DELETE("/:id", controller.GetDepartment().Delete)
	}

	// Employee routes
	employees := api.Group("/employees")
	{
		employees.GET("", controller.GetEmployee().List)
		employees.GET("/:id", controller.GetEmployee().FindById)
		employees.POST("", controller.GetEmployee().Create)
		employees.PUT("/:id", controller.GetEmployee().Update)
		employees.DELETE("/:id", controller.GetEmployee().Delete)
	}
}
