package routes

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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
			existing.ConductedBy = req.ConductedBy
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

		qa.POST("/import", func(c *gin.Context) {
			type ImportQARRequest struct {
				AssessmentTitle   string `json:"assessmentTitle" binding:"required"`
				Type              string `json:"type" binding:"required"`
				PeriodQuarter     string `json:"periodQuarter"`
				PeriodYear        string `json:"periodYear" binding:"required"`
				Result            string `json:"result" binding:"required"`
				Status            string `json:"status" binding:"required"`
				ConductedBy       string `json:"conductedBy"`
				Validator         string `json:"validator"`
				InternalEvaluator string `json:"internalEvaluator"`
				FileName          string `json:"fileName" binding:"required"`
				FileType          string `json:"fileType"`
				FileContent       string `json:"fileContent" binding:"required"` // Base64
			}

			var req ImportQARRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
				return
			}

			// Clean and decode base64
			base64Data := req.FileContent
			if idx := strings.Index(base64Data, ";base64,"); idx != -1 {
				base64Data = base64Data[idx+8:]
			}
			dec, err := base64.StdEncoding.DecodeString(base64Data)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid base64 file data: " + err.Error()})
				return
			}

			// Save to uploads/qar-reports
			uploadsDir := "./uploads/qar-reports"
			if err := os.MkdirAll(uploadsDir, 0755); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to create uploads directory: " + err.Error()})
				return
			}
			fileExt := filepath.Ext(req.FileName)
			baseName := strings.TrimSuffix(req.FileName, fileExt)
			uniqueFileName := fmt.Sprintf("%s-%d%s", baseName, time.Now().UnixNano(), fileExt)
			filePath := filepath.Join(uploadsDir, uniqueFileName)
			if err := os.WriteFile(filePath, dec, 0644); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to write file to storage: " + err.Error()})
				return
			}

			// Construct Period
			period := fmt.Sprintf("%s %s", req.PeriodQuarter, req.PeriodYear)
			period = strings.TrimSpace(period)
			if period == "" {
				period = req.PeriodYear
			}

			// Create QAReport
			report := models.QAReport{
				ID:                uuid.New(),
				Type:              req.Type,
				IsImported:        true,
				Period:            period,
				ReportName:        req.AssessmentTitle,
				Result:            req.Result,
				Status:            req.Status,
				ConductedBy:       req.ConductedBy,
				AssessmentTitle:   req.AssessmentTitle,
				Validator:         req.Validator,
				InternalEvaluator: req.InternalEvaluator,
				Attachment: &models.QAReportAttachment{
					Name:       req.FileName,
					Size:       fmt.Sprintf("%.2f MB", float64(len(dec))/(1024*1024)),
					UploadedAt: time.Now().Format("2006-01-02"),
					FilePath:   filePath,
				},
			}

			if err := db.Create(&report).Error; err != nil {
				os.Remove(filePath)
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to save QA report record: " + err.Error()})
				return
			}

			c.JSON(http.StatusCreated, report)
		})

		qa.GET("/:id/download", func(c *gin.Context) {
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

			if report.Attachment == nil || report.Attachment.FilePath == "" {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "This report does not have an attached file"})
				return
			}

			if _, err := os.Stat(report.Attachment.FilePath); os.IsNotExist(err) {
				c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Physical file not found on server"})
				return
			}

			c.Header("Content-Description", "File Transfer")
			c.Header("Content-Transfer-Encoding", "binary")
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", report.Attachment.Name))
			c.Header("Content-Type", "application/octet-stream")
			c.File(report.Attachment.FilePath)
		})

		qa.DELETE("/:id", func(c *gin.Context) {
			id, err := uuid.Parse(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID format"})
				return
			}

			var report models.QAReport
			if err := db.First(&report, "id = ?", id).Error; err == nil {
				if report.Attachment != nil && report.Attachment.FilePath != "" {
					os.Remove(report.Attachment.FilePath)
				}
			}

			if err := db.Delete(&models.QAReport{}, "id = ?", id).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to delete QA report"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "QA report deleted"})
		})
	}

	consulting := api.Group("/consulting-services")
	{
		consulting.GET("", func(c *gin.Context) {
			var services []models.ConsultingService
			if err := db.Order("created_at DESC").Find(&services).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to fetch consulting services"})
				return
			}
			c.JSON(http.StatusOK, services)
		})

		consulting.GET("/:id", func(c *gin.Context) {
			id, err := uuid.Parse(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID format"})
				return
			}
			var service models.ConsultingService
			if err := db.First(&service, "id = ?", id).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Consulting service not found"})
				return
			}
			c.JSON(http.StatusOK, service)
		})

		consulting.POST("", func(c *gin.Context) {
			var service models.ConsultingService
			if err := c.ShouldBindJSON(&service); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
				return
			}
			if service.ID == uuid.Nil {
				service.ID = uuid.New()
			}
			if err := db.Create(&service).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to create consulting service"})
				return
			}
			c.JSON(http.StatusCreated, service)
		})

		consulting.PUT("/:id", func(c *gin.Context) {
			id, err := uuid.Parse(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID format"})
				return
			}
			var existing models.ConsultingService
			if err := db.First(&existing, "id = ?", id).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Consulting service not found"})
				return
			}

			var req models.ConsultingService
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
				return
			}

			existing.Title = req.Title
			existing.Category = req.Category
			existing.RequestorDept = req.RequestorDept
			existing.Period = req.Period
			existing.ConsultantName = req.ConsultantName
			existing.Status = req.Status
			existing.Notes = req.Notes
			existing.Attachment = req.Attachment

			if err := db.Save(&existing).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update consulting service"})
				return
			}
			c.JSON(http.StatusOK, existing)
		})

		consulting.DELETE("/:id", func(c *gin.Context) {
			id, err := uuid.Parse(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID format"})
				return
			}

			var service models.ConsultingService
			if err := db.First(&service, "id = ?", id).Error; err == nil {
				if service.Attachment != nil && service.Attachment.FilePath != "" {
					os.Remove(service.Attachment.FilePath)
				}
			}

			if err := db.Delete(&models.ConsultingService{}, "id = ?", id).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to delete consulting service"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "message": "Consulting service deleted"})
		})

		consulting.POST("/import", func(c *gin.Context) {
			type ImportConsultingRequest struct {
				Title          string `json:"title" binding:"required"`
				Category       string `json:"category" binding:"required"`
				RequestorDept  string `json:"requestorDept" binding:"required"`
				Period         string `json:"period" binding:"required"`
				ConsultantName string `json:"consultantName" binding:"required"`
				Status         string `json:"status" binding:"required"`
				Notes          string `json:"notes"`
				FileName       string `json:"fileName" binding:"required"`
				FileType       string `json:"fileType"`
				FileContent    string `json:"fileContent" binding:"required"` // Base64
			}

			var req ImportConsultingRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
				return
			}

			base64Data := req.FileContent
			if idx := strings.Index(base64Data, ";base64,"); idx != -1 {
				base64Data = base64Data[idx+8:]
			}
			dec, err := base64.StdEncoding.DecodeString(base64Data)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid base64 file data: " + err.Error()})
				return
			}

			uploadsDir := "./uploads/consulting"
			if err := os.MkdirAll(uploadsDir, 0755); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to create uploads directory: " + err.Error()})
				return
			}
			fileExt := filepath.Ext(req.FileName)
			baseName := strings.TrimSuffix(req.FileName, fileExt)
			uniqueFileName := fmt.Sprintf("%s-%d%s", baseName, time.Now().UnixNano(), fileExt)
			filePath := filepath.Join(uploadsDir, uniqueFileName)
			if err := os.WriteFile(filePath, dec, 0644); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to write file to storage: " + err.Error()})
				return
			}

			service := models.ConsultingService{
				ID:             uuid.New(),
				Title:          req.Title,
				Category:       req.Category,
				RequestorDept:  req.RequestorDept,
				Period:         req.Period,
				ConsultantName: req.ConsultantName,
				Status:         req.Status,
				Notes:          req.Notes,
				Attachment: &models.ConsultingAttachment{
					Name:       req.FileName,
					Size:       fmt.Sprintf("%.2f MB", float64(len(dec))/(1024*1024)),
					UploadedAt: time.Now().Format("2006-01-02"),
					FilePath:   filePath,
				},
			}

			if err := db.Create(&service).Error; err != nil {
				os.Remove(filePath)
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to save consulting service record: " + err.Error()})
				return
			}

			c.JSON(http.StatusCreated, service)
		})

		consulting.GET("/:id/download", func(c *gin.Context) {
			id, err := uuid.Parse(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ID format"})
				return
			}
			var service models.ConsultingService
			if err := db.First(&service, "id = ?", id).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Consulting service not found"})
				return
			}

			if service.Attachment == nil || service.Attachment.FilePath == "" {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "This record does not have an attached file"})
				return
			}

			if _, err := os.Stat(service.Attachment.FilePath); os.IsNotExist(err) {
				c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Physical file not found on server"})
				return
			}

			c.Header("Content-Description", "File Transfer")
			c.Header("Content-Transfer-Encoding", "binary")
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", service.Attachment.Name))
			c.Header("Content-Type", "application/octet-stream")
			c.File(service.Attachment.FilePath)
		})
	}

	// Vision, Mission & Goals routes
	vmgs := api.Group("/vision-mission-goals")
	{
		vmgs.GET("", controller.GetVisionMissionGoals().List)
		vmgs.GET("/:id", controller.GetVisionMissionGoals().FindById)
		vmgs.GET("/company/:companyId", controller.GetVisionMissionGoals().FindByCompany)
		vmgs.POST("", controller.GetVisionMissionGoals().Create)
		vmgs.PUT("/:id", controller.GetVisionMissionGoals().Update)
		vmgs.DELETE("/:id", controller.GetVisionMissionGoals().Delete)
	}
}
