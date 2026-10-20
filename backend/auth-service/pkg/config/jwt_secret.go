package config

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"
)

// JWTSecretEnv is the environment variable that carries the JWT signing secret.
// It overrides jwt.secret from the config file.
const JWTSecretEnv = "JWT_SECRET"

// MinJWTSecretBytes is the minimum accepted length of the JWT signing secret.
const MinJWTSecretBytes = 64

// knownPublicJWTSecretSHA256 holds SHA-256 digests (never the values) of JWT
// secrets that were committed to the repository. A deployment that uses one of
// them is treated as public and refused at start-up.
var knownPublicJWTSecretSHA256 = []string{
	// Literal that was committed in the config.yaml files (rotated, see docs/jwt-rotation-runbook.md).
	"18c6875abd3ebc5ad6964d3ba2e2f0894880df859728c8d72ac045b695c2fbdd",
	// Placeholder that was committed in auth-service/.env.example.
	"103ab5dd9769664c34bb4dcecdbe1aa52a55f75a4a62243238c9abb3bc3d9e02",
}

// ValidateJWTSecret refuses a JWT signing secret that is empty, shorter than
// MinJWTSecretBytes, or equal to a known public value. The returned error never
// contains the secret.
func ValidateJWTSecret(secret string) error {
	if strings.TrimSpace(secret) == "" {
		return fmt.Errorf("%s is not set: provide it through the environment", JWTSecretEnv)
	}
	if len(secret) < MinJWTSecretBytes {
		return fmt.Errorf("%s is too short: %d bytes, minimum is %d", JWTSecretEnv, len(secret), MinJWTSecretBytes)
	}
	sum := sha256.Sum256([]byte(secret))
	digest := []byte(hex.EncodeToString(sum[:]))
	for _, known := range knownPublicJWTSecretSHA256 {
		if subtle.ConstantTimeCompare(digest, []byte(known)) == 1 {
			return fmt.Errorf("%s matches a publicly known value: generate a new secret", JWTSecretEnv)
		}
	}
	return nil
}

// ValidateJWT checks the loaded JWT signing secret. Serve commands call it before
// binding a port.
func (c *Config) ValidateJWT() error {
	return ValidateJWTSecret(c.JWT.Secret)
}
