package annualplan_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	ctrlAnnualPlan "audit-service/controllers/annual_plan"
	"audit-service/controllers/crud"
	"audit-service/models"
	"audit-service/pkg/activitycode/sqlitetest"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// activityRouter serves the annual plan routes with the same handlers as
// routes/handler.go (GET, POST, PUT, DELETE). The activity plan and audit
// activity tables exist because the update hook checks RKA usage.
func activityRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := sqlitetest.Open(t, &models.ActivityCodeSequence{}, &models.AuditAnnual{}, &models.AnnualSnapshotRef{},
		&models.ActivityPlan{}, &models.AuditActivity{}, &models.AuditExecution{})
	db.Logger = logger.Discard
	r := gin.New()
	g := r.Group("/annual-audit-plans")
	g.GET("", crud.ListWithItems(db, "AuditAnnual", func() interface{} { return &[]models.AuditAnnual{} }, ctrlAnnualPlan.DecorateList(db)))
	g.GET("/:id", ctrlAnnualPlan.GetByID(db))
	g.POST("", crud.CreateWithHook(db, "AuditAnnual", func() interface{} { return &models.AuditAnnual{} }, ctrlAnnualPlan.PrepareOnCreate))
	g.PUT("/:id", crud.UpdateWithHooks(db, "AuditAnnual", func() interface{} { return &models.AuditAnnual{} }, ctrlAnnualPlan.PrepareOnUpdate, ctrlAnnualPlan.RefreshSnapshotsAfterUpdate))
	g.DELETE("/:id", ctrlAnnualPlan.Delete(db))
	return r, db
}

func doJSON(t *testing.T, r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

// dataOf returns the "data" member of a response envelope.
func dataOf(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var env struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode response: %v: %s", err, w.Body.String())
	}
	return env.Data
}

// fullActivity uses every documented activity field with a distinct value.
const fullActivity = `{
  "itemNumber": "1",
  "activityCode": "ASR-2026-001",
  "name": "Audit pengadaan TI",
  "category": "Assurance",
  "auditUniverseId": "5f0c2a2e-8b6e-4c2b-9d1e-0a1b2c3d4e5f",
  "riskName": "Teknologi Informasi",
  "riskLevel": "High",
  "departmentId": "9a8b7c6d-1111-4222-8333-444455556666",
  "department": "Information Technology",
  "justification": "Risiko tinggi pada sistem pembayaran",
  "auditObjective": "Menilai kecukupan kontrol akses",
  "auditFocus": "Akses privileged",
  "timelineText": "Q2 2026",
  "auditorCount": 3,
  "totalMandays": 45,
  "supervisorName": "Budi Santoso"
}`

// TestActivityFieldsSurvivePutAndGet posts a plan with every activity field,
// PUTs the rows back as the client would (ids included), and checks that GET
// returns every field with the same value.
func TestActivityFieldsSurvivePutAndGet(t *testing.T) {
	r, _ := activityRouter(t)

	post := `{"year":2026,"title":"Rencana Audit 2026","activityType":"Assurance","activities":[` + fullActivity + `]}`
	w := doJSON(t, r, http.MethodPost, "/annual-audit-plans", post)
	if w.Code != http.StatusCreated {
		t.Fatalf("POST: %d %s", w.Code, w.Body.String())
	}
	created := dataOf(t, w)
	id, _ := created["id"].(string)
	if id == "" {
		t.Fatalf("POST returned no id: %s", w.Body.String())
	}

	got := getPlan(t, r, id)
	acts := got["activities"].([]interface{})
	if len(acts) != 1 {
		t.Fatalf("want 1 activity after POST, got %d", len(acts))
	}
	row := acts[0].(map[string]interface{})
	rowID, err := uuid.Parse(row["id"].(string))
	if err != nil || rowID == uuid.Nil {
		t.Fatalf("activity id not a UUID: %v", row["id"])
	}

	// PUT the plan back with the activity exactly as GET returned it, and
	// with a header change, as the frontend does when the user edits a plan.
	putBody, _ := json.Marshal(map[string]interface{}{
		"title":        "Rencana Audit 2026 (revisi)",
		"activityType": "Audit Khusus",
		"activities":   acts,
	})
	w = doJSON(t, r, http.MethodPut, "/annual-audit-plans/"+id, string(putBody))
	if w.Code != http.StatusOK {
		t.Fatalf("PUT: %d %s", w.Code, w.Body.String())
	}

	after := getPlan(t, r, id)
	if after["title"] != "Rencana Audit 2026 (revisi)" || after["activityType"] != "Audit Khusus" {
		t.Fatalf("header not updated: title=%v type=%v", after["title"], after["activityType"])
	}
	afterRow := after["activities"].([]interface{})[0].(map[string]interface{})
	if !reflect.DeepEqual(row, afterRow) {
		t.Fatalf("activity changed across PUT/GET\nbefore: %v\nafter:  %v", row, afterRow)
	}
	// Every documented field must be present, not just non-zero ones.
	for _, key := range activityKeys {
		if _, ok := afterRow[key]; !ok {
			t.Errorf("GET activity lacks field %q", key)
		}
	}
}

