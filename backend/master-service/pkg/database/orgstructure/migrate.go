// Package orgstructure holds the ONLY schema changes of the flexible
// organisation structure (docs/org-structure-model-plan.md 3.3 and 3.4).
//
// Hard rule: a legacy tenant (ORG_STRUCTURE_MODE unset) gets zero DDL from this
// package. The single legacy step re-creates the two global unique indexes that
// used to come from `uniqueIndex` tags, and only when they are missing on a
// database that is not flexible; on every existing tenant they exist, so
// nothing executes.
package orgstructure

import (
	"fmt"

	"master-service/pkg/orgmode"

	"gorm.io/gorm"
)

const (
	// Legacy global unique indexes (names GORM gave the old uniqueIndex tags).
	LegacyNameIndex = "idx_departments_department_name"
	LegacyCodeIndex = "idx_departments_department_code"

	// Flexible replacements.
	NameIndex   = "uq_departments_company_parent_name"
	ParentIndex = "idx_departments_parent"
	ParentFK    = "fk_departments_parent"
	ParentCheck = "chk_departments_parent_not_self"

	ChangeLogTable = "org_change_log"
)

// CodeScope is the department_code uniqueness policy for flexible tenants.
type CodeScope int

const (
	// CodeScopeGlobal: unique among live rows of the whole tenant.
	CodeScopeGlobal CodeScope = iota
	// CodeScopePerCompany: unique among live rows of one company.
	CodeScopePerCompany
)

// DepartmentCodeScope is decision D10 (still open). This constant is the one
// place to change: it drives both the gated index below (CodeIndexName,
// codeIndexDDL) and the service-level duplicate check
// (repositories/department OrgRepository.CodeTaken). Changing it affects only
// tenants migrated in flexible mode afterwards; an existing flexible tenant
// keeps its index (IF NOT EXISTS) until converted by hand.
const DepartmentCodeScope = CodeScopeGlobal

// CodeIndexName is the name of the flexible department_code unique index.
func CodeIndexName() string {
	if DepartmentCodeScope == CodeScopePerCompany {
		return "uq_departments_company_code_live"
	}
	return "uq_departments_code_live"
}

func codeIndexDDL() string {
	if DepartmentCodeScope == CodeScopePerCompany {
		return `CREATE UNIQUE INDEX IF NOT EXISTS uq_departments_company_code_live
	ON departments (company_id, department_code) WHERE deleted_at IS NULL`
	}
	return `CREATE UNIQUE INDEX IF NOT EXISTS uq_departments_code_live
	ON departments (department_code) WHERE deleted_at IS NULL`
}

