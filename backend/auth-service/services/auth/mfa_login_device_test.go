package services

import (
	"bufio"
	"context"
	"encoding/base32"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"auth-service/models"
	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/redis"
	"auth-service/pkg/utils"
	"auth-service/repositories"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

// fakeRedis is a minimal RESP2 server with GET/SET/DEL, enough for the MFA
// login token. Unknown commands (HELLO, CLIENT ...) get an error or +OK.
type fakeRedis struct {
	mu   sync.Mutex
	data map[string]string
}

func newFakeRedis(t *testing.T) *redis.Client {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	fr := &fakeRedis{data: map[string]string{}}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go fr.serve(c)
		}
	}()
	rdb := &redis.Client{Client: goredis.NewClient(&goredis.Options{Addr: ln.Addr().String(), Protocol: 2, DisableIndentity: true})}
	t.Cleanup(func() { _ = rdb.Close(); _ = ln.Close() })
	return rdb
}

func (f *fakeRedis) serve(c net.Conn) {
	defer c.Close()
	r := bufio.NewReader(c)
	for {
		args, err := readCommand(r)
		if err != nil {
			return
		}
		var reply string
		f.mu.Lock()
		switch strings.ToUpper(args[0]) {
		case "PING":
			reply = "+PONG\r\n"
		case "SET":
			f.data[args[1]] = args[2]
			reply = "+OK\r\n"
		case "GET":
			if v, ok := f.data[args[1]]; ok {
				reply = fmt.Sprintf("$%d\r\n%s\r\n", len(v), v)
			} else {
				reply = "$-1\r\n"
			}
		case "DEL":
			n := 0
			for _, k := range args[1:] {
				if _, ok := f.data[k]; ok {
					delete(f.data, k)
					n++
				}
			}
			reply = fmt.Sprintf(":%d\r\n", n)
		case "INCR":
			n, _ := strconv.Atoi(f.data[args[1]])
			n++
			f.data[args[1]] = strconv.Itoa(n)
			reply = fmt.Sprintf(":%d\r\n", n)
		case "TTL":
			// Expiry is not simulated: a present key reports a fixed 900s.
			if _, ok := f.data[args[1]]; ok {
				reply = ":900\r\n"
			} else {
				reply = ":-2\r\n"
			}
		case "EXISTS":
			n := 0
			for _, k := range args[1:] {
				if _, ok := f.data[k]; ok {
					n++
				}
			}
			reply = fmt.Sprintf(":%d\r\n", n)
		case "EXPIRE":
			if _, ok := f.data[args[1]]; ok {
				reply = ":1\r\n"
			} else {
				reply = ":0\r\n"
			}
		case "HELLO":
			reply = "-ERR unknown command 'HELLO'\r\n"
		default:
			reply = "+OK\r\n"
		}
		f.mu.Unlock()
		if _, err := io.WriteString(c, reply); err != nil {
			return
		}
	}
}

func readCommand(r *bufio.Reader) ([]string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(line, "*") {
		return nil, fmt.Errorf("unexpected %q", line)
	}
	n, err := strconv.Atoi(strings.TrimSpace(line[1:]))
	if err != nil {
		return nil, err
	}
	args := make([]string, 0, n)
	for i := 0; i < n; i++ {
		hdr, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		size, err := strconv.Atoi(strings.TrimSpace(hdr[1:]))
		if err != nil {
			return nil, err
		}
		buf := make([]byte, size+2)
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		args = append(args, string(buf[:size]))
	}
	return args, nil
}

// memDevices is an in-memory trusted device store with the real
// one-row-per-fingerprint and trust rules.
type memDevices struct {
	repositories.ITrustedDeviceRepository
	mu   sync.Mutex
	rows []*models.TrustedDevice
}

func (m *memDevices) FindByFingerprint(userID uuid.UUID, fp string) (*models.TrustedDevice, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, d := range m.rows {
		if d.UserID == userID && d.DeviceFingerprint == fp {
			cp := *d
			return &cp, nil
		}
	}
	return nil, apperrors.ErrNotFound
}

