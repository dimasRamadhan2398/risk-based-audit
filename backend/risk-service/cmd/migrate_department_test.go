package main

import (
	"sync"
	"testing"

	"risk-service/models"

	"gorm.io/gorm/schema"
)

// autoMigrate (cmd/migrate.go) already lists CorporateAuditUniverse, so the new
// column ships through the normal AutoMigrate: ALTER TABLE ... ADD COLUMN
// "department_id" uuid plus an index, no default and no NOT NULL, which leaves
// existing rows NULL. AutoMigrate cannot run on sqlite (gen_random_uuid
// defaults), so this pins the schema AutoMigrate would apply instead. It never
// touches a real or tenant database.
func TestCorporateAuditUniverseDepartmentIDSchema(t *testing.T) {
	s, err := schema.Parse(&models.CorporateAuditUniverse{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	f := s.LookUpField("department_id")
	if f == nil {
		t.Fatal("department_id column is not part of the model")
	}
	if f.NotNull || f.PrimaryKey || f.HasDefaultValue {
		t.Errorf("department_id must be plain nullable with no default: notnull=%v pk=%v default=%v", f.NotNull, f.PrimaryKey, f.HasDefaultValue)
	}
	indexed := false
	for _, idx := range s.ParseIndexes() {
		for _, of := range idx.Fields {
			if of.DBName == "department_id" {
				indexed = true
			}
		}
	}
	if !indexed {
		t.Error("department_id must be indexed")
	}
	if n := s.LookUpField("department_name"); n != nil && n.DBName != "" {
		t.Error("department_name must not be a stored column")
	}
}
