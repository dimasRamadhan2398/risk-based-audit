package media

import (
	"audit-service/pkg/response"
	"audit-service/pkg/safefile"
	"audit-service/pkg/upload"
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

// Upload stores a file (multipart field "file"). The body is capped before it
// is parsed; the media service validates the name, type, content and malware
// scan, and the folder is confined to the uploads tree ("..", absolute paths
// and backslashes are rejected with 400).
func (ctrl *MediaController) Upload(c *gin.Context) {
	if e := upload.ParseForm(c, upload.DefaultMaxBytes); e != nil {
		upload.Respond(c, e)
		return
	}
	header, e := upload.FormFile(c, "file", upload.DefaultMaxBytes)
	if e != nil {
		upload.Respond(c, e)
		return
	}
	file, err := header.Open()
	if err != nil {
		response.Error(c, http.StatusBadRequest, "BAD_REQUEST", "File is required", "")
		return
	}
	defer file.Close()

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
	}
	cleanFolder, e := upload.ConfineFolder(folder)
	if e != nil {
		upload.Respond(c, e)
		return
	}

	attachment, err := ctrl.mediaSvc.UploadFile(c.Request.Context(), file, header.Filename, cleanFolder)
	if err != nil {
		if ue, ok := upload.AsError(err); ok {
			upload.Respond(c, ue)
			return
		}
		response.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to upload file", "")
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
