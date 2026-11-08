package main

import (
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Medium rows scored under the old rule become audit priorities; nothing else
// changes, and a second run is a no-op.
func TestPrioritizeMediumUniverseYears(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:"+filepath.Join(t.TempDir(), "m.db")), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE audit_universe_years (id TEXT PRIMARY KEY, year INTEGER,
		risk_index REAL, risk_level TEXT, audit_priority BOOLEAN)`).Error; err != nil {
		t.Fatal(err)
	}
	rows := []struct {
		id, level string
		index     float64
		priority  bool
	}{
		{"high", "High", 85, true},
		{"mh", "Medium to High", 65, true},
		{"med", "Medium", 45, false},
		{"lm", "Low to Medium", 25, false},
		{"low", "Low", 10, false},
		{"unscored", "", 0, false},
	}
	for _, r := range rows {
		db.Exec(`INSERT INTO audit_universe_years (id, year, risk_index, risk_level, audit_priority) VALUES (?, 2026, ?, ?, ?)`, r.id, r.index, r.level, r.priority)
	}

	for run := 1; run <= 2; run++ {
		if err := prioritizeMediumUniverseYears(db); err != nil {
			t.Fatal(err)
		}
		want := map[string]bool{"high": true, "mh": true, "med": true, "lm": false, "low": false, "unscored": false}
		for id, w := range want {
			var got bool
			db.Raw(`SELECT audit_priority FROM audit_universe_years WHERE id = ?`, id).Scan(&got)
			if got != w {
				t.Fatalf("run %d: %s audit_priority = %v, want %v", run, id, got, w)
			}
		}
		var medIndex float64
		db.Raw(`SELECT risk_index FROM audit_universe_years WHERE id = 'med'`).Scan(&medIndex)
		if medIndex != 45 {
			t.Fatalf("run %d: risk_index changed to %v", run, medIndex)
		}
	}
}
