package executivesummary

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"audit-service/models"
	"audit-service/pkg/testdb"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func fuDB(t *testing.T) *gorm.DB {
	t.Helper()
	return testdb.New(t,
		&models.AssignmentLetter{},
		&models.AuditResultReport{},
		&models.FieldworkTestControl{},
		&models.WorkingPaperRisk{},
		&models.ActionTakenReport{},
		&models.ActionTakenReportEvidence{},
		&models.ExecutiveSummary{},
	)
}

// fuFixture is one approved LHA with a finding per case, built relative to base
type fuFixture struct {
	lha, draft                                  *models.AuditResultReport
	fClosed, fOverdue, fNoPlan14, fNoATR        string
	fCancelled, fPending, fDeleted, fOrphan     string
	atrClosed, atrOverdue, atrNoPlan, atrOrphan *models.ActionTakenReport
	atrPending                                  *models.ActionTakenReport
}

const fuReportNumber = "LHA-001/SKAI/2026"

func dayOf(t time.Time) *time.Time {
	d := models.ATRToday(t)
	return &d
}

func newFUFixture(t *testing.T, db *gorm.DB, base time.Time) fuFixture {
	t.Helper()
	fx := fuFixture{
		fClosed: uuid.NewString(), fOverdue: uuid.NewString(), fNoPlan14: uuid.NewString(), fNoATR: uuid.NewString(),
		fCancelled: uuid.NewString(), fPending: uuid.NewString(), fDeleted: uuid.NewString(), fOrphan: uuid.NewString(),
	}
	q4 := time.Date(base.Year(), 10, 5, 0, 0, 0, 0, time.UTC)
	fx.lha = &models.AuditResultReport{
		ID: uuid.New(), ReportNumber: fuReportNumber, Department: "Keuangan", Status: "Final", ReportDate: &q4,
		Findings: []models.AuditReportFinding{
			{ID: fx.fClosed, Title: "Closed finding"},
			{ID: fx.fOverdue, Title: "Overdue finding"},
			{ID: fx.fNoPlan14, Title: "No plan finding"},
			{ID: fx.fNoATR, Title: "Finding without ATR"},
			{ID: fx.fCancelled, Title: "Cancelled finding"},
			{ID: fx.fPending, Title: "Pending review finding"},
			{ID: fx.fDeleted, Title: "Deleted ATR finding"},
		},
	}
	mustCreate(t, db, fx.lha)

	created := base.AddDate(0, 0, -40) // LHA approved 40 days before base
	atr := func(finding, status string, due *time.Time, pic string) *models.ActionTakenReport {
		id := fx.lha.ID
		a := &models.ActionTakenReport{
			ID: uuid.New(), AuditResultReportID: &id, FindingID: finding, ReportNumber: fuReportNumber,
			FindingTitle: "copy", Status: status, DueDate: due, PICUserID: pic, PICName: pic, Progress: 40,
			CreatedAt: created, UpdatedAt: created,
		}
		mustCreate(t, db, a)
		return a
	}
	fx.atrClosed = atr(fx.fClosed, models.ATRStatusCompleted, dayOf(base.AddDate(0, 0, -100)), "u1")
	fx.atrOverdue = atr(fx.fOverdue, models.ATRStatusInProgress, dayOf(base.AddDate(0, 0, -13)), "u1")
	fx.atrNoPlan = atr(fx.fNoPlan14, models.ATRStatusPlanned, dayOf(base.AddDate(0, 0, 60)), "u2")
	atr(fx.fCancelled, models.ATRStatusCancelled, dayOf(base.AddDate(0, 0, -5)), "u1")
	fx.atrPending = atr(fx.fPending, models.ATRStatusPendingReview, nil, "u3")
	deleted := atr(fx.fDeleted, models.ATRStatusInProgress, nil, "u1")
	if err := db.Delete(deleted).Error; err != nil {
		t.Fatal(err)
	}
	// Finding removed from the LHA after work started: still followed up
	fx.atrOrphan = atr(fx.fOrphan, models.ATRStatusInProgress, nil, "u4")

	mustCreate(t, db, &models.ActionTakenReportEvidence{
		ID: uuid.New(), ActionTakenReportID: fx.atrPending.ID, FileName: "bukti.pdf", StoredName: "x-bukti.pdf", UploadedAt: base,
	})

	fx.draft = &models.AuditResultReport{
		ID: uuid.New(), ReportNumber: "LHA-002/SKAI/2026", Status: "Draft", ReportDate: &q4,
		Findings: []models.AuditReportFinding{{ID: uuid.NewString(), Title: "Draft finding"}},
	}
	mustCreate(t, db, fx.draft)
	return fx
}

