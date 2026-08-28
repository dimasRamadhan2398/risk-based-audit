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

type ImportedWorkingPaperController struct {
	DB *gorm.DB
}

func NewImportedWorkingPaperController(db *gorm.DB) *ImportedWorkingPaperController {
	uploadsDir := "./uploads/imported-working-papers"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		fmt.Printf("Failed to create uploads directory: %v\n", err)
	}
	return &ImportedWorkingPaperController{DB: db}
}

type ImportRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	FileName    string `json:"fileName" binding:"required"`
	FileType    string `json:"fileType"`
	FileContent string `json:"fileContent" binding:"required"` // Base64 encoded string
}

func (ctrl *ImportedWorkingPaperController) Import(c *gin.Context) {
	var req ImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Clean base64 string if it contains metadata prefix
	base64Data := req.FileContent
	if idx := strings.Index(base64Data, ";base64,"); idx != -1 {
		base64Data = base64Data[idx+8:]
	}

	// Decode base64
	dec, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		response.BadRequest(c, "Invalid base64 file data: "+err.Error())
		return
	}

	if int64(len(dec)) > 10*1024*1024 {
		response.BadRequest(c, "File size exceeds maximum limit of 10MB")
		return
	}

	// Generate a unique filename to avoid overwrites
	fileExt := filepath.Ext(req.FileName)
	baseName := strings.TrimSuffix(req.FileName, fileExt)
	uniqueFileName := fmt.Sprintf("%s-%d%s", baseName, time.Now().UnixNano(), fileExt)

	// Save file to disk
	uploadsDir := "./uploads/imported-working-papers"
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
	paper := models.ImportedWorkingPaper{
		ID:          uuid.New(),
		Title:       req.Title,
		Description: req.Description,
		FileName:    req.FileName,
		FilePath:    filePath,
		FileSize:    int64(len(dec)),
		FileContent: dec,
		FileType:    req.FileType,
	}

	if err := ctrl.DB.Create(&paper).Error; err != nil {
		os.Remove(filePath)
		response.InternalServerError(c, "Failed to save record to database: "+err.Error())
		return
	}

	response.Created(c, "Working paper imported successfully", paper)
}

func (ctrl *ImportedWorkingPaperController) List(c *gin.Context) {
	var papers []models.ImportedWorkingPaper
	if err := ctrl.DB.Order("created_at DESC").Find(&papers).Error; err != nil {
		response.InternalServerError(c, "Failed to retrieve imported working papers: "+err.Error())
		return
	}
	response.OK(c, "Imported working papers retrieved successfully", papers)
}

func (ctrl *ImportedWorkingPaperController) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid working paper ID")
		return
	}

	var paper models.ImportedWorkingPaper
	if err := ctrl.DB.First(&paper, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "Imported working paper not found")
			return
		}
		response.InternalServerError(c, "Failed to fetch imported working paper")
		return
	}

	if err := ctrl.DB.Delete(&paper).Error; err != nil {
		response.InternalServerError(c, "Failed to delete imported working paper record")
		return
	}

	if err := os.Remove(paper.FilePath); err != nil {
		fmt.Printf("Warning: Failed to delete physical file %s: %v\n", paper.FilePath, err)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Imported working paper deleted successfully",
	})
}

func (ctrl *ImportedWorkingPaperController) Download(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid working paper ID")
		return
	}

	var paper models.ImportedWorkingPaper
	if err := ctrl.DB.First(&paper, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "Imported working paper not found")
			return
		}
		response.InternalServerError(c, "Failed to fetch imported working paper")
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
