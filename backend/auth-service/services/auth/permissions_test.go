package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"auth-service/models"
	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/utils"
	"auth-service/repositories"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// permTestRepo is a real PermissionRepository over in-memory SQLite with:
// ADMIN: view_system_logs, export_system_logs, view_users
// EXECUTIVE: view_system_logs, view_reports
// VIEWER: view_reports
func permTestRepo(t *testing.T) repositories.IPermissionRepository {
	t.Helper()
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
	for _, p := range []string{"view_system_logs", "export_system_logs", "view_users", "view_reports"} {
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
	grant("EXECUTIVE", "view_system_logs", "view_reports")
	grant("VIEWER", "view_reports")
	grant("AUDITEE") // a role with no permissions
	return repositories.NewPermissionRepository(db)
}

func userWithRoles(roles ...string) *models.User {
	u := claimsTestUser(nil)
	u.Roles = nil
	for _, r := range roles {
		u.Roles = append(u.Roles, models.Role{Name: r})
	}
	return u
}

// loginJSONPermissions marshals the login response the way the controller
// does and returns the raw "permissions" value on user (nil if absent).
func loginJSONPermissions(t *testing.T, resp *models.LoginResponse) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	var out struct {
		User map[string]json.RawMessage `json:"user"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	return out.User["permissions"]
}

func TestCompleteLogin_PermissionsPerRoleSet(t *testing.T) {
	cases := []struct {
		name  string
		roles []string
		want  string
	}{
		{"multiple roles are a sorted, deduped union", []string{"EXECUTIVE", "admin"},
			`["export_system_logs","view_reports","view_system_logs","view_users"]`},
		{"viewer only gets its own", []string{"VIEWER"}, `["view_reports"]`},
		{"role without permissions", []string{"AUDITEE"}, `[]`},
		{"no roles is [] not null", nil, `[]`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, _ := newClaimsTestService(t, http.StatusNotFound, nil)
			svc.permRepo = permTestRepo(t)

			resp, err := svc.completeLogin(context.Background(), userWithRoles(tc.roles...), "", "", "", false)
			if err != nil {
				t.Fatal(err)
			}
			if got := string(loginJSONPermissions(t, resp)); got != tc.want {
				t.Fatalf("permissions = %s, want %s", got, tc.want)
			}
			// Not in the JWT.
			if strings.Contains(resp.Token, "view_") {
				t.Fatal("token mentions a permission")
			}
			claims, err := utils.ParseToken(resp.Token, jwtTestSecret)
			if err != nil {
				t.Fatal(err)
			}
			b, _ := json.Marshal(claims)
			if strings.Contains(string(b), "permission") {
				t.Fatalf("claims carry permissions: %s", b)
			}
		})
	}
}

// noMFARepo reports that the user has no MFA set up.
type noMFARepo struct {
	repositories.MFASetupRepositoryInterface
}

func (noMFARepo) FindByUserID(uuid.UUID) (*models.MFASetup, error) { return nil, apperrors.ErrNotFound }

// passwordLoginRepo serves FindByUsername and Update for the full Login path.
type passwordLoginRepo struct {
	stubUserRepo
	user *models.User
}

func (r passwordLoginRepo) FindByUsername(name string) (*models.User, error) {
	if r.user.Username == name {
		return r.user, nil
	}
	return nil, apperrors.ErrNotFound
}

func TestPasswordLogin_IncludesPermissions(t *testing.T) {
	hash, err := utils.HashPassword("correct-horse")
	if err != nil {
		t.Fatal(err)
	}
	user := userWithRoles("ADMIN", "EXECUTIVE")
	user.IsActive = true
	user.PasswordHash = hash

	svc, _ := newClaimsTestService(t, http.StatusNotFound, nil)
	svc.userRepo = passwordLoginRepo{user: user}
	svc.mfaRepo = noMFARepo{}
	svc.permRepo = permTestRepo(t)

	resp, err := svc.Login(context.Background(), &models.LoginRequest{Username: user.Username, Password: "correct-horse"}, "", "")
	if err != nil {
		t.Fatal(err)
	}
	want := `["export_system_logs","view_reports","view_system_logs","view_users"]`
	if got := string(loginJSONPermissions(t, resp)); got != want {
		t.Fatalf("permissions = %s, want %s", got, want)
	}
}

type failingPermRepo struct{}

func (failingPermRepo) NamesForRoles([]string) ([]string, error) { return nil, errors.New("db down") }

func TestCompleteLogin_PermissionLookupFailureOmitsField(t *testing.T) {
	svc, _ := newClaimsTestService(t, http.StatusNotFound, nil)
	svc.permRepo = failingPermRepo{}

	resp, err := svc.completeLogin(context.Background(), userWithRoles("ADMIN"), "", "", "", false)
	if err != nil {
		t.Fatalf("login must not fail on a permission lookup error: %v", err)
	}
	if raw := loginJSONPermissions(t, resp); raw != nil {
		t.Fatalf("permissions = %s, want field omitted so the frontend probes", raw)
	}
}
