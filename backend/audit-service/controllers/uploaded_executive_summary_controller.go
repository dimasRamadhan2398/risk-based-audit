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

type UploadedExecutiveSummaryController struct {
	DB *gorm.DB
}

const uploadedExecutiveSummariesDir = "./uploads/uploaded-executive-summaries"

func NewUploadedExecutiveSummaryController(db *gorm.DB) *UploadedExecutiveSummaryController {
	if err := os.MkdirAll(uploadedExecutiveSummariesDir, 0750); err != nil {
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
	name, filePath, ok := saveDocument(c, uploadedExecutiveSummariesDir, clientName, dec)
	if !ok {
		return
	}

	doc := models.UploadedExecutiveSummary{
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

func (ctrl *UploadedExecutiveSummaryController) find(c *gin.Context) (*models.UploadedExecutiveSummary, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid document ID")
		return nil, false
	}
	var doc models.UploadedExecutiveSummary
	if err := ctrl.DB.First(&doc, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "Uploaded executive summary document not found")
			return nil, false
		}
		response.InternalServerError(c, "Failed to fetch document")
		return nil, false
	}
	return &doc, true
}

func (ctrl *UploadedExecutiveSummaryController) Delete(c *gin.Context) {
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
		"message": "Uploaded executive summary document deleted successfully",
	})
}

// Download always sends the file as an attachment.
func (ctrl *UploadedExecutiveSummaryController) Download(c *gin.Context) {
	if doc, ok := ctrl.find(c); ok {
		upload.Serve(c, doc.FileName, doc.FileContent, doc.FilePath, false)
	}
}

// View shows PDF, images and plain text inline; other types are downloaded.
func (ctrl *UploadedExecutiveSummaryController) View(c *gin.Context) {
	if doc, ok := ctrl.find(c); ok {
		upload.Serve(c, doc.FileName, doc.FileContent, doc.FilePath, true)
	}
}
