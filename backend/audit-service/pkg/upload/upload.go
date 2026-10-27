// Package upload validates and stores user uploads.
//
// Every upload path in audit-service goes through the same steps:
//
//  1. LimitBody caps the request body before anything parses it.
//  2. The client file name is checked (no path separators, no "..") and
//     cleaned for display; it is never used as a path.
//  3. The extension must be on an allowlist, and the first bytes must match
//     that extension (magic bytes / http.DetectContentType).
//  4. The bytes are scanned by ClamAV (pkg/avscan) before they are written.
//  5. Store writes under a server-generated name with O_EXCL inside a
//     directory that must lie under the uploads root.
//
// Downloads go through Serve / ServeStored, which take the Content-Type from
// the extension allowlist (pkg/safefile), send nosniff and a sandbox CSP, and
// only allow inline display for PDF, images and plain text.
package upload

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"audit-service/pkg/avscan"
	"audit-service/pkg/response"
	"audit-service/pkg/safefile"

	"github.com/gin-gonic/gin"
)

// DefaultMaxBytes is the largest file accepted (10 MB, as before).
const DefaultMaxBytes int64 = 10 << 20

// envelope is room for multipart headers and the other form fields.
const envelope int64 = 1 << 20

// UploadsRoot is the directory every stored upload must lie under.
var UploadsRoot = "./uploads"

// DocumentExtensions are the file types accepted for uploaded documents and
// media. It matches the ATR evidence allowlist without archives.
var DocumentExtensions = map[string]bool{
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	".ppt": true, ".pptx": true, ".csv": true, ".txt": true,
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true,
}

// AllowedList is the human-readable form of DocumentExtensions.
const AllowedList = "pdf, doc, docx, xls, xlsx, ppt, pptx, csv, txt, png, jpg, jpeg, gif, webp"

// Error is an upload rejection with the HTTP status and code to send.
type Error struct {
	Status  int
	Code    string
	Message string
	Err     error
}

func (e *Error) Error() string { return e.Message }
func (e *Error) Unwrap() error { return e.Err }

func newErr(status int, code, msg string) *Error {
	return &Error{Status: status, Code: code, Message: msg}
}

// AsError returns err as an *Error when it is one.
func AsError(err error) (*Error, bool) {
	var ue *Error
	if errors.As(err, &ue) {
		return ue, true
	}
	return nil, false
}

// ErrTooLarge builds the 413 response for max bytes.
func ErrTooLarge(max int64) *Error {
	return newErr(http.StatusRequestEntityTooLarge, "FILE_TOO_LARGE", fmt.Sprintf("File exceeds the maximum size of %d MB", max>>20))
}

// Respond writes e in the service's standard error envelope.
func Respond(c *gin.Context, e *Error) {
	response.Error(c, e.Status, e.Code, e.Message, "")
}

// RespondErr writes err; anything that is not an *Error becomes fallback (500).
func RespondErr(c *gin.Context, err error, fallback string) {
	if ue, ok := AsError(err); ok {
		Respond(c, ue)
		return
	}
	response.InternalServerError(c, fallback)
}

// LimitBody caps the request body at max bytes of file plus room for the
// multipart envelope. Call it before c.PostForm, c.FormFile or JSON binding.
// When base64 is true the cap allows for base64 inflation (JSON uploads).
func LimitBody(c *gin.Context, max int64, base64 bool) {
	limit := max + envelope
	if base64 {
		limit = max/3*4 + 4 + envelope
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
}

// ParseForm caps the body and, for multipart requests, parses the form so
// that c.PostForm and c.FormFile work. JSON bodies are only capped (with the
// base64 allowance); bind them afterwards and check IsTooLarge on error.
// Call it first in the handler, before reading any field.
func ParseForm(c *gin.Context, max int64) *Error {
	// Some forms carry the file as a base64 field, so the body cap always has
	// room for base64; the per-file limit (max) is enforced separately.
	LimitBody(c, max, true)
	if c.ContentType() == "multipart/form-data" {
		if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
			if IsTooLarge(err) {
				return ErrTooLarge(max)
			}
			return newErr(http.StatusBadRequest, "INVALID_FORM", "Malformed multipart form")
		}
	}
	return nil
}

// IsTooLarge reports whether err comes from a body over the LimitBody cap.
func IsTooLarge(err error) bool {
	var mbe *http.MaxBytesError
	return errors.As(err, &mbe)
}

