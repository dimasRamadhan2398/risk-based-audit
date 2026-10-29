package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"auth-service/models"
	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/logger"
	"auth-service/pkg/utils"
	"auth-service/pkg/validations"
	"auth-service/repositories"
	userService "auth-service/services/user"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// provisionRepo is an in-memory user store for the account provisioning tests.
// It keeps users by ID so FindByID, Update and FindByEmployeeID behave like the
// real repository.
type provisionRepo struct {
	repositories.UserRepositoryInterface
	knownRoles map[string]bool
	users      map[uuid.UUID]*models.User
	assigned   map[uuid.UUID][]string
}

func newProvisionRepo() *provisionRepo {
	return &provisionRepo{
		knownRoles: map[string]bool{"AUDITEE": true},
		users:      map[uuid.UUID]*models.User{},
		assigned:   map[uuid.UUID][]string{},
	}
}

func (f *provisionRepo) FindByUsername(name string) (*models.User, error) {
	for _, u := range f.users {
		if u.Username == name {
			return u, nil
		}
	}
	return nil, apperrors.ErrNotFound
}

func (f *provisionRepo) FindByEmail(email string) (*models.User, error) {
	for _, u := range f.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, apperrors.ErrNotFound
}

func (f *provisionRepo) FindByEmployeeID(code string) (*models.User, error) {
	for _, u := range f.users {
		if u.EmployeeID != nil && *u.EmployeeID == code {
			return u, nil
		}
	}
	return nil, apperrors.ErrNotFound
}

func (f *provisionRepo) Create(u *models.User) error {
	u.ID = uuid.New()
	cp := *u
	f.users[u.ID] = &cp
	return nil
}

func (f *provisionRepo) Update(u *models.User) error {
	cp := *u
	f.users[u.ID] = &cp
	return nil
}

func (f *provisionRepo) Delete(id uuid.UUID) error {
	delete(f.users, id)
	return nil
}

func (f *provisionRepo) AssignRoles(id uuid.UUID, names []string) error {
	f.assigned[id] = names
	return nil
}

func (f *provisionRepo) FindByID(id uuid.UUID) (*models.User, error) {
	u, ok := f.users[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	cp := *u
	cp.Roles = nil
	for _, n := range f.assigned[id] {
		if f.knownRoles[n] {
			cp.Roles = append(cp.Roles, models.Role{Name: n})
		}
	}
	return &cp, nil
}

// onlyUser returns the single stored user, failing the test if there is not exactly one.
func (f *provisionRepo) onlyUser(t *testing.T) *models.User {
	t.Helper()
	if len(f.users) != 1 {
		t.Fatalf("expected exactly one stored user, got %d", len(f.users))
	}
	for _, u := range f.users {
		return u
	}
	return nil
}

func provisionRouter(repo *provisionRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	logger.Log = zap.NewNop()
	ctrl := NewUserController(validations.New(), userService.NewUserService(repo, nil))
	r := gin.New()
	r.POST("/users", ctrl.CreateUser)
	r.POST("/users/:id/link-employee", ctrl.LinkEmployee)
	return r
}

func doProvision(t *testing.T, r *gin.Engine, method, path, body string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var out map[string]any
	if w.Body.Len() > 0 {
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
			t.Fatalf("response is not JSON: %v (%s)", err, w.Body.String())
		}
	}
	return w.Code, out
}

// dataField returns the "data" object of an envelope response.
func dataField(t *testing.T, body map[string]any) map[string]any {
	t.Helper()
	d, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("response has no data object: %v", body)
	}
	return d
}

func TestCreateUserGeneratedPasswordSetsFlagAndLinksEmployee(t *testing.T) {
	repo := newProvisionRepo()
	body := `{"username":"fitri.handayani@aifl.co.id","email":"fitri.handayani@aifl.co.id","generate_password":true,"employee_id":"EMP-0007","roles":["AUDITEE"]}`
	code, resp := doProvision(t, provisionRouter(repo), http.MethodPost, "/users", body)
	if code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (%v)", code, resp)
	}

	data := dataField(t, resp)
	temp, _ := data["temporary_password"].(string)
	if temp == "" {
		t.Fatal("response must carry the generated temporary password once")
	}
	if data["username"] != "fitri.handayani@aifl.co.id" {
		t.Fatalf("username = %v", data["username"])
	}

	u := repo.onlyUser(t)
	if !u.MustChangePassword {
		t.Fatal("generated password must set must_change_password")
	}
	if u.EmployeeID == nil || *u.EmployeeID != "EMP-0007" {
		t.Fatalf("employee_id = %v, want EMP-0007", u.EmployeeID)
	}
	if !utils.CheckPassword(temp, u.PasswordHash) {
		t.Fatal("stored hash must match the returned temporary password")
	}
	if strings.Contains(u.PasswordHash, temp) {
		t.Fatal("password must not be stored in plain text")
	}
}

