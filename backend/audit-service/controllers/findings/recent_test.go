package findings

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"audit-service/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func day(d int) time.Time { return time.Date(2026, 9, d, 8, 0, 0, 0, time.UTC) }

func titles(items []RecentItem) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.Title
	}
	return out
}

func TestNormalizeCategory(t *testing.T) {
	cases := map[string]string{
		"Very Significant":       "Very Significant",
		"very_significant":       "Very Significant",
		"HIGH":                   "Very Significant",
		"critical":               "Very Significant",
		"Very High":              "Very Significant",
		"Significant":            "Significant",
		"MEDIUM":                 "Significant",
		"Moderate":               "Significant",
		"Quite Significant":      "Quite Significant",
		"Moderately Significant": "Quite Significant",
		"  low ":                 "Quite Significant",
		"Not Significant":        "Not Significant",
		"Insignificant":          "Not Significant",
		"":                       "Significant",
		"whatever":               "Significant",
	}
	for in, want := range cases {
		if got := NormalizeCategory(in); got != want {
			t.Errorf("NormalizeCategory(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseRecentLimit(t *testing.T) {
	cases := map[string]int{"": 5, "abc": 5, "0": 5, "-3": 5, "51": 5, "1": 1, "50": 50, "12": 12, " 7 ": 7}
	for in, want := range cases {
		if got := ParseRecentLimit(in); got != want {
			t.Errorf("ParseRecentLimit(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestDeriveKKAAndFieldwork(t *testing.T) {
	src := LiveSources{
		Causes: []models.WorkingPaperCause{
			{Condition: "  Stock opname not done ", CreatedAt: day(1), UpdatedAt: day(3)},
			{Condition: "", CreatedAt: day(1)},
			{Condition: "Approval override by staff", CreatedAt: day(2)}, // zero UpdatedAt -> CreatedAt
		},
		Plans: []models.WorkingPaperPlan{
			{ActionDescription: "Do stock opname monthly"},
			{Recommendation: "Second plan"},
			{Recommendation: "Remove override access"},
		},
		Risks: []models.WorkingPaperRisk{{RiskLevel: "MEDIUM"}},
		TestControls: []models.FieldworkTestControl{
			{ControlName: "Segregation", TestResult: "Ineffective", MitigationPlan: "Split duties", UpdatedAt: day(4)},
			{ControlName: "Review", TestResult: "partially effective", Recommendation: "Add reviewer", UpdatedAt: day(5)},
			{ControlName: "Backup", TestResult: "Effective", Finding: "Backup log missing", UpdatedAt: day(6)},
			{ControlName: "Access", TestResult: "Effective", UpdatedAt: day(7)},                                  // no finding
			{ControlName: "Dup", TestResult: "Ineffective", Finding: "stock opname not done", UpdatedAt: day(8)}, // dup of KKA
		},
	}
	got := Derive(src)
	want := []Finding{
		{Title: "Stock opname not done", Category: "Significant", Action: "Do stock opname monthly", Source: SourceWorkingPaper, Date: day(3)},
		{Title: "Approval override by staff", Category: "Very Significant", Action: "Remove override access", Source: SourceWorkingPaper, Date: day(2)},
		{Title: "Kelemahan Kontrol: Segregation", Category: "Very Significant", Action: "Split duties", Source: SourceFieldwork, Date: day(4)},
		{Title: "Kelemahan Kontrol: Review", Category: "Significant", Action: "Add reviewer", Source: SourceFieldwork, Date: day(5)},
		{Title: "Backup log missing", Category: "Significant", Source: SourceFieldwork, Date: day(6)},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d findings %+v, want %d", len(got), got, len(want))
	}
	for i := range want {
		g, w := got[i], want[i]
		if g.Title != w.Title || g.Category != w.Category || g.Action != w.Action || g.Source != w.Source || !g.Date.Equal(w.Date) {
			t.Errorf("finding %d = %+v, want %+v", i, g, w)
		}
	}

	// A HIGH risk in F02 raises the default KKA category
	src.Risks = append(src.Risks, models.WorkingPaperRisk{RiskLevel: "high"})
	if c := Derive(src)[0].Category; c != "Very Significant" {
		t.Errorf("category with HIGH risk = %q", c)
	}
}

func TestMergeRecentDedupePrefersReportWithNewerDate(t *testing.T) {
	reportID := uuid.New()
	reportDate := day(10)
	reports := []models.AuditResultReport{{
		ID: reportID, AssignmentLetterID: "AL-1", ReportDate: &reportDate, CreatedAt: day(1),
		Findings: []models.AuditReportFinding{
			{Title: "Stock Opname Not Done", Category: "Moderately Significant", Action: "report action"},
			{Title: " stock opname not done ", Category: "HIGH"}, // dup inside the same report
			{Title: "Old finding", Category: "Insignificant"},
		},
	}}
	live := []LetterFindings{
		{AssignmentLetterID: "AL-1", Findings: []Finding{
			{Title: "stock opname not done  ", Category: "Very Significant", Action: "live action", Source: SourceWorkingPaper, Date: day(20)},
			{Title: "Live only", Category: "Significant", Source: SourceFieldwork, Date: day(15)},
		}},
		// Same title on another letter is a different finding
		{AssignmentLetterID: "AL-2", Findings: []Finding{
			{Title: "Stock opname not done", Category: "LOW", Source: SourceFieldwork, Date: day(5)},
		}},
	}

	res := MergeRecent(reports, live, 50)
	if res.Total != 4 {
		t.Fatalf("total = %d, want 4 (%v)", res.Total, titles(res.Items))
	}
	first := res.Items[0]
	if first.Title != "Stock Opname Not Done" || first.Source != SourceAuditResultReport ||
		first.Action != "report action" || first.Category != "Quite Significant" ||
		first.ReportID == nil || *first.ReportID != reportID.String() ||
		first.Date != "2026-09-20T08:00:00Z" || first.AssignmentLetterID != "AL-1" {
		t.Errorf("merged entry = %+v", first)
	}
	if got := strings.Join(titles(res.Items), "|"); got != "Stock Opname Not Done|Live only|Old finding|Stock opname not done" {
		t.Errorf("order = %s", got)
	}
	if it := res.Items[1]; it.ReportID != nil || it.Date != "2026-09-15T08:00:00Z" {
		t.Errorf("live item = %+v", it)
	}
	if it := res.Items[2]; it.Category != "Not Significant" || it.Date != "2026-09-10T08:00:00Z" {
		t.Errorf("report item = %+v", it)
	}
	if it := res.Items[3]; it.AssignmentLetterID != "AL-2" || it.Category != "Quite Significant" {
		t.Errorf("AL-2 item = %+v", it)
	}
}

func TestMergeRecentReportDateFallbackAndEmptyLetter(t *testing.T) {
	r1, r2 := uuid.New(), uuid.New()
	reports := []models.AuditResultReport{
		{ID: r1, CreatedAt: day(3), Findings: []models.AuditReportFinding{{Title: "Same"}}},
		{ID: r2, CreatedAt: day(4), Findings: []models.AuditReportFinding{{Title: "same"}, {Title: "  "}}},
	}
	res := MergeRecent(reports, nil, 5)
	// No assignment letter: dedupe is per report, blank titles are dropped
	if res.Total != 2 {
		t.Fatalf("total = %d, want 2", res.Total)
	}
	if res.Items[0].Date != "2026-09-04T08:00:00Z" || *res.Items[0].ReportID != r2.String() {
		t.Errorf("first = %+v", res.Items[0])
	}
}

func TestMergeRecentOrderTiebreakLimitTotal(t *testing.T) {
	same := day(9)
	live := []LetterFindings{
		{AssignmentLetterID: "B", Findings: []Finding{{Title: "b2", Date: same}, {Title: "B1", Date: same}}},
		{AssignmentLetterID: "A", Findings: []Finding{{Title: "z", Date: same}, {Title: "old", Date: day(1)}, {Title: "new", Date: day(30)}}},
	}
	res := MergeRecent(nil, live, 50)
	if got := strings.Join(titles(res.Items), "|"); got != "new|z|B1|b2|old" {
		t.Errorf("order = %s", got)
	}
	for _, limit := range []int{0, -1, 51} {
		if r := MergeRecent(nil, live, limit); len(r.Items) != 5 || r.Total != 5 {
			t.Errorf("limit %d -> %d items, total %d", limit, len(r.Items), r.Total)
		}
	}
	r := MergeRecent(nil, live, 2)
	if len(r.Items) != 2 || r.Total != 5 || r.Items[1].Title != "z" {
		t.Errorf("limit 2 -> %+v total %d", titles(r.Items), r.Total)
	}
	if r := MergeRecent(nil, nil, 5); r.Items == nil || len(r.Items) != 0 || r.Total != 0 {
		t.Errorf("empty -> %+v", r)
	}
}

// --- end to end on sqlite through the gin handler ---

// createTable builds a sqlite table from the GORM schema; AutoMigrate cannot
// run here because the models default ids to gen_random_uuid()
func createTable(t *testing.T, db *gorm.DB, model interface{}) {
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
		case dt == "time":
			typ = "DATETIME"
		}
		cols = append(cols, "`"+f.DBName+"` "+typ)
	}
	if err := db.Exec("CREATE TABLE `" + stmt.Schema.Table + "` (" + strings.Join(cols, ", ") + ")").Error; err != nil {
		t.Fatal(err)
	}
}

func newRecentDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	for _, m := range []interface{}{&models.AuditResultReport{}, &models.WorkingPaperCause{}, &models.WorkingPaperPlan{},
		&models.WorkingPaperRisk{}, &models.FieldworkTestControl{}, &models.AssignmentLetter{}} {
		createTable(t, db, m)
	}
	return db
}

func mustCreate(t *testing.T, db *gorm.DB, v interface{}) {
	t.Helper()
	if err := db.Create(v).Error; err != nil {
		t.Fatal(err)
	}
}

type recentResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Data    RecentResult `json:"data"`
}

func getRecent(t *testing.T, db *gorm.DB, query string) recentResponse {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := r.Group("/audit-result-reports")
	g.GET("/recent-findings", Recent(db))
	g.GET("/:id", func(c *gin.Context) { c.String(http.StatusTeapot, "shadowed") })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/audit-result-reports/recent-findings"+query, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
	var resp recentResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("%v: %s", err, w.Body.String())
	}
	return resp
}

func TestRecentEndpoint(t *testing.T) {
	db := newRecentDB(t)
	letter, deletedLetter := uuid.New(), uuid.New()
	lid, dlid := letter.String(), deletedLetter.String()

	mustCreate(t, db, &models.AssignmentLetter{ID: letter, LetterNumber: "ST-1"})
	dl := &models.AssignmentLetter{ID: deletedLetter, LetterNumber: "ST-2"}
	mustCreate(t, db, dl)
	if err := db.Delete(dl).Error; err != nil {
		t.Fatal(err)
	}

	rd := day(10)
	report := &models.AuditResultReport{ID: uuid.New(), AssignmentLetterID: lid, ReportDate: &rd, CreatedAt: day(2), UpdatedAt: day(2),
		Findings: []models.AuditReportFinding{{Title: "Cash count variance", Category: "Very Significant", Action: "Reconcile"}}}
	mustCreate(t, db, report)
	gone := &models.AuditResultReport{ID: uuid.New(), AssignmentLetterID: lid, CreatedAt: day(25), UpdatedAt: day(25),
		Findings: []models.AuditReportFinding{{Title: "From deleted report"}}}
	mustCreate(t, db, gone)
	if err := db.Delete(gone).Error; err != nil {
		t.Fatal(err)
	}

	// KKA: one cause duplicates the report finding with a newer date
	mustCreate(t, db, &models.WorkingPaperCause{ID: uuid.New(), WorkingPaperID: lid, Condition: "cash count variance", CreatedAt: day(11), UpdatedAt: day(12)})
	mustCreate(t, db, &models.WorkingPaperCause{ID: uuid.New(), WorkingPaperID: lid, Condition: "Unapproved vendor", CreatedAt: day(13), UpdatedAt: day(13)})
	mustCreate(t, db, &models.WorkingPaperPlan{ID: uuid.New(), WorkingPaperID: lid, ActionDescription: "plan 1", CreatedAt: day(11), UpdatedAt: day(11)})
	mustCreate(t, db, &models.WorkingPaperPlan{ID: uuid.New(), WorkingPaperID: lid, ActionDescription: "plan 2", CreatedAt: day(12), UpdatedAt: day(12)})
	delCause := &models.WorkingPaperCause{ID: uuid.New(), WorkingPaperID: lid, Condition: "Soft deleted cause", CreatedAt: day(28), UpdatedAt: day(28)}
	mustCreate(t, db, delCause)
	if err := db.Delete(delCause).Error; err != nil {
		t.Fatal(err)
	}

	// Fieldwork
	mustCreate(t, db, &models.FieldworkTestControl{ID: uuid.New(), AssignmentLetterID: lid, ControlName: "Segregation", TestResult: "INEFFECTIVE", CreatedAt: day(14), UpdatedAt: day(14)})
	mustCreate(t, db, &models.FieldworkTestControl{ID: uuid.New(), AssignmentLetterID: lid, ControlName: "Fine", TestResult: "EFFECTIVE", CreatedAt: day(29), UpdatedAt: day(29)})
	// Fieldwork of a soft-deleted assignment letter is ignored
	mustCreate(t, db, &models.FieldworkTestControl{ID: uuid.New(), AssignmentLetterID: dlid, ControlName: "Ghost", TestResult: "INEFFECTIVE", CreatedAt: day(30), UpdatedAt: day(30)})

	resp := getRecent(t, db, "")
	if !resp.Success || resp.Data.Total != 3 {
		t.Fatalf("resp = %+v", resp)
	}
	if got := strings.Join(titles(resp.Data.Items), "|"); got != "Kelemahan Kontrol: Segregation|Unapproved vendor|Cash count variance" {
		t.Fatalf("order = %s", got)
	}
	cash := resp.Data.Items[2]
	if cash.ReportID == nil || *cash.ReportID != report.ID.String() || cash.Source != SourceAuditResultReport ||
		cash.Date != "2026-09-12T08:00:00Z" || cash.Action != "Reconcile" || cash.AssignmentLetterID != lid {
		t.Errorf("cash = %+v", cash)
	}
	if v := resp.Data.Items[1]; v.ReportID != nil || v.Action != "plan 2" || v.Source != SourceWorkingPaper || v.Date != "2026-09-13T08:00:00Z" {
		t.Errorf("vendor = %+v", v)
	}

	if r := getRecent(t, db, "?limit=1"); len(r.Data.Items) != 1 || r.Data.Total != 3 {
		t.Errorf("limit=1 -> %d items total %d", len(r.Data.Items), r.Data.Total)
	}
	if r := getRecent(t, db, "?limit=999"); len(r.Data.Items) != 3 {
		t.Errorf("limit=999 -> %d items", len(r.Data.Items))
	}
}

func TestRecentEndpointEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/x", Recent(newRecentDB(t)))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"data":{"items":[],"total":0}`) {
		t.Errorf("%d %s", w.Code, w.Body.String())
	}
}

