package repositories

import (
	"sort"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestNamesForRolesIsCaseInsensitive(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	for _, ddl := range []string{
		`CREATE TABLE permissions (id TEXT PRIMARY KEY, name TEXT)`,
		`CREATE TABLE roles (id TEXT PRIMARY KEY, name TEXT)`,
		`CREATE TABLE role_permissions (role_id TEXT, permission_id TEXT)`,
	} {
		if err := db.Exec(ddl).Error; err != nil {
			t.Fatal(err)
		}
	}
	perm := map[string]string{}
	for _, p := range []string{"view_system_logs", "export_system_logs", "view_users"} {
		perm[p] = uuid.NewString()
		db.Exec(`INSERT INTO permissions (id, name) VALUES (?, ?)`, perm[p], p)
	}
	grant := func(role string, perms ...string) {
		id := uuid.NewString()
		db.Exec(`INSERT INTO roles (id, name) VALUES (?, ?)`, id, role)
		for _, p := range perms {
			db.Exec(`INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)`, id, perm[p])
		}
	}
	grant("ADMIN", "view_system_logs", "export_system_logs", "view_users")
	grant("EXECUTIVE", "view_system_logs", "view_users")
	grant("VIEWER", "view_users")

	repo := NewPermissionRepository(db)
	cases := map[string][]string{
		"admin":            {"export_system_logs", "view_system_logs", "view_users"},
		"Executive,VIEWER": {"view_system_logs", "view_users"},
		"VIEWER":           {"view_users"},
		"NOPE":             {},
	}
	for roles, want := range cases {
		got, err := repo.NamesForRoles(strings.Split(roles, ","))
		if err != nil {
			t.Fatal(err)
		}
		sort.Strings(got)
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Errorf("%s: %v, want %v", roles, got, want)
		}
	}
	if got, _ := repo.NamesForRoles(nil); len(got) != 0 {
		t.Errorf("no roles: %v", got)
	}
}
