package mfa

import (
	"context"
	"encoding/base32"
	"errors"
	"sync"
	"testing"
	"time"

	"auth-service/models"
	apperrors "auth-service/pkg/errors"
	pkgKafka "auth-service/pkg/kafka"
	"auth-service/pkg/logger"
	"auth-service/repositories"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// memMFARepo keeps one mfa_setups row in memory. Rows are copied in and out,
// like a DB, so the service never shares a pointer with the store. User is
// never preloaded, as with the real FindByUserID.
type memMFARepo struct {
	repositories.MFASetupRepositoryInterface
	mu  sync.Mutex
	row *models.MFASetup
	// beforeEnable runs inside EnablePending before the conditional update,
	// to simulate a request that changes the row in between.
	beforeEnable func(r *memMFARepo)
}

func (r *memMFARepo) FindByUserID(userID uuid.UUID) (*models.MFASetup, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.row == nil || r.row.UserID != userID {
		return nil, apperrors.ErrNotFound
	}
	cp := *r.row
	return &cp, nil
}

func (r *memMFARepo) Create(m *models.MFASetup) error { return r.Update(m) }

func (r *memMFARepo) Update(m *models.MFASetup) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *m
	r.row = &cp
	return nil
}

func (r *memMFARepo) EnablePending(userID uuid.UUID, secret string) (bool, error) {
	if r.beforeEnable != nil {
		r.beforeEnable(r)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.row == nil || r.row.UserID != userID || r.row.SecretKey != secret || r.row.IsEnabled {
		return false, nil
	}
	r.row.IsEnabled = true
	return true, nil
}

type stubUsers struct {
	repositories.UserRepositoryInterface
	user *models.User
}

func (s stubUsers) FindByID(id uuid.UUID) (*models.User, error) {
	if s.user == nil || s.user.ID != id {
		return nil, apperrors.ErrNotFound
	}
	return s.user, nil
}

type recordingDevices struct {
	repositories.ITrustedDeviceRepository
	mu        sync.Mutex
	sightings []repositories.DeviceSighting
}

func (d *recordingDevices) RecordSeen(s repositories.DeviceSighting) (*models.TrustedDevice, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.sightings = append(d.sightings, s)
	return &models.TrustedDevice{UserID: s.UserID, DeviceFingerprint: s.Fingerprint}, nil
}

type countingPublisher struct {
	pkgKafka.IEventPublisher
	mu    sync.Mutex
	types []pkgKafka.EventType
}

func (p *countingPublisher) PublishMFAEvent(_ context.Context, _, _ string, t pkgKafka.EventType, _ bool, _ string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.types = append(p.types, t)
	return nil
}

func (p *countingPublisher) count(t pkgKafka.EventType) int {
	p.mu.Lock()
	defer p.mu.Unlock()
	n := 0
	for _, got := range p.types {
		if got == t {
			n++
		}
	}
	return n
}

type fixture struct {
	svc     *MfaService
	mfa     *memMFARepo
	devices *recordingDevices
	user    *models.User
}

func newFixture(t *testing.T, publisher pkgKafka.IEventPublisher) *fixture {
	t.Helper()
	if logger.Log == nil {
		logger.Log = zap.NewNop()
	}
	user := &models.User{
		ID:                   uuid.New(),
		Username:             "auditor1",
		Email:                "auditor1@example.test",
		LastLoginFingerprint: "fp-last-login",
	}
	f := &fixture{mfa: &memMFARepo{}, devices: &recordingDevices{}, user: user}
	f.svc = NewMfaService(f.mfa, stubUsers{user: user}, f.devices, publisher, nil, nil).(*MfaService)
	return f
}

// codeAt is the TOTP code offset steps (30 s each) away from now.
func codeAt(t *testing.T, secret string, offset int64) string {
	t.Helper()
	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		t.Fatal(err)
	}
	return generateHOTP(key, time.Now().Unix()/30+offset)
}

func appCode(err error) string {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return ""
}

func (f *fixture) setup(t *testing.T) string {
	t.Helper()
	res, err := f.svc.SetupMFA(context.Background(), f.user.ID, models.MFATypeTOTP, "203.0.113.7")
	if err != nil {
		t.Fatalf("SetupMFA: %v", err)
	}
	if res.Secret == "" || res.IsEnabled {
		t.Fatalf("setup response = %+v, want a secret and is_enabled=false", res)
	}
	return res.Secret
}

