package audit_execution

import (
	"encoding/json"
	"net/http"
	"testing"

	"audit-service/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type listedExecution struct {
	ID          string `json:"id"`
	Ref         string `json:"ref"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	LeadAuditor string `json:"lead_auditor"`
	Progress    int    `json:"progress"`
	Status      string `json:"status"`
}

func listExecutions(t *testing.T, r *gin.Engine) []listedExecution {
	t.Helper()
	w := doJSON(t, r, http.MethodGet, "/audit-executions", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", w.Code, w.Body.String())
	}
	var body struct {
		Data struct {
			Items []listedExecution `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v; body = %s", err, w.Body.String())
	}
	return body.Data.Items
}

type syncResponse struct {
	Data struct {
		Synced              int              `json:"synced"`
		Created             int              `json:"created"`
		Items               []any            `json:"items"`
		UnmatchedSourceRows map[string]int64 `json:"unmatched_source_rows"`
	} `json:"data"`
}

func postSync(t *testing.T, r *gin.Engine) syncResponse {
	t.Helper()
	w := doJSON(t, r, http.MethodPost, "/audit-executions/sync", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("sync status = %d, body = %s", w.Code, w.Body.String())
	}
	var body syncResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v; body = %s", err, w.Body.String())
	}
	return body
}

func letterBody(id uuid.UUID, number, title string) map[string]interface{} {
	return map[string]interface{}{
		"id":           id.String(),
		"letterNumber": number,
		"status":       "Draft",
		"auditTitle":   title,
		"category":     "Assurance",
		"membersList": []map[string]string{
			{"name": "Zeta Ramadhani", "role": "Chairperson"},
			{"name": "Rina Wulandari", "role": "Member"},
		},
	}
}

func countExecutionRows(t *testing.T, db *gorm.DB, unscoped bool) int64 {
	t.Helper()
	q := db.Model(&models.AuditExecution{})
	if unscoped {
		q = db.Unscoped().Model(&models.AuditExecution{})
	}
	var n int64
	if err := q.Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func TestLetterLifecycleDrivesExecution(t *testing.T) {
	db := writeSyncDB(t)
	r := writeSyncRouter(db)
	const number = "ST-001/SKAI/2026"
	id := uuid.New()

	// Create: a Draft letter gets an execution in the planning stage.
	w := doJSON(t, r, http.MethodPost, "/assignment-letters", letterBody(id, number, "Audit Kas"))
	if w.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", w.Code, w.Body.String())
	}
	items := listExecutions(t, r)
	if len(items) != 1 {
		t.Fatalf("executions after create = %d, want 1", len(items))
	}
	got := items[0]
	if got.Ref != number || got.Name != "Audit Kas" || got.Category != "Assurance" || got.LeadAuditor != "Zeta Ramadhani" {
		t.Fatalf("execution = %+v, want fields from the letter", got)
	}
	if got.Progress != 0 || got.Status != models.AuditStatusPlanning {
		t.Fatalf("progress %d status %q, want 0 planning", got.Progress, got.Status)
	}

	// Update: publishing and retitling the letter keeps one execution; the name
	// was not edited by a user, so it follows the letter.
	w = doJSON(t, r, http.MethodPut, "/assignment-letters/"+id.String(), map[string]interface{}{
		"status": "Published", "auditTitle": "Audit Kas Kecil", "scopeList": []string{"Kas kecil"},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("update status = %d, body = %s", w.Code, w.Body.String())
	}
	items = listExecutions(t, r)
	if len(items) != 1 {
		t.Fatalf("executions after update = %d, want 1", len(items))
	}
	if items[0].Name != "Audit Kas Kecil" {
		t.Errorf("name = %q, want it to follow the letter", items[0].Name)
	}
	// Published with scope and a team: entry band, 2 of 3 items present.
	if items[0].Status != models.AuditStatusEntryMeeting {
		t.Errorf("status = %q, want %q", items[0].Status, models.AuditStatusEntryMeeting)
	}
	var stored models.AuditExecution
	if err := db.Where("ref = ?", number).First(&stored).Error; err != nil {
		t.Fatal(err)
	}
	if stored.Status != models.AuditStatusEntryMeeting {
		t.Errorf("stored status = %q, the save-time sync must write it", stored.Status)
	}

	// Delete: the execution is soft-deleted and leaves the list.
	w = doJSON(t, r, http.MethodDelete, "/assignment-letters/"+id.String(), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("delete status = %d, body = %s", w.Code, w.Body.String())
	}
	if items = listExecutions(t, r); len(items) != 0 {
		t.Fatalf("executions after delete = %+v, want none", items)
	}
	if n := countExecutionRows(t, db, true); n != 1 {
		t.Fatalf("rows incl. deleted = %d, want 1 (soft delete)", n)
	}
	// The backfill does not bring it back.
	if res := postSync(t, r); res.Data.Created != 0 || res.Data.Synced != 0 {
		t.Fatalf("sync after delete = created %d synced %d, want 0/0", res.Data.Created, res.Data.Synced)
	}
}

func TestLetterRenumberKeepsOneExecution(t *testing.T) {
	db := writeSyncDB(t)
	r := writeSyncRouter(db)
	id := uuid.New()
	if w := doJSON(t, r, http.MethodPost, "/assignment-letters", letterBody(id, "ST-001/SKAI/2026", "Audit Kas")); w.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", w.Code, w.Body.String())
	}
	w := doJSON(t, r, http.MethodPut, "/assignment-letters/"+id.String(), map[string]interface{}{"letterNumber": "ST-050/SKAI/2026"})
	if w.Code != http.StatusOK {
		t.Fatalf("update status = %d, body = %s", w.Code, w.Body.String())
	}
	items := listExecutions(t, r)
	if len(items) != 1 || items[0].Ref != "ST-050/SKAI/2026" {
		t.Fatalf("executions = %+v, want one with the new number", items)
	}
	if n := countExecutionRows(t, db, true); n != 1 {
		t.Fatalf("rows incl. deleted = %d, want 1", n)
	}
}

