package employee

import (
	"fmt"
	"testing"

	"master-service/models"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// employeesDDL mirrors models.Employee in sqlite syntax. AutoMigrate emits
// Postgres-only defaults (gen_random_uuid) that sqlite rejects, so the table is
// created by hand. Every column GORM writes must be present.
const employeesDDL = `CREATE TABLE employees (
	id TEXT PRIMARY KEY,
	employee_code TEXT NOT NULL UNIQUE,
	full_name TEXT NOT NULL,
	email TEXT NOT NULL UNIQUE,
	phone TEXT,
	company_id TEXT,
	department_id TEXT,
	job_role_id TEXT,
	level_grade INTEGER NOT NULL DEFAULT 1,
	work_location_id TEXT,
	residence_address TEXT,
	residence_city TEXT,
	residence_province TEXT,
	residence_postal TEXT,
	manager_id TEXT,
	is_active NUMERIC,
	join_date DATETIME,
	created_at DATETIME,
	updated_at DATETIME,
	deleted_at DATETIME
)`

// newTestRepo opens a single-connection in-memory sqlite database (each
// ":memory:" connection is a separate database) with the employees table.
func newTestRepo(t *testing.T) (*EmployeeRepository, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql handle: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := db.Exec(employeesDDL).Error; err != nil {
		t.Fatalf("create employees: %v", err)
	}
	return NewEmployeeRepository(db).(*EmployeeRepository), db
}

func insertEmployee(t *testing.T, db *gorm.DB, code, name, email string) *models.Employee {
	t.Helper()
	e := &models.Employee{
		ID:           uuid.New(),
		EmployeeCode: code,
		FullName:     name,
		Email:        email,
		LevelGrade:   1,
		IsActive:     true,
	}
	if err := db.Create(e).Error; err != nil {
		t.Fatalf("insert %s: %v", code, err)
	}
	return e
}

func codesOf(rows []*models.Employee) []string {
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.EmployeeCode)
	}
	return out
}

func seedPaging(t *testing.T, db *gorm.DB, n int) {
	t.Helper()
	for i := 1; i <= n; i++ {
		insertEmployee(t, db, fmt.Sprintf("EMP-%04d", i), fmt.Sprintf("Person %d", i), fmt.Sprintf("person%d@example.com", i))
	}
}

func TestFindPage_OrdersByCodeAndSlicesPages(t *testing.T) {
	repo, db := newTestRepo(t)
	// Insert out of order to prove the ORDER BY, not insertion order, drives paging.
	for _, i := range []int{7, 2, 9, 1, 5, 3, 8, 4, 6} {
		insertEmployee(t, db, fmt.Sprintf("EMP-%04d", i), fmt.Sprintf("Person %d", i), fmt.Sprintf("p%d@example.com", i))
	}

	page1, err := repo.FindPage(0, 4, "")
	if err != nil {
		t.Fatalf("FindPage page 1: %v", err)
	}
	if got := codesOf(page1); fmt.Sprint(got) != "[EMP-0001 EMP-0002 EMP-0003 EMP-0004]" {
		t.Fatalf("page 1 = %v", got)
	}
	page2, _ := repo.FindPage(4, 4, "")
	if got := codesOf(page2); fmt.Sprint(got) != "[EMP-0005 EMP-0006 EMP-0007 EMP-0008]" {
		t.Fatalf("page 2 = %v", got)
	}
	last, _ := repo.FindPage(8, 4, "")
	if got := codesOf(last); fmt.Sprint(got) != "[EMP-0009]" {
		t.Fatalf("last page = %v", got)
	}
	past, _ := repo.FindPage(12, 4, "")
	if len(past) != 0 {
		t.Fatalf("page past end returned %d rows", len(past))
	}
}

func TestCount_MatchesFindPageTotal(t *testing.T) {
	repo, db := newTestRepo(t)
	seedPaging(t, db, 12)

	total, err := repo.Count("")
	if err != nil || total != 12 {
		t.Fatalf("Count(\"\") = %d, %v; want 12", total, err)
	}
	// EMP-0010, EMP-0011, EMP-0012 contain "EMP-001".
	filtered, err := repo.Count("EMP-001")
	if err != nil || filtered != 3 {
		t.Fatalf("Count(EMP-001) = %d, %v; want 3", filtered, err)
	}
}

func TestFindPage_SearchIsCaseInsensitiveAcrossCodeNameEmail(t *testing.T) {
	repo, db := newTestRepo(t)
	insertEmployee(t, db, "EMP-0001", "Siti Rahayu", "siti@bank.example")
	insertEmployee(t, db, "EMP-0002", "Budi Santoso", "budi.santoso@corp.example")
	insertEmployee(t, db, "EMP-0003", "Dewi Lestari", "dewi@other.example")

	cases := map[string][]string{
		"emp-0002":      {"EMP-0002"}, // code, lower case
		"RAHAYU":        {"EMP-0001"}, // name, upper case
		"corp.example":  {"EMP-0002"}, // email
		"example":       {"EMP-0001", "EMP-0002", "EMP-0003"},
		"no-such-value": {},
	}
	for search, want := range cases {
		rows, err := repo.FindPage(0, 10, search)
		if err != nil {
			t.Fatalf("FindPage(%q): %v", search, err)
		}
		if got := codesOf(rows); fmt.Sprint(got) != fmt.Sprint(want) {
			t.Errorf("FindPage(%q) = %v, want %v", search, got, want)
		}
		total, err := repo.Count(search)
		if err != nil || int(total) != len(want) {
			t.Errorf("Count(%q) = %d, %v; want %d", search, total, err, len(want))
		}
	}
}

func TestFindPage_SearchTreatsWildcardsLiterally(t *testing.T) {
	repo, db := newTestRepo(t)
	insertEmployee(t, db, "EMP-0001", "100% Ali", "ali@example.com")
	insertEmployee(t, db, "EMP-0002", "Under_score", "under@example.com")
	insertEmployee(t, db, "EMP-0003", "Plain Name", "plain@example.com")

	// Unescaped "%" would match every row and "_" would match any single
	// character. Both must match only the literal text.
	if total, _ := repo.Count("%"); total != 1 {
		t.Fatalf("Count(%%) = %d, want 1 (only the literal %% name)", total)
	}
	if total, _ := repo.Count("_"); total != 1 {
		t.Fatalf("Count(_) = %d, want 1 (only the literal _ name)", total)
	}
	if total, _ := repo.Count("!"); total != 0 {
		t.Fatalf("Count(!) = %d, want 0 (the escape character itself is literal)", total)
	}
}

func TestFindPage_SoftDeletedRowsAreExcluded(t *testing.T) {
	repo, db := newTestRepo(t)
	seedPaging(t, db, 3)
	if err := db.Where("employee_code = ?", "EMP-0002").Delete(&models.Employee{}).Error; err != nil {
		t.Fatalf("soft delete: %v", err)
	}

	total, _ := repo.Count("")
	if total != 2 {
		t.Fatalf("Count after soft delete = %d, want 2", total)
	}
	rows, _ := repo.FindPage(0, 10, "")
	if got := codesOf(rows); fmt.Sprint(got) != "[EMP-0001 EMP-0003]" {
		t.Fatalf("FindPage after soft delete = %v", got)
	}
}
