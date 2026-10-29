package media

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"audit-service/pkg/avscan"
	pkgmedia "audit-service/pkg/media"
	"audit-service/pkg/upload"
	"audit-service/services/media"

	"github.com/gin-gonic/gin"
)

var pdfBody = []byte("%PDF-1.4\n%%EOF")

func newLocal() pkgmedia.MediaProvider { return pkgmedia.NewLocalProvider() }

// scannerAddrFromEnv returns the live clamd address, or skips the test when
// CLAMAV_TEST_ADDR is not set (for example 127.0.0.1:13310).
func scannerAddrFromEnv(t *testing.T) string {
	t.Helper()
	addr := os.Getenv("CLAMAV_TEST_ADDR")
	if addr == "" {
		t.Skip("CLAMAV_TEST_ADDR not set; skipping live clamd upload test")
	}
	return addr
}

// uploadRouter wires the real media service over the local provider, with
// uploads written under a temp root.
func uploadRouter(t *testing.T, guard *avscan.Guard) (*gin.Engine, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	root := t.TempDir()
	oldRoot := upload.UploadsRoot
	upload.UploadsRoot = root
	t.Cleanup(func() { upload.UploadsRoot = oldRoot })
	restore := avscan.SetDefault(guard)
	t.Cleanup(restore)

	ctrl := NewMediaController(media.NewMediaService(newLocal()))
	r := gin.New()
	r.POST("/api/v1/media/upload", ctrl.Upload)
	return r, root
}

func postUpload(t *testing.T, r *gin.Engine, fields map[string]string, filename string, data []byte) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		_ = w.WriteField(k, v)
	}
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	part.Write(data)
	w.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/media/upload", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func errorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	return body.Error.Code
}

func TestUploadRejectsTraversalFolder(t *testing.T) {
	r, root := uploadRouter(t, &avscan.Guard{Mode: avscan.ModeOff})
	for _, folder := range []string{"Auditsphere/../../x", "../../etc", `Auditsphere\..\..\x`} {
		rec := postUpload(t, r, map[string]string{"folder": folder}, "plan.pdf", pdfBody)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("folder %q: status %d, want 400", folder, rec.Code)
		}
	}
	// Nothing may have been written above the root.
	if _, err := os.Stat(filepath.Join(filepath.Dir(root), "x")); err == nil {
		t.Fatal("traversal wrote outside the uploads root")
	}
}

func TestUploadRejectsTraversalFileName(t *testing.T) {
	r, root := uploadRouter(t, &avscan.Guard{Mode: avscan.ModeOff})
	// net/http reduces a multipart file name to its base name, so the client
	// name cannot carry a path. The file must land inside the root under a
	// server-generated name, never at the traversal target.
	rec := postUpload(t, r, map[string]string{"folder": "Auditsphere/sop"}, "../../../evil.pdf", pdfBody)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(root), "evil.pdf")); err == nil {
		t.Fatal("file written outside the uploads root")
	}
	files := listFiles(root)
	if len(files) != 1 || strings.Contains(files[0], "..") || !strings.HasPrefix(filepath.Base(files[0]), "evil-") {
		t.Fatalf("stored files %v", files)
	}
}

func TestUploadRejectsDisallowedExtension(t *testing.T) {
	r, _ := uploadRouter(t, &avscan.Guard{Mode: avscan.ModeOff})
	rec := postUpload(t, r, map[string]string{"folder": "Auditsphere/sop"}, "page.html", []byte("<script>alert(1)</script>"))
	if rec.Code != http.StatusUnsupportedMediaType || errorCode(t, rec) != "FILE_TYPE_NOT_ALLOWED" {
		t.Fatalf("html upload: status %d code %q", rec.Code, errorCode(t, rec))
	}
}

func TestUploadRejectsOversizeBody(t *testing.T) {
	r, _ := uploadRouter(t, &avscan.Guard{Mode: avscan.ModeOff})
	big := bytes.Repeat([]byte("a"), int(upload.DefaultMaxBytes)+1024)
	rec := postUpload(t, r, map[string]string{"folder": "Auditsphere/sop"}, "big.txt", big)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversize upload: status %d, want 413", rec.Code)
	}
}

func TestUploadRejectsEICAR(t *testing.T) {
	addr := scannerAddrFromEnv(t)
	r, root := uploadRouter(t, &avscan.Guard{Scanner: &avscan.Clamd{Address: addr}, Mode: avscan.ModeRequired, Logf: func(string, ...any) {}})
	eicar := []byte("X5O!P%@AP[4\\PZX54(P^)7CC)7}$" + "EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*")
	rec := postUpload(t, r, map[string]string{"folder": "Auditsphere/sop"}, "eicar.txt", eicar)
	if rec.Code != http.StatusUnprocessableEntity || errorCode(t, rec) != "MALWARE_DETECTED" {
		t.Fatalf("EICAR: status %d code %q body %s", rec.Code, errorCode(t, rec), rec.Body.String())
	}
	if files := listFiles(root); len(files) != 0 {
		t.Fatalf("infected file left on disk: %v", files)
	}
}

func TestUploadAcceptsNormalPDF(t *testing.T) {
	addr := scannerAddrFromEnv(t)
	r, root := uploadRouter(t, &avscan.Guard{Scanner: &avscan.Clamd{Address: addr}, Mode: avscan.ModeRequired, Logf: func(string, ...any) {}})
	rec := postUpload(t, r, map[string]string{"folder": "Auditsphere/sop"}, "SOP Audit 2026.pdf", pdfBody)
	if rec.Code != http.StatusOK {
		t.Fatalf("pdf: status %d body %s", rec.Code, rec.Body.String())
	}
	files := listFiles(root)
	if len(files) != 1 || !strings.HasSuffix(files[0], ".pdf") || !strings.Contains(files[0], "Auditsphere") {
		t.Fatalf("stored files %v", files)
	}
}

func listFiles(root string) []string {
	var out []string
	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			out = append(out, p)
		}
		return nil
	})
	return out
}