// FormFile returns the multipart file field. A body over the cap gives 413.
func FormFile(c *gin.Context, field string, max int64) (*multipart.FileHeader, *Error) {
	fh, err := c.FormFile(field)
	if err != nil {
		if IsTooLarge(err) {
			return nil, ErrTooLarge(max)
		}
		return nil, newErr(http.StatusBadRequest, "FILE_REQUIRED", fmt.Sprintf("A file is required (multipart field %q)", field))
	}
	if fh.Size > max {
		return nil, ErrTooLarge(max)
	}
	return fh, nil
}

// ReadAll reads at most max bytes from r. More than max gives 413.
func ReadAll(r io.Reader, max int64) ([]byte, *Error) {
	data, err := io.ReadAll(io.LimitReader(r, max+1))
	if err != nil {
		if IsTooLarge(err) {
			return nil, ErrTooLarge(max)
		}
		return nil, newErr(http.StatusBadRequest, "FILE_UNREADABLE", "Failed to read the uploaded file")
	}
	if int64(len(data)) > max {
		return nil, ErrTooLarge(max)
	}
	return data, nil
}

// ReadFormFile opens and reads a multipart file within max bytes.
func ReadFormFile(fh *multipart.FileHeader, max int64) ([]byte, *Error) {
	f, err := fh.Open()
	if err != nil {
		return nil, newErr(http.StatusBadRequest, "FILE_UNREADABLE", "Failed to read the uploaded file")
	}
	defer f.Close()
	return ReadAll(f, max)
}

