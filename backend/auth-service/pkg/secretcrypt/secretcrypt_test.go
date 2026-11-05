package secretcrypt

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
)

func randKey(t *testing.T) []byte {
	t.Helper()
	k := make([]byte, KeySize)
	if _, err := rand.Read(k); err != nil {
		t.Fatal(err)
	}
	return k
}

func mustCipher(t *testing.T, primary []byte, old ...[]byte) *Cipher {
	t.Helper()
	c, err := New(primary, old...)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

const totpSecret = "JBSWY3DPEHPK3PXPJBSWY3DPEHPK3PXP" // 20-byte seed, base32

var aad = []byte("mfa_setups.secret_key:00000000-0000-0000-0000-000000000001")

func TestRoundTrip(t *testing.T) {
	c := mustCipher(t, randKey(t))
	enc, err := c.Encrypt(totpSecret, aad)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(enc, PrefixV1) || strings.Contains(enc, totpSecret) {
		t.Fatalf("stored value %q is not an enc:v1 ciphertext", enc)
	}
	got, rewrite, err := c.Open(enc, aad)
	if err != nil || got != totpSecret || rewrite {
		t.Fatalf("Open = %q, rewrite=%v, err=%v", got, rewrite, err)
	}
}

func TestNonceIsRandomPerEncryption(t *testing.T) {
	c := mustCipher(t, randKey(t))
	a, _ := c.Encrypt(totpSecret, aad)
	b, _ := c.Encrypt(totpSecret, aad)
	if a == b {
		t.Fatal("two encryptions of the same secret are identical: nonce reused")
	}
}

func TestStoredLengthFitsVarchar255(t *testing.T) {
	c := mustCipher(t, randKey(t))
	enc, _ := c.Encrypt(totpSecret, aad)
	if len(enc) != 87 {
		t.Fatalf("len = %d, want 87 for a 32-char TOTP secret", len(enc))
	}
	// Largest plaintext that still fits: 255 - len(prefix) = 248 base64 chars = 186 bytes - 28 overhead.
	enc, _ = c.Encrypt(strings.Repeat("A", 158), aad)
	if len(enc) > 255 {
		t.Fatalf("158-byte secret encodes to %d chars, over varchar(255)", len(enc))
	}
}

func TestTamperDetection(t *testing.T) {
	c := mustCipher(t, randKey(t))
	enc, _ := c.Encrypt(totpSecret, aad)
	raw, _ := base64.StdEncoding.DecodeString(enc[len(PrefixV1):])
	for i := range raw {
		mod := bytes.Clone(raw)
		mod[i] ^= 0x01
		tampered := PrefixV1 + base64.StdEncoding.EncodeToString(mod)
		if _, err := c.Decrypt(tampered, aad); !errors.Is(err, ErrDecrypt) {
			t.Fatalf("byte %d flipped: err = %v, want ErrDecrypt", i, err)
		}
	}
	if _, err := c.Decrypt(PrefixV1+base64.StdEncoding.EncodeToString(raw[:len(raw)-1]), aad); !errors.Is(err, ErrDecrypt) {
		t.Fatalf("truncated: err = %v, want ErrDecrypt", err)
	}
	for _, bad := range []string{PrefixV1 + "!!!", PrefixV1, PrefixV1 + "AAAA", "enc:v9:" + enc[len(PrefixV1):]} {
		if _, err := c.Decrypt(bad, aad); !errors.Is(err, ErrFormat) {
			t.Fatalf("Decrypt(%q) err = %v, want ErrFormat", bad, err)
		}
	}
}

func TestAADBindsCiphertextToOwner(t *testing.T) {
	c := mustCipher(t, randKey(t))
	enc, _ := c.Encrypt(totpSecret, aad)
	if _, err := c.Decrypt(enc, []byte("mfa_setups.secret_key:someone-else")); !errors.Is(err, ErrDecrypt) {
		t.Fatalf("other owner: err = %v, want ErrDecrypt", err)
	}
}

func TestLegacyPlaintextPassthrough(t *testing.T) {
	c := mustCipher(t, randKey(t))
	got, rewrite, err := c.Open(totpSecret, aad)
	if err != nil || got != totpSecret || !rewrite {
		t.Fatalf("legacy with key: %q rewrite=%v err=%v, want passthrough + rewrite", got, rewrite, err)
	}
	var none *Cipher
	got, rewrite, err = none.Open(totpSecret, aad)
	if err != nil || got != totpSecret || rewrite {
		t.Fatalf("legacy without key: %q rewrite=%v err=%v", got, rewrite, err)
	}
	if got, rewrite, err := c.Open("", aad); err != nil || got != "" || rewrite {
		t.Fatalf("empty: %q rewrite=%v err=%v", got, rewrite, err)
	}
}

func TestWrongKey(t *testing.T) {
	enc, _ := mustCipher(t, randKey(t)).Encrypt(totpSecret, aad)
	if _, err := mustCipher(t, randKey(t)).Decrypt(enc, aad); !errors.Is(err, ErrDecrypt) {
		t.Fatalf("wrong key: err = %v, want ErrDecrypt", err)
	}
}

func TestNilCipher(t *testing.T) {
	var none *Cipher
	if none.Enabled() {
		t.Fatal("nil cipher reports enabled")
	}
	if got, err := none.Encrypt(totpSecret, aad); err != nil || got != totpSecret {
		t.Fatalf("nil Encrypt = %q, %v; want plaintext", got, err)
	}
	enc, _ := mustCipher(t, randKey(t)).Encrypt(totpSecret, aad)
	if _, err := none.Decrypt(enc, aad); !errors.Is(err, ErrNoKey) {
		t.Fatalf("nil Decrypt of ciphertext: err = %v, want ErrNoKey (fail closed)", err)
	}
}

func TestEmptyStaysEmpty(t *testing.T) {
	if got, err := mustCipher(t, randKey(t)).Encrypt("", aad); err != nil || got != "" {
		t.Fatalf("Encrypt(\"\") = %q, %v", got, err)
	}
}

func TestRotationOldKeyDecryptsAndAsksForRewrite(t *testing.T) {
	oldKey, newKey := randKey(t), randKey(t)
	enc, _ := mustCipher(t, oldKey).Encrypt(totpSecret, aad)
	rotated := mustCipher(t, newKey, oldKey)
	got, rewrite, err := rotated.Open(enc, aad)
	if err != nil || got != totpSecret || !rewrite {
		t.Fatalf("old-key value: %q rewrite=%v err=%v", got, rewrite, err)
	}
	fresh, _ := rotated.Encrypt(totpSecret, aad)
	if _, err := mustCipher(t, newKey).Decrypt(fresh, aad); err != nil {
		t.Fatalf("new writes must use the primary key: %v", err)
	}
	if _, err := mustCipher(t, oldKey).Decrypt(fresh, aad); err == nil {
		t.Fatal("new write decrypts with the old key")
	}
}

func TestParseKey(t *testing.T) {
	k := randKey(t)
	for name, s := range map[string]string{
		"std":        base64.StdEncoding.EncodeToString(k),
		"raw std":    base64.RawStdEncoding.EncodeToString(k),
		"url":        base64.URLEncoding.EncodeToString(k),
		"whitespace": " " + base64.StdEncoding.EncodeToString(k) + "\n",
	} {
		got, err := ParseKey(s)
		if err != nil || !bytes.Equal(got, k) {
			t.Fatalf("%s: err=%v", name, err)
		}
	}
	for name, s := range map[string]string{
		"empty":    "",
		"not b64":  "not base64 at all!!",
		"16 bytes": base64.StdEncoding.EncodeToString(k[:16]),
		"64 hex":   "6958c455ba7378688fac7200001d6bca8048fa66a28447ecc69f3b74b7ded4fa",
		"zeros":    base64.StdEncoding.EncodeToString(make([]byte, KeySize)),
	} {
		if _, err := ParseKey(s); err == nil {
			t.Fatalf("%s: accepted", name)
		} else if s != "" && strings.Contains(err.Error(), s) {
			t.Fatalf("%s: error leaks the key", name)
		}
	}
}

func TestNewRejectsBadKeyLength(t *testing.T) {
	if _, err := New(make([]byte, 16)); err == nil {
		t.Fatal("16-byte primary accepted")
	}
	if _, err := New(randKey(t), make([]byte, 31)); err == nil {
		t.Fatal("31-byte old key accepted")
	}
}
