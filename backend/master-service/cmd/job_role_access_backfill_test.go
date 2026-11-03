package main

import (
	"testing"

	"master-service/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestBackfillJobRoleAccessRoles(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`CREATE TABLE job_roles (
		id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))), job_role_code TEXT NOT NULL UNIQUE,
		job_role_name TEXT NOT NULL UNIQUE, job_role_description TEXT, job_position_type TEXT NOT NULL,
		default_access_role TEXT, is_active BOOLEAN DEFAULT TRUE,
		created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`).Error; err != nil {
		t.Fatalf("create table: %v", err)
	}
	rows := []struct{ code, name, pos, existing, want string }{
		{"A", "Director", "DIRECTOR", "", "EXECUTIVE"},
		{"B", "Commissioner", "COMMISSIONER", "", "EXECUTIVE"},
		{"C", "Staff", "STAFF", "", "AUDITEE"},
		{"D", "CAE", "CHIEF_AUDIT_EXECUTIVE", "", "AUDITOR"},
		{"E", "Manager", "MANAGER", "", ""},
		{"F", "Supervisor", "SUPERVISOR", "", ""},
		{"AUD-001", "Internal Auditor", "STAFF", "", "AUDITOR"},
		{"AUD-003", "Audit Manager", "MANAGER", "", "AUDITOR"},
		{"G", "Custom Director", "DIRECTOR", "VIEWER", "VIEWER"}, // never overwritten
	}
	for _, r := range rows {
		if err := db.Exec(`INSERT INTO job_roles (job_role_code, job_role_name, job_position_type, default_access_role) VALUES (?,?,?,NULLIF(?, ''))`,
			r.code, r.name, r.pos, r.existing).Error; err != nil {
			t.Fatalf("insert %s: %v", r.code, err)
		}
	}
	for i := 0; i < 2; i++ { // idempotent
		if err := backfillJobRoleAccessRoles(db); err != nil {
			t.Fatalf("backfill: %v", err)
		}
	}
	for _, r := range rows {
		var got models.JobRole
		if err := db.Where("job_role_code = ?", r.code).First(&got).Error; err != nil {
			t.Fatalf("load %s: %v", r.code, err)
		}
		if got.DefaultAccessRole != r.want {
			t.Errorf("%s: default_access_role = %q, want %q", r.code, got.DefaultAccessRole, r.want)
		}
	}
}
