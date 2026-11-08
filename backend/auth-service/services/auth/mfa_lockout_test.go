package services

import (
	"context"
	"net/http"
	"testing"

	"auth-service/models"
	apperrors "auth-service/pkg/errors"
)

func asAppErr(t *testing.T, err error) *apperrors.AppError {
	t.Helper()
	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		t.Fatalf("error = %v (%T), want *AppError", err, err)
	}
	return appErr
}

func TestVerifyMFALogin_WrongCodesReportRemainingThenLock(t *testing.T) {
	svc, _, user, rdb := newMFALoginService(t)
	ctx := context.Background()
	user.PasswordHash = mustHash(t, "correct-horse")

	// Each wrong code may use a fresh token: the count is per user.
	for i := 1; i < MFAMaxFailures; i++ {
		_, err := svc.VerifyMFALogin(ctx, &models.VerifyMFALoginRequest{MFAToken: startMFALogin(t, rdb, user.ID), Code: "000000"}, "ip", "UA")
		appErr := asAppErr(t, err)
		if appErr.Code != ErrCodeInvalidMFACode || appErr.StatusCode != http.StatusBadRequest {
			t.Fatalf("attempt %d: %s/%d, want INVALID_CODE/400", i, appErr.Code, appErr.StatusCode)
		}
		if got := appErr.Data["remaining_attempts"]; got != int64(MFAMaxFailures-i) {
			t.Fatalf("attempt %d: remaining_attempts = %v, want %d", i, got, MFAMaxFailures-i)
		}
	}

	token := startMFALogin(t, rdb, user.ID)
	_, err := svc.VerifyMFALogin(ctx, &models.VerifyMFALoginRequest{MFAToken: token, Code: "000000"}, "ip", "UA")
	appErr := asAppErr(t, err)
	if appErr.Code != ErrCodeAccountLocked || appErr.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("last attempt: %s/%d, want ACCOUNT_LOCKED/429", appErr.Code, appErr.StatusCode)
	}
	if appErr.Data["retry_after_seconds"] == nil {
		t.Fatal("lock response has no retry_after_seconds")
	}
	if n, _ := rdb.Exists(ctx, "mfa_login:"+token).Result(); n != 0 {
		t.Fatal("MFA token survived the lock")
	}

	// Locked: a correct code on another token is refused.
	_, err = svc.VerifyMFALogin(ctx, &models.VerifyMFALoginRequest{MFAToken: startMFALogin(t, rdb, user.ID), Code: currentTOTP(t, svc)}, "ip", "UA")
	if asAppErr(t, err).Code != ErrCodeAccountLocked {
		t.Fatalf("correct code during lock: %v, want ACCOUNT_LOCKED", err)
	}

	// Locked: the password step is refused too.
	_, err = svc.Login(ctx, &models.LoginRequest{Username: user.Username, Password: "correct-horse"}, "ip", "UA")
	if asAppErr(t, err).Code != ErrCodeAccountLocked {
		t.Fatalf("password login during lock: %v, want ACCOUNT_LOCKED", err)
	}

	// A wrong password during the lock still reads as wrong credentials.
	_, err = svc.Login(ctx, &models.LoginRequest{Username: user.Username, Password: "nope"}, "ip", "UA")
	if !apperrors.Is(err, apperrors.ErrInvalidCredentials) {
		t.Fatalf("wrong password during lock: %v, want INVALID_CREDENTIALS", err)
	}
}

func TestVerifyMFALogin_CorrectCodeResetsCounter(t *testing.T) {
	svc, _, user, rdb := newMFALoginService(t)
	ctx := context.Background()

	for i := 0; i < MFAMaxFailures-1; i++ {
		_, _ = svc.VerifyMFALogin(ctx, &models.VerifyMFALoginRequest{MFAToken: startMFALogin(t, rdb, user.ID), Code: "000000"}, "ip", "UA")
	}
	if _, err := svc.VerifyMFALogin(ctx, &models.VerifyMFALoginRequest{MFAToken: startMFALogin(t, rdb, user.ID), Code: currentTOTP(t, svc)}, "ip", "UA"); err != nil {
		t.Fatal(err)
	}
	_, err := svc.VerifyMFALogin(ctx, &models.VerifyMFALoginRequest{MFAToken: startMFALogin(t, rdb, user.ID), Code: "000000"}, "ip", "UA")
	if got := asAppErr(t, err).Data["remaining_attempts"]; got != int64(MFAMaxFailures-1) {
		t.Fatalf("remaining_attempts after reset = %v, want %d", got, MFAMaxFailures-1)
	}
}
