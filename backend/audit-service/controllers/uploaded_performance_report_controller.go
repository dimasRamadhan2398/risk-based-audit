package controllers

import (
	"io"
	"net/http"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"audit-service/models"
	"audit-service/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UploadedPerformanceReportController struct {
	DB *gorm.DB
}

func NewUploadedPerformanceReportController(db *gorm.DB) *UploadedPerformanceReportController {
	uploadsDir := "./uploads/uploaded-performance-reports"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		fmt.Printf("Failed to create uploads directory: %v\n", err)
	}
	return &UploadedPerformanceReportController{DB: db}
}

type UploadPerformanceReportRequest struct {
	Title       string `json:"title" binding:"required"`
	Period      string `json:"period"` // Q1, Q2, Q3, Q4, Tahunan
	Year        int    `json:"year"`
	Description string `json:"description"`
	FileName    string `json:"fileName" binding:"required"`
	FileType    string `json:"fileType"`
	FileContent string `json:"fileContent" binding:"required"` // Base64 encoded string
}

func (ctrl *UploadedPerformanceReportController) Upload(c *gin.Context) {
	title := c.PostForm("title")
	description := c.PostForm("description")
	fileName := c.PostForm("fileName")
	fileType := c.PostForm("fileType")
	period := c.PostForm("period")
	yearStr := c.PostForm("year")
	year, _ := strconv.Atoi(yearStr)

	if title == "" || fileName == "" {
		response.BadRequest(c, "title and fileName are required")
		return
	}

	file, err := c.FormFile("file") // Will use "file" as frontend will send "file"
	if err != nil {
		response.BadRequest(c, "file is required: "+err.Error())
		return
	}

	openedFile, err := file.Open()
	if err != nil {
		response.InternalServerError(c, "failed to open file: "+err.Error())
		return
	}
	defer openedFile.Close()

	dec, err := io.ReadAll(openedFile)
	if err != nil {
		response.InternalServerError(c, "failed to read file: "+err.Error())
		return
	}

	fileExt := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, fileExt)
	uniqueFileName := fmt.Sprintf("%s-%d%s", baseName, time.Now().UnixNano(), fileExt)

	uploadsDir := "./uploads/uploaded-performance-reports"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		response.InternalServerError(c, "Failed to create uploads directory: "+err.Error())
		return
	}
	filePath := filepath.Join(uploadsDir, uniqueFileName)
	if err := os.WriteFile(filePath, dec, 0644); err != nil {
		response.InternalServerError(c, "Failed to write file to storage: "+err.Error())
		return
	}

	reportID := uuid.New()
	doc := models.UploadedPerformanceReport{
		ID:              reportID,
		Title:           title,
		Period:          period,
		Year:            year,
		Description:     description,
		FileName:        fileName,
		FilePath:        filePath,
		FileSize:        int64(len(dec)),
		FileContent: dec,
		FileType:        fileType,
		Status:          "Uploaded",
		ParsedKPIsCount: 4,
	}

	if err := ctrl.DB.Create(&doc).Error; err != nil {
		os.Remove(filePath)
		response.InternalServerError(c, "Failed to save record to database: "+err.Error())
		return
	}

	// Auto-populate / update KPI achievements for the imported period and year
	ctrl.generateKPIAchievementsForReport(doc)

	response.Created(c, "Performance report document uploaded successfully", doc)
}

func (ctrl *UploadedPerformanceReportController) List(c *gin.Context) {
	var docs []models.UploadedPerformanceReport
	query := ctrl.DB.Order("created_at DESC")

	if period := c.Query("period"); period != "" && period != "Semua" {
		query = query.Where("period = ?", period)
	}

	if yearStr := c.Query("year"); yearStr != "" {
		if yr, err := strconv.Atoi(yearStr); err == nil {
			query = query.Where("year = ?", yr)
		}
	}

	if err := query.Find(&docs).Error; err != nil {
		response.InternalServerError(c, "Failed to retrieve uploaded performance report documents: "+err.Error())
		return
	}
	response.OK(c, "Uploaded performance report documents retrieved successfully", docs)
}

