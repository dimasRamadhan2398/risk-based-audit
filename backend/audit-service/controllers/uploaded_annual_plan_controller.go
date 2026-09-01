package controllers

import (
	"fmt"
	"io"
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

type UploadedAnnualPlanController struct {
	DB *gorm.DB
}

func NewUploadedAnnualPlanController(db *gorm.DB) *UploadedAnnualPlanController {
	uploadsDir := "./uploads/uploaded-annual-plans"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		fmt.Printf("Failed to create uploads directory: %v\n", err)
	}
	return &UploadedAnnualPlanController{DB: db}
}

type UploadAnnualPlanRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	FileName    string `json:"fileName" binding:"required"`
	FileType    string `json:"fileType"`
	FileContent string `json:"fileContent" binding:"required"` // Base64 encoded string
}

func (ctrl *UploadedAnnualPlanController) Upload(c *gin.Context) {
	title := c.PostForm("title")
	description := c.PostForm("description")
	fileName := c.PostForm("fileName")
	fileType := c.PostForm("fileType")

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

	uploadsDir := "./uploads/uploaded-annual-plans"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		response.InternalServerError(c, "Failed to create uploads directory: "+err.Error())
		return
	}
	filePath := filepath.Join(uploadsDir, uniqueFileName)
	if err := os.WriteFile(filePath, dec, 0644); err != nil {
		response.InternalServerError(c, "Failed to write file to storage: "+err.Error())
		return
	}

	doc := models.UploadedAnnualPlan{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
		FileName:    fileName,
		FilePath:    filePath,
		FileSize:    int64(len(dec)),
		FileContent: dec,
		FileType:    fileType,
	}

	if err := ctrl.DB.Create(&doc).Error; err != nil {
		os.Remove(filePath)
		response.InternalServerError(c, "Failed to save record to database: "+err.Error())
		return
	}

	response.Created(c, "Annual audit plan document uploaded successfully", doc)
}

func (ctrl *UploadedAnnualPlanController) List(c *gin.Context) {
	var docs []models.UploadedAnnualPlan
	if err := ctrl.DB.Order("created_at DESC").Find(&docs).Error; err != nil {
		response.InternalServerError(c, "Failed to retrieve uploaded annual audit plan documents: "+err.Error())
		return
	}
	response.OK(c, "Uploaded annual audit plan documents retrieved successfully", docs)
}

func (ctrl *UploadedAnnualPlanController) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid document ID")
		return
	}

	var doc models.UploadedAnnualPlan
	if err := ctrl.DB.First(&doc, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "Uploaded annual audit plan document not found")
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
		"message": "Uploaded annual audit plan document deleted successfully",
	})
}

func (ctrl *UploadedAnnualPlanController) Download(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid document ID")
		return
	}

	var doc models.UploadedAnnualPlan
	if err := ctrl.DB.First(&doc, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "Uploaded annual audit plan document not found")
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
