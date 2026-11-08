package config

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"
)

func testMFAKey(t *testing.T) string {
	t.Helper()
	k := make([]byte, 32)
	if _, err := rand.Read(k); err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(k)
}

func TestNewMFACipher(t *testing.T) {
	key, old := testMFAKey(t), testMFAKey(t)
	cases := []struct {
		name        string
		key, old    string
		required    bool
		wantErr     bool
		wantEnabled bool
	}{
		{"missing, not required: plaintext mode", "", "", false, false, false},
		{"missing, required: fatal", "", "", true, true, false},
		{"valid", key, "", true, false, true},
		{"valid with old keys", key, old + "," + testMFAKey(t), true, false, true},
		{"invalid key fails even when not required", "short", "", false, true, false},
		{"hex JWT-style secret is not a 32-byte base64 key", strings.Repeat("ab", 32), "", false, true, false},
		{"invalid old key", key, "nope", false, true, false},
		{"old key without new key", "", old, false, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := NewMFACipher(tc.key, tc.old, tc.required)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if c.Enabled() != tc.wantEnabled {
				t.Fatalf("Enabled() = %v, want %v", c.Enabled(), tc.wantEnabled)
			}
			if err != nil {
				for _, secret := range []string{tc.key, tc.old} {
					if len(secret) > 8 && strings.Contains(err.Error(), secret) {
						t.Fatal("error message leaks a key")
					}
				}
			}
		})
	}
}

func TestLoadMFAEncryptionFromEnv(t *testing.T) {
	path := writeConfig(t, "jwt:\n  expiry_hour: 24\napp:\n  app_env: development\n")
	key, old := testMFAKey(t), testMFAKey(t)
	t.Setenv(MFAEncryptionKeyEnv, key)
	t.Setenv(MFAEncryptionKeyOldEnv, old)
	t.Setenv(MFAEncryptionRequiredEnv, "true")

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MFA.EncryptionKey != key || cfg.MFA.EncryptionKeyOld != old || !cfg.MFA.EncryptionRequired {
		t.Fatalf("MFA env not bound: %+v", cfg.MFA.EncryptionRequired)
	}
	if err := cfg.ValidateMFAEncryption(); err != nil {
		t.Fatal(err)
	}
}

func TestMFAEncryptionRequiredInProduction(t *testing.T) {
	path := writeConfig(t, "jwt:\n  expiry_hour: 24\napp:\n  app_env: development\n")
	t.Setenv(MFAEncryptionKeyEnv, "")
	t.Setenv(MFAEncryptionRequiredEnv, "")
	t.Setenv("APP_APP_ENV", "production")

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.MFAEncryptionRequired() {
		t.Fatal("APP_APP_ENV=production must require the MFA key")
	}
	if err := cfg.ValidateMFAEncryption(); err == nil {
		t.Fatal("missing MFA_ENCRYPTION_KEY in production must fail start-up")
	}
}

func TestMFAEncryptionOptionalInDevelopment(t *testing.T) {
	path := writeConfig(t, "jwt:\n  expiry_hour: 24\napp:\n  app_env: development\n")
	t.Setenv(MFAEncryptionKeyEnv, "")
	t.Setenv(MFAEncryptionRequiredEnv, "")
	t.Setenv("APP_APP_ENV", "development")

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.ValidateMFAEncryption(); err != nil {
		t.Fatalf("development without a key should start (plaintext mode): %v", err)
	}
}
