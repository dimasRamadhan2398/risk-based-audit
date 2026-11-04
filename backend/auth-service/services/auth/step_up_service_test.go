package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"auth-service/models"
	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/kafka"
	"auth-service/pkg/logger"
	"auth-service/pkg/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

const (
	stepUpTestSecret   = "step-up-test-secret-at-least-64-bytes-step-up-test-secret-64-bytes"
	stepUpTestPassword = "Correct-Horse-9!"
	stepUpWrongPass    = "Wrong-Battery-7?"
)

type fakeStepUpUsers struct {
	user *models.User
	err  error
}

func (f fakeStepUpUsers) FindByID(id uuid.UUID) (*models.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.user == nil || f.user.ID != id {
		return nil, apperrors.ErrNotFound
	}
	return f.user, nil
}

// fakeAttempts mimics the CacheRepository login-attempt helpers.
type fakeAttempts struct {
	mu     sync.Mutex
	counts map[string]int64
	getErr error
	ids    []string
}

func (f *fakeAttempts) TrackLoginAttempt(_ context.Context, id string, exp time.Duration) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if exp != StepUpFailureWindow {
		panic("unexpected window")
	}
	f.ids = append(f.ids, id)
	f.counts[id]++
	return f.counts[id], nil
}

func (f *fakeAttempts) GetLoginAttempts(_ context.Context, id string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getErr != nil {
		return 0, f.getErr
	}
	return f.counts[id], nil
}

func (f *fakeAttempts) ClearLoginAttempts(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.counts, id)
	return nil
}

type fakePublisher struct {
	events []*kafka.Event
}

func (f *fakePublisher) Publish(_ context.Context, e *kafka.Event) error {
	f.events = append(f.events, e)
	return nil
}

type stepUpFixture struct {
	svc      *StepUpService
	user     *models.User
	attempts *fakeAttempts
	pub      *fakePublisher
	logs     *observer.ObservedLogs
}

var (
	stepUpHashOnce sync.Once
	stepUpHash     string
)

func newStepUpFixture(t *testing.T) *stepUpFixture {
	t.Helper()
	stepUpHashOnce.Do(func() {
		h, err := utils.HashPassword(stepUpTestPassword)
		if err != nil {
			t.Fatal(err)
		}
		stepUpHash = h
	})
	core, logs := observer.New(zap.DebugLevel)
	prev := logger.Log
	logger.Log = zap.New(core)
	t.Cleanup(func() { logger.Log = prev })

	user := &models.User{ID: uuid.New(), Username: "admin1", PasswordHash: stepUpHash, IsActive: true}
	attempts := &fakeAttempts{counts: map[string]int64{}}
	pub := &fakePublisher{}
	return &stepUpFixture{
		svc:      NewStepUpService(fakeStepUpUsers{user: user}, attempts, pub, stepUpTestSecret),
		user:     user,
		attempts: attempts,
		pub:      pub,
		logs:     logs,
	}
}

func (f *stepUpFixture) input(password, purpose string) VerifyPasswordInput {
	return VerifyPasswordInput{UserID: f.user.ID.String(), Username: f.user.Username, Password: password,
		Purpose: purpose, IP: "10.0.0.1", UserAgent: "test"}
}

// assertNoSecrets: neither the password nor any issued token may appear in
// a log line or in an emitted event.
func (f *stepUpFixture) assertNoSecrets(t *testing.T, secrets ...string) {
	t.Helper()
	secrets = append(secrets, stepUpTestPassword, stepUpWrongPass, stepUpHash)
	for _, e := range f.logs.All() {
		b, _ := json.Marshal(e.ContextMap())
		for _, s := range secrets {
			if s != "" && (strings.Contains(string(b), s) || strings.Contains(e.Message, s)) {
				t.Fatalf("log %q leaks a secret: %s", e.Message, b)
			}
		}
	}
	for _, ev := range f.pub.events {
		b, _ := json.Marshal(ev)
		for _, s := range secrets {
			if s != "" && strings.Contains(string(b), s) {
				t.Fatalf("event leaks a secret: %s", b)
			}
		}
	}
}

func (f *stepUpFixture) lastOutcome(t *testing.T) (string, *kafka.Event) {
	t.Helper()
	all := f.logs.All()
	if len(all) == 0 || len(f.pub.events) == 0 {
		t.Fatalf("no log (%d) or event (%d)", len(all), len(f.pub.events))
	}
	outcome, _ := all[len(all)-1].ContextMap()["outcome"].(string)
	return outcome, f.pub.events[len(f.pub.events)-1]
}

func TestVerifyPassword_SuccessIssuesVerifiableToken(t *testing.T) {
	f := newStepUpFixture(t)
	res, err := f.svc.VerifyPassword(context.Background(), f.input(stepUpTestPassword, utils.StepUpPurposeOrgLevels))
	if err != nil {
		t.Fatal(err)
	}
	if res.ExpiresIn != 300 || res.StepUpToken == "" {
		t.Fatalf("response = %+v", res)
	}
	claims := verifyLikeMaster(t, res.StepUpToken, stepUpTestSecret)
	if claims.Subject != f.user.ID.String() || claims.Purpose != utils.StepUpPurposeOrgLevels {
		t.Fatalf("claims = %+v", claims)
	}

	outcome, ev := f.lastOutcome(t)
	if outcome != StepUpOutcomeSuccess || ev.Type != kafka.EventStepUp || !ev.Success || ev.UserID != f.user.ID.String() || ev.Data["jti"] != claims.ID {
		t.Fatalf("outcome %q event %+v", outcome, ev)
	}
	f.assertNoSecrets(t, res.StepUpToken)
}

