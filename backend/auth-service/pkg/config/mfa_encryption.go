package config

import (
	"fmt"
	"strings"

	"auth-service/pkg/secretcrypt"
)

// MFA secret encryption at rest (mfa_setups.secret_key).
//
//	MFA_ENCRYPTION_KEY       32 random bytes, base64 (openssl rand -base64 32).
//	                         Encrypts every newly written TOTP secret. Each
//	                         deployment/tenant needs its own key; never reuse
//	                         or derive it from JWT_SECRET.
//	MFA_ENCRYPTION_KEY_OLD   optional previous key(s), comma separated, only
//	                         used to decrypt during a rotation.
//	MFA_ENCRYPTION_REQUIRED  "true" makes a missing key fatal at start-up.
//	                         Implied when APP_APP_ENV=production.
const (
	MFAEncryptionKeyEnv      = "MFA_ENCRYPTION_KEY"
	MFAEncryptionKeyOldEnv   = "MFA_ENCRYPTION_KEY_OLD"
	MFAEncryptionRequiredEnv = "MFA_ENCRYPTION_REQUIRED"
)

// MFAConfig holds the MFA secret encryption settings. Keys are read from the
// environment (bound in Load); they do not belong in config.yaml.
type MFAConfig struct {
	EncryptionKey      string `mapstructure:"encryption_key"`
	EncryptionKeyOld   string `mapstructure:"encryption_key_old"`
	EncryptionRequired bool   `mapstructure:"encryption_required"`
}

// MFAEncryptionRequired reports whether a missing MFA_ENCRYPTION_KEY must stop
// the service: explicitly via MFA_ENCRYPTION_REQUIRED, or in production.
func (c *Config) MFAEncryptionRequired() bool {
	if c.MFA.EncryptionRequired {
		return true
	}
	env := strings.ToLower(strings.TrimSpace(c.App.AppEnv))
	return env == "production" || env == "prod"
}

// MFACipher builds the cipher for MFA secrets. It returns (nil, nil) when no
// key is set and encryption is not required: secrets are then stored as
// plaintext and legacy rows keep working. An invalid key is always an error,
// as is a missing key when encryption is required. Errors never contain keys.
func (c *Config) MFACipher() (*secretcrypt.Cipher, error) {
	return NewMFACipher(c.MFA.EncryptionKey, c.MFA.EncryptionKeyOld, c.MFAEncryptionRequired())
}

// NewMFACipher is MFACipher on explicit values (for tests and tools).
func NewMFACipher(key, oldKeys string, required bool) (*secretcrypt.Cipher, error) {
	if strings.TrimSpace(key) == "" {
		if strings.TrimSpace(oldKeys) != "" {
			return nil, fmt.Errorf("%s is set but %s is not: set the new key as %s", MFAEncryptionKeyOldEnv, MFAEncryptionKeyEnv, MFAEncryptionKeyEnv)
		}
		if required {
			return nil, fmt.Errorf("%s is not set: provide a base64 32-byte key (openssl rand -base64 32) through the environment", MFAEncryptionKeyEnv)
		}
		return nil, nil
	}
	primary, err := secretcrypt.ParseKey(key)
	if err != nil {
		return nil, fmt.Errorf("%s is invalid: %v", MFAEncryptionKeyEnv, err)
	}
	var old [][]byte
	for i, part := range strings.Split(oldKeys, ",") {
		if strings.TrimSpace(part) == "" {
			continue
		}
		k, err := secretcrypt.ParseKey(part)
		if err != nil {
			return nil, fmt.Errorf("%s entry %d is invalid: %v", MFAEncryptionKeyOldEnv, i+1, err)
		}
		old = append(old, k)
	}
	return secretcrypt.New(primary, old...)
}

// ValidateMFAEncryption checks the MFA encryption settings. Serve commands call
// it before binding a port, next to ValidateJWT.
func (c *Config) ValidateMFAEncryption() error {
	_, err := c.MFACipher()
	return err
}
