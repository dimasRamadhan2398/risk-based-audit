package routes

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"risk-service/controllers"
	"risk-service/middleware"
	"risk-service/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// A synthetic value of the minimum length, never a real key.
var viewerTestSecret = strings.Repeat("v", middleware.MinJWTSecretBytes)

type stubRiskService struct{}

func (stubRiskService) GetAll(context.Context) ([]services.RiskResponse, error) {
	return []services.RiskResponse{}, nil
}
func (stubRiskService) Create(context.Context, *services.RiskRequest) (*services.RiskResponse, error) {
	return &services.RiskResponse{}, nil
}
func (stubRiskService) Update(context.Context, uuid.UUID, *services.RiskRequest) (*services.RiskResponse, error) {
	return &services.RiskResponse{}, nil
}
func (stubRiskService) Delete(uuid.UUID) error { return nil }

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
	auth, err := middleware.NewAuthMiddleware(viewerTestSecret)
	if err != nil {
		t.Fatal(err)
	}
	engine := gin.New()
	engine.Use(gin.RecoveryWithWriter(io.Discard)) // nil db handlers panic; that means "passed the guard"
	NewRiskRoute(
		controllers.NewRiskController(stubRiskService{}),
		&controllers.MitigationController{},
		&controllers.RiskFactorController{},
		&controllers.AuditUniverseController{},
		&controllers.RCMController{},
		&engine.RouterGroup,
		auth,
	).Run()
	return engine
}

func doViewerReq(engine *gin.Engine, method, path, token string) int {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w.Code
}

func TestViewerRiskRoutes(t *testing.T) {
	engine := viewerTestEngine(t)
	viewer := viewerTestToken(t, "VIEWER")
	viewerLower := viewerTestToken(t, "viewer")
	admin := viewerTestToken(t, "ADMIN")
	auditor := viewerTestToken(t, "AUDITOR")

	if c := doViewerReq(engine, "GET", "/api/v1/risks", viewer); c != http.StatusOK {
		t.Errorf("viewer GET /api/v1/risks = %d, want 200", c)
	}
	for _, p := range []string{"/api/v1/mitigations", "/api/v1/rcm", "/api/v1/audit-universe/standard"} {
		if c := doViewerReq(engine, "GET", p, viewer); c == 403 || c == 401 {
			t.Errorf("viewer GET %s = %d, want it to pass the guard", p, c)
		}
	}

	writes := []struct{ method, path string }{
		{"POST", "/api/v1/risks"},
		{"PUT", "/api/v1/risks/x"},
		{"DELETE", "/api/v1/risks/x"},
		{"POST", "/api/v1/mitigations"},
		{"POST", "/api/v1/risk-factors/corporate"},
		{"POST", "/api/v1/audit-universe/corporate/sync-master"},
		{"POST", "/api/v1/audit-universe/year/2026/score"},
		{"POST", "/api/v1/audit-universe/events/resync"},
		{"POST", "/api/v1/rcm"},
		{"DELETE", "/api/v1/rcm/x"},
	}
	for _, tc := range writes {
		for name, tok := range map[string]string{"VIEWER": viewer, "viewer": viewerLower} {
			if c := doViewerReq(engine, tc.method, tc.path, tok); c != http.StatusForbidden {
				t.Errorf("%s %s %s = %d, want 403", name, tc.method, tc.path, c)
			}
		}
		for name, tok := range map[string]string{"ADMIN": admin, "AUDITOR": auditor} {
			// Audit-universe writes are refused for AUDITOR by the write
			// allow-list (see write_rules_test.go), not the viewer guard.
			if name == "AUDITOR" && strings.HasPrefix(tc.path, "/api/v1/audit-universe/") {
				continue
			}
			if c := doViewerReq(engine, tc.method, tc.path, tok); c == http.StatusForbidden || c == http.StatusUnauthorized {
				t.Errorf("%s %s %s = %d, want it to pass the viewer guard", name, tc.method, tc.path, c)
			}
		}
	}
}

func TestViewerEveryRiskRoute(t *testing.T) {
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
		c := doViewerReq(engine, r.Method, path, viewer)
		if r.Method == "GET" && c == http.StatusForbidden {
			t.Errorf("viewer GET %s = 403, but it is a plain read", r.Path)
		}
		if r.Method != "GET" && c != http.StatusForbidden {
			t.Errorf("viewer %s %s = %d, want 403", r.Method, r.Path, c)
		}
	}
}

// serve.go mounts the same middleware on the risk-appetite group, which hangs
// off the engine rather than the group in Run.
func TestViewerRiskAppetiteGroupShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	auth, _ := middleware.NewAuthMiddleware(viewerTestSecret)
	r := gin.New()
	g := r.Group("/api/v1/risk-appetite")
	g.Use(auth.Authenticate(), middleware.DenyViewerWrites())
	ok := func(c *gin.Context) { c.Status(http.StatusOK) }
	g.GET("", ok)
	g.POST("", ok)
	g.PUT("/:id", ok)
	g.DELETE("/:id", ok)
	viewer := viewerTestToken(t, "VIEWER")
	admin := viewerTestToken(t, "ADMIN")
	for method, path := range map[string]string{"POST": "/api/v1/risk-appetite", "PUT": "/api/v1/risk-appetite/1", "DELETE": "/api/v1/risk-appetite/1"} {
		if c := doViewerReq(r, method, path, viewer); c != http.StatusForbidden {
			t.Errorf("viewer %s %s = %d, want 403", method, path, c)
		}
		if c := doViewerReq(r, method, path, admin); c != http.StatusOK {
			t.Errorf("admin %s %s = %d, want 200", method, path, c)
		}
	}
	if c := doViewerReq(r, "GET", "/api/v1/risk-appetite", viewer); c != http.StatusOK {
		t.Errorf("viewer GET risk-appetite = %d, want 200", c)
	}
}
