package audit_execution

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	ctrlAssignmentLetter "audit-service/controllers/assignment_letter"
	"audit-service/controllers/crud"
	"audit-service/models"
	"audit-service/pkg/logger"
	"audit-service/pkg/testdb"
	"audit-service/services/audit_execution_sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// writeSyncRouter mirrors the wiring in routes/handler.go for the routes under test.
func writeSyncRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	newInterview := func() interface{} { return &models.FieldworkInterview{} }
	newControl := func() interface{} { return &models.FieldworkTestControl{} }
	newLetter := func() interface{} { return &models.AssignmentLetter{} }

	r := gin.New()
	r.POST("/fieldwork/interviews", crud.CreateWithHooks(db, "FieldworkInterview", newInterview, nil, SyncAfterCreate(LetterKeysOf)))
	r.PUT("/fieldwork/test-controls/:id", SyncOnWrite(db, newControl, LetterKeysOf, crud.Update(db, "FieldworkTestControl", newControl)))
	r.DELETE("/fieldwork/test-controls/:id", SyncOnWrite(db, newControl, LetterKeysOf, crud.Delete(db, "FieldworkTestControl", newControl)))
	r.POST("/assignment-letters", crud.CreateWithHooks(db, "AssignmentLetter", newLetter, nil, SyncLetterAfterCreate()))
	r.PUT("/assignment-letters/:id", SyncLetterOnWrite(db, crud.UpdateWithHook(db, "AssignmentLetter", newLetter, ctrlAssignmentLetter.NormalizePurposeOnUpdate)))
	r.DELETE("/assignment-letters/:id", SyncLetterOnWrite(db, crud.Delete(db, "AssignmentLetter", newLetter)))
	r.GET("/audit-executions", ListWithSources(db))
	r.POST("/audit-executions/sync", NewSyncController(db).SyncFromSources)
	return r
}

func writeSyncDB(t *testing.T) *gorm.DB {
	t.Helper()
	logger.Log = zap.NewNop()
	return testdb.New(t,
		&models.AuditExecution{},
		&models.AuditActivity{},
		&models.AssignmentLetter{},
		&models.FieldworkInterview{},
		&models.FieldworkObservation{},
		&models.FieldworkDocument{},
		&models.FieldworkSample{},
		&models.FieldworkTestControl{},
		&models.WorkingPaperHeader{},
		&models.WorkingPaperRisk{},
		&models.WorkingPaperSample{},
		&models.WorkingPaperCause{},
		&models.WorkingPaperPlan{},
		&models.AuditResultReport{},
	)
}

func doJSON(t *testing.T, r http.Handler, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(method, path, &buf))
	return w
}

func progressOf(t *testing.T, db *gorm.DB, ref string) int {
	t.Helper()
	var e models.AuditExecution
	if err := db.Where("ref = ?", ref).First(&e).Error; err != nil {
		t.Fatalf("execution %s: %v", ref, err)
	}
	return e.Progress
}

