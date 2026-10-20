package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDebugRoutesOffByDefault(t *testing.T) {
	t.Setenv(DebugRoutesEnv, "")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerDebugRoutes(r)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/debug/routes", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("/debug/routes with flag unset: got %d, want 404", w.Code)
	}
}

func TestDebugRoutesOnlyWhenExplicitlyEnabled(t *testing.T) {
	t.Setenv(DebugRoutesEnv, "true")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	registerDebugRoutes(r)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/debug/routes", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("/debug/routes with flag true: got %d, want 200", w.Code)
	}
}

func TestSwaggerFlagIsStrictlyTrue(t *testing.T) {
	cases := map[string]bool{"": false, "false": false, "1": false, "yes": false, "true": true, " TRUE ": true}
	for v, want := range cases {
		t.Setenv(SwaggerEnv, v)
		if got := SwaggerEnabled(); got != want {
			t.Fatalf("ENABLE_SWAGGER=%q: enabled=%v, want %v", v, got, want)
		}
	}
}
