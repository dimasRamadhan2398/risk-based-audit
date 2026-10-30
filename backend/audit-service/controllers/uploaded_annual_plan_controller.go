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

type UploadedAnnualPlanController struct {
	DB *gorm.DB
}

const uploadedAnnualPlansDir = "./uploads/uploaded-annual-plans"

func NewUploadedAnnualPlanController(db *gorm.DB) *UploadedAnnualPlanController {
	if err := os.MkdirAll(uploadedAnnualPlansDir, 0750); err != nil {
		fmt.Printf("Failed to create uploads directory: %v\n", err)
	}
	return &UploadedAnnualPlanController{DB: db}
}

// The request is received as multipart/form-data, so we no longer need a struct for JSON binding.
// The fields are extracted manually from the form.
// Keeping the struct for documentation purposes only (not used in binding).
type UploadAnnualPlanRequest struct {
	Title       string
	Description string
	FileName    string
	FileType    string
	// FileContent is omitted because the file is streamed directly.
}

func (ctrl *UploadedAnnualPlanController) Upload(c *gin.Context) {
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
	name, filePath, ok := saveDocument(c, uploadedAnnualPlansDir, clientName, dec)
	if !ok {
		return
	}

	doc := models.UploadedAnnualPlan{
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

	removeDocumentFile(doc.FilePath)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Uploaded annual audit plan document deleted successfully",
	})
}

func (ctrl *UploadedAnnualPlanController) find(c *gin.Context) (*models.UploadedAnnualPlan, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid document ID")
		return nil, false
	}
	var doc models.UploadedAnnualPlan
	if err := ctrl.DB.First(&doc, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "Uploaded annual audit plan document not found")
			return nil, false
		}
		response.InternalServerError(c, "Failed to fetch document")
		return nil, false
	}
	return &doc, true
}

func (ctrl *UploadedAnnualPlanController) Download(c *gin.Context) {
	if doc, ok := ctrl.find(c); ok {
		upload.Serve(c, doc.FileName, doc.FileContent, doc.FilePath, false)
	}
}

func (ctrl *UploadedAnnualPlanController) View(c *gin.Context) {
	if doc, ok := ctrl.find(c); ok {
		upload.Serve(c, doc.FileName, doc.FileContent, doc.FilePath, true)
	}
}
