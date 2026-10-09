package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"auth-service/controllers"
	userCtrl "auth-service/controllers/user"
	"auth-service/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret"

// fakeUserController records which handler ran; other methods panic (nil embed)
type fakeUserController struct {
	userCtrl.UserControllerInterface
	hit string
}

func (f *fakeUserController) ListAssignableUsers(c *gin.Context) {
	f.hit = "assignable"
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (f *fakeUserController) GetUser(c *gin.Context) {
	f.hit = "get:" + c.Param("id")
	c.JSON(http.StatusOK, gin.H{"success": true})
}

type fakeControllers struct {
	controllers.IControllerRegistry
	user *fakeUserController
}

func (f *fakeControllers) GetUser() userCtrl.UserControllerInterface { return f.user }

func newUsersRouter(t *testing.T) (*gin.Engine, *fakeUserController) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	user := &fakeUserController{}
	reg := &Registry{
		group:          engine.Group("/api/v1"),
		authMiddleware: middleware.NewAuthMiddleware(testSecret),
		controller:     &fakeControllers{user: user},
	}
	reg.users()
	return engine, user
}

func token(t *testing.T, userID string, roles ...string) string {
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

func TestAssignableUsersRoleCheck(t *testing.T) {
	cases := []struct {
		roles []string
		want  int
	}{
		{[]string{"ADMIN"}, http.StatusOK},
		{[]string{"admin"}, http.StatusOK},
		{[]string{"AUDITOR"}, http.StatusOK},
		{[]string{"AUDIT_MANAGER"}, http.StatusOK},
		{[]string{"Audit Manager"}, http.StatusOK},
		{[]string{"CHIEF_AUDIT_EXECUTIVE"}, http.StatusOK},
		{[]string{"Chief Audit Executive"}, http.StatusOK},
		{[]string{"EXECUTIVE"}, http.StatusOK},
		{[]string{"AUDITEE", "AUDITOR"}, http.StatusOK},
		{[]string{"AUDITEE"}, http.StatusForbidden},
		{[]string{"VIEWER"}, http.StatusForbidden},
		{[]string{"DEPARTMENT_HEAD"}, http.StatusForbidden},
		{[]string{"MANAGER"}, http.StatusForbidden},
		{nil, http.StatusForbidden},
	}
	for _, tc := range cases {
		engine, user := newUsersRouter(t)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/assignable?search=x", nil)
		req.Header.Set("Authorization", "Bearer "+token(t, "11111111-1111-1111-1111-111111111111", tc.roles...))
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Errorf("roles %v: status %d, want %d (%s)", tc.roles, w.Code, tc.want, w.Body.String())
		}
		if tc.want == http.StatusOK && user.hit != "assignable" {
			t.Errorf("roles %v: handler %q ran, want assignable", tc.roles, user.hit)
		}
		if tc.want != http.StatusOK && user.hit != "" {
			t.Errorf("roles %v: handler %q ran on a forbidden request", tc.roles, user.hit)
		}
	}
}

func TestAssignableUsersRequiresToken(t *testing.T) {
	engine, user := newUsersRouter(t)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/users/assignable", nil))
	if w.Code != http.StatusUnauthorized || user.hit != "" {
		t.Fatalf("no token: status %d, handler %q", w.Code, user.hit)
	}
}

// /users/assignable must not be swallowed by /users/:id (which would let a
// user fetch "their own" profile with id "assignable", or 403 an admin path)
func TestAssignableUsersNotShadowedByUserID(t *testing.T) {
	engine, user := newUsersRouter(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/assignable", nil)
	req.Header.Set("Authorization", "Bearer "+token(t, "assignable", "AUDITOR"))
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if user.hit != "assignable" {
		t.Fatalf("GET /users/assignable ran %q", user.hit)
	}

	engine, user = newUsersRouter(t)
	id := "22222222-2222-2222-2222-222222222222"
	req = httptest.NewRequest(http.MethodGet, "/api/v1/users/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+token(t, id, "AUDITEE"))
	w = httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusOK || user.hit != "get:"+id {
		t.Fatalf("GET /users/:id (self): status %d, handler %q", w.Code, user.hit)
	}
}
