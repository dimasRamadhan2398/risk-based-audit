package seeders

import (
	"testing"

	"auth-service/models"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestAdminSeedFromEnvDefaults(t *testing.T) {
	t.Setenv("ADMIN_NAME", "")
	t.Setenv("ADMIN_EMAIL", "  ")
	t.Setenv("ADMIN_PASSWORD", "")
	a := AdminSeedFromEnv()
	if a.Name != DefaultAdminName || a.Email != DefaultAdminEmail || a.Password != DefaultAdminPassword || a.PasswordFromEnv {
		t.Fatalf("defaults = %+v", a)
	}
}

func TestAdminSeedFromEnvOverrides(t *testing.T) {
	t.Setenv("ADMIN_NAME", " Siti CAE ")
	t.Setenv("ADMIN_EMAIL", "cae@bai.co.id")
	t.Setenv("ADMIN_PASSWORD", " s3cret pw ")
	a := AdminSeedFromEnv()
	if a.Name != "Siti CAE" || a.Email != "cae@bai.co.id" || a.Password != " s3cret pw " || !a.PasswordFromEnv {
		t.Fatalf("overrides = %+v", a)
	}
}

func TestSeedDemoUsersSwitch(t *testing.T) {
	cases := []struct {
		demo, adminPw string
		want          bool
	}{
		{"", "", true},           // dev/demo compose stacks: unchanged
		{"", "random-pw", false}, // provisioned tenant, empty data
		{"true", "random-pw", true},
		{"false", "", false},
	}
	for _, tc := range cases {
		t.Setenv("SEED_DEMO_USERS", tc.demo)
		t.Setenv("ADMIN_PASSWORD", tc.adminPw)
		if got := SeedDemoUsers(); got != tc.want {
			t.Errorf("SEED_DEMO_USERS=%q ADMIN_PASSWORD set=%v: got %v", tc.demo, tc.adminPw != "", got)
		}
	}
}

func seedDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: gormlogger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	// gen_random_uuid() does not exist in SQLite: tables by hand, ids in a callback.
	for _, ddl := range []string{
		`CREATE TABLE users (id TEXT PRIMARY KEY, employee_id TEXT UNIQUE, username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL, password_hash TEXT, full_name TEXT, phone TEXT, department TEXT,
			position TEXT, is_active BOOLEAN DEFAULT 1, avatar_url TEXT, must_change_password BOOLEAN DEFAULT 0,
			locked_until DATETIME, last_login_fingerprint TEXT, last_login_at DATETIME,
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE roles (id TEXT PRIMARY KEY, name TEXT UNIQUE, description TEXT, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE user_roles (user_id TEXT, role_id TEXT, PRIMARY KEY (user_id, role_id))`,
	} {
		if err := db.Exec(ddl).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Callback().Create().Before("gorm:create").Register("test:user_id", func(tx *gorm.DB) {
		if u, ok := tx.Statement.Dest.(*models.User); ok && u.ID == uuid.Nil {
			u.ID = uuid.New()
		}
	}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"ADMIN", "AUDITOR", "EXECUTIVE", "AUDITEE", "VIEWER", "DEPARTMENT_HEAD"} {
		if err := db.Exec(`INSERT INTO roles (id, name) VALUES (?, ?)`, uuid.NewString(), name).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func TestSeedUsersProvisionedTenant(t *testing.T) {
	t.Setenv("ADMIN_NAME", "Siti CAE")
	t.Setenv("ADMIN_EMAIL", "cae@bai.co.id")
	t.Setenv("ADMIN_PASSWORD", "R4ndom-Initial-Pw")
	t.Setenv("SEED_DEMO_USERS", "")
	db := seedDB(t)
	s := NewSeeder(db)
	for i := 0; i < 2; i++ { // idempotent
		if err := s.SeedUsers(); err != nil {
			t.Fatal(err)
		}
	}
	var users []models.User
	db.Find(&users)
	if len(users) != 1 || users[0].Username != "admin" || users[0].Email != "cae@bai.co.id" || users[0].FullName != "Siti CAE" {
		t.Fatalf("users = %+v", users)
	}
	if bcrypt.CompareHashAndPassword([]byte(users[0].PasswordHash), []byte("R4ndom-Initial-Pw")) != nil {
		t.Error("admin hash does not match ADMIN_PASSWORD")
	}
	var links int64
	db.Table("user_roles").Count(&links)
	if links != 1 {
		t.Errorf("role links = %d", links)
	}

	// A later seed with other values leaves the existing admin alone.
	t.Setenv("ADMIN_PASSWORD", "changed")
	t.Setenv("ADMIN_EMAIL", "other@bai.co.id")
	if err := s.SeedUsers(); err != nil {
		t.Fatal(err)
	}
	var admin models.User
	db.Where("username = ?", "admin").First(&admin)
	if admin.Email != "cae@bai.co.id" || bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte("R4ndom-Initial-Pw")) != nil {
		t.Error("existing admin was modified")
	}
}

func TestSeedUsersDemoStack(t *testing.T) {
	t.Setenv("ADMIN_NAME", "")
	t.Setenv("ADMIN_EMAIL", "")
	t.Setenv("ADMIN_PASSWORD", "")
	t.Setenv("SEED_DEMO_USERS", "")
	db := seedDB(t)
	if err := NewSeeder(db).SeedUsers(); err != nil {
		t.Fatal(err)
	}
	var n int64
	db.Model(&models.User{}).Count(&n)
	if n != 6 {
		t.Fatalf("users = %d, want 6", n)
	}
}
