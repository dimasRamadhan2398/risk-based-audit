// Package upload applies the same upload checks as audit-service to the QA and
// consulting imports in master-service: a capped body, a file name that is
// never used as a path, an extension allowlist, a magic-byte check, a malware
// scan (pkg/avscan, fail-closed by default) and a server-generated stored name
// written with O_EXCL.
package upload

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"master-service/pkg/avscan"
)

// MaxBytes is the largest file accepted (10 MB, as before).
const MaxBytes int64 = 10 << 20

// Root is the directory uploads are stored under (the service's uploads tree).
var Root = "./uploads"

// Allowed lists the extensions accepted for QA reports and consulting documents.
var Allowed = map[string]bool{
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	".ppt": true, ".pptx": true, ".csv": true, ".txt": true,
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true,
}

// Error is a rejection with the HTTP status and a message for the client.
type Error struct {
	Status  int
	Message string
	Err     error
}

func (e *Error) Error() string { return e.Message }
func (e *Error) Unwrap() error { return e.Err }

func fail(status int, msg string) *Error { return &Error{Status: status, Message: msg} }

// CheckName validates a client file name. Path separators, NUL and "."/".."
// are rejected; the cleaned base name is returned for display.
func CheckName(name string) (string, *Error) {
	if strings.ContainsAny(name, `/\`) || strings.ContainsRune(name, 0) {
		return "", fail(http.StatusBadRequest, "File name must not contain path separators")
	}
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." {
		return "", fail(http.StatusBadRequest, "File name is required")
	}
	var b strings.Builder
	for _, r := range name {
		if unicode.IsControl(r) || r == '"' {
			continue
		}
		b.WriteRune(r)
	}
	out := b.String()
	if len([]rune(out)) > 200 {
		out = string([]rune(out)[:200])
	}
	if out == "" {
		return "", fail(http.StatusBadRequest, "File name is required")
	}
	return out, nil
}

// ExtensionOf returns the lower-cased extension of name if it is allowed.
func ExtensionOf(name string) (string, *Error) {
	ext := strings.ToLower(filepath.Ext(name))
	if !Allowed[ext] {
		return "", fail(http.StatusUnsupportedMediaType, "File type not allowed. Allowed: pdf, doc, docx, xls, xlsx, ppt, pptx, csv, txt, png, jpg, jpeg, gif, webp")
	}
	return ext, nil
}

// CheckContent verifies the first bytes match the extension.
func CheckContent(ext string, data []byte) *Error {
	ole2 := []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}
	zip := []byte("PK\x03\x04")
	ok := false
	switch ext {
	case ".pdf":
		ok = bytes.Contains(data[:min(len(data), 1024)], []byte("%PDF-"))
	case ".png":
		ok = bytes.HasPrefix(data, []byte("\x89PNG\r\n\x1a\n"))
	case ".jpg", ".jpeg":
		ok = bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF})
	case ".gif":
		ok = bytes.HasPrefix(data, []byte("GIF87a")) || bytes.HasPrefix(data, []byte("GIF89a"))
	case ".webp":
		ok = bytes.HasPrefix(data, []byte("RIFF")) && bytes.Contains(data[:min(len(data), 16)], []byte("WEBP"))
	case ".docx", ".xlsx", ".pptx":
		ok = bytes.HasPrefix(data, zip)
	case ".doc", ".xls", ".ppt":
		ok = bytes.HasPrefix(data, ole2) || bytes.HasPrefix(data, zip)
	case ".txt", ".csv":
		ok = !bytes.Contains(bytes.ToLower(data[:min(len(data), 512)]), []byte("<html")) &&
			!bytes.Contains(bytes.ToLower(data[:min(len(data), 512)]), []byte("<script")) &&
			!bytes.Contains(data[:min(len(data), 512)], []byte{0})
	}
	if !ok {
		return fail(http.StatusUnsupportedMediaType, fmt.Sprintf("File content does not match its %s extension", ext))
	}
	return nil
}

// Validate runs name, extension, size, content and malware checks on data
// and returns the cleaned display name.
func Validate(ctx context.Context, clientName string, data []byte) (string, *Error) {
	if len(data) == 0 {
		return "", fail(http.StatusBadRequest, "The file is empty")
	}
	if int64(len(data)) > MaxBytes {
		return "", fail(http.StatusRequestEntityTooLarge, "File size exceeds maximum limit of 10MB")
	}
	name, e := CheckName(clientName)
	if e != nil {
		return "", e
	}
	ext, e := ExtensionOf(name)
	if e != nil {
		return "", e
	}
	if e := CheckContent(ext, data); e != nil {
		return "", e
	}
	err := avscan.Default().CheckBytes(ctx, name, data)
	switch {
	case err == nil:
		return name, nil
	case avscan.IsInfected(err):
		return "", &Error{Status: http.StatusUnprocessableEntity, Message: "The file was rejected because it contains malware", Err: err}
	default:
		return "", &Error{Status: http.StatusServiceUnavailable, Message: "The file could not be checked for malware right now. Please try again later.", Err: err}
	}
}

// Store writes data under Root/<subdir> with a server-generated name and
// returns the path to keep in the database.
func Store(subdir, display string, data []byte) (string, error) {
	dir := filepath.Join(Root, subdir)
	absRoot, err := filepath.Abs(Root)
	if err != nil {
		return "", err
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	if absDir != absRoot && !strings.HasPrefix(absDir, absRoot+string(filepath.Separator)) {
		return "", errors.New("upload directory is outside the uploads root")
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", err
	}
	var rnd [6]byte
	for range 5 {
		if _, err := rand.Read(rnd[:]); err != nil {
			return "", err
		}
		base := safeBase(strings.TrimSuffix(display, filepath.Ext(display)))
		ext := strings.ToLower(filepath.Ext(display))
		name := base + "-" + hex.EncodeToString(rnd[:]) + ext
		path := filepath.Join(dir, name)
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		_, werr := f.Write(data)
		cerr := f.Close()
		if werr != nil || cerr != nil {
			_ = os.Remove(path)
			if werr != nil {
				return "", werr
			}
			return "", cerr
		}
		return path, nil
	}
	return "", errors.New("could not allocate a unique file name")
}

func safeBase(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	out := b.String()
	if len(out) > 80 {
		out = out[:80]
	}
	if out == "" {
		out = "file"
	}
	return out
}

// IsTooLarge reports whether err comes from a body over the Limit cap.
func IsTooLarge(err error) bool {
	var mbe *http.MaxBytesError
	return errors.As(err, &mbe)
}

// Limit caps a request body for a base64-or-multipart upload.
func Limit(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxBytes/3*4+(1<<20))
}
