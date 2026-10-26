package audit_execution_sync

import (
	"context"
	"strings"
	"testing"

	"audit-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func newLetter(number, status string) *models.AssignmentLetter {
	return &models.AssignmentLetter{
		ID:           uuid.New(),
		LetterNumber: number,
		Status:       status,
		AuditTitle:   "Audit Kas " + number,
		Category:     "Assurance",
		MembersList: []models.LetterMember{
			{Name: "Rina Wulandari", Role: "Member"},
			{Name: "Zeta Ramadhani", Role: "Chairperson"},
		},
	}
}

func liveExecutions(t *testing.T, db *gorm.DB) []models.AuditExecution {
	t.Helper()
	var out []models.AuditExecution
	if err := db.Order("ref ASC").Find(&out).Error; err != nil {
		t.Fatal(err)
	}
	return out
}

func mustEnsure(t *testing.T, db *gorm.DB, l, prev *models.AssignmentLetter) bool {
	t.Helper()
	created, err := EnsureExecutionForLetter(context.Background(), db, l, prev)
	if err != nil {
		t.Fatal(err)
	}
	return created
}

func TestEnsureExecutionForLetter_CreatesFromDraftLetter(t *testing.T) {
	db := syncTestDB(t)
	l := newLetter(letterKey, "Draft")
	mustCreate(t, db, l)

	if err := ReconcileLetter(context.Background(), db, l.ID, nil); err != nil {
		t.Fatal(err)
	}
	execs := liveExecutions(t, db)
	if len(execs) != 1 {
		t.Fatalf("executions = %d, want 1", len(execs))
	}
	e := execs[0]
	if e.Ref != letterKey || e.Name != l.AuditTitle || e.Category != "Assurance" || e.LeadAuditor != "Zeta Ramadhani" {
		t.Fatalf("execution = %+v, want ref/name/category/lead from the letter", e)
	}
	// Draft letter, no sources: planning.
	if e.Progress != 0 || e.Status != models.AuditStatusPlanning {
		t.Fatalf("progress %d status %q, want 0 planning", e.Progress, e.Status)
	}
}

func TestEnsureExecutionForLetter_LeaderFieldWinsAndBlankTitleUsesNumber(t *testing.T) {
	db := syncTestDB(t)
	l := newLetter(letterKey, "Published")
	l.Leader = "Andi Firmansyah"
	l.AuditTitle = "  "
	mustCreate(t, db, l)
	mustEnsure(t, db, l, nil)

	e := liveExecutions(t, db)[0]
	if e.LeadAuditor != "Andi Firmansyah" {
		t.Errorf("lead = %q, want the letter's leader", e.LeadAuditor)
	}
	if e.Name != letterKey {
		t.Errorf("name = %q, want the letter number", e.Name)
	}
}

func TestEnsureExecutionForLetter_TruncatesToColumnSize(t *testing.T) {
	db := syncTestDB(t)
	l := newLetter(letterKey, "Draft")
	l.AuditTitle = strings.Repeat("é", 300)
	mustCreate(t, db, l)
	mustEnsure(t, db, l, nil)
	if n := len([]rune(liveExecutions(t, db)[0].Name)); n != maxNameLen {
		t.Fatalf("name runes = %d, want %d", n, maxNameLen)
	}
}

func TestEnsureExecutionForLetter_Idempotent(t *testing.T) {
	db := syncTestDB(t)
	l := newLetter(letterKey, "Draft")
	mustCreate(t, db, l)
	if !mustEnsure(t, db, l, nil) {
		t.Fatal("first ensure did not create")
	}
	if mustEnsure(t, db, l, nil) {
		t.Fatal("second ensure created a duplicate")
	}
	if n := len(liveExecutions(t, db)); n != 1 {
		t.Fatalf("executions = %d, want 1", n)
	}
}

// A seeded execution with the letter number is the letter's execution: no new
// row, user-visible fields kept, blank ones filled.
func TestEnsureExecutionForLetter_SeededExecutionNotDuplicated(t *testing.T) {
	db := syncTestDB(t)
	seeded := &models.AuditExecution{ID: uuid.New(), Ref: letterKey, Name: "Audit Operasional Keuangan", Category: "Assurance", Progress: 80}
	mustCreate(t, db, seeded)
	l := newLetter(letterKey, "Published")
	mustCreate(t, db, l)

	if mustEnsure(t, db, l, nil) {
		t.Fatal("ensure created a duplicate of the seeded execution")
	}
	got := reload(t, db, seeded)
	if got.Name != "Audit Operasional Keuangan" {
		t.Errorf("name = %q, the seeded name must be kept", got.Name)
	}
	if got.LeadAuditor != "Zeta Ramadhani" {
		t.Errorf("lead = %q, a blank lead is filled from the letter", got.LeadAuditor)
	}
}

func TestEnsureExecutionForLetter_ExecutionReffingUUIDCounts(t *testing.T) {
	db := syncTestDB(t)
	l := newLetter(letterKey, "Draft")
	mustCreate(t, db, l)
	mustCreate(t, db, &models.AuditExecution{ID: uuid.New(), Ref: l.ID.String(), Name: "By id"})
	if mustEnsure(t, db, l, nil) {
		t.Fatal("ensure created a second execution for a letter already reffed by id")
	}
}

