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

type UploadedPlanDocumentController struct {
	DB *gorm.DB
}

func NewUploadedPlanDocumentController(db *gorm.DB) *UploadedPlanDocumentController {
	uploadsDir := "./uploads/uploaded-plan-documents"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		fmt.Printf("Failed to create uploads directory: %v\n", err)
	}
	return &UploadedPlanDocumentController{DB: db}
}

type UploadPlanRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	FileName    string `json:"fileName" binding:"required"`
	FileType    string `json:"fileType"`
	FileContent string `json:"fileContent" binding:"required"` // Base64 encoded string
}

func (ctrl *UploadedPlanDocumentController) Upload(c *gin.Context) {
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

	if int64(len(dec)) > 10*1024*1024 {
		response.BadRequest(c, "File size exceeds maximum limit of 10MB")
		return
	}

	// Generate a unique filename to avoid overwrites
	fileExt := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, fileExt)
	uniqueFileName := fmt.Sprintf("%s-%d%s", baseName, time.Now().UnixNano(), fileExt)

	// Save file to disk
	uploadsDir := "./uploads/uploaded-plan-documents"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		response.InternalServerError(c, "Failed to create uploads directory: "+err.Error())
		return
	}
	filePath := filepath.Join(uploadsDir, uniqueFileName)
	if err := os.WriteFile(filePath, dec, 0644); err != nil {
		response.InternalServerError(c, "Failed to write file to storage: "+err.Error())
		return
	}

	// Save metadata in database
	paper := models.UploadedPlanDocument{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
		FileName:    fileName,
		FilePath:    filePath,
		FileSize:    int64(len(dec)),
		FileContent: dec,
		FileType:    fileType,
	}

	if err := ctrl.DB.Create(&paper).Error; err != nil {
		os.Remove(filePath)
		response.InternalServerError(c, "Failed to save record to database: "+err.Error())
		return
	}

	response.Created(c, "Plan document uploaded successfully", paper)
}

func (ctrl *UploadedPlanDocumentController) List(c *gin.Context) {
	var papers []models.UploadedPlanDocument
	if err := ctrl.DB.Order("created_at DESC").Find(&papers).Error; err != nil {
		response.InternalServerError(c, "Failed to retrieve uploaded plan documents: "+err.Error())
		return
	}
	response.OK(c, "Uploaded plan documents retrieved successfully", papers)
}

func (ctrl *UploadedPlanDocumentController) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid plan document ID")
		return
	}

	var paper models.UploadedPlanDocument
	if err := ctrl.DB.First(&paper, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "Uploaded plan document not found")
			return
		}
		response.InternalServerError(c, "Failed to fetch uploaded plan document")
		return
	}

	if err := ctrl.DB.Delete(&paper).Error; err != nil {
		response.InternalServerError(c, "Failed to delete uploaded plan document record")
		return
	}

	if err := os.Remove(paper.FilePath); err != nil {
		fmt.Printf("Warning: Failed to delete physical file %s: %v\n", paper.FilePath, err)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Uploaded plan document deleted successfully",
	})
}

func (ctrl *UploadedPlanDocumentController) Download(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid plan document ID")
		return
	}

	var paper models.UploadedPlanDocument
	if err := ctrl.DB.First(&paper, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "Uploaded plan document not found")
			return
		}
		response.InternalServerError(c, "Failed to fetch uploaded plan document")
		return
	}

		if len(paper.FileContent) > 0 {
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", paper.FileName))
		if paper.FileType != "" {
			c.Header("Content-Type", paper.FileType)
		} else {
			c.Header("Content-Type", "application/octet-stream")
		}
		c.Data(http.StatusOK, paper.FileType, paper.FileContent)
		return
	}

	if _, err := os.Stat(paper.FilePath); os.IsNotExist(err) {
		response.NotFound(c, "Physical file not found on server")
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", paper.FileName))
	c.Header("Content-Type", paper.FileType)
	c.File(paper.FilePath)
}
