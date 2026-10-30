package routes

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"analytics-service/middleware"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// A synthetic value of the minimum length, never a real key.
var viewerTestSecret = strings.Repeat("a", middleware.MinJWTSecretBytes)

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
	// Any upstream call made by a request that passes the guard hits a closed port.
	t.Setenv("PYTHON_AI_URL", "http://127.0.0.1:1")
	t.Setenv("DATA_HUB_URL", "http://127.0.0.1:1")
	auth, err := middleware.NewAuthMiddleware(viewerTestSecret)
	if err != nil {
		t.Fatal(err)
	}
	engine := gin.New()
	engine.Use(gin.RecoveryWithWriter(io.Discard))
	SetupRoutes(engine, auth)
	return engine
}

func doViewerReq(engine *gin.Engine, method, path, token string) int {
	req := httptest.NewRequest(method, path, strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w.Code
}

func TestViewerAnalyticsRoutes(t *testing.T) {
	engine := viewerTestEngine(t)
	viewer := viewerTestToken(t, "VIEWER")
	viewerLower := viewerTestToken(t, "viewer")
	admin := viewerTestToken(t, "ADMIN")
	auditor := viewerTestToken(t, "AUDITOR")

	if c := doViewerReq(engine, "GET", "/api/analytics/report", viewer); c != http.StatusOK {
		t.Errorf("viewer GET /api/analytics/report = %d, want 200", c)
	}
	// POSTs that only compute a result stay open for a viewer.
	for _, p := range []string{"risk-score", "anomaly", "text-analysis", "performance-trend"} {
		if c := doViewerReq(engine, "POST", "/api/analytics/"+p, viewer); c == 403 || c == 401 {
			t.Errorf("viewer POST /api/analytics/%s = %d, want it to pass the guard", p, c)
		}
	}
	for _, p := range []string{"/api/analytics/caatt/benford", "/api/analytics/predict", "/api/analytics/risk-score/batch"} {
		if c := doViewerReq(engine, "GET", p, viewer); c == 403 || c == 401 {
			t.Errorf("viewer GET %s = %d, want it to pass the guard", p, c)
		}
	}

	for name, tok := range map[string]string{"VIEWER": viewer, "viewer": viewerLower} {
		if c := doViewerReq(engine, "POST", "/api/analytics/retrain/auto", tok); c != http.StatusForbidden {
			t.Errorf("%s POST /api/analytics/retrain/auto = %d, want 403", name, c)
		}
	}
	for name, tok := range map[string]string{"ADMIN": admin, "AUDITOR": auditor} {
		if c := doViewerReq(engine, "POST", "/api/analytics/retrain/auto", tok); c == http.StatusForbidden || c == http.StatusUnauthorized {
			t.Errorf("%s POST /api/analytics/retrain/auto = %d, want it to pass the viewer guard", name, c)
		}
	}
}

func TestViewerEveryAnalyticsRoute(t *testing.T) {
	engine := viewerTestEngine(t)
	viewer := viewerTestToken(t, "VIEWER")
	open := map[string]bool{
		"POST /api/analytics/risk-score":        true,
		"POST /api/analytics/anomaly":           true,
		"POST /api/analytics/text-analysis":     true,
		"POST /api/analytics/performance-trend": true,
	}
	for _, r := range engine.Routes() {
		c := doViewerReq(engine, r.Method, r.Path, viewer)
		wantOpen := r.Method == "GET" || open[r.Method+" "+r.Path]
		if wantOpen && c == http.StatusForbidden {
			t.Errorf("viewer %s %s = 403, but it is a read", r.Method, r.Path)
		}
		if !wantOpen && c != http.StatusForbidden {
			t.Errorf("viewer %s %s = %d, want 403", r.Method, r.Path, c)
		}
	}
}
