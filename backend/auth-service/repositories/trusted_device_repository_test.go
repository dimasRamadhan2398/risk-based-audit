package repositories

import (
	"testing"
	"time"

	"auth-service/models"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// devicesDB is an in-memory sqlite DB with trusted_devices created by hand
// (the model's gen_random_uuid() default is Postgres-only).
func devicesDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.Exec(`CREATE TABLE trusted_devices (
		id TEXT PRIMARY KEY, user_id TEXT NOT NULL, device_fingerprint TEXT NOT NULL,
		device_name TEXT, device_type TEXT, user_agent TEXT, ip_address TEXT,
		expires_at DATETIME, session_only NUMERIC NOT NULL DEFAULT false,
		last_seen_at DATETIME, created_at DATETIME)`).Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func countDevices(t *testing.T, db *gorm.DB, userID uuid.UUID) int64 {
	t.Helper()
	var n int64
	if err := db.Model(&models.TrustedDevice{}).Where("user_id = ?", userID).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

const chromeMacUA = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0 Safari/537.36"

func TestRecordSeen_NewDeviceIsSessionOnlyAndListed(t *testing.T) {
	db := devicesDB(t)
	repo := NewTrustedDeviceRepository(db)
	userID := uuid.New()

	d, err := repo.RecordSeen(DeviceSighting{UserID: userID, Fingerprint: "fp1", UserAgent: chromeMacUA, IPAddress: "203.0.113.7, 10.0.0.2"})
	if err != nil {
		t.Fatal(err)
	}
	if !d.SessionOnly || d.ExpiresAt != nil {
		t.Fatalf("new untrusted device = %+v, want session-only without expiry", d)
	}
	if repo.IsTrusted(d) {
		t.Fatal("session-only device would skip MFA")
	}
	if d.IPAddress != "203.0.113.7" {
		t.Fatalf("ip = %q, want the client address of the X-Forwarded-For chain", d.IPAddress)
	}
	if d.DeviceName != "Chrome on macOS" || d.DeviceType != "desktop" {
		t.Fatalf("name/type = %q/%q, want UA fallback", d.DeviceName, d.DeviceType)
	}
	if d.LastSeenAt == nil {
		t.Fatal("last_seen_at not set")
	}

	listed, err := repo.FindByUserID(userID)
	if err != nil || len(listed) != 1 || !listed[0].SessionOnly {
		t.Fatalf("FindByUserID = %+v, %v", listed, err)
	}
}

func TestRecordSeen_RepeatSightingsKeepOneRow(t *testing.T) {
	db := devicesDB(t)
	repo := NewTrustedDeviceRepository(db)
	userID := uuid.New()
	t0 := time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC)

	first, err := repo.RecordSeen(DeviceSighting{UserID: userID, Fingerprint: "fp1", Name: "Chrome on macOS", IPAddress: "198.51.100.1", Now: t0})
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 3; i++ {
		if _, err := repo.RecordSeen(DeviceSighting{UserID: userID, Fingerprint: "fp1", IPAddress: "198.51.100.9", Now: t0.Add(time.Duration(i) * time.Hour)}); err != nil {
			t.Fatal(err)
		}
	}
	if n := countDevices(t, db, userID); n != 1 {
		t.Fatalf("rows = %d, want 1", n)
	}
	got, err := repo.FindByFingerprint(userID, "fp1")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != first.ID {
		t.Fatal("row replaced instead of refreshed")
	}
	if got.IPAddress != "198.51.100.9" || !got.LastSeenAt.Equal(t0.Add(3*time.Hour)) {
		t.Fatalf("not refreshed: ip=%q last_seen=%v", got.IPAddress, got.LastSeenAt)
	}
	if got.DeviceName != "Chrome on macOS" {
		t.Fatalf("name = %q, a sighting without a name must keep it", got.DeviceName)
	}

	// Another device and another user each get their own row.
	if _, err := repo.RecordSeen(DeviceSighting{UserID: userID, Fingerprint: "fp2"}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.RecordSeen(DeviceSighting{UserID: uuid.New(), Fingerprint: "fp1"}); err != nil {
		t.Fatal(err)
	}
	if n := countDevices(t, db, userID); n != 2 {
		t.Fatalf("rows for user = %d, want 2", n)
	}
}

func TestRecordSeen_TrustUpgradesAndIsNotDowngraded(t *testing.T) {
	db := devicesDB(t)
	repo := NewTrustedDeviceRepository(db)
	userID := uuid.New()
	now := time.Now()

	if _, err := repo.RecordSeen(DeviceSighting{UserID: userID, Fingerprint: "fp1", Now: now}); err != nil {
		t.Fatal(err)
	}
	trusted, err := repo.RecordSeen(DeviceSighting{UserID: userID, Fingerprint: "fp1", Trust: true, Now: now})
	if err != nil {
		t.Fatal(err)
	}
	reloaded, _ := repo.FindByFingerprint(userID, "fp1")
	if trusted.SessionOnly || reloaded.SessionOnly || !repo.IsTrusted(reloaded) {
		t.Fatalf("trust not persisted: %+v", reloaded)
	}
	if reloaded.ExpiresAt == nil || reloaded.ExpiresAt.Sub(now) < TrustedDeviceTTL-time.Minute {
		t.Fatalf("expires_at = %v, want ~now+%v", reloaded.ExpiresAt, TrustedDeviceTTL)
	}

	// A later sighting without "trust this device" keeps the trust.
	if _, err := repo.RecordSeen(DeviceSighting{UserID: userID, Fingerprint: "fp1", Now: now.Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	reloaded, _ = repo.FindByFingerprint(userID, "fp1")
	if !repo.IsTrusted(reloaded) {
		t.Fatal("untrusted sighting downgraded a trusted device")
	}
	if n := countDevices(t, db, userID); n != 1 {
		t.Fatalf("rows = %d, want 1", n)
	}
}

func TestRecordSeen_RequiresFingerprint(t *testing.T) {
	repo := NewTrustedDeviceRepository(devicesDB(t))
	if _, err := repo.RecordSeen(DeviceSighting{UserID: uuid.New(), Fingerprint: "  "}); err == nil {
		t.Fatal("blank fingerprint accepted")
	}
}

func TestIsTrusted(t *testing.T) {
	repo := &TrustedDeviceRepository{}
	past, future := time.Now().Add(-time.Hour), time.Now().Add(time.Hour)
	for _, tc := range []struct {
		name string
		d    *models.TrustedDevice
		want bool
	}{
		{"nil", nil, false},
		{"legacy trusted, no expiry", &models.TrustedDevice{}, true},
		{"trusted, not expired", &models.TrustedDevice{ExpiresAt: &future}, true},
		{"trusted, expired", &models.TrustedDevice{ExpiresAt: &past}, false},
		{"session only", &models.TrustedDevice{SessionOnly: true}, false},
	} {
		if got := repo.IsTrusted(tc.d); got != tc.want {
			t.Errorf("%s: IsTrusted = %v, want %v", tc.name, got, tc.want)
		}
	}
}