var activityKeys = []string{
	"id", "itemNumber", "activityCode", "name", "category", "auditUniverseId",
	"riskName", "riskLevel", "departmentId", "department", "justification",
	"auditObjective", "auditFocus", "timelineText", "auditorCount", "totalMandays",
	"supervisorName",
}

func getPlan(t *testing.T, r *gin.Engine, id string) map[string]interface{} {
	t.Helper()
	w := doJSON(t, r, http.MethodGet, "/annual-audit-plans/"+id, "")
	if w.Code != http.StatusOK {
		t.Fatalf("GET: %d %s", w.Code, w.Body.String())
	}
	return dataOf(t, w)
}

// TestPutRewritesActivityIDs checks the id rules: a missing or invalid id gets a
// new UUID, a duplicate id is replaced on the second row, and undocumented keys
// are dropped.
func TestPutRewritesActivityIDs(t *testing.T) {
	r, _ := activityRouter(t)
	w := doJSON(t, r, http.MethodPost, "/annual-audit-plans", `{"year":2026,"activities":[`+fullActivity+`]}`)
	id := dataOf(t, w)["id"].(string)

	keep := "3c9e7f10-0000-4000-8000-000000000001"
	body := `{"activities":[
	  {"id":"` + keep + `","name":"A","unknownKey":"x"},
	  {"id":"` + keep + `","name":"B"},
	  {"id":"not-a-uuid","name":"C"},
	  {"name":"D"}
	]}`
	w = doJSON(t, r, http.MethodPut, "/annual-audit-plans/"+id, body)
	if w.Code != http.StatusOK {
		t.Fatalf("PUT: %d %s", w.Code, w.Body.String())
	}
	acts := getPlan(t, r, id)["activities"].([]interface{})
	if len(acts) != 4 {
		t.Fatalf("want 4 rows, got %d", len(acts))
	}
	seen := map[string]bool{}
	for i, a := range acts {
		row := a.(map[string]interface{})
		rid := row["id"].(string)
		if _, err := uuid.Parse(rid); err != nil {
			t.Errorf("row %d: id %q is not a UUID", i, rid)
		}
		if seen[rid] {
			t.Errorf("row %d: duplicate id %q", i, rid)
		}
		seen[rid] = true
		if _, ok := row["unknownKey"]; ok {
			t.Errorf("row %d: undocumented key was stored", i)
		}
	}
	if acts[0].(map[string]interface{})["id"] != keep {
		t.Errorf("first row lost its id: %v", acts[0])
	}
}

func TestPutRejectsInvalidHeaderAndRows(t *testing.T) {
	r, _ := activityRouter(t)
	w := doJSON(t, r, http.MethodPost, "/annual-audit-plans", `{"year":2026}`)
	id := dataOf(t, w)["id"].(string)

	cases := map[string]string{
		"activity type outside template": `{"activityType":"Lainnya"}`,
		"negative auditor count":         `{"activities":[{"name":"x","auditorCount":-1}]}`,
		"wrong type for mandays":         `{"activities":[{"name":"x","totalMandays":"many"}]}`,
	}
	for name, body := range cases {
		w := doJSON(t, r, http.MethodPut, "/annual-audit-plans/"+id, body)
		if w.Code != http.StatusBadRequest {
			t.Errorf("%s: want 400, got %d %s", name, w.Code, w.Body.String())
		}
	}
}

func TestPostAssignsActivityIDsAndKeepsEmptyType(t *testing.T) {
	r, _ := activityRouter(t)
	w := doJSON(t, r, http.MethodPost, "/annual-audit-plans", `{"year":2026,"activityType":"","activities":[{"id":"x","name":"a"}]}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("POST: %d %s", w.Code, w.Body.String())
	}
	acts := dataOf(t, w)["activities"].([]interface{})
	if _, err := uuid.Parse(acts[0].(map[string]interface{})["id"].(string)); err != nil {
		t.Fatalf("POST did not assign a UUID: %v", acts[0])
	}
}
