package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestRequireRolesForStatusChange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE letters (id TEXT PRIMARY KEY, status TEXT, deleted_at DATETIME)`).Error; err != nil {
		t.Fatal(err)
	}
	db.Exec(`INSERT INTO letters (id, status) VALUES ('draft-1', 'Draft'), ('pub-1', 'Published')`)

	guard := RequireRolesForStatusChange(db, "letters", "Draft", "ADMIN", "AUDIT_MANAGER")
	r := gin.New()
	withRoles := func(c *gin.Context) {
		if roles := c.GetHeader("X-Test-Roles"); roles != "" {
			c.Set("roles", strings.Split(roles, ","))
		}
	}
	// The handler echoes the body to prove the guard restored it
	echo := func(c *gin.Context) {
		body, _ := c.GetRawData()
		c.String(http.StatusOK, string(body))
	}
	r.POST("/letters", withRoles, guard, echo)
	r.PUT("/letters/:id", withRoles, guard, echo)

	cases := []struct {
		name, method, path, roles, body string
		want                            int
	}{
		{"auditor creates draft", "POST", "/letters", "AUDITOR", `{"status":"Draft","title":"x"}`, 200},
		{"auditor creates published", "POST", "/letters", "AUDITOR", `{"status":"Published"}`, 403},
		{"auditor edits draft keeping status", "PUT", "/letters/draft-1", "AUDITOR", `{"status":"Draft","title":"y"}`, 200},
		{"auditor edits without status", "PUT", "/letters/draft-1", "AUDITOR", `{"title":"y"}`, 200},
		{"auditor publishes", "PUT", "/letters/draft-1", "AUDITOR", `{"status":"Published"}`, 403},
		{"auditor edits published keeping status", "PUT", "/letters/pub-1", "AUDITOR", `{"status":"Published"}`, 200},
		{"auditor reverts published to draft", "PUT", "/letters/pub-1", "AUDITOR", `{"status":"Draft"}`, 403},
		{"no roles publishes", "PUT", "/letters/draft-1", "", `{"status":"Published"}`, 403},
		{"admin publishes", "PUT", "/letters/draft-1", "ADMIN", `{"status":"Published"}`, 200},
		{"manager publishes (lowercase role)", "PUT", "/letters/draft-1", "audit_manager", `{"status":"Published"}`, 200},
		{"unknown id passes through to handler", "PUT", "/letters/missing", "AUDITOR", `{"status":"Published"}`, 200},
	}

	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		req.Header.Set("Content-Type", "application/json")
		if tc.roles != "" {
			req.Header.Set("X-Test-Roles", tc.roles)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Errorf("%s: got %d, want %d", tc.name, w.Code, tc.want)
		}
		if w.Code == http.StatusOK && w.Body.String() != tc.body {
			t.Errorf("%s: handler got body %q, want %q", tc.name, w.Body.String(), tc.body)
		}
	}
}