func TestLetterFilter(t *testing.T) {
	liveID, delID := uuid.New(), uuid.New()
	deletedAt := gorm.DeletedAt{Time: day(1), Valid: true}
	rows := []LetterRow{
		{ID: liveID, LetterNumber: "ST-001/SKAI/2026"},
		{ID: delID, LetterNumber: "ST-009/SKAI/2026", DeletedAt: deletedAt},
	}
	f := NewLetterFilter(rows)
	cases := map[string]bool{
		"ST-001/SKAI/2026":               false, // live by number
		"  st-001/skai/2026 ":            false,
		liveID.String():                  false, // live by id
		strings.ToUpper(liveID.String()): false,
		"ST-009/SKAI/2026":               true, // deleted by number
		" st-009/skai/2026":              true,
		delID.String():                   true, // deleted by id
		"ST-404/SKAI/2026":               true, // orphan
		uuid.NewString():                 true, // orphan id
		"":                               false,
		"   ":                            false,
	}
	for key, want := range cases {
		if got := f.Drop(key); got != want {
			t.Errorf("Drop(%q) = %v, want %v", key, got, want)
		}
	}

	// Empty table: nothing is known, nothing is hidden
	empty := NewLetterFilter(nil)
	if empty.Drop("ST-404/SKAI/2026") || empty.Drop(uuid.NewString()) {
		t.Error("empty letters table must not drop orphans")
	}
	// Only deleted letters: those are still dropped, orphans are kept
	onlyDeleted := NewLetterFilter(rows[1:])
	if !onlyDeleted.Drop("ST-009/SKAI/2026") || !onlyDeleted.Drop(delID.String()) {
		t.Error("deleted letter must be dropped")
	}
	if onlyDeleted.Drop("ST-404/SKAI/2026") {
		t.Error("orphan must be kept when no live letter exists")
	}
	// Unreadable table -> zero filter keeps everything
	if (LetterFilter{}).Drop("anything") {
		t.Error("zero filter must keep everything")
	}
}

