package activityplan

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"audit-service/controllers/crud"
	"audit-service/models"
	"audit-service/pkg/activitycode/sqlitetest"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func planRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	db := sqlitetest.Open(t, &models.ActivityCodeSequence{}, &models.AuditActivity{}, &models.ActivityPlan{})
	newOne := func() interface{} { return &models.ActivityPlan{} }
	r := gin.New()
	g := r.Group("/activity-plans")
	g.GET("/:id", crud.GetByID(db, "ActivityPlan", newOne))
	g.POST("", crud.CreateWithHook(db, "ActivityPlan", newOne, AssignCodesOnCreate))
	g.PUT("/:id", crud.UpdateWithHook(db, "ActivityPlan", newOne, AssignCodesOnUpdate))
	return r, db
}

type planResp struct {
	ID                string `json:"id"`
	PlanYear          string `json:"planYear"`
	PlannedActivities []struct {
		ID           string  `json:"id"`
		ActivityCode string  `json:"activityCode"`
		Category     string  `json:"category"`
		AuditName    string  `json:"auditName"`
		Duration     float64 `json:"duration"`
	} `json:"plannedActivities"`
}

func call(t *testing.T, r *gin.Engine, method, path, body string) (int, planResp) {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var env struct {
		Data planResp `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &env)
	if w.Code >= 300 {
		t.Logf("%s %s -> %d %s", method, path, w.Code, w.Body.String())
	}
	return w.Code, env.Data
}

func codes(p planResp) []string {
	out := make([]string, len(p.PlannedActivities))
	for i, a := range p.PlannedActivities {
		out[i] = a.ActivityCode
	}
	return out
}

func wantCodes(t *testing.T, label string, got planResp, want ...string) {
	t.Helper()
	if strings.Join(codes(got), ",") != strings.Join(want, ",") {
		t.Errorf("%s: codes = %v, want %v", label, codes(got), want)
	}
}

func TestCreateAssignsCodesAndIgnoresClientValues(t *testing.T) {
	r, _ := planRouter(t)
	status, plan := call(t, r, http.MethodPost, "/activity-plans", `{
		"planTitle": "Plan A", "planYear": "2026",
		"plannedActivities": [
			{"id": "1", "auditName": "Cash", "category": "Assurance", "activityCode": "ASR-2026-999"},
			{"id": "2", "auditName": "Advice", "category": "Consulting Services"},
			{"id": "3", "auditName": "Bank", "category": "Assurance", "duration": 10}
		]}`)
	if status != http.StatusCreated {
		t.Fatalf("status %d", status)
	}
	wantCodes(t, "create", plan, "ASR-2026-001", "CNS-2026-001", "ASR-2026-002")

	// read back: stored, not just echoed
	_, got := call(t, r, http.MethodGet, "/activity-plans/"+plan.ID, "")
	wantCodes(t, "get", got, "ASR-2026-001", "CNS-2026-001", "ASR-2026-002")

	// the next plan in the same year continues the sequence per type
	_, plan2 := call(t, r, http.MethodPost, "/activity-plans", `{
		"planTitle": "Plan B", "planYear": "2026",
		"plannedActivities": [{"id": "x", "category": "Assurance"}, {"id": "y", "category": "Investigation"}]}`)
	wantCodes(t, "second plan", plan2, "ASR-2026-003", "INV-2026-001")

	// plan year decides the year, not today's date
	_, plan3 := call(t, r, http.MethodPost, "/activity-plans", `{
		"planTitle": "Plan C", "planYear": "2027",
		"plannedActivities": [{"id": "z", "category": "Assurance"}]}`)
	wantCodes(t, "2027 plan", plan3, "ASR-2027-001")
}

func TestUpdateKeepsCodesAndNumbersNewActivities(t *testing.T) {
	r, _ := planRouter(t)
	_, plan := call(t, r, http.MethodPost, "/activity-plans", `{
		"planTitle": "Plan", "planYear": "2026",
		"plannedActivities": [
			{"id": "1", "auditName": "Cash", "category": "Assurance"},
			{"id": "2", "auditName": "Bank", "category": "Assurance"}
		]}`)
	wantCodes(t, "create", plan, "ASR-2026-001", "ASR-2026-002")

	// Frontend-style edit: ids sent back (activityCode left out); first
	// activity's category and the plan year change, a row is removed and a
	// new one added. Existing code stays; the new one uses the new year.
	status, upd := call(t, r, http.MethodPut, "/activity-plans/"+plan.ID, `{
		"planYear": "2027",
		"plannedActivities": [
			{"id": "1", "auditName": "Cash (renamed)", "category": "Investigation", "duration": 5},
			{"id": "new", "auditName": "Stock", "category": "Assurance"}
		]}`)
	if status != http.StatusOK {
		t.Fatalf("status %d", status)
	}
	wantCodes(t, "update", upd, "ASR-2026-001", "ASR-2027-001")
	if upd.PlannedActivities[0].AuditName != "Cash (renamed)" || upd.PlannedActivities[0].Duration != 5 {
		t.Errorf("other fields not updated: %+v", upd.PlannedActivities[0])
	}

	// An update that does not touch plannedActivities leaves codes alone.
	_, upd2 := call(t, r, http.MethodPut, "/activity-plans/"+plan.ID, `{"planTitle": "Renamed"}`)
	wantCodes(t, "title only", upd2, "ASR-2026-001", "ASR-2027-001")

	// Matched by code when the id changed; ASR-2026-002 was removed above, so
	// it cannot be reclaimed, and a duplicated row gets its own code.
	_, upd3 := call(t, r, http.MethodPut, "/activity-plans/"+plan.ID, `{
		"plannedActivities": [
			{"id": "other", "activityCode": "ASR-2026-001", "category": "Investigation"},
			{"id": "other", "activityCode": "ASR-2026-001", "category": "Investigation"},
			{"id": "new2", "activityCode": "ASR-2026-002", "category": "Assurance"},
			{"id": "new", "category": "Assurance"}
		]}`)
	wantCodes(t, "rematch", upd3, "ASR-2026-001", "INV-2027-001", "ASR-2027-002", "ASR-2027-001")
}

func TestUpdateCannotClaimAnotherPlansCode(t *testing.T) {
	r, _ := planRouter(t)
	_, a := call(t, r, http.MethodPost, "/activity-plans", `{"planYear": "2026",
		"plannedActivities": [{"id": "1", "category": "Assurance"}]}`)
	_, b := call(t, r, http.MethodPost, "/activity-plans", `{"planYear": "2026",
		"plannedActivities": [{"id": "1", "category": "Assurance"}]}`)
	wantCodes(t, "b", b, "ASR-2026-002")
	_, upd := call(t, r, http.MethodPut, "/activity-plans/"+b.ID, `{
		"plannedActivities": [{"id": "1", "category": "Assurance"}, {"id": "2", "activityCode": "`+codes(a)[0]+`", "category": "Assurance"}]}`)
	wantCodes(t, "claim", upd, "ASR-2026-002", "ASR-2026-003")
}

func TestUpdateNumbersLegacyActivities(t *testing.T) {
	r, db := planRouter(t)
	// a plan stored before Activity IDs existed
	legacy := &models.ActivityPlan{PlanTitle: "old", PlanYear: "2026", PlannedActivities: []models.PlannedActivity{
		{ID: "1728371234567", Category: "Assurance"},
	}}
	if err := db.Create(legacy).Error; err != nil {
		t.Fatal(err)
	}
	_, got := call(t, r, http.MethodGet, "/activity-plans/"+legacy.ID.String(), "")
	wantCodes(t, "untouched on read", got, "")

	_, upd := call(t, r, http.MethodPut, "/activity-plans/"+legacy.ID.String(), `{
		"plannedActivities": [{"id": "1728371234567", "category": "Assurance"}]}`)
	wantCodes(t, "numbered on save", upd, "ASR-2026-001")
}

func TestUpdateRejectsNonArrayActivities(t *testing.T) {
	r, _ := planRouter(t)
	_, plan := call(t, r, http.MethodPost, "/activity-plans", `{"planYear": "2026",
		"plannedActivities": [{"id": "1", "category": "Assurance"}]}`)
	status, _ := call(t, r, http.MethodPut, "/activity-plans/"+plan.ID, `{"plannedActivities": {"id": "1"}}`)
	if status != http.StatusBadRequest {
		t.Errorf("status %d, want 400", status)
	}
	_, got := call(t, r, http.MethodGet, "/activity-plans/"+plan.ID, "")
	wantCodes(t, "unchanged", got, "ASR-2026-001")
}
