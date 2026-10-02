package controllers

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"audit-service/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// kbCreateTable creates the table by hand: the models' Postgres-only defaults
// (gen_random_uuid()) do not parse in SQLite, so AutoMigrate cannot be used.
func kbCreateTable(t *testing.T, db *gorm.DB, model interface{}) {
	t.Helper()
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(model); err != nil {
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
		case strings.Contains(dt, "decimal"), strings.Contains(dt, "float"):
			typ = "REAL"
		case dt == "time":
			typ = "DATETIME"
		}
		cols = append(cols, "`"+f.DBName+"` "+typ)
	}
	if err := db.Exec("CREATE TABLE `" + stmt.Schema.Table + "` (" + strings.Join(cols, ", ") + ")").Error; err != nil {
		t.Fatal(err)
	}
}

func newKbDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	kbCreateTable(t, db, &models.StrategicPlan{})
	kbCreateTable(t, db, &models.KPIAchievement{})
	return db
}

func kbCreate(t *testing.T, db *gorm.DB, v interface{}) {
	t.Helper()
	if err := db.Create(v).Error; err != nil {
		t.Fatal(err)
	}
}

func kbID(n byte) uuid.UUID {
	var u uuid.UUID
	u[15] = n
	return u
}

type kbResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Items      []KpiBreakdownItem `json:"items"`
		Pagination struct {
			Page       int `json:"page"`
			PageSize   int `json:"page_size"`
			Total      int `json:"total"`
			TotalPages int `json:"total_pages"`
		} `json:"pagination"`
	} `json:"data"`
}

func getKb(t *testing.T, db *gorm.DB, query string) (kbResponse, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	ctrl := NewPerformanceStatsController(db, nil)
	r.GET("/performance/kpi-breakdown", ctrl.GetKpiBreakdown)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/performance/kpi-breakdown"+query, nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("GET %s: status %d, body %s", query, w.Code, w.Body.String())
	}
	var resp kbResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode %s: %v", w.Body.String(), err)
	}
	if !resp.Success {
		t.Fatalf("success=false: %s", w.Body.String())
	}
	return resp, w.Body.String()
}

func kbMetrics(items []KpiBreakdownItem) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.Metric
	}
	return out
}

func kbByMetric(t *testing.T, items []KpiBreakdownItem, metric string) KpiBreakdownItem {
	t.Helper()
	for _, it := range items {
		if it.Metric == metric {
			return it
		}
	}
	t.Fatalf("no row %q in %v", metric, kbMetrics(items))
	return KpiBreakdownItem{}
}

