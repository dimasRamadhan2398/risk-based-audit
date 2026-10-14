package audit_execution_sync

import (
	"context"
	"sync"
	"testing"

	"audit-service/models"
	"audit-service/pkg/logger"
	"audit-service/pkg/testdb"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestMain(m *testing.M) {
	logger.Log = zap.NewNop()
	m.Run()
}

const letterKey = "ST-001/SKAI/2026"

func syncTestDB(t *testing.T) *gorm.DB {
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

func mustCreate(t *testing.T, db *gorm.DB, v interface{}) {
	t.Helper()
	if err := db.Create(v).Error; err != nil {
		t.Fatalf("create %T: %v", v, err)
	}
}

func seedExecution(t *testing.T, db *gorm.DB, progress int) *models.AuditExecution {
	t.Helper()
	exec := &models.AuditExecution{
		ID:       uuid.New(),
		Ref:      letterKey,
		Name:     "Audit Operasional Keuangan",
		Progress: progress,
		Status:   models.AuditStatusPlanning,
	}
	mustCreate(t, db, exec)
	return exec
}

func seedActivity(t *testing.T, db *gorm.DB, title string) *models.AuditActivity {
	t.Helper()
	act := &models.AuditActivity{
		ID:           uuid.New(),
		AnnualPlanID: uuid.New(),
		TargetUnitID: uuid.New(),
		ProjectCode:  "ASR-2026-001",
		Title:        title,
		Status:       "PLANNED",
	}
	mustCreate(t, db, act)
	return act
}

func seedLetter(t *testing.T, db *gorm.DB, number string, id uuid.UUID, status string) {
	t.Helper()
	mustCreate(t, db, &models.AssignmentLetter{ID: id, LetterNumber: number, Status: status})
}

func reload(t *testing.T, db *gorm.DB, exec *models.AuditExecution) *models.AuditExecution {
	t.Helper()
	var out models.AuditExecution
	if err := db.First(&out, "id = ?", exec.ID).Error; err != nil {
		t.Fatal(err)
	}
	return &out
}

// seedFullEvidence adds, under key, one interview, four test controls (two
// tested, one with mitigation), one cause and one complete plan. It stops
// before the LHA so each test chooses the report status.
func seedFullEvidence(t *testing.T, db *gorm.DB, key string) {
	t.Helper()
	mustCreate(t, db, &models.FieldworkInterview{ID: uuid.New(), AssignmentLetterID: key, Interviewee: "Kepala Kas"})
	for i, result := range []string{"Effective", "Effective", "", ""} {
		c := &models.FieldworkTestControl{
			ID: uuid.New(), AssignmentLetterID: key, ControlName: "C" + string(rune('A'+i)), TestResult: result,
		}
		if i == 0 {
			c.MitigationPlan = "Perbaiki SOP rekonsiliasi kas"
		}
		mustCreate(t, db, c)
	}
	mustCreate(t, db, &models.WorkingPaperCause{ID: uuid.New(), WorkingPaperID: key, Condition: "Selisih kas"})
	mustCreate(t, db, &models.WorkingPaperPlan{
		ID: uuid.New(), WorkingPaperID: key,
		ActionDescription: "Rekonsiliasi kas harian", Recommendation: "Lakukan rekonsiliasi harian",
	})
}

func TestSyncExecution_NoSourcesResetsManualProgress(t *testing.T) {
	db := syncTestDB(t)
	exec := seedExecution(t, db, 80) // hand-typed progress with no source data behind it

	res, err := SyncExecution(context.Background(), db, exec)
	if err != nil {
		t.Fatal(err)
	}
	if res.Progress != ProgressPlanned {
		t.Fatalf("progress = %d, want %d", res.Progress, ProgressPlanned)
	}
	got := reload(t, db, exec)
	if got.Progress != ProgressPlanned || got.Status != models.AuditStatusPlanning || got.StatusDetail == "" {
		t.Fatalf("stored = progress %d status %q detail %q", got.Progress, got.Status, got.StatusDetail)
	}
}

func TestSyncExecution_FollowsEachSourceModule(t *testing.T) {
	db := syncTestDB(t)
	ctx := context.Background()
	act := seedActivity(t, db, "Audit Operasional Keuangan")
	exec := seedExecution(t, db, 0)

	// Annual plan item is linked by title, so the execution is planned.
	res, err := SyncExecution(ctx, db, exec)
	if err != nil {
		t.Fatal(err)
	}
	if res.Progress != ProgressPlanned || !res.Sources.AnnualPlanLinked {
		t.Fatalf("annual plan only: progress %d, sources %+v", res.Progress, res.Sources)
	}
	if got := reload(t, db, exec); got.ActivityID == nil || *got.ActivityID != act.ID {
		t.Fatalf("activity link not saved: %v", got.ActivityID)
	}

	// Assignment letter issued (not draft), no scope recorded -> bottom of entry band.
	seedLetter(t, db, letterKey, uuid.New(), "Published")
	res, _ = SyncExecution(ctx, db, exec)
	if res.Progress != ProgressEntryMin || !res.Sources.AssignmentLetterIssued {
		t.Fatalf("letter issued: progress %d, sources %+v", res.Progress, res.Sources)
	}

	// Fieldwork evidence and 2 of 4 test controls tested -> fieldwork, 26 + 24*2/4 = 38.
	mustCreate(t, db, &models.FieldworkInterview{ID: uuid.New(), AssignmentLetterID: letterKey, Interviewee: "Kepala Kas"})
	for i, result := range []string{"Effective", "Effective", "", ""} {
		mustCreate(t, db, &models.FieldworkTestControl{
			ID: uuid.New(), AssignmentLetterID: letterKey, ControlName: "C" + string(rune('A'+i)), TestResult: result,
		})
	}
	res, _ = SyncExecution(ctx, db, exec)
	if res.Progress != 38 || res.Sources.TestControlsTested != 2 || res.Sources.TestControls != 4 {
		t.Fatalf("fieldwork: progress %d, sources %+v", res.Progress, res.Sources)
	}
	if got := reload(t, db, exec); got.Status != models.AuditStatusFieldwork {
		t.Fatalf("status = %q, want fieldwork", got.Status)
	}

	// Working paper cause without plans or mitigation -> still fieldwork, not draft findings.
	mustCreate(t, db, &models.WorkingPaperCause{ID: uuid.New(), WorkingPaperID: letterKey, Condition: "Selisih kas"})
	res, _ = SyncExecution(ctx, db, exec)
	if res.Progress != 38 || res.Sources.WorkingPaperCauses != 1 || res.Sources.WorkingPaperPlansComplete != 0 {
		t.Fatalf("cause without plan: progress %d, sources %+v", res.Progress, res.Sources)
	}

	// Complete plan and a mitigation plan -> draft findings, 55 + 20*1/1 = 75.
	if err := db.Model(&models.FieldworkTestControl{}).
		Where(colAssignmentLetterID+" = ? AND test_result = ?", letterKey, "Effective").
		Update(colMitigationPlan, "Perbaiki SOP rekonsiliasi kas").Error; err != nil {
		t.Fatal(err)
	}
	mustCreate(t, db, &models.WorkingPaperPlan{
		ID: uuid.New(), WorkingPaperID: letterKey,
		ActionDescription: "Rekonsiliasi kas harian", Recommendation: "Lakukan rekonsiliasi harian",
	})
	res, _ = SyncExecution(ctx, db, exec)
	if res.Progress != 75 || res.Sources.WorkingPaperPlansComplete != 1 || res.Sources.TestControlsWithMitigation != 2 {
		t.Fatalf("draft findings: progress %d, sources %+v", res.Progress, res.Sources)
	}

	// Draft LHA -> reporting draft, 76.
	lha := &models.AuditResultReport{ID: uuid.New(), AssignmentLetterID: letterKey, Status: "DRAFT"}
	mustCreate(t, db, lha)
	res, _ = SyncExecution(ctx, db, exec)
	if res.Progress != ProgressLHADraft || !res.Sources.LHAExists || res.Sources.LHAApproved {
		t.Fatalf("draft LHA: progress %d, sources %+v", res.Progress, res.Sources)
	}

	// LHA approved -> completed, and the annual plan activity follows.
	if err := db.Model(lha).Update("status", "Final").Error; err != nil {
		t.Fatal(err)
	}
	res, _ = SyncExecution(ctx, db, exec)
	if res.Progress != ProgressCompleted || res.Status != models.AuditStatusCompleted {
		t.Fatalf("approved LHA: progress %d status %q", res.Progress, res.Status)
	}
	var gotAct models.AuditActivity
	if err := db.First(&gotAct, "id = ?", act.ID).Error; err != nil {
		t.Fatal(err)
	}
	if gotAct.Status != "COMPLETED" {
		t.Fatalf("activity status = %q, want COMPLETED", gotAct.Status)
	}
}

// A draft letter gates everything else to 0, even with an approved LHA and a
// complete working paper behind it. Real fieldwork evidence is the only exception.
func TestSyncExecution_DraftLetterGatesProgress(t *testing.T) {
	db := syncTestDB(t)
	ctx := context.Background()
	exec := seedExecution(t, db, 0)
	seedLetter(t, db, letterKey, uuid.New(), "Draft")
	mustCreate(t, db, &models.AuditResultReport{ID: uuid.New(), AssignmentLetterID: letterKey, Status: "FINAL"})
	mustCreate(t, db, &models.WorkingPaperCause{ID: uuid.New(), WorkingPaperID: letterKey, Condition: "Selisih kas"})
	mustCreate(t, db, &models.WorkingPaperPlan{
		ID: uuid.New(), WorkingPaperID: letterKey,
		ActionDescription: "Tindakan", Recommendation: "Rekomendasi",
	})
	mustCreate(t, db, &models.FieldworkTestControl{
		ID: uuid.New(), AssignmentLetterID: letterKey, ControlName: "C", TestResult: "", MitigationPlan: "Mitigasi",
	})

	res, err := SyncExecution(ctx, db, exec)
	if err != nil {
		t.Fatal(err)
	}
	if res.Progress != ProgressPlanned || res.Status != models.AuditStatusPlanning {
		t.Fatalf("draft letter with later data: progress %d status %q, want 0 planning", res.Progress, res.Status)
	}

	// Letter published -> the LHA counts.
	if err := db.Model(&models.AssignmentLetter{}).Where("letter_number = ?", letterKey).
		Update("status", "Published").Error; err != nil {
		t.Fatal(err)
	}
	res, _ = SyncExecution(ctx, db, exec)
	if res.Progress != ProgressCompleted {
		t.Fatalf("published letter with approved LHA: progress %d, want %d", res.Progress, ProgressCompleted)
	}
}

// Real fieldwork evidence counts even while the letter is still a draft.
func TestSyncExecution_DraftLetterWithFieldworkEvidence(t *testing.T) {
	db := syncTestDB(t)
	exec := seedExecution(t, db, 0)
	seedLetter(t, db, letterKey, uuid.New(), "Draft")
	mustCreate(t, db, &models.FieldworkInterview{ID: uuid.New(), AssignmentLetterID: letterKey, Interviewee: "Kepala Kas"})

	res, err := SyncExecution(context.Background(), db, exec)
	if err != nil {
		t.Fatal(err)
	}
	if res.Progress != ProgressFieldworkMin || res.Sources.AssignmentLetterIssued {
		t.Fatalf("draft letter with interview: progress %d, sources %+v", res.Progress, res.Sources)
	}
}

// Letter statuses: anything but draft is issued. Missing letter is not issued.
func TestSyncExecution_LetterStatusRules(t *testing.T) {
	cases := []struct {
		name   string
		status string
		issued bool
	}{
		{"draft", "Draft", false},
		{"draft upper and padded", "  DRAFT ", false},
		{"blank", "", false},
		{"published", "Published", true},
		{"waiting approval", "Waiting Approval", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := syncTestDB(t)
			exec := seedExecution(t, db, 0)
			seedLetter(t, db, letterKey, uuid.New(), tc.status)
			res, err := SyncExecution(context.Background(), db, exec)
			if err != nil {
				t.Fatal(err)
			}
			if res.Sources.AssignmentLetterIssued != tc.issued {
				t.Fatalf("issued = %v, want %v", res.Sources.AssignmentLetterIssued, tc.issued)
			}
			want := ProgressPlanned
			if tc.issued {
				want = ProgressEntryMin
			}
			if res.Progress != want {
				t.Fatalf("progress = %d, want %d", res.Progress, want)
			}
		})
	}
}

// The entry band spreads with the scope, purpose and team recorded on the letter.
func TestSyncExecution_EntryAlignmentSpread(t *testing.T) {
	db := syncTestDB(t)
	exec := seedExecution(t, db, 0)
	mustCreate(t, db, &models.AssignmentLetter{
		ID: uuid.New(), LetterNumber: letterKey, Status: "Published",
		ScopeList: []string{"Kas kecil"}, PurposeList: []string{"Menilai kontrol kas"},
	})
	res, err := SyncExecution(context.Background(), db, exec)
	if err != nil {
		t.Fatal(err)
	}
	if res.Sources.EntryAlignment != 2 || res.Progress != 17 {
		t.Fatalf("scope and purpose: alignment %d progress %d, want 2 and 17", res.Sources.EntryAlignment, res.Progress)
	}

	var letter models.AssignmentLetter
	if err := db.First(&letter, "letter_number = ?", letterKey).Error; err != nil {
		t.Fatal(err)
	}
	letter.MembersList = []models.LetterMember{{Name: "Auditor A", Role: "Ketua"}}
	if err := db.Save(&letter).Error; err != nil {
		t.Fatal(err)
	}
	res, _ = SyncExecution(context.Background(), db, exec)
	if res.Sources.EntryAlignment != 3 || res.Progress != ProgressEntryMin+entrySpan {
		t.Fatalf("with team: alignment %d progress %d, want 3 and %d", res.Sources.EntryAlignment, res.Progress, ProgressEntryMin+entrySpan)
	}
}

// LHA status vocabulary: draft 76, in review 88, waiting on CAE 90, approved 100,
// rejected and cancelled do not count at all.
func TestSyncExecution_LHAStatuses(t *testing.T) {
	cases := []struct {
		name     string
		statuses []string
		want     int
	}{
		{"draft", []string{"DRAFT"}, ProgressLHADraft},
		{"draft mixed case", []string{"Draft"}, ProgressLHADraft},
		{"in review", []string{"IN_REVIEW"}, ProgressLHAReview},
		{"submitted", []string{"SUBMITTED"}, ProgressLHAReview},
		{"pending review", []string{"Pending Review"}, ProgressLHAReview},
		{"waiting approval", []string{"Waiting Approval"}, ProgressLHACAEPending},
		{"pending CAE", []string{"PENDING_CAE"}, ProgressLHACAEPending},
		{"CAE pending beside draft counts as CAE pending", []string{"DRAFT", "CAE PENDING"}, ProgressLHACAEPending},
		{"in review beside waiting CAE counts as CAE pending", []string{"IN_REVIEW", "Waiting Approval"}, ProgressLHACAEPending},
		{"waiting CAE beside final counts as final", []string{"Waiting Approval", "Final"}, ProgressCompleted},
		{"final", []string{"Final"}, ProgressCompleted},
		{"approved", []string{"APPROVED"}, ProgressCompleted},
		{"published", []string{"Published"}, ProgressCompleted},
		{"rejected", []string{"REJECTED"}, ProgressEntryMin},
		{"cancelled", []string{"Cancelled"}, ProgressEntryMin},
		{"canceled", []string{"CANCELED"}, ProgressEntryMin},
		{"rejected beside draft counts as draft", []string{"REJECTED", "DRAFT"}, ProgressLHADraft},
		{"cancelled beside final counts as final", []string{"CANCELLED", "Final"}, ProgressCompleted},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := syncTestDB(t)
			exec := seedExecution(t, db, 0)
			seedLetter(t, db, letterKey, uuid.New(), "Published")
			for _, st := range tc.statuses {
				mustCreate(t, db, &models.AuditResultReport{ID: uuid.New(), AssignmentLetterID: letterKey, Status: st})
			}
			res, err := SyncExecution(context.Background(), db, exec)
			if err != nil {
				t.Fatal(err)
			}
			if res.Progress != tc.want {
				t.Fatalf("progress = %d, want %d (sources %+v)", res.Progress, tc.want, res.Sources)
			}
		})
	}
}

