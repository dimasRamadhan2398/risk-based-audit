package media

import (
	"audit-service/pkg/response"
	"audit-service/pkg/safefile"
	"audit-service/services/media"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
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

	attachment, err := ctrl.mediaSvc.UploadFile(c.Request.Context(), file, header.Filename, folder)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to upload file", err.Error())
		return
	}

	response.Success(c, http.StatusOK, "File uploaded successfully", attachment)
}

// uploadsRoot is the only directory media downloads are read from.
var uploadsRoot = "uploads"

// Download serves an uploaded file. The route is under the /api/v1 group, so a
// valid token is required. Files are resolved inside uploadsRoot, private
// subdirectories are never served here, and the response uses safefile
// (allowlisted Content-Type, attachment unless the file is safe to show inline).
func (ctrl *MediaController) Download(c *gin.Context) {
	id := c.Param("id")
	if unescaped, err := url.QueryUnescape(id); err == nil {
		id = unescaped
	}

	// Also support query param "path" or "filepath"
	reqPath := c.Query("path")
	if reqPath == "" {
		reqPath = c.Query("filepath")
	}
	if unescaped, err := url.QueryUnescape(reqPath); err == nil {
		reqPath = unescaped
	}

	rel := id
	if reqPath != "" {
		rel = strings.TrimPrefix(strings.TrimPrefix(filepath.ToSlash(reqPath), "/"), "uploads/")
	}

	found := ""
	if !safefile.IsPrivatePath(rel) {
		if full, err := safefile.ResolveUnder(uploadsRoot, rel); err == nil {
			found = full
		}
	}
	if found == "" {
		found = findUploadByName(uploadsRoot, id)
	}
	if found == "" {
		response.Error(c, http.StatusNotFound, "NOT_FOUND", "File not found", "")
		return
	}

	download := c.Query("download") == "true"
	if err := safefile.ServeFile(c, found, filepath.Base(found), download); err != nil {
		response.Error(c, http.StatusNotFound, "NOT_FOUND", "File not found", "")
	}
}

// findUploadByName looks for a file called name (or ending in /name) under root.
// Private subdirectories are skipped, and symlinks are not followed.
func findUploadByName(root, name string) string {
	name = filepath.Base(filepath.FromSlash(name))
	if name == "" || name == "." || name == string(filepath.Separator) {
		return ""
	}
	var found string
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, relErr := filepath.Rel(root, p)
		if relErr != nil {
			return nil
		}
		if d.IsDir() {
			if rel != "." && safefile.IsPrivatePath(rel) {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}
		if d.Name() == name || strings.HasSuffix(filepath.ToSlash(p), "/"+name) {
			found = p
			return filepath.SkipAll
		}
		return nil
	})
	return found
}
