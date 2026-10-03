package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"audit-service/controllers/crud"
	"audit-service/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// echoRouter runs NormalizeRequest in front of a handler that returns the body
// it received
func echoRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/atr", NormalizeRequest(), func(c *gin.Context) {
		b, _ := io.ReadAll(c.Request.Body)
		c.Data(http.StatusOK, "application/json", b)
	})
	return r
}

func TestNormalizeRequest(t *testing.T) {
	cases := []struct {
		name       string
		body       string
		wantCode   int
		wantStatus interface{} // expected "status" in the forwarded body; nil = absent
	}{
		{"title case in progress", `{"title":"x","status":"In Progress"}`, 200, "IN_PROGRESS"},
		{"lower snake", `{"status":"in_progress"}`, 200, "IN_PROGRESS"},
		{"completed", `{"status":"Completed"}`, 200, "COMPLETED"},
		{"canonical kept", `{"status":"PLANNED"}`, 200, "PLANNED"},
		{"cancelled", `{"status":"cancelled"}`, 200, "CANCELLED"},
		{"absent status", `{"title":"x"}`, 200, nil},
		{"empty status dropped", `{"status":"  "}`, 200, nil},
		{"null status dropped", `{"status":null}`, 200, nil},
		{"unknown status", `{"status":"Closed"}`, 400, nil},
		{"overdue is not a status", `{"status":"OVERDUE"}`, 400, nil},
		{"non-string status", `{"status":3}`, 400, nil},
		{"capitalised key unknown", `{"Status":"Done"}`, 400, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/atr", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			echoRouter().ServeHTTP(w, req)
			if w.Code != tc.wantCode {
				t.Fatalf("code = %d, want %d (body %s)", w.Code, tc.wantCode, w.Body.String())
			}
			var got map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if tc.wantCode == 400 {
				if got["success"] != false || !strings.Contains(w.Body.String(), "PLANNED, IN_PROGRESS, COMPLETED, CANCELLED") {
					t.Fatalf("unexpected error body: %s", w.Body.String())
				}
				return
			}
			if s, ok := got["status"]; tc.wantStatus == nil && ok || tc.wantStatus != nil && s != tc.wantStatus {
				t.Fatalf("forwarded status = %v (present %v), want %v; body %s", s, ok, tc.wantStatus, w.Body.String())
			}
		})
	}
}

func TestNormalizeRequestStripsDerivedFieldsAndKeepsOthers(t *testing.T) {
	w := httptest.NewRecorder()
	body := `{"isOverdue":true,"daysOverdue":7,"is_overdue":true,"DaysOverdue":1,"title":"T","deadline":"2026-01-01","assignment_letter_id":null}`
	echoRouter().ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/atr", strings.NewReader(body)))
	var got map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"isOverdue", "daysOverdue", "is_overdue", "DaysOverdue"} {
		if _, ok := got[k]; ok {
			t.Fatalf("derived field %q was forwarded: %s", k, w.Body.String())
		}
	}
	if got["title"] != "T" || got["deadline"] != "2026-01-01" {
		t.Fatalf("other fields not preserved: %s", w.Body.String())
	}
	if v, ok := got["assignment_letter_id"]; !ok || v != nil {
		t.Fatalf("null field not preserved: %s", w.Body.String())
	}
}

func TestNormalizeRequestPassesThroughNonObject(t *testing.T) {
	w := httptest.NewRecorder()
	echoRouter().ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/atr", strings.NewReader(`not json`)))
	if w.Code != 200 || w.Body.String() != "not json" {
		t.Fatalf("got %d %q", w.Code, w.Body.String())
	}
}

// End to end through the real crud handlers, as wired in routes/handler.go

const atrTable = `CREATE TABLE action_taken_reports (
	id TEXT PRIMARY KEY, assignment_letter_id TEXT, audit_finding_id TEXT,
	audit_ref TEXT NOT NULL DEFAULT '', title TEXT NOT NULL DEFAULT '', department TEXT,
	audit_object TEXT, finding_category TEXT, condition TEXT, criteria TEXT,
	recommendation TEXT, pic TEXT, deadline TEXT, status TEXT DEFAULT 'PLANNED',
	attachment TEXT, progress_description TEXT,
	created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`

func atrRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(atrTable).Error; err != nil {
		t.Fatal(err)
	}
	newOne := func() interface{} { return &models.ActionTakenReport{} }
	r := gin.New()
	g := r.Group("/action-taken-reports")
	g.GET("", crud.List(db, "ActionTakenReport", func() interface{} { return &[]models.ActionTakenReport{} }, "AssignmentLetter", "AuditFinding"))
	g.GET("/:id", crud.GetByID(db, "ActionTakenReport", newOne, "AssignmentLetter", "AuditFinding"))
	g.POST("", NormalizeRequest(), crud.Create(db, "ActionTakenReport", newOne))
	g.PUT("/:id", NormalizeRequest(), crud.Update(db, "ActionTakenReport", newOne))
	return r
}