// Every band reached through a full chain of evidence, with the letter, fieldwork,
// working paper and LHA keyed by letter number.
func TestSyncExecution_FullEvidenceReachesCompleted(t *testing.T) {
	db := syncTestDB(t)
	exec := seedExecution(t, db, 0)
	seedLetter(t, db, letterKey, uuid.New(), "Published")
	seedFullEvidence(t, db, letterKey)
	mustCreate(t, db, &models.AuditResultReport{ID: uuid.New(), AssignmentLetterID: letterKey, Status: "Final"})

	res, err := SyncExecution(context.Background(), db, exec)
	if err != nil {
		t.Fatal(err)
	}
	if res.Progress != ProgressCompleted {
		t.Fatalf("progress = %d, want %d (sources %+v)", res.Progress, ProgressCompleted, res.Sources)
	}
}

// Fieldwork, working paper and LHA rows stored under the letter UUID count the
// same as rows stored under the letter number.
func TestSyncExecution_UUIDKeyedRowsCount(t *testing.T) {
	db := syncTestDB(t)
	exec := seedExecution(t, db, 0)
	letterID := uuid.New()
	seedLetter(t, db, letterKey, letterID, "Published")
	seedFullEvidence(t, db, letterID.String())
	mustCreate(t, db, &models.AuditResultReport{ID: uuid.New(), AssignmentLetterID: letterID.String(), Status: "APPROVED"})

	res, err := SyncExecution(context.Background(), db, exec)
	if err != nil {
		t.Fatal(err)
	}
	s := res.Sources
	if s.FieldworkRecords != 1 || s.TestControls != 4 || s.TestControlsTested != 2 ||
		s.WorkingPaperCauses != 1 || s.WorkingPaperPlansComplete != 1 || !s.LHAApproved {
		t.Fatalf("UUID-keyed rows not collected: %+v", s)
	}
	if res.Progress != ProgressCompleted {
		t.Fatalf("progress = %d, want %d", res.Progress, ProgressCompleted)
	}
}

