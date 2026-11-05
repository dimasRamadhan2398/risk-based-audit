package repositories

import (
	"crypto/rand"
	"strings"
	"testing"

	"auth-service/models"
	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/secretcrypt"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const encTestSecret = "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP"

func encMFADB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.Exec(`CREATE TABLE mfa_setups (
		id TEXT PRIMARY KEY, user_id TEXT NOT NULL, mfa_type TEXT NOT NULL,
		secret_key TEXT NOT NULL, is_enabled NUMERIC DEFAULT false,
		created_at DATETIME, updated_at DATETIME)`).Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func encCipher(t *testing.T, old ...[]byte) (*secretcrypt.Cipher, []byte) {
	t.Helper()
	k := make([]byte, secretcrypt.KeySize)
	if _, err := rand.Read(k); err != nil {
		t.Fatal(err)
	}
	c, err := secretcrypt.New(k, old...)
	if err != nil {
		t.Fatal(err)
	}
	return c, k
}

func storedSecret(t *testing.T, db *gorm.DB, id uuid.UUID) string {
	t.Helper()
	var s string
	if err := db.Raw("SELECT secret_key FROM mfa_setups WHERE id = ?", id).Scan(&s).Error; err != nil {
		t.Fatal(err)
	}
	return s
}

func insertRaw(t *testing.T, db *gorm.DB, userID uuid.UUID, mfaType models.MFAType, secret string, enabled bool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if err := db.Exec("INSERT INTO mfa_setups (id, user_id, mfa_type, secret_key, is_enabled) VALUES (?, ?, ?, ?, ?)",
		id, userID, mfaType, secret, enabled).Error; err != nil {
		t.Fatal(err)
	}
	return id
}

func TestMFARepoEncryptsOnWriteDecryptsOnRead(t *testing.T) {
	db := encMFADB(t)
	c, _ := encCipher(t)
	repo := NewMFASetupRepositoryWithCipher(db, c)
	userID := uuid.New()
	row := &models.MFASetup{ID: uuid.New(), UserID: userID, MFAType: models.MFATypeTOTP, SecretKey: encTestSecret}
	if err := repo.Create(row); err != nil {
		t.Fatal(err)
	}
	if row.SecretKey != encTestSecret {
		t.Fatal("Create must leave the caller's struct in plaintext")
	}
	stored := storedSecret(t, db, row.ID)
	if !strings.HasPrefix(stored, secretcrypt.PrefixV1) || strings.Contains(stored, encTestSecret) {
		t.Fatalf("stored value is not encrypted: %q", stored)
	}
	for name, find := range map[string]func() (*models.MFASetup, error){
		"FindByID":            func() (*models.MFASetup, error) { return repo.FindByID(row.ID) },
		"FindByUserID":        func() (*models.MFASetup, error) { return repo.FindByUserID(userID) },
		"FindByUserIDAndType": func() (*models.MFASetup, error) { return repo.FindByUserIDAndType(userID, models.MFATypeTOTP) },
	} {
		got, err := find()
		if err != nil || got.SecretKey != encTestSecret {
			t.Fatalf("%s: secret=%q err=%v", name, got.SecretKey, err)
		}
	}

	// Update re-encrypts with a fresh nonce.
	row.SecretKey = "KRSXG5DSNFXGOIDBNZSCA3LFONZWCZ3F"
	if err := repo.Update(row); err != nil {
		t.Fatal(err)
	}
	if s := storedSecret(t, db, row.ID); s == stored || !strings.HasPrefix(s, secretcrypt.PrefixV1) {
		t.Fatalf("Update did not store a new ciphertext: %q", s)
	}
	if got, _ := repo.FindByUserID(userID); got.SecretKey != "KRSXG5DSNFXGOIDBNZSCA3LFONZWCZ3F" {
		t.Fatal("updated secret not read back")
	}
}

func TestMFARepoReadsLegacyPlaintext(t *testing.T) {
	db := encMFADB(t)
	c, _ := encCipher(t)
	userID := uuid.New()
	insertRaw(t, db, userID, models.MFATypeTOTP, encTestSecret, true)
	got, err := NewMFASetupRepositoryWithCipher(db, c).FindByUserID(userID)
	if err != nil || got.SecretKey != encTestSecret {
		t.Fatalf("legacy row: %q, %v", got.SecretKey, err)
	}
}

func TestMFARepoWrongKeyFailsClosed(t *testing.T) {
	db := encMFADB(t)
	c, _ := encCipher(t)
	other, _ := encCipher(t)
	userID := uuid.New()
	if err := NewMFASetupRepositoryWithCipher(db, c).Create(&models.MFASetup{ID: uuid.New(), UserID: userID, MFAType: models.MFATypeTOTP, SecretKey: encTestSecret, IsEnabled: true}); err != nil {
		t.Fatal(err)
	}
	for name, repo := range map[string]IMFASetupRepository{
		"wrong key": NewMFASetupRepositoryWithCipher(db, other),
		"no key":    NewMFASetupRepository(db),
	} {
		got, err := repo.FindByUserID(userID)
		if got != nil || !apperrors.Is(err, ErrMFASecretUnreadable) {
			t.Fatalf("%s: row=%v err=%v, want ErrMFASecretUnreadable", name, got, err)
		}
		if apperrors.Is(err, apperrors.ErrNotFound) {
			t.Fatalf("%s: unreadable must not look like not-found (login would skip MFA)", name)
		}
	}
}

func TestMFARepoRowSwapIsRejected(t *testing.T) {
	db := encMFADB(t)
	c, _ := encCipher(t)
	repo := NewMFASetupRepositoryWithCipher(db, c)
	attacker, victim := uuid.New(), uuid.New()
	a := &models.MFASetup{ID: uuid.New(), UserID: attacker, MFAType: models.MFATypeTOTP, SecretKey: encTestSecret, IsEnabled: true}
	if err := repo.Create(a); err != nil {
		t.Fatal(err)
	}
	insertRaw(t, db, victim, models.MFATypeTOTP, storedSecret(t, db, a.ID), true)
	if _, err := repo.FindByUserID(victim); !apperrors.Is(err, ErrMFASecretUnreadable) {
		t.Fatalf("ciphertext copied to another user decrypted: err=%v", err)
	}
}

func TestEnablePendingWithEncryption(t *testing.T) {
	db := encMFADB(t)
	c, _ := encCipher(t)
	repo := NewMFASetupRepositoryWithCipher(db, c)
	userID := uuid.New()
	if err := repo.Create(&models.MFASetup{ID: uuid.New(), UserID: userID, MFAType: models.MFATypeTOTP, SecretKey: encTestSecret}); err != nil {
		t.Fatal(err)
	}
	if ok, err := repo.EnablePending(userID, "REPLACED"); err != nil || ok {
		t.Fatalf("other secret: ok=%v err=%v", ok, err)
	}
	if ok, err := repo.EnablePending(userID, encTestSecret); err != nil || !ok {
		t.Fatalf("pending secret (plaintext vs stored ciphertext): ok=%v err=%v", ok, err)
	}
	if ok, err := repo.EnablePending(userID, encTestSecret); err != nil || ok {
		t.Fatalf("already enabled: ok=%v err=%v", ok, err)
	}
	// The service re-reads and compares plaintext; both reads must agree.
	a, _ := repo.FindByUserID(userID)
	b, _ := repo.FindByUserID(userID)
	if !a.IsEnabled || a.SecretKey != b.SecretKey || a.SecretKey != encTestSecret {
		t.Fatal("re-read row does not compare equal in plaintext")
	}
}

func TestEnablePendingLegacyRow(t *testing.T) {
	db := encMFADB(t)
	c, _ := encCipher(t)
	userID := uuid.New()
	insertRaw(t, db, userID, models.MFATypeTOTP, encTestSecret, false)
	if ok, err := NewMFASetupRepositoryWithCipher(db, c).EnablePending(userID, encTestSecret); err != nil || !ok {
		t.Fatalf("legacy pending row: ok=%v err=%v", ok, err)
	}
}

func TestEncryptMFASecretsMigration(t *testing.T) {
	db := encMFADB(t)
	oldC, oldKey := encCipher(t)
	legacyUser, emailUser, oldUser, curUser := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	legacyID := insertRaw(t, db, legacyUser, models.MFATypeTOTP, encTestSecret, true)
	emailID := insertRaw(t, db, emailUser, models.MFATypeEmail, "", true)
	oldEnc, _ := oldC.Encrypt(encTestSecret, mfaSecretAAD(oldUser))
	oldID := insertRaw(t, db, oldUser, models.MFATypeTOTP, oldEnc, true)

	c, newKey := encCipher(t, oldKey)
	if err := NewMFASetupRepositoryWithCipher(db, c).Create(&models.MFASetup{ID: uuid.New(), UserID: curUser, MFAType: models.MFATypeTOTP, SecretKey: encTestSecret}); err != nil {
		t.Fatal(err)
	}

	// Dry run: counts, writes nothing.
	res, err := EncryptMFASecrets(db, c, true)
	if err != nil || res.Legacy != 1 || res.Rotated != 1 || res.Current != 1 || res.Empty != 1 || res.Total != 4 {
		t.Fatalf("dry run: %+v err=%v", res, err)
	}
	if storedSecret(t, db, legacyID) != encTestSecret {
		t.Fatal("dry run wrote to the database")
	}

	res, err = EncryptMFASecrets(db, c, false)
	if err != nil || res.Changed() != 2 {
		t.Fatalf("first run: %+v err=%v", res, err)
	}
	if s := storedSecret(t, db, legacyID); !strings.HasPrefix(s, secretcrypt.PrefixV1) {
		t.Fatalf("legacy row not encrypted: %q", s)
	}
	if storedSecret(t, db, emailID) != "" {
		t.Fatal("empty email-MFA secret must stay empty")
	}
	if storedSecret(t, db, oldID) == oldEnc {
		t.Fatal("old-key row not re-encrypted")
	}

	// Idempotent: second run changes nothing.
	snapshot := map[uuid.UUID]string{legacyID: storedSecret(t, db, legacyID), oldID: storedSecret(t, db, oldID)}
	res, err = EncryptMFASecrets(db, c, false)
	if err != nil || res.Changed() != 0 || res.Current != 3 || res.Empty != 1 {
		t.Fatalf("second run: %+v err=%v", res, err)
	}
	for id, want := range snapshot {
		if storedSecret(t, db, id) != want {
			t.Fatal("second run rewrote a row")
		}
	}

	// After rotation completes, the new key alone reads every row.
	newOnly, _ := secretcrypt.New(newKey)
	repo := NewMFASetupRepositoryWithCipher(db, newOnly)
	for _, u := range []uuid.UUID{legacyUser, oldUser, curUser} {
		if got, err := repo.FindByUserID(u); err != nil || got.SecretKey != encTestSecret {
			t.Fatalf("user %s after migration: %v", u, err)
		}
	}
}

func TestEncryptMFASecretsRollsBackOnUnreadableRow(t *testing.T) {
	db := encMFADB(t)
	stranger, _ := encCipher(t)
	legacyID := insertRaw(t, db, uuid.New(), models.MFATypeTOTP, encTestSecret, true)
	u := uuid.New()
	enc, _ := stranger.Encrypt(encTestSecret, mfaSecretAAD(u))
	insertRaw(t, db, u, models.MFATypeTOTP, enc, true)

	c, _ := encCipher(t)
	res, err := EncryptMFASecrets(db, c, false)
	if err == nil || res.Unreadable != 1 {
		t.Fatalf("want error for unreadable row, got %+v err=%v", res, err)
	}
	if storedSecret(t, db, legacyID) != encTestSecret {
		t.Fatal("transaction not rolled back: legacy row was rewritten")
	}
}

func TestEncryptMFASecretsRequiresKey(t *testing.T) {
	if _, err := EncryptMFASecrets(encMFADB(t), nil, false); err == nil {
		t.Fatal("migration without a key must fail")
	}
}
