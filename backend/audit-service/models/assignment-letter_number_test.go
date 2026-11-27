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
		leader_employee_id VARCHAR(64),
		leader_employee_code VARCHAR(64),
		category VARCHAR(100),
		audit_year VARCHAR(10),
		audit_team VARCHAR(100),
		start_period VARCHAR(100),
		finish_period VARCHAR(100),
		working_unit VARCHAR(255),
		working_unit_department_id TEXT,
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

// openPartialLetterDB mirrors the migrated production/local schema where the
// unique index on letter_number applies only to active rows (WHERE deleted_at IS NULL).
func openPartialLetterDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`CREATE TABLE assignment_letters (
		id TEXT PRIMARY KEY,
		letter_number VARCHAR(100) NOT NULL,
		status VARCHAR(50) DEFAULT 'Draft',
		audit_title VARCHAR(255),
		leader VARCHAR(200),
		leader_employee_id VARCHAR(64),
		leader_employee_code VARCHAR(64),
		category VARCHAR(100),
		audit_year VARCHAR(10),
		audit_team VARCHAR(100),
		start_period VARCHAR(100),
		finish_period VARCHAR(100),
		working_unit VARCHAR(255),
		working_unit_department_id TEXT,
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
	if err := db.Exec(`CREATE UNIQUE INDEX idx_assignment_letters_letter_number ON assignment_letters (letter_number) WHERE deleted_at IS NULL`).Error; err != nil {
		t.Fatalf("create partial index: %v", err)
	}
	return db
}

// Each audit team (SKAI, SPI, CAE, ...) and year must have an independent
// sequence starting at ST-001.
func TestLetterNumberStartsAtOnePerAuditTeamAndYear(t *testing.T) {
	db := openPartialLetterDB(t)

	for i, want := range []string{
		"ST-001/SKAI/2026",
		"ST-002/SKAI/2026",
		"ST-003/SKAI/2026",
		"ST-004/SKAI/2026",
	} {
		l := newTestLetter("SKAI", "2026")
		if err := db.Create(l).Error; err != nil {
			t.Fatalf("create SKAI #%d: %v", i+1, err)
		}
		if l.LetterNumber != want {
			t.Fatalf("SKAI #%d = %q, want %q", i+1, l.LetterNumber, want)
		}
	}

	// SPI must start at ST-001/SPI/2026, not continue SKAI's counter.
	spi1 := newTestLetter("SPI", "2026")
	if err := db.Create(spi1).Error; err != nil {
		t.Fatalf("create SPI #1: %v", err)
	}
	if spi1.LetterNumber != "ST-001/SPI/2026" {
		t.Fatalf("SPI #1 = %q, want ST-001/SPI/2026", spi1.LetterNumber)
	}

	spi2 := newTestLetter("SPI", "2026")
	if err := db.Create(spi2).Error; err != nil {
		t.Fatalf("create SPI #2: %v", err)
	}
	if spi2.LetterNumber != "ST-002/SPI/2026" {
		t.Fatalf("SPI #2 = %q, want ST-002/SPI/2026", spi2.LetterNumber)
	}

	// CAE must also start independently at ST-001/CAE/2026.
	cae1 := newTestLetter("CAE", "2026")
	if err := db.Create(cae1).Error; err != nil {
		t.Fatalf("create CAE #1: %v", err)
	}
	if cae1.LetterNumber != "ST-001/CAE/2026" {
		t.Fatalf("CAE #1 = %q, want ST-001/CAE/2026", cae1.LetterNumber)
	}

	// SKAI in a different year starts at 001.
	skai2025 := newTestLetter("SKAI", "2025")
	if err := db.Create(skai2025).Error; err != nil {
		t.Fatalf("create SKAI 2025 #1: %v", err)
	}
	if skai2025.LetterNumber != "ST-001/SKAI/2025" {
		t.Fatalf("SKAI 2025 #1 = %q, want ST-001/SKAI/2025", skai2025.LetterNumber)
	}

	// Subsequent SKAI 2026 continues sequentially at 005.
	skai5 := newTestLetter("SKAI", "2026")
	if err := db.Create(skai5).Error; err != nil {
		t.Fatalf("create SKAI #5: %v", err)
	}
	if skai5.LetterNumber != "ST-005/SKAI/2026" {
		t.Fatalf("SKAI #5 = %q, want ST-005/SKAI/2026", skai5.LetterNumber)
	}
}

// Regression test: soft-deleted outliers (such as deleted ST-005, ST-019, ST-9994)
// or active outlier numbers (ST-020 / ST-9995) must not cause numbers after #4
// to jump to 20 or 9995 under the partial unique index.
func TestLetterNumberDoesNotJumpOnDeletedOrOutlierNumbers(t *testing.T) {
	db := openPartialLetterDB(t)

	// Seed and soft-delete old outliers (simulating deleted ST-005, ST-019, ST-9994).
	for _, oldNum := range []string{"ST-005/SKAI/2026", "ST-019/SKAI/2026", "ST-9994/SKAI/2026"} {
		old := newTestLetter("SKAI", "2026")
		old.LetterNumber = oldNum
		if err := db.Create(old).Error; err != nil {
			t.Fatalf("create old %s: %v", oldNum, err)
		}
		if err := db.Delete(old).Error; err != nil {
			t.Fatalf("delete old %s: %v", oldNum, err)
		}
	}

	// Also create an active high outlier (ST-9995/SKAI/2026) and a legacy non-ST letter.
	outlier := newTestLetter("SKAI", "2026")
	outlier.LetterNumber = "ST-9995/SKAI/2026"
	if err := db.Create(outlier).Error; err != nil {
		t.Fatalf("create outlier: %v", err)
	}
	legacy := newTestLetter("SKAI", "2023")
	legacy.LetterNumber = "020/ST/01/KSIAD/2023"
	if err := db.Create(legacy).Error; err != nil {
		t.Fatalf("create legacy: %v", err)
	}

	// Sequential creations 1..6 must proceed 001..006 without jumping to 9996!
	for i := 1; i <= 6; i++ {
		l := newTestLetter("SKAI", "2026")
		if err := db.Create(l).Error; err != nil {
			t.Fatalf("create #%d: %v", i, err)
		}
		want := "ST-00" + string(rune('0'+i)) + "/SKAI/2026"
		if l.LetterNumber != want {
			t.Fatalf("#%d = %q, want %q", i, l.LetterNumber, want)
		}
	}

	// Even when the client sends ST-005/SKAI/2026 while soft-deleted ST-005 exists,
	// partial unique index allows reusing ST-005 if no active ST-005 exists.
	// Wait: here ST-005 was just created above, so sending ST-005 should move to ST-007.
	dup := newTestLetter("SKAI", "2026")
	dup.LetterNumber = "ST-005/SKAI/2026"
	if err := db.Create(dup).Error; err != nil {
		t.Fatalf("create colliding #7: %v", err)
	}
	if dup.LetterNumber != "ST-007/SKAI/2026" {
		t.Fatalf("colliding #7 = %q, want ST-007/SKAI/2026", dup.LetterNumber)
	}
}
