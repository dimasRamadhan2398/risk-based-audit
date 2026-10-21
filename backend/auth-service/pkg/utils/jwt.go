package utils

import (
	"time"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	// EmployeeID is the master-data employees.id (UUID) linked to the user.
	// Empty when the account is not linked or master-service was unreachable.
	EmployeeID string `json:"employee_id,omitempty"`
	// EmployeeCode is users.employee_id (employees.employee_code).
	// Empty when the account is not linked to an employee.
	EmployeeCode string `json:"employee_code,omitempty"`
	jwt.RegisteredClaims
}

// GenerateToken generates a JWT token without employee claims
func GenerateToken(userID, username string, roles []string, secret string, expiryHour int) (string, error) {
	return GenerateTokenWithEmployee(userID, username, roles, "", "", secret, expiryHour)
}

// GenerateTokenWithEmployee generates a JWT token carrying the employee link.
// Empty employeeID or employeeCode leaves the matching claim out of the token.
func GenerateTokenWithEmployee(userID, username string, roles []string, employeeID, employeeCode, secret string, expiryHour int) (string, error) {
	claims := Claims{
		UserID:       userID,
		Username:     username,
		Roles:        roles,
		EmployeeID:   employeeID,
		EmployeeCode: employeeCode,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(expiryHour))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken parses and validates a JWT token
func ParseToken(tokenString, secret string) (*Claims, error) {
	// GenerateToken above always signs HS256 and always sets exp; pin both so a
	// token claiming another algorithm, or none at all, is rejected outright.
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
	)

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
