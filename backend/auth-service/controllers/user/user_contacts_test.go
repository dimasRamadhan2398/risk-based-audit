package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"auth-service/models"
	"auth-service/pkg/logger"
	"auth-service/pkg/validations"
	"auth-service/repositories"
	userService "auth-service/services/user"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// contactsRepo records the FindContacts arguments
type contactsRepo struct {
	repositories.UserRepositoryInterface
	ids   []uuid.UUID
	dept  string
	calls int
	users []*models.User
}

func (r *contactsRepo) FindContacts(ids []uuid.UUID, dept string) ([]*models.User, error) {
	r.calls++
	r.ids, r.dept = ids, dept
	return r.users, nil
}

func contactsRouter(repo *contactsRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	logger.Log = zap.NewNop()
	ctrl := NewUserController(validations.New(), userService.NewUserService(repo, nil))
	r := gin.New()
	r.GET("/users/contacts", ctrl.ListUserContacts)
	return r
}

func getContacts(t *testing.T, r *gin.Engine, query string) (int, map[string]any) {
	t.Helper()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/users/contacts"+query, nil))
	var body map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	return w.Code, body
}

func TestListUserContactsReturnsEmailAndNoSecrets(t *testing.T) {
	id := uuid.New()
	repo := &contactsRepo{users: []*models.User{{
		ID: id, FullName: "Budi", Email: "budi@x.id", Department: "Finance",
		Username: "budi", Phone: "0812", PasswordHash: "secret",
	}}}
	code, body := getContacts(t, contactsRouter(repo), "?ids="+id.String())
	if code != http.StatusOK {
		t.Fatalf("status = %d, body = %v", code, body)
	}
	users := body["data"].(map[string]any)["users"].([]any)
	if len(users) != 1 {
		t.Fatalf("users = %v", users)
	}
	u := users[0].(map[string]any)
	if u["email"] != "budi@x.id" || u["full_name"] != "Budi" || u["department"] != "Finance" {
		t.Fatalf("unexpected contact: %v", u)
	}
	for _, k := range []string{"username", "phone", "password_hash", "roles"} {
		if _, ok := u[k]; ok {
			t.Fatalf("contact leaks %q: %v", k, u)
		}
	}
	if len(repo.ids) != 1 || repo.ids[0] != id {
		t.Fatalf("ids passed = %v", repo.ids)
	}
}

func TestListUserContactsDepartmentHead(t *testing.T) {
	repo := &contactsRepo{}
	code, _ := getContacts(t, contactsRouter(repo), "?department_head_of=%20Finance%20")
	if code != http.StatusOK {
		t.Fatalf("status = %d", code)
	}
	if repo.dept != "Finance" {
		t.Fatalf("department passed = %q", repo.dept)
	}
}

func TestListUserContactsRejectsNoFilter(t *testing.T) {
	repo := &contactsRepo{}
	code, _ := getContacts(t, contactsRouter(repo), "")
	if code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", code)
	}
	if repo.calls != 0 {
		t.Fatal("repository must not be queried without a filter")
	}
}

func TestListUserContactsRejectsBadAndTooManyIDs(t *testing.T) {
	repo := &contactsRepo{}
	r := contactsRouter(repo)
	if code, _ := getContacts(t, r, "?ids=not-a-uuid"); code != http.StatusBadRequest {
		t.Fatalf("bad id status = %d", code)
	}
	ids := ""
	for i := 0; i < 21; i++ {
		if i > 0 {
			ids += ","
		}
		ids += uuid.New().String()
	}
	if code, _ := getContacts(t, r, "?ids="+ids); code != http.StatusBadRequest {
		t.Fatalf("too many ids status = %d", code)
	}
	if repo.calls != 0 {
		t.Fatal("repository must not be queried for invalid input")
	}
}
