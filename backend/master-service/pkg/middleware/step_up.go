package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"net/http"
	"strings"
	"time"

	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	"master-service/pkg/logger"
	"master-service/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Step-up tokens (docs/org-structure-model-plan.md 9.3). auth-service issues
// one after POST /auth/verify-password; it is a short-lived HS256 JWT signed
// with a key DERIVED from the stack's jwt.secret, so it can never pass as an
// access token (and an access token never passes as a step-up token).
const (
	StepUpHeader = "X-Step-Up-Token"
	// StepUpKeyLabel is the HMAC message of the key derivation. auth-service
	// must use the same value.
	StepUpKeyLabel = "auditsphere-step-up-v1"
	// StepUpPurposeOrgLevels is the purpose of org level changes.
	StepUpPurposeOrgLevels = "org_levels.change"
	// StepUpTTL is the token lifetime and the one-time jti retention.
	StepUpTTL = 5 * time.Minute
	// stepUpLeeway tolerates clock skew between auth-service and this service.
	stepUpLeeway = 30 * time.Second
)

// DeriveStepUpKey returns HMAC-SHA256(jwtSecret, StepUpKeyLabel).
func DeriveStepUpKey(jwtSecret string) []byte {
	mac := hmac.New(sha256.New, []byte(jwtSecret))
	mac.Write([]byte(StepUpKeyLabel))
	return mac.Sum(nil)
}

// StepUpClaims: sub (user id), purpose, jti, iat, exp.
type StepUpClaims struct {
	Purpose string `json:"purpose"`
	jwt.RegisteredClaims
}

// JTIStore consumes a token id once.
type JTIStore interface {
	// ConsumeOnce returns true the first time jti is seen within ttl.
	ConsumeOnce(ctx context.Context, jti string, ttl time.Duration) (bool, error)
}

// RedisJTIStore keeps used ids in the tenant's own Redis DB (SET NX EX).
type RedisJTIStore struct {
	Client *redis.Client
}

func (s RedisJTIStore) ConsumeOnce(ctx context.Context, jti string, ttl time.Duration) (bool, error) {
	return s.Client.SetNX(ctx, "stepup:"+jti, 1, ttl).Result()
}

var (
	errStepUpRequired = apperrors.New(apperrors.CodeStepUpRequired,
		"Please confirm this change with your password.", http.StatusForbidden)
	errStepUpInvalid = apperrors.New(apperrors.CodeStepUpInvalid,
		"The password confirmation is invalid, expired or already used. Please confirm again.", http.StatusForbidden)
	errStepUpUnavailable = apperrors.New(apperrors.CodeStepUpUnavailable,
		"The password confirmation cannot be checked right now. Please try again.", http.StatusServiceUnavailable)
)

// RequireStepUp verifies the X-Step-Up-Token header: signature with the
// derived key, purpose, expiry, sub equal to the access token's user_id, and
// one-time use of jti. A nil store (Redis unavailable) answers 503. Must run
// after Authenticate.
func (m *AuthMiddleware) RequireStepUp(purpose string, store JTIStore) gin.HandlerFunc {
	key := DeriveStepUpKey(m.secret)
	return func(c *gin.Context) {
		// fail logs the rejection (reason, user ids, jti; never the token).
		claims := &StepUpClaims{}
		fail := func(reason string, err *apperrors.AppError) {
			logStepUp(c, reason, claims, err.StatusCode)
			if err.StatusCode >= http.StatusInternalServerError {
				// RespondError would normalise any 5xx to INTERNAL_ERROR/500.
				response.Error(c, err.StatusCode, err.Code, err.Message, "")
			} else {
				base.RespondError(c, err)
			}
			c.Abort()
		}
		raw := strings.TrimSpace(c.GetHeader(StepUpHeader))
		if raw == "" {
			fail("missing_token", errStepUpRequired)
			return
		}
		token, err := jwt.ParseWithClaims(raw, claims, func(*jwt.Token) (interface{}, error) { return key, nil },
			jwt.WithValidMethods([]string{"HS256"}),
			jwt.WithExpirationRequired(),
			jwt.WithIssuedAt(),
			jwt.WithLeeway(stepUpLeeway),
		)
		if err != nil || !token.Valid {
			reason := "invalid_token"
			switch {
			case errors.Is(err, jwt.ErrTokenExpired):
				reason = "expired"
			case errors.Is(err, jwt.ErrTokenSignatureInvalid):
				reason = "bad_signature"
			}
			*claims = StepUpClaims{} // unverified claims are not logged
			fail(reason, errStepUpInvalid)
			return
		}
		userID := c.GetString("user_id")
		if claims.Purpose != purpose {
			fail("purpose_mismatch", errStepUpInvalid)
			return
		}
		if claims.ID == "" || claims.Subject == "" || userID == "" || claims.Subject != userID {
			fail("subject_mismatch", errStepUpInvalid)
			return
		}
		// Refuse long-lived tokens even when correctly signed.
		if claims.IssuedAt == nil || claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time) > StepUpTTL+stepUpLeeway {
			fail("lifetime_too_long", errStepUpInvalid)
			return
		}
		if store == nil {
			fail("store_unavailable", errStepUpUnavailable)
			return
		}
		first, err := store.ConsumeOnce(c.Request.Context(), claims.ID, StepUpTTL)
		if err != nil {
			fail("store_unavailable", errStepUpUnavailable)
			return
		}
		if !first {
			fail("replayed", errStepUpInvalid)
			return
		}
		logStepUp(c, "", claims, http.StatusOK)
		c.Next()
	}
}

// logStepUp writes one structured line per step-up check: reason "" means
// accepted. It logs the access token's user_id and the verified token's
// sub/jti/purpose, never the token itself.
func logStepUp(c *gin.Context, reason string, claims *StepUpClaims, status int) {
	log := logger.GetLogger()
	if log == nil {
		return
	}
	fields := []zap.Field{
		zap.String("event", "security.step_up_check"),
		zap.String("user_id", c.GetString("user_id")),
		zap.String("username", c.GetString("username")),
		zap.String("path", c.FullPath()),
		zap.String("ip", c.ClientIP()),
		zap.Int("status", status),
	}
	if claims != nil {
		fields = append(fields,
			zap.String("token_sub", boundedLogValue(claims.Subject)),
			zap.String("jti", boundedLogValue(claims.ID)),
			zap.String("purpose", boundedLogValue(claims.Purpose)))
	}
	switch {
	case reason == "":
		log.Info("step-up accepted", fields...)
	case status >= http.StatusInternalServerError:
		log.Error("step-up rejected", append(fields, zap.String("reason", reason))...)
	default:
		log.Warn("step-up rejected", append(fields, zap.String("reason", reason))...)
	}
}

// boundedLogValue caps signed-but-foreign claim values written to the log.
func boundedLogValue(s string) string {
	if len(s) > 100 {
		return s[:100]
	}
	return s
}
