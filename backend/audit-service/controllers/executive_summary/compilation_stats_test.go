package executivesummary

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"audit-service/models"
	"audit-service/pkg/testdb"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var testNow = time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)

func ptrTime(y int, m time.Month, d int) *time.Time {
	t := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return &t
}

func newDB(t *testing.T) *gorm.DB {
	t.Helper()
	return testdb.New(t,
		&models.AssignmentLetter{},
		&models.AuditResultReport{},
		&models.FieldworkTestControl{},
		&models.WorkingPaperRisk{},
		&models.ActionTakenReport{},
		&models.ActionTakenReportEvidence{},
	)
}

func mustCreate(t *testing.T, db *gorm.DB, v interface{}) {
	t.Helper()
	if err := db.Create(v).Error; err != nil {
		t.Fatal(err)
	}
}

// lhaOn creates an LHA of a letter (by number or UUID) with a status and an optional report date
func lhaOn(t *testing.T, db *gorm.DB, letter, status string, reportDate *time.Time) *models.AuditResultReport {
	t.Helper()
	r := &models.AuditResultReport{ID: uuid.New(), AssignmentLetterID: letter, Status: status, ReportDate: reportDate}
	mustCreate(t, db, r)
	return r
}

func letterOf(t *testing.T, db *gorm.DB, number, year string) {
	t.Helper()
	mustCreate(t, db, &models.AssignmentLetter{ID: uuid.New(), LetterNumber: number, AuditYear: year})
}

func tcOf(t *testing.T, db *gorm.DB, letter, rec string) {
	t.Helper()
	mustCreate(t, db, &models.FieldworkTestControl{ID: uuid.New(), AssignmentLetterID: letter, Recommendation: rec})
}

func riskOf(t *testing.T, db *gorm.DB, letter, level string) {
	t.Helper()
	mustCreate(t, db, &models.WorkingPaperRisk{ID: uuid.New(), WorkingPaperID: letter, RiskLevel: level})
}

// atrOf creates an ATR of an LHA (nil LHA: legacy row without a link)
func atrOf(t *testing.T, db *gorm.DB, report *models.AuditResultReport, status string, due *time.Time) *models.ActionTakenReport {
	t.Helper()
	a := &models.ActionTakenReport{ID: uuid.New(), FindingID: uuid.NewString(), Status: status, DueDate: due}
	if report != nil {
		id := report.ID
		a.AuditResultReportID = &id
	}
	mustCreate(t, db, a)
	return a
}

func followUp(s CompilationStats, status string) int64 {
	for _, row := range s.FollowUpTable {
		if row.Status == status {
			return row.Jumlah
		}
	}
	return -1
}

