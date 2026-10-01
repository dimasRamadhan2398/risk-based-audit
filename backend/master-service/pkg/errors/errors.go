package errors

import (
	stderrors "errors"
	"fmt"
	"net/http"
)

// AppError represents an application error.
//
// Code is a stable, machine-readable UPPER_SNAKE identifier the frontend uses
// as its i18n key. Message is a human-readable English default and must never
// contain SQL, driver output or Go type names. Err holds the underlying
// technical cause; it is logged server-side and never sent to clients.
// Fields optionally maps request field names (JSON names) to field-level codes
// such as REQUIRED or INVALID_FORMAT.
type AppError struct {
	Code       string
	Message    string
	StatusCode int
	Err        error
	Fields     map[string]string
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is checks if an error is of type AppError
func Is(err error, target *AppError) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == target.Code
	}
	return false
}

// As returns the first *AppError in err's chain.
func As(err error) (*AppError, bool) {
	var appErr *AppError
	if stderrors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// New creates a new AppError
func New(code string, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// Wrap wraps an existing error
func Wrap(code string, message string, statusCode int, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}

// WithFields returns a copy of e carrying the given field-level codes.
// The receiver is not modified, so it is safe to call on shared sentinels.
func (e *AppError) WithFields(fields map[string]string) *AppError {
	cp := *e
	cp.Fields = fields
	return &cp
}

// WithErr returns a copy of e wrapping err (for server-side logging).
func (e *AppError) WithErr(err error) *AppError {
	cp := *e
	cp.Err = err
	return &cp
}

// Common errors
var (
	ErrBadRequest         = New("BAD_REQUEST", "Bad request", http.StatusBadRequest)
	ErrUnauthorized       = New(CodeUnauthorized, "Unauthorized", http.StatusUnauthorized)
	ErrForbidden          = New(CodeForbidden, "Forbidden", http.StatusForbidden)
	ErrNotFound           = New(CodeNotFound, MsgNotFound, http.StatusNotFound)
	ErrConflict           = New("CONFLICT", "Resource conflict", http.StatusConflict)
	ErrInternalServer     = New(CodeInternal, "Internal server error", http.StatusInternalServerError)
	ErrDatabase           = New("DATABASE_ERROR", "Database error", http.StatusInternalServerError)
	ErrValidation         = New("VALIDATION_ERROR", "Validation failed", http.StatusBadRequest)
	ErrInvalidCredentials = New("INVALID_CREDENTIALS", "Invalid credentials", http.StatusUnauthorized)
	ErrInvalidToken       = New("INVALID_TOKEN", "Invalid token", http.StatusUnauthorized)
	ErrTokenExpired       = New("TOKEN_EXPIRED", "Token expired", http.StatusUnauthorized)
	ErrUnknownDevice      = New("UNKNOWN_DEVICE", "Unknown device", http.StatusUnauthorized)
	ErrDuplicateEntry     = New(CodeDuplicateEntry, "Duplicate entry", http.StatusConflict)
)