func do(t *testing.T, r *gin.Engine, method, path, body string) (int, map[string]interface{}) {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("%s %s: bad JSON %q", method, path, w.Body.String())
	}
	return w.Code, resp
}

func TestATREndpointsEndToEnd(t *testing.T) {
	r := atrRouter(t)
	const id = "11111111-1111-1111-1111-111111111111"

	// Create: status normalised, client-sent derived fields ignored, overdue computed
	code, resp := do(t, r, http.MethodPost, "/action-taken-reports",
		`{"id":"`+id+`","auditRef":"ST-1","title":"T","deadline":"2000-01-01","status":"In Progress","isOverdue":false,"daysOverdue":0}`)
	if code != http.StatusCreated {
		t.Fatalf("create: %d %v", code, resp)
	}
	item := resp["data"].(map[string]interface{})
	if item["status"] != "IN_PROGRESS" || item["isOverdue"] != true || item["daysOverdue"].(float64) < 9000 {
		t.Fatalf("create response: %v", item)
	}

	// Create without status: DB default PLANNED; future deadline not overdue
	code, resp = do(t, r, http.MethodPost, "/action-taken-reports",
		`{"id":"22222222-2222-2222-2222-222222222222","auditRef":"ST-2","title":"U","deadline":"2999-12-31","isOverdue":true,"daysOverdue":4}`)
	if code != http.StatusCreated {
		t.Fatalf("create 2: %d %v", code, resp)
	}
	item = resp["data"].(map[string]interface{})
	if item["isOverdue"] != false || item["daysOverdue"].(float64) != 0 {
		t.Fatalf("create 2 response: %v", item)
	}
	_, resp = do(t, r, http.MethodGet, "/action-taken-reports/22222222-2222-2222-2222-222222222222", "")
	if s := resp["data"].(map[string]interface{})["status"]; s != "PLANNED" {
		t.Fatalf("default status = %v", s)
	}

	// Unknown status rejected on create
	if code, resp = do(t, r, http.MethodPost, "/action-taken-reports", `{"auditRef":"ST-3","title":"V","status":"Done"}`); code != http.StatusBadRequest {
		t.Fatalf("create bad status: %d %v", code, resp)
	}

	// Get by id
	_, resp = do(t, r, http.MethodGet, "/action-taken-reports/"+id, "")
	item = resp["data"].(map[string]interface{})
	if item["isOverdue"] != true || item["status"] != "IN_PROGRESS" {
		t.Fatalf("get: %v", item)
	}

	// List: every item carries the derived fields
	_, resp = do(t, r, http.MethodGet, "/action-taken-reports", "")
	items := resp["data"].(map[string]interface{})["items"].([]interface{})
	if len(items) != 2 {
		t.Fatalf("list: %v", resp)
	}
	for _, it := range items {
		m := it.(map[string]interface{})
		if _, ok := m["isOverdue"]; !ok {
			t.Fatalf("list item missing isOverdue: %v", m)
		}
		if _, ok := m["daysOverdue"]; !ok {
			t.Fatalf("list item missing daysOverdue: %v", m)
		}
	}

	// Update: lowercase status normalised, derived fields in body ignored
	code, resp = do(t, r, http.MethodPut, "/action-taken-reports/"+id, `{"status":"completed","isOverdue":true,"daysOverdue":3}`)
	if code != http.StatusOK {
		t.Fatalf("update: %d %v", code, resp)
	}
	item = resp["data"].(map[string]interface{})
	if item["status"] != "COMPLETED" || item["isOverdue"] != false || item["daysOverdue"].(float64) != 0 {
		t.Fatalf("update response: %v", item)
	}

	// Update with empty status leaves it unchanged; unknown status is rejected
	code, resp = do(t, r, http.MethodPut, "/action-taken-reports/"+id, `{"status":"","title":"T2"}`)
	if code != http.StatusOK || resp["data"].(map[string]interface{})["status"] != "COMPLETED" {
		t.Fatalf("update empty status: %d %v", code, resp)
	}
	if code, resp = do(t, r, http.MethodPut, "/action-taken-reports/"+id, `{"status":"OVERDUE"}`); code != http.StatusBadRequest {
		t.Fatalf("update bad status: %d %v", code, resp)
	}
}
