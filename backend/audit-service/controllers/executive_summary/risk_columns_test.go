package executivesummary

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"audit-service/controllers/crud"
	"audit-service/models"
	"audit-service/pkg/testdb"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type riskCounts struct{ High, ModHigh, Mod, LowMod, Low, Tinggi, Sedang, Rendah int }

func countsOf(e models.ExecutiveSummary) riskCounts {
	return riskCounts{e.RisikoHigh, e.RisikoModerateHigh, e.RisikoModerate, e.RisikoLowModerate, e.RisikoLow, e.RisikoTinggi, e.RisikoSedang, e.RisikoRendah}
}

func loadSummary(t *testing.T, db *gorm.DB, id uuid.UUID) models.ExecutiveSummary {
	t.Helper()
	var e models.ExecutiveSummary
	if err := db.Unscoped().First(&e, "id = ?", id).Error; err != nil {
		t.Fatal(err)
	}
	return e
}

// createLegacy inserts a row as an old deploy would have: no hooks, five-level columns 0
func createLegacy(t *testing.T, db *gorm.DB, tinggi, sedang, rendah int) models.ExecutiveSummary {
	t.Helper()
	ts := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	e := models.ExecutiveSummary{ID: uuid.New(), RisikoTinggi: tinggi, RisikoSedang: sedang, RisikoRendah: rendah, CreatedAt: ts, UpdatedAt: ts}
	if err := db.Session(&gorm.Session{SkipHooks: true}).Create(&e).Error; err != nil {
		t.Fatal(err)
	}
	return e
}

func TestNormalizeExecutiveSummaryRiskCounts(t *testing.T) {
	db := testdb.New(t, &models.ExecutiveSummary{})

	legacy := createLegacy(t, db, 3, 2, 1)
	deleted := createLegacy(t, db, 0, 0, 4)
	if err := db.Delete(&deleted).Error; err != nil {
		t.Fatal(err)
	}
	empty := createLegacy(t, db, 0, 0, 0)
	// Already on five levels: untouched even though legacy counts differ
	five := createLegacy(t, db, 9, 9, 9)
	if err := db.Model(&models.ExecutiveSummary{}).Where("id = ?", five.ID).UpdateColumn("risiko_moderate_high", 1).Error; err != nil {
		t.Fatal(err)
	}

	n, err := models.NormalizeExecutiveSummaryRiskCounts(db)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("rows normalized = %d, want 2", n)
	}
	got := loadSummary(t, db, legacy.ID)
	if c := countsOf(got); c != (riskCounts{3, 0, 2, 0, 1, 3, 2, 1}) {
		t.Errorf("legacy row = %+v", c)
	}
	if !got.UpdatedAt.Equal(legacy.UpdatedAt) {
		t.Errorf("updated_at changed: %v -> %v", legacy.UpdatedAt, got.UpdatedAt)
	}
	if c := countsOf(loadSummary(t, db, deleted.ID)); c != (riskCounts{0, 0, 0, 0, 4, 0, 0, 4}) {
		t.Errorf("soft-deleted row = %+v", c)
	}
	if c := countsOf(loadSummary(t, db, empty.ID)); c != (riskCounts{}) {
		t.Errorf("empty row = %+v", c)
	}
	if c := countsOf(loadSummary(t, db, five.ID)); c != (riskCounts{0, 1, 0, 0, 0, 9, 9, 9}) {
		t.Errorf("five-level row = %+v", c)
	}

	// Second run is a no-op
	if n, err := models.NormalizeExecutiveSummaryRiskCounts(db); err != nil || n != 0 {
		t.Errorf("second run: n=%d err=%v", n, err)
	}
}

