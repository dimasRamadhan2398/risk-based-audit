package safefile

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupRoot(t *testing.T) (root, outside string) {
	t.Helper()
	base := t.TempDir()
	root = filepath.Join(base, "uploads")
	outside = filepath.Join(base, "secret.txt")
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "plan.pdf"), []byte("%PDF-1.4 test"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "page.html"), []byte("<script>alert(1)</script>"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(outside, []byte("outside the root"), 0o600); err != nil {
		t.Fatal(err)
	}
	return root, outside
}

func TestResolveUnderConfinesTraversal(t *testing.T) {
	root, _ := setupRoot(t)
	rootAbs, _ := filepath.Abs(root)
	for _, rel := range []string{"../secret.txt", "docs/../../secret.txt", "/../../secret.txt", `..\..\secret.txt`} {
		got, err := ResolveUnder(root, rel)
		if err == nil && !strings.HasPrefix(got, rootAbs+string(filepath.Separator)) {
			t.Fatalf("ResolveUnder(%q) escaped the root: %s", rel, got)
		}
	}
	if _, err := ResolveUnder(root, "../secret.txt"); err == nil {
		t.Fatal("a file that exists only outside the root must not resolve")
	}
	if got, err := ResolveUnder(root, "/docs/plan.pdf"); err != nil || !strings.HasSuffix(got, filepath.Join("docs", "plan.pdf")) {
		t.Fatalf("valid path: got %q, %v", got, err)
	}
}

func TestResolveUnderRefusesSymlinkOutsideRoot(t *testing.T) {
	root, outside := setupRoot(t)
	if err := os.Symlink(outside, filepath.Join(root, "link.txt")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := ResolveUnder(root, "/link.txt"); err == nil {
		t.Fatal("a symlink pointing outside the root must be refused")
	}
}

func TestContentTypeNeverActive(t *testing.T) {
	for _, name := range []string{"x.html", "x.htm", "x.svg", "x.xhtml", "x.js", "x.xml", "x.mhtml", "x.bin", "noext"} {
		ct := ContentTypeFor(name)
		if ct != "application/octet-stream" {
			t.Fatalf("ContentTypeFor(%q) = %q, want application/octet-stream", name, ct)
		}
	}
	if ct := ContentTypeFor("report.PDF"); ct != "application/pdf" {
		t.Fatalf("pdf: got %q", ct)
	}
	if IsInline(ContentTypeFor("x.svg")) || IsInline(ContentTypeFor("x.html")) {
		t.Fatal("html and svg must never be inline")
	}
}

func TestIsPrivatePath(t *testing.T) {
	private := []string{"/action-taken-reports/a.pdf", "action-taken-reports/a.pdf", "/Action-Taken-Reports/a.pdf", "/docs/../action-taken-reports/a.pdf", `\action-taken-reports\a.pdf`}
	for _, p := range private {
		if !IsPrivatePath(p) {
			t.Fatalf("IsPrivatePath(%q) = false, want true", p)
		}
	}
	if IsPrivatePath("/docs/plan.pdf") {
		t.Fatal("ordinary path must not be private")
	}
	// Feature document stores are private by prefix, in any case.
	for _, p := range []string{"uploaded-assignment-letters/a.pdf", "/Uploaded-Plan-Documents/b.docx", "imported-working-papers/c.xlsx"} {
		if !IsPrivatePath(p) {
			t.Fatalf("IsPrivatePath(%q) = false, want true", p)
		}
	}
	// Media uploads are not private.
	for _, p := range []string{"Auditsphere/fieldwork/x.pdf", "fieldwork/x.pdf", "Auditsphere/uploaded-x/y.pdf"} {
		if IsPrivatePath(p) {
			t.Fatalf("IsPrivatePath(%q) = true, want false", p)
		}
	}
}

func serveOnce(t *testing.T, fullPath, name string, force bool) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/f", func(c *gin.Context) {
		if err := ServeFile(c, fullPath, name, force); err != nil {
			c.Status(http.StatusNotFound)
		}
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/f", nil))
	return w
}

func TestServeFileHTMLIsAttachmentOctetStream(t *testing.T) {
	root, _ := setupRoot(t)
	w := serveOnce(t, filepath.Join(root, "docs", "page.html"), "page.html", false)
	if ct := w.Header().Get("Content-Type"); ct != "application/octet-stream" {
		t.Fatalf("Content-Type = %q, want application/octet-stream", ct)
	}
	if cd := w.Header().Get("Content-Disposition"); !strings.HasPrefix(cd, "attachment") {
		t.Fatalf("Content-Disposition = %q, want attachment", cd)
	}
	if w.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("nosniff header missing")
	}
	if !strings.Contains(w.Header().Get("Content-Security-Policy"), "sandbox") {
		t.Fatal("sandbox CSP missing")
	}
}

func TestServeFilePDFInlineUnlessDownloadRequested(t *testing.T) {
	root, _ := setupRoot(t)
	full := filepath.Join(root, "docs", "plan.pdf")
	if cd := serveOnce(t, full, "plan.pdf", false).Header().Get("Content-Disposition"); !strings.HasPrefix(cd, "inline") {
		t.Fatalf("pdf view: Content-Disposition = %q, want inline", cd)
	}
	if cd := serveOnce(t, full, "plan.pdf", true).Header().Get("Content-Disposition"); !strings.HasPrefix(cd, "attachment") {
		t.Fatalf("pdf download: Content-Disposition = %q, want attachment", cd)
	}
}

func TestServeFileRefusesDirectory(t *testing.T) {
	root, _ := setupRoot(t)
	if w := serveOnce(t, filepath.Join(root, "docs"), "docs", false); w.Code != http.StatusNotFound {
		t.Fatalf("directory: got %d, want 404", w.Code)
	}
}
