package services

import (
	"context"
	"net/http"
	"time"

	"auth-service/pkg/errors"
	"auth-service/pkg/utils"

	"github.com/google/uuid"
)

// MFA sign-in lockout. Wrong one-time codes are counted per user (not per MFA
// login token), so signing in again for a fresh token does not reset the count.
// The MFAMaxFailures-th wrong code locks the account for MFALockDuration: the
// MFA step and the password step are both refused until the lock expires.
const (
	MFAMaxFailures = 3
	// MFALockDuration is both the lock length and the window the failure
	// counter lives for, so stale failures do not pile up across days.
	MFALockDuration = 15 * time.Minute

	mfaFailKeyPrefix = "auth:mfa_fail:"
	mfaLockKeyPrefix = "auth:mfa_lock:"
)

// Error codes the frontend matches on.
const (
	ErrCodeInvalidMFACode = "INVALID_CODE"
	ErrCodeAccountLocked  = "ACCOUNT_LOCKED"
)

func mfaFailKey(userID uuid.UUID) string { return mfaFailKeyPrefix + userID.String() }
func mfaLockKey(userID uuid.UUID) string { return mfaLockKeyPrefix + userID.String() }

// errAccountLocked is the 429 sent while a lock is active. retry_after_seconds
// lets the client say how long to wait.
func errAccountLocked(retryAfter time.Duration) *errors.AppError {
	secs := int(retryAfter.Round(time.Second) / time.Second)
	if secs < 1 {
		secs = 1
	}
	return errors.New(ErrCodeAccountLocked,
		"Too many incorrect verification codes. Your account is locked for 15 minutes.",
		http.StatusTooManyRequests,
	).WithData(map[string]interface{}{
		"retry_after_seconds": secs,
		"remaining_attempts":  0,
		"max_attempts":        MFAMaxFailures,
	})
}

// mfaLockRemaining returns how long the user's MFA lock still runs (0 when not
// locked). A Redis error is returned so callers can fail closed.
func (s *AuthService) mfaLockRemaining(ctx context.Context, userID uuid.UUID) (time.Duration, error) {
	ttl, err := s.redis.TTL(ctx, mfaLockKey(userID)).Result()
	if err != nil {
		return 0, err
	}
	// go-redis reports "no expiry" as -1 and "no key" as -2. A lock without
	// an expiry is never written here; treat it as a full lock, not as none.
	if ttl == -1 {
		return MFALockDuration, nil
	}
	if ttl <= 0 {
		return 0, nil
	}
	return ttl, nil
}

// checkMFALock returns errAccountLocked while the user is locked, and
// ErrInternalServer when the lock state cannot be read (fail closed).
func (s *AuthService) checkMFALock(ctx context.Context, userID uuid.UUID) error {
	remaining, err := s.mfaLockRemaining(ctx, userID)
	if err != nil {
		s.LogError("MFA lock lookup failed", utils.LogField("user_id", userID), utils.LogField("error", err.Error()))
		return errors.ErrInternalServer
	}
	if remaining > 0 {
		return errAccountLocked(remaining)
	}
	return nil
}

// registerMFAFailure counts a wrong code and returns the error to send: an
// INVALID_CODE carrying the attempts left, or ACCOUNT_LOCKED once the limit is
// reached (the counter is then replaced by the lock key).
func (s *AuthService) registerMFAFailure(ctx context.Context, userID uuid.UUID) error {
	failKey := mfaFailKey(userID)

	count, err := s.redis.Incr(ctx, failKey).Result()
	if err != nil {
		s.LogError("MFA failure counter not updated", utils.LogField("user_id", userID), utils.LogField("error", err.Error()))
		return errors.ErrInternalServer
	}
	// Expire only when the key has no TTL yet, so the window starts at the
	// first failure (EXPIRE NX, Redis 7). Retried on every failure, so a
	// missed call cannot leave a counter that never expires.
	if err := s.redis.ExpireNX(ctx, failKey, MFALockDuration).Err(); err != nil {
		s.LogWarning("MFA failure counter expiry not set", utils.LogField("user_id", userID), utils.LogField("error", err.Error()))
	}

	if count >= MFAMaxFailures {
		if err := s.redis.Set(ctx, mfaLockKey(userID), "1", MFALockDuration).Err(); err != nil {
			s.LogError("MFA lock not written", utils.LogField("user_id", userID), utils.LogField("error", err.Error()))
			return errors.ErrInternalServer
		}
		s.redis.Del(ctx, failKey)
		s.LogWarning("Account locked after too many wrong MFA codes",
			utils.LogField("user_id", userID), utils.LogField("failed_attempts", count))
		return errAccountLocked(MFALockDuration)
	}

	return errors.New(ErrCodeInvalidMFACode, "Invalid MFA code", http.StatusBadRequest).
		WithData(map[string]interface{}{
			"failed_attempts":    count,
			"remaining_attempts": MFAMaxFailures - count,
			"max_attempts":       MFAMaxFailures,
		})
}

// clearMFAFailures resets the counter after a correct code. Failure is only
// logged: the sign-in itself succeeded.
func (s *AuthService) clearMFAFailures(ctx context.Context, userID uuid.UUID) {
	if err := s.redis.Del(ctx, mfaFailKey(userID)).Err(); err != nil {
		s.LogWarning("MFA failure counter not cleared", utils.LogField("user_id", userID), utils.LogField("error", err.Error()))
	}
}