// An execution whose ref is the letter UUID finds the letter and its number-keyed rows.
func TestSyncExecution_RefIsLetterUUID(t *testing.T) {
	db := syncTestDB(t)
	letterID := uuid.New()
	seedLetter(t, db, letterKey, letterID, "Published")
	mustCreate(t, db, &models.FieldworkInterview{ID: uuid.New(), AssignmentLetterID: letterKey, Interviewee: "Kepala Kas"})
	exec := &models.AuditExecution{ID: uuid.New(), Ref: letterID.String(), Name: "Audit UUID", Progress: 0}
	mustCreate(t, db, exec)

	res, err := SyncExecution(context.Background(), db, exec)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Sources.AssignmentLetterIssued || res.Sources.FieldworkRecords != 1 || res.Progress != ProgressFieldworkMin {
		t.Fatalf("ref as UUID: progress %d, sources %+v", res.Progress, res.Sources)
	}
}

// Unmatched rows are rows whose letter reference matches no live letter.
func TestUnmatchedSourceRows(t *testing.T) {
	db := syncTestDB(t)
	seedLetter(t, db, letterKey, uuid.New(), "Published")
	mustCreate(t, db, &models.FieldworkInterview{ID: uuid.New(), AssignmentLetterID: "ST-404/SKAI/2026", Interviewee: "X"})
	mustCreate(t, db, &models.FieldworkInterview{ID: uuid.New(), AssignmentLetterID: letterKey, Interviewee: "Y"})
	mustCreate(t, db, &models.FieldworkTestControl{ID: uuid.New(), AssignmentLetterID: uuid.New().String(), ControlName: "C"})
	mustCreate(t, db, &models.WorkingPaperCause{ID: uuid.New(), WorkingPaperID: "ST-404/SKAI/2026", Condition: "X"})
	mustCreate(t, db, &models.AuditResultReport{ID: uuid.New(), AssignmentLetterID: letterKey, Status: "DRAFT"})

	got, err := UnmatchedSourceRows(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]int64{
		"fieldwork_interviews":    1,
		"fieldwork_test_controls": 1,
		"working_paper_causes":    1,
	}
	if len(got) != len(want) {
		t.Fatalf("unmatched = %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("unmatched[%s] = %d, want %d", k, got[k], v)
		}
	}
}

