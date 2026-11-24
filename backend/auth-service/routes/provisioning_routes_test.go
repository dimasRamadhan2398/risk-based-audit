package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"auth-service/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// Only ADMIN reaches the provisioning proxy. With no PROVISIONER_URL the
// proxy answers 503, which is how "passed the role gate" shows here.
func TestProvisioningRoutesAdminOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	reg := &Registry{group: engine.Group("/api/v1"), authMiddleware: middleware.NewAuthMiddleware(testSecret)}
	reg.provisioningRoutes()

	paths := []struct{ method, path string }{
		{"POST", "/api/v1/provisioning/jobs"},
		{"GET", "/api/v1/provisioning/jobs"},
		{"GET", "/api/v1/provisioning/jobs/0b6a3c1e-1d2f-4a5b-8c7d-9e0f1a2b3c4d"},
		{"GET", "/api/v1/provisioning/jobs/0b6a3c1e-1d2f-4a5b-8c7d-9e0f1a2b3c4d/events"},
		{"GET", "/api/v1/provisioning/ports/next"},
	}
	cases := []struct {
		name  string
		roles []string
		want  int
	}{
		{"admin", []string{"ADMIN"}, http.StatusServiceUnavailable},
		{"admin lower case", []string{"admin"}, http.StatusServiceUnavailable},
		{"auditor", []string{"AUDITOR"}, http.StatusForbidden},
		{"executive", []string{"EXECUTIVE"}, http.StatusForbidden},
		{"audit manager", []string{"AUDIT_MANAGER"}, http.StatusForbidden},
		{"viewer", []string{"VIEWER"}, http.StatusForbidden},
	}
	for _, p := range paths {
		req := httptest.NewRequest(p.method, p.path, nil)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("no token %s %s = %d", p.method, p.path, w.Code)
		}
		for _, tc := range cases {
			req := httptest.NewRequest(p.method, p.path, nil)
			req.Header.Set("Authorization", "Bearer "+token(t, "u1", tc.roles...))
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			if w.Code != tc.want {
				t.Errorf("%s %s %s = %d, want %d", tc.name, p.method, p.path, w.Code, tc.want)
			}
		}
	}
}
