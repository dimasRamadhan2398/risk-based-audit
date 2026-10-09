package action_taken_report

import (
	"strings"
	"testing"

	"audit-service/models"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// The ON CONFLICT target must be inferable from a plain unique index on
// exactly (audit_result_report_id, finding_id): Postgres refuses a partial
// index (WHERE ...) unless the insert repeats the predicate.
func TestUniqueFindingIndexMatchesConflictTarget(t *testing.T) {
	stmt := &gorm.Statement{DB: newDB(t)}
	if err := stmt.Parse(&models.ActionTakenReport{}); err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, idx := range stmt.Schema.ParseIndexes() {
		if idx.Name != UniqueFindingIndex {
			continue
		}
		found = true
		if !strings.EqualFold(idx.Class, "UNIQUE") || idx.Where != "" || len(idx.Fields) != 2 ||
			idx.Fields[0].DBName != "audit_result_report_id" || idx.Fields[1].DBName != "finding_id" {
			t.Fatalf("index %s: class %q where %q fields %v", idx.Name, idx.Class, idx.Where, idx.Fields)
		}
	}
	if !found {
		t.Fatalf("model has no %s", UniqueFindingIndex)
	}

	// The insert SyncFromReport sends, in the Postgres dialect
	pg, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=127.0.0.1 dbname=none"}), &gorm.Config{DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.New()
	res := pg.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "audit_result_report_id"}, {Name: "finding_id"}},
		DoNothing: true,
	}).Omit(clause.Associations).Create(&models.ActionTakenReport{ID: id, AuditResultReportID: &id, FindingID: "f"})
	if res.Error != nil {
		t.Fatal(res.Error)
	}
	sql := res.Statement.SQL.String()
	if !strings.Contains(sql, `ON CONFLICT ("audit_result_report_id","finding_id") DO NOTHING`) || strings.Contains(sql, "WHERE") {
		t.Fatalf("insert SQL: %s", sql)
	}
}

// A table that lost the unique index (as sqlite's column-drop rebuild does)
// gets it back before the backfill inserts, instead of failing
func TestBackfillRecreatesMissingUniqueIndex(t *testing.T) {
	db := newDB(t)
	if err := db.Exec("DROP INDEX " + UniqueFindingIndex).Error; err != nil {
		t.Fatal(err)
	}
	r := createReport(t, db, "Final", "A", "B")
	res, err := BackfillApprovedReports(db, now)
	if err != nil {
		t.Fatalf("backfill without the index: %v", err)
	}
	if res.CreatedATRs != 2 {
		t.Fatalf("created %d", res.CreatedATRs)
	}
	if !db.Migrator().HasIndex(&models.ActionTakenReport{}, UniqueFindingIndex) {
		t.Fatal("index not recreated")
	}
	if res, _ := BackfillApprovedReports(db, now); res.CreatedATRs != 0 {
		t.Fatalf("second run created %d", res.CreatedATRs)
	}
	if n := len(atrsOf(t, db, r.ID)); n != 2 {
		t.Fatalf("%d ATRs", n)
	}
	// EnsureIndexes is a no-op when everything exists
	if err := EnsureIndexes(db); err != nil {
		t.Fatal(err)
	}
}