// A blank title or project code in the annual plan never links to an execution
// that has no name or ref.
func TestFindActivityID_BlankNameOrRefNeverMatches(t *testing.T) {
	db := syncTestDB(t)
	mustCreate(t, db, &models.AuditActivity{
		ID: uuid.New(), AnnualPlanID: uuid.New(), TargetUnitID: uuid.New(),
		ProjectCode: "", Title: "", Status: "PLANNED",
	})

	cases := []struct {
		name string
		exec models.AuditExecution
	}{
		{"blank name with ref", models.AuditExecution{Name: "", Ref: letterKey}},
		{"name with blank ref", models.AuditExecution{Name: "Audit Baru", Ref: ""}},
		{"both blank", models.AuditExecution{Name: "  ", Ref: ""}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id, err := findActivityID(db, &tc.exec)
			if err != nil {
				t.Fatal(err)
			}
			if id != nil {
				t.Fatalf("linked to activity %s, want no link", id)
			}
		})
	}
}

func TestSyncExecution_IgnoresSourcesOfOtherLetters(t *testing.T) {
	db := syncTestDB(t)
	exec := seedExecution(t, db, 0)
	mustCreate(t, db, &models.AuditResultReport{ID: uuid.New(), AssignmentLetterID: "ST-999/SKAI/2026", Status: "FINAL"})

	res, err := SyncExecution(context.Background(), db, exec)
	if err != nil {
		t.Fatal(err)
	}
	if res.Progress != ProgressPlanned || res.Sources.LHAExists {
		t.Fatalf("other letter's LHA leaked in: progress %d, sources %+v", res.Progress, res.Sources)
	}
}

