package models_test

// These tests document that the backend puts no rule on how an annual audit
// plan's months/quarters are distributed. The "Beban kerja Triwulan I terlalu
// tinggi (>40%). Mohon ratakan jadwal" error shown by the Add Annual Audit
// Plan modal is produced by the frontend only; POST/PUT /annual-audit-plans go
// straight to the generic crud handlers (routes/handler.go) and AuditAnnual has
// no hooks, so Q1-heavy and Q1-only plans are stored as sent.

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"audit-service/controllers/crud"
	"audit-service/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// sqliteUUID stands in for Postgres' gen_random_uuid(), so a POST without an
// id gets one from the database as it does in production
const sqliteUUID = `(lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' ||
	substr(lower(hex(randomblob(2))), 2) || '-' || substr('89ab', 1 + (abs(random()) % 4), 1) ||
	substr(lower(hex(randomblob(2))), 2) || '-' || lower(hex(randomblob(6))))`

// annualPlanDB creates audit_annuals from the GORM schema (AutoMigrate cannot
// run on sqlite because ids default to gen_random_uuid()), keeping the
// model's literal defaults (version, status, is_active)
func annualPlanDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(&models.AuditAnnual{}); err != nil {
		t.Fatal(err)
	}
	var cols []string
	for _, f := range stmt.Schema.Fields {
		if f.DBName == "" {
			continue
		}
		typ := "TEXT"
		switch dt := strings.ToLower(string(f.DataType)); {
		case strings.Contains(dt, "int"):
			typ = "INTEGER"
		case dt == "bool":
			typ = "BOOLEAN"
		case dt == "time":
			typ = "DATETIME"
		}
		col := "`" + f.DBName + "` " + typ
		switch {
		case f.PrimaryKey:
			col += " PRIMARY KEY DEFAULT " + sqliteUUID
		case f.DefaultValue != "" && !strings.Contains(f.DefaultValue, "("):
			def := f.DefaultValue
			if typ == "TEXT" { // GORM strips the quotes from string defaults
				def = "'" + strings.Trim(def, "'") + "'"
			}
			col += " DEFAULT " + def
		}
		if f.NotNull {
			col += " NOT NULL"
		}
		cols = append(cols, col)
	}
	if err := db.Exec("CREATE TABLE " + stmt.Schema.Table + " (" + strings.Join(cols, ", ") + ")").Error; err != nil {
		t.Fatal(err)
	}
	return db
}

// annualPlanRouter wires the annual plan routes exactly as routes/handler.go
// does (minus auth, which is not what is under test)
func annualPlanRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/annual-audit-plans")
	g.GET("", crud.List(db, "AuditAnnual", func() interface{} { return &[]models.AuditAnnual{} }))
	g.GET("/:id", crud.GetByID(db, "AuditAnnual", func() interface{} { return &models.AuditAnnual{} }))
	g.POST("", crud.Create(db, "AuditAnnual", func() interface{} { return &models.AuditAnnual{} }))
	g.PUT("/:id", crud.Update(db, "AuditAnnual", func() interface{} { return &models.AuditAnnual{} }))
	return r
}

type annualPlanResponse struct {
	Success bool               `json:"success"`
	Message string             `json:"message"`
	Data    models.AuditAnnual `json:"data"`
}

// callAnnualPlan sends one request and returns the status, decoded body and
// raw body. Every body is checked for the frontend's Q1 workload message.
func callAnnualPlan(t *testing.T, r *gin.Engine, method, path, body string) (int, annualPlanResponse, string) {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	raw := w.Body.String()
	assertNoQ1WorkloadMessage(t, method+" "+path, raw)
	var resp annualPlanResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("%s %s: bad JSON %q: %v", method, path, raw, err)
	}
	return w.Code, resp, raw
}

func assertNoQ1WorkloadMessage(t *testing.T, where, body string) {
	t.Helper()
	for _, s := range []string{"Triwulan", "40%", "ratakan"} {
		if strings.Contains(body, s) {
			t.Errorf("%s: response contains %q: %s", where, s, body)
		}
	}
}

