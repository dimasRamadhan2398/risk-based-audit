package annualplan_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"

	ctrlAnnualPlan "audit-service/controllers/annual_plan"
	"audit-service/controllers/crud"
	"audit-service/models"
	"audit-service/pkg/activitycode/sqlitetest"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func planRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := sqlitetest.Open(t, &models.ActivityCodeSequence{}, &models.AuditAnnual{}, &models.AnnualSnapshotRef{})
	db.Logger = logger.Discard
	r := gin.New()
	g := r.Group("/annual-audit-plans")
	g.POST("", crud.CreateWithHook(db, "AuditAnnual", func() interface{} { return &models.AuditAnnual{} }, ctrlAnnualPlan.AssignCodeOnCreate))
	g.DELETE("/:id", crud.Delete(db, "AuditAnnual", func() interface{} { return &models.AuditAnnual{} }))
	return r, db
}

// post creates a plan through the route and returns the status and the code
// the server stored.
func post(t *testing.T, r *gin.Engine, body string) (int, string) {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/annual-audit-plans", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var env struct {
		Data models.AuditAnnual `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &env)
	return w.Code, env.Data.Code
}

func planBody(year int, clientCode string) string {
	return fmt.Sprintf(`{"code":%q,"year":%d,"status":"DRAFT"}`, clientCode, year)
}

func mustPost(t *testing.T, r *gin.Engine, year int) string {
	t.Helper()
	code, got := post(t, r, planBody(year, ""))
	if code != http.StatusCreated {
		t.Fatalf("POST year %d: status %d", year, code)
	}
	return got
}

func softDelete(t *testing.T, r *gin.Engine, code string, db *gorm.DB) {
	t.Helper()
	var p models.AuditAnnual
	if err := db.Where("code = ?", code).First(&p).Error; err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/annual-audit-plans/"+p.ID.String(), nil))
	if w.Code != http.StatusOK {
		t.Fatalf("DELETE %s: status %d: %s", code, w.Code, w.Body.String())
	}
}

func TestCreateAssignsServerCodeAndIgnoresClientCode(t *testing.T) {
	r, _ := planRouter(t)
	code, got := post(t, r, planBody(2026, "PKAT-2026-ASR-999"))
	if code != http.StatusCreated {
		t.Fatalf("status %d", code)
	}
	if got != "PKAT-2026-001" {
		t.Errorf("code = %q, want PKAT-2026-001", got)
	}
}

func TestTwoCreatesInSameYearGetConsecutiveNumbers(t *testing.T) {
	r, _ := planRouter(t)
	if got := mustPost(t, r, 2026); got != "PKAT-2026-001" {
		t.Errorf("first = %q", got)
	}
	if got := mustPost(t, r, 2026); got != "PKAT-2026-002" {
		t.Errorf("second = %q", got)
	}
}

func TestNewYearRestartsAtOne(t *testing.T) {
	r, _ := planRouter(t)
	mustPost(t, r, 2026)
	mustPost(t, r, 2026)
	if got := mustPost(t, r, 2027); got != "PKAT-2027-001" {
		t.Errorf("2027 = %q, want PKAT-2027-001", got)
	}
	if got := mustPost(t, r, 2026); got != "PKAT-2026-003" {
		t.Errorf("2026 after 2027 = %q, want PKAT-2026-003", got)
	}
}

func TestSoftDeletedPlanKeepsItsNumber(t *testing.T) {
	r, db := planRouter(t)
	first := mustPost(t, r, 2026)  // 001
	second := mustPost(t, r, 2026) // 002
	softDelete(t, r, second, db)
	if got := mustPost(t, r, 2026); got != "PKAT-2026-003" {
		t.Errorf("after soft delete of 002 got %q, want PKAT-2026-003", got)
	}
	// hard delete of the first one must not free 001 either
	if err := db.Unscoped().Where("code = ?", first).Delete(&models.AuditAnnual{}).Error; err != nil {
		t.Fatal(err)
	}
	if got := mustPost(t, r, 2026); got != "PKAT-2026-004" {
		t.Errorf("after hard delete got %q, want PKAT-2026-004", got)
	}
}

func TestCodesStoredOutsideTheCounterAreRespected(t *testing.T) {
	r, db := planRouter(t)
	// seed-style row, a soft-deleted higher one, and a legacy category-segment code
	for _, c := range []string{"PKAT-2026-007", "PKAT-2025-001", "PKAT-2026-ASR-001"} {
		if err := db.Create(&models.AuditAnnual{Code: c, Year: 2026}).Error; err != nil {
			t.Fatal(err)
		}
	}
	gone := &models.AuditAnnual{Code: "PKAT-2026-020", Year: 2026}
	if err := db.Create(gone).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Delete(gone).Error; err != nil {
		t.Fatal(err)
	}
	if got := mustPost(t, r, 2026); got != "PKAT-2026-021" {
		t.Errorf("2026 got %q, want PKAT-2026-021", got)
	}
	if got := mustPost(t, r, 2025); got != "PKAT-2025-002" {
		t.Errorf("2025 got %q, want PKAT-2025-002", got)
	}
}

func TestRejectsMissingOrInvalidYear(t *testing.T) {
	r, db := planRouter(t)
	for _, body := range []string{`{"status":"DRAFT"}`, `{"year":0}`, `{"year":26}`} {
		if code, _ := post(t, r, body); code != http.StatusBadRequest {
			t.Errorf("POST %s: status %d, want 400", body, code)
		}
	}
	var n int64
	db.Model(&models.AuditAnnual{}).Count(&n)
	if n != 0 {
		t.Errorf("%d plans stored for invalid input", n)
	}
}

func TestConcurrentCreatesGetDistinctCodes(t *testing.T) {
	r, _ := planRouter(t)
	const n = 20
	codes := make([]string, n)
	statuses := make([]int, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			statuses[i], codes[i] = post(t, r, planBody(2026, ""))
		}(i)
	}
	wg.Wait()
	for i, s := range statuses {
		if s != http.StatusCreated {
			t.Fatalf("worker %d: status %d", i, s)
		}
	}
	sort.Strings(codes)
	for i, c := range codes {
		if want := fmt.Sprintf("PKAT-2026-%03d", i+1); c != want {
			t.Fatalf("codes = %v; position %d is %q, want %q", codes, i, c, want)
		}
	}
}

func TestPartialUniqueIndexOnLivePlanCodes(t *testing.T) {
	_, db := planRouter(t)
	first := &models.AuditAnnual{Code: "PKAT-2026-001", Year: 2026}
	if err := db.Create(first).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&models.AuditAnnual{Code: "PKAT-2026-001", Year: 2026}).Error; err == nil {
		t.Fatal("duplicate live code accepted")
	}
	if err := db.Delete(first).Error; err != nil {
		t.Fatal(err)
	}
	// the soft-deleted row no longer holds the code
	if err := db.Create(&models.AuditAnnual{Code: "PKAT-2026-001", Year: 2026}).Error; err != nil {
		t.Fatalf("code not reusable after soft delete (index must be partial): %v", err)
	}
}

// Plans saved before codes existed have an empty code; several of them must
// coexist with the unique index.
func TestLegacyEmptyCodesDoNotCollide(t *testing.T) {
	_, db := planRouter(t)
	for i := 0; i < 2; i++ {
		if err := db.Create(&models.AuditAnnual{Year: 2024}).Error; err != nil {
			t.Fatalf("empty code row %d rejected: %v", i, err)
		}
	}
}

// The model declares the index as partial; check the schema the migration
// is built from, so a dropped where: tag fails here.
func TestPlanCodeIndexDeclaration(t *testing.T) {
	sch, err := schema.Parse(&models.AuditAnnual{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	for _, idx := range sch.ParseIndexes() {
		if idx.Name == "idx_audit_annuals_code_live" {
			// empty codes (rows from before the code was assigned) must not collide
			if idx.Class != "UNIQUE" || idx.Where != "deleted_at IS NULL AND code <> ''" {
				t.Errorf("index = class %q where %q", idx.Class, idx.Where)
			}
			return
		}
	}
	t.Fatal("idx_audit_annuals_code_live not declared on AuditAnnual.Code")
}
