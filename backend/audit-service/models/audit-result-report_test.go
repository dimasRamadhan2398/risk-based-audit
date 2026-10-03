package models_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"audit-service/controllers/crud"
	"audit-service/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// arrDB creates audit_result_reports from the GORM schema (AutoMigrate cannot
// run on sqlite because ids default to gen_random_uuid())
func arrDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(&models.AuditResultReport{}); err != nil {
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
		case dt == "time":
			typ = "DATETIME"
		}
		cols = append(cols, "`"+f.DBName+"` "+typ)
	}
	if err := db.Exec("CREATE TABLE audit_result_reports (" + strings.Join(cols, ", ") + ")").Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func storedCount(t *testing.T, db *gorm.DB, id uuid.UUID) (int, int) {
	t.Helper()
	var r models.AuditResultReport
	if err := db.First(&r, "id = ?", id).Error; err != nil {
		t.Fatal(err)
	}
	return r.FindingsCount, len(r.Findings)
}

func TestFindingsCountHookDirect(t *testing.T) {
	db := arrDB(t)
	r := &models.AuditResultReport{ID: uuid.New(), FindingsCount: 99,
		Findings: []models.AuditReportFinding{{Title: "a"}, {Title: "b"}}}
	if err := db.Create(r).Error; err != nil {
		t.Fatal(err)
	}
	if c, n := storedCount(t, db, r.ID); c != 2 || n != 2 {
		t.Fatalf("after create: count %d, findings %d", c, n)
	}

	r.Findings = r.Findings[:1]
	r.FindingsCount = 7
	if err := db.Save(r).Error; err != nil {
		t.Fatal(err)
	}
	if c, _ := storedCount(t, db, r.ID); c != 1 {
		t.Fatalf("after save: count %d", c)
	}

	// Map update that only touches findings_count is reset to the stored length
	if err := db.Model(r).Updates(map[string]interface{}{"findings_count": 42}).Error; err != nil {
		t.Fatal(err)
	}
	if c, _ := storedCount(t, db, r.ID); c != 1 {
		t.Fatalf("after findings_count-only update: count %d", c)
	}
	// Map update without findings/findings_count leaves the count alone
	if err := db.Model(r).Updates(map[string]interface{}{"title": "x"}).Error; err != nil {
		t.Fatal(err)
	}
	if c, _ := storedCount(t, db, r.ID); c != 1 {
		t.Fatalf("after title update: count %d", c)
	}
	// Clearing findings
	if err := db.Model(r).Updates(map[string]interface{}{"findings": nil, "findings_count": 5}).Error; err != nil {
		t.Fatal(err)
	}
	if c, n := storedCount(t, db, r.ID); c != 0 || n != 0 {
		t.Fatalf("after clearing: count %d, findings %d", c, n)
	}
}

// Through the generic handlers wired for /audit-result-reports
func TestFindingsCountHookViaCrud(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := arrDB(t)
	newOne := func() interface{} { return &models.AuditResultReport{} }
	r := gin.New()
	r.POST("/arr", crud.Create(db, "AuditResultReport", newOne))
	r.PUT("/arr/:id", crud.Update(db, "AuditResultReport", newOne))

	do := func(method, path, body string) map[string]interface{} {
		t.Helper()
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		if w.Code >= 300 {
			t.Fatalf("%s %s: %d %s", method, path, w.Code, w.Body.String())
		}
		var resp struct {
			Data map[string]interface{} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		return resp.Data
	}

	id := uuid.New()
	created := do(http.MethodPost, "/arr", `{"id":"`+id.String()+`","findingsCount":10,"findings":[{"title":"a"},{"title":"b"},{"title":"c"}]}`)
	if created["findingsCount"] != float64(3) {
		t.Fatalf("create response findingsCount = %v", created["findingsCount"])
	}
	if c, n := storedCount(t, db, id); c != 3 || n != 3 {
		t.Fatalf("after POST: count %d, findings %d", c, n)
	}

	updated := do(http.MethodPut, "/arr/"+id.String(), `{"findingsCount":0,"findings":[{"title":"only"}]}`)
	if updated["findingsCount"] != float64(1) {
		t.Fatalf("update response findingsCount = %v", updated["findingsCount"])
	}
	if c, n := storedCount(t, db, id); c != 1 || n != 1 {
		t.Fatalf("after PUT: count %d, findings %d", c, n)
	}

	do(http.MethodPut, "/arr/"+id.String(), `{"findingsCount":9}`)
	if c, _ := storedCount(t, db, id); c != 1 {
		t.Fatalf("after count-only PUT: count %d", c)
	}

	do(http.MethodPut, "/arr/"+id.String(), `{"findings":[]}`)
	if c, n := storedCount(t, db, id); c != 0 || n != 0 {
		t.Fatalf("after empty findings PUT: count %d, findings %d", c, n)
	}
}