func TestLoadCompilationStats(t *testing.T) {
	db := newDB(t)

	// Letters: two of 2026 (one soft-deleted), one of 2025
	letterOf(t, db, "ST-001/A/2026", "2026")
	letterOf(t, db, "ST-002/A/2026", "2026")
	deletedLetter := models.AssignmentLetter{ID: uuid.New(), LetterNumber: "ST-003/A/2026", AuditYear: "2026"}
	mustCreate(t, db, &deletedLetter)
	db.Delete(&deletedLetter)
	letterOf(t, db, "ST-001/A/2025", "2025")

	// LHAs of 2026: approved ones count (Final, APPROVED); a draft, a deleted
	// one and an LHA of another year do not. lha3 shares letter ST-002 by UUID.
	lha1 := lhaOn(t, db, "ST-001/A/2026", "Final", nil) // no report date: still in the year
	lha2 := lhaOn(t, db, "ST-002/A/2026", "APPROVED", nil)
	lhaShared := lhaOn(t, db, letterUUID(t, db, "ST-002/A/2026"), "Final", nil)
	draft := lhaOn(t, db, "ST-002/A/2026", "Draft", nil)
	del := lhaOn(t, db, "ST-002/A/2026", "Final", nil)
	db.Delete(del)
	lhaOn(t, db, "ST-001/A/2025", "Final", nil)

	// Test controls: recommendation counted when non-blank, year and soft delete respected
	tcOf(t, db, "ST-001/A/2026", "Perkuat MFA")
	tcOf(t, db, "ST-001/A/2026", "   ")
	tcOf(t, db, "ST-002/A/2026", "")
	tcOf(t, db, "ST-002/A/2026", "Revisi SOP")
	tcOf(t, db, "ST-001/A/2025", "Tahun lalu")
	gone := models.FieldworkTestControl{ID: uuid.New(), AssignmentLetterID: "ST-001/A/2026", Recommendation: "Dihapus"}
	mustCreate(t, db, &gone)
	if err := db.Delete(&gone).Error; err != nil {
		t.Fatal(err)
	}

	// Working paper risks keyed by letter number
	riskOf(t, db, "ST-001/A/2026", "high")
	riskOf(t, db, "ST-001/A/2026", "CRITICAL")
	riskOf(t, db, "ST-001/A/2026", "moderate-high")
	riskOf(t, db, "ST-001/A/2026", "Moderate to High")
	riskOf(t, db, "ST-002/A/2026", "moderate")
	riskOf(t, db, "ST-002/A/2026", "Medium")
	riskOf(t, db, "ST-002/A/2026", "SEDANG")
	riskOf(t, db, "ST-002/A/2026", "low-moderate")
	riskOf(t, db, "ST-002/A/2026", "low")
	riskOf(t, db, "ST-002/A/2026", "Rendah")
	riskOf(t, db, "ST-002/A/2026", "")
	riskOf(t, db, "ST-002/A/2026", "extreme")
	riskOf(t, db, "ST-001/A/2025", "high")

	// ATRs of the approved 2026 LHAs
	atrOf(t, db, lha1, models.ATRStatusCompleted, ptrTime(2026, 9, 1))
	atrOf(t, db, lha1, "completed", nil)
	atrOf(t, db, lha1, models.ATRStatusPlanned, ptrTime(2026, 10, 1))    // overdue
	atrOf(t, db, lha2, models.ATRStatusInProgress, ptrTime(2026, 12, 1)) // in progress
	atrOf(t, db, lha2, models.ATRStatusPendingReview, nil)               // in progress
	atrOf(t, db, lha2, "Unknown legacy", ptrTime(2026, 9, 30))           // open, overdue
	atrOf(t, db, lha2, models.ATRStatusCancelled, ptrTime(2026, 9, 30))  // cancelled, not in table
	softATR := atrOf(t, db, lha2, models.ATRStatusCompleted, nil)
	db.Delete(softATR)                                      // soft-deleted
	atrOf(t, db, del, models.ATRStatusPlanned, nil)         // LHA deleted: excluded
	atrOf(t, db, draft, models.ATRStatusPlanned, nil)       // LHA not approved: excluded
	atrOf(t, db, nil, models.ATRStatusPlanned, nil)         // legacy, no LHA: excluded
	atrOf(t, db, lhaShared, models.ATRStatusCompleted, nil) // shared letter, own LHA: counted
	other := lhaOn(t, db, "ST-001/A/2025", "Final", nil)
	atrOf(t, db, other, models.ATRStatusCompleted, nil) // other year: excluded

	s, err := LoadCompilationStats(db, 2026, 0, testNow)
	if err != nil {
		t.Fatal(err)
	}

	// Approved LHAs of 2026: lha1, lha2, lhaShared (lhaShared shares the letter of lha2)
	if s.JumlahLaporan != 3 {
		t.Errorf("jumlahLaporan = %d, want 3", s.JumlahLaporan)
	}
	// Letter-level count: ST-001 has 1, ST-002 has 1; lhaShared does not repeat ST-002
	if s.JumlahRekomendasi != 2 {
		t.Errorf("jumlahRekomendasi = %d, want 2", s.JumlahRekomendasi)
	}
	if s.RisikoHigh != 2 || s.RisikoModerateHigh != 2 || s.RisikoModerate != 3 || s.RisikoLowModerate != 1 || s.RisikoLow != 2 || s.RisikoLainnya != 2 {
		t.Errorf("risk breakdown = %+v", s)
	}
	if s.TotalTemuan != 10 {
		t.Errorf("totalTemuan = %d, want 10 (five levels, without risikoLainnya)", s.TotalTemuan)
	}
	// Closed: 2 COMPLETED on lha1 + 1 COMPLETED on lhaShared -> 3
	// Overdue: PLANNED due 2026-10-01 and "Unknown legacy" due 2026-09-30 -> 2
	// In Progress: IN_PROGRESS (future due), PENDING_REVIEW (no due) -> 2
	want := map[string]int64{FollowUpClosed: 3, FollowUpInProgress: 2, FollowUpOverdue: 2}
	if len(s.FollowUpTable) != 3 {
		t.Fatalf("followUpTable = %+v", s.FollowUpTable)
	}
	for _, row := range s.FollowUpTable {
		if row.Jumlah != want[row.Status] {
			t.Errorf("%s = %d, want %d", row.Status, row.Jumlah, want[row.Status])
		}
	}
	if s.FollowUpCancelled != 1 {
		t.Errorf("followUpCancelled = %d, want 1", s.FollowUpCancelled)
	}
	if s.Quarter != 0 {
		t.Errorf("quarter = %d, want 0 for the full year", s.Quarter)
	}
	if !s.SyncedAt.Equal(testNow) {
		t.Errorf("syncedAt = %v, want %v", s.SyncedAt, testNow)
	}
}

