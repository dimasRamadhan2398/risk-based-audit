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

type UploadedPlanDocumentController struct {
	DB *gorm.DB
}

const uploadedPlanDocumentsDir = "./uploads/uploaded-plan-documents"

func NewUploadedPlanDocumentController(db *gorm.DB) *UploadedPlanDocumentController {
	if err := os.MkdirAll(uploadedPlanDocumentsDir, 0750); err != nil {
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

// Upload accepts multipart/form-data with a "file" field, a base64
// "fileContent" form field, or a JSON body (UploadPlanRequest).
func (ctrl *UploadedPlanDocumentController) Upload(c *gin.Context) {
	if !beginDocumentUpload(c) {
		return
	}
	title := c.PostForm("title")
	description := c.PostForm("description")
	fileName := c.PostForm("fileName")
	fileType := c.PostForm("fileType")

	var dec []byte
	clientName := fileName

	if _, err := c.FormFile("file"); err == nil {
		var ok bool
		dec, clientName, ok = readFormDocument(c, fileName)
		if !ok {
			return
		}
	} else {
		base64Data := c.PostForm("fileContent")
		if base64Data == "" && c.ContentType() == "application/json" {
			var req UploadPlanRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				if upload.IsTooLarge(err) {
					upload.Respond(c, upload.ErrTooLarge(uploadedDocMaxBytes))
					return
				}
			} else {
				title = req.Title
				description = req.Description
				fileName = req.FileName
				fileType = req.FileType
				base64Data = req.FileContent
				clientName = fileName
			}
		}
		if base64Data == "" {
			response.BadRequest(c, "file is required")
			return
		}
		var ok bool
		if dec, ok = decodeBase64Document(c, base64Data); !ok {
			return
		}
	}

	if title == "" || fileName == "" {
		response.BadRequest(c, "title and fileName are required")
		return
	}

	name, filePath, ok := saveDocument(c, uploadedPlanDocumentsDir, clientName, dec)
	if !ok {
		return
	}

	// Save metadata in database
	paper := models.UploadedPlanDocument{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
		FileName:    name,
		FilePath:    filePath,
		FileSize:    int64(len(dec)),
		FileContent: dec,
		FileType:    fileType,
	}

	if err := ctrl.DB.Create(&paper).Error; err != nil {
		removeDocumentFile(filePath)
		response.InternalServerError(c, "Failed to save record to database")
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

func (ctrl *UploadedPlanDocumentController) find(c *gin.Context) (*models.UploadedPlanDocument, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid plan document ID")
		return nil, false
	}
	var paper models.UploadedPlanDocument
	if err := ctrl.DB.First(&paper, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "Uploaded plan document not found")
			return nil, false
		}
		response.InternalServerError(c, "Failed to fetch uploaded plan document")
		return nil, false
	}
	return &paper, true
}

func (ctrl *UploadedPlanDocumentController) Delete(c *gin.Context) {
	paper, ok := ctrl.find(c)
	if !ok {
		return
	}

	if err := ctrl.DB.Delete(paper).Error; err != nil {
		response.InternalServerError(c, "Failed to delete uploaded plan document record")
		return
	}

	removeDocumentFile(paper.FilePath)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Uploaded plan document deleted successfully",
	})
}

func (ctrl *UploadedPlanDocumentController) Download(c *gin.Context) {
	if paper, ok := ctrl.find(c); ok {
		upload.Serve(c, paper.FileName, paper.FileContent, paper.FilePath, false)
	}
}

func (ctrl *UploadedPlanDocumentController) View(c *gin.Context) {
	if paper, ok := ctrl.find(c); ok {
		upload.Serve(c, paper.FileName, paper.FileContent, paper.FilePath, true)
	}
}
