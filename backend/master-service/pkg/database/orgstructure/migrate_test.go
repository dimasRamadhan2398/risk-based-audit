package orgstructure

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// recorder captures every SQL statement GORM sends.
type recorder struct {
	mu   sync.Mutex
	stmt []string
}

func (r *recorder) LogMode(logger.LogLevel) logger.Interface      { return r }
func (r *recorder) Info(context.Context, string, ...interface{})  {}
func (r *recorder) Warn(context.Context, string, ...interface{})  {}
func (r *recorder) Error(context.Context, string, ...interface{}) {}
func (r *recorder) Trace(_ context.Context, _ time.Time, fc func() (string, int64), _ error) {
	sql, _ := fc()
	r.mu.Lock()
	r.stmt = append(r.stmt, sql)
	r.mu.Unlock()
}

func (r *recorder) reset() { r.mu.Lock(); r.stmt = nil; r.mu.Unlock() }

// ddl returns the recorded statements that change schema or data.
func (r *recorder) writes() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []string
	for _, s := range r.stmt {
		u := strings.ToUpper(strings.TrimSpace(s))
		for _, verb := range []string{"CREATE", "ALTER", "DROP", "INSERT", "UPDATE", "DELETE", "DO "} {
			if strings.HasPrefix(u, verb) {
				out = append(out, s)
			}
		}
	}
	return out
}

func open(t *testing.T) (*gorm.DB, *recorder) {
	t.Helper()
	rec := &recorder{}
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: rec})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	return db, rec
}

// legacyTenant creates departments exactly as an existing tenant has it: the
// two global unique indexes the old uniqueIndex tags produced.
func legacyTenant(t *testing.T, db *gorm.DB) {
	t.Helper()
	for _, s := range []string{
		`CREATE TABLE departments (id TEXT PRIMARY KEY, department_code TEXT NOT NULL, department_name TEXT NOT NULL, company_id TEXT, deleted_at DATETIME)`,
		`CREATE UNIQUE INDEX idx_departments_department_code ON departments (department_code)`,
		`CREATE UNIQUE INDEX idx_departments_department_name ON departments (department_name)`,
	} {
		if err := db.Exec(s).Error; err != nil {
			t.Fatal(err)
		}
	}
}

// Existing legacy tenant, flag off: the org-structure step issues no DDL and
// no data change, on every run.
func TestLegacyMigrateIssuesNoDDL(t *testing.T) {
	db, rec := open(t)
	legacyTenant(t, db)
	for run := 0; run < 2; run++ {
		rec.reset()
		if err := Migrate(db, false); err != nil {
			t.Fatal(err)
		}
		if w := rec.writes(); len(w) != 0 {
			t.Fatalf("run %d: legacy migrate wrote: %v", run, w)
		}
	}
	if db.Migrator().HasTable("org_levels") || db.Migrator().HasColumn("departments", "parent_department_id") {
		t.Fatal("flexible schema appeared on a legacy tenant")
	}
}

// From-scratch legacy DB (no indexes yet): the step creates both legacy
// unique indexes, then is a no-op.
func TestLegacyMigrateCreatesMissingIndexesOnFreshDB(t *testing.T) {
	db, rec := open(t)
	if err := db.Exec(`CREATE TABLE departments (id TEXT PRIMARY KEY, department_code TEXT, department_name TEXT)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := Migrate(db, false); err != nil {
		t.Fatal(err)
	}
	for _, idx := range []string{LegacyCodeIndex, LegacyNameIndex} {
		if !db.Migrator().HasIndex("departments", idx) {
			t.Fatalf("%s missing", idx)
		}
	}
	rec.reset()
	if err := Migrate(db, false); err != nil {
		t.Fatal(err)
	}
	if w := rec.writes(); len(w) != 0 {
		t.Fatalf("second run wrote: %v", w)
	}
}

// Flag drift: org_levels exists but the flag is off. Nothing runs, and the
// global name index is NOT re-created (that would undo the index swap).
func TestFlagDriftDoesNotRecreateLegacyIndexes(t *testing.T) {
	db, rec := open(t)
	for _, s := range []string{
		`CREATE TABLE departments (id TEXT PRIMARY KEY, department_code TEXT, department_name TEXT)`,
		`CREATE TABLE org_levels (id TEXT PRIMARY KEY)`,
	} {
		if err := db.Exec(s).Error; err != nil {
			t.Fatal(err)
		}
	}
	rec.reset()
	if err := Migrate(db, false); err != nil {
		t.Fatal(err)
	}
	if w := rec.writes(); len(w) != 0 {
		t.Fatalf("drifted tenant got: %v", w)
	}
}

func TestFlexibleStatementsAreIdempotentAndComplete(t *testing.T) {
	joined := strings.Join(FlexibleStatements(), "\n")
	for _, s := range FlexibleStatements() {
		u := strings.ToUpper(s)
		if !strings.Contains(u, "IF NOT EXISTS") && !strings.Contains(u, "IF EXISTS") {
			t.Fatalf("statement not idempotent: %s", s)
		}
	}
	for _, want := range []string{"org_levels", "org_change_log", "parent_department_id", ParentFK, ParentCheck, ParentIndex,
		"DROP INDEX IF EXISTS idx_departments_department_name", NameIndex,
		"DROP INDEX IF EXISTS idx_departments_department_code", CodeIndexName()} {
		if !strings.Contains(joined, want) {
			t.Fatalf("flexible DDL lacks %s", want)
		}
	}
}

// D10 is one constant: the index name and DDL follow it.
func TestCodeScopeDrivesIndex(t *testing.T) {
	if DepartmentCodeScope != CodeScopeGlobal || CodeIndexName() != "uq_departments_code_live" ||
		!strings.Contains(codeIndexDDL(), "(department_code) WHERE deleted_at IS NULL") {
		t.Fatalf("default D10 policy changed: %s", codeIndexDDL())
	}
}
