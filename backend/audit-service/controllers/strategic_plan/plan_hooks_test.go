package strategicplan

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"audit-service/controllers/crud"
	"audit-service/models"
	"audit-service/pkg/activitycode/sqlitetest"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// planRouter serves POST/PUT/DELETE /strategic-plans as routes/handler.go does.
func planRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := sqlitetest.Open(t, &models.ActivityCodeSequence{}, &models.StrategicPlan{})
	newPlan := func() interface{} { return &models.StrategicPlan{} }
	r := gin.New()
	g := r.Group("/strategic-plans")
	g.POST("", ValidateCategory(), crud.CreateWithHook(db, "StrategicPlan", newPlan, PrepareOnCreate))
	g.PUT("/:id", ValidateCategory(), crud.UpdateWithHook(db, "StrategicPlan", newPlan, PrepareOnUpdate))
	g.DELETE("/:id", crud.Delete(db, "StrategicPlan", newPlan))
	return r, db
}

func doPlan(t *testing.T, r *gin.Engine, method, path, body string) (int, models.StrategicPlan, string) {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var env struct {
		Data models.StrategicPlan `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &env)
	return w.Code, env.Data, w.Body.String()
}

const validPlan = `{"code":"SO-IA01","strategicObjective":"SO","yearStart":2026,"yearEnd":2028,` +
	`"kpiTargets":{"2026":"90","2027":"92.5","2028":""},"kpiActuals":{"2026":"88"},"target":"90","actual":"88"}`

func TestCreateAssignsCodeAboveSeededCodes(t *testing.T) {
	r, db := planRouter(t)
	// Seed data fills SO-IA01..SO-IA100; one of them soft-deleted, plus a non-SO-IA code.
	for _, c := range []string{"SO-IA01", "SO-IA100", "SO-IA07", "CUSTOM-1"} {
		if err := db.Create(&models.StrategicPlan{Code: c}).Error; err != nil {
			t.Fatal(err)
		}
	}
	db.Where("code = ?", "SO-IA100").Delete(&models.StrategicPlan{})

	var got []string
	for i := 0; i < 3; i++ {
		status, plan, body := doPlan(t, r, http.MethodPost, "/strategic-plans", validPlan)
		if status != http.StatusCreated {
			t.Fatalf("POST: %d %s", status, body)
		}
		got = append(got, plan.Code)
	}
	want := []string{"SO-IA101", "SO-IA102", "SO-IA103"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("codes = %v, want %v (client code must be ignored)", got, want)
		}
	}
}

func TestCreateCodesAreUniqueUnderConcurrency(t *testing.T) {
	r, _ := planRouter(t)
	const n = 10
	codes := make(chan string, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			status, plan, body := doPlan(t, r, http.MethodPost, "/strategic-plans", validPlan)
			if status != http.StatusCreated {
				t.Errorf("POST: %d %s", status, body)
			}
			codes <- plan.Code
		}()
	}
	wg.Wait()
	close(codes)
	seen := map[string]bool{}
	for c := range codes {
		if seen[c] {
			t.Fatalf("duplicate code %s", c)
		}
		seen[c] = true
	}
}

func TestCreateRejectsInvalidInput(t *testing.T) {
	r, db := planRouter(t)
	cases := map[string]string{
		"end before start":    `{"yearStart":2028,"yearEnd":2026}`,
		"text target":         `{"yearStart":2026,"yearEnd":2027,"kpiTargets":{"2026":"ninety"}}`,
		"percent realization": `{"yearStart":2026,"yearEnd":2027,"kpiActuals":{"2026":"88%"}}`,
		"rupiah target":       `{"target":"Rp 1.000.000"}`,
	}
	for name, body := range cases {
		status, _, resp := doPlan(t, r, http.MethodPost, "/strategic-plans", body)
		if status != http.StatusBadRequest {
			t.Errorf("%s: status %d, want 400: %s", name, status, resp)
		}
	}
	var count int64
	db.Model(&models.StrategicPlan{}).Count(&count)
	if count != 0 {
		t.Fatalf("%d rows stored after rejected creates", count)
	}
}

func TestUpdateValidatesAndKeepsCode(t *testing.T) {
	r, _ := planRouter(t)
	_, plan, _ := doPlan(t, r, http.MethodPost, "/strategic-plans", validPlan)
	path := "/strategic-plans/" + plan.ID.String()

	// Only yearEnd sent: compared with the stored yearStart (2026).
	if status, _, body := doPlan(t, r, http.MethodPut, path, `{"yearEnd":2025}`); status != http.StatusBadRequest {
		t.Fatalf("yearEnd before stored yearStart: %d %s", status, body)
	}
	if status, _, body := doPlan(t, r, http.MethodPut, path, `{"kpiActuals":{"2026":"abc"}}`); status != http.StatusBadRequest {
		t.Fatalf("text realization: %d %s", status, body)
	}

	status, updated, body := doPlan(t, r, http.MethodPut, path, `{"code":"SO-IA999","yearStart":2027,"yearEnd":2027,"kpiTargets":{"2027":"-1.5"}}`)
	if status != http.StatusOK {
		t.Fatalf("valid update: %d %s", status, body)
	}
	if updated.Code != plan.Code {
		t.Fatalf("code changed to %s, want %s", updated.Code, plan.Code)
	}
	if updated.YearStart != 2027 || updated.KPITargets["2027"] != "-1.5" {
		t.Fatalf("update not applied: %+v", updated)
	}
}
