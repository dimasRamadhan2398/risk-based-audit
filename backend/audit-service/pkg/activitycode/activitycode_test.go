package activitycode_test

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"audit-service/models"
	"audit-service/pkg/activitycode"
	"audit-service/pkg/activitycode/sqlitetest"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestTypeCode(t *testing.T) {
	cases := map[string]string{
		// frontend AuditCategory values
		"Assurance":                "ASR",
		"Special Audit":            "SPC",
		"Specific Reason":          "SPR",
		"Consulting Services":      "CNS",
		"Investigation":            "INV",
		"Quality Assurance Review": "QAR",
		"Follow-Up Audit":          "FUA",
		// enum keys, stored short forms, case and spacing
		"ASSURANCE":                      "ASR",
		"CONSULTING":                     "CNS",
		"consulting_services":            "CNS",
		"SPECIAL_AUDIT":                  "SPC",
		"Special":                        "SPC",
		"FOLLOWUP_AUDIT":                 "FUA",
		"Quality Assurance Review (QAR)": "QAR",
		"  assurance ":                   "ASR",
		// empty and unknown
		"":                          "AUD",
		"   ":                       "AUD",
		"Operational":               "OPE",
		"IT":                        "IT",
		"Information Technology":    "IT",
		"Accounting Systems Review": "AUD", // initials would clash with ASR
	}
	for in, want := range cases {
		if got := activitycode.TypeCode(in); got != want {
			t.Errorf("TypeCode(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFormatAndParse(t *testing.T) {
	if got := activitycode.Format("ASR", 2026, 3); got != "ASR-2026-003" {
		t.Errorf("Format = %q", got)
	}
	if got := activitycode.Format("CNS", 2026, 1234); got != "CNS-2026-1234" {
		t.Errorf("Format past 999 = %q", got)
	}
	tc, y, seq, ok := activitycode.Parse("SPC-2025-012")
	if !ok || tc != "SPC" || y != 2025 || seq != 12 {
		t.Errorf("Parse = %q %d %d %v", tc, y, seq, ok)
	}
	for _, bad := range []string{"", "1728371234567", "ASR-26-001", "ASR-2026-01", "asr-2026-001", "ASR-2026-001x"} {
		if _, _, _, ok := activitycode.Parse(bad); ok {
			t.Errorf("Parse(%q) ok, want not ok", bad)
		}
	}
}

func TestPlanYear(t *testing.T) {
	if y := activitycode.PlanYear("2026", "2025-01-01"); y != 2026 {
		t.Errorf("plan year first: %d", y)
	}
	if y := activitycode.PlanYear("", "2027-04-01"); y != 2027 {
		t.Errorf("period start fallback: %d", y)
	}
	if y := activitycode.PlanYear("FY 2028/2029"); y != 2028 {
		t.Errorf("embedded year: %d", y)
	}
	if y := activitycode.PlanYear("", "n/a"); y != time.Now().Year() {
		t.Errorf("no year: %d", y)
	}
}

func openDB(t *testing.T) *gorm.DB {
	return sqlitetest.Open(t, &models.ActivityCodeSequence{}, &models.AuditActivity{}, &models.ActivityPlan{})
}

func next(t *testing.T, db *gorm.DB, auditType string, year int) string {
	t.Helper()
	var code string
	if err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		code, err = activitycode.Next(tx, auditType, year)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	return code
}

func TestNextCountsPerTypeAndYear(t *testing.T) {
	db := openDB(t)
	steps := []struct {
		typ  string
		year int
		want string
	}{
		{"Assurance", 2026, "ASR-2026-001"},
		{"Assurance", 2026, "ASR-2026-002"},
		{"Consulting Services", 2026, "CNS-2026-001"},
		{"Assurance", 2025, "ASR-2025-001"},
		{"ASSURANCE", 2026, "ASR-2026-003"},
		{"", 2026, "AUD-2026-001"},
	}
	for _, s := range steps {
		if got := next(t, db, s.typ, s.year); got != s.want {
			t.Errorf("Next(%q, %d) = %q, want %q", s.typ, s.year, got, s.want)
		}
	}
}

func TestRollbackReturnsNumber(t *testing.T) {
	db := openDB(t)
	next(t, db, "Assurance", 2026) // 001
	_ = db.Transaction(func(tx *gorm.DB) error {
		if _, err := activitycode.Next(tx, "Assurance", 2026); err != nil {
			t.Fatal(err)
		}
		return fmt.Errorf("insert failed")
	})
	if got := next(t, db, "Assurance", 2026); got != "ASR-2026-002" {
		t.Errorf("after rollback got %q, want ASR-2026-002", got)
	}
}

func insertActivity(t *testing.T, db *gorm.DB, code string) *models.AuditActivity {
	t.Helper()
	a := &models.AuditActivity{AnnualPlanID: uuid.New(), TargetUnitID: uuid.New(), ProjectCode: code, Title: code}
	if err := db.Create(a).Error; err != nil {
		t.Fatal(err)
	}
	return a
}

func TestNumbersAreNotReusedAfterDelete(t *testing.T) {
	db := openDB(t)
	a1 := insertActivity(t, db, next(t, db, "Assurance", 2026))
	a2 := insertActivity(t, db, next(t, db, "Assurance", 2026))
	// soft delete the newest and hard delete the other
	if err := db.Delete(a2).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Unscoped().Delete(a1).Error; err != nil {
		t.Fatal(err)
	}
	if got := next(t, db, "Assurance", 2026); got != "ASR-2026-003" {
		t.Errorf("after deletes got %q, want ASR-2026-003", got)
	}
}

func TestStartsAfterCodesStoredOutsideTheCounter(t *testing.T) {
	db := openDB(t)
	// seed-style rows: a project code, a soft-deleted one, and a planned
	// activity code inside an activity plan's JSON
	insertActivity(t, db, "ASR-2026-004")
	deleted := insertActivity(t, db, "ASR-2026-006")
	if err := db.Delete(deleted).Error; err != nil {
		t.Fatal(err)
	}
	plan := &models.ActivityPlan{PlanTitle: "p", PlanYear: "2026", PlannedActivities: []models.PlannedActivity{
		{ID: "1", ActivityCode: "ASR-2026-009", Category: "Assurance"},
		{ID: "2", ActivityCode: "CNS-2026-002", Category: "Consulting Services"},
	}}
	if err := db.Create(plan).Error; err != nil {
		t.Fatal(err)
	}
	insertActivity(t, db, "ACT-2026-050") // legacy format, other prefix: ignored

	if got := next(t, db, "Assurance", 2026); got != "ASR-2026-010" {
		t.Errorf("got %q, want ASR-2026-010", got)
	}
	if got := next(t, db, "Consulting Services", 2026); got != "CNS-2026-003" {
		t.Errorf("got %q, want CNS-2026-003", got)
	}
	// counter row exists now; a code stored later above it is still respected
	insertActivity(t, db, "ASR-2026-020")
	if got := next(t, db, "Assurance", 2026); got != "ASR-2026-021" {
		t.Errorf("got %q, want ASR-2026-021", got)
	}
}

func TestNextManyKeepsRequestOrder(t *testing.T) {
	db := openDB(t)
	var codes []string
	err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		codes, err = activitycode.NextMany(tx, 2026, []activitycode.Request{
			{AuditType: "Consulting Services"}, {AuditType: "Assurance"}, {AuditType: "Consulting Services"}, {AuditType: "Assurance"},
		})
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"CNS-2026-001", "ASR-2026-001", "CNS-2026-002", "ASR-2026-002"}
	if strings.Join(codes, ",") != strings.Join(want, ",") {
		t.Errorf("NextMany = %v, want %v", codes, want)
	}
}

// Concurrent creates, each in its own transaction (code + insert), must get
// distinct consecutive codes. sqlite serialises the transactions; on Postgres
// the ON CONFLICT row lock does (see TestReserveOnPostgres).
func TestConcurrentCreatesGetDistinctCodes(t *testing.T) {
	db := openDB(t)
	assertConcurrentCodes(t, db, 25)
}

func assertConcurrentCodes(t *testing.T, db *gorm.DB, n int) {
	t.Helper()
	var wg sync.WaitGroup
	codes := make([]string, n)
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = db.Transaction(func(tx *gorm.DB) error {
				code, err := activitycode.Next(tx, "Assurance", 2026)
				if err != nil {
					return err
				}
				codes[i] = code
				return tx.Create(&models.AuditActivity{AnnualPlanID: uuid.New(), TargetUnitID: uuid.New(), ProjectCode: code, Title: code}).Error
			})
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("worker %d: %v", i, err)
		}
	}
	sort.Strings(codes)
	for i, c := range codes {
		if want := activitycode.Format("ASR", 2026, i+1); c != want {
			t.Fatalf("codes = %v; position %d is %q, want %q", codes, i, c, want)
		}
	}
}