// seedByNumber seeds findings keyed by letter NUMBER, as the frontend saves them
func seedByNumber(t *testing.T, db *gorm.DB) (keptReport uuid.UUID) {
	t.Helper()
	add := func(letter, title string, d int) {
		mustCreate(t, db, &models.WorkingPaperCause{ID: uuid.New(), WorkingPaperID: letter, Condition: title, CreatedAt: day(d), UpdatedAt: day(d)})
	}
	keptReport = uuid.New()
	mustCreate(t, db, &models.AuditResultReport{ID: keptReport, AssignmentLetterID: " st-001/skai/2026 ", CreatedAt: day(1), UpdatedAt: day(1),
		Findings: []models.AuditReportFinding{{Title: "Live report finding"}}})
	mustCreate(t, db, &models.AuditResultReport{ID: uuid.New(), AssignmentLetterID: "ST-009/SKAI/2026", CreatedAt: day(2), UpdatedAt: day(2),
		Findings: []models.AuditReportFinding{{Title: "Deleted-letter report finding"}}})
	mustCreate(t, db, &models.AuditResultReport{ID: uuid.New(), AssignmentLetterID: "", CreatedAt: day(3), UpdatedAt: day(3),
		Findings: []models.AuditReportFinding{{Title: "No-letter report finding"}}})
	add("ST-001/SKAI/2026", "Live KKA finding", 4)
	add("ST-009/SKAI/2026", "Deleted-letter KKA finding", 5)
	add("ST-404/SKAI/2026", "Orphan KKA finding", 6)
	mustCreate(t, db, &models.FieldworkTestControl{ID: uuid.New(), AssignmentLetterID: "ST-009/SKAI/2026", ControlName: "Gone", TestResult: "INEFFECTIVE", CreatedAt: day(7), UpdatedAt: day(7)})
	mustCreate(t, db, &models.FieldworkTestControl{ID: uuid.New(), AssignmentLetterID: "ST-404/SKAI/2026", ControlName: "Orphan", TestResult: "INEFFECTIVE", CreatedAt: day(8), UpdatedAt: day(8)})
	return keptReport
}

