package findings

import (
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// The Postgres SQL LoadRecent builds: soft-deleted rows excluded everywhere
// except the assignment-letter lookup (which needs deleted letters' ids and
// numbers), and deterministic ordering
func TestLoadRecentSQLShape(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=127.0.0.1 dbname=none"}), &gorm.Config{
		DryRun:               true,
		DisableAutomaticPing: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	var sqls []string
	capture := func(tx *gorm.DB) {
		sqls = append(sqls, tx.Statement.SQL.String())
		tx.Statement.SQL.Reset()
		tx.Statement.Vars = nil
	}
	_ = db.Callback().Query().After("gorm:query").Register("test:capture", capture)

	if _, err := LoadRecent(db, 5); err != nil {
		t.Fatal(err)
	}
	want := []string{
		`SELECT "id","letter_number","deleted_at" FROM "assignment_letters"`,
		`SELECT "id","assignment_letter_id","findings","report_date","created_at" FROM "audit_result_reports" WHERE "audit_result_reports"."deleted_at" IS NULL ORDER BY created_at ASC,id ASC`,
		`SELECT * FROM "working_paper_causes" WHERE "working_paper_causes"."deleted_at" IS NULL ORDER BY working_paper_id ASC,created_at ASC,id ASC`,
		`SELECT * FROM "working_paper_plans" WHERE "working_paper_plans"."deleted_at" IS NULL ORDER BY working_paper_id ASC,created_at ASC,id ASC`,
		`SELECT * FROM "working_paper_risks" WHERE "working_paper_risks"."deleted_at" IS NULL ORDER BY working_paper_id ASC,created_at ASC,id ASC`,
		`SELECT * FROM "fieldwork_test_controls" WHERE "fieldwork_test_controls"."deleted_at" IS NULL ORDER BY assignment_letter_id ASC,created_at ASC,id ASC`,
	}
	if len(sqls) != len(want) {
		t.Fatalf("got %d statements:\n%s", len(sqls), strings.Join(sqls, "\n"))
	}
	for i := range want {
		if sqls[i] != want[i] {
			t.Errorf("statement %d:\n got  %s\n want %s", i, sqls[i], want[i])
		}
	}
}
