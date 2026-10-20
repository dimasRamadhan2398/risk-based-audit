package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// sessionTestSecret is a synthetic test value, never a real key.
const sessionTestSecret = "session-test-"

var sessionTestKey = sessionTestSecret + strings.Repeat("s", 64)

func sessionTestToken(t *testing.T) string {
	t.Helper()
	claims := &Claims{
		UserID: "user-1", Username: "tester", Roles: []string{"VIEWER"},
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(sessionTestKey))
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

// sessionStore stands in for the Redis "token:<jwt>" keys written at login.
type sessionStore map[string]bool

func (s sessionStore) check(_ context.Context, token string) (bool, error) {
	return s[token], nil
}

func sessionRouter(m *AuthMiddleware) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/me", m.Authenticate(), func(c *gin.Context) { c.Status(http.StatusOK) })
	return r
}

func callWithBearer(r *gin.Engine, tok string) int {
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

// A logged-out token must stop working on protected routes before its JWT expiry.
func TestLogoutEndsSessionForProtectedRoutes(t *testing.T) {
	store := sessionStore{}
	m := NewAuthMiddleware(sessionTestKey).WithSessionCheck(store.check)
	r := sessionRouter(m)
	tok := sessionTestToken(t)

	store[tok] = true // login
	if code := callWithBearer(r, tok); code != http.StatusOK {
		t.Fatalf("live session: got %d, want 200", code)
	}

	delete(store, tok) // logout deletes the session key
	if code := callWithBearer(r, tok); code != http.StatusUnauthorized {
		t.Fatalf("after logout: got %d, want 401", code)
	}
}

func TestSessionCheckErrorFailsClosed(t *testing.T) {
	m := NewAuthMiddleware(sessionTestKey).WithSessionCheck(func(context.Context, string) (bool, error) {
		return true, errors.New("redis unavailable")
	})
	if code := callWithBearer(sessionRouter(m), sessionTestToken(t)); code != http.StatusUnauthorized {
		t.Fatalf("redis error must not authenticate: got %d, want 401", code)
	}
}

func TestWithoutSessionCheckJWTAloneIsEnough(t *testing.T) {
	if code := callWithBearer(sessionRouter(NewAuthMiddleware(sessionTestKey)), sessionTestToken(t)); code != http.StatusOK {
		t.Fatalf("no session check configured: got %d, want 200", code)
	}
}