func TestLoadFollowUp(t *testing.T) {
	db := fuDB(t)
	fx := newFUFixture(t, db, testNow)

	lhas, err := approvedLHAsByID(db, fx.lha.ID)
	if err != nil {
		t.Fatal(err)
	}
	s, err := LoadFollowUp(db, lhas, testNow, true)
	if err != nil {
		t.Fatal(err)
	}
	want := FollowUpBuckets{Closed: 1, InProgress: 2, Overdue: 1, NoAction: 2}
	if s.Buckets != want || s.Total != 6 || s.Cancelled != 1 {
		t.Fatalf("buckets = %+v total %d cancelled %d, want %+v total 6 cancelled 1", s.Buckets, s.Total, s.Cancelled, want)
	}
	if len(s.Findings) != 6 {
		t.Fatalf("findings = %d, want 6", len(s.Findings))
	}

	byID := map[string]FollowUpFinding{}
	for _, f := range s.Findings {
		byID[f.FindingID] = f
	}
	for _, gone := range []string{fx.fCancelled, fx.fDeleted} {
		if _, ok := byID[gone]; ok {
			t.Errorf("finding %s must be excluded", gone)
		}
	}
	check := func(id, bucket string, reason *string) FollowUpFinding {
		t.Helper()
		f, ok := byID[id]
		if !ok {
			t.Fatalf("finding %s missing", id)
		}
		if f.Bucket != bucket {
			t.Errorf("%s: bucket %s, want %s", f.Title, f.Bucket, bucket)
		}
		if (reason == nil) != (f.NoActionReason == nil) || (reason != nil && *reason != *f.NoActionReason) {
			t.Errorf("%s: noActionReason %v, want %v", f.Title, f.NoActionReason, reason)
		}
		return f
	}
	str := func(s string) *string { return &s }

	check(fx.fClosed, models.FollowUpBucketClosed, nil)
	if f := check(fx.fOverdue, models.FollowUpBucketOverdue, nil); f.OverdueDays != 13 || f.DueDate == nil || *f.DueDate != "2026-10-01" {
		t.Errorf("overdue: days %d due %v", f.OverdueDays, f.DueDate)
	}
	check(fx.fNoPlan14, models.FollowUpBucketNoAction, str(models.NoActionReasonNoPlanEscalated))
	if f := check(fx.fNoATR, models.FollowUpBucketNoAction, str(models.NoActionReasonNoPIC)); f.ATRID != nil || f.ATRStatus != nil || f.DueDate != nil || f.Progress != nil || f.PICName != nil || f.Evidence == nil {
		t.Errorf("no ATR row: %+v", f)
	}
	if f := check(fx.fPending, models.FollowUpBucketInProgress, nil); len(f.Evidence) != 1 || f.Evidence[0].FileName != "bukti.pdf" || *f.ATRStatus != models.ATRStatusPendingReview {
		t.Errorf("pending: %+v", f)
	}
	if f := check(fx.fOrphan, models.FollowUpBucketInProgress, nil); f.Title != "copy" || f.Divisi != "Keuangan" || f.ReportNumber != fuReportNumber {
		t.Errorf("orphan: %+v", f)
	}
	if f := byID[fx.fClosed]; f.Title != "Closed finding" || f.PICName == nil || *f.PICName != "u1" || f.Progress == nil || *f.Progress != 40 || *f.ATRID != fx.atrClosed.ID.String() {
		t.Errorf("closed fields: %+v", f)
	}

	// The 14 days count from the LHA approval (first ATR created): 10 days later is NO_PLAN
	early := testNow.AddDate(0, 0, -30)
	s, err = LoadFollowUp(db, lhas, early, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range s.Findings {
		if f.FindingID == fx.fNoPlan14 && (f.NoActionReason == nil || *f.NoActionReason != models.NoActionReasonNoPlan) {
			t.Errorf("10 days after approval: reason %v, want NO_PLAN", f.NoActionReason)
		}
	}

	// A draft LHA has no follow-up
	drafts, err := approvedLHAsByID(db, fx.draft.ID)
	if err != nil || len(drafts) != 0 {
		t.Fatalf("draft LHA: %v %v", drafts, err)
	}
}

// The compilation statistics use the same buckets
func TestCompilationStatsUsesFollowUpBuckets(t *testing.T) {
	db := fuDB(t)
	newFUFixture(t, db, testNow)
	s, err := LoadCompilationStats(db, 2026, 4, testNow)
	if err != nil {
		t.Fatal(err)
	}
	if s.FollowUpBuckets != (FollowUpBuckets{Closed: 1, InProgress: 2, Overdue: 1, NoAction: 2}) || s.FollowUpCancelled != 1 {
		t.Fatalf("buckets %+v cancelled %d", s.FollowUpBuckets, s.FollowUpCancelled)
	}
	if len(s.FollowUpTable) != 3 || s.FollowUpTable[0].Jumlah != 1 || s.FollowUpTable[1].Jumlah != 2 || s.FollowUpTable[2].Jumlah != 1 {
		t.Fatalf("followUpTable = %+v", s.FollowUpTable)
	}
}

func fuRouter(db *gorm.DB, roles []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "u-test")
		c.Set("username", "tester")
		c.Set("roles", roles)
		c.Next()
	})
	g := r.Group("/api/v1/executive-summaries")
	g.GET("/follow-up-status", FollowUpStatusHandler(db))
	g.POST("/:id/approve", ApproveHandler(db))
	return r
}

