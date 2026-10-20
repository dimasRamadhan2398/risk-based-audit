package models

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// openLegacyLetterDB mirrors a production database created before the
// letter_number index became partial: the unique index covers soft-deleted rows.
func openLegacyLetterDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	// Table as the legacy schema had it: the letter_number unique index covers
	// soft-deleted rows. (AutoMigrate cannot be used here: sqlite has no gen_random_uuid.)
	if err := db.Exec(`CREATE TABLE assignment_letters (
		id TEXT PRIMARY KEY,
		letter_number VARCHAR(100) NOT NULL,
		status VARCHAR(50) DEFAULT 'Draft',
		audit_title VARCHAR(255),
		leader VARCHAR(200),
		category VARCHAR(100),
		audit_year VARCHAR(10),
		audit_team VARCHAR(100),
		start_period VARCHAR(100),
		finish_period VARCHAR(100),
		working_unit VARCHAR(255),
		company_id TEXT,
		company_name VARCHAR(255),
		execution_period VARCHAR(255),
		audit_purpose TEXT,
		letter_date DATETIME,
		cae_signature TEXT,
		members_list TEXT,
		purpose_list TEXT,
		scope_list TEXT,
		cc_list TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create table: %v", err)
	}
	if err := db.Exec(`CREATE UNIQUE INDEX idx_assignment_letters_letter_number ON assignment_letters (letter_number)`).Error; err != nil {
		t.Fatalf("create index: %v", err)
	}
	return db
}

func newTestLetter(team, year string) *AssignmentLetter {
	return &AssignmentLetter{ID: uuid.New(), AuditTeam: team, AuditYear: year, Status: "Draft"}
}

// A number used by a soft-deleted letter must not be handed out again, because
// the legacy unique index still counts that row.
func TestCreateLetterSkipsNumberOfSoftDeletedLetter(t *testing.T) {
	db := openLegacyLetterDB(t)

	first := newTestLetter("SKAI", "2026")
	if err := db.Create(first).Error; err != nil {
		t.Fatalf("create first: %v", err)
	}
	if first.LetterNumber != "ST-001/SKAI/2026" {
		t.Fatalf("unexpected first number %q", first.LetterNumber)
	}
	if err := db.Delete(first).Error; err != nil {
		t.Fatalf("soft delete: %v", err)
	}

	second := newTestLetter("SKAI", "2026")
	if err := db.Create(second).Error; err != nil {
		t.Fatalf("create after soft delete failed: %v", err)
	}
	if second.LetterNumber == first.LetterNumber {
		t.Fatalf("reused soft-deleted number %q", second.LetterNumber)
	}
}

// A client-supplied number that matches a soft-deleted letter is replaced.
func TestCreateLetterReplacesClientNumberOfSoftDeletedLetter(t *testing.T) {
	db := openLegacyLetterDB(t)

	old := newTestLetter("SKAI", "2026")
	old.LetterNumber = "ST-005/SKAI/2026"
	if err := db.Create(old).Error; err != nil {
		t.Fatalf("create old: %v", err)
	}
	if err := db.Delete(old).Error; err != nil {
		t.Fatalf("soft delete: %v", err)
	}

	next := newTestLetter("SKAI", "2026")
	next.LetterNumber = "ST-005/SKAI/2026"
	if err := db.Create(next).Error; err != nil {
		t.Fatalf("create with soft-deleted number failed: %v", err)
	}
	if next.LetterNumber == "ST-005/SKAI/2026" {
		t.Fatalf("kept soft-deleted number")
	}
}