// CheckName validates a client-supplied file name and returns the cleaned
// display name. Names with path separators, NUL or a ".." element are
// rejected rather than silently rewritten, so a traversal attempt is visible.
func CheckName(name string) (string, *Error) {
	if strings.ContainsAny(name, `/\`) || strings.ContainsRune(name, 0) {
		return "", newErr(http.StatusBadRequest, "INVALID_FILE_NAME", "File name must not contain path separators")
	}
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", newErr(http.StatusBadRequest, "INVALID_FILE_NAME", "File name is required")
	}
	if trimmed == "." || trimmed == ".." {
		return "", newErr(http.StatusBadRequest, "INVALID_FILE_NAME", "File name is not valid")
	}
	return DisplayName(trimmed), nil
}

// DisplayName cleans a name for storing in the database and showing: base
// name only, no control characters or quotes, at most 200 characters.
func DisplayName(name string) string {
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
		return "file"
	}
	return name
}

// storageSafe keeps [A-Za-z0-9._-] (others become "_"), at most n
// characters, never starting with a dot.
func storageSafe(s string, n int) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	out := strings.TrimLeft(b.String(), ".")
	if len(out) > n {
		out = out[:n]
	}
	return out
}

// plainExtension reports whether ext is a dot followed by 1 to 10 ASCII
// letters or digits (".pdf", ".docx"). Anything else is dropped from stored names.
func plainExtension(ext string) bool {
	if len(ext) < 2 || len(ext) > 11 || ext[0] != '.' {
		return false
	}
	for _, r := range ext[1:] {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9') {
			return false
		}
	}
	return true
}

// StoredName builds a server-generated file name for display: the safe base
// name, a random 12-hex suffix and the lower-cased extension, for example
// "Laporan_Audit-3f2a9c1b7d4e.pdf".
func StoredName(display string) string {
	ext := strings.ToLower(filepath.Ext(display))
	if !plainExtension(ext) {
		ext = ""
	}
	base := storageSafe(strings.TrimSuffix(display, filepath.Ext(display)), 80)
	if base == "" {
		base = "file"
	}
	var rnd [6]byte
	if _, err := rand.Read(rnd[:]); err != nil {
		panic(err)
	}
	return base + "-" + hex.EncodeToString(rnd[:]) + ext
}

// CheckExtension returns the lower-cased extension of name if allowed.
func CheckExtension(name string, allowed map[string]bool) (string, *Error) {
	ext := strings.ToLower(filepath.Ext(name))
	if !allowed[ext] {
		return "", newErr(http.StatusUnsupportedMediaType, "FILE_TYPE_NOT_ALLOWED", "File type not allowed. Allowed: "+AllowedList)
	}
	return ext, nil
}

var (
	magicOLE2 = []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}
	magicZIP  = []byte("PK\x03\x04")
	magicRTF  = []byte(`{\rtf`)
)

// CheckContent verifies that data looks like ext. Office files may be either
// the legacy OLE2 container or OOXML (zip), because renamed files are common;
// .doc may also be RTF. Text files must sniff as plain text (not HTML or XML).
func CheckContent(ext string, data []byte) *Error {
	if len(data) == 0 {
		return newErr(http.StatusBadRequest, "FILE_EMPTY", "The file is empty")
	}
	head := data
	if len(head) > 1024 {
		head = head[:1024]
	}
	sniffed := http.DetectContentType(data)
	ok := false
	switch ext {
	case ".pdf":
		// Readers accept a few bytes of junk before the header.
		ok = bytes.Contains(head, []byte("%PDF-"))
	case ".png":
		ok = sniffed == "image/png"
	case ".jpg", ".jpeg":
		ok = sniffed == "image/jpeg"
	case ".gif":
		ok = sniffed == "image/gif"
	case ".webp":
		ok = sniffed == "image/webp"
	case ".docx", ".xlsx", ".pptx":
		ok = bytes.HasPrefix(data, magicZIP)
	case ".zip":
		ok = bytes.HasPrefix(data, magicZIP) || bytes.HasPrefix(data, []byte("PK\x05\x06"))
	case ".doc":
		ok = bytes.HasPrefix(data, magicOLE2) || bytes.HasPrefix(data, magicZIP) || bytes.HasPrefix(data, magicRTF)
	case ".xls", ".ppt":
		ok = bytes.HasPrefix(data, magicOLE2) || bytes.HasPrefix(data, magicZIP)
	case ".txt", ".csv":
		ok = strings.HasPrefix(sniffed, "text/plain")
	}
	if !ok {
		return newErr(http.StatusUnsupportedMediaType, "FILE_CONTENT_MISMATCH", fmt.Sprintf("File content does not match its %s extension", ext))
	}
	return nil
}

// Scan runs the malware scanner on data.
func Scan(ctx context.Context, label string, data []byte) *Error {
	err := avscan.Default().CheckBytes(ctx, label, data)
	return scanErr(err)
}

// ScanReader runs the malware scanner on a stream.
func ScanReader(ctx context.Context, label string, r io.Reader) *Error {
	return scanErr(avscan.Default().Check(ctx, label, r))
}

func scanErr(err error) *Error {
	if err == nil {
		return nil
	}
	if avscan.IsInfected(err) {
		return &Error{Status: http.StatusUnprocessableEntity, Code: "MALWARE_DETECTED", Message: "The file was rejected because it contains malware", Err: err}
	}
	return &Error{Status: http.StatusServiceUnavailable, Code: "SCANNER_UNAVAILABLE", Message: "The file could not be checked for malware right now. Please try again later.", Err: err}
}

// Validate runs the name, extension, content and malware checks. clientName
// is the name the client supplied; the cleaned display name is returned.
func Validate(ctx context.Context, clientName string, data []byte, max int64, allowed map[string]bool) (string, *Error) {
	name, e := CheckName(clientName)
	if e != nil {
		return "", e
	}
	ext, e := CheckExtension(name, allowed)
	if e != nil {
		return "", e
	}
	if int64(len(data)) > max {
		return "", ErrTooLarge(max)
	}
	if e := CheckContent(ext, data); e != nil {
		return "", e
	}
	if e := Scan(ctx, name, data); e != nil {
		return "", e
	}
	return name, nil
}

// within reports whether p equals root or lies below it.
func within(root, p string) bool {
	return p == root || strings.HasPrefix(p, root+string(filepath.Separator))
}

// confinedDir returns dir as an absolute path, which must lie under the
// uploads root.
func confinedDir(dir string) (string, error) {
	root, err := filepath.Abs(UploadsRoot)
	if err != nil {
		return "", err
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	if !within(root, abs) {
		return "", fmt.Errorf("upload directory %q is outside the uploads root", dir)
	}
	return abs, nil
}

// Store writes data to dir (which must lie under UploadsRoot) under a
// server-generated name built from display. The file is created with O_EXCL,
// so nothing is ever overwritten. It returns the stored file name and the path
// joined as filepath.Join(dir, name) (the form kept in FilePath columns).
func Store(dir, display string, data []byte) (storedName, path string, err error) {
	if _, err := confinedDir(dir); err != nil {
		return "", "", err
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", "", err
	}
	for range 5 {
		storedName = StoredName(display)
		path = filepath.Join(dir, storedName)
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return "", "", err
		}
		_, werr := f.Write(data)
		cerr := f.Close()
		if werr != nil || cerr != nil {
			_ = os.Remove(path)
			if werr != nil {
				return "", "", werr
			}
			return "", "", cerr
		}
		return storedName, path, nil
	}
	return "", "", errors.New("could not allocate a unique file name")
}

// ResolveStored maps a stored FilePath (as written by Store or older code) to
// an absolute path under UploadsRoot. Paths that leave the root, are not
// regular files or are symlinks out of the root are refused.
func ResolveStored(stored string) (string, error) {
	if stored == "" {
		return "", os.ErrNotExist
	}
	root, err := filepath.Abs(UploadsRoot)
	if err != nil {
		return "", err
	}
	abs, err := filepath.Abs(filepath.FromSlash(stored))
	if err != nil {
		return "", err
	}
	if !within(root, abs) || abs == root {
		return "", safefile.ErrOutsideRoot
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return "", safefile.ErrOutsideRoot
	}
	full, err := safefile.ResolveUnder(root, filepath.ToSlash(rel))
	if err != nil {
		return "", err
	}
	info, err := os.Stat(full)
	if err != nil {
		return "", err
	}
	if !info.Mode().IsRegular() {
		return "", safefile.ErrOutsideRoot
	}
	return full, nil
}

// RemoveStored deletes a stored file, but only when it lies under
// UploadsRoot. Missing files and refused paths are reported, not fatal.
func RemoveStored(stored string) error {
	full, err := ResolveStored(stored)
	if err != nil {
		return err
	}
	return os.Remove(full)
}

// ConfineFolder validates a client-supplied media folder such as
// "Auditsphere/fieldwork/ST-001" and returns it cleaned (forward slashes, no
// empty elements). Absolute paths, backslashes, NUL and "." or ".." elements
// are rejected. Other characters outside [A-Za-z0-9 ._-] become "_". The
// result always starts with "Auditsphere".
func ConfineFolder(folder string) (string, *Error) {
	bad := newErr(http.StatusBadRequest, "INVALID_FOLDER", "Folder must be a relative path without '..' elements")
	if strings.ContainsAny(folder, "\\\x00") {
		return "", bad
	}
	// A leading "/" is tolerated (dropped), as before.
	var parts []string
	for p := range strings.SplitSeq(folder, "/") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if p == "." || p == ".." {
			return "", bad
		}
		var b strings.Builder
		for _, r := range p {
			switch {
			case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-', r == '_', r == ' ':
				b.WriteRune(r)
			default:
				b.WriteByte('_')
			}
		}
		s := strings.TrimLeft(b.String(), ".")
		if s == "" {
			s = "_"
		}
		if len(s) > 100 {
			s = s[:100]
		}
		parts = append(parts, s)
	}
	if len(parts) > 8 {
		return "", newErr(http.StatusBadRequest, "INVALID_FOLDER", "Folder is nested too deeply")
	}
	if len(parts) == 0 || !strings.EqualFold(parts[0], "Auditsphere") {
		parts = append([]string{"Auditsphere"}, parts...)
	}
	parts[0] = "Auditsphere"
	return strings.Join(parts, "/"), nil
}

// headers sets the safe-serving headers and returns the Content-Type.
func headers(c *gin.Context, name string, inline bool) string {
	ct := safefile.ContentTypeFor(name)
	disposition := "attachment"
	if inline && safefile.IsInline(ct) {
		disposition = "inline"
	}
	c.Header("Content-Type", ct)
	c.Header("Content-Disposition", safefile.DispositionHeader(disposition, name))
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Security-Policy", "default-src 'none'; sandbox")
	c.Header("Cache-Control", "private, no-store")
	return ct
}

// Serve sends an uploaded document. Bytes stored in the database are used
// when present, otherwise the file at storedPath (which must lie under
// UploadsRoot). inline asks for in-browser display, which is honoured only for
// PDF, images and plain text; everything else is an attachment.
func Serve(c *gin.Context, name string, content []byte, storedPath string, inline bool) {
	if name == "" {
		name = "document"
	}
	if len(content) > 0 {
		ct := headers(c, name, inline)
		c.Data(http.StatusOK, ct, content)
		return
	}
	// Paths outside the uploads root (legacy rows) get the same 404 as missing files.
	full, err := ResolveStored(storedPath)
	if err != nil {
		response.NotFound(c, "Physical file not found on server")
		return
	}
	f, err := os.Open(full)
	if err != nil {
		response.NotFound(c, "Physical file not found on server")
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		response.InternalServerError(c, "Failed to read file")
		return
	}
	headers(c, name, inline)
	http.ServeContent(c.Writer, c.Request, "", info.ModTime(), f)
}
