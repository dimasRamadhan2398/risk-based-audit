package models

import (
	"time"

	"github.com/google/uuid"
)

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required" validate:"required,min=3,max=50"`
	Password string `json:"password" binding:"required" validate:"required,min=6"`
}

// RegisterRequest represents a register request
type RegisterRequest struct {
	Username  string `json:"username" binding:"required" validate:"required,min=3,max=50"`
	Email     string `json:"email" binding:"required" validate:"required,email"`
	Password  string `json:"password" binding:"required" validate:"required,min=6"`
	FullName  string `json:"full_name" validate:"max=100"`
	Phone     string `json:"phone" validate:"max=20"`
	Department string `json:"department" validate:"max=100"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string   `json:"token"`
	ExpiresAt int64    `json:"expires_at"`
	User      UserInfo `json:"user"`
}

// UserInfo represents user information
type UserInfo struct {
	ID        string   `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	FullName  string   `json:"full_name"`
	Phone     string   `json:"phone"`
	Department string  `json:"department"`
	Roles     []string `json:"roles"`
}

// ChangePasswordRequest represents a change password request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required" validate:"required"`
	NewPassword string `json:"new_password" binding:"required" validate:"required,min=6"`
}

// CreateUserRequest represents a create user request
type CreateUserRequest struct {
	Username   string   `json:"username" binding:"required" validate:"required,min=3,max=50"`
	Email      string   `json:"email" binding:"required" validate:"required,email"`
	Password   string   `json:"password" binding:"required" validate:"required,min=6"`
	FullName   string   `json:"full_name" validate:"max=100"`
	Phone      string   `json:"phone" validate:"max=20"`
	Department string   `json:"department" validate:"max=100"`
	Roles      []string `json:"roles"`
}

// UpdateUserRequest represents an update user request
type UpdateUserRequest struct {
	FullName   *string  `json:"full_name" validate:"omitempty,max=100"`
	Phone      *string  `json:"phone" validate:"omitempty,max=20"`
	Department *string  `json:"department" validate:"omitempty,max=100"`
	IsActive   *bool    `json:"is_active"`
}

// ListUsersRequest represents a list users request
type ListUsersRequest struct {
	Page      int    `form:"page" validate:"min=1"`
	PageSize  int    `form:"page_size" validate:"min=1,max=100"`
	Search    string `form:"search"`
	Department string `form:"department"`
	IsActive  *bool  `form:"is_active"`
}

// UserResponse represents a user response
type UserResponse struct {
	ID         string  `json:"id"`
	Username   string  `json:"username"`
	Email      string  `json:"email"`
	FullName   string  `json:"full_name"`
	Phone      string  `json:"phone"`
	Department string  `json:"department"`
	IsActive   bool    `json:"is_active"`
	Roles      []string `json:"roles"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

// MFA DTOs

// SetupMfaRequest represents a request to setup MFA
type SetupMfaRequest struct {
	MFAType MFAType `json:"mfa_type" binding:"required" validate:"required"`
}

// VerifyMfaRequest represents a request to verify MFA code
type VerifyMfaRequest struct {
	Code string `json:"code" binding:"required" validate:"required,len=6"`
}

type MfaSetupRequest struct {
    MFAType MFAType `json:"mfa_type" binding:"required"`
}

// MfaSetupResponse represents the MFA setup response
type MfaSetupResponse struct {
	Secret         string `json:"secret,omitempty"`
	QRCodeURL      string `json:"qr_code_url,omitempty"`
	MFAType        MFAType `json:"mfa_type"`
	IsEnabled      bool   `json:"is_enabled"`
	BackupCodes    []string `json:"backup_codes,omitempty"`
}

// MfaVerificationResponse represents MFA verification response
type MfaVerificationResponse struct {
	Verified bool `json:"verified"`
}

// EnableMfaRequest represents a request to enable MFA
type EnableMfaRequest struct {
	Code string `json:"code" binding:"required" validate:"required,len=6"`
}

// DisableMfaRequest represents a request to disable MFA
type DisableMfaRequest struct {
	Password string `json:"password" binding:"required" validate:"required"`
}

type EmailCodeRequest struct {
	Email string `json:"email" binding:"required" validate:"required,email"`
}

type TrustedDeviceRequest struct {
	DeviceFingerprint string `json:"device_fingerprint" binding:"required" validate:"required"`
	DeviceName        string `json:"device_name" binding:"required" validate:"required"`
	DeviceType        string `json:"device_type" binding:"required" validate:"required"`
}

type TrustedDeviceResponse struct {
	ID                uuid.UUID  `json:"id"`
	DeviceFingerprint string     `json:"device_fingerprint"`
	DeviceName        string     `json:"device_name"`
	DeviceType        string     `json:"device_type"`
	UserAgent 		  string     `json:"user_agent"`
	IPAddress         string     `json:"ip_address"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

type QREnrollmentResponse struct {
	Token     string `json:"token"`
	QRCodeURL string `json:"qr_code_url"`
	ExpiresAt time.Time  `json:"expires_at"`
}

type CreateRoleRequest struct {
    Name        string `json:"name" binding:"required"`
    Description string `json:"description"`
}

type CreateRoleResponse struct{
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
type UpdateRoleRequest struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}
type UpdateRoleResponse struct{
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
type PaginatedResponse struct {
    Data   interface{} `json:"data"`
    Total  int64       `json:"total"`
    Offset int         `json:"offset"`
    Limit  int         `json:"limit"`
}