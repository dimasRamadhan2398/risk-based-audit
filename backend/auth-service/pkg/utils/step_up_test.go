package utils

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Cross-service test vector. master-service
// pkg/middleware/step_up_compat_test.go asserts the SAME secret -> key.
const (
	stepUpVectorSecret = "test-jwt-secret-0123456789abcdef0123456789"
	stepUpVectorKeyHex = "756760ac34c7a9234ec8a23f9572fed4d3ce8d205c74588eea8cb628bca2721d"
)

func TestDeriveStepUpKey_CrossServiceVector(t *testing.T) {
	if StepUpKeyLabel != "auditsphere-step-up-v1" {
		t.Fatalf("StepUpKeyLabel = %q; master-service verifies with auditsphere-step-up-v1", StepUpKeyLabel)
	}
	if got := hex.EncodeToString(DeriveStepUpKey(stepUpVectorSecret)); got != stepUpVectorKeyHex {
		t.Fatalf("DeriveStepUpKey = %s, want %s", got, stepUpVectorKeyHex)
	}
}

// The token must carry exactly sub, purpose, jti, iat, exp and be HS256.
func TestGenerateStepUpToken_Shape(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	tok, jti, err := GenerateStepUpToken("user-1", StepUpPurposeOrgLevels, stepUpVectorSecret, now)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		t.Fatalf("not a JWS: %d parts", len(parts))
	}
	var header map[string]any
	decodeSegment(t, parts[0], &header)
	if header["alg"] != "HS256" {
		t.Fatalf("alg = %v", header["alg"])
	}
	var payload map[string]any
	decodeSegment(t, parts[1], &payload)
	keys := make([]string, 0, len(payload))
	for k := range payload {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if strings.Join(keys, ",") != "exp,iat,jti,purpose,sub" {
		t.Fatalf("claims = %v", keys)
	}
	if payload["sub"] != "user-1" || payload["purpose"] != StepUpPurposeOrgLevels || payload["jti"] != jti || jti == "" {
		t.Fatalf("payload = %v", payload)
	}
	if exp, iat := payload["exp"].(float64), payload["iat"].(float64); exp-iat != 300 {
		t.Fatalf("exp - iat = %v, want 300", exp-iat)
	}
}

// A step-up token is not an access token and an access token is not a
// step-up token: the two are signed with different keys.
func TestStepUpAndAccessTokensDoNotCrossVerify(t *testing.T) {
	stepUp, _, err := GenerateStepUpToken("user-1", StepUpPurposeOrgLevels, stepUpVectorSecret, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseToken(stepUp, stepUpVectorSecret); err == nil {
		t.Fatal("step-up token accepted by ParseToken (access-token verification)")
	}
	access, err := GenerateToken("user-1", "u", []string{"ADMIN"}, stepUpVectorSecret, 1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = jwt.ParseWithClaims(access, &StepUpClaims{}, func(*jwt.Token) (interface{}, error) {
		return DeriveStepUpKey(stepUpVectorSecret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err == nil {
		t.Fatal("access token verified with the step-up key")
	}
}

func decodeSegment(t *testing.T, seg string, out any) {
	t.Helper()
	b, err := base64.RawURLEncoding.DecodeString(seg)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, out); err != nil {
		t.Fatal(err)
	}
}