func (m *memDevices) IsTrusted(d *models.TrustedDevice) bool {
	return (&repositories.TrustedDeviceRepository{}).IsTrusted(d)
}

func (m *memDevices) RecordSeen(s repositories.DeviceSighting) (*models.TrustedDevice, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	var d *models.TrustedDevice
	for _, row := range m.rows {
		if row.UserID == s.UserID && row.DeviceFingerprint == s.Fingerprint {
			d = row
		}
	}
	if d == nil {
		d = &models.TrustedDevice{ID: uuid.New(), UserID: s.UserID, DeviceFingerprint: s.Fingerprint, SessionOnly: true, CreatedAt: now}
		m.rows = append(m.rows, d)
	}
	if s.Name != "" {
		d.DeviceName = s.Name
	}
	d.UserAgent, d.IPAddress, d.LastSeenAt = s.UserAgent, s.IPAddress, &now
	if s.Trust {
		exp := now.Add(repositories.TrustedDeviceTTL)
		d.SessionOnly, d.ExpiresAt = false, &exp
	}
	cp := *d
	return &cp, nil
}

func (m *memDevices) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.rows)
}

type enabledMFARepo struct {
	repositories.MFASetupRepositoryInterface
	row models.MFASetup
}

func (r enabledMFARepo) FindByUserID(id uuid.UUID) (*models.MFASetup, error) {
	if id != r.row.UserID {
		return nil, apperrors.ErrNotFound
	}
	cp := r.row
	return &cp, nil
}

type mfaLoginUsers struct {
	stubUserRepo
	user *models.User
}

func (r mfaLoginUsers) FindByID(id uuid.UUID) (*models.User, error) {
	if id != r.user.ID {
		return nil, apperrors.ErrNotFound
	}
	return r.user, nil
}

func (r mfaLoginUsers) FindByUsername(name string) (*models.User, error) {
	if name != r.user.Username {
		return nil, apperrors.ErrNotFound
	}
	return r.user, nil
}