func TestSyncAll_SyncsEveryExecution(t *testing.T) {
	db := syncTestDB(t)
	seedExecution(t, db, 10)
	exec2 := &models.AuditExecution{ID: uuid.New(), Ref: "ST-002/SKAI/2026", Name: "Audit IT", Progress: 90}
	mustCreate(t, db, exec2)
	seedLetter(t, db, "ST-002/SKAI/2026", uuid.New(), "Published")
	mustCreate(t, db, &models.AuditResultReport{ID: uuid.New(), AssignmentLetterID: "ST-002/SKAI/2026", Status: "APPROVED"})

	results, err := SyncAll(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("synced %d executions, want 2", len(results))
	}
	byRef := map[string]int{}
	for _, r := range results {
		byRef[r.Ref] = r.Progress
	}
	if byRef[letterKey] != ProgressPlanned || byRef["ST-002/SKAI/2026"] != ProgressCompleted {
		t.Fatalf("progress by ref = %v", byRef)
	}
}

// The column constants must match the GORM schema, so a model rename fails here
// rather than producing zero counts.
func TestColumnNames(t *testing.T) {
	check := func(model interface{}, field, want string) {
		t.Helper()
		s, err := schema.Parse(model, &sync.Map{}, schema.NamingStrategy{})
		if err != nil {
			t.Fatal(err)
		}
		f := s.LookUpField(field)
		if f == nil {
			t.Fatalf("%T has no field %s", model, field)
		}
		if f.DBName != want {
			t.Errorf("%T.%s column = %q, want %q", model, field, f.DBName, want)
		}
	}
	check(&models.FieldworkTestControl{}, "AssignmentLetterID", colAssignmentLetterID)
	check(&models.FieldworkTestControl{}, "MitigationPlan", colMitigationPlan)
	check(&models.FieldworkInterview{}, "AssignmentLetterID", colAssignmentLetterID)
	check(&models.AuditResultReport{}, "AssignmentLetterID", colAssignmentLetterID)
	check(&models.WorkingPaperCause{}, "WorkingPaperID", colWorkingPaperID)
	check(&models.WorkingPaperPlan{}, "WorkingPaperID", colWorkingPaperID)
}

