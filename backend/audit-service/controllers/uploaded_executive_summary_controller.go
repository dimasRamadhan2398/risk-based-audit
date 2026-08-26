package controllers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"audit-service/models"
	"audit-service/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UploadedExecutiveSummaryController struct {
	DB *gorm.DB
}

func NewUploadedExecutiveSummaryController(db *gorm.DB) *UploadedExecutiveSummaryController {
	uploadsDir := "./uploads/uploaded-executive-summaries"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		fmt.Printf("Failed to create uploads directory: %v\n", err)
	}
	return &UploadedExecutiveSummaryController{DB: db}
}

type UploadExecutiveSummaryRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	FileName    string `json:"fileName" binding:"required"`
	FileType    string `json:"fileType"`
	FileContent string `json:"fileContent" binding:"required"`
}

func (ctrl *UploadedExecutiveSummaryController) Upload(c *gin.Context) {
	var req UploadExecutiveSummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	base64Data := req.FileContent
	if idx := strings.Index(base64Data, ";base64,"); idx != -1 {
		base64Data = base64Data[idx+8:]
	}

	dec, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		response.BadRequest(c, "Invalid base64 file data: "+err.Error())
		return
	}

	fileExt := filepath.Ext(req.FileName)
	baseName := strings.TrimSuffix(req.FileName, fileExt)
	uniqueFileName := fmt.Sprintf("%s-%d%s", baseName, time.Now().UnixNano(), fileExt)

	uploadsDir := "./uploads/uploaded-executive-summaries"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		response.InternalServerError(c, "Failed to create uploads directory: "+err.Error())
		return
	}
	filePath := filepath.Join(uploadsDir, uniqueFileName)
	if err := os.WriteFile(filePath, dec, 0644); err != nil {
		response.InternalServerError(c, "Failed to write file to storage: "+err.Error())
		return
	}

	doc := models.UploadedExecutiveSummary{
		ID:          uuid.New(),
		Title:       req.Title,
		Description: req.Description,
		FileName:    req.FileName,
		FilePath:    filePath,
		FileSize:    int64(len(dec)),
		FileContent: dec,
		FileType:    req.FileType,
	}

	if err := ctrl.DB.Create(&doc).Error; err != nil {
		os.Remove(filePath)
		response.InternalServerError(c, "Failed to save record to database: "+err.Error())
		return
	}

	response.Created(c, "Executive summary document uploaded successfully", doc)
}

func (ctrl *UploadedExecutiveSummaryController) List(c *gin.Context) {
	var docs []models.UploadedExecutiveSummary
	if err := ctrl.DB.Order("created_at DESC").Find(&docs).Error; err != nil {
		response.InternalServerError(c, "Failed to retrieve uploaded executive summary documents: "+err.Error())
		return
	}
	response.OK(c, "Uploaded executive summary documents retrieved successfully", docs)
}

func (ctrl *UploadedExecutiveSummaryController) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid document ID")
		return
	}

	var doc models.UploadedExecutiveSummary
	if err := ctrl.DB.First(&doc, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "Uploaded executive summary document not found")
			return
		}
		response.InternalServerError(c, "Failed to fetch document")
		return
	}

	if err := ctrl.DB.Delete(&doc).Error; err != nil {
		response.InternalServerError(c, "Failed to delete document record")
		return
	}

	if err := os.Remove(doc.FilePath); err != nil {
		fmt.Printf("Warning: Failed to delete physical file %s: %v\n", doc.FilePath, err)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Uploaded executive summary document deleted successfully",
	})
}

func (ctrl *UploadedExecutiveSummaryController) Download(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid document ID")
		return
	}

	var doc models.UploadedExecutiveSummary
	if err := ctrl.DB.First(&doc, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "Uploaded executive summary document not found")
			return
		}
		response.InternalServerError(c, "Failed to fetch document")
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
		response.NotFound(c, "Physical file not found on server")
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", doc.FileName))
	c.Header("Content-Type", doc.FileType)
	c.File(doc.FilePath)
}
