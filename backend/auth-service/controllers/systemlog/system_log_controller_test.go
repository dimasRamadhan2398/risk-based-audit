package systemlog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"auth-service/models"
	"auth-service/pkg/activitylog"
	"auth-service/pkg/validations"
	"auth-service/repositories"
	systemLogService "auth-service/services/systemlog"

	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeRepo struct {
	repositories.ISysAlertLogRepository
	rows []models.SysAlertLog
}

func (f *fakeRepo) List(_ context.Context, flt repositories.SysAlertLogFilter) ([]models.SysAlertLog, error) {
	return append([]models.SysAlertLog(nil), f.rows...), nil
}

type fakePerms map[string][]string

func (f fakePerms) NamesForRoles(roles []string) ([]string, error) {
	var out []string
	for _, r := range roles {
		out = append(out, f[strings.ToUpper(r)]...)
	}
	return out, nil
}

func engine(rows []models.SysAlertLog) *gin.Engine {
	gin.SetMode(gin.TestMode)
	svc := systemLogService.NewService(&fakeRepo{rows: rows}, fakePerms{
		"ADMIN":     {systemLogService.PermissionView, systemLogService.PermissionExport},
		"EXECUTIVE": {systemLogService.PermissionView},
	}, activitylog.NopRecorder{})
	ctrl := NewController(validations.New(), svc)
	e := gin.New()
	e.GET("/api/v1/system-logs", func(c *gin.Context) {
		// stands in for Authenticate: roles from the token
		if r := c.GetHeader("X-Test-Roles"); r != "" {
			c.Set("roles", strings.Split(r, ","))
		}
		c.Next()
	}, ctrl.List)
	return e
}

func get(e *gin.Engine, roles, query string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/system-logs"+query, nil)
	if roles != "" {
		req.Header.Set("X-Test-Roles", roles)
	}
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)
	return w
}

func TestSystemLogsPermission(t *testing.T) {
	e := engine(nil)
	for roles, want := range map[string]int{
		"ADMIN":           http.StatusOK,
		"admin":           http.StatusOK,
		"EXECUTIVE":       http.StatusOK,
		"VIEWER":          http.StatusForbidden,
		"AUDITOR":         http.StatusForbidden,
		"DEPARTMENT_HEAD": http.StatusForbidden,
		"":                http.StatusForbidden,
	} {
		if w := get(e, roles, ""); w.Code != want {
			t.Errorf("roles %q: %d, want %d (%s)", roles, w.Code, want, w.Body.String())
		}
	}
}

func TestSystemLogsResponseShape(t *testing.T) {
	actor := uuid.New()
	rows := []models.SysAlertLog{{
		ID: uuid.New(), OccurredAt: time.Date(2026, 11, 4, 1, 0, 0, 0, time.UTC), CreatedAt: time.Now(),
		ActorUserID: &actor, ActorUsername: "alice", Action: "user.login_failed", EntityType: "user",
		Result: "failure", IPAddress: "203.0.113.50", UserAgent: "UA", Summary: "Failed sign-in for alice",
		Details: models.JSONMap{"reason": "invalid_password", "password": "hunter2", "temporary_password": "Tmp#1"},
	}}
	e := engine(rows)

	w := get(e, "EXECUTIVE", "?limit=10")
	if w.Code != http.StatusOK {
		t.Fatalf("%d %s", w.Code, w.Body.String())
	}
	if w.Header().Get("Cache-Control") != "no-store" {
		t.Error("missing Cache-Control: no-store")
	}
	body := w.Body.String()
	for _, secret := range []string{"hunter2", "Tmp#1", "203.0.113.50"} {
		if strings.Contains(body, secret) {
			t.Errorf("response leaks %q: %s", secret, body)
		}
	}
	var got map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"success", "data", "next_cursor", "latest_cursor"} {
		if _, ok := got[k]; !ok {
			t.Errorf("missing top-level %q", k)
		}
	}
	if got["next_cursor"] != nil {
		t.Errorf("next_cursor = %v, want null", got["next_cursor"])
	}
	item := got["data"].([]interface{})[0].(map[string]interface{})
	for _, k := range []string{"id", "occurred_at", "actor_user_id", "actor_username", "action", "entity_type",
		"entity_id", "result", "ip_address", "user_agent", "summary", "details"} {
		if _, ok := item[k]; !ok {
			t.Errorf("item missing %q", k)
		}
	}
	if len(item) != 12 {
		t.Errorf("item has %d keys, want exactly 12: %v", len(item), item)
	}
	if item["ip_address"] != "203.0.x.x" || item["actor_user_id"] != actor.String() {
		t.Errorf("item = %v", item)
	}

	// Admin (export_system_logs) sees the full address.
	if w := get(e, "ADMIN", ""); !strings.Contains(w.Body.String(), "203.0.113.50") {
		t.Errorf("admin should see the full IP: %s", w.Body.String())
	}
	// Validation errors are 400.
	if w := get(e, "ADMIN", "?result=maybe"); w.Code != http.StatusBadRequest {
		t.Errorf("bad filter = %d", w.Code)
	}
}
