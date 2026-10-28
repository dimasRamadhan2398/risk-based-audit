package routes

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"audit-service/models"
	"audit-service/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

const viewerTestSecret = "viewer-test-secret-viewer-test-secret-0123456789"

func viewerTestToken(t *testing.T, roles ...string) string {
	t.Helper()
	claims := &middleware.Claims{
		UserID: "user-1", Username: "tester", Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(viewerTestSecret))
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

// viewerTestEngine registers the real route table (plus /uploads) over an
// in-memory database. Recovery turns the nil-controller stubs into 500s, so a
// request that gets past the viewer guard shows up as "not 403".
func viewerTestEngine(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`CREATE TABLE strategic_plans (id TEXT PRIMARY KEY, created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`).Error; err != nil {
		t.Fatalf("create table: %v", err)
	}
	_ = models.StrategicPlan{}

	engine := gin.New()
	engine.Use(gin.RecoveryWithWriter(io.Discard))
	auth := middleware.NewAuthMiddleware(viewerTestSecret)

	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "plan.pdf"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	RegisterUploads(engine, root, auth.Authenticate())

	h := NewRouteHandler(engine, auth, db, nil)
	registry := NewRouteRegistry()
	registry.SetAuditCharterController(stubCharterCtrl{})
	registry.SetAuditMandateController(stubMandateCtrl{})
	registry.SetAuditAssignmentController(stubAssignmentCtrl{})
	registry.SetAuditActivityController(stubActivityCtrl{})
	h.SetRegistry(registry)
	h.RegisterRoutes()
	return engine
}

func doViewerReq(engine *gin.Engine, method, path, token string) int {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w.Code
}

func TestViewerRouteMatrix(t *testing.T) {
	engine := viewerTestEngine(t)
	viewer := viewerTestToken(t, "VIEWER")
	viewerLower := viewerTestToken(t, "viewer")
	admin := viewerTestToken(t, "ADMIN")
	auditor := viewerTestToken(t, "AUDITOR")
	mixed := viewerTestToken(t, "VIEWER", "AUDITOR")

	const id = "00000000-0000-0000-0000-000000000001"
	// Reads that stay open for a viewer: they must not be 403/401.
	reads := []string{
		"/api/v1/strategic-plans",
		"/api/v1/audit-charters/active",
		"/api/v1/annual-audit-plans/" + id,
		"/api/v1/assignment-letters",
		"/api/v1/action-taken-reports",
		"/api/v1/performance/kpi",
		"/api/v1/uploaded-annual-plans", // metadata list only, no file bytes
	}
	for _, p := range reads {
		if c := doViewerReq(engine, "GET", p, viewer); c == 403 || c == 401 {
			t.Errorf("viewer GET %s = %d, want it to pass the guard", p, c)
		}
	}
	if c := doViewerReq(engine, "GET", "/api/v1/strategic-plans", viewer); c != http.StatusOK {
		t.Errorf("viewer GET /api/v1/strategic-plans = %d, want 200", c)
	}

	deny := []struct{ method, path string }{
		// writes
		{"POST", "/api/v1/strategic-plans"},
		{"PUT", "/api/v1/strategic-plans/" + id},
		{"DELETE", "/api/v1/strategic-plans/" + id},
		{"POST", "/api/v1/audit-charters"},
		{"POST", "/api/v1/audit-charters/" + id + "/activate"},
		{"PUT", "/api/v1/audit-assignments/" + id + "/status"},
		{"POST", "/api/v1/action-taken-reports/" + id + "/submit"},
		{"POST", "/api/v1/action-taken-reports/" + id + "/review"},
		{"POST", "/api/v1/action-taken-reports/" + id + "/evidence"},
		{"POST", "/api/v1/uploaded-annual-plans"},
		{"POST", "/api/v1/working-papers/imports"},
		{"POST", "/api/v1/media/upload"},
		{"DELETE", "/api/v1/performance/kpi/" + id},
		// file access
		{"GET", "/api/v1/uploaded-annual-plans/" + id + "/download"},
		{"GET", "/api/v1/uploaded-annual-plans/" + id + "/view"},
		{"GET", "/api/v1/uploaded-annual-plans/" + id + "/view/a.pdf"},
		{"GET", "/api/v1/uploaded-audit-result-reports/" + id + "/download"},
		{"GET", "/api/v1/working-papers/imports/" + id + "/download"},
		{"GET", "/api/v1/working-papers/imports/" + id + "/view"},
		{"GET", "/api/v1/audit-charters/" + id + "/download"},
		{"GET", "/api/v1/action-taken-reports/" + id + "/evidence/" + id},
		{"GET", "/api/v1/media/download/" + id},
		{"GET", "/uploads/plan.pdf"},
		{"HEAD", "/uploads/plan.pdf"},
		// exports
		{"GET", "/api/v1/performance/export-pdf"},
		{"GET", "/api/v1/audit-result-reports/" + id + "/download-docx"},
	}
	for _, tc := range deny {
		for name, tok := range map[string]string{"VIEWER": viewer, "viewer": viewerLower} {
			if c := doViewerReq(engine, tc.method, tc.path, tok); c != http.StatusForbidden {
				t.Errorf("%s %s %s = %d, want 403", name, tc.method, tc.path, c)
			}
		}
		for name, tok := range map[string]string{"ADMIN": admin, "AUDITOR": auditor, "VIEWER+AUDITOR": mixed} {
			if c := doViewerReq(engine, tc.method, tc.path, tok); c == http.StatusForbidden || c == http.StatusUnauthorized {
				t.Errorf("%s %s %s = %d, want it to pass the viewer guard", name, tc.method, tc.path, c)
			}
		}
	}
}

// Deny-by-default: every registered route is checked, so a route added later
// without thought for VIEWER fails here instead of shipping open.
func TestViewerEveryRegisteredRoute(t *testing.T) {
	engine := viewerTestEngine(t)
	viewer := viewerTestToken(t, "VIEWER")
	for _, r := range engine.Routes() {
		if !strings.HasPrefix(r.Path, "/api/v1/") && !strings.HasPrefix(r.Path, "/uploads") {
			continue
		}
		path := strings.NewReplacer(":id", "00000000-0000-0000-0000-000000000001", ":filename", "a.pdf", "*filepath", "/plan.pdf").Replace(r.Path)
		for _, seg := range strings.Split(r.Path, "/") {
			if strings.HasPrefix(seg, ":") {
				path = strings.ReplaceAll(path, seg, "x")
			}
		}
		c := doViewerReq(engine, r.Method, path, viewer)
		wantDenied := (r.Method != "GET" && r.Method != "HEAD") || middleware.IsFileOrExportRoute(r.Path)
		if wantDenied && c != http.StatusForbidden {
			t.Errorf("viewer %s %s = %d, want 403", r.Method, r.Path, c)
		}
		if !wantDenied && c == http.StatusForbidden {
			t.Errorf("viewer %s %s = 403, but it is a plain read", r.Method, r.Path)
		}
	}
}
