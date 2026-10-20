package controllers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"auth-service/models"
	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/logger"
	"auth-service/pkg/validations"
	"auth-service/repositories"
	userService "auth-service/services/user"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// createRepo records what CreateUser does. knownRoles lists the role names that
// exist in the roles table; AssignRoles attaches only those.
type createRepo struct {
	repositories.UserRepositoryInterface
	knownRoles map[string]bool
	created    *models.User
	assigned   []string
	deleted    bool
}

func (f *createRepo) FindByUsername(string) (*models.User, error) { return nil, apperrors.ErrNotFound }
func (f *createRepo) FindByEmail(string) (*models.User, error)    { return nil, apperrors.ErrNotFound }
func (f *createRepo) Create(u *models.User) error {
	u.ID = uuid.New()
	f.created = u
	return nil
}
func (f *createRepo) AssignRoles(_ uuid.UUID, names []string) error {
	f.assigned = names
	return nil
}
func (f *createRepo) FindByID(id uuid.UUID) (*models.User, error) {
	u := *f.created
	u.Roles = nil
	for _, n := range f.assigned {
		if f.knownRoles[n] {
			u.Roles = append(u.Roles, models.Role{Name: n})
		}
	}
	return &u, nil
}
func (f *createRepo) Delete(uuid.UUID) error { f.deleted = true; return nil }

func postCreateUser(t *testing.T, repo *createRepo, body string) int {
	t.Helper()
	gin.SetMode(gin.TestMode)
	logger.Log = zap.NewNop()
	ctrl := NewUserController(validations.New(), userService.NewUserService(repo, nil))
	r := gin.New()
	r.POST("/users", ctrl.CreateUser)
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

const createBodyNoRoles = `{"username":"newuser","email":"new@example.com","password":"secret1","roles":[]}`

func TestCreateUserRefusesAccountWithoutRole(t *testing.T) {
	repo := &createRepo{knownRoles: map[string]bool{"AUDITOR": true}}
	if code := postCreateUser(t, repo, createBodyNoRoles); code != http.StatusBadRequest {
		t.Fatalf("no roles: got %d, want 400", code)
	}
	if repo.created != nil {
		t.Fatal("user must not be created when no role is requested")
	}

	repo = &createRepo{knownRoles: map[string]bool{"AUDITOR": true}}
	if code := postCreateUser(t, repo, `{"username":"newuser","email":"new@example.com","password":"secret1","roles":["  "]}`); code != http.StatusBadRequest {
		t.Fatalf("blank role name: got %d, want 400", code)
	}
	if repo.created != nil {
		t.Fatal("user must not be created when the only role is blank")
	}
}

func TestCreateUserAssignsRequestedRoles(t *testing.T) {
	repo := &createRepo{knownRoles: map[string]bool{"AUDITOR": true}}
	body := `{"username":"newuser","email":"new@example.com","password":"secret1","roles":["AUDITOR"]}`
	if code := postCreateUser(t, repo, body); code != http.StatusCreated {
		t.Fatalf("valid roles: got %d, want 201", code)
	}
	if repo.created == nil || !repo.created.IsActive {
		t.Fatal("expected an active user to be created")
	}
	if len(repo.assigned) != 1 || repo.assigned[0] != "AUDITOR" {
		t.Fatalf("roles assigned = %v, want [AUDITOR]", repo.assigned)
	}
	if repo.deleted {
		t.Fatal("a user with a role must not be deleted")
	}
}

func TestCreateUserRemovesAccountWhenNoRequestedRoleExists(t *testing.T) {
	repo := &createRepo{knownRoles: map[string]bool{"AUDITOR": true}}
	body := `{"username":"newuser","email":"new@example.com","password":"secret1","roles":["NO_SUCH_ROLE"]}`
	if code := postCreateUser(t, repo, body); code != http.StatusBadRequest {
		t.Fatalf("unknown role: got %d, want 400", code)
	}
	if !repo.deleted {
		t.Fatal("an account left without any role must be removed")
	}
}
