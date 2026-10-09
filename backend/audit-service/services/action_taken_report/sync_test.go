package action_taken_report

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"audit-service/controllers/crud"
	"audit-service/models"
	"audit-service/pkg/testdb"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var now = time.Date(2026, 10, 9, 10, 0, 0, 0, time.UTC)

func newDB(t *testing.T) *gorm.DB {
	t.Helper()
	return testdb.New(t, &models.AuditResultReport{}, &models.ActionTakenReport{}, &models.ActionTakenReportEvidence{})
}

func createReport(t *testing.T, db *gorm.DB, status string, titles ...string) *models.AuditResultReport {
	t.Helper()
	r := &models.AuditResultReport{
		ID: uuid.New(), ReportNumber: "LHA-0" + uuid.NewString()[:4] + "/SKAI/2026", ReportTitle: "Audit Kas", Status: status,
		Findings: []models.AuditReportFinding{},
	}
	for _, title := range titles {
		r.Findings = append(r.Findings, models.AuditReportFinding{Title: title, Category: "Major", Action: "Fix " + title})
	}
	if err := db.Create(r).Error; err != nil {
		t.Fatal(err)
	}
	return r
}

func atrsOf(t *testing.T, db *gorm.DB, reportID uuid.UUID) []models.ActionTakenReport {
	t.Helper()
	var out []models.ActionTakenReport
	if err := db.Unscoped().Where("audit_result_report_id = ?", reportID).Order("finding_title").Find(&out).Error; err != nil {
		t.Fatal(err)
	}
	return out
}

