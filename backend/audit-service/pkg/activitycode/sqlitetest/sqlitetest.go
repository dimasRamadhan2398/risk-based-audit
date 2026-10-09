// Package sqlitetest opens a file-backed sqlite database with tables built
// from GORM models, for tests of code that writes Activity IDs. Only test
// files import it.
//
// AutoMigrate cannot run on sqlite here because ids default to
// gen_random_uuid(), so tables are created from the parsed schema with a
// sqlite expression that generates a v4 UUID instead.
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

// Open returns a database in t.TempDir(). Transactions start with BEGIN
// IMMEDIATE and wait up to 10s for the write lock, so concurrent writers
// queue instead of failing with SQLITE_BUSY.
func Open(t *testing.T, models ...interface{}) *gorm.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	dsn := "file:" + path + "?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_txlock=immediate"
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

// CreateTable creates the table of model, with its primary key, NOT NULL,
// literal defaults and unique indexes.
func CreateTable(t *testing.T, db *gorm.DB, model interface{}) {
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
		case dt == "bool":
			typ = "BOOLEAN"
		case dt == "time":
			typ = "DATETIME"
		}
		col := "`" + f.DBName + "` " + typ
		switch {
		case f.PrimaryKey && strings.Contains(f.DefaultValue, "gen_random_uuid"):
			col += " PRIMARY KEY DEFAULT " + sqliteUUID
		case f.PrimaryKey:
			col += " PRIMARY KEY"
		case f.DefaultValue != "" && !strings.Contains(f.DefaultValue, "("):
			def := f.DefaultValue
			if typ == "TEXT" {
				def = "'" + strings.Trim(def, "'") + "'"
			}
			col += " DEFAULT " + def
		}
		if f.NotNull {
			col += " NOT NULL"
		}
		if f.Unique {
			col += " UNIQUE"
		}
		cols = append(cols, col)
	}
	ddl := "CREATE TABLE `" + stmt.Schema.Table + "` (" + strings.Join(cols, ", ") + ")"
	if err := db.Exec(ddl).Error; err != nil {
		t.Fatalf("%s: %v", ddl, err)
	}
	for _, idx := range stmt.Schema.ParseIndexes() {
		if idx.Class != "UNIQUE" {
			continue
		}
		var names []string
		for _, o := range idx.Fields {
			names = append(names, "`"+o.DBName+"`")
		}
		q := "CREATE UNIQUE INDEX `" + idx.Name + "` ON `" + stmt.Schema.Table + "` (" + strings.Join(names, ", ") + ")"
		if err := db.Exec(q).Error; err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
}
