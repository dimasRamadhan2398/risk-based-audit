package media

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// setupMediaRoot points downloads at a temp uploads directory that has a file
// outside it, so a traversal would be visible in the response body.
func setupMediaRoot(t *testing.T) {
	t.Helper()
	base := t.TempDir()
	root := filepath.Join(base, "uploads")
	write := func(p, body string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(root, "docs", "plan.pdf"), "pdf-body")
	write(filepath.Join(root, "docs", "page.html"), "<script>x</script>")
	write(filepath.Join(root, "action-taken-reports", "evidence.pdf"), "evidence-body")
	write(filepath.Join(base, "secret.txt"), "outside-body")

	old := uploadsRoot
	uploadsRoot = root
	t.Cleanup(func() { uploadsRoot = old })
}

func mediaRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	ctrl := NewMediaController(nil)
	r := gin.New()
	r.GET("/api/v1/media/download/:id", ctrl.Download)
	return r
}

func download(r *gin.Engine, target string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, target, nil))
	return w
}

func TestMediaDownloadServesFileInsideRoot(t *testing.T) {
	setupMediaRoot(t)
	w := download(mediaRouter(), "/api/v1/media/download/plan.pdf")
	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", w.Code)
	}
	if cd := w.Header().Get("Content-Disposition"); !strings.HasPrefix(cd, "inline") {
		t.Fatalf("pdf view: Content-Disposition = %q, want inline", cd)
	}
	if w := download(mediaRouter(), "/api/v1/media/download/plan.pdf?download=true"); !strings.HasPrefix(w.Header().Get("Content-Disposition"), "attachment") {
		t.Fatal("download=true must send an attachment")
	}
}

func TestMediaDownloadNeverReadsOutsideRoot(t *testing.T) {
	setupMediaRoot(t)
	r := mediaRouter()
	for _, target := range []string{
		"/api/v1/media/download/x?path=uploads/../../secret.txt",
		"/api/v1/media/download/x?path=../secret.txt",
		"/api/v1/media/download/..%2F..%2Fsecret.txt",
		"/api/v1/media/download/secret.txt?path=/../../secret.txt",
	} {
		w := download(r, target)
		if strings.Contains(w.Body.String(), "outside-body") {
			t.Fatalf("%s leaked a file outside the uploads root", target)
		}
	}
}

func TestMediaDownloadRefusesPrivateDirectory(t *testing.T) {
	setupMediaRoot(t)
	r := mediaRouter()
	for _, target := range []string{
		"/api/v1/media/download/evidence.pdf",
		"/api/v1/media/download/x?path=action-taken-reports/evidence.pdf",
	} {
		if w := download(r, target); strings.Contains(w.Body.String(), "evidence-body") {
			t.Fatalf("%s served a private ATR file", target)
		}
	}
}

func TestMediaDownloadActiveContentIsAttachment(t *testing.T) {
	setupMediaRoot(t)
	w := download(mediaRouter(), "/api/v1/media/download/page.html")
	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/octet-stream" {
		t.Fatalf("html Content-Type = %q, want application/octet-stream", ct)
	}
	if cd := w.Header().Get("Content-Disposition"); !strings.HasPrefix(cd, "attachment") {
		t.Fatalf("html Content-Disposition = %q, want attachment", cd)
	}
}