// FlexibleStatements returns the gated DDL (3.3 steps 1-6), in order. Every
// statement is idempotent. Postgres only.
func FlexibleStatements() []string {
	return []string{
		// 1. org_levels
		`CREATE TABLE IF NOT EXISTS org_levels (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	level integer NOT NULL CONSTRAINT chk_org_levels_level CHECK (level BETWEEN 1 AND 999),
	name varchar(50) NOT NULL,
	name_key varchar(30) NULL CONSTRAINT chk_org_levels_name_key
		CHECK (name_key IS NULL OR name_key IN ('directorate', 'division', 'department', 'section')),
	is_active boolean NOT NULL DEFAULT true,
	created_at timestamptz NOT NULL DEFAULT now(),
	updated_at timestamptz NOT NULL DEFAULT now(),
	deleted_at timestamptz NULL
)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_org_levels_level ON org_levels (level) WHERE deleted_at IS NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_org_levels_name ON org_levels (lower(name)) WHERE deleted_at IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_org_levels_deleted_at ON org_levels (deleted_at)`,

		// 2. org_change_log (append-only)
		`CREATE TABLE IF NOT EXISTS org_change_log (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	action varchar(50) NOT NULL,
	actor_user_id varchar(100) NOT NULL DEFAULT '',
	actor_username varchar(255) NOT NULL DEFAULT '',
	actor_roles jsonb NULL,
	ip varchar(64) NOT NULL DEFAULT '',
	user_agent varchar(500) NOT NULL DEFAULT '',
	base_version varchar(64) NOT NULL DEFAULT '',
	before jsonb NULL,
	after jsonb NULL,
	impact jsonb NULL,
	created_at timestamptz NOT NULL DEFAULT now()
)`,
		`CREATE INDEX IF NOT EXISTS idx_org_change_log_created_at ON org_change_log (created_at DESC)`,

		// 3. departments.parent_department_id, FK, check, index
		`ALTER TABLE departments ADD COLUMN IF NOT EXISTS parent_department_id uuid NULL`,
		`DO $$ BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_departments_parent' AND conrelid = 'departments'::regclass) THEN
		ALTER TABLE departments ADD CONSTRAINT fk_departments_parent
			FOREIGN KEY (parent_department_id) REFERENCES departments (id) ON DELETE RESTRICT;
	END IF;
END $$`,
		`DO $$ BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'chk_departments_parent_not_self' AND conrelid = 'departments'::regclass) THEN
		ALTER TABLE departments ADD CONSTRAINT chk_departments_parent_not_self
			CHECK (parent_department_id IS NULL OR parent_department_id <> id);
	END IF;
END $$`,
		`CREATE INDEX IF NOT EXISTS idx_departments_parent ON departments (company_id, parent_department_id)`,

		// 4-5. name: global unique -> unique per (company, parent), case-insensitive, live rows
		`DROP INDEX IF EXISTS idx_departments_department_name`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_departments_company_parent_name ON departments (
	company_id,
	COALESCE(parent_department_id, '00000000-0000-0000-0000-000000000000'::uuid),
	lower(department_name)
) WHERE deleted_at IS NULL`,

		// 6. code (D10): legacy global index -> partial index per DepartmentCodeScope
		`DROP INDEX IF EXISTS idx_departments_department_code`,
		codeIndexDDL(),
	}
}

// DropStatements removes the flexible-only tables (used by `master fresh` in
// flexible mode, before the regular tables are dropped).
func DropStatements() []string {
	return []string{
		`DROP TABLE IF EXISTS org_change_log`,
		`DROP TABLE IF EXISTS org_levels`,
	}
}

// Migrate runs after the regular AutoMigrate. flagOn is ORG_STRUCTURE_MODE.
//
//   - flag on: the gated DDL, in one transaction.
//   - flag off: only the legacy index step (normally a no-op, see package doc).
func Migrate(db *gorm.DB, flagOn bool) error {
	if flagOn {
		return migrateFlexible(db)
	}
	return ensureLegacyIndexes(db)
}

func migrateFlexible(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, stmt := range FlexibleStatements() {
			if err := tx.Exec(stmt).Error; err != nil {
				return fmt.Errorf("org structure DDL failed: %w\n%s", err, stmt)
			}
		}
		return nil
	})
}

// legacyIndexes are created only when missing on a non-flexible database.
var legacyIndexes = []struct{ name, ddl string }{
	{LegacyCodeIndex, `CREATE UNIQUE INDEX IF NOT EXISTS idx_departments_department_code ON departments (department_code)`},
	{LegacyNameIndex, `CREATE UNIQUE INDEX IF NOT EXISTS idx_departments_department_name ON departments (department_name)`},
}

// ensureLegacyIndexes re-creates the legacy unique indexes on a from-scratch
// legacy database (local dev, CI, a fresh rb_audit). On an existing tenant both
// exist and nothing executes; on a flexible database that lost its flag
// (org_levels present) it never runs, so the index swap is not undone.
func ensureLegacyIndexes(db *gorm.DB) error {
	m := db.Migrator()
	if m.HasTable(orgmode.LevelsTable) {
		return nil
	}
	if !m.HasTable("departments") {
		return nil
	}
	for _, idx := range legacyIndexes {
		if m.HasIndex("departments", idx.name) {
			continue
		}
		if err := db.Exec(idx.ddl).Error; err != nil {
			return fmt.Errorf("create legacy index %s: %w", idx.name, err)
		}
	}
	return nil
}
