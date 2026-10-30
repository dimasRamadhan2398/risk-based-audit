package controllers

import (
	"fmt"
	"net/http"
	"os"

	"audit-service/models"
	"audit-service/pkg/response"
	"audit-service/pkg/upload"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UploadedAuditResultReportController struct {
	DB *gorm.DB
}

const uploadedAuditResultReportsDir = "./uploads/uploaded-audit-result-reports"

func NewUploadedAuditResultReportController(db *gorm.DB) *UploadedAuditResultReportController {
	if err := os.MkdirAll(uploadedAuditResultReportsDir, 0750); err != nil {
		fmt.Printf("Failed to create uploads directory: %v\n", err)
	}
	return &UploadedAuditResultReportController{DB: db}
}

type UploadAuditResultReportRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	FileName    string `json:"fileName" binding:"required"`
	FileType    string `json:"fileType"`
	FileContent string `json:"fileContent" binding:"required"`
}

func (ctrl *UploadedAuditResultReportController) Upload(c *gin.Context) {
	if !beginDocumentUpload(c) {
		return
	}
	title := c.PostForm("title")
	description := c.PostForm("description")
	fileName := c.PostForm("fileName")
	fileType := c.PostForm("fileType")

	if title == "" || fileName == "" {
		response.BadRequest(c, "title and fileName are required")
		return
	}

	dec, clientName, ok := readFormDocument(c, fileName)
	if !ok {
		return
	}
	name, filePath, ok := saveDocument(c, uploadedAuditResultReportsDir, clientName, dec)
	if !ok {
		return
	}

	doc := models.UploadedAuditResultReport{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
		FileName:    name,
		FilePath:    filePath,
		FileSize:    int64(len(dec)),
		FileContent: dec,
		FileType:    fileType,
	}

	if err := ctrl.DB.Create(&doc).Error; err != nil {
		removeDocumentFile(filePath)
		response.InternalServerError(c, "Failed to save record to database")
		return
	}

	response.Created(c, "Audit result report document uploaded successfully", doc)
}

func (ctrl *UploadedAuditResultReportController) List(c *gin.Context) {
	var docs []models.UploadedAuditResultReport
	if err := ctrl.DB.Order("created_at DESC").Find(&docs).Error; err != nil {
		response.InternalServerError(c, "Failed to retrieve uploaded audit result report documents: "+err.Error())
		return
	}
	response.OK(c, "Uploaded audit result report documents retrieved successfully", docs)
}

func (ctrl *UploadedAuditResultReportController) find(c *gin.Context) (*models.UploadedAuditResultReport, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid document ID")
		return nil, false
	}
	var doc models.UploadedAuditResultReport
	if err := ctrl.DB.First(&doc, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "Uploaded audit result report document not found")
			return nil, false
		}
		response.InternalServerError(c, "Failed to fetch document")
		return nil, false
	}
	return &doc, true
}

func (ctrl *UploadedAuditResultReportController) Delete(c *gin.Context) {
	doc, ok := ctrl.find(c)
	if !ok {
		return
	}

	if err := ctrl.DB.Delete(doc).Error; err != nil {
		response.InternalServerError(c, "Failed to delete document record")
		return
	}

	removeDocumentFile(doc.FilePath)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Uploaded audit result report document deleted successfully",
	})
}

// Download always sends the file as an attachment.
func (ctrl *UploadedAuditResultReportController) Download(c *gin.Context) {
	if doc, ok := ctrl.find(c); ok {
		upload.Serve(c, doc.FileName, doc.FileContent, doc.FilePath, false)
	}
}

// View shows PDF, images and plain text inline; other types are downloaded.
func (ctrl *UploadedAuditResultReportController) View(c *gin.Context) {
	if doc, ok := ctrl.find(c); ok {
		upload.Serve(c, doc.FileName, doc.FileContent, doc.FilePath, true)
	}
}
