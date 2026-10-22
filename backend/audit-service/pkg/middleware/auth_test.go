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
	mw := NewAuthMiddleware(authTestSecret)
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
	for _, name := range []string{"auth_token", "auth-token", "token"} {
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

// signAuthTestTokenWithClaims signs a token with the given claims. It is used
// for tokens that include employee claims.
func signAuthTestTokenWithClaims(t *testing.T, secret string, c Claims) string {
	t.Helper()
	c.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &c).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

// authTestContextRouter records the context values the middleware sets.
func authTestContextRouter(t *testing.T, got *map[string]any) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	mw := NewAuthMiddleware(authTestSecret)
	r := gin.New()
	r.GET("/p", mw.Authenticate(), func(c *gin.Context) {
		values := map[string]any{}
		for _, k := range []string{"user_id", "username", "employee_id", "employee_code"} {
			v, _ := c.Get(k)
			values[k] = v
		}
		*got = values
		c.Status(http.StatusOK)
	})
	return r
}

func TestAuthenticateExposesEmployeeClaimsOnContext(t *testing.T) {
	var got map[string]any
	r := authTestContextRouter(t, &got)
	const empUUID = "6f1c1b1e-0000-4000-8000-0000000000e1"
	tok := signAuthTestTokenWithClaims(t, authTestSecret, Claims{
		UserID:       "user-1",
		Username:     "tester",
		Roles:        []string{"AUDITOR"},
		EmployeeID:   empUUID,
		EmployeeCode: "EMP001",
	})

	if code := serve(r, "/p", map[string]string{"Authorization": "Bearer " + tok}); code != http.StatusOK {
		t.Fatalf("status = %d, want 200", code)
	}
	if got["employee_id"] != empUUID {
		t.Fatalf("employee_id on context = %v, want %s", got["employee_id"], empUUID)
	}
	if got["employee_code"] != "EMP001" {
		t.Fatalf("employee_code on context = %v, want EMP001", got["employee_code"])
	}
	if got["user_id"] != "user-1" || got["username"] != "tester" {
		t.Fatalf("existing context values changed: %v", got)
	}
}

func TestAuthenticateAcceptsTokenWithoutEmployeeClaims(t *testing.T) {
	// Tokens issued before the employee claims existed must still work. The
	// employee values are empty strings, not missing keys and not an error.
	var got map[string]any
	r := authTestContextRouter(t, &got)
	tok := signAuthTestToken(t, authTestSecret)

	if code := serve(r, "/p", map[string]string{"Authorization": "Bearer " + tok}); code != http.StatusOK {
		t.Fatalf("legacy token: status = %d, want 200", code)
	}
	if got["employee_id"] != "" || got["employee_code"] != "" {
		t.Fatalf("legacy token employee values = %v / %v, want empty", got["employee_id"], got["employee_code"])
	}
	if got["user_id"] != "user-1" {
		t.Fatalf("user_id = %v", got["user_id"])
	}
}
