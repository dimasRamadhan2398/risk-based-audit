package main

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/config"
	"master-service/pkg/database/seeders"
	"master-service/pkg/kafka"
	deptRepo "master-service/repositories/department"
	levelRepo "master-service/repositories/org_level"
	deptSvc "master-service/services/department"
	orgLevelSvc "master-service/services/org_level"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Real-Postgres checks of the gated DDL (docs/org-structure-model-plan.md
// section 14, checks 1, 2, 4, 6). Skipped unless MASTER_TEST_PG_DSN points to
// a THROWAWAY server (e.g. "host=127.0.0.1 port=55432 user=postgres
// dbname=postgres sslmode=disable"). Each test creates and drops its own
// database. Never point this at a tenant or production server.

func pgDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("MASTER_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("MASTER_TEST_PG_DSN not set")
	}
	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	name := "ms_orgtest_" + strings.ReplaceAll(uuid.NewString()[:8], "-", "")
	if err := admin.Exec("CREATE DATABASE " + name).Error; err != nil {
		t.Fatal(err)
	}
	db, err := gorm.Open(postgres.Open(dsn+" dbname="+name), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
		admin.Exec("DROP DATABASE IF EXISTS " + name)
		if sqlDB, err := admin.DB(); err == nil {
			sqlDB.Close()
		}
	})
	return db
}

func withMode(t *testing.T, mode string) {
	t.Helper()
	prev := cfg
	cfg = &config.Config{OrgStructure: config.OrgStructureConfig{Mode: mode}}
	t.Cleanup(func() { cfg = prev })
}

// schema is a comparable description of the public schema.
func schema(t *testing.T, db *gorm.DB) map[string][]string {
	t.Helper()
	q := map[string]string{
		"columns": `SELECT table_name||'.'||column_name||' '||data_type||' '||is_nullable||' '||coalesce(column_default,'')
			FROM information_schema.columns WHERE table_schema='public' ORDER BY 1`,
		"indexes":     `SELECT indexname||' '||indexdef FROM pg_indexes WHERE schemaname='public' ORDER BY 1`,
		"constraints": `SELECT conrelid::regclass||' '||conname||' '||pg_get_constraintdef(oid) FROM pg_constraint WHERE connamespace='public'::regnamespace ORDER BY 1`,
	}
	out := map[string][]string{}
	for k, sql := range q {
		var rows []string
		if err := db.Raw(sql).Scan(&rows).Error; err != nil {
			t.Fatal(err)
		}
		out[k] = rows
	}
	return out
}

func hasIndex(db *gorm.DB, name string) bool { return db.Migrator().HasIndex("departments", name) }

func TestPGLegacyMigrateIsStableAndKeepsGlobalIndexes(t *testing.T) {
	db := pgDB(t)
	withMode(t, "")
	if err := migrateAll(db); err != nil {
		t.Fatal(err)
	}
	first := schema(t, db)
	for i := 0; i < 2; i++ {
		if err := migrateAll(db); err != nil {
			t.Fatal(err)
		}
		if again := schema(t, db); !reflect.DeepEqual(first, again) {
			t.Fatalf("legacy re-migrate changed the schema:\nbefore %v\nafter  %v", first, again)
		}
	}
	if !hasIndex(db, "idx_departments_department_name") || !hasIndex(db, "idx_departments_department_code") {
		t.Fatal("from-scratch legacy DB lacks the global unique indexes")
	}
	if db.Migrator().HasTable("org_levels") || db.Migrator().HasTable("org_change_log") ||
		db.Migrator().HasColumn("departments", "parent_department_id") {
		t.Fatal("flexible schema on a legacy DB")
	}
}

