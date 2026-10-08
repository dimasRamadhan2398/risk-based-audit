package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"

	"audit-service/models"
	"audit-service/pkg/activitycode/sqlitetest"
	"audit-service/repositories"
	svcActivity "audit-service/services/audit_activity"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func activityRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	db := sqlitetest.Open(t, &models.ActivityCodeSequence{}, &models.AuditAnnual{}, &models.AuditActivity{}, &models.ActivityPlan{})
	repo := repositories.NewAuditActivityRepository(repositories.NewBaseRepository(db))
	ctrl := NewAuditActivityController(svcActivity.NewAuditActivityService(repo))
	r := gin.New()
	g := r.Group("/audit-activities")
	g.POST("", ctrl.CreateActivity)
	g.PUT("/:id", ctrl.UpdateActivity)
	g.GET("/:id", ctrl.GetActivity)
	return r, db
}

func annualPlan(t *testing.T, db *gorm.DB, year int) uuid.UUID {
	t.Helper()
	p := &models.AuditAnnual{Year: year}
	if err := db.Create(p).Error; err != nil {
		t.Fatal(err)
	}
	return p.ID
}

type activityResp struct {
	ID          string `json:"id"`
	ProjectCode string `json:"project_code"`
	AuditType   string `json:"audit_type"`
	Title       string `json:"title"`
}

func send(t *testing.T, r *gin.Engine, method, path, body string) (int, activityResp) {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var env struct {
		Data activityResp `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &env)
	return w.Code, env.Data
}

func createBody(plan uuid.UUID, extra string) string {
	return fmt.Sprintf(`{"annual_plan_id": %q, "target_unit_id": %q, "title": "Audit",
		"planned_start": "2026-01-15T00:00:00Z", "planned_end": "2026-03-31T00:00:00Z"%s}`, plan, uuid.New(), extra)
}

func TestCreateActivityGeneratesProjectCode(t *testing.T) {
	r, db := activityRouter(t)
	plan2026 := annualPlan(t, db, 2026)
	plan2025 := annualPlan(t, db, 2025)

	steps := []struct {
		plan  uuid.UUID
		extra string
		want  string
	}{
		{plan2026, `, "audit_type": "Assurance"`, "ASR-2026-001"},                               // no project_code
		{plan2026, `, "audit_type": "Assurance", "project_code": "MY-OWN-CODE"`, "ASR-2026-002"}, // client value ignored
		{plan2026, `, "audit_type": "Assurance", "project_code": "ASR-2026-001"`, "ASR-2026-003"},
		{plan2026, `, "audit_type": "Consulting Services"`, "CNS-2026-001"},
		{plan2026, ``, "AUD-2026-001"},
		// the annual plan's year, not the planned dates or today's date
		{plan2025, `, "audit_type": "Assurance"`, "ASR-2025-001"},
	}
	for i, s := range steps {
		status, got := send(t, r, http.MethodPost, "/audit-activities", createBody(s.plan, s.extra))
		if status != http.StatusCreated {
			t.Fatalf("step %d: status %d", i, status)
		}
		if got.ProjectCode != s.want {
			t.Errorf("step %d: project_code = %q, want %q", i, got.ProjectCode, s.want)
		}
	}
}

func TestCreateActivityValidation(t *testing.T) {
	r, db := activityRouter(t)
	plan := annualPlan(t, db, 2026)

	if status, _ := send(t, r, http.MethodPost, "/audit-activities", createBody(uuid.New(), "")); status != http.StatusBadRequest {
		t.Errorf("unknown annual plan: status %d, want 400", status)
	}
	missingTitle := strings.Replace(createBody(plan, ""), `"title": "Audit",`, "", 1)
	if status, _ := send(t, r, http.MethodPost, "/audit-activities", missingTitle); status != http.StatusBadRequest {
		t.Errorf("missing title: status %d, want 400", status)
	}
	long := `, "audit_type": "` + strings.Repeat("x", 101) + `"`
	if status, _ := send(t, r, http.MethodPost, "/audit-activities", createBody(plan, long)); status != http.StatusBadRequest {
		t.Errorf("audit_type over 100 chars: status %d, want 400", status)
	}
	// nothing was numbered by the failed requests
	_, got := send(t, r, http.MethodPost, "/audit-activities", createBody(plan, `, "audit_type": "Assurance"`))
	if got.ProjectCode != "ASR-2026-001" {
		t.Errorf("project_code = %q, want ASR-2026-001", got.ProjectCode)
	}
}

func TestUpdateActivityKeepsProjectCode(t *testing.T) {
	r, db := activityRouter(t)
	plan := annualPlan(t, db, 2026)
	_, created := send(t, r, http.MethodPost, "/audit-activities", createBody(plan, `, "audit_type": "Assurance"`))

	status, upd := send(t, r, http.MethodPut, "/audit-activities/"+created.ID,
		`{"title": "Renamed", "planned_start": "2027-02-01T00:00:00Z", "project_code": "XYZ-2027-001"}`)
	if status != http.StatusOK {
		t.Fatalf("status %d", status)
	}
	if upd.ProjectCode != created.ProjectCode || upd.Title != "Renamed" {
		t.Errorf("after update: %+v, want code %q", upd, created.ProjectCode)
	}
}

func TestConcurrentCreateActivityCodesAreUnique(t *testing.T) {
	r, db := activityRouter(t)
	plan := annualPlan(t, db, 2026)
	const n = 20
	var wg sync.WaitGroup
	codes := make([]string, n)
	statuses := make([]int, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			statuses[i], _ = func() (int, activityResp) {
				status, got := send(t, r, http.MethodPost, "/audit-activities", createBody(plan, `, "audit_type": "Assurance"`))
				codes[i] = got.ProjectCode
				return status, got
			}()
		}(i)
	}
	wg.Wait()
	for i, s := range statuses {
		if s != http.StatusCreated {
			t.Fatalf("request %d: status %d", i, s)
		}
	}
	sort.Strings(codes)
	for i, c := range codes {
		if want := fmt.Sprintf("ASR-2026-%03d", i+1); c != want {
			t.Fatalf("codes = %v; position %d = %q, want %q", codes, i, c, want)
		}
	}
}
