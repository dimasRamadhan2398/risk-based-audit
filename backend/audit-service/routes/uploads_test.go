package routes

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"audit-service/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// uploadsTestSecret is a synthetic test value, never a real key.
var uploadsTestSecret = "uploads-test-" + strings.Repeat("u", 64)

func uploadsTestToken(t *testing.T) string {
	t.Helper()
	claims := &middleware.Claims{
		UserID: "user-1", Username: "tester", Roles: []string{"VIEWER"},
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(uploadsTestSecret))
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func uploadsTestRouter(t *testing.T) (*gin.Engine, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	base := t.TempDir()
	root := filepath.Join(base, "uploads")
	for dir, name := range map[string]string{
		"docs":                 "plan.pdf",
		"":                     "root.txt",
		"action-taken-reports": "evidence.pdf",
	} {
		d := filepath.Join(root, dir)
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(d, name), []byte("content"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "page.html"), []byte("<script>x</script>"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "secret.txt"), []byte("outside"), 0o600); err != nil {
		t.Fatal(err)
	}

	engine := gin.New()
	auth := middleware.NewAuthMiddleware(uploadsTestSecret)
	RegisterUploads(engine, root, auth.Authenticate())
	return engine, root
}

func getUpload(r *gin.Engine, target, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestUploadsRequireValidToken(t *testing.T) {
	r, _ := uploadsTestRouter(t)
	if w := getUpload(r, "/uploads/docs/plan.pdf", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous request: got %d, want 401", w.Code)
	}
	if w := getUpload(r, "/uploads/docs/plan.pdf?token="+uploadsTestToken(t), ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("token in query string: got %d, want 401", w.Code)
	}
	if w := getUpload(r, "/uploads/docs/plan.pdf", uploadsTestToken(t)); w.Code != http.StatusOK {
		t.Fatalf("authenticated request: got %d, want 200", w.Code)
	}
}

func TestUploadsServeActiveContentAsAttachment(t *testing.T) {
	r, _ := uploadsTestRouter(t)
	tok := uploadsTestToken(t)

	w := getUpload(r, "/uploads/docs/page.html", tok)
	if w.Code != http.StatusOK {
		t.Fatalf("html: got %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/octet-stream" {
		t.Fatalf("html Content-Type = %q, want application/octet-stream", ct)
	}
	if cd := w.Header().Get("Content-Disposition"); !strings.HasPrefix(cd, "attachment") {
		t.Fatalf("html Content-Disposition = %q, want attachment", cd)
	}
}

func TestUploadsPrivateDirectoryIsNotServed(t *testing.T) {
	r, _ := uploadsTestRouter(t)
	tok := uploadsTestToken(t)
	for _, target := range []string{
		"/uploads/action-taken-reports/evidence.pdf",
		"/uploads/Action-Taken-Reports/evidence.pdf",
		"/uploads/docs/../action-taken-reports/evidence.pdf",
	} {
		if w := getUpload(r, target, tok); w.Code != http.StatusNotFound {
			t.Fatalf("%s: got %d, want 404", target, w.Code)
		}
	}
}

// Feature document stores are private by default: /uploads refuses them even
// to an authenticated caller who knows the stored name.
func TestUploadsFeatureDocumentsAreNotServed(t *testing.T) {
	r, root := uploadsTestRouter(t)
	tok := uploadsTestToken(t)
	for dir, name := range map[string]string{
		"uploaded-assignment-letters": "letter-abc123.pdf",
		"imported-working-papers":     "paper-ghi789.xlsx",
	} {
		d := filepath.Join(root, dir)
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(d, name), []byte("feature-body"), 0o600); err != nil {
			t.Fatal(err)
		}
		target := "/uploads/" + dir + "/" + name
		if w := getUpload(r, target, tok); w.Code != http.StatusNotFound {
			t.Fatalf("%s: got %d, want 404", target, w.Code)
		}
	}
}

func TestUploadsTraversalStaysInsideRoot(t *testing.T) {
	r, _ := uploadsTestRouter(t)
	tok := uploadsTestToken(t)
	for _, target := range []string{"/uploads/../secret.txt", "/uploads/%2e%2e/secret.txt", "/uploads/docs/../../secret.txt"} {
		w := getUpload(r, target, tok)
		if strings.Contains(w.Body.String(), "outside") {
			t.Fatalf("%s returned a file outside the uploads root", target)
		}
	}
}

func TestUploadsDirectoryIsNotListed(t *testing.T) {
	r, _ := uploadsTestRouter(t)
	if w := getUpload(r, "/uploads/docs", uploadsTestToken(t)); w.Code != http.StatusNotFound {
		t.Fatalf("directory: got %d, want 404", w.Code)
	}
}
