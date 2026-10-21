package controllers

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"risk-service/models"
	"risk-service/pkg/kafka"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type scoringFixture struct {
	db       *gorm.DB
	router   *gin.Engine
	stdA     models.StandardRiskFactor
	stdB     models.StandardRiskFactor
	cfA      models.CorporateRiskFactor
	cfB      models.CorporateRiskFactor
	e1, e2   models.CorporateAuditUniverse
	scored   models.AuditUniverseYear // E1 2026, scores A=5 B=3
	unscored models.AuditUniverseYear // E2 2026, no scores
}

func mustCreate(t *testing.T, db *gorm.DB, v any) {
	t.Helper()
	if err := db.Create(v).Error; err != nil {
		t.Fatal(err)
	}
}

// newScoringFixture: two factors weighted 50/50, entity E1 scored for 2026
// (A=5, B=3 -> index 80, High), entity E2 established for 2026 and unscored.
func newScoringFixture(t *testing.T) *scoringFixture {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := openTestDB(t)
	now := time.Now()
	f := &scoringFixture{db: db}

	f.stdA = models.StandardRiskFactor{ID: uuid.New(), Name: "Financial Materiality", CreatedAt: now, UpdatedAt: now}
	f.stdB = models.StandardRiskFactor{ID: uuid.New(), Name: "Fraud Exposure", CreatedAt: now, UpdatedAt: now}
	mustCreate(t, db, &f.stdA)
	mustCreate(t, db, &f.stdB)
	f.cfA = models.CorporateRiskFactor{ID: uuid.New(), StandardRiskFactorID: f.stdA.ID, Weight: 0.5, CreatedAt: now, UpdatedAt: now}
	f.cfB = models.CorporateRiskFactor{ID: uuid.New(), StandardRiskFactorID: f.stdB.ID, Weight: 0.5, CreatedAt: now, UpdatedAt: now}
	mustCreate(t, db, &f.cfA)
	mustCreate(t, db, &f.cfB)

	f.e1 = models.CorporateAuditUniverse{ID: uuid.New(), Name: "Finance", CreatedAt: now, UpdatedAt: now}
	f.e2 = models.CorporateAuditUniverse{ID: uuid.New(), Name: "Procurement", CreatedAt: now, UpdatedAt: now}
	mustCreate(t, db, &f.e1)
	mustCreate(t, db, &f.e2)

	scoredAt := now.Add(-time.Hour)
	f.scored = models.AuditUniverseYear{ID: uuid.New(), CorporateAuditUniverseID: f.e1.ID, Year: 2026,
		RiskIndex: 80, RiskLevel: models.RiskLevelHigh, AuditPriority: true, ScoredAt: &scoredAt, CreatedAt: now, UpdatedAt: now}
	f.unscored = models.AuditUniverseYear{ID: uuid.New(), CorporateAuditUniverseID: f.e2.ID, Year: 2026, CreatedAt: now, UpdatedAt: now}
	mustCreate(t, db, &f.scored)
	mustCreate(t, db, &f.unscored)
	mustCreate(t, db, &models.AuditUniverseRiskScore{ID: uuid.New(), AuditUniverseYearID: f.scored.ID, CorporateRiskFactorID: f.cfA.ID, Score: 5, WeightedScore: 2.5, CreatedAt: now, UpdatedAt: now})
	mustCreate(t, db, &models.AuditUniverseRiskScore{ID: uuid.New(), AuditUniverseYearID: f.scored.ID, CorporateRiskFactorID: f.cfB.ID, Score: 3, WeightedScore: 1.5, CreatedAt: now, UpdatedAt: now})

	events := kafka.NewPublisher(db, "risk-service-test")
	rf := NewRiskFactorController(db, events)
	au := NewAuditUniverseController(db, events)
	r := gin.New()
	r.DELETE("/risk-factors/standard/:id", rf.DeleteStandardRiskFactor)
	r.POST("/risk-factors/corporate", rf.UpdateCorporateRiskFactors)
	r.DELETE("/audit-universe/corporate/:id", au.DeleteCorporateNode)
	r.GET("/audit-universe/year/:year", au.ListYearlyAuditUniverse)
	r.POST("/audit-universe/year/:year", au.EstablishYearlyUniverse)
	r.POST("/audit-universe/year/:year/score", au.ScoreYearlyEntity)
	f.router = r
	return f
}

