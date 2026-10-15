package controllers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

// fakeRepo is an in-memory UserRepositoryInterface; unimplemented methods panic via the nil embed.
type fakeRepo struct {
	repositories.UserRepositoryInterface
	user *models.User
}

func (f *fakeRepo) FindByID(id uuid.UUID) (*models.User, error) {
	cp := *f.user
	return &cp, nil
}
func (f *fakeRepo) Update(u *models.User) error { cp := *u; f.user = &cp; return nil }

const pngB64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg=="

func pngURL() string { return "data:image/png;base64," + pngB64 }

func setup(t *testing.T, roles []string, existing *string) (*gin.Engine, *fakeRepo, uuid.UUID) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	logger.Log = zap.NewNop()
	id := uuid.New()
	repo := &fakeRepo{user: &models.User{
		ID: id, Username: "u", Email: "u@x.com", FullName: "Old Name", Phone: "123",
		Department: "IT", Position: "Dev", IsActive: true, AvatarURL: existing,
	}}
	ctrl := NewUserController(validations.New(), userService.NewUserService(repo, nil))
	r := gin.New()
	r.PUT("/users/:id", func(c *gin.Context) { c.Set("roles", roles) }, ctrl.UpdateUser)
	return r, repo, id
}

func put(r *gin.Engine, id uuid.UUID, body string) (*httptest.ResponseRecorder, map[string]any) {
	req := httptest.NewRequest(http.MethodPut, "/users/"+id.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var out map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	return w, out
}

func jsonBody(t *testing.T, m map[string]any) string {
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestUpdateUserAvatar(t *testing.T) {
	old := pngURL()
	oversize := "data:image/png;base64," + strings.Repeat("A", 300000)
	tests := []struct {
		name       string
		existing   *string
		body       string
		wantStatus int
		wantAvatar string // expected response avatar_url when 200
	}{
		{"set valid png", nil, jsonBody(t, map[string]any{"avatar_url": pngURL()}), 200, pngURL()},
		{"set valid jpeg", nil, jsonBody(t, map[string]any{"avatar_url": "data:image/jpeg;base64," + pngB64}), 200, "data:image/jpeg;base64," + pngB64},
		{"set valid webp", nil, jsonBody(t, map[string]any{"avatar_url": "data:image/webp;base64," + pngB64}), 200, "data:image/webp;base64," + pngB64},
		{"clear with empty string", &old, `{"avatar_url":""}`, 200, ""},
		{"absent leaves unchanged", &old, `{"full_name":"New"}`, 200, old},
		{"null leaves unchanged", &old, `{"avatar_url":null}`, 200, old},
		{"svg rejected", nil, `{"avatar_url":"data:image/svg+xml;base64,` + pngB64 + `"}`, 400, ""},
		{"gif rejected", nil, `{"avatar_url":"data:image/gif;base64,` + pngB64 + `"}`, 400, ""},
		{"text/html rejected", nil, `{"avatar_url":"data:text/html;base64,` + pngB64 + `"}`, 400, ""},
		{"not a data url", nil, `{"avatar_url":"https://evil.example/a.png"}`, 400, ""},
		{"invalid base64 chars", nil, `{"avatar_url":"data:image/png;base64,@@@@"}`, 400, ""},
		{"invalid base64 padding", nil, `{"avatar_url":"data:image/png;base64,abcde"}`, 400, ""},
		{"empty payload", nil, `{"avatar_url":"data:image/png;base64,"}`, 400, ""},
		{"oversize", nil, jsonBody(t, map[string]any{"avatar_url": oversize}), 400, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, repo, id := setup(t, []string{"auditor"}, tc.existing)
			w, out := put(r, id, tc.body)
			if w.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d: %s", w.Code, tc.wantStatus, w.Body.String())
			}
			if tc.wantStatus == 400 {
				e, _ := out["error"].(map[string]any)
				if e["code"] != "INVALID_AVATAR" || e["message"] == "" {
					t.Fatalf("unexpected error body: %v", out)
				}
				if repo.user.AvatarURL != tc.existing && (repo.user.AvatarURL == nil || tc.existing == nil || *repo.user.AvatarURL != *tc.existing) {
					t.Fatal("avatar changed despite 400")
				}
				return
			}
			data := out["data"].(map[string]any)
			if data["avatar_url"] != tc.wantAvatar {
				t.Fatalf("response avatar_url mismatch: %v", data["avatar_url"])
			}
			if (tc.wantAvatar == "") != (repo.user.AvatarURL == nil) {
				t.Fatalf("stored avatar mismatch: %v", repo.user.AvatarURL)
			}
		})
	}
}

func TestUpdateUserPartialDoesNotBlank(t *testing.T) {
	r, repo, id := setup(t, []string{"auditor"}, nil)
	w, out := put(r, id, `{"position":"Lead"}`)
	if w.Code != 200 {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}
	u := repo.user
	if u.FullName != "Old Name" || u.Phone != "123" || u.Department != "IT" || u.Position != "Lead" || !u.IsActive {
		t.Fatalf("partial update altered other fields: %+v", u)
	}
	data := out["data"].(map[string]any)
	for _, k := range []string{"id", "username", "email", "full_name", "phone", "department", "position", "is_active", "avatar_url", "roles"} {
		if _, ok := data[k]; !ok {
			t.Errorf("response missing %q", k)
		}
	}
	if out["message"] != "User updated successfully" {
		t.Errorf("message = %v", out["message"])
	}
}

func TestUpdateUserIsActiveAdminOnly(t *testing.T) {
	r, repo, id := setup(t, []string{"auditor"}, nil)
	if w, _ := put(r, id, `{"is_active":false}`); w.Code != http.StatusForbidden {
		t.Fatalf("non-admin status = %d, want 403", w.Code)
	}
	if !repo.user.IsActive {
		t.Fatal("non-admin flipped is_active")
	}

	r, repo, id = setup(t, []string{"Admin"}, nil) // case-insensitive
	if w, _ := put(r, id, `{"is_active":false}`); w.Code != 200 {
		t.Fatalf("admin status = %d", w.Code)
	}
	if repo.user.IsActive {
		t.Fatal("admin could not deactivate")
	}
}

func TestValidateAvatarDataURLRoundTrip(t *testing.T) {
	if _, err := base64.StdEncoding.DecodeString(pngB64); err != nil {
		t.Fatal(err)
	}
	if err := userService.ValidateAvatarDataURL(pngURL()); err != nil {
		t.Fatal(err)
	}
}