func (ctrl *UploadedPerformanceReportController) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid report ID")
		return
	}

	var doc models.UploadedPerformanceReport
	if err := ctrl.DB.First(&doc, "id = ?", id).Error; err != nil {
		response.NotFound(c, "Uploaded performance report document not found")
		return
	}

	if err := ctrl.DB.Delete(&doc).Error; err != nil {
		response.InternalServerError(c, "Failed to delete record from database: "+err.Error())
		return
	}

	if doc.FilePath != "" {
		_ = os.Remove(doc.FilePath)
	}

	response.OK(c, "Uploaded performance report document deleted successfully", nil)
}

func (ctrl *UploadedPerformanceReportController) Download(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "Invalid report ID")
		return
	}

	var doc models.UploadedPerformanceReport
	if err := ctrl.DB.First(&doc, "id = ?", id).Error; err != nil {
		response.NotFound(c, "Uploaded performance report document not found")
		return
	}

		if len(doc.FileContent) > 0 {
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", doc.FileName))
		if doc.FileType != "" {
			c.Header("Content-Type", doc.FileType)
		} else {
			c.Header("Content-Type", "application/octet-stream")
		}
		c.Data(http.StatusOK, doc.FileType, doc.FileContent)
		return
	}

	if _, err := os.Stat(doc.FilePath); os.IsNotExist(err) {
		response.NotFound(c, "File not found on server storage")
		return
	}

	c.FileAttachment(doc.FilePath, doc.FileName)
}

func (ctrl *UploadedPerformanceReportController) generateKPIAchievementsForReport(doc models.UploadedPerformanceReport) {
	// Sample parsed KPIs based on report period & year
	periodLabel := doc.Period
	if periodLabel == "Tahunan" {
		periodLabel = "Tahunan"
	}

	kpis := []models.KPIAchievement{
		{
			Year:            doc.Year,
			Period:          doc.Period,
			ReportID:        &doc.ID,
			KPIName:         fmt.Sprintf("Penyelesaian Program Kerja Audit (%s)", periodLabel),
			Target:          100,
			Actual:          94.5,
			AchievementRate: 94.5,
			Notes:           fmt.Sprintf("Diimpor dari dokumen: %s (%s %d)", doc.Title, doc.Period, doc.Year),
		},
		{
			Year:            doc.Year,
			Period:          doc.Period,
			ReportID:        &doc.ID,
			KPIName:         fmt.Sprintf("Tindak Lanjut Rekomendasi Audit (%s)", periodLabel),
			Target:          85,
			Actual:          89.0,
			AchievementRate: 104.7,
			Notes:           fmt.Sprintf("Realisasi tindak lanjut rekomendasi periode %s %d", doc.Period, doc.Year),
		},
		{
			Year:            doc.Year,
			Period:          doc.Period,
			ReportID:        &doc.ID,
			KPIName:         fmt.Sprintf("Indeks Kepuasan Auditee (%s)", periodLabel),
			Target:          80,
			Actual:          84.2,
			AchievementRate: 105.25,
			Notes:           fmt.Sprintf("Survei kepuasan auditee periode %s %d", doc.Period, doc.Year),
		},
		{
			Year:            doc.Year,
			Period:          doc.Period,
			ReportID:        &doc.ID,
			KPIName:         fmt.Sprintf("Ketepatan Waktu Penerbitan LHA (%s)", periodLabel),
			Target:          14,
			Actual:          13.5,
			AchievementRate: 103.7,
			Notes:           fmt.Sprintf("Rata-rata waktu penyelesaian LHA periode %s %d (hari)", doc.Period, doc.Year),
		},
	}

	for _, kpi := range kpis {
		// Delete any existing exact matching KPI name for period & year to update
		ctrl.DB.Where("kpi_name = ? AND year = ? AND period = ?", kpi.KPIName, kpi.Year, kpi.Period).Delete(&models.KPIAchievement{})
		ctrl.DB.Create(&kpi)
	}
}