// The counter statement as Postgres receives it.
func TestReserveSQLShapeOnPostgres(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=127.0.0.1 dbname=none"}), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
		Logger:               logger.Discard,
	})
	if err != nil {
		t.Fatal(err)
	}
	var sqls []string
	capture := func(tx *gorm.DB) { sqls = append(sqls, tx.Statement.SQL.String()) }
	_ = db.Callback().Row().After("gorm:row").Register("test:capture", capture)
	_ = db.Callback().Query().After("gorm:query").Register("test:capture", capture)

	_, _ = activitycode.Reserve(db, "ASR-2026", 2)
	joined := strings.Join(sqls, "\n")
	for _, want := range []string{
		`SELECT "project_code" FROM "audit_activities" WHERE project_code LIKE $1`,
		`SELECT "planned_activities" FROM "activity_plans" WHERE planned_activities LIKE $1`,
		"INSERT INTO activity_code_sequences (prefix, last_value, updated_at)\nVALUES ($1, CAST($2 AS INTEGER), $3)",
		"ON CONFLICT (prefix) DO UPDATE SET",
		"RETURNING last_value",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("SQL missing %q:\n%s", want, joined)
		}
	}
	if strings.Contains(joined, "deleted_at") {
		t.Errorf("stored-code scan must include soft-deleted rows:\n%s", joined)
	}
}

