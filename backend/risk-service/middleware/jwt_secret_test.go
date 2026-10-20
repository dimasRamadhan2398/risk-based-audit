package middleware

import (
	"strings"
	"testing"
)

func TestNewAuthMiddlewareRefusesWeakSecrets(t *testing.T) {
	for _, secret := range []string{"", "   ", strings.Repeat("a", 63)} {
		if _, err := NewAuthMiddleware(secret); err == nil {
			t.Fatalf("secret of length %d must be refused", len(secret))
		}
	}
	if _, err := NewAuthMiddleware(strings.Repeat("a", 64)); err != nil {
		t.Fatalf("64-byte secret should be accepted: %v", err)
	}
}
