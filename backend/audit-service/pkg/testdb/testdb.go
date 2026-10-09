// Package testdb opens an in-memory sqlite database with tables built from
// GORM models, for tests. AutoMigrate cannot run on sqlite here because the
// models default their ids to Postgres' gen_random_uuid(), so the tables are
// created from the parsed schema instead (column names and indexes, including
// composite unique indexes, match what AutoMigrate creates on Postgres).
//
// Only test code may import this package.
package testdb

import (
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// New opens a fresh in-memory database and creates a table for every model
func New(t testing.TB, models ...interface{}) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	// One connection: every query sees the same in-memory database
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	for _, m := range models {
		CreateTable(t, db, m)
	}
	return db
}

// CreateTable creates the table of model with its indexes
func CreateTable(t testing.TB, db *gorm.DB, model interface{}) {
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
		case dt == "float":
			typ = "REAL"
		case dt == "time":
			typ = "DATETIME"
		}
		col := "`" + f.DBName + "` " + typ
		if f.PrimaryKey {
			col += " PRIMARY KEY"
		}
		if f.HasDefaultValue && f.DefaultValue != "" && !strings.Contains(f.DefaultValue, "(") {
			col += " DEFAULT " + f.DefaultValue
		}
		cols = append(cols, col)
	}
	if err := db.Exec("CREATE TABLE `" + stmt.Schema.Table + "` (" + strings.Join(cols, ", ") + ")").Error; err != nil {
		t.Fatal(err)
	}
	for _, idx := range stmt.Schema.ParseIndexes() {
		var names []string
		for _, f := range idx.Fields {
			names = append(names, "`"+f.DBName+"`")
		}
		kind := "INDEX"
		if strings.EqualFold(idx.Class, "UNIQUE") {
			kind = "UNIQUE INDEX"
		}
		sql := "CREATE " + kind + " `" + idx.Name + "` ON `" + stmt.Schema.Table + "` (" + strings.Join(names, ", ") + ")"
		if err := db.Exec(sql).Error; err != nil {
			t.Fatal(err)
		}
	}
}