// letterUUID returns the UUID of a letter, to reference it from an LHA
func letterUUID(t *testing.T, db *gorm.DB, number string) string {
	t.Helper()
	var l models.AssignmentLetter
	if err := db.Where("letter_number = ?", number).Take(&l).Error; err != nil {
		t.Fatal(err)
	}
	return l.ID.String()
}

// Quarters are read from the LHA report date; letters are kept to those with an
// approved LHA in the period; boundary dates fall in the quarter they start.
func TestLoadCompilationStatsQuarter(t *testing.T) {
	db := newDB(t)
	letterOf(t, db, "ST-010/A/2026", "2026") // A
	letterOf(t, db, "ST-011/A/2026", "2026") // B
	letterOf(t, db, "ST-012/A/2026", "2026") // C

	// A: two Q1 LHAs (shared letter) and one Q2 LHA on the first day of Q2
	a1 := lhaOn(t, db, "ST-010/A/2026", "Final", ptrTime(2026, 2, 10))
	lhaOn(t, db, "ST-010/A/2026", "Approved", ptrTime(2026, 3, 31)) // last day of Q1
	aQ2 := lhaOn(t, db, "ST-010/A/2026", "Final", ptrTime(2026, 4, 1))
	// C: one Q1 LHA
	c1 := lhaOn(t, db, "ST-012/A/2026", "Final", ptrTime(2026, 2, 15))
	// B: approved LHA without a date (year only), approved LHA in Q3, draft LHA in Q2 (never counted)
	bNoDate := lhaOn(t, db, "ST-011/A/2026", "Final", nil)
	bQ3 := lhaOn(t, db, "ST-011/A/2026", "Final", ptrTime(2026, 7, 1))
	lhaOn(t, db, "ST-011/A/2026", "Draft", ptrTime(2026, 5, 1))

	tcOf(t, db, "ST-010/A/2026", "Rek A1")
	tcOf(t, db, "ST-010/A/2026", "Rek A2")
	tcOf(t, db, "ST-012/A/2026", "Rek C1")
	tcOf(t, db, "ST-011/A/2026", "Rek B1")

	riskOf(t, db, "ST-010/A/2026", "high")
	riskOf(t, db, "ST-012/A/2026", "low")
	riskOf(t, db, "ST-011/A/2026", "moderate")

	atrOf(t, db, a1, models.ATRStatusPlanned, ptrTime(2026, 1, 31))   // Q1, overdue
	atrOf(t, db, c1, models.ATRStatusInProgress, nil)                 // Q1, in progress
	atrOf(t, db, aQ2, models.ATRStatusCompleted, nil)                 // Q2, closed
	atrOf(t, db, bQ3, models.ATRStatusCancelled, ptrTime(2026, 1, 1)) // Q3, cancelled
	atrOf(t, db, bNoDate, models.ATRStatusPlanned, nil)               // year only, open
	lhaDraftQ2 := lhaOn(t, db, "ST-011/A/2026", "Draft", ptrTime(2026, 6, 1))
	atrOf(t, db, lhaDraftQ2, models.ATRStatusPlanned, ptrTime(2026, 1, 1)) // draft: never counted

	type want struct {
		laporan, rekomendasi, temuan, closed, inProgress, overdue, cancelled int64
		high, low, moderate                                                  int64
	}
	cases := []struct {
		quarter int
		want    want
	}{
		// Q1: lhaA1, lhaA2, c1 (A's two LHAs count one A letter's recommendations)
		{1, want{laporan: 3, rekomendasi: 3, temuan: 2, closed: 0, inProgress: 1, overdue: 1, cancelled: 0, high: 1, low: 1, moderate: 0}},
		// Q2: aQ2 only; letter A is in scope, letter B has no approved LHA in Q2
		{2, want{laporan: 1, rekomendasi: 2, temuan: 1, closed: 1, inProgress: 0, overdue: 0, cancelled: 0, high: 1, low: 0, moderate: 0}},
		// Q3: bQ3 only; its cancelled ATR is out of the table
		{3, want{laporan: 1, rekomendasi: 1, temuan: 1, closed: 0, inProgress: 0, overdue: 0, cancelled: 1, high: 0, low: 0, moderate: 1}},
		// Q4: nothing, but the table rows are present
		{4, want{}},
		// Year: all approved LHAs of the year, including bNoDate (open, no due date: in progress)
		{0, want{laporan: 6, rekomendasi: 4, temuan: 3, closed: 1, inProgress: 2, overdue: 1, cancelled: 1, high: 1, low: 1, moderate: 1}},
	}
	for _, c := range cases {
		s, err := LoadCompilationStats(db, 2026, c.quarter, testNow)
		if err != nil {
			t.Fatalf("quarter %d: %v", c.quarter, err)
		}
		w := c.want
		if s.Quarter != c.quarter {
			t.Errorf("quarter %d: quarter field = %d", c.quarter, s.Quarter)
		}
		if s.JumlahLaporan != w.laporan || s.JumlahRekomendasi != w.rekomendasi || s.TotalTemuan != w.temuan {
			t.Errorf("quarter %d: laporan/rekomendasi/temuan = %d/%d/%d, want %d/%d/%d", c.quarter,
				s.JumlahLaporan, s.JumlahRekomendasi, s.TotalTemuan, w.laporan, w.rekomendasi, w.temuan)
		}
		if s.RisikoHigh != w.high || s.RisikoLow != w.low || s.RisikoModerate != w.moderate {
			t.Errorf("quarter %d: risks high/moderate/low = %d/%d/%d", c.quarter, s.RisikoHigh, s.RisikoModerate, s.RisikoLow)
		}
		if followUp(s, FollowUpClosed) != w.closed || followUp(s, FollowUpInProgress) != w.inProgress || followUp(s, FollowUpOverdue) != w.overdue {
			t.Errorf("quarter %d: follow-up closed/inProgress/overdue = %d/%d/%d", c.quarter,
				followUp(s, FollowUpClosed), followUp(s, FollowUpInProgress), followUp(s, FollowUpOverdue))
		}
		if s.FollowUpCancelled != w.cancelled {
			t.Errorf("quarter %d: followUpCancelled = %d, want %d", c.quarter, s.FollowUpCancelled, w.cancelled)
		}
		if len(s.FollowUpTable) != 3 {
			t.Errorf("quarter %d: followUpTable rows = %d", c.quarter, len(s.FollowUpTable))
		}
	}
}

