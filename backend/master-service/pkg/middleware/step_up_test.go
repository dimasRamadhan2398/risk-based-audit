package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-at-least-64-bytes-long-test-secret-at-least-64-bytes-long"

type memStore struct {
	seen map[string]bool
	err  error
}

func (m *memStore) ConsumeOnce(_ context.Context, jti string, _ time.Duration) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	if m.seen[jti] {
		return false, nil
	}
	m.seen[jti] = true
	return true, nil
}

func sign(t *testing.T, key []byte, claims jwt.Claims) string {
	t.Helper()
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func stepUpClaims(sub, purpose, jti string, ttl time.Duration) StepUpClaims {
	now := time.Now()
	return StepUpClaims{Purpose: purpose, RegisteredClaims: jwt.RegisteredClaims{
		Subject: sub, ID: jti, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}}
}

func stepUpRouter(store JTIStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	m := NewAuthMiddleware(testSecret)
	r := gin.New()
	r.POST("/apply", func(c *gin.Context) { c.Set("user_id", "u1"); c.Next() },
		m.RequireStepUp(StepUpPurposeOrgLevels, store), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	return r
}

func doStepUp(r *gin.Engine, token string) int {
	req := httptest.NewRequest(http.MethodPost, "/apply", nil)
	if token != "" {
		req.Header.Set(StepUpHeader, token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestRequireStepUp(t *testing.T) {
	key := DeriveStepUpKey(testSecret)
	store := &memStore{seen: map[string]bool{}}
	r := stepUpRouter(store)

	if c := doStepUp(r, ""); c != http.StatusUnauthorized {
		t.Fatalf("missing token: %d", c)
	}
	good := sign(t, key, stepUpClaims("u1", StepUpPurposeOrgLevels, "j1", StepUpTTL))
	if c := doStepUp(r, good); c != http.StatusNoContent {
		t.Fatalf("valid token: %d", c)
	}
	if c := doStepUp(r, good); c != http.StatusUnauthorized {
		t.Fatalf("replayed token: %d", c)
	}
	cases := map[string]string{
		"signed with the raw jwt secret (an access-token key)": sign(t, []byte(testSecret), stepUpClaims("u1", StepUpPurposeOrgLevels, "j2", StepUpTTL)),
		"other user":     sign(t, key, stepUpClaims("u2", StepUpPurposeOrgLevels, "j3", StepUpTTL)),
		"other purpose":  sign(t, key, stepUpClaims("u1", "something.else", "j4", StepUpTTL)),
		"expired":        sign(t, key, stepUpClaims("u1", StepUpPurposeOrgLevels, "j5", -time.Minute)),
		"too long-lived": sign(t, key, stepUpClaims("u1", StepUpPurposeOrgLevels, "j6", time.Hour)),
		"no jti":         sign(t, key, stepUpClaims("u1", StepUpPurposeOrgLevels, "", StepUpTTL)),
	}
	for name, tok := range cases {
		if c := doStepUp(r, tok); c != http.StatusUnauthorized {
			t.Fatalf("%s: got %d", name, c)
		}
	}
	// Redis down / absent -> 503.
	if c := doStepUp(stepUpRouter(&memStore{err: errors.New("down")}), sign(t, key, stepUpClaims("u1", StepUpPurposeOrgLevels, "j7", StepUpTTL))); c != http.StatusServiceUnavailable {
		t.Fatalf("store error: %d", c)
	}
	if c := doStepUp(stepUpRouter(nil), sign(t, key, stepUpClaims("u1", StepUpPurposeOrgLevels, "j8", StepUpTTL))); c != http.StatusServiceUnavailable {
		t.Fatalf("nil store: %d", c)
	}
}

// A step-up token must never pass as an access token.
func TestStepUpTokenIsNotAnAccessToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := NewAuthMiddleware(testSecret)
	r := gin.New()
	r.GET("/x", m.Authenticate(), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	tok := sign(t, DeriveStepUpKey(testSecret), stepUpClaims("u1", StepUpPurposeOrgLevels, "j", StepUpTTL))
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("step-up token accepted as access token: %d", w.Code)
	}
}

func TestRequireRolesIsCaseInsensitive(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := NewAuthMiddleware(testSecret)
	for roles, want := range map[string]int{"admin": 204, "Executive": 204, "AUDITOR": 403, "CAE": 403} {
		r := gin.New()
		role := roles
		r.GET("/x", func(c *gin.Context) { c.Set("roles", []string{role}); c.Next() },
			m.RequireRoles("ADMIN", "EXECUTIVE"), func(c *gin.Context) { c.Status(http.StatusNoContent) })
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))
		if w.Code != want {
			t.Fatalf("role %s: got %d want %d", role, w.Code, want)
		}
	}
}