func (f *scoringFixture) do(t *testing.T, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)
	return w
}

func (f *scoringFixture) year(t *testing.T, id uuid.UUID) models.AuditUniverseYear {
	t.Helper()
	var y models.AuditUniverseYear
	if err := f.db.First(&y, "id = ?", id).Error; err != nil {
		t.Fatal(err)
	}
	return y
}

type outboxEvent struct {
	EventType string          `json:"event_type"`
	Data      json.RawMessage `json:"data"`
}

func (f *scoringFixture) outbox(t *testing.T) []outboxEvent {
	t.Helper()
	var rows []models.KafkaOutboxEvent
	if err := f.db.Order("id ASC").Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	out := make([]outboxEvent, 0, len(rows))
	for _, r := range rows {
		var e outboxEvent
		if err := json.Unmarshal([]byte(r.Payload), &e); err != nil {
			t.Fatal(err)
		}
		out = append(out, e)
	}
	return out
}

func assertUnscored(t *testing.T, y models.AuditUniverseYear) {
	t.Helper()
	if y.ScoredAt != nil || y.RiskIndex != 0 || y.RiskLevel != "" || y.AuditPriority {
		t.Fatalf("row %s should be unscored (nil/0/\"\"/false), got scored_at=%v index=%v level=%q priority=%v",
			y.ID, y.ScoredAt, y.RiskIndex, y.RiskLevel, y.AuditPriority)
	}
}