func TestVerifyPassword_WrongPasswordCountsAndLocksAfterFive(t *testing.T) {
	f := newStepUpFixture(t)
	ctx := context.Background()
	for i := 1; i <= 4; i++ {
		_, err := f.svc.VerifyPassword(ctx, f.input(stepUpWrongPass, utils.StepUpPurposeOrgLevels))
		if !apperrors.Is(err, ErrStepUpPasswordInvalid) || err.(*apperrors.AppError).StatusCode != http.StatusForbidden {
			t.Fatalf("attempt %d: err = %v", i, err)
		}
		outcome, ev := f.lastOutcome(t)
		if outcome != StepUpOutcomeWrongPassword || ev.Success || ev.ErrorReason != StepUpOutcomeWrongPassword || ev.Data["failed_attempts"] != int64(i) {
			t.Fatalf("attempt %d: outcome %q event %+v", i, outcome, ev)
		}
	}
	// 5th failure locks.
	if _, err := f.svc.VerifyPassword(ctx, f.input(stepUpWrongPass, utils.StepUpPurposeOrgLevels)); !apperrors.Is(err, ErrStepUpLocked) {
		t.Fatalf("5th failure: err = %v", err)
	}
	if outcome, _ := f.lastOutcome(t); outcome != StepUpOutcomeLockedOut {
		t.Fatalf("5th failure outcome %q", outcome)
	}
	// Locked: even the right password is refused, and not counted.
	if _, err := f.svc.VerifyPassword(ctx, f.input(stepUpTestPassword, utils.StepUpPurposeOrgLevels)); !apperrors.Is(err, ErrStepUpLocked) ||
		err.(*apperrors.AppError).StatusCode != http.StatusTooManyRequests {
		t.Fatalf("locked user with right password: err = %v", err)
	}
	if outcome, _ := f.lastOutcome(t); outcome != StepUpOutcomeLockedOut {
		t.Fatalf("locked outcome %q", outcome)
	}
	// Own counter, not the login one.
	want := "stepup:" + f.user.ID.String()
	for _, id := range f.attempts.ids {
		if id != want {
			t.Fatalf("counter identifier %q, want %q", id, want)
		}
	}
	if f.attempts.counts[want] != 5 {
		t.Fatalf("counter = %d, want 5", f.attempts.counts[want])
	}
	f.assertNoSecrets(t)
}

func TestVerifyPassword_SuccessClearsFailures(t *testing.T) {
	f := newStepUpFixture(t)
	ctx := context.Background()
	_, _ = f.svc.VerifyPassword(ctx, f.input(stepUpWrongPass, utils.StepUpPurposeOrgLevels))
	if _, err := f.svc.VerifyPassword(ctx, f.input(stepUpTestPassword, utils.StepUpPurposeOrgLevels)); err != nil {
		t.Fatal(err)
	}
	if n := f.attempts.counts["stepup:"+f.user.ID.String()]; n != 0 {
		t.Fatalf("counter after success = %d", n)
	}
}

func TestVerifyPassword_PurposeWhitelist(t *testing.T) {
	f := newStepUpFixture(t)
	_, err := f.svc.VerifyPassword(context.Background(), f.input(stepUpTestPassword, "users.delete"))
	if !apperrors.Is(err, ErrStepUpPurposeInvalid) || err.(*apperrors.AppError).StatusCode != http.StatusBadRequest {
		t.Fatalf("err = %v", err)
	}
	outcome, ev := f.lastOutcome(t)
	if outcome != StepUpOutcomeInvalidPurpose || ev.Success || ev.Data["purpose"] != "" {
		t.Fatalf("outcome %q event %+v", outcome, ev)
	}
	if len(f.attempts.ids) != 0 {
		t.Fatal("a bad purpose must not count as a password failure")
	}
	f.assertNoSecrets(t)
}

func TestVerifyPassword_FailsClosedWithoutRedis(t *testing.T) {
	f := newStepUpFixture(t)
	f.attempts.getErr = errors.New("redis down")
	_, err := f.svc.VerifyPassword(context.Background(), f.input(stepUpTestPassword, utils.StepUpPurposeOrgLevels))
	if !apperrors.Is(err, ErrStepUpUnavailable) {
		t.Fatalf("err = %v", err)
	}
	nilStore := NewStepUpService(fakeStepUpUsers{user: f.user}, nil, nil, stepUpTestSecret)
	if _, err := nilStore.VerifyPassword(context.Background(), f.input(stepUpTestPassword, utils.StepUpPurposeOrgLevels)); !apperrors.Is(err, ErrStepUpUnavailable) {
		t.Fatalf("nil store: err = %v", err)
	}
	f.assertNoSecrets(t)
}

func TestVerifyPassword_InactiveUser(t *testing.T) {
	f := newStepUpFixture(t)
	f.user.IsActive = false
	if _, err := f.svc.VerifyPassword(context.Background(), f.input(stepUpTestPassword, utils.StepUpPurposeOrgLevels)); !apperrors.Is(err, apperrors.ErrUnauthorized) {
		t.Fatalf("err = %v", err)
	}
}

// verifyLikeMaster applies master-service RequireStepUp's parse options.
func verifyLikeMaster(t *testing.T, token, secret string) *utils.StepUpClaims {
	t.Helper()
	claims := &utils.StepUpClaims{}
	tok, err := jwt.ParseWithClaims(token, claims, func(*jwt.Token) (interface{}, error) { return utils.DeriveStepUpKey(secret), nil },
		jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithLeeway(30*time.Second))
	if err != nil || !tok.Valid {
		t.Fatalf("token rejected: %v", err)
	}
	return claims
}