type fuEnvelope struct {
	Success bool `json:"success"`
	Data    struct {
		AsOf               string                       `json:"asOf"`
		Buckets            map[string]int64             `json:"buckets"`
		Total              int64                        `json:"total"`
		Findings           []map[string]json.RawMessage `json:"findings"`
		FindingsRestricted bool                         `json:"findingsRestricted"`
	} `json:"data"`
}

func getFU(t *testing.T, r *gin.Engine, query string) (*httptest.ResponseRecorder, fuEnvelope) {
	t.Helper()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/executive-summaries/follow-up-status"+query, nil))
	var env fuEnvelope
	if w.Code == http.StatusOK {
		if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
			t.Fatalf("decode: %v", err)
		}
	}
	return w, env
}

func TestFollowUpStatusHandler(t *testing.T) {
	db := fuDB(t)
	fx := newFUFixture(t, db, time.Now())
	wantBuckets := map[string]int64{"closed": 1, "inProgress": 2, "overdue": 1, "noAction": 2}

	auditor := fuRouter(db, []string{"auditor"})
	for _, q := range []string{
		"?audit_result_report_id=" + fx.lha.ID.String(),
		"?report_number=" + "LHA-001%2FSKAI%2F2026",
		"?year=" + time.Now().Format("2006") + "&quarter=4",
	} {
		w, env := getFU(t, auditor, q)
		if w.Code != http.StatusOK {
			t.Fatalf("%s: status %d %s", q, w.Code, w.Body.String())
		}
		if _, err := time.Parse(time.RFC3339, env.Data.AsOf); err != nil {
			t.Errorf("%s: asOf %q not RFC3339", q, env.Data.AsOf)
		}
		for k, v := range wantBuckets {
			if env.Data.Buckets[k] != v {
				t.Errorf("%s: buckets = %v", q, env.Data.Buckets)
			}
		}
		if env.Data.Total != 6 || len(env.Data.Findings) != 6 || env.Data.FindingsRestricted {
			t.Errorf("%s: total %d findings %d restricted %v", q, env.Data.Total, len(env.Data.Findings), env.Data.FindingsRestricted)
		}
	}

	// Contract: every finding carries every key, null where there is no ATR
	_, env := getFU(t, auditor, "?audit_result_report_id="+fx.lha.ID.String())
	keys := []string{"findingId", "auditResultReportId", "reportNumber", "atrId", "title", "divisi", "atrStatus", "bucket", "noActionReason", "dueDate", "progress", "picName", "overdueDays", "evidence"}
	for _, f := range env.Data.Findings {
		for _, k := range keys {
			if _, ok := f[k]; !ok {
				t.Errorf("finding missing key %s: %v", k, f)
			}
		}
		var id string
		_ = json.Unmarshal(f["findingId"], &id)
		if id == fx.fNoATR {
			for _, k := range []string{"atrId", "atrStatus", "dueDate", "progress", "picName"} {
				if string(f[k]) != "null" {
					t.Errorf("no-ATR finding: %s = %s, want null", k, f[k])
				}
			}
			if string(f["noActionReason"]) != `"NO_PIC"` || string(f["evidence"]) != "[]" {
				t.Errorf("no-ATR finding: %v", f)
			}
		}
	}

	// Without ATR view_all: counts only
	for _, roles := range [][]string{{"VIEWER"}, {"AUDITEE"}, {"DEPARTMENT_HEAD"}, nil} {
		w, env := getFU(t, fuRouter(db, roles), "?audit_result_report_id="+fx.lha.ID.String())
		if w.Code != http.StatusOK || env.Data.Total != 6 || !env.Data.FindingsRestricted || env.Data.Findings == nil || len(env.Data.Findings) != 0 {
			t.Errorf("roles %v: status %d total %d restricted %v findings %v", roles, w.Code, env.Data.Total, env.Data.FindingsRestricted, env.Data.Findings)
		}
	}
	// EXECUTIVE (CAE) has view_all
	if _, env := getFU(t, fuRouter(db, []string{"EXECUTIVE"}), "?audit_result_report_id="+fx.lha.ID.String()); len(env.Data.Findings) != 6 {
		t.Errorf("EXECUTIVE: findings %d", len(env.Data.Findings))
	}

	// A draft LHA: zero counts
	if w, env := getFU(t, auditor, "?audit_result_report_id="+fx.draft.ID.String()); w.Code != http.StatusOK || env.Data.Total != 0 || len(env.Data.Findings) != 0 {
		t.Errorf("draft: %d %s", w.Code, w.Body.String())
	}

	for q, code := range map[string]int{
		"":                             http.StatusBadRequest,
		"?audit_result_report_id=nope": http.StatusBadRequest,
		"?audit_result_report_id=" + uuid.NewString(): http.StatusNotFound,
		"?report_number=LHA-404":                      http.StatusNotFound,
		"?year=abc":                                   http.StatusBadRequest,
		"?year=2026&quarter=5":                        http.StatusBadRequest,
	} {
		if w, _ := getFU(t, auditor, q); w.Code != code {
			t.Errorf("query %q: status %d, want %d", q, w.Code, code)
		}
	}
}