func TestLoadCompilationStatsInvalidQuarter(t *testing.T) {
	db := newDB(t)
	for _, q := range []int{-1, 5} {
		if _, err := LoadCompilationStats(db, 2026, q, testNow); !errors.Is(err, ErrInvalidQuarter) {
			t.Errorf("quarter %d: err = %v, want ErrInvalidQuarter", q, err)
		}
	}
}

func TestLoadCompilationStatsEmptyYear(t *testing.T) {
	db := newDB(t)
	s, err := LoadCompilationStats(db, 2030, 0, testNow)
	if err != nil {
		t.Fatal(err)
	}
	if s.JumlahLaporan != 0 || s.JumlahRekomendasi != 0 || s.TotalTemuan != 0 || s.RisikoLainnya != 0 || s.FollowUpCancelled != 0 {
		t.Fatalf("stats %+v", s)
	}
	if len(s.FollowUpTable) != 3 || s.FollowUpTable[0].Status != FollowUpClosed || s.FollowUpTable[1].Status != FollowUpInProgress || s.FollowUpTable[2].Status != FollowUpOverdue {
		t.Fatalf("table rows must always be present: %+v", s.FollowUpTable)
	}
	if s.SyncedAt.IsZero() {
		t.Error("syncedAt must be set on empty data")
	}
}

