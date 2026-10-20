package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// authTestSecret is synthetic and only used inside these tests.
var authTestSecret = "test-" + strings.Repeat("y", 64)

func signAuthTestToken(t *testing.T, secret string) string {
	t.Helper()
	claims := &Claims{
		UserID:   "user-1",
		Username: "tester",
		Roles:    []string{"VIEWER"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func authTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	mw := newTestAuth(t)
	r := gin.New()
	r.GET("/p", mw.Authenticate(), func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func serve(r *gin.Engine, target string, headers map[string]string, cookies ...*http.Cookie) int {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, ck := range cookies {
		req.AddCookie(ck)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestAuthenticateAcceptsBearerHeader(t *testing.T) {
	r := authTestRouter(t)
	tok := signAuthTestToken(t, authTestSecret)
	if code := serve(r, "/p", map[string]string{"Authorization": "Bearer " + tok}); code != http.StatusOK {
		t.Fatalf("valid bearer token: got %d, want 200", code)
	}
}

func TestAuthenticateRefusesQueryStringToken(t *testing.T) {
	r := authTestRouter(t)
	tok := signAuthTestToken(t, authTestSecret)
	if code := serve(r, "/p?token="+tok, nil); code != http.StatusUnauthorized {
		t.Fatalf("?token= must not authenticate: got %d, want 401", code)
	}
}

func TestAuthenticateRefusesCookieToken(t *testing.T) {
	r := authTestRouter(t)
	tok := signAuthTestToken(t, authTestSecret)
	for _, name := range []string{"auth_token", "token"} {
		if code := serve(r, "/p", nil, &http.Cookie{Name: name, Value: tok}); code != http.StatusUnauthorized {
			t.Fatalf("cookie %q must not authenticate: got %d, want 401", name, code)
		}
	}
}

func TestAuthenticateRefusesMissingAndForeignTokens(t *testing.T) {
	r := authTestRouter(t)
	if code := serve(r, "/p", nil); code != http.StatusUnauthorized {
		t.Fatalf("no token: got %d, want 401", code)
	}
	foreign := signAuthTestToken(t, "other-"+strings.Repeat("z", 64))
	if code := serve(r, "/p", map[string]string{"Authorization": "Bearer " + foreign}); code != http.StatusUnauthorized {
		t.Fatalf("token signed with another secret: got %d, want 401", code)
	}
}

func newTestAuth(t *testing.T) *AuthMiddleware {
	t.Helper()
	m, err := NewAuthMiddleware(authTestSecret)
	if err != nil {
		t.Fatal(err)
	}
	return m
}