func TestWeightChangeRecomputesScoredRowsOnly(t *testing.T) {
	f := newScoringFixture(t)

	// 50/50 -> 20/80: E1 = (0.2*5 + 0.8*3)/5*100 = 68 -> Medium to High.
	w := f.do(t, http.MethodPost, "/risk-factors/corporate", []map[string]any{
		{"standard_risk_factor_id": f.stdA.ID.String(), "weight": 20},
		{"standard_risk_factor_id": f.stdB.ID.String(), "weight": 80},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}

	got := f.year(t, f.scored.ID)
	if math.Abs(got.RiskIndex-68) > 1e-9 || got.RiskLevel != models.RiskLevelMediumToHigh || !got.AuditPriority || got.ScoredAt == nil {
		t.Fatalf("scored row not recomputed: index=%v level=%q priority=%v scored_at=%v", got.RiskIndex, got.RiskLevel, got.AuditPriority, got.ScoredAt)
	}

	var scores []models.AuditUniverseRiskScore
	f.db.Where("audit_universe_year_id = ?", f.scored.ID).Find(&scores)
	for _, sc := range scores {
		want := map[uuid.UUID]float64{f.cfA.ID: 1.0, f.cfB.ID: 2.4}[sc.CorporateRiskFactorID]
		if math.Abs(sc.WeightedScore-want) > 1e-9 {
			t.Fatalf("weighted_score for factor %s = %v, want %v", sc.CorporateRiskFactorID, sc.WeightedScore, want)
		}
	}

	assertUnscored(t, f.year(t, f.unscored.ID))
	var unscoredScores int64
	f.db.Model(&models.AuditUniverseRiskScore{}).Where("audit_universe_year_id = ?", f.unscored.ID).Count(&unscoredScores)
	if unscoredScores != 0 {
		t.Fatalf("unscored row got %d invented score rows", unscoredScores)
	}

	events := f.outbox(t)
	if len(events) != 1 || events[0].EventType != kafka.EventYearScored {
		t.Fatalf("want exactly one year.scored event (changed scored row only), got %+v", events)
	}
	var d kafka.YearScoreData
	_ = json.Unmarshal(events[0].Data, &d)
	if d.CorporateAuditUniverseID != f.e1.ID.String() || d.Year != 2026 || d.RiskLevel != models.RiskLevelMediumToHigh || !d.AuditPriority {
		t.Fatalf("unexpected year.scored data: %+v", d)
	}
}

func TestRemovingAFactorRecomputesOverRemainingScores(t *testing.T) {
	f := newScoringFixture(t)

	// Drop B entirely: E1 is now scored on A alone (5) -> 100, High.
	w := f.do(t, http.MethodPost, "/risk-factors/corporate", []map[string]any{
		{"standard_risk_factor_id": f.stdA.ID.String(), "weight": 100},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}
	got := f.year(t, f.scored.ID)
	if math.Abs(got.RiskIndex-100) > 1e-9 || got.RiskLevel != models.RiskLevelHigh || !got.AuditPriority {
		t.Fatalf("index=%v level=%q priority=%v, want 100/High/true", got.RiskIndex, got.RiskLevel, got.AuditPriority)
	}
	var n int64
	f.db.Model(&models.AuditUniverseRiskScore{}).Where("corporate_risk_factor_id = ?", f.cfB.ID).Count(&n)
	if n != 0 {
		t.Fatalf("scores of the removed factor were kept: %d", n)
	}
}

func TestAddingAFactorDoesNotInventScores(t *testing.T) {
	f := newScoringFixture(t)
	now := time.Now()
	stdC := models.StandardRiskFactor{ID: uuid.New(), Name: "HSE", CreatedAt: now, UpdatedAt: now}
	mustCreate(t, f.db, &stdC)

	// A=25, B=25, C=50 (C unscored): E1 = (0.25*5+0.25*3)/0.5/5*100 = 80.
	w := f.do(t, http.MethodPost, "/risk-factors/corporate", []map[string]any{
		{"standard_risk_factor_id": f.stdA.ID.String(), "weight": 25},
		{"standard_risk_factor_id": f.stdB.ID.String(), "weight": 25},
		{"standard_risk_factor_id": stdC.ID.String(), "weight": 50},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}
	got := f.year(t, f.scored.ID)
	if math.Abs(got.RiskIndex-80) > 1e-9 || got.RiskLevel != models.RiskLevelHigh {
		t.Fatalf("index=%v level=%q, want 80/High over the scored factors only", got.RiskIndex, got.RiskLevel)
	}
	var n int64
	f.db.Model(&models.AuditUniverseRiskScore{}).Where("audit_universe_year_id = ?", f.scored.ID).Count(&n)
	if n != 2 {
		t.Fatalf("score rows = %d, want 2 (no default score for the new factor)", n)
	}
	if len(f.outbox(t)) != 0 {
		t.Fatal("no published value changed, so no event is expected")
	}
}

func TestScoreYearlyEntitySetsScoredAt(t *testing.T) {
	f := newScoringFixture(t)
	before := time.Now().Add(-time.Second)

	w := f.do(t, http.MethodPost, "/audit-universe/year/2026/score", map[string]any{
		"audit_universe_year_id": f.unscored.ID.String(),
		"scores": []map[string]any{
			{"corporate_risk_factor_id": f.cfA.ID.String(), "score": 4},
			{"corporate_risk_factor_id": f.cfB.ID.String(), "score": 2},
		},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}
	got := f.year(t, f.unscored.ID)
	if got.ScoredAt == nil || got.ScoredAt.Before(before) {
		t.Fatalf("scored_at = %v, want set to now", got.ScoredAt)
	}
	// (0.5*4 + 0.5*2)/5*100 = 60 -> Medium to High.
	if math.Abs(got.RiskIndex-60) > 1e-9 || got.RiskLevel != models.RiskLevelMediumToHigh || !got.AuditPriority {
		t.Fatalf("index=%v level=%q priority=%v", got.RiskIndex, got.RiskLevel, got.AuditPriority)
	}
	if !strings.Contains(w.Body.String(), `"scored_at":"`) {
		t.Fatalf("response does not carry scored_at: %s", w.Body.String())
	}
	events := f.outbox(t)
	if len(events) != 1 || events[0].EventType != kafka.EventYearScored {
		t.Fatalf("want one year.scored event, got %+v", events)
	}
}

func TestEstablishCreatesUnscoredRowsAndAnnouncesChanges(t *testing.T) {
	f := newScoringFixture(t)

	w := f.do(t, http.MethodPost, "/audit-universe/year/2027", map[string]any{
		"corporate_audit_universe_ids": []string{f.e1.ID.String(), f.e2.ID.String()},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}
	var rows []models.AuditUniverseYear
	f.db.Where("year = ?", 2027).Find(&rows)
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}
	for _, y := range rows {
		assertUnscored(t, y)
		var n int64
		f.db.Model(&models.AuditUniverseRiskScore{}).Where("audit_universe_year_id = ?", y.ID).Count(&n)
		if n != 0 {
			t.Fatalf("new row got %d default score rows", n)
		}
	}

	// GET returns scored_at (null) on every row.
	g := f.do(t, http.MethodGet, "/audit-universe/year/2027", nil)
	var resp struct {
		Data []map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(g.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("GET rows = %d", len(resp.Data))
	}
	for _, row := range resp.Data {
		v, ok := row["scored_at"]
		if !ok || string(v) != "null" {
			t.Fatalf("scored_at missing or not null: %s", v)
		}
		if string(row["risk_level"]) != `""` || string(row["risk_index"]) != "0" || string(row["audit_priority"]) != "false" {
			t.Fatalf("unexpected unscored row: %v", row)
		}
	}

	// Drop E1 from 2027.
	w = f.do(t, http.MethodPost, "/audit-universe/year/2027", map[string]any{
		"corporate_audit_universe_ids": []string{f.e2.ID.String()},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}

	events := f.outbox(t)
	var types []string
	for _, e := range events {
		types = append(types, e.EventType)
	}
	want := []string{kafka.EventYearScored, kafka.EventYearScored, kafka.EventYearRemoved}
	if strings.Join(types, ",") != strings.Join(want, ",") {
		t.Fatalf("events = %v, want %v", types, want)
	}
	var removed kafka.YearRemovedData
	_ = json.Unmarshal(events[2].Data, &removed)
	if removed.CorporateAuditUniverseID != f.e1.ID.String() || removed.Year != 2027 {
		t.Fatalf("unexpected year.removed data: %+v", removed)
	}
	var added kafka.YearScoreData
	_ = json.Unmarshal(events[0].Data, &added)
	if added.RiskLevel != "" || added.AuditPriority || added.RiskIndex != 0 {
		t.Fatalf("added row must be announced unscored: %+v", added)
	}
}

func TestDeleteWeightedStandardFactorConflicts(t *testing.T) {
	f := newScoringFixture(t)

	w := f.do(t, http.MethodDelete, "/risk-factors/standard/"+f.stdA.ID.String(), nil)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409: %s", w.Code, w.Body.String())
	}
	var body map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body["code"] != "FACTOR_HAS_CORPORATE_WEIGHT" || !strings.Contains(body["error"].(string), "50%") {
		t.Fatalf("unexpected body: %v", body)
	}
	var n int64
	f.db.Model(&models.StandardRiskFactor{}).Where("id = ?", f.stdA.ID).Count(&n)
	if n != 1 {
		t.Fatal("standard factor was deleted despite the conflict")
	}
	f.db.Model(&models.AuditUniverseRiskScore{}).Where("corporate_risk_factor_id = ?", f.cfA.ID).Count(&n)
	if n != 1 {
		t.Fatal("scores were deleted despite the conflict")
	}
}

func TestDeleteUnweightedStandardFactorSucceeds(t *testing.T) {
	f := newScoringFixture(t)
	now := time.Now()
	stdC := models.StandardRiskFactor{ID: uuid.New(), Name: "Unused", CreatedAt: now, UpdatedAt: now}
	mustCreate(t, f.db, &stdC)

	w := f.do(t, http.MethodDelete, "/risk-factors/standard/"+stdC.ID.String(), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}
}

// tree: root -> child -> grandchild, each with a 2026 row and a score; plus
// E1 (from the fixture) which must survive.
func buildTree(t *testing.T, f *scoringFixture) (root, child, grandchild models.CorporateAuditUniverse) {
	t.Helper()
	now := time.Now()
	root = models.CorporateAuditUniverse{ID: uuid.New(), Name: "Subsidiaries", CreatedAt: now, UpdatedAt: now}
	child = models.CorporateAuditUniverse{ID: uuid.New(), Name: "PT Anak", ParentID: &root.ID, CreatedAt: now, UpdatedAt: now}
	grandchild = models.CorporateAuditUniverse{ID: uuid.New(), Name: "Cabang Anak", ParentID: &child.ID, CreatedAt: now, UpdatedAt: now}
	for _, n := range []*models.CorporateAuditUniverse{&root, &child, &grandchild} {
		mustCreate(t, f.db, n)
		y := models.AuditUniverseYear{ID: uuid.New(), CorporateAuditUniverseID: n.ID, Year: 2026, ScoredAt: &now, CreatedAt: now, UpdatedAt: now}
		mustCreate(t, f.db, &y)
		mustCreate(t, f.db, &models.AuditUniverseRiskScore{ID: uuid.New(), AuditUniverseYearID: y.ID, CorporateRiskFactorID: f.cfA.ID, Score: 3, WeightedScore: 1.5, CreatedAt: now, UpdatedAt: now})
	}
	return root, child, grandchild
}

func TestDeleteCorporateNodeRemovesAllDescendants(t *testing.T) {
	f := newScoringFixture(t)
	root, child, grandchild := buildTree(t, f)

	w := f.do(t, http.MethodDelete, "/audit-universe/corporate/"+root.ID.String(), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}

	ids := []uuid.UUID{root.ID, child.ID, grandchild.ID}
	var n int64
	f.db.Model(&models.CorporateAuditUniverse{}).Where("id IN ?", ids).Count(&n)
	if n != 0 {
		t.Fatalf("%d subtree nodes left (grandchild must go too)", n)
	}
	f.db.Model(&models.AuditUniverseYear{}).Where("corporate_audit_universe_id IN ?", ids).Count(&n)
	if n != 0 {
		t.Fatalf("%d subtree year rows left", n)
	}
	// Only E1's two scores remain.
	f.db.Model(&models.AuditUniverseRiskScore{}).Count(&n)
	if n != 2 {
		t.Fatalf("score rows = %d, want 2 (E1's scores untouched)", n)
	}
	f.db.Model(&models.AuditUniverseYear{}).Where("corporate_audit_universe_id = ?", f.e1.ID).Count(&n)
	if n != 1 {
		t.Fatal("an unrelated entity's year row was deleted")
	}

	var deleted []string
	for _, e := range f.outbox(t) {
		if e.EventType != kafka.EventEntityDeleted {
			t.Fatalf("unexpected event %s", e.EventType)
		}
		var d kafka.EntityData
		_ = json.Unmarshal(e.Data, &d)
		deleted = append(deleted, d.Name)
	}
	if strings.Join(deleted, ",") != "Cabang Anak,PT Anak,Subsidiaries" {
		t.Fatalf("entity.deleted order = %v, want deepest first", deleted)
	}
}

func TestDeleteCorporateNodeRollsBackAndReportsErrors(t *testing.T) {
	f := newScoringFixture(t)
	root, child, grandchild := buildTree(t, f)

	// Make the score delete fail mid-transaction.
	if err := f.db.Exec("DROP TABLE audit_universe_risk_scores").Error; err != nil {
		t.Fatal(err)
	}

	w := f.do(t, http.MethodDelete, "/audit-universe/corporate/"+root.ID.String(), nil)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500: %s", w.Code, w.Body.String())
	}
	var n int64
	f.db.Model(&models.CorporateAuditUniverse{}).Where("id IN ?", []uuid.UUID{root.ID, child.ID, grandchild.ID}).Count(&n)
	if n != 3 {
		t.Fatalf("nodes left = %d, want 3 (transaction rolled back)", n)
	}
	if len(f.outbox(t)) != 0 {
		t.Fatal("no event may be published for a failed delete")
	}
}