// countExecutionUpdates counts UPDATE statements that write audit_executions.
func countExecutionUpdates(db *gorm.DB) *int {
	n := new(int)
	_ = db.Callback().Update().After("gorm:update").Register("test:count_execution_updates", func(tx *gorm.DB) {
		if tx.Statement.Table == "audit_executions" {
			*n++
		}
	})
	return n
}

// SyncExecutionsForLetter re-derives the executions whose ref is the letter
// number, or the letter UUID, and leaves executions of other letters alone.
func TestSyncExecutionsForLetter_ScopesToOneLetter(t *testing.T) {
	db := syncTestDB(t)
	ctx := context.Background()
	letterID := uuid.New()
	seedLetter(t, db, letterKey, letterID, "Published")
	mustCreate(t, db, &models.FieldworkInterview{ID: uuid.New(), AssignmentLetterID: letterKey, Interviewee: "Kepala Kas"})

	byNumber := seedExecution(t, db, 0)
	byUUID := &models.AuditExecution{ID: uuid.New(), Ref: letterID.String(), Name: "Audit UUID", Progress: 0}
	mustCreate(t, db, byUUID)
	other := &models.AuditExecution{ID: uuid.New(), Ref: "ST-999/SKAI/2026", Name: "Audit Lain", Progress: 7}
	mustCreate(t, db, other)

	// Stored progress 7 with no evidence behind it: a full sync would reset it to 0.
	res, err := SyncExecutionsForLetter(ctx, db, letterKey)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 2 {
		t.Fatalf("synced %d executions, want 2 (number and UUID ref)", len(res))
	}
	for _, e := range []*models.AuditExecution{byNumber, byUUID} {
		if got := reload(t, db, e); got.Progress != ProgressFieldworkMin {
			t.Errorf("execution %s progress = %d, want %d", got.Ref, got.Progress, ProgressFieldworkMin)
		}
	}
	if got := reload(t, db, other); got.Progress != 7 {
		t.Fatalf("other letter's execution changed: progress %d, want 7", got.Progress)
	}

	// The UUID form of the key finds the same executions.
	res, err = SyncExecutionsForLetter(ctx, db, letterID.String())
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 2 {
		t.Fatalf("UUID key synced %d executions, want 2", len(res))
	}
}

