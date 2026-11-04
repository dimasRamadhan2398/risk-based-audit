package routes

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"master-service/controllers"
	"master-service/pkg/middleware"
	"master-service/pkg/validations"
	"master-service/services"

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

func viewerTestEngine(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	engine := gin.New()
	engine.Use(gin.RecoveryWithWriter(io.Discard))
	ctrl := controllers.NewControllerRegistry(services.NewServiceRegistry(db, nil, nil, services.OrgOptions{}), validations.New())
	RegisterRoutes(engine, ctrl, db, middleware.NewAuthMiddleware(viewerTestSecret), nil, OrgRouteOptions{})
	return engine
}

func doViewerReq(engine *gin.Engine, method, path, token string) int {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w.Code
}

func TestViewerRefusedOnMasterData(t *testing.T) {
	engine := viewerTestEngine(t)
	viewer := viewerTestToken(t, "VIEWER")
	viewerLower := viewerTestToken(t, "viewer")
	admin := viewerTestToken(t, "ADMIN")
	auditor := viewerTestToken(t, "AUDITOR")
	mixed := viewerTestToken(t, "VIEWER", "AUDITOR")

	cases := []struct{ method, path string }{
		{"GET", "/api/v1/companies"},
		{"GET", "/api/v1/departments"},
		{"GET", "/api/v1/employees"},
		{"GET", "/api/v1/employees/by-code/E1"},
		{"GET", "/api/v1/locations"},
		{"GET", "/api/v1/business-units"},
		{"GET", "/api/v1/job-roles"},
		{"GET", "/api/v1/vision-mission-goals"},
		{"GET", "/api/v1/quality-assurance"},
		{"GET", "/api/v1/quality-assurance/x/download"},
		{"GET", "/api/v1/quality-assurance/x/view"},
		{"GET", "/api/v1/consulting-services"},
		{"GET", "/api/v1/consulting-services/x/download"},
		{"GET", "/api/v1/data-sources"},
		{"POST", "/api/v1/companies"},
		{"PUT", "/api/v1/companies/x"},
		{"DELETE", "/api/v1/companies/x"},
		{"POST", "/api/v1/quality-assurance/import"},
		{"POST", "/api/v1/data-sources/sync-all"},
	}
	for _, tc := range cases {
		for name, tok := range map[string]string{"VIEWER": viewer, "viewer": viewerLower} {
			if c := doViewerReq(engine, tc.method, tc.path, tok); c != http.StatusForbidden {
				t.Errorf("%s %s %s = %d, want 403", name, tc.method, tc.path, c)
			}
		}
		for name, tok := range map[string]string{"ADMIN": admin, "AUDITOR": auditor, "VIEWER+AUDITOR": mixed} {
			// Master-data and QA/consulting writes are refused for AUDITOR by
			// the write allow-lists (see write_rules_test.go), not the viewer
			// guard. Only data-sources writes are still open to AUDITOR.
			if name != "ADMIN" && tc.method != "GET" && !strings.HasPrefix(tc.path, "/api/v1/data-sources") {
				continue
			}
			if c := doViewerReq(engine, tc.method, tc.path, tok); c == http.StatusForbidden || c == http.StatusUnauthorized {
				t.Errorf("%s %s %s = %d, want it to pass the viewer guard", name, tc.method, tc.path, c)
			}
		}
	}
}

// Every route under /api/v1 must refuse a viewer, including ones added later.
func TestViewerRefusedOnEveryMasterRoute(t *testing.T) {
	engine := viewerTestEngine(t)
	viewer := viewerTestToken(t, "VIEWER")
	for _, r := range engine.Routes() {
		if !strings.HasPrefix(r.Path, "/api/v1/") {
			continue
		}
		path := r.Path
		for _, seg := range strings.Split(r.Path, "/") {
			if strings.HasPrefix(seg, ":") {
				path = strings.ReplaceAll(path, seg, "x")
			}
		}
		if c := doViewerReq(engine, r.Method, path, viewer); c != http.StatusForbidden {
			t.Errorf("viewer %s %s = %d, want 403", r.Method, r.Path, c)
		}
	}
}