func storedAnnualPlan(t *testing.T, db *gorm.DB, id string) models.AuditAnnual {
	t.Helper()
	var p models.AuditAnnual
	if err := db.First(&p, "id = ?", id).Error; err != nil {
		t.Fatalf("plan %s not stored: %v", id, err)
	}
	return p
}

func assertSchedule(t *testing.T, where string, p models.AuditAnnual, months []int, quarters []string, year int) {
	t.Helper()
	if !reflect.DeepEqual(p.SelectedMonths, months) {
		t.Errorf("%s: selectedMonths = %v, want %v", where, p.SelectedMonths, months)
	}
	if !reflect.DeepEqual(p.Quarters, quarters) {
		t.Errorf("%s: quarters = %v, want %v", where, p.Quarters, quarters)
	}
	if p.Year != year {
		t.Errorf("%s: year = %d, want %d", where, p.Year, year)
	}
}

// createPayload mirrors the body built by addPlan in frontend/stores/annual-audit.ts
func createPayload(t *testing.T, months []int, quarters []string, year int) string {
	t.Helper()
	b, err := json.Marshal(map[string]interface{}{
		"code":            "PKAT-TEST-001",
		"version":         "v1.0",
		"revisionHistory": []interface{}{},
		"activities": []map[string]string{
			{"name": "Financial Audit Q1", "category": "Assurance", "department": "Finance", "riskName": "Payment Overrides", "riskLevel": "High"},
		},
		"status":               "DRAFT",
		"selectedMonths":       months,
		"quarters":             quarters,
		"auditorCount":         3,
		"daysPerAuditor":       10,
		"totalMandays":         30,
		"supervisorId":         "S01",
		"supervisorName":       "Budi Santoso (Mgr)",
		"notes":                "Q1-heavy schedule",
		"year":                 year,
		"attachmentCategory":   "Plan",
		"attachments":          []interface{}{},
		"attachmentUploadedBy": "Tester",
		"attachmentUploadDate": "2026-01-05",
		"isActive":             true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestAnnualPlanCreateAcceptsQ1HeavySchedule(t *testing.T) {
	cases := []struct {
		name     string
		months   []int
		quarters []string
		rawJSON  string // exact selectedMonths fragment expected in the response
	}{
		// 3 of 4 months in Q1 (75%), 0-based as the frontend sends them;
		// month 0 (January) must not be dropped or treated as empty
		{"0-based Jan-Apr", []int{0, 1, 2, 3}, []string{"Q1", "Q2"}, `"selectedMonths":[0,1,2,3]`},
		// same schedule, 1-based as cmd/seed.go stores it
		{"1-based Jan-Apr", []int{1, 2, 3, 4}, []string{"Q1", "Q2"}, `"selectedMonths":[1,2,3,4]`},
		// everything in Q1 (100%)
		{"0-based Q1 only", []int{0, 1, 2}, []string{"Q1"}, `"selectedMonths":[0,1,2]`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := annualPlanDB(t)
			r := annualPlanRouter(db)

			code, resp, raw := callAnnualPlan(t, r, http.MethodPost, "/annual-audit-plans",
				createPayload(t, tc.months, tc.quarters, 2026))
			if code != http.StatusCreated || !resp.Success {
				t.Fatalf("POST: got %d, want 201: %s", code, raw)
			}
			if !strings.Contains(raw, tc.rawJSON) || !strings.Contains(raw, `"quarters":["Q1"`) {
				t.Errorf("POST response does not echo the schedule unchanged: %s", raw)
			}
			assertSchedule(t, "POST response", resp.Data, tc.months, tc.quarters, 2026)

			id := resp.Data.ID.String()
			if resp.Data.ID == [16]byte{} {
				t.Fatalf("no id assigned: %s", raw)
			}
			assertSchedule(t, "stored", storedAnnualPlan(t, db, id), tc.months, tc.quarters, 2026)

			code, resp, raw = callAnnualPlan(t, r, http.MethodGet, "/annual-audit-plans/"+id, "")
			if code != http.StatusOK {
				t.Fatalf("GET: got %d: %s", code, raw)
			}
			if !strings.Contains(raw, tc.rawJSON) {
				t.Errorf("GET response lost the schedule: %s", raw)
			}
			assertSchedule(t, "GET", resp.Data, tc.months, tc.quarters, 2026)
		})
	}
}

func TestAnnualPlanUpdateToQ1OnlyIsPersisted(t *testing.T) {
	// Today is past most of 2026, so 2026 Q1 is in the past: the backend
	// does not care either way
	for _, year := range []int{2026, 2027} {
		t.Run(strconv.Itoa(year), func(t *testing.T) {
			db := annualPlanDB(t)
			r := annualPlanRouter(db)

			// start from an evenly spread plan
			code, resp, raw := callAnnualPlan(t, r, http.MethodPost, "/annual-audit-plans",
				createPayload(t, []int{0, 3, 6, 9}, []string{"Q1", "Q2", "Q3", "Q4"}, 2025))
			if code != http.StatusCreated {
				t.Fatalf("POST: got %d: %s", code, raw)
			}
			id := resp.Data.ID.String()

			// PUT body mirrors updatePlan in frontend/stores/annual-audit.ts:
			// the whole form spread in, including keys that are not columns
			// (file, id); numbers arrive as float64 in crud.Update's map
			body, err := json.Marshal(map[string]interface{}{
				"id":             id,
				"code":           "PKAT-TEST-001",
				"status":         "DRAFT",
				"selectedMonths": []int{0, 1, 2},
				"quarters":       []string{"Q1"},
				"auditorCount":   3,
				"daysPerAuditor": 10,
				"totalMandays":   30,
				"supervisorId":   "S01",
				"supervisorName": "Budi Santoso (Mgr)",
				"notes":          "moved everything to Q1",
				"year":           year,
				"activities":     []map[string]string{{"name": "Financial Audit Q1"}},
				"attachments":    []interface{}{},
				"file":           []interface{}{},
				"isActive":       true,
			})
			if err != nil {
				t.Fatal(err)
			}
			code, resp, raw = callAnnualPlan(t, r, http.MethodPut, "/annual-audit-plans/"+id, string(body))
			if code != http.StatusOK || !resp.Success {
				t.Fatalf("PUT: got %d, want 200: %s", code, raw)
			}
			assertSchedule(t, "PUT response", resp.Data, []int{0, 1, 2}, []string{"Q1"}, year)
			assertSchedule(t, "stored", storedAnnualPlan(t, db, id), []int{0, 1, 2}, []string{"Q1"}, year)

			code, resp, raw = callAnnualPlan(t, r, http.MethodGet, "/annual-audit-plans/"+id, "")
			if code != http.StatusOK {
				t.Fatalf("GET: got %d: %s", code, raw)
			}
			if !strings.Contains(raw, `"selectedMonths":[0,1,2]`) || !strings.Contains(raw, `"quarters":["Q1"]`) {
				t.Errorf("GET response does not hold the Q1-only schedule: %s", raw)
			}
			assertSchedule(t, "GET", resp.Data, []int{0, 1, 2}, []string{"Q1"}, year)
			if resp.Data.Notes != "moved everything to Q1" {
				t.Errorf("notes not updated: %q", resp.Data.Notes)
			}
		})
	}
}

// The workload message must not exist anywhere in the service's code or
// config, so removing it from the frontend is the whole fix. Test files are
// skipped because this one names the phrases. Plain "Triwulan" is not checked
// here: seed data legitimately uses it in SOP names and narratives.
func TestAnnualPlanBackendHasNoQ1WorkloadRule(t *testing.T) {
	phrases := []string{"terlalu tinggi", "ratakan jadwal", "beban kerja triwulan", "40%"}
	root := ".." // the audit-service module root
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("module root not found from %s: %v", root, err)
	}
	scanned := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "vendor", "node_modules", "uploads":
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}
		switch filepath.Ext(path) {
		case ".go", ".json", ".yaml", ".yml", ".sql", ".tmpl", ".toml":
		default:
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		scanned++
		text := strings.ToLower(string(b))
		for _, p := range phrases {
			if strings.Contains(text, p) {
				t.Errorf("%s contains %q", path, p)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if scanned == 0 {
		t.Fatal("no source files scanned")
	}
}
