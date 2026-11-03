package seeders

import (
	"testing"

	"master-service/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSeedJobRolesIsAdditiveAndValid(t *testing.T) {
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
	// A tenant-edited row with a seeded code, and a custom row squatting on a seeded name.
	db.Exec(`INSERT INTO job_roles (job_role_code, job_role_name, job_position_type, default_access_role) VALUES ('AUD-001','Renamed Auditor','STAFF','VIEWER')`)
	db.Exec(`INSERT INTO job_roles (job_role_code, job_role_name, job_position_type) VALUES ('X-1','Intern','STAFF')`)

	s := &Seeder{DB: db}
	for i := 0; i < 2; i++ {
		if err := s.SeedJobRoles(); err != nil {
			t.Fatalf("seed run %d: %v", i, err)
		}
	}

	var edited models.JobRole
	db.Where("job_role_code = ?", "AUD-001").First(&edited)
	if edited.JobRoleName != "Renamed Auditor" || edited.DefaultAccessRole != "VIEWER" {
		t.Errorf("existing row was overwritten: %+v", edited)
	}
	var squat models.JobRole
	db.Where("job_role_name = ?", "Intern").First(&squat)
	if squat.JobRoleCode != "X-1" || squat.DefaultAccessRole != "" {
		t.Errorf("name-squatting row was modified: %+v", squat)
	}
	var all []models.JobRole
	db.Find(&all)
	byName := map[string]string{}
	for _, r := range all {
		if _, ok := models.NormalizeAccessRole(r.DefaultAccessRole); !ok {
			t.Errorf("%s has invalid access role %q", r.JobRoleCode, r.DefaultAccessRole)
		}
		byName[r.JobRoleName] = r.DefaultAccessRole
	}
	for name, want := range map[string]string{
		"CEO": "EXECUTIVE", "Senior Manager": "EXECUTIVE", "General Manager": "EXECUTIVE",
		"Commissioner": "EXECUTIVE", "Chief Audit Executive": "AUDITOR", "IT Auditor": "AUDITOR",
		"Audit Manager": "AUDITOR", "Junior Staff": "AUDITEE", "Risk Officer": "AUDITEE",
	} {
		if byName[name] != want {
			t.Errorf("%s: default_access_role = %q, want %q", name, byName[name], want)
		}
	}
}
