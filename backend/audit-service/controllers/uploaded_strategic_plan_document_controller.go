package controllers

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

type UploadedStrategicPlanDocumentController struct {
	DB *gorm.DB
}

func NewUploadedStrategicPlanDocumentController(db *gorm.DB) *UploadedStrategicPlanDocumentController {
	uploadsDir := "./uploads/uploaded-strategic-plan-documents"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		fmt.Printf("Failed to create uploads directory: %v\n", err)
	}
	return &UploadedStrategicPlanDocumentController{DB: db}
}

type UploadStrategicPlanRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	FileName    string `json:"fileName" binding:"required"`
	FileType    string `json:"fileType"`
	FileContent string `json:"fileContent" binding:"required"`
}

func (ctrl *UploadedStrategicPlanDocumentController) Upload(c *gin.Context) {
	title := c.PostForm("title")
	description := c.PostForm("description")
	fileName := c.PostForm("fileName")
	fileType := c.PostForm("fileType")

	var dec []byte

	file, err := c.FormFile("file")
	if err == nil {
		openedFile, err := file.Open()
		if err != nil {
			response.InternalServerError(c, "failed to open file: "+err.Error())
			return
		}
		defer openedFile.Close()

		dec, err = io.ReadAll(openedFile)
		if err != nil {
			response.InternalServerError(c, "failed to read file: "+err.Error())
			return
		}
	} else {
		var req UploadStrategicPlanRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "invalid request: "+err.Error())
			return
		}

		title = req.Title
		description = req.Description
		fileName = req.FileName
		fileType = req.FileType

		dec, err = base64.StdEncoding.DecodeString(req.FileContent)
		if err != nil {
			response.BadRequest(c, "invalid base64 fileContent: "+err.Error())
			return
		}
	}

	if title == "" || fileName == "" {
		response.BadRequest(c, "title and fileName are required")
		return
	}

	if int64(len(dec)) > 10*1024*1024 {
		response.BadRequest(c, "File size exceeds maximum limit of 10MB")
		return
	}

	uploadsDir := "./uploads/uploaded-strategic-plan-documents"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		response.InternalServerError(c, "Failed to create uploads directory: "+err.Error())
		return
	}

	fileExt := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, fileExt)
	uniqueFileName := fmt.Sprintf("%s-%d%s", baseName, time.Now().UnixNano(), fileExt)
	filePath := filepath.Join(uploadsDir, uniqueFileName)

	if err := os.WriteFile(filePath, dec, 0644); err != nil {
		response.InternalServerError(c, "Failed to write file to storage: "+err.Error())
		return
	}

	doc := models.UploadedStrategicPlanDocument{
		Title:       title,
		Description: description,
		FileName:    fileName,
		FilePath:    filePath,
		FileSize:    int64(len(dec)),
		FileType:    fileType,
		FileContent: dec,
	}

	if err := ctrl.DB.Create(&doc).Error; err != nil {
		_ = os.Remove(filePath)
		response.InternalServerError(c, "Failed to save record to database: "+err.Error())
		return
	}

	response.Created(c, "Uploaded strategic audit plan document created successfully", doc)
}

func (ctrl *UploadedStrategicPlanDocumentController) List(c *gin.Context) {
	var docs []models.UploadedStrategicPlanDocument
	if err := ctrl.DB.Order("created_at desc").Find(&docs).Error; err != nil {
		response.InternalServerError(c, "Failed to retrieve documents: "+err.Error())
		return
	}

	response.OK(c, "Uploaded strategic audit plan documents retrieved successfully", docs)
}

func (ctrl *UploadedStrategicPlanDocumentController) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid strategic audit plan document ID")
		return
	}

	var doc models.UploadedStrategicPlanDocument
	if err := ctrl.DB.First(&doc, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "Uploaded strategic audit plan document not found")
			return
		}
		response.InternalServerError(c, "Failed to fetch document")
		return
	}

	if err := ctrl.DB.Delete(&doc).Error; err != nil {
		response.InternalServerError(c, "Failed to delete record: "+err.Error())
		return
	}

	_ = os.Remove(doc.FilePath)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Uploaded strategic audit plan document deleted successfully",
	})
}

func (ctrl *UploadedStrategicPlanDocumentController) Download(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid strategic audit plan document ID")
		return
	}

	var doc models.UploadedStrategicPlanDocument
	if err := ctrl.DB.First(&doc, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "Uploaded strategic audit plan document not found")
			return
		}
		response.InternalServerError(c, "Failed to fetch document")
		return
	}

	contentType := getDocumentContentType(doc.FileName, doc.FileType)
	encodedFileName := url.PathEscape(doc.FileName)

	if len(doc.FileContent) > 0 {
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", doc.FileName, encodedFileName))
		c.Header("Content-Type", contentType)
		c.Data(http.StatusOK, contentType, doc.FileContent)
		return
	}

	if _, err := os.Stat(doc.FilePath); os.IsNotExist(err) {
		response.NotFound(c, "Physical file not found on server")
		return
	}

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", doc.FileName, encodedFileName))
	c.Header("Content-Type", contentType)
	c.File(doc.FilePath)
}

func (ctrl *UploadedStrategicPlanDocumentController) View(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid strategic audit plan document ID")
		return
	}

	var doc models.UploadedStrategicPlanDocument
	if err := ctrl.DB.First(&doc, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "Uploaded strategic audit plan document not found")
			return
		}
		response.InternalServerError(c, "Failed to fetch document")
		return
	}

	contentType := getDocumentContentType(doc.FileName, doc.FileType)
	encodedFileName := url.PathEscape(doc.FileName)

	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"; filename*=UTF-8''%s", doc.FileName, encodedFileName))
	c.Header("Content-Type", contentType)

	if len(doc.FileContent) > 0 {
		c.Data(http.StatusOK, contentType, doc.FileContent)
		return
	}

	if _, err := os.Stat(doc.FilePath); os.IsNotExist(err) {
		response.NotFound(c, "Physical file not found on server")
		return
	}

	c.File(doc.FilePath)
}
