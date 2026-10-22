// Package sqlitetest opens a file-backed sqlite database with tables built from
// GORM models. Only test files import it, so the sqlite driver is never linked
// into the service binary.
//
// AutoMigrate cannot run on sqlite because ids default to gen_random_uuid(), so
// each table is created from the parsed schema: column types, the primary key
// (with a sqlite v4-UUID default where the model uses gen_random_uuid()) and
// nothing else. Unique indexes and foreign keys are not created or enforced.
package sqlitetest

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const sqliteUUID = `(lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' ||
	substr(lower(hex(randomblob(2))), 2) || '-' || substr('89ab', 1 + (abs(random()) % 4), 1) ||
	substr(lower(hex(randomblob(2))), 2) || '-' || lower(hex(randomblob(6))))`

// Open returns a database in t.TempDir() with a table for each model.
func Open(t *testing.T, models ...any) *gorm.DB {
	t.Helper()
	dsn := "file:" + filepath.Join(t.TempDir(), "test.db") + "?_pragma=busy_timeout(10000)"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	for _, m := range models {
		CreateTable(t, db, m)
	}
	return db
}

// CreateTable creates the table of model.
func CreateTable(t *testing.T, db *gorm.DB, model any) {
	t.Helper()
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(model); err != nil {
		t.Fatal(err)
	}
	var cols []string
	for _, f := range stmt.Schema.Fields {
		if f.DBName == "" {
			continue
		}
		typ := "TEXT"
		switch dt := strings.ToLower(string(f.DataType)); {
		case strings.Contains(dt, "int"):
			typ = "INTEGER"
		case strings.Contains(dt, "float"), strings.Contains(dt, "decimal"), strings.Contains(dt, "numeric"):
			typ = "REAL"
		case dt == "bool":
			typ = "BOOLEAN"
		case dt == "time":
			typ = "DATETIME"
		}
		col := "`" + f.DBName + "` " + typ
		if f.PrimaryKey {
			col += " PRIMARY KEY"
			if strings.Contains(f.DefaultValue, "gen_random_uuid") {
				col += " DEFAULT " + sqliteUUID
			}
		}
		cols = append(cols, col)
	}
	if err := db.Exec("CREATE TABLE `" + stmt.Schema.Table + "` (" + strings.Join(cols, ", ") + ")").Error; err != nil {
		t.Fatal(err)
	}
}
