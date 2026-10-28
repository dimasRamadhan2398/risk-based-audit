package routes

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"auth-service/controllers"
	authCtrl "auth-service/controllers/auth"
	confidentialityCtrl "auth-service/controllers/confidentiality"
	mfaCtrl "auth-service/controllers/mfa"
	resendCtrl "auth-service/controllers/resend"
	roleCtrl "auth-service/controllers/role"
	trustedDevicesCtrl "auth-service/controllers/trusted-devices"
	userCtrl "auth-service/controllers/user"
	"auth-service/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func ok200(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"success": true}) }

// The fakes implement only the handlers a viewer may reach. Every other method
// panics through the nil embedded interface; the test engine recovers that as
// a 500, which reads as "passed the viewer guard".
type vAuth struct {
	authCtrl.AuthControllerInterface
}

func (vAuth) Me(c *gin.Context)             { ok200(c) }
func (vAuth) Logout(c *gin.Context)         { ok200(c) }
func (vAuth) ChangePassword(c *gin.Context) { ok200(c) }

type vUser struct {
	userCtrl.UserControllerInterface
}

func (vUser) GetUser(c *gin.Context) { ok200(c) }

type vConf struct {
	confidentialityCtrl.ConfidentialityControllerInterface
}

func (vConf) GetStatus(c *gin.Context) { ok200(c) }
func (vConf) Accept(c *gin.Context)    { ok200(c) }

type vControllers struct {
	controllers.IControllerRegistry
}

func (vControllers) GetAuth() authCtrl.AuthControllerInterface { return vAuth{} }
func (vControllers) GetUser() userCtrl.UserControllerInterface { return vUser{} }
func (vControllers) GetRole() roleCtrl.RoleControllerInterface {
	return struct {
		roleCtrl.RoleControllerInterface
	}{}
}
func (vControllers) GetMfa() mfaCtrl.MfaControllerInterface {
	return struct{ mfaCtrl.MfaControllerInterface }{}
}
func (vControllers) GetTrustedDevices() trustedDevicesCtrl.TrustedDevicesControllerInterface {
	return struct {
		trustedDevicesCtrl.TrustedDevicesControllerInterface
	}{}
}
func (vControllers) GetConfidentiality() confidentialityCtrl.ConfidentialityControllerInterface {
	return vConf{}
}
func (vControllers) GetResend() resendCtrl.ResendControllerInterface {
	return struct {
		resendCtrl.ResendControllerInterface
	}{}
}

func viewerAuthEngine(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(gin.RecoveryWithWriter(io.Discard))
	reg := &Registry{
		group:          engine.Group("/api/v1"),
		authMiddleware: middleware.NewAuthMiddleware(testSecret),
		controller:     vControllers{},
	}
	reg.Serve()
	return engine
}

func viewerReq(engine *gin.Engine, method, path, tok string) int {
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w.Code
}

