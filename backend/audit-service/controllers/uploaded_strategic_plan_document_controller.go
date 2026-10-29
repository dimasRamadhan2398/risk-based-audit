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

type UploadedStrategicPlanDocumentController struct {
	DB *gorm.DB
}

const uploadedStrategicPlanDocumentsDir = "./uploads/uploaded-strategic-plan-documents"

func NewUploadedStrategicPlanDocumentController(db *gorm.DB) *UploadedStrategicPlanDocumentController {
	if err := os.MkdirAll(uploadedStrategicPlanDocumentsDir, 0750); err != nil {
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

// Upload accepts multipart/form-data with a "file" field, or a JSON body
// (UploadStrategicPlanRequest) with base64 fileContent.
func (ctrl *UploadedStrategicPlanDocumentController) Upload(c *gin.Context) {
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
		var req UploadStrategicPlanRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			if upload.IsTooLarge(err) {
				upload.Respond(c, upload.ErrTooLarge(uploadedDocMaxBytes))
				return
			}
			response.BadRequest(c, "invalid request: "+err.Error())
			return
		}

		title = req.Title
		description = req.Description
		fileName = req.FileName
		fileType = req.FileType
		clientName = fileName

		var ok bool
		if dec, ok = decodeBase64Document(c, req.FileContent); !ok {
			return
		}
	}

	if title == "" || fileName == "" {
		response.BadRequest(c, "title and fileName are required")
		return
	}

	name, filePath, ok := saveDocument(c, uploadedStrategicPlanDocumentsDir, clientName, dec)
	if !ok {
		return
	}

	doc := models.UploadedStrategicPlanDocument{
		Title:       title,
		Description: description,
		FileName:    name,
		FilePath:    filePath,
		FileSize:    int64(len(dec)),
		FileType:    fileType,
		FileContent: dec,
	}

	if err := ctrl.DB.Create(&doc).Error; err != nil {
		removeDocumentFile(filePath)
		response.InternalServerError(c, "Failed to save record to database")
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

func (ctrl *UploadedStrategicPlanDocumentController) find(c *gin.Context) (*models.UploadedStrategicPlanDocument, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid strategic audit plan document ID")
		return nil, false
	}
	var doc models.UploadedStrategicPlanDocument
	if err := ctrl.DB.First(&doc, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "Uploaded strategic audit plan document not found")
			return nil, false
		}
		response.InternalServerError(c, "Failed to fetch document")
		return nil, false
	}
	return &doc, true
}

func (ctrl *UploadedStrategicPlanDocumentController) Delete(c *gin.Context) {
	doc, ok := ctrl.find(c)
	if !ok {
		return
	}

	if err := ctrl.DB.Delete(doc).Error; err != nil {
		response.InternalServerError(c, "Failed to delete record: "+err.Error())
		return
	}

	removeDocumentFile(doc.FilePath)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Uploaded strategic audit plan document deleted successfully",
	})
}

func (ctrl *UploadedStrategicPlanDocumentController) Download(c *gin.Context) {
	if doc, ok := ctrl.find(c); ok {
		upload.Serve(c, doc.FileName, doc.FileContent, doc.FilePath, false)
	}
}

func (ctrl *UploadedStrategicPlanDocumentController) View(c *gin.Context) {
	if doc, ok := ctrl.find(c); ok {
		upload.Serve(c, doc.FileName, doc.FileContent, doc.FilePath, true)
	}
}
