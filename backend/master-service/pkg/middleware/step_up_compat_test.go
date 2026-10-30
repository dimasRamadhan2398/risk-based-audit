package middleware

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"master-service/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// Cross-service test vector. auth-service pkg/utils/step_up_test.go asserts
// the SAME secret -> key; if either side changes the label or the
// construction, one of the two tests fails.
const (
	stepUpVectorSecret = "test-jwt-secret-0123456789abcdef0123456789"
	stepUpVectorKeyHex = "756760ac34c7a9234ec8a23f9572fed4d3ce8d205c74588eea8cb628bca2721d"
)

func TestDeriveStepUpKey_CrossServiceVector(t *testing.T) {
	if StepUpKeyLabel != "auditsphere-step-up-v1" {
		t.Fatalf("StepUpKeyLabel = %q; auth-service signs with auditsphere-step-up-v1", StepUpKeyLabel)
	}
	if got := hex.EncodeToString(DeriveStepUpKey(stepUpVectorSecret)); got != stepUpVectorKeyHex {
		t.Fatalf("DeriveStepUpKey = %s, want %s", got, stepUpVectorKeyHex)
	}
}

const compatSecret = "compat-secret-at-least-64-bytes-long-compat-secret-at-least-64-bytes"

type compatStore struct {
	mu   sync.Mutex
	seen map[string]bool
}

func (s *compatStore) ConsumeOnce(_ context.Context, jti string, _ time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.seen[jti] {
		return false, nil
	}
	s.seen[jti] = true
	return true, nil
}

func compatSign(t *testing.T, key []byte, claims jwt.Claims) string {
	t.Helper()
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(key)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// compatAccessToken mirrors auth-service utils.GenerateToken (user_id claim).
func compatAccessToken(t *testing.T, userID string) string {
	now := time.Now()
	return compatSign(t, []byte(compatSecret), Claims{UserID: userID, Username: "admin1", Roles: []string{"ADMIN"},
		RegisteredClaims: jwt.RegisteredClaims{IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)), ID: "access-jti"}})
}

func compatStepUp(t *testing.T, sub, purpose, jti string, iat time.Time, ttl time.Duration) string {
	return compatSign(t, DeriveStepUpKey(compatSecret), StepUpClaims{Purpose: purpose, RegisteredClaims: jwt.RegisteredClaims{
		Subject: sub, ID: jti, IssuedAt: jwt.NewNumericDate(iat), ExpiresAt: jwt.NewNumericDate(iat.Add(ttl)),
	}})
}

// compatRouter: the real Authenticate then RequireStepUp, as on the apply route.
func compatRouter(store JTIStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	m := NewAuthMiddleware(compatSecret)
	r := gin.New()
	r.POST("/api/v1/org-levels/changes/apply", m.Authenticate(), m.RequireStepUp(StepUpPurposeOrgLevels, store),
		func(c *gin.Context) { c.Status(http.StatusNoContent) })
	return r
}

func compatDo(r *gin.Engine, access, stepUp string) (int, string) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/org-levels/changes/apply", nil)
	req.Header.Set("Authorization", "Bearer "+access)
	if stepUp != "" {
		req.Header.Set(StepUpHeader, stepUp)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	return w.Code, body.Error.Code
}

func withObservedLogger(t *testing.T) *observer.ObservedLogs {
	t.Helper()
	core, logs := observer.New(zap.DebugLevel)
	prev := logger.Log
	logger.Log = zap.New(core)
	t.Cleanup(func() { logger.Log = prev })
	return logs
}

// assertNoTokenInLogs fails if any logged field contains a raw token.
func assertNoTokenInLogs(t *testing.T, logs *observer.ObservedLogs, tokens ...string) {
	t.Helper()
	for _, e := range logs.All() {
		for k, v := range e.ContextMap() {
			s, _ := json.Marshal(v)
			for _, tok := range tokens {
				if tok != "" && (strings.Contains(string(s), tok) || strings.Contains(e.Message, tok)) {
					t.Fatalf("log %q field %q contains a token", e.Message, k)
				}
			}
		}
	}
}

func lastReason(logs *observer.ObservedLogs) string {
	all := logs.FilterMessage("step-up rejected").All()
	if len(all) == 0 {
		return ""
	}
	r, _ := all[len(all)-1].ContextMap()["reason"].(string)
	return r
}

