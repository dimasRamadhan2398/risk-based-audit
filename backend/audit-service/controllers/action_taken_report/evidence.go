package actiontakenreport

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"audit-service/models"
	"audit-service/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AllowedEvidenceExtensions are the file types accepted as evidence
var AllowedEvidenceExtensions = map[string]bool{
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	".ppt": true, ".pptx": true, ".csv": true, ".txt": true,
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true,
	".zip": true,
}

// errUnsafePath is returned for a stored name that would leave the evidence
// directory or does not name a regular file
var errUnsafePath = errors.New("unsafe evidence path")

// displayFileName cleans a client file name for storing and showing: the base
// name only, no control characters, at most 200 characters
func displayFileName(name string) string {
	name = strings.ReplaceAll(name, `\`, "/")
	name = filepath.Base(name)
	name = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || r == '"' || r == '/' {
			return -1
		}
		return r
	}, name)
	name = strings.TrimSpace(name)
	if r := []rune(name); len(r) > 200 {
		ext := filepath.Ext(name)
		if len([]rune(ext)) > 20 {
			ext = ""
		}
		name = string(r[:200-len([]rune(ext))]) + ext
	}
	if name == "" || name == "." || name == ".." {
		return "evidence"
	}
	return name
}

// storageSafeName keeps only [A-Za-z0-9._-] of name (others become "_"),
// at most 120 characters, never starting with a dot
func storageSafeName(name string) string {
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	s := strings.TrimLeft(b.String(), ".")
	if len(s) > 120 {
		ext := filepath.Ext(s)
		if len(ext) > 20 {
			ext = ""
		}
		s = s[:120-len(ext)] + ext
	}
	if s == "" {
		s = "evidence"
	}
	return s
}

// ResolveEvidencePath returns the absolute path of stored inside dir. stored
// must be a plain file name (no separators, not "." or ".."); the joined path
// must stay inside dir and be a regular file (symlinks are refused).
func ResolveEvidencePath(dir, stored string) (string, error) {
	if stored == "" || stored == "." || stored == ".." ||
		strings.ContainsAny(stored, `/\`) || strings.ContainsRune(stored, 0) ||
		stored != filepath.Base(stored) {
		return "", errUnsafePath
	}
	base, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	full := filepath.Join(base, stored)
	rel, err := filepath.Rel(base, full)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", errUnsafePath
	}
	info, err := os.Lstat(full)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", errUnsafePath
	}
	return full, nil
}

// UploadEvidence stores a file (multipart field "file") for the ATR. The PIC
// or an admin; PLANNED/IN_PROGRESS only. The file is written as
// "<evidence uuid>-<safe name>" in the evidence directory, never served
// statically.
func (ctl *Controller) UploadEvidence(c *gin.Context) {
	a := actor(c)
	r, ok := ctl.load(c)
	if !ok {
		return
	}
	if !a.CanActAsPIC(r) {
		forbidden(c, "Only the assigned PIC can upload evidence")
		return
	}
	if !statusIn(r.Status, models.ATRStatusPlanned, models.ATRStatusInProgress) {
		conflict(c, "Evidence cannot be added while the action taken report is "+r.Status)
		return
	}

	maxBytes := ctl.MaxEvidenceBytes
	if maxBytes <= 0 {
		maxBytes = DefaultMaxEvidenceBytes
	}
	// Room for the multipart envelope on top of the file itself
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes+1<<20)
	fh, err := c.FormFile("file")
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			response.Error(c, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", fmt.Sprintf("File exceeds the maximum size of %d MB", maxBytes>>20), "")
			return
		}
		invalid(c, "A file is required (multipart field \"file\")")
		return
	}
	if fh.Size > maxBytes {
		response.Error(c, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", fmt.Sprintf("File exceeds the maximum size of %d MB", maxBytes>>20), "")
		return
	}
	if fh.Size == 0 {
		invalid(c, "The file is empty")
		return
	}
	name := displayFileName(fh.Filename)
	ext := strings.ToLower(filepath.Ext(name))
	if !AllowedEvidenceExtensions[ext] {
		invalid(c, "File type not allowed. Allowed: pdf, doc, docx, xls, xlsx, ppt, pptx, csv, txt, png, jpg, jpeg, gif, webp, zip")
		return
	}

	src, err := fh.Open()
	if err != nil {
		response.InternalServerError(c, "Failed to read the uploaded file")
		return
	}
	defer src.Close()

	if err := os.MkdirAll(ctl.EvidenceDir, 0o750); err != nil {
		response.InternalServerError(c, "Failed to prepare evidence storage")
		return
	}
	evidenceID := uuid.New()
	stored := evidenceID.String() + "-" + storageSafeName(name)
	path := filepath.Join(ctl.EvidenceDir, stored)
	dst, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		response.InternalServerError(c, "Failed to store the file")
		return
	}
	written, copyErr := io.Copy(dst, io.LimitReader(src, maxBytes+1))
	closeErr := dst.Close()
	if copyErr != nil || closeErr != nil || written > maxBytes {
		_ = os.Remove(path)
		if written > maxBytes {
			response.Error(c, http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", fmt.Sprintf("File exceeds the maximum size of %d MB", maxBytes>>20), "")
			return
		}
		response.InternalServerError(c, "Failed to store the file")
		return
	}

	contentType := fh.Header.Get("Content-Type")
	if len(contentType) > 150 {
		contentType = ""
	}
	ev := models.ActionTakenReportEvidence{
		ID:                  evidenceID,
		ActionTakenReportID: r.ID,
		FileName:            name,
		StoredName:          stored,
		ContentType:         contentType,
		FileSize:            written,
		UploadedBy:          a.DisplayName(),
		UploadedByUserID:    a.UserID,
		UploadedAt:          ctl.now(),
	}
	if err := ctl.DB.Create(&ev).Error; err != nil {
		_ = os.Remove(path)
		response.InternalServerError(c, "Failed to save the evidence record")
		return
	}
	ctl.respondOne(c, a, r.ID, "Evidence uploaded")
}

// DownloadEvidence sends an evidence file to a caller who may view the ATR.
// It is always sent as an attachment with a generic content type.
func (ctl *Controller) DownloadEvidence(c *gin.Context) {
	a := actor(c)
	r, ok := ctl.loadVisible(c, a)
	if !ok {
		return
	}
	evidenceID, err := uuid.Parse(c.Param("evidenceId"))
	if err != nil {
		response.BadRequest(c, "Invalid evidence ID")
		return
	}
	var ev models.ActionTakenReportEvidence
	err = ctl.DB.Where("id = ? AND action_taken_report_id = ?", evidenceID, r.ID).First(&ev).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.NotFound(c, "Evidence not found")
		return
	}
	if err != nil {
		response.InternalServerError(c, "Failed to fetch evidence")
		return
	}

	path, err := ResolveEvidencePath(ctl.EvidenceDir, ev.StoredName)
	if errors.Is(err, errUnsafePath) {
		forbidden(c, "Evidence file path is not allowed")
		return
	}
	if err != nil {
		response.NotFound(c, "Evidence file not found on the server")
		return
	}
	f, err := os.Open(path)
	if err != nil {
		response.NotFound(c, "Evidence file not found on the server")
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		response.InternalServerError(c, "Failed to read evidence file")
		return
	}

	disposition := mime.FormatMediaType("attachment", map[string]string{"filename": ev.FileName})
	if disposition == "" {
		disposition = `attachment; filename="evidence"`
	}
	c.DataFromReader(http.StatusOK, info.Size(), "application/octet-stream", f, map[string]string{
		"Content-Disposition":    disposition,
		"X-Content-Type-Options": "nosniff",
		"Cache-Control":          "private, no-store",
	})
}