func mustHash(t *testing.T, pw string) string {
	t.Helper()
	h, err := utils.HashPassword(pw)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

const mfaLoginSecret ="JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP"

func currentTOTP(t *testing.T, svc *AuthService) string {
	t.Helper()
	key, err := base32.StdEncoding.DecodeString(mfaLoginSecret)
	if err != nil {
		t.Fatal(err)
	}
	return svc.generateHOTP(key, time.Now().Unix()/30)
}

func newMFALoginService(t *testing.T) (*AuthService, *memDevices, *models.User, *redis.Client) {
	t.Helper()
	svc, _ := newClaimsTestService(t, 404, nil)
	user := claimsTestUser(nil)
	user.IsActive = true
	devices := &memDevices{}
	rdb := newFakeRedis(t)
	svc.redis = rdb
	svc.userRepo = mfaLoginUsers{user: user}
	svc.mfaRepo = enabledMFARepo{row: models.MFASetup{UserID: user.ID, MFAType: models.MFATypeTOTP, SecretKey: mfaLoginSecret, IsEnabled: true}}
	svc.trustedRepo = devices
	return svc, devices, user, rdb
}

func startMFALogin(t *testing.T, rdb *redis.Client, userID uuid.UUID) string {
	t.Helper()
	token := uuid.NewString()
	if err := rdb.Set(context.Background(), "mfa_login:"+token, userID.String(), time.Minute).Err(); err != nil {
		t.Fatal(err)
	}
	return token
}

func TestVerifyMFALogin_RegistersDeviceWithoutDuplicates(t *testing.T) {
	svc, devices, user, rdb := newMFALoginService(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		req := &models.VerifyMFALoginRequest{
			MFAToken:          startMFALogin(t, rdb, user.ID),
			Code:              currentTOTP(t, svc),
			DeviceFingerprint: "fp-browser",
			DeviceName:        "Chrome on macOS",
			DeviceType:        "desktop",
		}
		resp, err := svc.VerifyMFALogin(ctx, req, "203.0.113.7", "Mozilla/5.0")
		if err != nil {
			t.Fatalf("login %d: %v", i, err)
		}
		if resp.Token == "" {
			t.Fatalf("login %d: no token", i)
		}
	}
	if n := devices.count(); n != 1 {
		t.Fatalf("device rows = %d, want 1 after repeat logins", n)
	}
	d, _ := devices.FindByFingerprint(user.ID, "fp-browser")
	if !d.SessionOnly || d.LastSeenAt == nil || d.IPAddress != "203.0.113.7" || d.UserAgent != "Mozilla/5.0" {
		t.Fatalf("device = %+v, want a session-only row with ip, user agent and last seen", d)
	}
}

func TestVerifyMFALogin_SessionOnlyDeviceStillNeedsMFA(t *testing.T) {
	svc, devices, user, rdb := newMFALoginService(t)
	ctx := context.Background()
	user.PasswordHash = mustHash(t, "correct-horse")

	_, err := svc.VerifyMFALogin(ctx, &models.VerifyMFALoginRequest{
		MFAToken: startMFALogin(t, rdb, user.ID), Code: currentTOTP(t, svc), DeviceFingerprint: "fp-browser",
	}, "ip", "UA")
	if err != nil {
		t.Fatal(err)
	}

	resp, err := svc.Login(ctx, &models.LoginRequest{Username: user.Username, Password: "correct-horse", DeviceFingerprint: "fp-browser"}, "ip", "UA")
	if err != nil {
		t.Fatal(err)
	}
	if !resp.MFARequired {
		t.Fatal("a device recorded without \"trust this device\" skipped MFA")
	}
	if devices.count() != 1 {
		t.Fatalf("device rows = %d, want 1", devices.count())
	}
}

func TestVerifyMFALogin_TrustDeviceUpgradesSameRowAndSkipsMFA(t *testing.T) {
	svc, devices, user, rdb := newMFALoginService(t)
	ctx := context.Background()
	user.PasswordHash = mustHash(t, "correct-horse")

	for _, trust := range []bool{false, true} {
		if _, err := svc.VerifyMFALogin(ctx, &models.VerifyMFALoginRequest{
			MFAToken: startMFALogin(t, rdb, user.ID), Code: currentTOTP(t, svc), DeviceFingerprint: "fp-browser", TrustDevice: trust,
		}, "ip", "UA"); err != nil {
			t.Fatal(err)
		}
	}
	if devices.count() != 1 {
		t.Fatalf("device rows = %d, want 1", devices.count())
	}

	resp, err := svc.Login(ctx, &models.LoginRequest{Username: user.Username, Password: "correct-horse", DeviceFingerprint: "fp-browser"}, "ip", "UA2")
	if err != nil {
		t.Fatal(err)
	}
	if resp.MFARequired || resp.Token == "" {
		t.Fatal("trusted device did not skip MFA")
	}
	if d, _ := devices.FindByFingerprint(user.ID, "fp-browser"); d.UserAgent != "UA2" {
		t.Fatal("trusted sign-in did not refresh the device's last seen details")
	}
	if devices.count() != 1 {
		t.Fatalf("device rows = %d, want 1", devices.count())
	}
}

func TestVerifyMFALogin_NoFingerprintRecordsNothing(t *testing.T) {
	svc, devices, user, rdb := newMFALoginService(t)
	if _, err := svc.VerifyMFALogin(context.Background(), &models.VerifyMFALoginRequest{
		MFAToken: startMFALogin(t, rdb, user.ID), Code: currentTOTP(t, svc),
	}, "ip", "UA"); err != nil {
		t.Fatal(err)
	}
	if devices.count() != 0 {
		t.Fatal("recorded a device without a fingerprint")
	}
}
