// Package orgtest builds an in-memory sqlite database shaped like a FLEXIBLE
// tenant (departments with parent_department_id, org_levels, org_change_log,
// outbox) for tests. Test helper only; never imported by service code.
package orgtest

import (
	"testing"

	"master-service/models"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// sqlite equivalents of the gated Postgres DDL (pkg/database/orgstructure),
// including the partial unique indexes, plus the minimal referenced tables.
var ddl = []string{
	`CREATE TABLE companies (id TEXT PRIMARY KEY, company_code TEXT, company_name TEXT, deleted_at DATETIME)`,
	`CREATE TABLE business_units (id TEXT PRIMARY KEY, company_id TEXT, deleted_at DATETIME)`,
	`CREATE TABLE employees (id TEXT PRIMARY KEY, company_id TEXT, department_id TEXT, deleted_at DATETIME)`,
	`CREATE TABLE departments (
		id TEXT PRIMARY KEY, department_code TEXT NOT NULL, department_name TEXT NOT NULL,
		department_description TEXT, pic_id TEXT, level INTEGER NOT NULL, is_active NUMERIC DEFAULT 1,
		created_at DATETIME, updated_at DATETIME, deleted_at DATETIME,
		company_id TEXT NOT NULL, business_unit_id TEXT,
		parent_department_id TEXT NULL REFERENCES departments(id) ON DELETE RESTRICT,
		CHECK (parent_department_id IS NULL OR parent_department_id <> id))`,
	`CREATE UNIQUE INDEX uq_departments_company_parent_name ON departments (company_id,
		COALESCE(parent_department_id, '00000000-0000-0000-0000-000000000000'), lower(department_name)) WHERE deleted_at IS NULL`,
	`CREATE UNIQUE INDEX uq_departments_code_live ON departments (department_code) WHERE deleted_at IS NULL`,
	`CREATE TABLE org_levels (id TEXT PRIMARY KEY, level INTEGER NOT NULL CHECK (level BETWEEN 1 AND 999),
		name TEXT NOT NULL, name_key TEXT NULL, is_active NUMERIC NOT NULL DEFAULT 1,
		created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
	`CREATE UNIQUE INDEX uq_org_levels_level ON org_levels (level) WHERE deleted_at IS NULL`,
	`CREATE UNIQUE INDEX uq_org_levels_name ON org_levels (lower(name)) WHERE deleted_at IS NULL`,
	`CREATE TABLE org_change_log (id TEXT PRIMARY KEY, action TEXT NOT NULL, actor_user_id TEXT, actor_username TEXT,
		actor_roles TEXT, ip TEXT, user_agent TEXT, base_version TEXT, before TEXT, after TEXT, impact TEXT, created_at DATETIME)`,
}

// DB opens the flexible-shaped sqlite database.
func DB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatal(err)
	}
	for _, stmt := range ddl {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("ddl: %v\n%s", err, stmt)
		}
	}
	if err := db.AutoMigrate(&models.KafkaOutboxEvent{}); err != nil {
		t.Fatal(err)
	}
	return db
}

// Company inserts a company and returns its id.
func Company(t *testing.T, db *gorm.DB) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if err := db.Exec(`INSERT INTO companies (id, company_code, company_name) VALUES (?, ?, ?)`, id, id.String()[:8], "Co "+id.String()[:4]).Error; err != nil {
		t.Fatal(err)
	}
	return id
}

// Level inserts an org level.
func Level(t *testing.T, db *gorm.DB, level int, name string, key string) models.OrgLevel {
	t.Helper()
	l := models.OrgLevel{ID: uuid.New(), Level: level, Name: name, IsActive: true}
	if key != "" {
		l.NameKey = &key
	}
	if err := db.Create(&l).Error; err != nil {
		t.Fatal(err)
	}
	return l
}

// BankLevels inserts 10 Direktorat, 20 Divisi, 30 Departemen, 40 Seksi.
func BankLevels(t *testing.T, db *gorm.DB) []models.OrgLevel {
	return []models.OrgLevel{
		Level(t, db, 10, "Direktorat", models.OrgLevelKeyDirectorate),
		Level(t, db, 20, "Divisi", models.OrgLevelKeyDivision),
		Level(t, db, 30, "Departemen", models.OrgLevelKeyDepartment),
		Level(t, db, 40, "Seksi", models.OrgLevelKeySection),
	}
}

// Employee inserts an employee in a department.
func Employee(t *testing.T, db *gorm.DB, companyID, departmentID uuid.UUID) {
	t.Helper()
	if err := db.Exec(`INSERT INTO employees (id, company_id, department_id) VALUES (?, ?, ?)`, uuid.New(), companyID, departmentID).Error; err != nil {
		t.Fatal(err)
	}
}

// Outbox returns the queued outbox rows in order.
func Outbox(t *testing.T, db *gorm.DB) []models.KafkaOutboxEvent {
	t.Helper()
	var rows []models.KafkaOutboxEvent
	if err := db.Order("id ASC").Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	return rows
}
