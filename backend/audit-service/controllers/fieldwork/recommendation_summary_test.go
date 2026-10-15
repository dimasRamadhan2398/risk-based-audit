package fieldwork

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"audit-service/models"
	"audit-service/pkg/testdb"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func newDB(t *testing.T) *gorm.DB {
	t.Helper()
	return testdb.New(t,
		&models.AssignmentLetter{},
		&models.AuditResultReport{},
		&models.FieldworkTestControl{},
	)
}

func mustCreate(t *testing.T, db *gorm.DB, v interface{}) {
	t.Helper()
	if err := db.Create(v).Error; err != nil {
		t.Fatal(err)
	}
}

type fixture struct {
	lhaByNumber, lhaByUUID, lhaOther, lhaNoLetter, lhaDeleted uuid.UUID
}

func seed(t *testing.T, db *gorm.DB) fixture {
	t.Helper()
	l1 := models.AssignmentLetter{ID: uuid.New(), LetterNumber: "ST-001/A/2026", AuditYear: "2026"}
	l2 := models.AssignmentLetter{ID: uuid.New(), LetterNumber: "ST-002/A/2026", AuditYear: "2026"}
	mustCreate(t, db, &l1)
	mustCreate(t, db, &l2)

	lha := func(letter string) uuid.UUID {
		r := &models.AuditResultReport{ID: uuid.New(), AssignmentLetterID: letter}
		mustCreate(t, db, r)
		return r.ID
	}
	f := fixture{
		lhaByNumber: lha(l1.LetterNumber),
		lhaByUUID:   lha(l2.ID.String()),  // LHA stores the letter UUID
		lhaOther:    lha("ST-404/A/2026"), // letter that does not exist
		lhaNoLetter: lha(""),
	}
	f.lhaDeleted = lha(l1.LetterNumber)
	if err := db.Delete(&models.AuditResultReport{}, "id = ?", f.lhaDeleted).Error; err != nil {
		t.Fatal(err)
	}

	tc := func(letter, rec string) *models.FieldworkTestControl {
		c := &models.FieldworkTestControl{ID: uuid.New(), AssignmentLetterID: letter, Recommendation: rec}
		mustCreate(t, db, c)
		return c
	}
	// Letter 1: 2 by number + 1 by UUID; blank and whitespace ignored; deleted ignored
	tc(l1.LetterNumber, "Perkuat MFA")
	tc(l1.LetterNumber, "Revisi SOP")
	tc(l1.ID.String(), "Rekonsiliasi bulanan")
	tc(l1.LetterNumber, "")
	tc(l1.LetterNumber, "   ")
	gone := tc(l1.LetterNumber, "Dihapus")
	if err := db.Delete(gone).Error; err != nil {
		t.Fatal(err)
	}
	// Letter 2 (LHA stores its UUID): 1 by number + 1 by UUID
	tc(l2.LetterNumber, "Segregasi tugas")
	tc(l2.ID.String(), "Review akses")
	// Unresolvable key, matched literally
	tc("ST-404/A/2026", "Tanpa surat tugas")
	return f
}

func TestLoadRecommendationSummary(t *testing.T) {
	db := newDB(t)
	f := seed(t, db)

	unknown := uuid.NewString()
	ids := []string{f.lhaByNumber.String(), f.lhaByUUID.String(), f.lhaOther.String(), f.lhaNoLetter.String(), f.lhaDeleted.String(), unknown, "not-a-uuid"}
	s, err := LoadRecommendationSummary(db, ids)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Items) != len(ids) {
		t.Fatalf("items = %+v", s.Items)
	}
	want := []int64{3, 2, 1, 0, 0, 0, 0}
	for i, it := range s.Items {
		if it.AuditResultReportID != ids[i] {
			t.Errorf("item %d id = %q, want %q (order must follow the request)", i, it.AuditResultReportID, ids[i])
		}
		if it.Count != want[i] {
			t.Errorf("item %d (%s) count = %d, want %d", i, it.AssignmentLetterID, it.Count, want[i])
		}
	}
	if s.Items[0].AssignmentLetterID != "ST-001/A/2026" || s.Items[5].AssignmentLetterID != "" {
		t.Errorf("assignmentLetterId = %q / %q", s.Items[0].AssignmentLetterID, s.Items[5].AssignmentLetterID)
	}
	if s.Total != 6 {
		t.Errorf("total = %d, want 6", s.Total)
	}
}