func TestPGFlexibleMigrateSeedAndApply(t *testing.T) {
	db := pgDB(t)
	withMode(t, "flexible")
	for i := 0; i < 2; i++ { // idempotent
		if err := migrateAll(db); err != nil {
			t.Fatalf("flexible migrate run %d: %v", i, err)
		}
	}
	m := db.Migrator()
	if !m.HasTable("org_levels") || !m.HasTable("org_change_log") || !m.HasColumn("departments", "parent_department_id") {
		t.Fatal("flexible schema missing")
	}
	if hasIndex(db, "idx_departments_department_name") || hasIndex(db, "idx_departments_department_code") {
		t.Fatal("legacy global indexes still present on a flexible DB")
	}
	for _, idx := range []string{"uq_departments_company_parent_name", "uq_departments_code_live", "idx_departments_parent"} {
		if !hasIndex(db, idx) {
			t.Fatalf("%s missing", idx)
		}
	}

	// seed-org-levels (bank), twice.
	for i := 0; i < 2; i++ {
		if _, err := orgLevelSvc.SeedPreset(levelRepo.New(db), orgLevelSvc.PresetBank, nil); err != nil {
			t.Fatal(err)
		}
	}
	var n int64
	db.Model(&models.OrgLevel{}).Count(&n)
	if n != 4 {
		t.Fatalf("want 4 levels, got %d", n)
	}

	// Demo seeders, preset-aware, twice.
	for i := 0; i < 2; i++ {
		if err := seeders.NewSeeder(db).WithOrgStructure(true).RunAll(); err != nil {
			t.Fatalf("seed run %d: %v", i, err)
		}
	}
	var level1 int64
	db.Model(&models.Department{}).Where("level = 1").Count(&level1)
	if level1 != 0 {
		t.Fatalf("%d departments at Level 1", level1)
	}
	svc := deptSvc.NewOrgDepartmentService(deptRepo.NewDepartmentRepository(db), deptRepo.NewOrgRepository(db), nil)
	tree, err := svc.Tree(base.NewBaseService(), nil)
	if err != nil {
		t.Fatal(err)
	}
	paths := map[string]string{}
	var walk func(nodes []*models.DepartmentTreeNode)
	walk = func(nodes []*models.DepartmentTreeNode) {
		for _, nd := range nodes {
			var names []string
			for _, s := range nd.Path {
				names = append(names, s.Name)
			}
			paths[nd.DepartmentCode] = strings.Join(names, " › ")
			walk(nd.Children)
		}
	}
	walk(tree)
	for code, want := range map[string]string{
		"DEPT-003":     "Direktorat Utama › Internal Audit",
		"DEPT-PJK-AKT": "Direktorat Keuangan › Divisi Akuntansi › Departemen Pajak",
		"DEPT-PJK-TRS": "Direktorat Keuangan › Divisi Treasury › Departemen Pajak",
		"SKS-INFRA":    "Direktorat Operasional › Divisi Teknologi Informasi › Information Technology › Seksi Infrastruktur",
		"DEPT-004":     "Direktorat Operasional › Operations",
	} {
		if paths[code] != want {
			t.Fatalf("%s path = %q, want %q", code, paths[code], want)
		}
	}

	// DB-level guard: same name under the same parent is refused by the index.
	var akt models.Department
	db.Where("department_code = ?", "DIV-AKT").First(&akt)
	dup := uuid.New()
	err = db.Exec(`INSERT INTO departments (id, department_code, department_name, level, company_id, parent_department_id, is_active, created_at, updated_at)
		VALUES (?, 'X-DUP', 'DEPARTEMEN PAJAK', 30, ?, ?, true, now(), now())`, dup, akt.CompanyID, akt.ID).Error
	if err == nil || !strings.Contains(err.Error(), "uq_departments_company_parent_name") {
		t.Fatalf("duplicate leaf under one parent not refused by the DB: %v", err)
	}

	// Renumber 10/20/30/40 -> 100/200/300/400 on Postgres (park trick, jsonb log, FOR UPDATE).
	if err := db.AutoMigrate(&models.KafkaOutboxEvent{}); err != nil {
		t.Fatal(err)
	}
	ls := orgLevelSvc.NewService(db, kafka.NewDomainPublisher(db, "test"))
	cur, _ := ls.List(context.Background())
	var in []orgLevelSvc.LevelInput
	for i, l := range cur {
		id := l.ID
		in = append(in, orgLevelSvc.LevelInput{ID: &id, Level: (i + 1) * 100, Name: l.Name})
	}
	// Swap two names too, to exercise the unique name index.
	in[0].Name, in[1].Name = cur[1].Name, cur[0].Name
	req := orgLevelSvc.ChangeRequest{Levels: in}
	pv, err := ls.Preview(context.Background(), req)
	if err != nil || !pv.CanApply {
		t.Fatalf("preview: %+v %v", pv, err)
	}
	res, err := ls.Apply(context.Background(), orgLevelSvc.ApplyRequest{ChangeRequest: req, BaseVersion: pv.BaseVersion},
		orgLevelSvc.Actor{UserID: "u1", Username: "admin", Roles: []string{"ADMIN"}, IP: "127.0.0.1"})
	if err != nil {
		t.Fatalf("apply on postgres: %v", err)
	}
	var levels []int
	db.Model(&models.Department{}).Distinct().Order("level").Pluck("level", &levels)
	if fmt.Sprint(levels) != "[100 200 300 400]" {
		t.Fatalf("departments not renumbered: %v", levels)
	}
	var logs []models.OrgChangeLog
	if err := db.Find(&logs).Error; err != nil || len(logs) != 1 || !strings.Contains(strings.ReplaceAll(string(logs[0].Before), " ", ""), `"level":10`) {
		t.Fatalf("change log: %v %+v", err, logs)
	}
	if res.EventsQueued == 0 {
		t.Fatal("no department events queued")
	}
}
