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

// assignableRepo records the FindAssignable arguments
type assignableRepo struct {
	repositories.UserRepositoryInterface
	offset, limit int
	search        string
	id            *uuid.UUID
	users         []*models.User
	total         int64
}

func (r *assignableRepo) FindAssignable(offset, limit int, search string, id *uuid.UUID) ([]*models.User, int64, error) {
	r.offset, r.limit, r.search, r.id = offset, limit, search, id
	return r.users, r.total, nil
}

func assignableRouter(repo *assignableRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	logger.Log = zap.NewNop()
	ctrl := NewUserController(validations.New(), userService.NewUserService(repo, nil))
	r := gin.New()
	r.GET("/users/assignable", ctrl.ListAssignableUsers)
	return r
}

func getAssignable(t *testing.T, r *gin.Engine, query string) (int, map[string]any) {
	t.Helper()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/users/assignable"+query, nil))
	var body map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	return w.Code, body
}

func TestListAssignableUsersShape(t *testing.T) {
	id := uuid.New()
	avatar := "data:image/png;base64,xx"
	repo := &assignableRepo{
		users: []*models.User{{
			ID: id, FullName: "Budi", Department: "Finance", Position: "Manager",
			// Present on the model but must never be returned
			Email: "budi@x.id", Username: "budi", Phone: "0812", AvatarURL: &avatar,
			Roles: []models.Role{{Name: "ADMIN"}},
		}},
		total: 41,
	}
	code, body := getAssignable(t, assignableRouter(repo), "?search=bud&page=3&page_size=20")
	if code != http.StatusOK || body["success"] != true {
		t.Fatalf("status %d body %v", code, body)
	}
	data := body["data"].(map[string]any)
	users := data["users"].([]any)
	if len(users) != 1 {
		t.Fatalf("users = %v", users)
	}
	u := users[0].(map[string]any)
	if len(u) != 4 || u["id"] != id.String() || u["full_name"] != "Budi" || u["department"] != "Finance" || u["position"] != "Manager" {
		t.Fatalf("user = %v, want exactly id/full_name/department/position", u)
	}
	p := data["pagination"].(map[string]any)
	if p["page"] != float64(3) || p["page_size"] != float64(20) || p["total"] != float64(41) || p["total_pages"] != float64(3) {
		t.Fatalf("pagination = %v", p)
	}
	if repo.offset != 40 || repo.limit != 20 || repo.search != "bud" || repo.id != nil {
		t.Fatalf("repo called with offset %d limit %d search %q id %v", repo.offset, repo.limit, repo.search, repo.id)
	}
}

func TestListAssignableUsersPaging(t *testing.T) {
	cases := []struct {
		query          string
		offset, limit  int
		page, pageSize float64
	}{
		{"", 0, 20, 1, 20},
		{"?page_size=500", 0, 100, 1, 100},
		{"?page_size=100", 0, 100, 1, 100},
		{"?page_size=0", 0, 20, 1, 20},
		{"?page_size=-5", 0, 20, 1, 20},
		{"?page_size=abc", 0, 20, 1, 20},
		{"?page=0", 0, 20, 1, 20},
		{"?page=-1", 0, 20, 1, 20},
		{"?page=x", 0, 20, 1, 20},
		{"?page=2&page_size=5", 5, 5, 2, 5},
	}
	for _, tc := range cases {
		repo := &assignableRepo{}
		code, body := getAssignable(t, assignableRouter(repo), tc.query)
		if code != http.StatusOK {
			t.Fatalf("%s: status %d", tc.query, code)
		}
		if repo.offset != tc.offset || repo.limit != tc.limit {
			t.Errorf("%s: offset %d limit %d, want %d %d", tc.query, repo.offset, repo.limit, tc.offset, tc.limit)
		}
		p := body["data"].(map[string]any)["pagination"].(map[string]any)
		if p["page"] != tc.page || p["page_size"] != tc.pageSize || p["total"] != float64(0) || p["total_pages"] != float64(0) {
			t.Errorf("%s: pagination %v", tc.query, p)
		}
		// An empty page is an empty array, not null
		if users, ok := body["data"].(map[string]any)["users"].([]any); !ok || len(users) != 0 {
			t.Errorf("%s: users = %v, want []", tc.query, body["data"].(map[string]any)["users"])
		}
	}
}

func TestListAssignableUsersByID(t *testing.T) {
	id := uuid.New()
	repo := &assignableRepo{}
	if code, _ := getAssignable(t, assignableRouter(repo), "?id="+id.String()); code != http.StatusOK {
		t.Fatalf("status %d", code)
	}
	if repo.id == nil || *repo.id != id {
		t.Fatalf("id filter not passed: %v", repo.id)
	}
	if code, _ := getAssignable(t, assignableRouter(&assignableRepo{}), "?id=not-a-uuid"); code != http.StatusBadRequest {
		t.Fatalf("invalid id: status %d, want 400", code)
	}
}