func TestRiskBucket(t *testing.T) {
	cases := map[string]string{
		"high": RiskHigh, "High": RiskHigh, "HIGH": RiskHigh, " very high ": RiskHigh, "CRITICAL": RiskHigh, "Tinggi": RiskHigh,
		"moderate-high": RiskModerateHigh, "MODERATE_HIGH": RiskModerateHigh, "Moderate to High": RiskModerateHigh, "moderate  to  high": RiskModerateHigh,
		"moderate": RiskModerate, "MEDIUM": RiskModerate, "Moderate": RiskModerate, "sedang": RiskModerate,
		"low-moderate": RiskLowModerate, "LOW_MODERATE": RiskLowModerate, "Low to Moderate": RiskLowModerate,
		"low": RiskLow, "LOW": RiskLow, "rendah": RiskLow, "RENDAH": RiskLow,
		"VERY_LOW": "", "high-moderate": "", "": "", "unknown": "",
	}
	for in, want := range cases {
		got, _ := RiskBucket(in)
		if got != want {
			t.Errorf("RiskBucket(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCompilationStatsHandlerJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newDB(t)
	letterOf(t, db, "ST-001/A/2026", "2026")

	r := gin.New()
	r.GET("/executive-summary/compilation-stats", CompilationStatsHandler(db))

	get := func(query string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/executive-summary/compilation-stats"+query, nil))
		return w
	}

	w := get("?year=2026")
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	for _, old := range []string{`"risikoTinggi"`, `"risikoSedang"`, `"risikoRendah"`} {
		if strings.Contains(body, old) {
			t.Errorf("response still carries %s: %s", old, body)
		}
	}
	for _, key := range []string{`"year":2026`, `"quarter":0`, `"syncedAt":"`, `"jumlahLaporan":0`, `"jumlahRekomendasi":0`, `"totalTemuan":0`, `"risikoHigh":0`, `"risikoModerateHigh":0`, `"risikoModerate":0`, `"risikoLowModerate":0`, `"risikoLow":0`, `"risikoLainnya":0`, `"followUpTable":[`, `"followUpCancelled":0`} {
		if !strings.Contains(body, key) {
			t.Errorf("response missing %s: %s", key, body)
		}
	}

	if w := get("?year=2026&quarter=2"); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"quarter":2`) {
		t.Errorf("quarter=2: status %d: %s", w.Code, w.Body.String())
	}
	for _, q := range []string{"?quarter=0", "?quarter=5", "?quarter=abc", "?quarter=-1"} {
		if w := get(q); w.Code != http.StatusBadRequest {
			t.Errorf("query %q: status %d, want 400", q, w.Code)
		}
	}
}

// The handler reads the whole response from one decode, so a JSON round trip
// checks the field names the frontend uses.
func TestCompilationStatsJSONContract(t *testing.T) {
	raw, err := json.Marshal(CompilationStats{Year: 2026, Quarter: 1, SyncedAt: testNow, FollowUpTable: []FollowUpRow{{Status: FollowUpClosed}}})
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"year", "quarter", "syncedAt", "jumlahLaporan", "jumlahRekomendasi", "totalTemuan", "risikoHigh", "risikoModerateHigh", "risikoModerate", "risikoLowModerate", "risikoLow", "risikoLainnya", "followUpTable", "followUpCancelled"} {
		if _, ok := fields[k]; !ok {
			t.Errorf("json field %s missing", k)
		}
	}
}
