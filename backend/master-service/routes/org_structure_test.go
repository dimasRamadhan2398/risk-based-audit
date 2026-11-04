package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	orgLevelCtrl "master-service/controllers/org_level"
	tenantConfigCtrl "master-service/controllers/tenant_config"
	"master-service/internal/orgtest"
	"master-service/pkg/middleware"
	"master-service/pkg/orgmode"
	"master-service/pkg/validations"
	orgLevelSvc "master-service/services/org_level"
	tenantConfigSvc "master-service/services/tenant_config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

const secret = "routes-test-secret-routes-test-secret-routes-test-secret-routes-test"

func accessToken(t *testing.T, roles ...string) string {
	t.Helper()
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, middleware.Claims{
		UserID: "u1", Username: "tester", Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))},
	}).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func engine(t *testing.T, db *gorm.DB, mode orgmode.Mode) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	auth := middleware.NewAuthMiddleware(secret)
	r := gin.New()
	api := r.Group("/api/v1")
	api.Use(auth.Authenticate())
	v := validations.New()
	tenant := tenantConfigCtrl.NewTenantConfigController(tenantConfigSvc.NewService(db, mode), v)
	var levels orgLevelCtrl.OrgLevelControllerInterface
	if mode.IsFlexible() {
		levels = orgLevelCtrl.NewOrgLevelController(orgLevelSvc.NewService(db, nil), v)
	}
	registerOrgStructureRoutes(api, tenant, levels, auth, OrgRouteOptions{Flexible: mode.IsFlexible()})
	api.GET("/departments/tree", middleware.RequireFlexible(mode.IsFlexible()), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	return r
}

func send(t *testing.T, r *gin.Engine, method, path, token, body string, headers ...string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	for i := 0; i+1 < len(headers); i += 2 {
		req.Header.Set(headers[i], headers[i+1])
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w.Code, out
}

func TestLegacyOrgRoutes(t *testing.T) {
	// No database: legacy routes must not touch one.
	r := engine(t, nil, orgmode.Legacy)
	tok := accessToken(t, "ADMIN")

	code, body := send(t, r, http.MethodGet, "/api/v1/tenant-config", tok, "")
	data, _ := body["data"].(map[string]any)
	org, _ := data["org_structure"].(map[string]any)
	if code != 200 || org["mode"] != "legacy" || len(org) != 1 {
		t.Fatalf("tenant-config legacy: %d %v", code, body)
	}
	for _, p := range []string{"/api/v1/org-levels", "/api/v1/org-levels/changes/preview", "/api/v1/org-levels/changes/apply", "/api/v1/departments/tree"} {
		method := http.MethodGet
		if strings.HasSuffix(p, "preview") || strings.HasSuffix(p, "apply") {
			method = http.MethodPost
		}
		code, body := send(t, r, method, p, tok, "{}")
		if code != http.StatusNotFound || body["code"] != "ORG_STRUCTURE_NOT_ENABLED" {
			t.Fatalf("%s %s: %d %v", method, p, code, body)
		}
	}
	if code, _ := send(t, r, http.MethodGet, "/api/v1/tenant-config", "", ""); code != http.StatusUnauthorized {
		t.Fatalf("tenant-config without token: %d", code)
	}
}

func TestFlexibleLevelChangesAreAdminOrExecutiveOnly(t *testing.T) {
	db := orgtest.DB(t)
	orgtest.BankLevels(t, db)
	r := engine(t, db, orgmode.Flexible)

	code, body := send(t, r, http.MethodGet, "/api/v1/tenant-config", accessToken(t, "AUDITOR"), "")
	data, _ := body["data"].(map[string]any)
	org, _ := data["org_structure"].(map[string]any)
	if code != 200 || org["mode"] != "flexible" || org["max_levels"] != float64(6) || len(org["levels"].([]any)) != 4 {
		t.Fatalf("tenant-config flexible: %d %v", code, body)
	}
	// Any authenticated user may list levels.
	if code, _ := send(t, r, http.MethodGet, "/api/v1/org-levels", accessToken(t, "VIEWER"), ""); code != 200 {
		t.Fatalf("list levels: %d", code)
	}

	preview := `{"levels":[{"level":10,"name":"Unit"}],"delete_ids":[]}`
	for _, role := range []string{"AUDITOR", "VIEWER", "AUDITEE", "DEPARTMENT_HEAD", "CAE", "CHIEF_AUDIT_EXECUTIVE"} {
		if code, _ := send(t, r, http.MethodPost, "/api/v1/org-levels/changes/preview", accessToken(t, role), preview); code != http.StatusForbidden {
			t.Fatalf("%s preview: %d", role, code)
		}
		if code, _ := send(t, r, http.MethodGet, "/api/v1/org-levels/changes", accessToken(t, role), ""); code != http.StatusForbidden {
			t.Fatalf("%s change log: %d", role, code)
		}
		// Step-up failures are 403 too, so check the role guard answered.
		if code, body := send(t, r, http.MethodPost, "/api/v1/org-levels/changes/apply", accessToken(t, role), preview); code != http.StatusForbidden || body["code"] != "FORBIDDEN" {
			t.Fatalf("%s apply: %d %v", role, code, body)
		}
	}
	for _, role := range []string{"ADMIN", "executive"} {
		code, body := send(t, r, http.MethodPost, "/api/v1/org-levels/changes/preview", accessToken(t, role), preview)
		if code != 200 {
			t.Fatalf("%s preview: %d %v", role, code, body)
		}
		if code, _ := send(t, r, http.MethodGet, "/api/v1/org-levels/changes", accessToken(t, role), ""); code != 200 {
			t.Fatalf("%s change log: %d", role, code)
		}
		// Apply without a step-up token: 403 (not 401, which signs the user out).
		code, body = send(t, r, http.MethodPost, "/api/v1/org-levels/changes/apply", accessToken(t, role), preview)
		if code != http.StatusForbidden || body["code"] != "STEP_UP_REQUIRED" {
			t.Fatalf("%s apply without step-up: %d %v", role, code, body)
		}
	}
	// A valid step-up token but no Redis: 503, never an unverified apply.
	now := time.Now()
	stepUp, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, middleware.StepUpClaims{Purpose: middleware.StepUpPurposeOrgLevels,
		RegisteredClaims: jwt.RegisteredClaims{Subject: "u1", ID: "jti-1", IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute))}}).
		SignedString(middleware.DeriveStepUpKey(secret))
	if code, _ := send(t, r, http.MethodPost, "/api/v1/org-levels/changes/apply", accessToken(t, "ADMIN"), preview, middleware.StepUpHeader, stepUp); code != http.StatusServiceUnavailable {
		t.Fatalf("apply without redis: %d", code)
	}
	if code, _ := send(t, r, http.MethodGet, "/api/v1/departments/tree", accessToken(t, "VIEWER"), ""); code != http.StatusNoContent {
		t.Fatalf("tree on flexible: %d", code)
	}
}
