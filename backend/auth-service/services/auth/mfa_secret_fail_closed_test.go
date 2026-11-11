package services

import (
	"context"
	"net/http"
	"testing"

	"auth-service/models"
	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/utils"
	"auth-service/repositories"

	"github.com/google/uuid"
)

// unreadableMFARepo is an MFA row whose secret cannot be decrypted (missing or
// wrong MFA_ENCRYPTION_KEY, or a tampered value).
type unreadableMFARepo struct {
	repositories.MFASetupRepositoryInterface
}

func (unreadableMFARepo) FindByUserID(uuid.UUID) (*models.MFASetup, error) {
	return nil, repositories.ErrMFASecretUnreadable
}

// TestPasswordLogin_UnreadableMFASecretFailsClosed: before encryption the login
// treated every lookup error as "no MFA" and issued tokens. An unreadable secret
// must stop the login instead of skipping the MFA step.
func TestPasswordLogin_UnreadableMFASecretFailsClosed(t *testing.T) {
	hash, err := utils.HashPassword("correct-horse")
	if err != nil {
		t.Fatal(err)
	}
	user := userWithRoles("ADMIN")
	user.IsActive = true
	user.PasswordHash = hash

	svc, _ := newClaimsTestService(t, http.StatusNotFound, nil)
	svc.userRepo = passwordLoginRepo{user: user}
	svc.mfaRepo = unreadableMFARepo{}
	svc.permRepo = permTestRepo(t)

	resp, err := svc.Login(context.Background(), &models.LoginRequest{Username: user.Username, Password: "correct-horse"}, "", "")
	if err == nil {
		t.Fatalf("login succeeded without MFA: %+v", resp)
	}
	if !apperrors.Is(err, apperrors.ErrInternalServer) {
		t.Fatalf("err = %v, want internal error", err)
	}
	if resp != nil && resp.Token != "" {
		t.Fatal("tokens issued despite an unreadable MFA secret")
	}
}
