package config

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testSecret is a synthetic value used only by these tests. It is not a real key.
var testSecret = "test-" + strings.Repeat("x", 64)

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestValidateJWTSecret(t *testing.T) {
	cases := []struct {
		name    string
		secret  string
		wantErr bool
	}{
		{"empty", "", true},
		{"whitespace only", "   ", true},
		{"too short", strings.Repeat("a", 63), true},
		{"exactly 64 bytes", strings.Repeat("a", 64), false},
		{"long enough", testSecret, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateJWTSecret(tc.secret)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateJWTSecret() error = %v, wantErr %v", err, tc.wantErr)
			}
			if err != nil && tc.secret != "" && strings.Contains(err.Error(), tc.secret) {
				t.Fatal("error message must not contain the secret")
			}
		})
	}
}

func TestLoadJWTSecretEnvOverridesConfig(t *testing.T) {
	path := writeConfig(t, "jwt:\n  secret: \"value-from-file\"\n  expiry_hour: 24\n")
	t.Setenv("JWT_SECRET", testSecret)

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.JWT.Secret != testSecret {
		t.Fatal("JWT_SECRET from the environment did not override jwt.secret from the file")
	}
	if err := cfg.ValidateJWT(); err != nil {
		t.Fatalf("env secret should validate: %v", err)
	}
}

func TestLoadJWTSecretFromEnvWhenFileHasNoSecretKey(t *testing.T) {
	// Without BindEnv viper would drop the key entirely, because it is not in the file.
	path := writeConfig(t, "jwt:\n  expiry_hour: 24\n")
	t.Setenv("JWT_SECRET", testSecret)

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.JWT.Secret != testSecret {
		t.Fatal("JWT_SECRET was not read when the file has no jwt.secret key")
	}
}

func TestLoadEmptyJWTSecretFailsValidation(t *testing.T) {
	path := writeConfig(t, "jwt:\n  secret: \"\"\n  expiry_hour: 24\n")
	t.Setenv("JWT_SECRET", "")

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.ValidateJWT(); err == nil {
		t.Fatal("an empty JWT_SECRET must fail validation")
	}
}

// TestValidateJWTSecretRefusesKnownPublicDigest checks the blocklist mechanism.
// The list holds digests only, so the test adds the digest of a synthetic value
// to it for the duration of the test rather than storing any real secret.
func TestValidateJWTSecretRefusesKnownPublicDigest(t *testing.T) {
	sum := sha256.Sum256([]byte(testSecret))
	saved := knownPublicJWTSecretSHA256
	knownPublicJWTSecretSHA256 = append([]string{hex.EncodeToString(sum[:])}, saved...)
	t.Cleanup(func() { knownPublicJWTSecretSHA256 = saved })

	if err := ValidateJWTSecret(testSecret); err == nil {
		t.Fatal("a secret whose digest is on the known-public list must be refused")
	}
}