func sync(t *testing.T, db *gorm.DB, id uuid.UUID) SyncResult {
	t.Helper()
	var res SyncResult
	err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		res, err = SyncReportByID(tx, id, now)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestCreateGivesFindingsStableIDs(t *testing.T) {
	db := newDB(t)
	r := createReport(t, db, "Draft", "A", "B")
	var stored models.AuditResultReport
	if err := db.First(&stored, "id = ?", r.ID).Error; err != nil {
		t.Fatal(err)
	}
	if len(stored.Findings) != 2 || stored.FindingsCount != 2 {
		t.Fatalf("findings %v count %d", stored.Findings, stored.FindingsCount)
	}
	for _, f := range stored.Findings {
		if _, err := uuid.Parse(f.ID); err != nil {
			t.Fatalf("finding %q has no uuid id: %q", f.Title, f.ID)
		}
	}
	if stored.Findings[0].ID == stored.Findings[1].ID {
		t.Fatal("findings share an id")
	}
}

func TestBackfillFindingIDsIsIdempotent(t *testing.T) {
	db := newDB(t)
	keep := uuid.NewString()
	// Rows written before ids existed (raw JSON, no hooks): no id, a valid id,
	// a duplicate and a junk id
	raw := `[{"title":"A"},{"id":"` + keep + `","title":"B"},{"id":"` + keep + `","title":"C"},{"id":"x","title":"D"}]`
	legacy := uuid.New()
	deleted := uuid.New()
	clean := uuid.New()
	for _, id := range []uuid.UUID{legacy, deleted, clean} {
		findings := raw
		if id == clean {
			findings = `[{"id":"` + uuid.NewString() + `","title":"ok"}]`
		}
		if err := db.Exec("INSERT INTO audit_result_reports (id, status, findings, findings_count) VALUES (?, 'Final', ?, 0)", id.String(), findings).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Exec("UPDATE audit_result_reports SET deleted_at = ? WHERE id = ?", now, deleted.String()).Error; err != nil {
		t.Fatal(err)
	}

	n, err := BackfillFindingIDs(db)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("first run updated %d reports, want 2 (the clean one untouched, soft-deleted one included)", n)
	}
	var r models.AuditResultReport
	if err := db.Unscoped().First(&r, "id = ?", legacy).Error; err != nil {
		t.Fatal(err)
	}
	if r.FindingsCount != 4 || models.FindingsNeedIDs(r.Findings) {
		t.Fatalf("after backfill: %+v count %d", r.Findings, r.FindingsCount)
	}
	if r.Findings[1].ID != keep {
		t.Fatalf("existing id not kept: %q", r.Findings[1].ID)
	}
	first := append([]models.AuditReportFinding(nil), r.Findings...)

	n, err = BackfillFindingIDs(db)
	if err != nil || n != 0 {
		t.Fatalf("second run updated %d (err %v), want 0", n, err)
	}
	var again models.AuditResultReport
	_ = db.Unscoped().First(&again, "id = ?", legacy).Error
	for i := range first {
		if first[i].ID != again.Findings[i].ID {
			t.Fatalf("finding %d id changed on rerun: %s -> %s", i, first[i].ID, again.Findings[i].ID)
		}
	}
}

func TestSyncCreatesOneATRPerFindingOnlyWhenApproved(t *testing.T) {
	db := newDB(t)
	r := createReport(t, db, "Draft", "Kas", "Stok")

	if res := sync(t, db, r.ID); res.Created != 0 {
		t.Fatalf("draft created %d ATRs", res.Created)
	}
	if got := atrsOf(t, db, r.ID); len(got) != 0 {
		t.Fatalf("draft has ATRs: %v", got)
	}

	for _, status := range []string{"Final", "final", " APPROVED ", "Published"} {
		if !models.IsLHAApproved(status) {
			t.Fatalf("%q not treated as approved", status)
		}
	}
	for _, status := range []string{"Draft", "Review", "", "Rejected"} {
		if models.IsLHAApproved(status) {
			t.Fatalf("%q treated as approved", status)
		}
	}

	db.Model(&models.AuditResultReport{}).Where("id = ?", r.ID).UpdateColumn("status", "Final")
	if res := sync(t, db, r.ID); res.Created != 2 {
		t.Fatalf("approval created %d ATRs, want 2", res.Created)
	}
	got := atrsOf(t, db, r.ID)
	var stored models.AuditResultReport
	db.First(&stored, "id = ?", r.ID)
	byFinding := map[string]models.ActionTakenReport{}
	for _, a := range got {
		byFinding[a.FindingID] = a
	}
	for _, f := range stored.Findings {
		a, ok := byFinding[f.ID]
		if !ok {
			t.Fatalf("no ATR for finding %s", f.Title)
		}
		if a.Status != models.ATRStatusPlanned || a.FindingTitle != f.Title || a.FindingCategory != "Major" ||
			a.Recommendation != "Fix "+f.Title || a.ReportNumber != r.ReportNumber || a.ReportTitle != "Audit Kas" {
			t.Fatalf("ATR copy wrong: %+v", a)
		}
	}

	// Re-approval / re-save is idempotent
	if res := sync(t, db, r.ID); res.Created != 0 {
		t.Fatalf("re-sync created %d", res.Created)
	}
	if got := atrsOf(t, db, r.ID); len(got) != 2 {
		t.Fatalf("after re-sync %d ATRs", len(got))
	}
}

func TestSyncWhenApprovedFindingsChange(t *testing.T) {
	db := newDB(t)
	r := createReport(t, db, "Final", "Kas", "Stok", "Vendor")
	sync(t, db, r.ID)
	var stored models.AuditResultReport
	db.First(&stored, "id = ?", r.ID)
	ids := map[string]string{}
	for _, f := range stored.Findings {
		ids[f.Title] = f.ID
	}
	// Stok is being worked on; Vendor's ATR was deleted by an admin
	db.Model(&models.ActionTakenReport{}).Where("finding_id = ?", ids["Stok"]).UpdateColumn("status", models.ATRStatusInProgress)
	db.Where("finding_id = ?", ids["Vendor"]).Delete(&models.ActionTakenReport{})

	// New finding list: Kas edited, Stok and Vendor removed, Baru added
	stored.Findings = []models.AuditReportFinding{
		{ID: ids["Kas"], Title: "Kas (revisi)", Category: "Minor", Action: "Rekonsiliasi"},
		{Title: "Baru", Category: "Major", Action: "Tindak"},
	}
	if err := db.Save(&stored).Error; err != nil {
		t.Fatal(err)
	}
	res := sync(t, db, r.ID)
	if res.Created != 1 {
		t.Fatalf("created %d, want 1 (only the new finding)", res.Created)
	}

	all := atrsOf(t, db, r.ID)
	by := map[string]models.ActionTakenReport{}
	for _, a := range all {
		by[a.FindingID] = a
	}
	kas := by[ids["Kas"]]
	if kas.FindingTitle != "Kas (revisi)" || kas.FindingCategory != "Minor" || kas.Recommendation != "Rekonsiliasi" || kas.Status != models.ATRStatusPlanned {
		t.Fatalf("PLANNED ATR copy not refreshed: %+v", kas)
	}
	if s := by[ids["Stok"]]; s.Status != models.ATRStatusInProgress || s.FindingTitle != "Stok" {
		t.Fatalf("worked-on ATR of a removed finding changed: %+v", s)
	}
	if v := by[ids["Vendor"]]; !v.DeletedAt.Valid {
		t.Fatalf("deleted ATR came back: %+v", v)
	}
	if len(all) != 4 {
		t.Fatalf("%d ATR rows, want 4 (Kas, Stok, deleted Vendor, Baru)", len(all))
	}

	// Removing an untouched finding cancels its ATR
	db.First(&stored, "id = ?", r.ID)
	var baruID string
	for _, f := range stored.Findings {
		if f.Title == "Baru" {
			baruID = f.ID
		}
	}
	stored.Findings = stored.Findings[:1]
	db.Save(&stored)
	if res := sync(t, db, r.ID); res.Cancelled != 1 {
		t.Fatalf("cancelled %d, want 1", res.Cancelled)
	}
	var baru models.ActionTakenReport
	db.First(&baru, "finding_id = ?", baruID)
	if baru.Status != models.ATRStatusCancelled || baru.ReviewNote != RemovedFindingNote || baru.ReviewedBy != SystemActor {
		t.Fatalf("removed finding's ATR: %+v", baru)
	}
}

func TestBackfillApprovedReports(t *testing.T) {
	db := newDB(t)
	a := createReport(t, db, "Final", "A1", "A2")
	b := createReport(t, db, " published ", "B1")
	draft := createReport(t, db, "Draft", "D1")
	gone := createReport(t, db, "Approved", "G1")
	db.Delete(&models.AuditResultReport{}, "id = ?", gone.ID)

	res, err := BackfillApprovedReports(db, now)
	if err != nil {
		t.Fatal(err)
	}
	if res.ApprovedReports != 2 || res.CreatedATRs != 3 {
		t.Fatalf("first run: %+v, want 2 reports / 3 ATRs", res)
	}
	if n := len(atrsOf(t, db, a.ID)) + len(atrsOf(t, db, b.ID)); n != 3 {
		t.Fatalf("%d ATRs for approved reports", n)
	}
	if len(atrsOf(t, db, draft.ID)) != 0 || len(atrsOf(t, db, gone.ID)) != 0 {
		t.Fatal("draft or deleted report got ATRs")
	}
	res, err = BackfillApprovedReports(db, now)
	if err != nil || res.CreatedATRs != 0 {
		t.Fatalf("second run: %+v err %v, want nothing created", res, err)
	}
}

// Through the generic handlers wired for /audit-result-reports (routes uses
// AfterReportWrite as the after-hook)
func TestApprovalThroughCrudHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newDB(t)
	newOne := func() interface{} { return &models.AuditResultReport{} }
	r := gin.New()
	r.GET("/arr", crud.List(db, "AuditResultReport", func() interface{} { return &[]models.AuditResultReport{} }))
	r.POST("/arr", crud.CreateWithHooks(db, "AuditResultReport", newOne, nil, AfterReportWrite))
	r.PUT("/arr/:id", crud.UpdateWithHooks(db, "AuditResultReport", newOne, nil, AfterReportWrite))

	do := func(method, path, body string) (int, string) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		return w.Code, w.Body.String()
	}

	id := uuid.New()
	if code, body := do(http.MethodPost, "/arr", `{"id":"`+id.String()+`","reportNumber":"LHA-100/SKAI/2026","status":"Draft","findings":[{"title":"Kas"},{"title":"Stok"}]}`); code >= 300 {
		t.Fatalf("POST: %d %s", code, body)
	}
	if n := len(atrsOf(t, db, id)); n != 0 {
		t.Fatalf("draft created %d ATRs", n)
	}

	// The list exposes the finding ids
	code, body := do(http.MethodGet, "/arr", "")
	if code != http.StatusOK {
		t.Fatalf("GET: %d %s", code, body)
	}
	var list struct {
		Data struct {
			Items []struct {
				Findings []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
				} `json:"findings"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &list); err != nil || len(list.Data.Items) != 1 || len(list.Data.Items[0].Findings) != 2 {
		t.Fatalf("list body: %s (%v)", body, err)
	}
	kasID := list.Data.Items[0].Findings[0].ID
	if _, err := uuid.Parse(kasID); err != nil {
		t.Fatalf("finding id not exposed: %s", body)
	}

	// Approve; the client drops the ids but keeps the titles: ids are kept
	if code, body := do(http.MethodPut, "/arr/"+id.String(), `{"status":"Final","findings":[{"title":"Kas"},{"title":"Stok"},{"title":"Vendor"}]}`); code >= 300 {
		t.Fatalf("PUT approve: %d %s", code, body)
	}
	atrs := atrsOf(t, db, id)
	if len(atrs) != 3 {
		t.Fatalf("approval created %d ATRs, want 3", len(atrs))
	}
	found := false
	for _, a := range atrs {
		if a.FindingID == kasID {
			found = true
		}
	}
	if !found {
		t.Fatalf("Kas lost its id on an update without ids")
	}

	// Saving again does not duplicate
	if code, body := do(http.MethodPut, "/arr/"+id.String(), `{"status":"Final","executive_summary":"x"}`); code >= 300 {
		t.Fatalf("PUT again: %d %s", code, body)
	}
	if n := len(atrsOf(t, db, id)); n != 3 {
		t.Fatalf("re-save: %d ATRs", n)
	}

	// A failing sync rolls the LHA write back
	if err := db.Exec("DROP TABLE action_taken_reports").Error; err != nil {
		t.Fatal(err)
	}
	if code, _ := do(http.MethodPut, "/arr/"+id.String(), `{"status":"Final","title":"must not stick"}`); code < 400 {
		t.Fatalf("PUT with broken ATR table answered %d", code)
	}
	var stored models.AuditResultReport
	db.First(&stored, "id = ?", id)
	if stored.Title == "must not stick" {
		t.Fatal("LHA update was not rolled back when the ATR sync failed")
	}
}

func TestAfterReportWriteIgnoresOtherEntities(t *testing.T) {
	if err := AfterReportWrite(nil, &models.ActionTakenReport{}); err != nil {
		t.Fatal(err)
	}
	if err := AfterReportWrite(nil, &models.AuditResultReport{}); err != nil {
		t.Fatal(err)
	}
	if errors.Is(nil, gorm.ErrRecordNotFound) {
		t.Fatal("unreachable")
	}
}