// Runs against a real Postgres when ACTIVITYCODE_TEST_PG_DSN is set, in a
// throwaway schema that is dropped afterwards. Never point it at a database
// holding real data.
func TestReserveOnPostgres(t *testing.T) {
	dsn := os.Getenv("ACTIVITYCODE_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("ACTIVITYCODE_TEST_PG_DSN not set")
	}
	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("activitycode_test_%d", time.Now().UnixNano())
	if err := admin.Exec(`CREATE SCHEMA "` + schema + `"`).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { admin.Exec(`DROP SCHEMA "` + schema + `" CASCADE`) })

	db, err := gorm.Open(postgres.Open(dsn+" search_path="+schema), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.ActivityCodeSequence{}, &models.AuditAnnual{}, &models.AuditActivity{}, &models.ActivityPlan{}); err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(20)
	// audit_activities.annual_plan_id has a foreign key on Postgres
	ann := &models.AuditAnnual{Year: 2026}
	if err := db.Create(ann).Error; err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	n := 40
	codes := make([]string, n)
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errs[i] = db.Transaction(func(tx *gorm.DB) error {
				code, err := activitycode.Next(tx, "Assurance", 2026)
				if err != nil {
					return err
				}
				codes[i] = code
				return tx.Create(&models.AuditActivity{AnnualPlanID: ann.ID, TargetUnitID: uuid.New(), ProjectCode: code, Title: code}).Error
			})
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("worker %d: %v", i, err)
		}
	}
	sort.Strings(codes)
	for i, c := range codes {
		if want := activitycode.Format("ASR", 2026, i+1); c != want {
			t.Fatalf("codes = %v; position %d is %q, want %q", codes, i, c, want)
		}
	}
	if got := next(t, db, "Consulting Services", 2026); got != "CNS-2026-001" {
		t.Errorf("got %q", got)
	}
}
