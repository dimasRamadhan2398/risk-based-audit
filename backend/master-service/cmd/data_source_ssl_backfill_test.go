package main

import (
	"testing"

	"master-service/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// The data_source_connections table is created by hand: AutoMigrate of the full model
// uses Postgres-only defaults. Only the columns the backfill touches are needed.
func strPtr(s string) *string { return &s }

func TestBackfillDataSourceSSLMode(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`CREATE TABLE data_source_connections (
		id TEXT PRIMARY KEY, ssl BOOLEAN, ssl_mode TEXT, updated_at DATETIME, deleted_at DATETIME)`).Error; err != nil {
		t.Fatalf("create table: %v", err)
	}
	fixtures := []struct {
		id      string
		ssl     bool
		mode    *string // nil = NULL (legacy row)
		deleted bool
	}{
		{"a", true, nil, false},
		{"b", false, nil, false},
		{"c", true, nil, true},
		{"d", false, strPtr("verify-full"), false},
	}
	for _, f := range fixtures {
		var del any
		if f.deleted {
			del = "2026-01-01 00:00:00"
		}
		var mode any
		if f.mode != nil {
			mode = *f.mode
		}
		if err := db.Exec(`INSERT INTO data_source_connections (id, ssl, ssl_mode, deleted_at) VALUES (?, ?, ?, ?)`,
			f.id, f.ssl, mode, del).Error; err != nil {
			t.Fatalf("insert %s: %v", f.id, err)
		}
	}

	if err := backfillDataSourceSSLMode(db); err != nil {
		t.Fatalf("backfill: %v", err)
	}
	// Second run must not change anything.
	if err := backfillDataSourceSSLMode(db); err != nil {
		t.Fatalf("backfill rerun: %v", err)
	}

	want := map[string]string{"a": "require", "b": "disable", "c": "require", "d": "verify-full"}
	for id, mode := range want {
		var got string
		if err := db.Unscoped().Model(&models.DataSourceConnection{}).Select("ssl_mode").Where("id = ?", id).Scan(&got).Error; err != nil {
			t.Fatalf("read %s: %v", id, err)
		}
		if got != mode {
			t.Errorf("row %s: ssl_mode = %q, want %q", id, got, mode)
		}
	}
}
