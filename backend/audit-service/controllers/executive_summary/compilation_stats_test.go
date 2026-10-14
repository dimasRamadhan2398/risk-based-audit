package executivesummary

import (
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

func TestLoadCompilationStats(t *testing.T) {
	db := newDB(t)

	// Letters: two of 2026 (one soft-deleted), one of 2025
	mustCreate(t, db, &models.AssignmentLetter{ID: uuid.New(), LetterNumber: "ST-001/A/2026", AuditYear: "2026"})
	mustCreate(t, db, &models.AssignmentLetter{ID: uuid.New(), LetterNumber: "ST-002/A/2026", AuditYear: "2026"})
	deletedLetter := models.AssignmentLetter{ID: uuid.New(), LetterNumber: "ST-003/A/2026", AuditYear: "2026"}
	mustCreate(t, db, &deletedLetter)
	db.Delete(&deletedLetter)
	mustCreate(t, db, &models.AssignmentLetter{ID: uuid.New(), LetterNumber: "ST-001/A/2025", AuditYear: "2025"})

	// LHA: two in 2026, one soft-deleted, one in 2025
	lha := func(letter string) *models.AuditResultReport {
		r := &models.AuditResultReport{ID: uuid.New(), AssignmentLetterID: letter}
		mustCreate(t, db, r)
		return r
	}
	lha1 := lha("ST-001/A/2026")
	lha2 := lha("ST-002/A/2026")
	del := lha("ST-002/A/2026")
	db.Delete(del)
	lha("ST-001/A/2025")

	// Test controls: recommendation counted when non-blank, year and soft delete respected
	tc := func(letter, rec string) {
		mustCreate(t, db, &models.FieldworkTestControl{ID: uuid.New(), AssignmentLetterID: letter, Recommendation: rec})
	}
	tc("ST-001/A/2026", "Perkuat MFA")
	tc("ST-001/A/2026", "   ")
	tc("ST-002/A/2026", "")
	tc("ST-002/A/2026", "Revisi SOP")
	tc("ST-001/A/2025", "Tahun lalu")
	gone := models.FieldworkTestControl{ID: uuid.New(), AssignmentLetterID: "ST-001/A/2026", Recommendation: "Dihapus"}
	mustCreate(t, db, &gone)
	if err := db.Delete(&gone).Error; err != nil {
		t.Fatal(err)
	}

	// Working paper risks keyed by letter number
	risk := func(letter, level string) {
		mustCreate(t, db, &models.WorkingPaperRisk{ID: uuid.New(), WorkingPaperID: letter, RiskLevel: level})
	}
	risk("ST-001/A/2026", "High")
	risk("ST-001/A/2026", "CRITICAL")
	risk("ST-002/A/2026", "Medium")
	risk("ST-002/A/2026", "low")
	risk("ST-002/A/2026", "")
	risk("ST-001/A/2025", "High")

	// ATRs of the 2026 LHAs
	atr := func(report *models.AuditResultReport, status string, due *time.Time) *models.ActionTakenReport {
		a := &models.ActionTakenReport{ID: uuid.New(), FindingID: uuid.NewString(), Status: status, DueDate: due}
		if report != nil {
			id := report.ID
			a.AuditResultReportID = &id
		}
		mustCreate(t, db, a)
		return a
	}
	atr(lha1, models.ATRStatusCompleted, ptrTime(2026, 9, 1))
	atr(lha1, "completed", nil)
	atr(lha1, models.ATRStatusPlanned, ptrTime(2026, 10, 1))    // overdue
	atr(lha2, models.ATRStatusInProgress, ptrTime(2026, 12, 1)) // in progress
	atr(lha2, models.ATRStatusPendingReview, nil)               // in progress
	atr(lha2, "Unknown legacy", ptrTime(2026, 9, 30))           // open, overdue
	atr(lha2, models.ATRStatusCancelled, ptrTime(2026, 9, 30))  // cancelled, not in table
	softATR := atr(lha2, models.ATRStatusCompleted, nil)
	db.Delete(softATR)                        // soft-deleted
	atr(del, models.ATRStatusPlanned, nil)    // LHA deleted: excluded
	atr(nil, models.ATRStatusPlanned, nil)    // legacy, no LHA: excluded
	atr(lha1, models.ATRStatusCompleted, nil) // counted below
	atrOther := &models.ActionTakenReport{ID: uuid.New(), FindingID: uuid.NewString(), Status: models.ATRStatusCompleted}
	mustCreate(t, db, atrOther)
	other := lha("ST-001/A/2025")
	atr(other, models.ATRStatusCompleted, nil) // other year: excluded

	s, err := LoadCompilationStats(db, 2026, testNow)
	if err != nil {
		t.Fatal(err)
	}

	if s.JumlahLaporan != 2 {
		t.Errorf("jumlahLaporan = %d, want 2", s.JumlahLaporan)
	}
	if s.JumlahRekomendasi != 2 {
		t.Errorf("jumlahRekomendasi = %d, want 2", s.JumlahRekomendasi)
	}
	if s.RisikoTinggi != 2 || s.RisikoSedang != 1 || s.RisikoRendah != 1 || s.RisikoLainnya != 1 {
		t.Errorf("risk breakdown = %+v", s)
	}
	// Closed: 2 COMPLETED on lha1 + 1 COMPLETED on lha1 (the "completed" one and the last one) -> 3
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
}

func TestLoadCompilationStatsEmptyYear(t *testing.T) {
	db := newDB(t)
	s, err := LoadCompilationStats(db, 2030, testNow)
	if err != nil {
		t.Fatal(err)
	}
	if s.JumlahLaporan != 0 || s.JumlahRekomendasi != 0 || s.RisikoTinggi+s.RisikoSedang+s.RisikoRendah+s.RisikoLainnya != 0 || s.FollowUpCancelled != 0 {
		t.Fatalf("stats %+v", s)
	}
	if len(s.FollowUpTable) != 3 || s.FollowUpTable[0].Status != FollowUpClosed || s.FollowUpTable[1].Status != FollowUpInProgress || s.FollowUpTable[2].Status != FollowUpOverdue {
		t.Fatalf("table rows must always be present: %+v", s.FollowUpTable)
	}
}

func TestRiskBucket(t *testing.T) {
	cases := map[string]string{
		"High": "Tinggi", " very high ": "Tinggi", "CRITICAL": "Tinggi", "Tinggi": "Tinggi",
		"Medium": "Sedang", "moderate": "Sedang", "Sedang": "Sedang",
		"Low": "Rendah", "rendah": "Rendah", "VERY_LOW": "",
		"": "", "unknown": "",
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
	mustCreate(t, db, &models.AssignmentLetter{ID: uuid.New(), LetterNumber: "ST-001/A/2026", AuditYear: "2026"})

	r := gin.New()
	r.GET("/executive-summary/compilation-stats", CompilationStatsHandler(db))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/executive-summary/compilation-stats?year=2026", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	for _, key := range []string{`"year":2026`, `"jumlahLaporan":0`, `"jumlahRekomendasi":0`, `"risikoTinggi":0`, `"risikoSedang":0`, `"risikoRendah":0`, `"risikoLainnya":0`, `"followUpTable":[`, `"followUpCancelled":0`} {
		if !strings.Contains(body, key) {
			t.Errorf("response missing %s: %s", key, body)
		}
	}
}
