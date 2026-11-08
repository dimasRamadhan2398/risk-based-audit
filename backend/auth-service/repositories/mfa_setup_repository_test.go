package repositories

import (
	"testing"

	"auth-service/models"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func mfaDB(t *testing.T) *gorm.DB {
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

func TestEnablePending(t *testing.T) {
	db := mfaDB(t)
	repo := NewMFASetupRepository(db)
	userID := uuid.New()
	if err := repo.Create(&models.MFASetup{ID: uuid.New(), UserID: userID, MFAType: models.MFATypeTOTP, SecretKey: "PENDING"}); err != nil {
		t.Fatal(err)
	}

	if ok, err := repo.EnablePending(userID, "REPLACED"); err != nil || ok {
		t.Fatalf("other secret: ok=%v err=%v, want no switch", ok, err)
	}
	if row, _ := repo.FindByUserID(userID); row.IsEnabled {
		t.Fatal("enabled with a secret that is not the pending one")
	}

	if ok, err := repo.EnablePending(userID, "PENDING"); err != nil || !ok {
		t.Fatalf("pending secret: ok=%v err=%v, want switched on", ok, err)
	}
	if row, _ := repo.FindByUserID(userID); !row.IsEnabled {
		t.Fatal("not enabled")
	}

	// Second call: already on, nothing to switch.
	if ok, err := repo.EnablePending(userID, "PENDING"); err != nil || ok {
		t.Fatalf("already enabled: ok=%v err=%v, want no switch", ok, err)
	}
}