func approx(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

// seedKb seeds the fixture used by most handler tests. For year 2026 the
// breakdown must contain exactly these rows, in this order:
//
//	Audit Completion Rate     plan P1  (kpiTargets[2026]=90, actual from newest matching achievement = 95)
//	Client Satisfaction       plan P3  (active only through kpiTargets["2026"])
//	Cost Variance             plan P8  (2025 achievement with the same name must NOT override)
//	Fraud Cases Investigated  achievement A2 (matches no plan)
//	Quarterly WBS             plan P4  (active only through kpiTargets["Q2-2026"], no target)
//	Report Timeliness (days)  plan P2  (HIB)
func seedKb(t *testing.T, db *gorm.DB) {
	t.Helper()
	plans := []models.StrategicPlan{
		{ID: kbID(1), KPI: "Audit Completion Rate", Unit: "%", HibHig: "HIG", SelectedPeriod: "2026",
			YearStart: 2024, YearEnd: 2028, KPITargets: map[string]string{"2025": "88", "2026": "90"}, Target: "85", Actual: "70"},
		{ID: kbID(2), KPI: "Report Timeliness (days)", Unit: "Day", HibHig: "HIB", SelectedPeriod: "2026",
			YearStart: 2026, YearEnd: 2026, Target: "10", Actual: "12"},
		{ID: kbID(3), KPI: "Client Satisfaction", Unit: "Score", HibHig: "HIG", SelectedPeriod: "2026",
			KPITargets: map[string]string{"2026": "4.5"}, Target: "4.0", Actual: "3"},
		{ID: kbID(4), KPI: "Quarterly WBS", Unit: "%", HibHig: "HIG", SelectedPeriod: "Q2-2026",
			KPITargets: map[string]string{"Q2-2026": "95"}, Target: "0", Actual: "40"},
		// inactive in 2026: range ends 2024, kpiTargets only for 2025
		{ID: kbID(5), KPI: "Old Plan", HibHig: "HIG", YearStart: 2020, YearEnd: 2024,
			KPITargets: map[string]string{"2025": "1"}, Target: "1", Actual: "1"},
		// active in 2026 but soft-deleted below
		{ID: kbID(6), KPI: "Deleted Plan", HibHig: "HIG", YearStart: 2026, YearEnd: 2026, Target: "1", Actual: "1"},
		// a blank kpiTargets entry for the year does not make the plan active
		{ID: kbID(7), KPI: "Empty Target Key", HibHig: "HIG", KPITargets: map[string]string{"2026": " "}, Target: "1", Actual: "1"},
		{ID: kbID(8), KPI: "Cost Variance", Unit: "%", HibHig: "HIG", SelectedPeriod: "Q1",
			YearStart: 2026, YearEnd: 2027, Target: "100", Actual: "85"},
	}
	for i := range plans {
		kbCreate(t, db, &plans[i])
	}
	if err := db.Delete(&models.StrategicPlan{}, "id = ?", kbID(6)).Error; err != nil {
		t.Fatal(err)
	}

	base := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	achs := []models.KPIAchievement{
		// two achievements match P1; the newest one (A1, actual 95) wins
		{ID: kbID(100), Year: 2026, Period: "Q1", KPIName: "audit completion", Target: 90, Actual: 50, CreatedAt: base},
		{ID: kbID(101), Year: 2026, Period: "Tahunan", KPIName: "AUDIT COMPLETION RATE FY", Target: 90, Actual: 95, CreatedAt: base.Add(time.Hour)},
		// unmatched in 2026 → its own row
		{ID: kbID(102), Year: 2026, Period: "Tahunan", KPIName: "Fraud Cases Investigated", Target: 10, Actual: 8, CreatedAt: base},
		// other year: neither overrides P8 nor appears as a row
		{ID: kbID(103), Year: 2025, Period: "Tahunan", KPIName: "Cost Variance", Target: 100, Actual: 1, CreatedAt: base},
		{ID: kbID(104), Year: 2025, Period: "Tahunan", KPIName: "Legacy KPI", Target: 1, Actual: 1, CreatedAt: base},
		// soft-deleted below
		{ID: kbID(105), Year: 2026, Period: "Tahunan", KPIName: "Deleted KPI", Target: 1, Actual: 1, CreatedAt: base},
	}
	for i := range achs {
		kbCreate(t, db, &achs[i])
	}
	if err := db.Delete(&models.KPIAchievement{}, "id = ?", kbID(105)).Error; err != nil {
		t.Fatal(err)
	}
}

var kb2026Order = []string{
	"Audit Completion Rate",
	"Client Satisfaction",
	"Cost Variance",
	"Fraud Cases Investigated",
	"Quarterly WBS",
	"Report Timeliness (days)",
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// handler tests (SQLite)
// ---------------------------------------------------------------------------

func TestKpiBreakdown_YearFilterMergeAndOrder(t *testing.T) {
	db := newKbDB(t)
	seedKb(t, db)

	resp, body := getKb(t, db, "?year=2026&page_size=100")
	if got := kbMetrics(resp.Data.Items); !equalStrings(got, kb2026Order) {
		t.Fatalf("rows = %v, want %v", got, kb2026Order)
	}
	p := resp.Data.Pagination
	if p.Total != 6 || p.TotalPages != 1 || p.Page != 1 || p.PageSize != 100 {
		t.Fatalf("pagination = %+v", p)
	}
	if !strings.Contains(body, `"achievementRate":null`) {
		t.Errorf("expected a null achievementRate in %s", body)
	}

	items := resp.Data.Items

	// P1: target from kpiTargets[2026], actual from the newest matching achievement
	p1 := kbByMetric(t, items, "Audit Completion Rate")
	if p1.ID != kbID(1).String() || p1.Source != KpiSourceStrategicPlan || p1.Unit != "%" || p1.Period != "2026" || p1.Category != "" {
		t.Errorf("P1 identity = %+v", p1)
	}
	if !approx(p1.Target, 90) || !approx(p1.Actual, 95) || !approx(p1.Gap, 5) || !p1.GapIsPositive {
		t.Errorf("P1 numbers = %+v", p1)
	}
	if p1.AchievementRate == nil || !approx(*p1.AchievementRate, 95.0/90*100) || p1.Status != KpiStatusExceeded {
		t.Errorf("P1 status = %+v", p1)
	}

	// P2: HIB → gap = target - actual; no kpiTargets → plan target; no achievement → plan actual
	p2 := kbByMetric(t, items, "Report Timeliness (days)")
	if !approx(p2.Target, 10) || !approx(p2.Actual, 12) || !approx(p2.Gap, -2) || p2.GapIsPositive {
		t.Errorf("P2 numbers = %+v", p2)
	}
	if p2.AchievementRate == nil || !approx(*p2.AchievementRate, 120) || p2.Status != KpiStatusExceeded {
		t.Errorf("P2 status = %+v", p2)
	}

	// P3: active through kpiTargets only; rate 66.7 → Needs Attention
	p3 := kbByMetric(t, items, "Client Satisfaction")
	if !approx(p3.Target, 4.5) || !approx(p3.Actual, 3) || !approx(p3.Gap, -1.5) || p3.GapIsPositive || p3.Status != KpiStatusNeedsAttention {
		t.Errorf("P3 = %+v", p3)
	}

	// P4: active through a quarterly key; target 0 → No Target, null rate
	p4 := kbByMetric(t, items, "Quarterly WBS")
	if !approx(p4.Target, 0) || p4.AchievementRate != nil || p4.Status != KpiStatusNoTarget || p4.Period != "Q2-2026" {
		t.Errorf("P4 = %+v", p4)
	}

	// P8: the 2025 achievement with the same name does not override the 2026 actual
	p8 := kbByMetric(t, items, "Cost Variance")
	if !approx(p8.Actual, 85) || p8.Status != KpiStatusOnTrack || p8.AchievementRate == nil || !approx(*p8.AchievementRate, 85) {
		t.Errorf("P8 = %+v", p8)
	}

	// A2: unmatched achievement row (treated as higher-is-good), rate exactly 80 → On Track
	a2 := kbByMetric(t, items, "Fraud Cases Investigated")
	if a2.ID != kbID(102).String() || a2.Source != KpiSourceKpiAchievement || a2.Period != "Tahunan" || a2.Unit != "" || a2.Category != "" {
		t.Errorf("A2 identity = %+v", a2)
	}
	if !approx(a2.Target, 10) || !approx(a2.Actual, 8) || !approx(a2.Gap, -2) || a2.GapIsPositive || a2.Status != KpiStatusOnTrack {
		t.Errorf("A2 numbers = %+v", a2)
	}
}

func TestKpiBreakdown_OtherYear(t *testing.T) {
	db := newKbDB(t)
	seedKb(t, db)

	resp, _ := getKb(t, db, "?year=2025&page_size=100")
	// 2025: P1 (range 2024-2028), P5 (kpiTargets["2025"]); P2/P3/P4/P8 are not active.
	// Achievements 2025: "Cost Variance" (P8 not active → unmatched), "Legacy KPI".
	want := []string{"Audit Completion Rate", "Cost Variance", "Legacy KPI", "Old Plan"}
	if got := kbMetrics(resp.Data.Items); !equalStrings(got, want) {
		t.Fatalf("rows = %v, want %v", got, want)
	}
	p1 := kbByMetric(t, resp.Data.Items, "Audit Completion Rate")
	// target kpiTargets["2025"]=88; no 2025 achievement matches → plan actual 70
	if !approx(p1.Target, 88) || !approx(p1.Actual, 70) {
		t.Errorf("P1 in 2025 = %+v", p1)
	}
	cv := kbByMetric(t, resp.Data.Items, "Cost Variance")
	if cv.Source != KpiSourceKpiAchievement {
		t.Errorf("Cost Variance 2025 should be the achievement row, got %+v", cv)
	}
}

func TestKpiBreakdown_Filters(t *testing.T) {
	db := newKbDB(t)
	seedKb(t, db)

	cases := []struct {
		query string
		want  []string
	}{
		{"&status=Exceeded", []string{"Audit Completion Rate", "Report Timeliness (days)"}},
		{"&status=On+Track", []string{"Cost Variance", "Fraud Cases Investigated"}},
		{"&status=Needs+Attention", []string{"Client Satisfaction"}},
		{"&status=No+Target", []string{"Quarterly WBS"}},
		{"&status=exceeded", []string{}}, // exact match
		{"&period=Q1", []string{"Cost Variance"}},
		{"&period=Tahunan", []string{"Fraud Cases Investigated"}},
		{"&period=2026", []string{"Audit Completion Rate", "Client Satisfaction", "Report Timeliness (days)"}},
		{"&period=Q2", []string{}}, // exact match, not prefix
		{"&category=Operational", []string{}},
		{"&category=", kb2026Order},
		{"&search=COST", []string{"Cost Variance"}},
		{"&search=rate", []string{"Audit Completion Rate"}},
		{"&search=%20wbs%20", []string{"Quarterly WBS"}},
		{"&search=nothing-matches", []string{}},
		{"&search=a&status=Exceeded&period=2026", []string{"Audit Completion Rate", "Report Timeliness (days)"}},
		{"&search=cost&status=Exceeded", []string{}},
	}
	for _, tc := range cases {
		resp, body := getKb(t, db, "?year=2026&page_size=100"+tc.query)
		got := kbMetrics(resp.Data.Items)
		if !equalStrings(got, tc.want) {
			t.Errorf("%s: rows = %v, want %v", tc.query, got, tc.want)
		}
		if resp.Data.Pagination.Total != len(tc.want) {
			t.Errorf("%s: total = %d, want %d", tc.query, resp.Data.Pagination.Total, len(tc.want))
		}
		if len(tc.want) == 0 {
			if resp.Data.Pagination.TotalPages != 0 {
				t.Errorf("%s: total_pages = %d, want 0", tc.query, resp.Data.Pagination.TotalPages)
			}
			if !strings.Contains(body, `"items":[]`) {
				t.Errorf("%s: items should be [] in %s", tc.query, body)
			}
		}
	}
}

func TestKpiBreakdown_Pagination(t *testing.T) {
	db := newKbDB(t)
	seedKb(t, db)

	cases := []struct {
		query                string
		page, size, totPages int
		want                 []string
	}{
		{"&page=1&page_size=4", 1, 4, 2, kb2026Order[:4]},
		{"&page=2&page_size=4", 2, 4, 2, kb2026Order[4:]},
		{"&page=3&page_size=4", 3, 4, 2, []string{}}, // past the end
		{"&page=99999999999&page_size=4", 99999999999, 4, 2, []string{}},
		{"&page=6&page_size=1", 6, 1, 6, kb2026Order[5:]},
		{"&page=1&page_size=6", 1, 6, 1, kb2026Order},
		{"&page=1&page_size=5", 1, 5, 2, kb2026Order[:5]},
		{"", 1, 10, 1, kb2026Order}, // defaults
		{"&page=0", 1, 10, 1, kb2026Order},
		{"&page=-2", 1, 10, 1, kb2026Order},
		{"&page=abc", 1, 10, 1, kb2026Order},
		{"&page_size=0", 1, 10, 1, kb2026Order},
		{"&page_size=-5", 1, 10, 1, kb2026Order},
		{"&page_size=101", 1, 10, 1, kb2026Order},
		{"&page_size=abc", 1, 10, 1, kb2026Order},
		{"&page_size=100", 1, 100, 1, kb2026Order},
		{"&page_size=1", 1, 1, 6, kb2026Order[:1]},
	}
	for _, tc := range cases {
		resp, body := getKb(t, db, "?year=2026"+tc.query)
		p := resp.Data.Pagination
		if p.Page != tc.page || p.PageSize != tc.size || p.Total != 6 || p.TotalPages != tc.totPages {
			t.Errorf("%s: pagination = %+v, want page=%d size=%d total=6 pages=%d", tc.query, p, tc.page, tc.size, tc.totPages)
		}
		if got := kbMetrics(resp.Data.Items); !equalStrings(got, tc.want) {
			t.Errorf("%s: rows = %v, want %v", tc.query, got, tc.want)
		}
		if len(tc.want) == 0 && !strings.Contains(body, `"items":[]`) {
			t.Errorf("%s: items should be [] in %s", tc.query, body)
		}
	}
}

func TestKpiBreakdown_PagesCoverEveryRowOnce(t *testing.T) {
	db := newKbDB(t)
	seedKb(t, db)
	var all []string
	for page := 1; page <= 3; page++ {
		resp, _ := getKb(t, db, "?year=2026&page_size=2&page="+string(rune('0'+page)))
		all = append(all, kbMetrics(resp.Data.Items)...)
	}
	if !equalStrings(all, kb2026Order) {
		t.Fatalf("pages concatenated = %v, want %v", all, kb2026Order)
	}
}

func TestKpiBreakdown_EmptyResult(t *testing.T) {
	db := newKbDB(t)
	resp, body := getKb(t, db, "?year=2026")
	p := resp.Data.Pagination
	if p.Total != 0 || p.TotalPages != 0 || p.Page != 1 || p.PageSize != 10 || len(resp.Data.Items) != 0 {
		t.Fatalf("empty: %+v items=%v", p, resp.Data.Items)
	}
	if !strings.Contains(body, `"items":[]`) {
		t.Fatalf("items should be [] in %s", body)
	}
	if !strings.Contains(body, `"message":"KPI breakdown fetched successfully"`) {
		t.Fatalf("envelope: %s", body)
	}

	// seeded DB, year with no plans or achievements
	seedKb(t, db)
	resp, _ = getKb(t, db, "?year=1999")
	if resp.Data.Pagination.Total != 0 || resp.Data.Pagination.TotalPages != 0 || len(resp.Data.Items) != 0 {
		t.Fatalf("1999: %+v", resp.Data)
	}
}

func TestKpiBreakdown_DefaultYearIsCurrentYear(t *testing.T) {
	db := newKbDB(t)
	y := time.Now().Year()
	kbCreate(t, db, &models.StrategicPlan{ID: kbID(1), KPI: "This Year", HibHig: "HIG", YearStart: y, YearEnd: y, Target: "1", Actual: "1"})
	kbCreate(t, db, &models.StrategicPlan{ID: kbID(2), KPI: "Last Year", HibHig: "HIG", YearStart: y - 1, YearEnd: y - 1, Target: "1", Actual: "1"})

	for _, q := range []string{"", "?year=", "?year=abc", "?year=0", "?year=-1"} {
		resp, _ := getKb(t, db, q)
		if got := kbMetrics(resp.Data.Items); !equalStrings(got, []string{"This Year"}) {
			t.Errorf("%q: rows = %v, want [This Year]", q, got)
		}
	}
}

// ---------------------------------------------------------------------------
// unit tests for the rules
// ---------------------------------------------------------------------------

func TestParseKpiBreakdownQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Date(2031, 5, 1, 0, 0, 0, 0, time.UTC)
	parse := func(raw string) KpiBreakdownQuery {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/x?"+raw, nil)
		return ParseKpiBreakdownQuery(c, now)
	}
	if q := parse(""); q.Year != 2031 || q.Page != 1 || q.PageSize != 10 || q.Search != "" || q.Category != "" || q.Status != "" || q.Period != "" {
		t.Errorf("defaults = %+v", q)
	}
	if q := parse("year=2024&page=3&page_size=25&search=%20Foo%20&category=C&status=On+Track&period=Q1"); q.Year != 2024 || q.Page != 3 || q.PageSize != 25 ||
		q.Search != "Foo" || q.Category != "C" || q.Status != "On Track" || q.Period != "Q1" {
		t.Errorf("explicit = %+v", q)
	}
	for _, ps := range []string{"0", "-1", "101", "1000", "x", "1.5"} {
		if q := parse("page_size=" + ps); q.PageSize != 10 {
			t.Errorf("page_size=%s → %d, want 10", ps, q.PageSize)
		}
	}
	for _, ps := range []string{"1", "100"} {
		if q := parse("page_size=" + ps); q.PageSize == 10 {
			t.Errorf("page_size=%s should be kept", ps)
		}
	}
}

func TestKpiBreakdown_GapAndStatusRules(t *testing.T) {
	cases := []struct {
		name           string
		target, actual float64
		hig            bool
		gap            float64
		positive       bool
		status         string
		rate           float64 // -1 = null
	}{
		{"HIG above target", 90, 99, true, 9, true, KpiStatusExceeded, 110},
		{"HIG at target", 90, 90, true, 0, true, KpiStatusExceeded, 100},
		{"HIG 80% boundary", 100, 80, true, -20, false, KpiStatusOnTrack, 80},
		{"HIG just under 80%", 100, 79.99, true, -20.01, false, KpiStatusNeedsAttention, 79.99},
		{"HIB below target", 10, 8, false, 2, true, KpiStatusOnTrack, 80},
		{"HIB above target", 10, 12, false, -2, false, KpiStatusExceeded, 120},
		{"zero target", 0, 5, true, 5, true, KpiStatusNoTarget, -1},
		{"negative target", -3, 5, false, -8, false, KpiStatusNoTarget, -1},
		{"zero actual", 50, 0, true, -50, false, KpiStatusNeedsAttention, 0},
	}
	for _, tc := range cases {
		it := newKpiBreakdownItem("id", KpiSourceStrategicPlan, "m", "", "", tc.target, tc.actual, tc.hig)
		if !approx(it.Gap, tc.gap) || it.GapIsPositive != tc.positive || it.Status != tc.status {
			t.Errorf("%s: gap=%v positive=%v status=%q", tc.name, it.Gap, it.GapIsPositive, it.Status)
		}
		if tc.rate < 0 {
			if it.AchievementRate != nil {
				t.Errorf("%s: rate = %v, want null", tc.name, *it.AchievementRate)
			}
		} else if it.AchievementRate == nil || math.Abs(*it.AchievementRate-tc.rate) > 1e-9 {
			t.Errorf("%s: rate = %v, want %v", tc.name, it.AchievementRate, tc.rate)
		}
	}

	// HIB/HIG flag is matched case-insensitively; anything else is treated as HIB
	for in, want := range map[string]bool{"HIG": true, "hig": true, " HIG ": true, "HIB": false, "": false, "LIB": false} {
		if got := higherIsGood(in); got != want {
			t.Errorf("higherIsGood(%q) = %v", in, got)
		}
	}

	// overflow never yields Inf/NaN (encoding/json would fail)
	it := newKpiBreakdownItem("id", KpiSourceStrategicPlan, "m", "", "", 1e-300, 1e300, true)
	if math.IsInf(*it.AchievementRate, 0) || math.IsNaN(*it.AchievementRate) || math.IsInf(it.Gap, 0) {
		t.Errorf("overflow: %+v", it)
	}
	if _, err := json.Marshal(it); err != nil {
		t.Errorf("marshal overflow row: %v", err)
	}
}

func TestKpiNamesMatch(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"Audit Completion Rate", "audit completion", true},
		{"audit", "AUDIT COMPLETION RATE", true},
		{"Report Timeliness", "Audit Completion", false},
		{"", "anything", false},
		{"anything", "  ", false},
	}
	for _, tc := range cases {
		if got := kpiNamesMatch(tc.a, tc.b); got != tc.want {
			t.Errorf("kpiNamesMatch(%q, %q) = %v", tc.a, tc.b, got)
		}
	}
}

