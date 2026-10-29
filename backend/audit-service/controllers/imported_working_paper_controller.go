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

type ImportedWorkingPaperController struct {
	DB *gorm.DB
}

const importedWorkingPapersDir = "./uploads/imported-working-papers"

func NewImportedWorkingPaperController(db *gorm.DB) *ImportedWorkingPaperController {
	if err := os.MkdirAll(importedWorkingPapersDir, 0750); err != nil {
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
	name, filePath, ok := saveDocument(c, importedWorkingPapersDir, clientName, dec)
	if !ok {
		return
	}

	// Save metadata in database
	paper := models.ImportedWorkingPaper{
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

func (ctrl *ImportedWorkingPaperController) find(c *gin.Context) (*models.ImportedWorkingPaper, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid working paper ID")
		return nil, false
	}
	var paper models.ImportedWorkingPaper
	if err := ctrl.DB.First(&paper, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.NotFound(c, "Imported working paper not found")
			return nil, false
		}
		response.InternalServerError(c, "Failed to fetch imported working paper")
		return nil, false
	}
	return &paper, true
}

func (ctrl *ImportedWorkingPaperController) Delete(c *gin.Context) {
	paper, ok := ctrl.find(c)
	if !ok {
		return
	}

	if err := ctrl.DB.Delete(paper).Error; err != nil {
		response.InternalServerError(c, "Failed to delete imported working paper record")
		return
	}

	removeDocumentFile(paper.FilePath)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Imported working paper deleted successfully",
	})
}

func (ctrl *ImportedWorkingPaperController) Download(c *gin.Context) {
	if paper, ok := ctrl.find(c); ok {
		upload.Serve(c, paper.FileName, paper.FileContent, paper.FilePath, false)
	}
}

func (ctrl *ImportedWorkingPaperController) View(c *gin.Context) {
	if paper, ok := ctrl.find(c); ok {
		upload.Serve(c, paper.FileName, paper.FileContent, paper.FilePath, true)
	}
}