func TestRecentEndpointLetterNumbers(t *testing.T) {
	db := newRecentDB(t)
	mustCreate(t, db, &models.AssignmentLetter{ID: uuid.New(), LetterNumber: "ST-001/SKAI/2026"})
	del := &models.AssignmentLetter{ID: uuid.New(), LetterNumber: "ST-009/SKAI/2026"}
	mustCreate(t, db, del)
	if err := db.Delete(del).Error; err != nil {
		t.Fatal(err)
	}
	kept := seedByNumber(t, db)

	resp := getRecent(t, db, "?limit=50")
	if got := strings.Join(titles(resp.Data.Items), "|"); got != "Live KKA finding|No-letter report finding|Live report finding" {
		t.Fatalf("items = %s", got)
	}
	if resp.Data.Total != 3 {
		t.Errorf("total = %d", resp.Data.Total)
	}
	if it := resp.Data.Items[2]; it.ReportID == nil || *it.ReportID != kept.String() || it.AssignmentLetterID != "st-001/skai/2026" {
		t.Errorf("report item = %+v", it)
	}
}

func TestRecentEndpointEmptyLettersTableKeepsOrphans(t *testing.T) {
	db := newRecentDB(t)
	seedByNumber(t, db)
	resp := getRecent(t, db, "?limit=50")
	if resp.Data.Total != 8 {
		t.Errorf("total = %d, want 8 (%v)", resp.Data.Total, titles(resp.Data.Items))
	}
}

func TestRecentEndpointUnreadableLettersTable(t *testing.T) {
	db := newRecentDB(t)
	seedByNumber(t, db)
	if err := db.Exec("DROP TABLE assignment_letters").Error; err != nil {
		t.Fatal(err)
	}
	resp := getRecent(t, db, "?limit=50")
	if !resp.Success || resp.Data.Total != 8 {
		t.Errorf("success %v total %d, want filter skipped", resp.Success, resp.Data.Total)
	}
}
