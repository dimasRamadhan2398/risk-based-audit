// Package safefile serves uploaded files without trusting their names or types.
//
// Paths are confined to a root directory, Content-Type comes from an allowlist
// built from the file extension (never from the client), and anything that a
// browser could render as active content is sent as an attachment.
package safefile

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// ErrOutsideRoot is returned when a requested path resolves outside its root.
var ErrOutsideRoot = errors.New("path is outside the uploads root")

// PrivateDirs are subdirectories of the uploads root that are never served by
// the generic file routes. Their files are reachable only through authorised
// endpoints (for example ATR evidence).
var PrivateDirs = []string{"action-taken-reports"}

const octetStream = "application/octet-stream"

// extTypes maps a file extension to the only Content-Type the service will send.
// Anything not listed is served as application/octet-stream.
var extTypes = map[string]string{
	".pdf":  "application/pdf",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".gif":  "image/gif",
	".webp": "image/webp",
	".txt":  "text/plain; charset=utf-8",
	".csv":  "text/csv; charset=utf-8",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".ppt":  "application/vnd.ms-powerpoint",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
}

// inlineTypes may be displayed in the browser. Everything else is downloaded.
var inlineTypes = map[string]bool{
	"application/pdf":           true,
	"image/png":                 true,
	"image/jpeg":                true,
	"image/gif":                 true,
	"image/webp":                true,
	"text/plain; charset=utf-8": true,
}

// ContentTypeFor returns the Content-Type to send for a file name. HTML, SVG,
// XML and script types are never returned, whatever the upload claimed.
func ContentTypeFor(fileName string) string {
	if ct, ok := extTypes[strings.ToLower(path.Ext(fileName))]; ok {
		return ct
	}
	return octetStream
}

// IsInline reports whether a Content-Type may be shown inline.
func IsInline(contentType string) bool {
	return inlineTypes[contentType]
}

// IsPrivatePath reports whether a path (relative to the uploads root) falls in a
// private subdirectory. The comparison is case-insensitive and runs after
// cleaning, so "..", doubled slashes and backslashes do not get around it.
func IsPrivatePath(rel string) bool {
	cleaned := strings.ToLower(path.Clean("/" + strings.ReplaceAll(rel, `\`, "/")))
	for _, dir := range PrivateDirs {
		d := "/" + strings.ToLower(dir)
		if cleaned == d || strings.HasPrefix(cleaned, d+"/") {
			return true
		}
	}
	return false
}

// ResolveUnder maps a relative path onto root and returns the absolute path.
// The path is cleaned as if it were rooted, so ".." cannot climb above root.
// Symlinks that lead outside root are refused.
func ResolveUnder(root, rel string) (string, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	cleaned := path.Clean("/" + strings.ReplaceAll(rel, `\`, "/"))
	full := filepath.Join(rootAbs, filepath.FromSlash(cleaned))
	if !within(rootAbs, full) {
		return "", ErrOutsideRoot
	}
	resolved, err := filepath.EvalSymlinks(full)
	if err != nil {
		return "", err
	}
	rootReal, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return "", err
	}
	if !within(rootReal, resolved) {
		return "", ErrOutsideRoot
	}
	return resolved, nil
}

func within(root, p string) bool {
	return p == root || strings.HasPrefix(p, root+string(filepath.Separator))
}

// ServeFile streams a file that has already been resolved and authorised. The
// file is sent as inline only when its type is safe to display and the caller
// did not ask for a download. Directories are refused.
func ServeFile(c *gin.Context, fullPath, displayName string, forceAttachment bool) error {
	f, err := os.Open(fullPath)
	if err != nil {
		return err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory", displayName)
	}

	contentType := ContentTypeFor(displayName)
	disposition := "attachment"
	if !forceAttachment && IsInline(contentType) {
		disposition = "inline"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", dispositionHeader(disposition, displayName))
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Security-Policy", "default-src 'none'; sandbox")
	// Audit evidence is tenant-scoped and access-controlled: never cache it in a shared cache.
	c.Header("Cache-Control", "private, max-age=300")
	c.Header("Vary", "Authorization")
	http.ServeContent(c.Writer, c.Request, displayName, info.ModTime(), f)
	return nil
}

// DispositionHeader builds a Content-Disposition value ("inline" or
// "attachment") with an ASCII fallback filename plus an RFC 5987 UTF-8 form.
func DispositionHeader(disposition, name string) string {
	return dispositionHeader(disposition, name)
}

// dispositionHeader builds an ASCII fallback filename plus an RFC 5987 UTF-8 form.
func dispositionHeader(disposition, name string) string {
	var b strings.Builder
	for _, r := range name {
		switch {
		case r < 0x20 || r > 0x7e || r == '"' || r == '\\':
			b.WriteRune('_')
		default:
			b.WriteRune(r)
		}
	}
	return fmt.Sprintf(`%s; filename="%s"; filename*=UTF-8''%s`, disposition, b.String(), url.PathEscape(name))
}