func TestPostFieldworkRecordSyncsItsLetterExecution(t *testing.T) {
	db := writeSyncDB(t)
	r := writeSyncRouter(db)
	const key = "ST-001/SKAI/2026"
	mustCreateRow(t, db, &models.AssignmentLetter{ID: uuid.New(), LetterNumber: key, Status: "Published"})
	mustCreateRow(t, db, &models.AuditExecution{ID: uuid.New(), Ref: key, Name: "Audit A", Progress: 0})

	if got := progressOf(t, db, key); got != 0 {
		t.Fatalf("before save progress = %d, want 0", got)
	}
	w := doJSON(t, r, http.MethodPost, "/fieldwork/interviews", map[string]string{
		"assignmentLetterId": key, "interviewee": "Kepala Kas",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	// Letter issued (entry band) plus fieldwork evidence: fieldwork band opens at 26.
	if got := progressOf(t, db, key); got != 26 {
		t.Fatalf("after save progress = %d, want 26", got)
	}
}

func TestPutTestControlSyncsOldAndNewLetter(t *testing.T) {
	db := writeSyncDB(t)
	r := writeSyncRouter(db)
	const letterA, letterB = "ST-001/SKAI/2026", "ST-002/SKAI/2026"
	for _, key := range []string{letterA, letterB} {
		mustCreateRow(t, db, &models.AssignmentLetter{ID: uuid.New(), LetterNumber: key, Status: "Published"})
		mustCreateRow(t, db, &models.AuditExecution{ID: uuid.New(), Ref: key, Name: key})
	}
	control := &models.FieldworkTestControl{ID: uuid.New(), AssignmentLetterID: letterA, ControlName: "C1"}
	mustCreateRow(t, db, control)

	// Tested control under A: fieldwork band, 26 + 24 = 50.
	w := doJSON(t, r, http.MethodPut, "/fieldwork/test-controls/"+control.ID.String(), map[string]string{"testResult": "Effective"})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if got := progressOf(t, db, letterA); got != 50 {
		t.Fatalf("A after test = %d, want 50", got)
	}
	if got := progressOf(t, db, letterB); got != 0 {
		t.Fatalf("B untouched = %d, want 0", got)
	}

	// Move the control to B: A loses its evidence (entry band, 1), B gains it (50).
	w = doJSON(t, r, http.MethodPut, "/fieldwork/test-controls/"+control.ID.String(), map[string]string{"assignmentLetterId": letterB})
	if w.Code != http.StatusOK {
		t.Fatalf("move status = %d, body = %s", w.Code, w.Body.String())
	}
	if got := progressOf(t, db, letterA); got != 1 {
		t.Fatalf("A after move = %d, want 1 (entry band)", got)
	}
	if got := progressOf(t, db, letterB); got != 50 {
		t.Fatalf("B after move = %d, want 50", got)
	}
}

func TestDeleteTestControlSyncsLetter(t *testing.T) {
	db := writeSyncDB(t)
	r := writeSyncRouter(db)
	const key = "ST-001/SKAI/2026"
	mustCreateRow(t, db, &models.AssignmentLetter{ID: uuid.New(), LetterNumber: key, Status: "Published"})
	mustCreateRow(t, db, &models.AuditExecution{ID: uuid.New(), Ref: key, Name: "Audit A"})
	control := &models.FieldworkTestControl{ID: uuid.New(), AssignmentLetterID: key, ControlName: "C1", TestResult: "Effective"}
	mustCreateRow(t, db, control)
	// Derive the pre-delete state explicitly: the rows above were written without a save.
	if _, err := audit_execution_sync.SyncExecutionsForLetter(context.Background(), db, key); err != nil {
		t.Fatal(err)
	}
	if got := progressOf(t, db, key); got != 50 {
		t.Fatalf("before delete = %d, want 50", got)
	}
	w := doJSON(t, r, http.MethodDelete, "/fieldwork/test-controls/"+control.ID.String(), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if got := progressOf(t, db, key); got != 1 {
		t.Fatalf("after delete = %d, want 1 (entry band)", got)
	}
}

// A failing sync is logged. The save still commits and returns 201.
func TestPostSaveSurvivesSyncFailure(t *testing.T) {
	db := writeSyncDB(t)
	r := writeSyncRouter(db)
	const key = "ST-001/SKAI/2026"
	mustCreateRow(t, db, &models.AssignmentLetter{ID: uuid.New(), LetterNumber: key, Status: "Published"})
	mustCreateRow(t, db, &models.AuditExecution{ID: uuid.New(), Ref: key, Name: "Audit A", Progress: 0})
	if err := db.Migrator().DropTable(&models.FieldworkTestControl{}); err != nil {
		t.Fatal(err)
	}

	w := doJSON(t, r, http.MethodPost, "/fieldwork/interviews", map[string]string{
		"assignmentLetterId": key, "interviewee": "Kepala Kas",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 even though the sync fails; body = %s", w.Code, w.Body.String())
	}
	var n int64
	if err := db.Model(&models.FieldworkInterview{}).Where("assignment_letter_id = ?", key).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("interview rows = %d, want 1: the save must commit", n)
	}
	if got := progressOf(t, db, key); got != 0 {
		t.Fatalf("progress = %d, want 0 (sync rolled back)", got)
	}
}

func TestPostLetterSyncsExecutionRefingItsNumber(t *testing.T) {
	db := writeSyncDB(t)
	r := writeSyncRouter(db)
	const key = "ST-010/SKAI/2026"
	mustCreateRow(t, db, &models.AuditExecution{ID: uuid.New(), Ref: key, Name: "Audit Baru"})

	w := doJSON(t, r, http.MethodPost, "/assignment-letters", map[string]string{"id": uuid.NewString(), "letterNumber": key, "status": "Published"})
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if got := progressOf(t, db, key); got != 1 {
		t.Fatalf("progress after letter = %d, want 1 (entry band)", got)
	}
}

// The list carries sources, and progress is derived from them. The stored row is
// not written by the list.
func TestListWithSourcesReturnsSourcesAndDerivedProgress(t *testing.T) {
	db := writeSyncDB(t)
	r := writeSyncRouter(db)
	const key = "ST-001/SKAI/2026"
	mustCreateRow(t, db, &models.AssignmentLetter{ID: uuid.New(), LetterNumber: key, Status: "Published"})
	exec := &models.AuditExecution{ID: uuid.New(), Ref: key, Name: "Audit A", Progress: 90}
	mustCreateRow(t, db, exec)
	mustCreateRow(t, db, &models.FieldworkTestControl{ID: uuid.New(), AssignmentLetterID: key, ControlName: "C1", TestResult: "Effective", MitigationPlan: "Perbaiki SOP"})
	mustCreateRow(t, db, &models.FieldworkTestControl{ID: uuid.New(), AssignmentLetterID: key, ControlName: "C2"})

	w := doJSON(t, r, http.MethodGet, "/audit-executions", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var body struct {
		Data struct {
			Items []struct {
				Ref      string         `json:"ref"`
				Progress int            `json:"progress"`
				Status   string         `json:"status"`
				Sources  map[string]any `json:"sources"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v; body = %s", err, w.Body.String())
	}
	if len(body.Data.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(body.Data.Items))
	}
	item := body.Data.Items[0]
	// One tested control of two: fieldwork band 26 + 24*1/2 = 38. Stored 90 is not used.
	if item.Progress != 38 || item.Status != models.AuditStatusFieldwork {
		t.Fatalf("progress = %d status %q, want 38 fieldwork", item.Progress, item.Status)
	}
	want := map[string]float64{
		"test_controls":              2,
		"test_controls_tested":       1,
		"test_controls_with_finding": 0,
		"mitigation_plans":           1,
		"working_paper_plans":        0,
		"fieldwork_records":          0,
	}
	for k, v := range want {
		got, ok := item.Sources[k].(float64)
		if !ok || got != v {
			t.Errorf("sources.%s = %v, want %v", k, item.Sources[k], v)
		}
	}
	for _, k := range []string{"annual_plan", "assignment_letter", "working_paper_causes", "result_report", "result_report_approved"} {
		if _, ok := item.Sources[k]; !ok {
			t.Errorf("sources.%s missing", k)
		}
	}

	var stored models.AuditExecution
	if err := db.First(&stored, "id = ?", exec.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.Progress != 90 {
		t.Fatalf("list wrote the stored row: progress %d, want 90", stored.Progress)
	}
}

func TestLetterKeysOf(t *testing.T) {
	id := uuid.New()
	cases := []struct {
		name   string
		entity interface{}
		want   []string
	}{
		{"interview", &models.FieldworkInterview{AssignmentLetterID: "L1"}, []string{"L1"}},
		{"test control", &models.FieldworkTestControl{AssignmentLetterID: "L1"}, []string{"L1"}},
		{"WP header", &models.WorkingPaperHeader{AssignmentLetterID: "L1"}, []string{"L1"}},
		{"WP cause", &models.WorkingPaperCause{WorkingPaperID: "L1"}, []string{"L1"}},
		{"WP plan", &models.WorkingPaperPlan{WorkingPaperID: "L1"}, []string{"L1"}},
		{"LHA", &models.AuditResultReport{AssignmentLetterID: "L1"}, []string{"L1"}},
		{"letter keeps number and id", &models.AssignmentLetter{ID: id, LetterNumber: "L1"}, []string{"L1", id.String()}},
		{"letter without id", &models.AssignmentLetter{LetterNumber: "L1"}, []string{"L1"}},
		{"unknown type", &models.AuditExecution{}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := LetterKeysOf(tc.entity)
			if len(got) != len(tc.want) {
				t.Fatalf("keys = %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("keys = %v, want %v", got, tc.want)
				}
			}
		})
	}
}

func mustCreateRow(t *testing.T, db *gorm.DB, v interface{}) {
	t.Helper()
	if err := db.Create(v).Error; err != nil {
		t.Fatalf("create %T: %v", v, err)
	}
}
