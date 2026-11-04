package services

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"testing"

	"auth-service/models"
	"auth-service/pkg/activitylog"
	baseService "auth-service/pkg/base"
	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/logger"
	"auth-service/pkg/utils"
	"auth-service/repositories"

	"go.uber.org/zap"
)

type captureRecorder struct {
	mu      sync.Mutex
	entries []activitylog.Entry
}

func (c *captureRecorder) Record(_ context.Context, e activitylog.Entry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = append(c.entries, e)
}

// loginUserRepo answers FindByUsername from a fixed user.
type loginUserRepo struct {
	repositories.UserRepositoryInterface
	user *models.User
}

func (r loginUserRepo) FindByUsername(name string) (*models.User, error) {
	if r.user != nil && r.user.Username == name {
		return r.user, nil
	}
	return nil, apperrors.ErrNotFound
}

func TestLoginFailuresAreRecordedWithoutPassword(t *testing.T) {
	hash, err := utils.HashPassword("correct-horse")
	if err != nil {
		t.Fatal(err)
	}
	active := claimsTestUser(nil)
	active.IsActive = true
	active.PasswordHash = hash

	rec := &captureRecorder{}
	if logger.Log == nil {
		logger.Log = zap.NewNop()
	}
	svc := &AuthService{BaseService: baseService.NewBaseService(), userRepo: loginUserRepo{user: active}, activity: rec}

	attempts := []struct {
		username, password, reason string
	}{
		{"ghost", "s3cret-typed-1", loginFailUnknownUser},
		{active.Username, "s3cret-typed-2", loginFailInvalidPassword},
	}
	for _, a := range attempts {
		if _, err := svc.Login(context.Background(), &models.LoginRequest{Username: a.username, Password: a.password}, "203.0.113.4", "UA"); err == nil {
			t.Fatalf("%s: login succeeded", a.username)
		}
	}
	if len(rec.entries) != 2 {
		t.Fatalf("%d entries, want 2", len(rec.entries))
	}
	for i, e := range rec.entries {
		if e.Action != activitylog.ActionLoginFailed || e.Result != activitylog.ResultFailure ||
			e.ActorUsername != attempts[i].username || e.Details["reason"] != attempts[i].reason || e.IPAddress != "203.0.113.4" {
			t.Errorf("entry %d = %+v", i, e)
		}
		b, _ := json.Marshal(e)
		if strings.Contains(string(b), attempts[i].password) {
			t.Errorf("entry %d contains the typed password: %s", i, b)
		}
	}
	if rec.entries[0].ActorUserID != "" || rec.entries[1].ActorUserID != active.ID.String() {
		t.Errorf("actor ids = %q, %q", rec.entries[0].ActorUserID, rec.entries[1].ActorUserID)
	}
}

func TestLoginSuccessIsRecordedWithoutToken(t *testing.T) {
	svc, _ := newClaimsTestService(t, http.StatusNotFound, nil)
	rec := &captureRecorder{}
	svc.activity = rec

	resp, err := svc.completeLogin(context.Background(), claimsTestUser(nil), "", "198.51.100.1", "UA", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.entries) != 1 {
		t.Fatalf("%d entries", len(rec.entries))
	}
	e := rec.entries[0]
	if e.Action != activitylog.ActionLogin || e.Result != activitylog.ResultSuccess || e.ActorUsername != "auditor1" || e.Details["new_device"] != true {
		t.Fatalf("entry = %+v", e)
	}
	b, _ := json.Marshal(e)
	if strings.Contains(string(b), resp.Token) {
		t.Fatal("login entry contains the JWT")
	}
}
