package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"auth-service/pkg/utils"

	"github.com/gin-gonic/gin"
)

func TestUserRouteAccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const secret = "test-secret"
	m := NewAuthMiddleware(secret)

	r := gin.New()
	users := r.Group("/users", m.Authenticate())
	selfOrAdmin := m.RequireSelfOrRoles("id", "ADMIN")
	ok := func(c *gin.Context) { c.Status(http.StatusOK) }
	users.GET("/:id", selfOrAdmin, ok)
	admin := users.Group("", m.RequireRoles("ADMIN"))
	admin.GET("", ok)
	admin.DELETE("/:id", ok)

	token := func(userID string, roles ...string) string {
		tok, err := utils.GenerateToken(userID, "u", roles, secret, 1)
		if err != nil {
			t.Fatal(err)
		}
		return tok
	}

	cases := []struct {
		name, method, path, token string
		want                      int
	}{
		{"no token", "GET", "/users/u1", "", http.StatusUnauthorized},
		{"own profile", "GET", "/users/u1", token("u1", "AUDITEE"), http.StatusOK},
		{"other profile", "GET", "/users/u2", token("u1", "AUDITEE"), http.StatusForbidden},
		{"admin reads other profile", "GET", "/users/u2", token("a1", "ADMIN"), http.StatusOK},
		{"non-admin list", "GET", "/users", token("u1", "AUDITOR"), http.StatusForbidden},
		{"admin list", "GET", "/users", token("a1", "admin"), http.StatusOK},
		{"non-admin delete self", "DELETE", "/users/u1", token("u1", "AUDITEE"), http.StatusForbidden},
	}

	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		if tc.token != "" {
			req.Header.Set("Authorization", "Bearer "+tc.token)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Errorf("%s: got %d, want %d", tc.name, w.Code, tc.want)
		}
	}
}
