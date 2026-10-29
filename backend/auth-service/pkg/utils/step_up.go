package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Step-up tokens (docs/org-structure-model-plan.md 9.3). Issued by
// POST /auth/verify-password and verified by master-service
// pkg/middleware/step_up.go (RequireStepUp). Both sides MUST agree on:
//   - the key: HMAC-SHA256(jwt.secret, StepUpKeyLabel), raw 32 bytes;
//   - HS256;
//   - the claims: sub (access-token user_id), purpose, jti, iat, exp.
//
// The derived key means a step-up token never verifies as an access token
// and an access token never verifies as a step-up token. The test vector in
// step_up_test.go is duplicated in master-service; change both or neither.
const (
	// StepUpKeyLabel is the HMAC message of the key derivation
	// (master-service middleware.StepUpKeyLabel).
	StepUpKeyLabel = "auditsphere-step-up-v1"
	// StepUpPurposeOrgLevels is the only purpose accepted today
	// (master-service middleware.StepUpPurposeOrgLevels).
	StepUpPurposeOrgLevels = "org_levels.change"
	// StepUpTTL is the token lifetime (master-service middleware.StepUpTTL).
	StepUpTTL = 5 * time.Minute
)

// DeriveStepUpKey returns HMAC-SHA256(jwtSecret, StepUpKeyLabel).
func DeriveStepUpKey(jwtSecret string) []byte {
	mac := hmac.New(sha256.New, []byte(jwtSecret))
	mac.Write([]byte(StepUpKeyLabel))
	return mac.Sum(nil)
}

// StepUpClaims has the same JSON shape as master-service middleware.StepUpClaims.
type StepUpClaims struct {
	Purpose string `json:"purpose"`
	jwt.RegisteredClaims
}

// GenerateStepUpToken signs a one-time step-up token for subject (the access
// token's user_id, unchanged) and purpose. It returns the token and its jti.
func GenerateStepUpToken(subject, purpose, jwtSecret string, now time.Time) (string, string, error) {
	jti := uuid.New().String()
	claims := StepUpClaims{
		Purpose: purpose,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(StepUpTTL)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(DeriveStepUpKey(jwtSecret))
	if err != nil {
		return "", "", err
	}
	return signed, jti, nil
}
