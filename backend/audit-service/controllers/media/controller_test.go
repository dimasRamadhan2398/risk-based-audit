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
	write(filepath.Join(root, "uploaded-assignment-letters", "letter-abc123.pdf"), "letter-body")
	write(filepath.Join(root, "imported-working-papers", "paper-ghi789.xlsx"), "paper-body")
	write(filepath.Join(root, "Auditsphere", "fieldwork", "evidence-def456.pdf"), "fieldwork-body")
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

// Feature document stores (uploaded-*, imported-*) are private: the generic
// route refuses them whether the caller names the stored file or its path.
// Their files are served only by the owning feature's endpoints.
func TestMediaDownloadRefusesFeatureDocuments(t *testing.T) {
	setupMediaRoot(t)
	r := mediaRouter()
	for _, target := range []string{
		"/api/v1/media/download/letter-abc123.pdf",
		"/api/v1/media/download/letter-abc123.pdf?download=true",
		"/api/v1/media/download/x?path=uploaded-assignment-letters/letter-abc123.pdf",
		"/api/v1/media/download/x?path=/Uploaded-Assignment-Letters/letter-abc123.pdf",
		"/api/v1/media/download/uploaded-assignment-letters%2Fletter-abc123.pdf",
		"/api/v1/media/download/paper-ghi789.xlsx",
		"/api/v1/media/download/x?path=imported-working-papers/paper-ghi789.xlsx",
	} {
		w := download(r, target)
		if w.Code != http.StatusNotFound {
			t.Fatalf("%s: got %d, want 404", target, w.Code)
		}
		if body := w.Body.String(); strings.Contains(body, "letter-body") || strings.Contains(body, "paper-body") {
			t.Fatalf("%s served a feature document", target)
		}
	}
}

// An authorised media file (a stored name in a sub-folder, as the fieldwork
// store uploads it) is still served by bare name.
func TestMediaDownloadStillServesMediaUploads(t *testing.T) {
	setupMediaRoot(t)
	w := download(mediaRouter(), "/api/v1/media/download/evidence-def456.pdf")
	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", w.Code)
	}
	if body := w.Body.String(); body != "fieldwork-body" {
		t.Fatalf("body = %q, want fieldwork-body", body)
	}
}

// An explicit path that does not resolve is a 404. A file with the same base
// name in another folder (docs/plan.pdf here) must not be returned in its place.
func TestMediaDownloadExplicitPathNeverFallsBackToBaseName(t *testing.T) {
	setupMediaRoot(t)
	r := mediaRouter()
	for _, target := range []string{
		"/api/v1/media/download/plan.pdf?path=reports/plan.pdf",
		"/api/v1/media/download/plan.pdf?path=uploads/reports/plan.pdf",
		"/api/v1/media/download/plan.pdf?filepath=/reports/plan.pdf",
		"/api/v1/media/download/plan.pdf?path=plan.pdf",
	} {
		w := download(r, target)
		if w.Code != http.StatusNotFound {
			t.Fatalf("%s: got %d, want 404", target, w.Code)
		}
		if strings.Contains(w.Body.String(), "pdf-body") {
			t.Fatalf("%s returned docs/plan.pdf in place of the requested path", target)
		}
	}
}