func TestSetupStoresPendingSecretNotEnabled(t *testing.T) {
	f := newFixture(t, nil)
	secret := f.setup(t)

	row, _ := f.mfa.FindByUserID(f.user.ID)
	if row.IsEnabled {
		t.Fatal("setup enabled MFA before the code was verified")
	}
	if row.SecretKey != secret {
		t.Fatal("setup did not store the secret it returned")
	}
}

// Regression: EnableMFA dereferenced mfa.User (never preloaded) after the row
// was already switched on. The request panicked into a 500, the UI showed a
// failure, and the user's retry with the same code got MFA_ALREADY_ENABLED.
func TestEnableWithoutPreloadedUserOrPublisherDoesNotPanic(t *testing.T) {
	f := newFixture(t, nil)
	secret := f.setup(t)

	res, err := f.svc.EnableMFA(context.Background(), f.user.ID, models.VerifyMfaRequest{Code: codeAt(t, secret, 0)}, "203.0.113.7", "UA")
	if err != nil {
		t.Fatalf("EnableMFA: %v", err)
	}
	if res.AlreadyEnabled {
		t.Fatal("first verify reported already_enabled")
	}
	row, _ := f.mfa.FindByUserID(f.user.ID)
	if !row.IsEnabled {
		t.Fatal("MFA not enabled after a valid code")
	}
}

func TestEnableIsIdempotentForRetryWithValidCode(t *testing.T) {
	pub := &countingPublisher{}
	f := newFixture(t, pub)
	secret := f.setup(t)
	code := codeAt(t, secret, 0)
	ctx := context.Background()

	if _, err := f.svc.EnableMFA(ctx, f.user.ID, models.VerifyMfaRequest{Code: code}, "ip", "UA"); err != nil {
		t.Fatalf("first verify: %v", err)
	}
	res, err := f.svc.EnableMFA(ctx, f.user.ID, models.VerifyMfaRequest{Code: code}, "ip", "UA")
	if err != nil {
		t.Fatalf("retried verify with the same valid code: %v", err)
	}
	if !res.AlreadyEnabled {
		t.Fatal("retry did not report already_enabled")
	}
	if got := pub.count(pkgKafka.EventMFAEnabled); got != 2 { // 1 from setup, 1 from the first verify
		t.Fatalf("mfa.enabled events = %d, want 2 (a retry must not publish again)", got)
	}
}

func TestEnableAlreadyEnabledWithWrongCodeStillErrors(t *testing.T) {
	f := newFixture(t, nil)
	secret := f.setup(t)
	ctx := context.Background()
	if _, err := f.svc.EnableMFA(ctx, f.user.ID, models.VerifyMfaRequest{Code: codeAt(t, secret, 0)}, "ip", "UA"); err != nil {
		t.Fatal(err)
	}

	_, err := f.svc.EnableMFA(ctx, f.user.ID, models.VerifyMfaRequest{Code: codeAt(t, secret, 5)}, "ip", "UA")
	if appCode(err) != "MFA_ALREADY_ENABLED" {
		t.Fatalf("err = %v, want MFA_ALREADY_ENABLED", err)
	}
}

func TestEnableRejectsInvalidCodeAndStaysPending(t *testing.T) {
	f := newFixture(t, nil)
	secret := f.setup(t)

	_, err := f.svc.EnableMFA(context.Background(), f.user.ID, models.VerifyMfaRequest{Code: codeAt(t, secret, 5)}, "ip", "UA")
	if appCode(err) != "INVALID_CODE" {
		t.Fatalf("err = %v, want INVALID_CODE", err)
	}
	if row, _ := f.mfa.FindByUserID(f.user.ID); row.IsEnabled {
		t.Fatal("invalid code enabled MFA")
	}
	if len(f.devices.sightings) != 0 {
		t.Fatal("device recorded for an invalid code")
	}
}

func TestTOTPAcceptsOneStepOfSkew(t *testing.T) {
	f := newFixture(t, nil)
	secret := f.setup(t)
	for _, tc := range []struct {
		offset int64
		want   bool
	}{{-2, false}, {-1, true}, {0, true}, {1, true}, {2, false}} {
		got, _ := f.svc.verifyTOTP(secret, codeAt(t, secret, tc.offset))
		if got != tc.want {
			t.Errorf("offset %+d steps: valid = %v, want %v", tc.offset, got, tc.want)
		}
	}
}

