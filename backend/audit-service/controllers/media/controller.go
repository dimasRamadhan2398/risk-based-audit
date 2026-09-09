package media

import (
	"audit-service/pkg/response"
	"audit-service/services/media"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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

	if header.Size > 10*1024*1024 {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "File size exceeds maximum limit of 10MB", "Max allowed file size is 10MB")
		return
	}

<<<<<<< HEAD
	folder := c.DefaultPostForm("folder", "audit")
=======
	folder := c.PostForm("folder")
	featureName := c.PostForm("feature_name")
	documentID := c.PostForm("document_id")

	if folder == "" {
		if featureName != "" && documentID != "" {
			folder = fmt.Sprintf("Auditsphere/%s/%s", featureName, documentID)
		} else if featureName != "" {
			folder = fmt.Sprintf("Auditsphere/%s", featureName)
		} else {
			folder = "Auditsphere/audit"
		}
	} else if !strings.HasPrefix(folder, "Auditsphere/") && folder != "Auditsphere" {
		folder = "Auditsphere/" + strings.TrimPrefix(folder, "/")
	}
>>>>>>> cbc5d39943d30c4e2499ca6e2d081099b879c9ca

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
	if _, err := os.Stat(filePath); err == nil {
		c.File(filePath)
		return
	}

	// Search recursively in uploads directory for matching file
	var foundPath string
	_ = filepath.Walk("uploads", func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && (info.Name() == id || strings.HasSuffix(info.Name(), id)) {
			foundPath = path
			return filepath.SkipAll
		}
		return nil
	})

	if foundPath != "" {
		c.File(foundPath)
		return
	}

	c.File(filePath)
}
