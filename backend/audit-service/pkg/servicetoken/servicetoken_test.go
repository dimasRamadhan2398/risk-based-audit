package servicetoken

import (
	"errors"
	"strings"
	"testing"
	"time"

	"audit-service/pkg/middleware"

	"github.com/golang-jwt/jwt/v5"
)

// testSecret is synthetic and only used inside these tests.
var testSecret = "test-" + strings.Repeat("z", 64)

// parseLikeServices validates the way master-service and risk-service do:
// HS256 pinned, exp required (same options as their Authenticate).
func parseLikeServices(t *testing.T, tok, secret string, now time.Time) (*Claims, error) {
	t.Helper()
	parsed, err := jwt.ParseWithClaims(tok, &Claims{},
		func(*jwt.Token) (interface{}, error) { return []byte(secret), nil },
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(func() time.Time { return now }),
	)
	if err != nil {
		return nil, err
	}
	return parsed.Claims.(*Claims), nil
}

func TestSignClaims(t *testing.T) {
	now := time.Date(2026, 10, 27, 8, 0, 0, 0, time.UTC)
	tok, err := Sign(testSecret, now, DefaultTTL)
	if err != nil {
		t.Fatal(err)
	}
	c, err := parseLikeServices(t, tok, testSecret, now.Add(time.Minute))
	if err != nil {
		t.Fatalf("token rejected by service-style validation: %v", err)
	}
	if c.Subject != Subject || c.UserID != Subject || c.Username != Subject || c.Issuer != Issuer {
		t.Fatalf("identity claims = sub %q user_id %q username %q iss %q", c.Subject, c.UserID, c.Username, c.Issuer)
	}
	if len(c.Roles) != 0 {
		t.Fatalf("roles = %v, want none (least privilege)", c.Roles)
	}
	if c.ID == "" {
		t.Fatal("jti missing")
	}
	if got := c.ExpiresAt.Time.Sub(c.IssuedAt.Time); got != DefaultTTL {
		t.Fatalf("lifetime = %v, want %v", got, DefaultTTL)
	}
	if DefaultTTL > 10*time.Minute {
		t.Fatalf("DefaultTTL %v is not short-lived", DefaultTTL)
	}
	if aud := []string(c.Audience); len(aud) != 2 || aud[0] != "master-service" || aud[1] != "risk-service" {
		t.Fatalf("aud = %v", aud)
	}
	// roles must be an empty array, never null, so a decoder sees []string{}.
	if !strings.Contains(decodePayload(t, tok), `"roles":[]`) {
		t.Fatalf("payload has no empty roles array: %s", decodePayload(t, tok))
	}
}

func TestSignedTokenExpiresAndNeedsSameSecret(t *testing.T) {
	now := time.Date(2026, 10, 27, 8, 0, 0, 0, time.UTC)
	tok, err := Sign(testSecret, now, DefaultTTL)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parseLikeServices(t, tok, testSecret, now.Add(DefaultTTL+time.Second)); !errors.Is(err, jwt.ErrTokenExpired) {
		t.Fatalf("expired token: err = %v, want ErrTokenExpired", err)
	}
	other := "test-" + strings.Repeat("q", 64)
	if _, err := parseLikeServices(t, tok, other, now); err == nil {
		t.Fatal("token accepted with another deployment's secret")
	}
}

func TestSignRefusesBadSecretOrTTL(t *testing.T) {
	if _, err := Sign("", time.Now(), DefaultTTL); err == nil {
		t.Fatal("signed with an empty secret")
	}
	if _, err := Sign("short", time.Now(), DefaultTTL); err == nil {
		t.Fatal("signed with a short secret")
	}
	if _, err := Sign(testSecret, time.Now(), 0); err == nil {
		t.Fatal("signed with ttl 0")
	}
}

// audit-service's own middleware has the same shape as master's and risk's;
// it must accept the token (proves the json claim names line up) and see no
// roles, so RequireRoles refuses it.
func TestTokenPassesAuthenticateButHoldsNoRole(t *testing.T) {
	tok, err := Sign(testSecret, time.Now(), DefaultTTL)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := jwt.ParseWithClaims(tok, &middleware.Claims{},
		func(*jwt.Token) (interface{}, error) { return []byte(testSecret), nil },
		jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())
	if err != nil {
		t.Fatal(err)
	}
	c := parsed.Claims.(*middleware.Claims)
	if c.UserID != Subject || len(c.Roles) != 0 {
		t.Fatalf("middleware sees user_id %q roles %v", c.UserID, c.Roles)
	}
}

func TestAuthorizationOverrideTakesPrecedence(t *testing.T) {
	h, src, err := Authorization("  bearer user.jwt.value ", testSecret, time.Now())
	if err != nil || h != "Bearer user.jwt.value" || src != SourceOverride {
		t.Fatalf("override: %q %q %v", h, src, err)
	}
	// The override wins even when the secret could not sign.
	if h, src, err := Authorization("user.jwt.value", "", time.Now()); err != nil || src != SourceOverride || h != "Bearer user.jwt.value" {
		t.Fatalf("override with no secret: %q %q %v", h, src, err)
	}
}

func TestAuthorizationSelfSignsWithoutOverride(t *testing.T) {
	now := time.Now()
	h, src, err := Authorization("", testSecret, now)
	if err != nil || src != SourceSelfSigned || !strings.HasPrefix(h, "Bearer ") {
		t.Fatalf("self-signed: src %q err %v", src, err)
	}
	if _, err := parseLikeServices(t, strings.TrimPrefix(h, "Bearer "), testSecret, now); err != nil {
		t.Fatalf("self-signed token invalid: %v", err)
	}
}

func TestAuthorizationMalformedOverrideIsError(t *testing.T) {
	// Set but unusable: fail loudly instead of silently self-signing.
	for _, v := range []string{"Bearer ", "two parts"} {
		if _, _, err := Authorization(v, testSecret, time.Now()); err == nil {
			t.Fatalf("override %q accepted", v)
		}
	}
	if _, _, err := Authorization("", "", time.Now()); err == nil {
		t.Fatal("no override and no secret did not fail")
	}
}

func TestBearer(t *testing.T) {
	cases := map[string]string{
		"":               "",
		"   ":            "",
		"Bearer ":        "",
		"abc.def.ghi":    "Bearer abc.def.ghi",
		" abc.def.ghi\n": "Bearer abc.def.ghi",
		"Bearer abc":     "Bearer abc",
		"bearer abc":     "Bearer abc",
		"BEARER   abc  ": "Bearer abc",
		"a b":            "",
	}
	for in, want := range cases {
		if got := Bearer(in); got != want {
			t.Errorf("Bearer(%q) = %q, want %q", in, got, want)
		}
	}
}

func decodePayload(t *testing.T, tok string) string {
	t.Helper()
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		t.Fatalf("not a JWT")
	}
	b, err := jwt.NewParser().DecodeSegment(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
