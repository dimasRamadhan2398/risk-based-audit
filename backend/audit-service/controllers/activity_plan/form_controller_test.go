package activityplan

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"audit-service/models"
	"audit-service/pkg/activitycode/sqlitetest"

	"github.com/gin-gonic/gin"
)

// TestOptionsRouteNotShadowedByGetByID guards against GET /activity-plans/:id
// capturing /activity-plans/annual-activity-options. Such a capture makes the
// generic handler read "annual-activity-options" as a plan ID and answer
// 400 "Invalid ActivityPlan ID".
func TestOptionsRouteNotShadowedByGetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := sqlitetest.Open(t, &models.AuditAnnual{}, &models.AuditActivity{}, &models.ActivityPlan{}, &models.ActivityCodeSequence{})

	r := gin.New()
	NewFormController(db).Register(r.Group("/api/v1/activity-plans"))

	q := url.Values{}
	q.Set("annual_plan_id", "a364ed6b-016e-47bb-945c-2638d00e2185")
	q.Set("activity_type", models.ActivityTypeInvestigation)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/activity-plans/annual-activity-options?"+q.Encode(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if strings.Contains(w.Body.String(), "Invalid ActivityPlan ID") {
		t.Fatalf("annual-activity-options was handled as GET /:id: %d %s", w.Code, w.Body.String())
	}
	// The DB has no PKAT row for this ID, so Options reports it as not found.
	// That proves the request reached Options and its UUID parsing passed.
	if w.Code != http.StatusNotFound || !strings.Contains(w.Body.String(), "ANNUAL_PLAN_NOT_FOUND") {
		t.Fatalf("status = %d, want 404 ANNUAL_PLAN_NOT_FOUND from Options; body: %s", w.Code, w.Body.String())
	}
}