// An execution name a user edited is kept when the letter's title changes.
func TestLetterUpdateKeepsUserEditedName(t *testing.T) {
	db := writeSyncDB(t)
	r := writeSyncRouter(db)
	id := uuid.New()
	if w := doJSON(t, r, http.MethodPost, "/assignment-letters", letterBody(id, "ST-001/SKAI/2026", "Audit Kas")); w.Code != http.StatusCreated {
		t.Fatalf("create status = %d, body = %s", w.Code, w.Body.String())
	}
	if err := db.Model(&models.AuditExecution{}).Where("ref = ?", "ST-001/SKAI/2026").Update("name", "Nama manual").Error; err != nil {
		t.Fatal(err)
	}
	if w := doJSON(t, r, http.MethodPut, "/assignment-letters/"+id.String(), map[string]interface{}{"auditTitle": "Judul baru"}); w.Code != http.StatusOK {
		t.Fatalf("update status = %d, body = %s", w.Code, w.Body.String())
	}
	if items := listExecutions(t, r); len(items) != 1 || items[0].Name != "Nama manual" {
		t.Fatalf("executions = %+v, want one keeping the manual name", items)
	}
}

// Letters that predate the hook (and a seeded demo execution) are backfilled by
// POST /audit-executions/sync; a second run creates nothing.
func TestSyncBackfillsLettersIdempotently(t *testing.T) {
	db := writeSyncDB(t)
	r := writeSyncRouter(db)
	// Seeded demo execution with the number of an existing letter.
	seeded := &models.AuditExecution{ID: uuid.New(), Ref: "ST-001/SKAI/2026", Name: "Audit Operasional Keuangan", LeadAuditor: "Zeta Ramadhani"}
	mustCreateRow(t, db, seeded)
	for _, n := range []string{"ST-001/SKAI/2026", "ST-002/SKAI/2026", "ST-003/SKAI/2026"} {
		mustCreateRow(t, db, &models.AssignmentLetter{ID: uuid.New(), LetterNumber: n, Status: "Draft", AuditTitle: "Audit " + n})
	}

	first := postSync(t, r)
	if first.Data.Created != 2 || first.Data.Synced != 3 || len(first.Data.Items) != 3 {
		t.Fatalf("first sync = created %d synced %d items %d, want 2/3/3", first.Data.Created, first.Data.Synced, len(first.Data.Items))
	}
	if first.Data.UnmatchedSourceRows == nil {
		t.Errorf("unmatched_source_rows missing from the response")
	}
	second := postSync(t, r)
	if second.Data.Created != 0 || second.Data.Synced != 3 {
		t.Fatalf("second sync = created %d synced %d, want 0/3", second.Data.Created, second.Data.Synced)
	}
	if n := countExecutionRows(t, db, true); n != 3 {
		t.Fatalf("execution rows = %d, want 3", n)
	}
	var kept models.AuditExecution
	if err := db.First(&kept, "id = ?", seeded.ID).Error; err != nil {
		t.Fatal(err)
	}
	if kept.Name != "Audit Operasional Keuangan" {
		t.Fatalf("seeded name = %q, want it kept", kept.Name)
	}
}
