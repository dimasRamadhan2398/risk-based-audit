package action_taken_report

import (
	"fmt"

	"audit-service/models"

	"gorm.io/gorm"
)

// UniqueFindingIndex is the unique index SyncFromReport's
// INSERT ... ON CONFLICT (audit_result_report_id, finding_id) relies on. It
// is a plain unique index (no WHERE predicate), so Postgres infers it for
// that conflict target; soft-deleted rows keep their slot on purpose.
const UniqueFindingIndex = "idx_atr_report_finding"

// EnsureIndexes creates every index of the ATR model that is missing, in
// particular UniqueFindingIndex. AutoMigrate creates them, but a table
// rebuilt by a column drop (sqlite rebuilds the table; a failed or partial
// earlier migration may leave it out) must not reach an ON CONFLICT insert
// without it, or every LHA approval fails. Idempotent.
func EnsureIndexes(db *gorm.DB) error {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(&models.ActionTakenReport{}); err != nil {
		return err
	}
	m := db.Migrator()
	for _, idx := range stmt.Schema.ParseIndexes() {
		if m.HasIndex(&models.ActionTakenReport{}, idx.Name) {
			continue
		}
		if err := m.CreateIndex(&models.ActionTakenReport{}, idx.Name); err != nil {
			return fmt.Errorf("create index %s: %w", idx.Name, err)
		}
	}
	return nil
}
