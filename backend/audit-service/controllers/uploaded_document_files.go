package controllers

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"audit-service/pkg/upload"

	"github.com/gin-gonic/gin"
)

// Shared upload steps for the uploaded-document and imported-working-paper
// controllers. See pkg/upload for the checks themselves.

// uploadedDocMaxBytes is the per-file limit for uploaded documents.
const uploadedDocMaxBytes = upload.DefaultMaxBytes

// beginDocumentUpload caps the body and parses the form. It must run before
// any c.PostForm call. It writes the error response and returns false on failure.
func beginDocumentUpload(c *gin.Context) bool {
	if e := upload.ParseForm(c, uploadedDocMaxBytes); e != nil {
		upload.Respond(c, e)
		return false
	}
	return true
}

// readFormDocument reads the multipart "file" field. The client name is the
// fileName form field when given, else the multipart file name.
func readFormDocument(c *gin.Context, fileName string) (data []byte, clientName string, ok bool) {
	fh, e := upload.FormFile(c, "file", uploadedDocMaxBytes)
	if e != nil {
		upload.Respond(c, e)
		return nil, "", false
	}
	data, e = upload.ReadFormFile(fh, uploadedDocMaxBytes)
	if e != nil {
		upload.Respond(c, e)
		return nil, "", false
	}
	if fileName == "" {
		fileName = fh.Filename
	}
	return data, fileName, true
}

// decodeBase64Document decodes a base64 (optionally data-URL) payload within the size limit.
func decodeBase64Document(c *gin.Context, raw string) ([]byte, bool) {
	if idx := strings.Index(raw, ";base64,"); idx != -1 {
		raw = raw[idx+8:]
	}
	if int64(base64.StdEncoding.DecodedLen(len(raw))) > uploadedDocMaxBytes+3 {
		upload.Respond(c, upload.ErrTooLarge(uploadedDocMaxBytes))
		return nil, false
	}
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		upload.Respond(c, &upload.Error{Status: 400, Code: "INVALID_BASE64", Message: "Invalid base64 file data"})
		return nil, false
	}
	return data, true
}

// saveDocument validates data (name, extension, magic bytes, size, malware)
// and stores it in dir under a server-generated name. It returns the cleaned
// display name and the stored path. On failure the response is written.
func saveDocument(c *gin.Context, dir, clientName string, data []byte) (name, filePath string, ok bool) {
	name, e := upload.Validate(c.Request.Context(), clientName, data, uploadedDocMaxBytes, upload.DocumentExtensions)
	if e != nil {
		upload.Respond(c, e)
		return "", "", false
	}
	_, filePath, err := upload.Store(dir, name, data)
	if err != nil {
		upload.RespondErr(c, err, "Failed to store the file")
		return "", "", false
	}
	return name, filePath, true
}

// removeDocumentFile deletes a stored file, refusing paths outside the uploads root.
func removeDocumentFile(path string) {
	if path == "" {
		return
	}
	if err := upload.RemoveStored(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Printf("Warning: not deleting physical file %q: %v\n", path, err)
	}
}
