package main

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// The pre-deploy table (no scored_at) gets the column and every existing row is
// marked scored; a second migrate must not touch rows that are unscored by then.
func TestAddAuditUniverseScoredAtBackfillsOnce(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+filepath.Join(t.TempDir(), "m.db")), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE audit_universe_years (id TEXT PRIMARY KEY, corporate_audit_universe_id TEXT, year INTEGER,
		risk_index REAL, risk_level TEXT, audit_priority BOOLEAN, created_at DATETIME, updated_at DATETIME)`).Error; err != nil {
		t.Fatal(err)
	}
	updated := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	created := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	db.Exec(`INSERT INTO audit_universe_years (id, year, risk_index, risk_level, audit_priority, created_at, updated_at) VALUES (?, 2026, 60, 'Medium to High', 1, ?, ?)`, "a", created, updated)
	db.Exec(`INSERT INTO audit_universe_years (id, year, risk_index, risk_level, audit_priority, created_at, updated_at) VALUES (?, 2026, 60, 'Medium to High', 1, ?, NULL)`, "b", created)

	if err := addAuditUniverseScoredAt(db); err != nil {
		t.Fatal(err)
	}
	var unscored int64
	db.Raw(`SELECT COUNT(*) FROM audit_universe_years WHERE scored_at IS NULL`).Scan(&unscored)
	if unscored != 0 {
		t.Fatalf("%d existing rows left unscored after the backfill", unscored)
	}
	// Compared in SQL: sqlite does not parse a column declared timestamptz.
	var matching int64
	db.Raw(`SELECT COUNT(*) FROM audit_universe_years WHERE (id = 'a' AND scored_at = updated_at) OR (id = 'b' AND scored_at = created_at)`).Scan(&matching)
	if matching != 2 {
		t.Fatalf("scored_at should be updated_at, falling back to created_at; %d of 2 rows match", matching)
	}

	// A row established after the deploy is unscored and must stay that way.
	db.Exec(`INSERT INTO audit_universe_years (id, year, risk_index, risk_level, audit_priority, created_at, updated_at, scored_at) VALUES ('c', 2027, 0, '', 0, ?, ?, NULL)`, updated, updated)
	if err := addAuditUniverseScoredAt(db); err != nil {
		t.Fatal(err)
	}
	db.Raw(`SELECT COUNT(*) FROM audit_universe_years WHERE scored_at IS NULL`).Scan(&unscored)
	if unscored != 1 {
		t.Fatalf("second migrate changed unscored rows: %d still unscored, want 1", unscored)
	}
}