func TestExecutiveSummaryBeforeCreateSyncsRiskCounts(t *testing.T) {
	db := testdb.New(t, &models.ExecutiveSummary{})

	five := models.ExecutiveSummary{RisikoHigh: 2, RisikoModerateHigh: 1, RisikoModerate: 3, RisikoLowModerate: 4, RisikoLow: 5, RisikoTinggi: 99}
	if err := db.Create(&five).Error; err != nil {
		t.Fatal(err)
	}
	if c := countsOf(loadSummary(t, db, five.ID)); c != (riskCounts{2, 1, 3, 4, 5, 3, 3, 9}) {
		t.Errorf("five-level create = %+v", c)
	}

	legacy := models.ExecutiveSummary{RisikoTinggi: 3, RisikoSedang: 2, RisikoRendah: 1}
	if err := db.Create(&legacy).Error; err != nil {
		t.Fatal(err)
	}
	if c := countsOf(loadSummary(t, db, legacy.ID)); c != (riskCounts{3, 0, 2, 0, 1, 3, 2, 1}) {
		t.Errorf("legacy create = %+v", c)
	}
}

func TestExecutiveSummaryUpdateSyncsRiskCounts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := testdb.New(t, &models.ExecutiveSummary{})
	r := gin.New()
	r.PUT("/executive-summaries/:id", crud.UpdateWithHook(db, "ExecutiveSummary", func() interface{} { return &models.ExecutiveSummary{} }, SyncRiskColumnsBeforeUpdate))

	put := func(id uuid.UUID, payload map[string]interface{}) {
		t.Helper()
		b, _ := json.Marshal(payload)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/executive-summaries/"+id.String(), bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("PUT %v: status %d: %s", payload, w.Code, w.Body.String())
		}
	}

	e := createLegacy(t, db, 3, 2, 1)
	if _, err := models.NormalizeExecutiveSummaryRiskCounts(db); err != nil {
		t.Fatal(err)
	}

	// Partial five-level payload: the others keep their stored values, null is 0
	put(e.ID, map[string]interface{}{"risikoModerateHigh": 4, "risikoLowModerate": nil, "risikoLow": "2"})
	if c := countsOf(loadSummary(t, db, e.ID)); c != (riskCounts{3, 4, 2, 0, 2, 7, 2, 2}) {
		t.Errorf("after five-level PUT = %+v", c)
	}

	// Legacy-only payload from an old client
	put(e.ID, map[string]interface{}{"risikoTinggi": 5})
	if c := countsOf(loadSummary(t, db, e.ID)); c != (riskCounts{5, 0, 2, 0, 2, 5, 2, 2}) {
		t.Errorf("after legacy PUT = %+v", c)
	}

	// Saving all five as 0 clears the legacy counts too, so the startup
	// normalization does not bring the old values back
	put(e.ID, map[string]interface{}{"risikoHigh": 0, "risikoModerateHigh": 0, "risikoModerate": 0, "risikoLowModerate": 0, "risikoLow": 0})
	if n, err := models.NormalizeExecutiveSummaryRiskCounts(db); err != nil || n != 0 {
		t.Errorf("normalization after all-zero save: n=%d err=%v", n, err)
	}
	if c := countsOf(loadSummary(t, db, e.ID)); c != (riskCounts{}) {
		t.Errorf("after all-zero PUT = %+v", c)
	}

	// Unrelated payload leaves the counts alone
	put(e.ID, map[string]interface{}{"narrative": "x"})
	if c := countsOf(loadSummary(t, db, e.ID)); c != (riskCounts{}) {
		t.Errorf("after narrative PUT = %+v", c)
	}
}

func TestExecutiveSummaryJSONKeys(t *testing.T) {
	b, err := json.Marshal(models.ExecutiveSummary{})
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	for _, k := range []string{"risikoHigh", "risikoModerateHigh", "risikoModerate", "risikoLowModerate", "risikoLow", "risikoTinggi", "risikoSedang", "risikoRendah"} {
		if _, ok := m[k]; !ok {
			t.Errorf("ExecutiveSummary JSON missing %q", k)
		}
	}
}