func approveFU(t *testing.T, db *gorm.DB, e *models.ExecutiveSummary) models.ExecutiveSummary {
	t.Helper()
	mustCreate(t, db, e)
	w := doAppr(t, fuRouter(db, []string{"ADMIN"}), http.MethodPost, "/api/v1/executive-summaries/"+e.ID.String()+"/approve", approvePayload(testPNGDataURL))
	if w.Code != http.StatusOK {
		t.Fatalf("approve: %d %s", w.Code, w.Body.String())
	}
	return reloadSummary(t, db, e.ID)
}

func decodeRows(t *testing.T, s string) []map[string]interface{} {
	t.Helper()
	var rows []map[string]interface{}
	if err := json.Unmarshal([]byte(s), &rows); err != nil {
		t.Fatalf("decode %q: %v", s, err)
	}
	return rows
}

func TestApproveFreezesFollowUp(t *testing.T) {
	db := fuDB(t)
	fx := newFUFixture(t, db, time.Now())

	topFindings, _ := json.Marshal([]map[string]interface{}{
		{"findingId": fx.fNoPlan14, "judulTemuan": "No plan finding", "statusTL": "In Progress"},
		{"judulTemuan": "Old row without id", "statusTL": "In Progress"},
	})
	matriks, _ := json.Marshal([]map[string]interface{}{
		{"findingId": fx.fOverdue, "judulTemuan": "Overdue finding", "status": "In Progress", "progres": 40},
		{"findingId": fx.fCancelled, "judulTemuan": "Cancelled finding", "status": "In Progress"},
		{"judulTemuan": "Old row without id", "status": "Closed"},
	})
	e := approveFU(t, db, &models.ExecutiveSummary{
		ID: uuid.New(), Status: StatusDraft, NomorDokumen: fuReportNumber, Tahun: 2026,
		FollowUpTable:    `[{"status":"Closed","jumlah":9,"persentase":100,"keterangan":"catatan closed"},{"status":"In Progress","jumlah":0,"persentase":0,"keterangan":""},{"status":"Overdue","jumlah":0,"persentase":0,"keterangan":"telat"}]`,
		TopFindings:      string(topFindings),
		MatriksKompilasi: string(matriks),
	})
	if e.ApprovedAt == nil {
		t.Fatal("approved_at not set")
	}

	table := decodeRows(t, e.FollowUpTable)
	if len(table) != 3 {
		t.Fatalf("followUpTable must keep 3 rows: %s", e.FollowUpTable)
	}
	want := []struct {
		status string
		jumlah float64
		pct    float64
		note   string
	}{{"Closed", 1, 25, "catatan closed"}, {"In Progress", 2, 50, ""}, {"Overdue", 1, 25, "telat"}}
	for i, w := range want {
		r := table[i]
		if r["status"] != w.status || r["jumlah"] != w.jumlah || r["persentase"] != w.pct || r["keterangan"] != w.note {
			t.Errorf("row %d = %v, want %+v", i, r, w)
		}
	}

	top := decodeRows(t, e.TopFindings)
	if top[0]["statusTL"] != FollowUpNoAction || top[1]["statusTL"] != "In Progress" {
		t.Errorf("topFindings = %v", top)
	}
	mk := decodeRows(t, e.MatriksKompilasi)
	if mk[0]["status"] != FollowUpOverdue || mk[0]["progres"] != float64(40) || mk[1]["status"] != "In Progress" || mk[2]["status"] != "Closed" {
		t.Errorf("matriksKompilasi = %v", mk)
	}
}

