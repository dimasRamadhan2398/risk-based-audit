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

type UploadedConsultingDocumentController struct {
	DB *gorm.DB
}

const uploadedConsultingDocumentsDir = "./uploads/uploaded-consulting-documents"

func NewUploadedConsultingDocumentController(db *gorm.DB) *UploadedConsultingDocumentController {
	if err := os.MkdirAll(uploadedConsultingDocumentsDir, 0750); err != nil {
		fmt.Printf("Failed to create uploads directory: %v\n", err)
	}
	return &UploadedConsultingDocumentController{DB: db}
}

type UploadConsultingDocumentRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	FileName    string `json:"fileName" binding:"required"`
	FileType    string `json:"fileType"`
	FileContent string `json:"fileContent" binding:"required"`
}

func (ctrl *UploadedConsultingDocumentController) Upload(c *gin.Context) {
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
	name, filePath, ok := saveDocument(c, uploadedConsultingDocumentsDir, clientName, dec)
	if !ok {
		return
	}

	doc := models.UploadedConsultingDocument{
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

	response.Created(c, "Consulting document uploaded successfully", doc)
}

func (ctrl *UploadedConsultingDocumentController) List(c *gin.Context) {
	var docs []models.UploadedConsultingDocument
	if err := ctrl.DB.Order("created_at DESC").Find(&docs).Error; err != nil {
		response.InternalServerError(c, "Failed to retrieve uploaded consulting documents: "+err.Error())
		return
	}
	response.OK(c, "Uploaded consulting documents retrieved successfully", docs)
}

func (ctrl *UploadedConsultingDocumentController) find(c *gin.Context) (*models.UploadedConsultingDocument, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid document ID")
		return nil, false
	}
	var doc models.UploadedConsultingDocument
	if err := ctrl.DB.First(&doc, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "Uploaded consulting document not found")
			return nil, false
		}
		response.InternalServerError(c, "Failed to fetch document")
		return nil, false
	}
	return &doc, true
}

func (ctrl *UploadedConsultingDocumentController) Delete(c *gin.Context) {
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
		"message": "Uploaded consulting document deleted successfully",
	})
}

// Download always sends the file as an attachment.
func (ctrl *UploadedConsultingDocumentController) Download(c *gin.Context) {
	if doc, ok := ctrl.find(c); ok {
		upload.Serve(c, doc.FileName, doc.FileContent, doc.FilePath, false)
	}
}

// View shows PDF, images and plain text inline; other types are downloaded.
func (ctrl *UploadedConsultingDocumentController) View(c *gin.Context) {
	if doc, ok := ctrl.find(c); ok {
		upload.Serve(c, doc.FileName, doc.FileContent, doc.FilePath, true)
	}
}