func TestSyncExecutionsForLetter_NoopCases(t *testing.T) {
	db := syncTestDB(t)
	ctx := context.Background()
	other := seedExecution(t, db, 7)

	if res, err := SyncExecutionsForLetter(ctx, db, "  "); err != nil || res != nil {
		t.Fatalf("blank key: res %v err %v, want nil nil", res, err)
	}
	if res, err := SyncExecutionsForLetter(ctx, db, "ST-404/SKAI/2026"); err != nil || len(res) != 0 {
		t.Fatalf("key with no execution: res %v err %v, want none", res, err)
	}
	if got := reload(t, db, other); got.Progress != 7 {
		t.Fatalf("unrelated execution changed: progress %d", got.Progress)
	}
}

// The execution row is written only when progress, status or detail changed.
func TestSyncExecutionsForLetter_WritesOnlyOnChange(t *testing.T) {
	db := syncTestDB(t)
	ctx := context.Background()
	exec := seedExecution(t, db, 0)
	seedLetter(t, db, letterKey, uuid.New(), "Published")
	writes := countExecutionUpdates(db)

	if _, err := SyncExecutionsForLetter(ctx, db, letterKey); err != nil {
		t.Fatal(err)
	}
	if *writes != 1 {
		t.Fatalf("first sync wrote %d times, want 1 (entry band)", *writes)
	}

	*writes = 0
	if _, err := SyncExecutionsForLetter(ctx, db, letterKey); err != nil {
		t.Fatal(err)
	}
	if *writes != 0 {
		t.Fatalf("unchanged sync wrote %d times, want 0", *writes)
	}

	mustCreate(t, db, &models.FieldworkInterview{ID: uuid.New(), AssignmentLetterID: letterKey, Interviewee: "Kepala Kas"})
	*writes = 0
	if _, err := SyncExecutionsForLetter(ctx, db, letterKey); err != nil {
		t.Fatal(err)
	}
	if *writes != 1 {
		t.Fatalf("changed sync wrote %d times, want 1", *writes)
	}
	if got := reload(t, db, exec); got.Progress != ProgressFieldworkMin {
		t.Fatalf("progress = %d, want %d", got.Progress, ProgressFieldworkMin)
	}
}

// A failing source read is returned, so the caller can roll back or log it.
func TestSyncExecutionsForLetter_ReturnsSourceError(t *testing.T) {
	db := syncTestDB(t)
	seedExecution(t, db, 0)
	seedLetter(t, db, letterKey, uuid.New(), "Published")
	if err := db.Migrator().DropTable(&models.FieldworkTestControl{}); err != nil {
		t.Fatal(err)
	}
	if _, err := SyncExecutionsForLetter(context.Background(), db, letterKey); err == nil {
		t.Fatal("expected an error when a source table is missing")
	}
}