func TestBuildKpiBreakdownRows_PlanNameFallbackAndEmptyNames(t *testing.T) {
	plans := []models.StrategicPlan{
		// no KPI text: the objective is used for matching and as the metric
		{ID: kbID(1), StrategicObjective: "Board Reporting Quality", HibHig: "HIG", YearStart: 2026, YearEnd: 2026, Target: "10", Actual: "1"},
	}
	achs := []models.KPIAchievement{
		{ID: kbID(10), Year: 2026, KPIName: "board reporting quality", Target: 10, Actual: 9},
		{ID: kbID(11), Year: 2026, KPIName: "", Target: 1, Actual: 1}, // empty name never matches a plan
	}
	rows := BuildKpiBreakdownRows(plans, achs, 2026)
	if len(rows) != 2 {
		t.Fatalf("rows = %+v", rows)
	}
	if rows[0].Metric != "Board Reporting Quality" || !approx(rows[0].Actual, 9) {
		t.Errorf("plan row = %+v", rows[0])
	}
	if rows[1].Source != KpiSourceKpiAchievement || rows[1].ID != kbID(11).String() {
		t.Errorf("achievement row = %+v", rows[1])
	}
}

func TestPlanActiveInYear(t *testing.T) {
	cases := []struct {
		name string
		p    models.StrategicPlan
		want bool
	}{
		{"in range", models.StrategicPlan{YearStart: 2024, YearEnd: 2028}, true},
		{"range start", models.StrategicPlan{YearStart: 2026, YearEnd: 2027}, true},
		{"range end", models.StrategicPlan{YearStart: 2025, YearEnd: 2026}, true},
		{"before range", models.StrategicPlan{YearStart: 2027, YearEnd: 2028}, false},
		{"no range", models.StrategicPlan{}, false},
		{"year key", models.StrategicPlan{KPITargets: map[string]string{"2026": "5"}}, true},
		{"blank year key", models.StrategicPlan{KPITargets: map[string]string{"2026": ""}}, false},
		{"quarter key", models.StrategicPlan{KPITargets: map[string]string{"Q4-2026": "5"}}, true},
		{"quarter key other year", models.StrategicPlan{KPITargets: map[string]string{"Q4-2025": "5"}}, false},
		{"bare quarter key", models.StrategicPlan{KPITargets: map[string]string{"Q1": "5"}}, false},
		{"Q5 is not a quarter", models.StrategicPlan{KPITargets: map[string]string{"Q5-2026": "5"}}, false},
		{"longer year-like key", models.StrategicPlan{KPITargets: map[string]string{"20260": "5"}}, false},
	}
	for _, tc := range cases {
		if got := planActiveInYear(tc.p, 2026); got != tc.want {
			t.Errorf("%s: %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestYearTargetAndLeadingFloat(t *testing.T) {
	p := models.StrategicPlan{KPITargets: map[string]string{"2026": "abc", "2027": "92.5%", "2028": ""}}
	if _, ok := yearTarget(p, 2026); ok {
		t.Error("non-numeric kpiTargets value should fall back to the plan target")
	}
	if v, ok := yearTarget(p, 2027); !ok || !approx(v, 92.5) {
		t.Errorf("2027 = %v %v", v, ok)
	}
	if _, ok := yearTarget(p, 2028); ok {
		t.Error("blank kpiTargets value should fall back to the plan target")
	}
	if _, ok := yearTarget(models.StrategicPlan{}, 2026); ok {
		t.Error("nil kpiTargets")
	}

	cases := map[string]float64{
		"84.0": 84, " 90% ": 90, "4.5 / 5": 4.5, "-3": -3, "+2.5": 2.5, ".5": 0.5, "1e2": 100,
		"1e": 1, "1.2.3": 1.2, "": 0, "abc": 0, "-": 0, ".": 0, "NaN": 0, "Infinity": 0, "1e999": 0,
	}
	for in, want := range cases {
		if got := parseLeadingFloat(in); !approx(got, want) {
			t.Errorf("parseLeadingFloat(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestSortKpiBreakdownRows(t *testing.T) {
	rows := []KpiBreakdownItem{
		{ID: "b", Metric: "banana"},
		{ID: "z", Metric: "Apple"},
		{ID: "a", Metric: "apple"},
		{ID: "c", Metric: "Apple"},
		{ID: "x", Metric: ""},
	}
	SortKpiBreakdownRows(rows)
	var got []string
	for _, r := range rows {
		got = append(got, r.Metric+"/"+r.ID)
	}
	want := []string{"/x", "Apple/c", "Apple/z", "apple/a", "banana/b"}
	if !equalStrings(got, want) {
		t.Fatalf("order = %v, want %v", got, want)
	}
}

// ---------------------------------------------------------------------------
// SQL shape (Postgres DryRun)
// ---------------------------------------------------------------------------

// Both queries exclude soft-deleted rows, the achievements are limited to the
// requested year with a bound parameter, and the ordering is deterministic.
func TestKpiBreakdownSQLShape(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=127.0.0.1 dbname=none"}), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var sqls []string
	var vars [][]interface{}
	capture := func(tx *gorm.DB) {
		sqls = append(sqls, tx.Statement.SQL.String())
		vars = append(vars, tx.Statement.Vars)
		tx.Statement.SQL.Reset()
		tx.Statement.Vars = nil
	}
	_ = db.Callback().Query().After("gorm:query").Register("test:capture", capture)

	getKb(t, db, "?year=2026&search=x'%3B--&status=Exceeded&period=Q1&category=c")

	want := []string{
		`SELECT * FROM "strategic_plans" WHERE "strategic_plans"."deleted_at" IS NULL ORDER BY "id"`,
		`SELECT * FROM "kpi_achievements" WHERE year = $1 AND "kpi_achievements"."deleted_at" IS NULL ORDER BY "created_at" DESC,"id"`,
	}
	if len(sqls) != len(want) {
		t.Fatalf("got %d statements:\n%s", len(sqls), strings.Join(sqls, "\n"))
	}
	for i := range want {
		if sqls[i] != want[i] {
			t.Errorf("statement %d:\n got  %s\n want %s", i, sqls[i], want[i])
		}
	}
	if len(vars[1]) != 1 || vars[1][0] != 2026 {
		t.Errorf("achievement query vars = %v, want [2026]", vars[1])
	}
}