func viewerTok(t *testing.T, userID string, roles ...string) string {
	t.Helper()
	claims := middleware.Claims{
		UserID: userID, Username: "tester", Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
	}
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestViewerAuthServiceAllowList(t *testing.T) {
	engine := viewerAuthEngine(t)
	viewer := viewerTok(t, "u1", "VIEWER")
	viewerLower := viewerTok(t, "u1", "viewer")

	allowed := []struct{ method, path string }{
		{"GET", "/api/v1/auth/me"},
		{"POST", "/api/v1/auth/logout"},
		{"POST", "/api/v1/auth/change-password"},
		{"GET", "/api/v1/users/u1"}, // own profile
		{"GET", "/api/v1/confidentiality/status"},
		{"POST", "/api/v1/confidentiality/accept"},
	}
	for _, tc := range allowed {
		for name, tok := range map[string]string{"VIEWER": viewer, "viewer": viewerLower} {
			if c := viewerReq(engine, tc.method, tc.path, tok); c != http.StatusOK {
				t.Errorf("%s %s %s = %d, want 200", name, tc.method, tc.path, c)
			}
		}
	}

	denied := []struct{ method, path string }{
		{"GET", "/api/v1/users"},
		{"GET", "/api/v1/users/assignable"},
		{"GET", "/api/v1/users/contacts"},
		{"GET", "/api/v1/users/lookup"},
		{"GET", "/api/v1/users/u2"}, // someone else's profile
		{"PUT", "/api/v1/users/u1"}, // profile update is not on the viewer list
		{"POST", "/api/v1/users"},
		{"DELETE", "/api/v1/users/u1"},
		{"POST", "/api/v1/users/u1/reset-password"},
		{"GET", "/api/v1/roles"},
		{"GET", "/api/v1/roles/all"},
		{"GET", "/api/v1/roles/x"},
		{"POST", "/api/v1/roles"},
		{"GET", "/api/v1/permissions"},
		{"GET", "/api/v1/mfa/status"},
		{"POST", "/api/v1/mfa/enroll"},
		{"POST", "/api/v1/mfa/disable"},
		{"GET", "/api/v1/devices"},
		{"POST", "/api/v1/devices/enroll"},
		{"DELETE", "/api/v1/devices/d1"},
	}
	for _, tc := range denied {
		for name, tok := range map[string]string{"VIEWER": viewer, "viewer": viewerLower} {
			if c := viewerReq(engine, tc.method, tc.path, tok); c != http.StatusForbidden {
				t.Errorf("%s %s %s = %d, want 403", name, tc.method, tc.path, c)
			}
		}
	}

	// Other roles are unaffected: they pass the viewer guard wherever their
	// own role rules let them (a 500 here is the nil fake handler).
	admin := viewerTok(t, "a1", "ADMIN")
	auditor := viewerTok(t, "u1", "AUDITOR")
	mixed := viewerTok(t, "u1", "VIEWER", "AUDITOR")
	for _, tc := range []struct {
		name, tok, method, path string
	}{
		{"ADMIN", admin, "GET", "/api/v1/users"},
		{"ADMIN", admin, "GET", "/api/v1/roles"},
		{"ADMIN", admin, "GET", "/api/v1/users/u2"},
		{"ADMIN", admin, "GET", "/api/v1/mfa/status"},
		{"AUDITOR", auditor, "GET", "/api/v1/users/assignable"},
		{"AUDITOR", auditor, "GET", "/api/v1/roles"},
		{"AUDITOR", auditor, "GET", "/api/v1/mfa/status"},
		{"AUDITOR", auditor, "GET", "/api/v1/devices"},
		{"VIEWER+AUDITOR", mixed, "GET", "/api/v1/users/assignable"},
	} {
		if c := viewerReq(engine, tc.method, tc.path, tc.tok); c == http.StatusForbidden || c == http.StatusUnauthorized {
			t.Errorf("%s %s %s = %d, want it to pass", tc.name, tc.method, tc.path, c)
		}
	}
	if c := viewerReq(engine, "GET", "/api/v1/users/u1", auditor); c != http.StatusOK {
		t.Errorf("AUDITOR own profile = %d, want 200", c)
	}
	if c := viewerReq(engine, "GET", "/api/v1/auth/me", admin); c != http.StatusOK {
		t.Errorf("ADMIN /auth/me = %d, want 200", c)
	}
}

// Deny-by-default: every authenticated route, including ones added later, must
// refuse a viewer unless it is on the allow-list.
func TestViewerAuthServiceEveryProtectedRoute(t *testing.T) {
	engine := viewerAuthEngine(t)
	viewer := viewerTok(t, "u1", "VIEWER")
	public := map[string]bool{
		"POST /api/v1/auth/login":            true,
		"POST /api/v1/auth/verify-mfa-login": true,
		"POST /api/v1/resend/webhook":        true,
	}
	allow := map[string]bool{
		"GET /api/v1/auth/me":                 true,
		"POST /api/v1/auth/logout":            true,
		"POST /api/v1/auth/change-password":   true,
		"GET /api/v1/users/:id":               true,
		"GET /api/v1/confidentiality/status":  true,
		"POST /api/v1/confidentiality/accept": true,
	}
	for _, r := range engine.Routes() {
		key := r.Method + " " + r.Path
		if public[key] {
			continue
		}
		path := strings.ReplaceAll(r.Path, "/:id", "/u1")
		for _, seg := range strings.Split(path, "/") {
			if strings.HasPrefix(seg, ":") {
				path = strings.ReplaceAll(path, seg, "x")
			}
		}
		c := viewerReq(engine, r.Method, path, viewer)
		if allow[key] && c != http.StatusOK {
			t.Errorf("viewer %s = %d, want 200", key, c)
		}
		if !allow[key] && c != http.StatusForbidden {
			t.Errorf("viewer %s = %d, want 403", key, c)
		}
	}
}
