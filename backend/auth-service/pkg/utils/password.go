package utils

import (
	"crypto/rand"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword checks if a password matches a hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

const (
	tempPasswordUpper  = "ABCDEFGHJKLMNPQRSTUVWXYZ"
	tempPasswordLower  = "abcdefghijkmnopqrstuvwxyz"
	tempPasswordDigits = "23456789"
	tempPasswordLength = 12
)

// GenerateTemporaryPassword returns a random password that always contains an
// uppercase letter, a lowercase letter and a digit, so it passes the frontend
// password rules. Look-alike characters (0/O, 1/l/I) are left out because the
// admin reads the password out to the user.
func GenerateTemporaryPassword() (string, error) {
	all := tempPasswordUpper + tempPasswordLower + tempPasswordDigits
	required := []string{tempPasswordUpper, tempPasswordLower, tempPasswordDigits}

	buf := make([]byte, tempPasswordLength)
	for i := range buf {
		charset := all
		if i < len(required) {
			charset = required[i]
		}
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		buf[i] = charset[n.Int64()]
	}

	// Shuffle so the guaranteed characters are not always at the front
	for i := len(buf) - 1; i > 0; i-- {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		j := n.Int64()
		buf[i], buf[j] = buf[j], buf[i]
	}

	return string(buf), nil
}
