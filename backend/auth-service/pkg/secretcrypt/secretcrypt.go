// Package secretcrypt encrypts small secrets (MFA TOTP seeds) for storage at rest.
//
// Stored format: "enc:v1:" + base64(nonce || ciphertext || tag), AES-256-GCM
// with a random 12-byte nonce per encryption. A value without the "enc:" prefix
// is legacy plaintext and is returned unchanged by Decrypt, so rows written
// before encryption keep working until the migration rewrites them.
//
// Rotation: a Cipher holds one primary key (used for every encryption) and
// optional previous keys that are only tried on decryption. GCM authenticates,
// so the key that opens a value is unambiguous. Re-encrypting old-key rows with
// the primary key is done by `auth migrate encrypt-mfa-secrets`.
package secretcrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

// KeySize is the required key length in bytes (AES-256).
const KeySize = 32

// PrefixV1 marks a value encrypted with format version 1.
const PrefixV1 = "enc:v1:"

// encPrefix is common to every encrypted format version.
const encPrefix = "enc:"

const nonceSize = 12

var (
	// ErrNoKey is returned when an encrypted value is read or a value must be
	// encrypted but no key is configured.
	ErrNoKey = errors.New("secretcrypt: no encryption key configured")
	// ErrDecrypt is returned when no configured key opens the value: wrong key,
	// tampered ciphertext, or a value bound to different associated data.
	ErrDecrypt = errors.New("secretcrypt: value cannot be decrypted with the configured keys")
	// ErrFormat is returned for a value that has the "enc:" prefix but is not a
	// well-formed v1 ciphertext.
	ErrFormat = errors.New("secretcrypt: malformed encrypted value")
)

// Cipher encrypts with its primary key and decrypts with any of its keys.
// A nil *Cipher is valid: it stores plaintext and refuses encrypted values.
type Cipher struct {
	primary cipher.AEAD
	old     []cipher.AEAD
}

// ParseKey decodes a base64 key (standard or URL alphabet, padded or not) and
// checks it is exactly KeySize bytes. The error never contains the key.
func ParseKey(encoded string) ([]byte, error) {
	s := strings.TrimSpace(encoded)
	if s == "" {
		return nil, errors.New("key is empty")
	}
	var key []byte
	var err error
	for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding} {
		if key, err = enc.DecodeString(s); err == nil {
			break
		}
	}
	if err != nil {
		return nil, errors.New("key is not valid base64")
	}
	if len(key) != KeySize {
		return nil, fmt.Errorf("key decodes to %d bytes, need %d (generate with: openssl rand -base64 32)", len(key), KeySize)
	}
	allSame := true
	for _, b := range key[1:] {
		if b != key[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return nil, errors.New("key is a constant byte pattern: generate a random key")
	}
	return key, nil
}

// New builds a Cipher. primary encrypts; old keys are tried after primary on
// decryption. Every key must be KeySize bytes.
func New(primary []byte, old ...[]byte) (*Cipher, error) {
	p, err := newAEAD(primary)
	if err != nil {
		return nil, fmt.Errorf("primary key: %w", err)
	}
	c := &Cipher{primary: p}
	for i, k := range old {
		a, err := newAEAD(k)
		if err != nil {
			return nil, fmt.Errorf("old key %d: %w", i+1, err)
		}
		c.old = append(c.old, a)
	}
	return c, nil
}

func newAEAD(key []byte) (cipher.AEAD, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("key must be %d bytes, got %d", KeySize, len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCMWithNonceSize(block, nonceSize)
}

// Enabled reports whether c can encrypt.
func (c *Cipher) Enabled() bool { return c != nil && c.primary != nil }

// IsEncrypted reports whether stored carries an encryption prefix.
func IsEncrypted(stored string) bool { return strings.HasPrefix(stored, encPrefix) }

// Encrypt seals plaintext with the primary key. aad binds the ciphertext to
// its context (e.g. the owning row) so it cannot be moved to another row.
// The empty string stays empty: there is nothing secret to protect.
// A nil Cipher returns plaintext unchanged (encryption not configured).
func (c *Cipher) Encrypt(plaintext string, aad []byte) (string, error) {
	if plaintext == "" || !c.Enabled() {
		return plaintext, nil
	}
	nonce := make([]byte, nonceSize, nonceSize+len(plaintext)+c.primary.Overhead())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("secretcrypt: nonce: %w", err)
	}
	sealed := c.primary.Seal(nonce, nonce, []byte(plaintext), aad)
	return PrefixV1 + base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt returns the plaintext of stored. Legacy plaintext (no "enc:" prefix)
// is returned unchanged.
func (c *Cipher) Decrypt(stored string, aad []byte) (string, error) {
	plain, _, err := c.Open(stored, aad)
	return plain, err
}

// Open is Decrypt that also reports whether stored should be rewritten with
// the primary key: true for legacy plaintext (when a key is configured) and for
// values that only an old key opens.
func (c *Cipher) Open(stored string, aad []byte) (plaintext string, needsRewrite bool, err error) {
	if !IsEncrypted(stored) {
		return stored, stored != "" && c.Enabled(), nil
	}
	if !c.Enabled() {
		return "", false, ErrNoKey
	}
	if !strings.HasPrefix(stored, PrefixV1) {
		return "", false, ErrFormat
	}
	raw, err := base64.StdEncoding.DecodeString(stored[len(PrefixV1):])
	if err != nil || len(raw) < nonceSize+c.primary.Overhead() {
		return "", false, ErrFormat
	}
	nonce, body := raw[:nonceSize], raw[nonceSize:]
	if out, err := c.primary.Open(nil, nonce, body, aad); err == nil {
		return string(out), false, nil
	}
	for _, a := range c.old {
		if out, err := a.Open(nil, nonce, body, aad); err == nil {
			return string(out), true, nil
		}
	}
	return "", false, ErrDecrypt
}