func TestParseReportIDs(t *testing.T) {
	got := ParseReportIDs(" a, ,b,a,, c ")
	if strings.Join(got, "|") != "a|b|c" {
		t.Fatalf("ParseReportIDs = %v", got)
	}
	if len(ParseReportIDs("")) != 0 || len(ParseReportIDs(" , ")) != 0 {
		t.Fatal("empty input must yield no ids")
	}
}

func TestRecommendationSummaryHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newDB(t)
	f := seed(t, db)

	r := gin.New()
	r.GET("/fieldwork/test-controls/recommendations/summary", RecommendationSummaryHandler(db))
	r.GET("/fieldwork/test-controls/:id", func(c *gin.Context) { c.Status(http.StatusTeapot) })

	get := func(query string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/fieldwork/test-controls/recommendations/summary"+query, nil))
		return w
	}

	w := get("?audit_result_report_ids=" + f.lhaByNumber.String() + "," + f.lhaByUUID.String())
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
	var body struct {
		Success bool                  `json:"success"`
		Data    RecommendationSummary `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.Success || body.Data.Total != 5 || len(body.Data.Items) != 2 {
		t.Fatalf("body %s", w.Body.String())
	}
	for _, key := range []string{`"total":5`, `"items":[`, `"auditResultReportId":"` + f.lhaByNumber.String() + `"`, `"assignmentLetterId":"ST-001/A/2026"`, `"count":3`} {
		if !strings.Contains(w.Body.String(), key) {
			t.Errorf("response missing %s: %s", key, w.Body.String())
		}
	}

	for _, q := range []string{"", "?audit_result_report_ids=", "?audit_result_report_ids=%20,%20"} {
		if w := get(q); w.Code != http.StatusBadRequest {
			t.Errorf("query %q: status %d, want 400", q, w.Code)
		}
	}

	many := make([]string, MaxRecommendationSummaryIDs+1)
	for i := range many {
		many[i] = uuid.NewString()
	}
	if w := get("?audit_result_report_ids=" + strings.Join(many, ",")); w.Code != http.StatusBadRequest {
		t.Errorf("over cap: status %d, want 400", w.Code)
	}
	if w := get("?audit_result_report_ids=" + strings.Join(many[:MaxRecommendationSummaryIDs], ",")); w.Code != http.StatusOK {
		t.Errorf("at cap: status %d, want 200", w.Code)
	}
}

// A soft-deleted letter does not resolve: its UUID-keyed test controls are
// not counted, only those holding the LHA's stored key literally.
func TestRecommendationSummaryDeletedLetter(t *testing.T) {
	db := newDB(t)
	l := models.AssignmentLetter{ID: uuid.New(), LetterNumber: "ST-009/A/2026", AuditYear: "2026"}
	mustCreate(t, db, &l)
	if err := db.Delete(&l).Error; err != nil {
		t.Fatal(err)
	}
	r := &models.AuditResultReport{ID: uuid.New(), AssignmentLetterID: l.LetterNumber}
	mustCreate(t, db, r)
	mustCreate(t, db, &models.FieldworkTestControl{ID: uuid.New(), AssignmentLetterID: l.LetterNumber, Recommendation: "A"})
	mustCreate(t, db, &models.FieldworkTestControl{ID: uuid.New(), AssignmentLetterID: l.ID.String(), Recommendation: "B"})

	s, err := LoadRecommendationSummary(db, []string{r.ID.String()})
	if err != nil {
		t.Fatal(err)
	}
	if s.Total != 1 || s.Items[0].Count != 1 {
		t.Fatalf("summary %+v", s)
	}
}