func TestEnableAfterSetupReplacedSecretIsRejected(t *testing.T) {
	f := newFixture(t, nil)
	secret := f.setup(t)
	f.mfa.beforeEnable = func(r *memMFARepo) {
		r.mu.Lock()
		r.row.SecretKey = "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP" // a second setup call won the race
		r.mu.Unlock()
	}

	_, err := f.svc.EnableMFA(context.Background(), f.user.ID, models.VerifyMfaRequest{Code: codeAt(t, secret, 0)}, "ip", "UA")
	if appCode(err) != "MFA_SETUP_CHANGED" {
		t.Fatalf("err = %v, want MFA_SETUP_CHANGED", err)
	}
	if row, _ := f.mfa.FindByUserID(f.user.ID); row.IsEnabled {
		t.Fatal("MFA enabled with a secret the user never scanned")
	}
}

func TestEnableConcurrentDoubleSubmitBothSucceed(t *testing.T) {
	f := newFixture(t, nil)
	secret := f.setup(t)
	code := codeAt(t, secret, 0)

	var wg sync.WaitGroup
	errs := make([]error, 2)
	results := make([]*models.EnableMfaResult, 2)
	for i := range errs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i], errs[i] = f.svc.EnableMFA(context.Background(), f.user.ID, models.VerifyMfaRequest{Code: code}, "ip", "UA")
		}(i)
	}
	wg.Wait()
	fresh := 0
	for i, err := range errs {
		if err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
		if !results[i].AlreadyEnabled {
			fresh++
		}
	}
	if fresh != 1 {
		t.Fatalf("%d requests reported a fresh enable, want exactly 1", fresh)
	}
}

func TestEnableRecordsSessionOnlyDevice(t *testing.T) {
	t.Run("fingerprint from request", func(t *testing.T) {
		f := newFixture(t, nil)
		secret := f.setup(t)
		req := models.VerifyMfaRequest{Code: codeAt(t, secret, 0), DeviceFingerprint: "fp-req", DeviceName: "Chrome on macOS", DeviceType: "desktop"}
		if _, err := f.svc.EnableMFA(context.Background(), f.user.ID, req, "203.0.113.7", "Mozilla/5.0"); err != nil {
			t.Fatal(err)
		}
		if len(f.devices.sightings) != 1 {
			t.Fatalf("sightings = %d, want 1", len(f.devices.sightings))
		}
		s := f.devices.sightings[0]
		if s.Fingerprint != "fp-req" || s.Name != "Chrome on macOS" || s.IPAddress != "203.0.113.7" || s.UserAgent != "Mozilla/5.0" {
			t.Fatalf("sighting = %+v", s)
		}
		if s.Trust {
			t.Fatal("enabling MFA must not trust the device (it would skip MFA)")
		}
	})
	t.Run("falls back to last sign-in fingerprint", func(t *testing.T) {
		f := newFixture(t, nil)
		secret := f.setup(t)
		if _, err := f.svc.EnableMFA(context.Background(), f.user.ID, models.VerifyMfaRequest{Code: codeAt(t, secret, 0)}, "ip", "UA"); err != nil {
			t.Fatal(err)
		}
		if len(f.devices.sightings) != 1 || f.devices.sightings[0].Fingerprint != "fp-last-login" {
			t.Fatalf("sightings = %+v, want one with fp-last-login", f.devices.sightings)
		}
	})
	t.Run("idempotent retry records the device too", func(t *testing.T) {
		f := newFixture(t, nil)
		secret := f.setup(t)
		code := codeAt(t, secret, 0)
		for i := 0; i < 2; i++ {
			if _, err := f.svc.EnableMFA(context.Background(), f.user.ID, models.VerifyMfaRequest{Code: code}, "ip", "UA"); err != nil {
				t.Fatal(err)
			}
		}
		if len(f.devices.sightings) != 2 {
			t.Fatalf("sightings = %d, want 2 (the repository dedupes by fingerprint)", len(f.devices.sightings))
		}
	})
}

func TestSetupRefusedWhenEnabled(t *testing.T) {
	f := newFixture(t, nil)
	secret := f.setup(t)
	if _, err := f.svc.EnableMFA(context.Background(), f.user.ID, models.VerifyMfaRequest{Code: codeAt(t, secret, 0)}, "ip", "UA"); err != nil {
		t.Fatal(err)
	}
	_, err := f.svc.SetupMFA(context.Background(), f.user.ID, models.MFATypeTOTP, "ip")
	if appCode(err) != "MFA_ALREADY_ENABLED" {
		t.Fatalf("err = %v, want MFA_ALREADY_ENABLED", err)
	}
	if row, _ := f.mfa.FindByUserID(f.user.ID); row.SecretKey != secret {
		t.Fatal("setup on an enabled account replaced the secret")
	}
}