func TestEnsureExecutionForLetter_UpdateFollowsLetterUnlessEdited(t *testing.T) {
	db := syncTestDB(t)
	l := newLetter(letterKey, "Draft")
	mustCreate(t, db, l)
	mustEnsure(t, db, l, nil)
	exec := liveExecutions(t, db)[0]
	// A user renames the execution; category and lead still hold the letter's values.
	if err := db.Model(&exec).Update("name", "Nama dari user").Error; err != nil {
		t.Fatal(err)
	}

	prev := *l
	next := *l
	next.AuditTitle = "Judul baru"
	next.Category = "Consulting"
	next.MembersList = []models.LetterMember{{Name: "Budi Santoso", Role: "chairperson"}}
	if mustEnsure(t, db, &next, &prev) {
		t.Fatal("update created an execution")
	}
	got := reload(t, db, &exec)
	if got.Name != "Nama dari user" {
		t.Errorf("name = %q, the user's edit must be kept", got.Name)
	}
	if got.Category != "Consulting" || got.LeadAuditor != "Budi Santoso" {
		t.Errorf("category %q lead %q, want them to follow the letter", got.Category, got.LeadAuditor)
	}
}

func TestEnsureExecutionForLetter_RenumberMovesExecution(t *testing.T) {
	db := syncTestDB(t)
	l := newLetter(letterKey, "Draft")
	mustCreate(t, db, l)
	mustEnsure(t, db, l, nil)

	prev := *l
	const newNumber = "ST-099/SKAI/2026"
	if err := db.Model(l).Update("letter_number", newNumber).Error; err != nil {
		t.Fatal(err)
	}
	if mustEnsure(t, db, l, &prev) {
		t.Fatal("renumbering created a second execution")
	}
	execs := liveExecutions(t, db)
	if len(execs) != 1 || execs[0].Ref != newNumber {
		t.Fatalf("executions = %+v, want one with ref %s", execs, newNumber)
	}
}

func TestRetireExecutionsForLetter(t *testing.T) {
	db := syncTestDB(t)
	l := newLetter(letterKey, "Published")
	mustCreate(t, db, l)
	mustEnsure(t, db, l, nil)
	other := newLetter("ST-002/SKAI/2026", "Draft")
	mustCreate(t, db, other)
	mustEnsure(t, db, other, nil)

	// The letter is live: nothing is retired.
	if n, err := RetireExecutionsForLetter(context.Background(), db, l); err != nil || n != 0 {
		t.Fatalf("retire live letter = %d, %v; want 0", n, err)
	}

	if err := db.Delete(l).Error; err != nil {
		t.Fatal(err)
	}
	if err := ReconcileLetter(context.Background(), db, l.ID, l); err != nil {
		t.Fatal(err)
	}
	execs := liveExecutions(t, db)
	if len(execs) != 1 || execs[0].Ref != other.LetterNumber {
		t.Fatalf("live executions = %+v, want only %s", execs, other.LetterNumber)
	}
	var all int64
	if err := db.Unscoped().Model(&models.AuditExecution{}).Count(&all).Error; err != nil {
		t.Fatal(err)
	}
	if all != 2 {
		t.Fatalf("rows incl. deleted = %d, want 2 (soft delete)", all)
	}
	// Again: idempotent.
	if n, err := RetireExecutionsForLetter(context.Background(), db, l); err != nil || n != 0 {
		t.Fatalf("second retire = %d, %v; want 0", n, err)
	}
}

func TestEnsureExecutionsForAllLetters_BackfillIsIdempotent(t *testing.T) {
	db := syncTestDB(t)
	// Seeded execution for the first letter; none for the others.
	mustCreate(t, db, &models.AuditExecution{ID: uuid.New(), Ref: letterKey, Name: "Seeded"})
	// Seeded execution with no letter (a 2025 demo row): left alone.
	mustCreate(t, db, &models.AuditExecution{ID: uuid.New(), Ref: "ST-001/SKAI/2025", Name: "Old"})
	for _, n := range []string{letterKey, "ST-002/SKAI/2026", "ST-003/SKAI/2026"} {
		mustCreate(t, db, newLetter(n, "Draft"))
	}
	deleted := newLetter("ST-004/SKAI/2026", "Draft")
	mustCreate(t, db, deleted)
	if err := db.Delete(deleted).Error; err != nil {
		t.Fatal(err)
	}

	created, err := EnsureExecutionsForAllLetters(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if created != 2 {
		t.Fatalf("created = %d, want 2", created)
	}
	created, err = EnsureExecutionsForAllLetters(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if created != 0 {
		t.Fatalf("second run created = %d, want 0", created)
	}
	var refs []string
	for _, e := range liveExecutions(t, db) {
		refs = append(refs, e.Ref)
	}
	want := "ST-001/SKAI/2025,ST-001/SKAI/2026,ST-002/SKAI/2026,ST-003/SKAI/2026"
	if strings.Join(refs, ",") != want {
		t.Fatalf("refs = %v, want %s", refs, want)
	}
}