func TestRequireStepUp_FullChainAndLogging(t *testing.T) {
	logs := withObservedLogger(t)
	const user = "6f1c1b1e-0000-4000-8000-0000000000a1"
	access := compatAccessToken(t, user)
	now := time.Now()
	r := compatRouter(&compatStore{seen: map[string]bool{}})

	good := compatStepUp(t, user, StepUpPurposeOrgLevels, "jti-ok", now, StepUpTTL)
	if code, _ := compatDo(r, access, good); code != http.StatusNoContent {
		t.Fatalf("valid step-up token: %d", code)
	}
	if n := logs.FilterMessage("step-up accepted").Len(); n != 1 {
		t.Fatalf("accepted log lines = %d, want 1", n)
	}

	expiredInLeeway := compatStepUp(t, user, StepUpPurposeOrgLevels, "jti-leeway", now.Add(-StepUpTTL-20*time.Second), StepUpTTL)
	if code, _ := compatDo(r, access, expiredInLeeway); code != http.StatusNoContent {
		t.Fatalf("token expired 20s ago (inside 30s leeway): %d", code)
	}

	otherUser := compatStepUp(t, "6f1c1b1e-0000-4000-8000-0000000000b2", StepUpPurposeOrgLevels, "jti-other", now, StepUpTTL)
	expired := compatStepUp(t, user, StepUpPurposeOrgLevels, "jti-expired", now.Add(-StepUpTTL-45*time.Second), StepUpTTL)
	cases := []struct {
		name, token, wantCode, wantReason string
		wantStatus                        int
	}{
		{"missing header", "", "STEP_UP_REQUIRED", "missing_token", http.StatusUnauthorized},
		{"access token as step-up token", access, "STEP_UP_INVALID", "bad_signature", http.StatusUnauthorized},
		{"expired beyond leeway", expired, "STEP_UP_INVALID", "expired", http.StatusUnauthorized},
		{"wrong purpose", compatStepUp(t, user, "users.delete", "jti-purpose", now, StepUpTTL), "STEP_UP_INVALID", "purpose_mismatch", http.StatusUnauthorized},
		{"sub is another user", otherUser, "STEP_UP_INVALID", "subject_mismatch", http.StatusUnauthorized},
		{"replayed jti", good, "STEP_UP_INVALID", "replayed", http.StatusUnauthorized},
	}
	for _, tc := range cases {
		status, code := compatDo(r, access, tc.token)
		if status != tc.wantStatus || code != tc.wantCode {
			t.Fatalf("%s: got %d %s, want %d %s", tc.name, status, code, tc.wantStatus, tc.wantCode)
		}
		if got := lastReason(logs); got != tc.wantReason {
			t.Fatalf("%s: logged reason %q, want %q", tc.name, got, tc.wantReason)
		}
		last := logs.FilterMessage("step-up rejected").All()
		if uid, _ := last[len(last)-1].ContextMap()["user_id"].(string); uid != user {
			t.Fatalf("%s: logged user_id %q, want %q", tc.name, uid, user)
		}
	}

	// nil store (no Redis): 503 STEP_UP_UNAVAILABLE, logged as an error.
	status, code := compatDo(compatRouter(nil), access, compatStepUp(t, user, StepUpPurposeOrgLevels, "jti-nil", now, StepUpTTL))
	if status != http.StatusServiceUnavailable || code != "STEP_UP_UNAVAILABLE" {
		t.Fatalf("nil store: %d %s", status, code)
	}
	if lastReason(logs) != "store_unavailable" {
		t.Fatalf("nil store: reason %q", lastReason(logs))
	}

	assertNoTokenInLogs(t, logs, access, good, expiredInLeeway, otherUser, expired)
}

// An access token never passes RequireStepUp even when it carries matching
// claims, because it is signed with the raw secret, not the derived key.
func TestRequireStepUp_RawSecretTokenWithStepUpClaims(t *testing.T) {
	withObservedLogger(t)
	const user = "6f1c1b1e-0000-4000-8000-0000000000a1"
	now := time.Now()
	forged := compatSign(t, []byte(compatSecret), StepUpClaims{Purpose: StepUpPurposeOrgLevels, RegisteredClaims: jwt.RegisteredClaims{
		Subject: user, ID: "jti-forged", IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(now.Add(StepUpTTL)),
	}})
	if status, code := compatDo(compatRouter(&compatStore{seen: map[string]bool{}}), compatAccessToken(t, user), forged); status != http.StatusUnauthorized || code != "STEP_UP_INVALID" {
		t.Fatalf("raw-secret token: %d %s", status, code)
	}
}
