// Package servicetoken issues the short-lived JWT audit-service presents to
// master-service and risk-service when there is no user request to forward a
// token from (the annual snapshot reconcile, at start-up or from the CLI).
//
// Why a self-signed token works, and what it can do:
//
//   - master-service (pkg/middleware/auth.go) and risk-service
//     (middleware/auth.go) only verify the token: HS256 signature with the
//     stack's jwt.secret, and a present, unexpired "exp". Neither looks the
//     user up, checks a session / blacklist / token_version, or needs a
//     tenant, company or MFA claim.
//   - The endpoints the reconcile reads (GET /api/v1/departments,
//     /api/v1/audit-universe/corporate, /api/v1/audit-universe/year/:year)
//     sit behind Authenticate only, with no RequireRoles, and their handlers
//     do not read the caller's identity.
//
// So the token carries no role at all: every route guarded by RequireRoles
// (or a role check in a handler) refuses it. It lives DefaultTTL, names a
// non-human subject, and is signed with this deployment's own secret, so a
// tenant's token is only valid inside that tenant's stack.
//
// Never log the token or the Authorization header built from it.
package servicetoken

import (
	"fmt"
	"strings"
	"time"

	"audit-service/pkg/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	// Subject is the non-human identity of the token (sub, user_id, username).
	Subject = "system:audit-service"
	// Issuer is the iss claim.
	Issuer = "audit-service"
	// DefaultTTL is how long a signed token is valid. One reconcile reads its
	// sources well within this; each attempt signs a new token.
	DefaultTTL = 5 * time.Minute
	// OverrideEnv may hold a user JWT to present instead of the self-signed
	// one (CLI only). "Bearer " prefix optional.
	OverrideEnv = "SNAPSHOT_SYNC_TOKEN"

	SourceOverride   = "override"
	SourceSelfSigned = "self-signed"
)

// Audience lists the services the token is meant for. master-service and
// risk-service do not check aud today; it documents intent.
var Audience = []string{"master-service", "risk-service"}

// Claims has the json shape auth-service issues (user_id, username, roles), so
// every service's middleware decodes it.
type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

// Sign returns a token for Subject valid from now for ttl, signed HS256 with
// secret. It refuses a secret the services themselves would refuse.
func Sign(secret string, now time.Time, ttl time.Duration) (string, error) {
	if err := config.ValidateJWTSecret(secret); err != nil {
		return "", fmt.Errorf("cannot sign service token: %w", err)
	}
	if ttl <= 0 {
		return "", fmt.Errorf("cannot sign service token: ttl must be positive")
	}
	now = now.UTC()
	claims := Claims{
		UserID:   Subject,
		Username: Subject,
		Roles:    []string{},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   Subject,
			Issuer:    Issuer,
			Audience:  Audience,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        uuid.NewString(),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// Bearer turns a raw token (with or without a "Bearer" prefix, any case) into
// the header value the services accept, which is exactly "Bearer <jwt>". A JWT
// has no spaces, so anything else is "" (no usable token).
func Bearer(raw string) string {
	fields := strings.Fields(raw)
	if len(fields) > 0 && strings.EqualFold(fields[0], "bearer") {
		fields = fields[1:]
	}
	if len(fields) != 1 {
		return ""
	}
	return "Bearer " + fields[0]
}

// Authorization returns the Authorization header for one reconcile: the
// override token when override holds one, else a freshly self-signed token.
// source is SourceOverride or SourceSelfSigned (safe to log; the header is not).
func Authorization(override, secret string, now time.Time) (header, source string, err error) {
	if h := Bearer(override); h != "" {
		return h, SourceOverride, nil
	}
	if strings.TrimSpace(override) != "" {
		return "", "", fmt.Errorf("%s is set but does not hold a token", OverrideEnv)
	}
	tok, err := Sign(secret, now, DefaultTTL)
	if err != nil {
		return "", "", err
	}
	return "Bearer " + tok, SourceSelfSigned, nil
}
