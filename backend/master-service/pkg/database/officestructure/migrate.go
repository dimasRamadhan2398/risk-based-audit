// Package officestructure holds the schema steps of the office structure that
// GORM tags cannot express (docs/office-structure-plan.md section 2): partial
// unique indexes and the foreign keys of the new columns, plus the idempotent
// seed of the tenant default office levels. It applies to every tenant (D8);
// every statement is idempotent and additive.
package officestructure

import (
	"fmt"

	"master-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Index and constraint names (also registered in pkg/errors/constraints.go).
const (
	CompanyLevelIndex = "uq_office_levels_company_level"
	DefaultLevelIndex = "uq_office_levels_default_level"
	OfficeCodeIndex   = "uq_locations_company_office_code"
	HeadOfficeIndex   = "uq_locations_company_head_office"
)

// IndexStatements are valid on Postgres and SQLite (tests).
//
// company_id NULL is the tenant default scheme. Postgres treats NULLs as
// distinct in a unique index, so the default scheme gets its own index.
func IndexStatements() []string {
	return []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS ` + CompanyLevelIndex + ` ON office_levels (company_id, level)
	WHERE deleted_at IS NULL AND company_id IS NOT NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ` + DefaultLevelIndex + ` ON office_levels (level)
	WHERE deleted_at IS NULL AND company_id IS NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ` + OfficeCodeIndex + ` ON locations (company_id, office_code)
	WHERE deleted_at IS NULL AND office_code IS NOT NULL`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ` + HeadOfficeIndex + ` ON locations (company_id)
	WHERE deleted_at IS NULL AND is_active AND office_function = 'head_office'`,
	}
}

// foreignKeyStatements are Postgres only. The columns are new and NULL on
// existing rows, so adding the constraints never fails on old data.
func foreignKeyStatements() []string {
	fk := func(name, table, column, ref string) string {
		return fmt.Sprintf(`DO $$ BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = '%s' AND conrelid = '%s'::regclass) THEN
		ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (id);
	END IF;
END $$`, name, table, table, name, column, ref)
	}
	return []string{
		fk("fk_locations_company", "locations", "company_id", "companies"),
		fk("fk_locations_parent", "locations", "parent_location_id", "locations"),
		fk("fk_office_levels_company", "office_levels", "company_id", "companies"),
		fk("fk_location_companies_location", "location_companies", "location_id", "locations"),
		fk("fk_location_companies_company", "location_companies", "company_id", "companies"),
	}
}

// Migrate creates office_levels and location_companies (AutoMigrate; the new
// locations columns come from models.Location in the main AutoMigrate list),
// then the indexes, the Postgres foreign keys and the default levels.
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&models.OfficeLevel{}, &models.LocationCompany{}); err != nil {
		return fmt.Errorf("office structure: automigrate: %w", err)
	}
	stmts := IndexStatements()
	if db.Dialector.Name() == "postgres" {
		stmts = append(stmts, foreignKeyStatements()...)
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			return fmt.Errorf("office structure: %w", err)
		}
	}
	return EnsureDefaultLevels(db)
}

// EnsureDefaultLevels seeds the tenant default scheme (10 Kantor Pusat, 20
// Kantor Wilayah, 30 Kantor Cabang) when it has no live level. It never
// touches an existing scheme, so admin edits survive every deploy.
func EnsureDefaultLevels(db *gorm.DB) error {
	var n int64
	if err := db.Model(&models.OfficeLevel{}).Where("company_id IS NULL").Count(&n).Error; err != nil {
		return fmt.Errorf("office structure: count default levels: %w", err)
	}
	if n > 0 {
		return nil
	}
	for _, l := range models.DefaultOfficeLevels() {
		l.ID = uuid.New()
		if err := db.Select("*").Create(&l).Error; err != nil {
			return fmt.Errorf("office structure: seed default level %d: %w", l.Level, err)
		}
	}
	return nil
}
