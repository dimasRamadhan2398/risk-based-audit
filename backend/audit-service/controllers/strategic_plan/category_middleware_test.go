package strategicplan

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

// echoRouter runs ValidateCategory in front of a handler that returns the body
// it received
func echoRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/sp", ValidateCategory(), func(c *gin.Context) {
		b, _ := io.ReadAll(c.Request.Body)
		c.Data(http.StatusOK, "application/json", b)
	})
	return r
}

func TestValidateCategory(t *testing.T) {
	const absent = "<absent>"
	cases := []struct {
		name     string
		body     string
		wantCode int
		key      string // key to look up in the forwarded body
		want     string // expected value, or absent
	}{
		{"operational", `{"category":"Operational"}`, 200, "category", "Operational"},
		{"financial", `{"category":"Financial"}`, 200, "category", "Financial"},
		{"quality", `{"category":"Quality"}`, 200, "category", "Quality"},
		{"issue", `{"category":"Issue"}`, 200, "category", "Issue"},
		{"efficiency", `{"category":"Efficiency"}`, 200, "category", "Efficiency"},
		{"case-insensitive, canonicalised", `{"category":"financial"}`, 200, "category", "Financial"},
		{"trimmed", `{"category":"  EFFICIENCY "}`, 200, "category", "Efficiency"},
		{"empty kept (clears on update)", `{"category":""}`, 200, "category", ""},
		{"blank becomes empty", `{"category":"   "}`, 200, "category", ""},
		{"null dropped", `{"category":null,"kpi":"x"}`, 200, "category", absent},
		{"absent", `{"kpi":"x"}`, 200, "category", absent},
		{"capitalised key", `{"Category":"issue"}`, 200, "Category", "Issue"},
		{"unknown", `{"category":"Strategic"}`, 400, "", ""},
		{"prefix is not enough", `{"category":"Oper"}`, 400, "", ""},
		{"plural is unknown", `{"category":"Issues"}`, 400, "", ""},
		{"number", `{"category":3}`, 400, "", ""},
		{"array", `{"category":["Quality"]}`, 400, "", ""},
		{"object", `{"category":{"name":"Quality"}}`, 400, "", ""},
		{"capitalised key unknown", `{"CATEGORY":"Other"}`, 400, "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/sp", strings.NewReader(tc.body))
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
				if got["success"] != false || !strings.Contains(w.Body.String(), "Operational, Financial, Quality, Issue, Efficiency or empty") {
					t.Fatalf("unexpected error body: %s", w.Body.String())
				}
				return
			}
			v, ok := got[tc.key]
			if tc.want == absent {
				if ok {
					t.Fatalf("%q forwarded as %v: %s", tc.key, v, w.Body.String())
				}
				return
			}
			if !ok || v != tc.want {
				t.Fatalf("forwarded %q = %v (present %v), want %q", tc.key, v, ok, tc.want)
			}
		})
	}
}

func TestValidateCategoryKeepsOtherFieldsAndPassesThrough(t *testing.T) {
	// no category: the body is forwarded byte for byte
	in := `{"kpi":"K", "kpiTargets":{"2026":"5"},"yearStart":2026,"code":null}`
	w := httptest.NewRecorder()
	echoRouter().ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/sp", strings.NewReader(in)))
	if w.Code != 200 || w.Body.String() != in {
		t.Fatalf("got %d %q", w.Code, w.Body.String())
	}

	// with a category: other fields (including nulls and nested objects) survive
	w = httptest.NewRecorder()
	echoRouter().ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/sp",
		strings.NewReader(`{"category":"quality","kpi":"K","kpiTargets":{"2026":"5"},"yearStart":2026,"code":null}`)))
	var got map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got["category"] != "Quality" || got["kpi"] != "K" || got["yearStart"] != float64(2026) ||
		got["kpiTargets"].(map[string]interface{})["2026"] != "5" {
		t.Fatalf("fields not preserved: %s", w.Body.String())
	}
	if v, ok := got["code"]; !ok || v != nil {
		t.Fatalf("null field not preserved: %s", w.Body.String())
	}

	// not a JSON object: untouched, so crud reports the binding error as before
	for _, in := range []string{`not json`, `[1,2]`, `null`} {
		w = httptest.NewRecorder()
		echoRouter().ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/sp", strings.NewReader(in)))
		if w.Code != 200 || w.Body.String() != in {
			t.Fatalf("%s: got %d %q", in, w.Code, w.Body.String())
		}
	}
}

// End to end through the real crud handlers, as wired in routes/handler.go

const spTable = `CREATE TABLE strategic_plans (
	id TEXT PRIMARY KEY, code TEXT, goal_id TEXT, strategic_objective TEXT, kpi TEXT,
	unit TEXT, hib_hig TEXT, period_type TEXT, selected_period TEXT,
	year_start INTEGER, year_end INTEGER, kpi_targets TEXT, kpi_actuals TEXT,
	internal_audit_so TEXT, actual TEXT, target TEXT, calculation TEXT, status TEXT,
	category TEXT DEFAULT '',
	created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`

func spRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err := db.Exec(spTable).Error; err != nil {
		t.Fatal(err)
	}
	newOne := func() interface{} { return &models.StrategicPlan{} }
	r := gin.New()
	g := r.Group("/strategic-plans")
	g.GET("/:id", crud.GetByID(db, "StrategicPlan", newOne))
	g.POST("", ValidateCategory(), crud.Create(db, "StrategicPlan", newOne))
	g.PUT("/:id", ValidateCategory(), crud.Update(db, "StrategicPlan", newOne))
	return r, db
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

func storedCategory(t *testing.T, db *gorm.DB, id string) string {
	t.Helper()
	var p models.StrategicPlan
	if err := db.First(&p, "id = ?", id).Error; err != nil {
		t.Fatal(err)
	}
	return p.Category
}

func TestStrategicPlanCategoryEndToEnd(t *testing.T) {
	r, db := spRouter(t)
	const a = "11111111-1111-1111-1111-111111111111"
	const b = "22222222-2222-2222-2222-222222222222"
	path := "/strategic-plans/"

	// Create with a category (canonicalised) and without one ("")
	code, resp := do(t, r, http.MethodPost, "/strategic-plans", `{"id":"`+a+`","kpi":"A","category":"financial","kpiTargets":{"2026":"5"}}`)
	if code != http.StatusCreated || resp["data"].(map[string]interface{})["category"] != "Financial" {
		t.Fatalf("create A: %d %v", code, resp)
	}
	code, resp = do(t, r, http.MethodPost, "/strategic-plans", `{"id":"`+b+`","kpi":"B"}`)
	if code != http.StatusCreated || resp["data"].(map[string]interface{})["category"] != "" {
		t.Fatalf("create B: %d %v", code, resp)
	}
	if storedCategory(t, db, a) != "Financial" || storedCategory(t, db, b) != "" {
		t.Fatalf("stored: A=%q B=%q", storedCategory(t, db, a), storedCategory(t, db, b))
	}
	code, resp = do(t, r, http.MethodGet, path+a, "")
	if code != http.StatusOK || resp["data"].(map[string]interface{})["category"] != "Financial" {
		t.Fatalf("get A: %d %v", code, resp)
	}

	// Create with an unknown category: 400 and nothing is stored
	code, resp = do(t, r, http.MethodPost, "/strategic-plans", `{"id":"33333333-3333-3333-3333-333333333333","kpi":"C","category":"Risk"}`)
	if code != http.StatusBadRequest || resp["success"] != false {
		t.Fatalf("create bad: %d %v", code, resp)
	}
	var n int64
	db.Model(&models.StrategicPlan{}).Count(&n)
	if n != 2 {
		t.Fatalf("rows after rejected create = %d", n)
	}

	// Update without category keeps it; other fields still update
	code, resp = do(t, r, http.MethodPut, path+a, `{"kpi":"A2","target":"9"}`)
	if code != http.StatusOK {
		t.Fatalf("update no category: %d %v", code, resp)
	}
	data := resp["data"].(map[string]interface{})
	if data["category"] != "Financial" || data["kpi"] != "A2" || data["target"] != "9" || storedCategory(t, db, a) != "Financial" {
		t.Fatalf("update no category: %v", data)
	}

	// Update with null keeps it
	if code, resp = do(t, r, http.MethodPut, path+a, `{"category":null}`); code != http.StatusOK || storedCategory(t, db, a) != "Financial" {
		t.Fatalf("update null: %d %v stored=%q", code, resp, storedCategory(t, db, a))
	}

	// Update with an unknown category: 400, existing value and other fields untouched
	if code, resp = do(t, r, http.MethodPut, path+a, `{"category":"Strategic","kpi":"SHOULD NOT SAVE"}`); code != http.StatusBadRequest {
		t.Fatalf("update bad: %d %v", code, resp)
	}
	var p models.StrategicPlan
	db.First(&p, "id = ?", a)
	if p.Category != "Financial" || p.KPI != "A2" {
		t.Fatalf("after rejected update: %+v", p)
	}

	// Update to another category (any casing) and back to "" (cleared)
	if code, resp = do(t, r, http.MethodPut, path+a, `{"Category":"QUALITY"}`); code != http.StatusOK || storedCategory(t, db, a) != "Quality" {
		t.Fatalf("update Quality: %d %v stored=%q", code, resp, storedCategory(t, db, a))
	}
	if code, resp = do(t, r, http.MethodPut, path+a, `{"category":""}`); code != http.StatusOK || storedCategory(t, db, a) != "" {
		t.Fatalf("clear: %d %v stored=%q", code, resp, storedCategory(t, db, a))
	}

	// B was never touched
	if storedCategory(t, db, b) != "" {
		t.Fatalf("B changed: %q", storedCategory(t, db, b))
	}
}

func TestNormalizeStrategicPlanCategory(t *testing.T) {
	for in, want := range map[string]string{
		"Operational": "Operational", "financial": "Financial", " QUALITY ": "Quality",
		"issue": "Issue", "Efficiency": "Efficiency", "": "", "  ": "",
	} {
		if got, ok := models.NormalizeStrategicPlanCategory(in); !ok || got != want {
			t.Errorf("%q -> %q %v, want %q", in, got, ok, want)
		}
	}
	for _, in := range []string{"Other", "Issues", "Op", "Opera tional", "\u200bQuality"} {
		if got, ok := models.NormalizeStrategicPlanCategory(in); ok {
			t.Errorf("%q accepted as %q", in, got)
		}
	}
}
