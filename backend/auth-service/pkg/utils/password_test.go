package utils

import (
	"strings"
	"testing"
)

func TestGenerateTemporaryPassword(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 200; i++ {
		pw, err := GenerateTemporaryPassword()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pw) != tempPasswordLength {
			t.Fatalf("expected length %d, got %d (%q)", tempPasswordLength, len(pw), pw)
		}
		if !strings.ContainsAny(pw, tempPasswordUpper) || !strings.ContainsAny(pw, tempPasswordLower) || !strings.ContainsAny(pw, tempPasswordDigits) {
			t.Fatalf("password %q is missing a required character class", pw)
		}
		if seen[pw] {
			t.Fatalf("duplicate password generated: %q", pw)
		}
		seen[pw] = true
	}
}