func TestApproveFreezesFourRowsOnlyWhenPresent(t *testing.T) {
	db := fuDB(t)
	newFUFixture(t, db, time.Now())

	e := approveFU(t, db, &models.ExecutiveSummary{
		ID: uuid.New(), Status: StatusDraft, NomorDokumen: "ESC-1/" + time.Now().Format("2006"), Tahun: time.Now().Year(), Quarter: 4,
		FollowUpTable: `[{"status":"Closed","jumlah":0},{"status":"In Progress","jumlah":0},{"status":"Overdue","jumlah":0},{"status":"No Action","jumlah":0,"keterangan":"belum ada rencana"}]`,
	})
	table := decodeRows(t, e.FollowUpTable)
	if len(table) != 4 || table[3]["status"] != FollowUpNoAction || table[3]["jumlah"] != float64(2) || table[3]["keterangan"] != "belum ada rencana" {
		t.Fatalf("followUpTable = %s", e.FollowUpTable)
	}
	// Percentages over all four rows (6 findings)
	if pct := table[1]["persentase"].(float64); pct < 33.3 || pct > 33.4 {
		t.Errorf("In Progress persentase = %v", pct)
	}
}

func TestApproveLeavesFollowUpWhenScopeUnknown(t *testing.T) {
	db := fuDB(t)
	newFUFixture(t, db, time.Now())

	const typed = `[{"status":"Closed","jumlah":7,"persentase":100,"keterangan":"manual"}]`
	e := approveFU(t, db, &models.ExecutiveSummary{
		ID: uuid.New(), Status: StatusDraft, NomorDokumen: "LHA-404/SKAI/2026", Tahun: 2026, FollowUpTable: typed,
	})
	if e.FollowUpTable != typed || e.Status != StatusApproved {
		t.Errorf("followUpTable = %s status %s", e.FollowUpTable, e.Status)
	}
}
