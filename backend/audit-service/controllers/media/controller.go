package media

import (
	"audit-service/pkg/response"
	"audit-service/services/media"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type MediaController struct {
	mediaSvc media.MediaServiceInterface
}

func NewMediaController(mediaSvc media.MediaServiceInterface) *MediaController {
	return &MediaController{
		mediaSvc: mediaSvc,
	}
}

func (ctrl *MediaController) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "File is required", err.Error())
		return
	}
	defer file.Close()

	folder := c.DefaultPostForm("folder", "audit")

	attachment, err := ctrl.mediaSvc.UploadFile(c.Request.Context(), file, header.Filename, folder)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to upload file", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "File uploaded successfully", attachment)
}

func (ctrl *MediaController) Download(c *gin.Context) {
	id := c.Param("id")
	filePath := filepath.Join("uploads", id)
	c.File(filePath)
}