func TestCreateUserAdminPasswordDoesNotForceChangeOrReturnPassword(t *testing.T) {
	repo := newProvisionRepo()
	body := `{"username":"plain","email":"plain@example.com","password":"secret1","roles":["AUDITEE"]}`
	code, resp := doProvision(t, provisionRouter(repo), http.MethodPost, "/users", body)
	if code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (%v)", code, resp)
	}
	if _, present := dataField(t, resp)["temporary_password"]; present {
		t.Fatal("no temporary password is returned when the admin supplied one")
	}
	u := repo.onlyUser(t)
	if u.MustChangePassword {
		t.Fatal("admin-supplied password must not set must_change_password")
	}
	if u.EmployeeID != nil {
		t.Fatalf("employee_id = %v, want unset", *u.EmployeeID)
	}
}

func TestCreateUserPasswordRules(t *testing.T) {
	cases := map[string]string{
		"both password and generate":    `{"username":"a1","email":"a1@example.com","password":"secret1","generate_password":true,"roles":["AUDITEE"]}`,
		"neither password nor generate": `{"username":"a1","email":"a1@example.com","roles":["AUDITEE"]}`,
		"short admin password":          `{"username":"a1","email":"a1@example.com","password":"abc","roles":["AUDITEE"]}`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			repo := newProvisionRepo()
			if code, _ := doProvision(t, provisionRouter(repo), http.MethodPost, "/users", body); code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", code)
			}
			if len(repo.users) != 0 {
				t.Fatal("no user may be created when the password rule fails")
			}
		})
	}
}

func TestCreateUserRefusesSecondAccountForLinkedEmployee(t *testing.T) {
	repo := newProvisionRepo()
	existing := uuid.New()
	code := "EMP-0007"
	repo.users[existing] = &models.User{ID: existing, Username: "old", Email: "old@example.com", EmployeeID: &code}

	body := `{"username":"new@aifl.co.id","email":"new@aifl.co.id","generate_password":true,"employee_id":"EMP-0007","roles":["AUDITEE"]}`
	status, resp := doProvision(t, provisionRouter(repo), http.MethodPost, "/users", body)
	if status != http.StatusConflict {
		t.Fatalf("status = %d, want 409 (%v)", status, resp)
	}
	if len(repo.users) != 1 {
		t.Fatal("no second account may be created for an employee that is already linked")
	}
}

func TestLinkEmployeeLinksUnlinkedAccount(t *testing.T) {
	repo := newProvisionRepo()
	id := uuid.New()
	repo.users[id] = &models.User{ID: id, Username: "u", Email: "u@aifl.co.id"}

	status, _ := doProvision(t, provisionRouter(repo), http.MethodPost, "/users/"+id.String()+"/link-employee", `{"employee_code":" EMP-0009 "}`)
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	if repo.users[id].EmployeeID == nil || *repo.users[id].EmployeeID != "EMP-0009" {
		t.Fatalf("employee_id = %v, want EMP-0009", repo.users[id].EmployeeID)
	}
}

func TestLinkEmployeeRules(t *testing.T) {
	repo := newProvisionRepo()
	linked, other := uuid.New(), uuid.New()
	code9, code10 := "EMP-0009", "EMP-0010"
	repo.users[linked] = &models.User{ID: linked, Username: "linked", Email: "l@aifl.co.id", EmployeeID: &code9}
	repo.users[other] = &models.User{ID: other, Username: "other", Email: "o@aifl.co.id", EmployeeID: &code10}
	free := uuid.New()
	repo.users[free] = &models.User{ID: free, Username: "free", Email: "f@aifl.co.id"}
	r := provisionRouter(repo)

	// Same code again is a no-op.
	if status, _ := doProvision(t, r, http.MethodPost, "/users/"+linked.String()+"/link-employee", `{"employee_code":"EMP-0009"}`); status != http.StatusOK {
		t.Fatalf("re-link same code: status = %d, want 200", status)
	}
	// Account already linked to a different employee.
	if status, _ := doProvision(t, r, http.MethodPost, "/users/"+linked.String()+"/link-employee", `{"employee_code":"EMP-0011"}`); status != http.StatusConflict {
		t.Fatalf("account linked elsewhere: status = %d, want 409", status)
	}
	// Employee already has another account.
	if status, _ := doProvision(t, r, http.MethodPost, "/users/"+free.String()+"/link-employee", `{"employee_code":"EMP-0010"}`); status != http.StatusConflict {
		t.Fatalf("employee linked elsewhere: status = %d, want 409", status)
	}
	if repo.users[free].EmployeeID != nil {
		t.Fatal("a rejected link must not change the account")
	}
	// Unknown account.
	if status, _ := doProvision(t, r, http.MethodPost, "/users/"+uuid.New().String()+"/link-employee", `{"employee_code":"EMP-0012"}`); status != http.StatusNotFound {
		t.Fatalf("unknown user: status = %d, want 404", status)
	}
	// Missing code.
	if status, _ := doProvision(t, r, http.MethodPost, "/users/"+free.String()+"/link-employee", `{"employee_code":"  "}`); status != http.StatusBadRequest {
		t.Fatalf("blank code: status = %d, want 400", status)
	}
}
