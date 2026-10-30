package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"auth-service/controllers"
	authCtrl "auth-service/controllers/auth"
	"auth-service/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// fakeAuthHandlers answers every auth route; registration must not be among them.
type fakeAuthHandlers struct {
	authCtrl.AuthControllerInterface
}

func (fakeAuthHandlers) Login(c *gin.Context)          { c.Status(http.StatusOK) }
func (fakeAuthHandlers) VerifyMFALogin(c *gin.Context) { c.Status(http.StatusOK) }
func (fakeAuthHandlers) Logout(c *gin.Context)         { c.Status(http.StatusOK) }
func (fakeAuthHandlers) ChangePassword(c *gin.Context) { c.Status(http.StatusOK) }
func (fakeAuthHandlers) Me(c *gin.Context)             { c.Status(http.StatusOK) }

type fakeAuthControllers struct {
	controllers.IControllerRegistry
}

func (fakeAuthControllers) GetAuth() authCtrl.AuthControllerInterface { return fakeAuthHandlers{} }

type fakeStepUpHandlers struct{}

func (fakeStepUpHandlers) VerifyPassword(c *gin.Context) { c.Status(http.StatusOK) }

func (fakeAuthControllers) GetStepUp() authCtrl.StepUpControllerInterface { return fakeStepUpHandlers{} }

// Self-registration was open and created active users with no role. It must not
// be routed at all, whether or not the caller is authenticated.
func TestPublicRegisterRouteIsRemoved(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	reg := &Registry{
		group:          engine.Group("/api/v1"),
		authMiddleware: middleware.NewAuthMiddleware(testSecret),
		controller:     fakeAuthControllers{},
	}
	reg.auth()

	for _, route := range engine.Routes() {
		if strings.HasSuffix(route.Path, "/register") {
			t.Fatalf("registration route still exposed: %s %s", route.Method, route.Path)
		}
	}

	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader("{}")))
	if w.Code != http.StatusNotFound {
		t.Fatalf("POST /api/v1/auth/register = %d, want 404", w.Code)
	}
}
