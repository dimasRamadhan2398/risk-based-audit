package routes

import (
	"net/http"

	"master-service/controllers"
	"master-service/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RegisterRoutes wires all HTTP routes for the master-service.
func RegisterRoutes(router *gin.Engine, controller controllers.IControllerRegistry, db *gorm.DB) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"service": "master-service", "status": "ok"})
	})

	api := router.Group("/api/v1")

	// Organizational structure
	companies := api.Group("/companies")
	{
		companies.GET("", controller.GetCompany().FindAll)
		companies.GET("/:id", controller.GetCompany().FindById)
		companies.POST("", controller.GetCompany().Create)
		companies.PUT("/:id", controller.GetCompany().Update)
		companies.DELETE("/:id", controller.GetCompany().Delete)
	}

	businessUnits := api.Group("/business-units")
	{
		businessUnits.GET("", controller.GetBusinessUnit().FindAll)
		businessUnits.GET("/:id", controller.GetBusinessUnit().FindById)
		businessUnits.POST("", controller.GetBusinessUnit().Create)
		businessUnits.PUT("/:id", controller.GetBusinessUnit().Update)
		businessUnits.DELETE("/:id", controller.GetBusinessUnit().Delete)
	}

	departments := api.Group("/departments")
	{
		departments.GET("", controller.GetDepartment().List)
		departments.GET("/:id", controller.GetDepartment().FindById)
		departments.POST("", controller.GetDepartment().Create)
		departments.PUT("/:id", controller.GetDepartment().Update)
		departments.DELETE("/:id", controller.GetDepartment().Delete)
	}

	qa := api.Group("/quality-assurance")
	{
		qa.GET("", func(c *gin.Context) {
			var reports []models.QAReport
			if err := db.Order("created_at DESC").Find(&reports).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to fetch QA reports"})
				return
			}
			c.JSON(http.StatusOK, reports)
		})

		qa.GET("/:id", func(c *gin.Context) {
			id, err := uuid.Parse(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID format"})
				return
			}
			var report models.QAReport
			if err := db.First(&report, "id = ?", id).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "QA report not found"})
				return
			}
			c.JSON(http.StatusOK, report)
		})

		qa.POST("", func(c *gin.Context) {
			var report models.QAReport
			if err := c.ShouldBindJSON(&report); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
				return
			}
			if report.ID == uuid.Nil {
				report.ID = uuid.New()
			}
			if err := db.Create(&report).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to create QA report"})
				return
			}
			c.JSON(http.StatusCreated, report)
		})

		qa.PUT("/:id", func(c *gin.Context) {
			id, err := uuid.Parse(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID format"})
				return
			}
			var existing models.QAReport
			if err := db.First(&existing, "id = ?", id).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "QA report not found"})
				return
			}

			var req models.QAReport
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
				return
			}

			existing.Type = req.Type
			existing.Period = req.Period
			existing.ReportName = req.ReportName
			existing.Result = req.Result
			existing.Status = req.Status
			existing.AssessmentTitle = req.AssessmentTitle
			existing.Validator = req.Validator
			existing.InternalEvaluator = req.InternalEvaluator
			existing.Attachment = req.Attachment

			if err := db.Save(&existing).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update QA report"})
				return
			}
			c.JSON(http.StatusOK, existing)
		})

		qa.DELETE("/:id", func(c *gin.Context) {
			id, err := uuid.Parse(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID format"})
				return
			}
			if err := db.Delete(&models.QAReport{}, "id = ?", id).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to delete QA report"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "QA report deleted"})
		})
	}
}
