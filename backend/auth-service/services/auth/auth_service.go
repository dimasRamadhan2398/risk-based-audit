package services

import (
	"auth-service/models"
	baseService "auth-service/pkg/base"
	"auth-service/pkg/config"
	"auth-service/pkg/errors"
	"auth-service/pkg/kafka"
	"auth-service/pkg/redis"
	"auth-service/pkg/utils"
	"auth-service/repositories"
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AuthServiceInterface defines the auth service interface
type AuthServiceInterface interface {
	Login(ctx context.Context, req *models.LoginRequest, ipAddress, userAgent string) (*models.LoginResponse, error)
	VerifyMFALogin(ctx context.Context, req *models.VerifyMFALoginRequest, ipAddress, userAgent string) (*models.LoginResponse, error)
	Register(ctx context.Context, req *models.RegisterRequest, ipAddress string) error
	Logout(ctx context.Context, token, userID, username, ipAddress, userAgent string) error
	ValidateToken(token string) (*utils.Claims, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, req *models.ChangePasswordRequest, ipAddress string) error
}

// AuthService handles authentication logic
type AuthService struct {
	*baseService.BaseService
	userRepo      repositories.UserRepositoryInterface
	mfaRepo       repositories.MFASetupRepositoryInterface
	trustedRepo   repositories.ITrustedDeviceRepository
	redis         *redis.Client
	config        *config.Config
	kafkaProducer *kafka.Producer
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo repositories.UserRepositoryInterface,
	mfaRepo repositories.MFASetupRepositoryInterface,
	trustedRepo repositories.ITrustedDeviceRepository,
	rdb *redis.Client,
	cfg *config.Config,
	kafkaProducer *kafka.Producer,
) AuthServiceInterface {
	return &AuthService{
		BaseService:   baseService.NewBaseService(),
		userRepo:      userRepo,
		mfaRepo:       mfaRepo,
		trustedRepo:   trustedRepo,
		redis:         rdb,
		config:        cfg,
		kafkaProducer: kafkaProducer,
	}
}

// Login authenticates a user and returns a token
func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest, ipAddress, userAgent string) (*models.LoginResponse, error) {
	// Find user by username
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			s.LogInfo("Login attempt with non-existent username", utils.LogField("username", req.Username))
			return nil, errors.ErrInvalidCredentials
		}
		s.LogError("Failed to find user during login", utils.LogField("error", err))
		return nil, errors.ErrInternalServer
	}

	// Check if user is active
	if !user.IsActive {
		s.LogInfo("Login attempt for inactive user", utils.LogField("user_id", user.ID))
		return nil, errors.ErrUnauthorized
	}

	// Verify password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		s.LogInfo("Login attempt with invalid password", utils.LogField("user_id", user.ID))
		return nil, errors.ErrInvalidCredentials
	}

	isNewDevice := false
	if req.DeviceFingerprint != "" && user.LastLoginFingerprint != "" && user.LastLoginFingerprint != req.DeviceFingerprint {
		isNewDevice = true
	}

	// Check for MFA
	mfa, err := s.mfaRepo.FindByUserID(user.ID)
	if err == nil && mfa != nil && mfa.IsEnabled {
		// Check if device is trusted
		isTrusted := false
		if req.DeviceFingerprint != "" {
			device, _ := s.trustedRepo.FindByFingerprint(user.ID, req.DeviceFingerprint)
			if device != nil && !s.trustedRepo.IsExpired(device) {
				isTrusted = true
			}
		}

		if !isTrusted {
			// Generate MFA token
			mfaToken := uuid.New().String()
			key := "mfa_login:" + mfaToken
			if err := s.redis.Set(ctx, key, user.ID.String(), 10*time.Minute).Err(); err != nil {
				return nil, errors.ErrInternalServer
			}

			return &models.LoginResponse{
				MFARequired: true,
				MFAToken:    mfaToken,
				IsNewDevice: isNewDevice,
			}, nil
		}
	}

	return s.completeLogin(ctx, user, req.DeviceFingerprint, ipAddress, userAgent, isNewDevice)
}

func (s *AuthService) completeLogin(ctx context.Context, user *models.User, fingerprint, ipAddress, userAgent string, isNewDevice bool) (*models.LoginResponse, error) {
	// Update user's last login info
	now := time.Now()
	user.LastLoginAt = &now
	if fingerprint != "" {
		user.LastLoginFingerprint = fingerprint
	}
	s.userRepo.Update(user)

	// Generate roles list
	roles := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = role.Name
	}

	// Generate JWT token
	token, err := utils.GenerateToken(
		user.ID.String(),
		user.Username,
		roles,
		s.config.JWT.Secret,
		s.config.JWT.ExpiryHour,
	)
	if err != nil {
		s.LogError("Failed to generate token", utils.LogField("error", err))
		return nil, errors.ErrInternalServer
	}

	// Store token in Redis for logout functionality
	expiresAt := time.Now().Add(time.Hour * time.Duration(s.config.JWT.ExpiryHour))
	key := "token:" + token
	if err := s.redis.Set(ctx, key, user.ID.String(), time.Until(expiresAt)).Err(); err != nil {
		s.LogWarning("Failed to store token in redis", utils.LogField("error", err))
	}

	s.LogInfo("User logged in successfully", utils.LogField("user_id", user.ID))

	// Send to Kafka for centralized logging
	if s.kafkaProducer != nil && s.kafkaProducer.IsEnabled() {
		s.kafkaProducer.Info(ctx, "User logged in successfully", map[string]interface{}{
			"user_id":    user.ID.String(),
			"username":   user.Username,
			"ip_address": ipAddress,
			"user_agent": userAgent,
		})
	}

	return &models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt.Unix(),
		User: models.UserInfo{
			ID:         user.ID.String(),
			Username:   user.Username,
			Email:      user.Email,
			FullName:   user.FullName,
			Phone:      user.Phone,
			Department: user.Department,
			Roles:      roles,
		},
		IsNewDevice: isNewDevice,
	}, nil
}

// VerifyMFALogin verifies MFA code during login
func (s *AuthService) VerifyMFALogin(ctx context.Context, req *models.VerifyMFALoginRequest, ipAddress, userAgent string) (*models.LoginResponse, error) {
	key := "mfa_login:" + req.MFAToken
	userIDStr, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, errors.Wrap("INVALID_MFA_TOKEN", "Invalid or expired MFA login token", 400, nil)
	}

	userID, _ := uuid.Parse(userIDStr)
	mfa, err := s.mfaRepo.FindByUserID(userID)
	if err != nil || !mfa.IsEnabled {
		return nil, errors.Wrap("MFA_NOT_ENABLED", "MFA is not enabled for this user", 400, nil)
	}

	// Verify code (simulated logic for now, using a helper if available)
	// In a real implementation, we should call MfaService.VerifyMFA but we need to avoid circular dependency
	// For this task, we will assume a simple TOTP verification helper exists or implement it here
	// Since MfaService already has VerifyMFA, we might need to restructure or use a shared package for TOTP
	// For simplicity, let's assume we have access to a verify function.

	// As per MfaService, it uses s.verifyTOTP(mfa.SecretKey, code)
	// We'll duplicate the simple TOTP verification logic or call MfaService if possible.
	// To avoid circular dependency, we'll implement a basic check or assume the user wants us to use the existing MfaService.
	// But AuthService doesn't have MfaService.

	// Let's implement verifyTOTP logic here as well for now or use utils if available.
	if !s.verifyTOTP(mfa.SecretKey, req.Code) {
		return nil, errors.Wrap("INVALID_CODE", "Invalid MFA code", 400, nil)
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.ErrNotFound
	}

	if req.TrustDevice && req.DeviceFingerprint != "" {
		expiresAt := time.Now().Add(90 * 24 * time.Hour)
		device := &models.TrustedDevice{
			ID:                uuid.New(),
			UserID:            userID,
			DeviceFingerprint: req.DeviceFingerprint,
			DeviceName:        req.DeviceName,
			DeviceType:        req.DeviceType,
			UserAgent:         userAgent,
			IPAddress:         ipAddress,
			ExpiresAt:         &expiresAt,
			CreatedAt:         time.Now(),
		}
		s.trustedRepo.Create(device)
	}

	s.redis.Del(ctx, key)

	isNewDevice := false
	if req.DeviceFingerprint != "" && user.LastLoginFingerprint != "" && user.LastLoginFingerprint != req.DeviceFingerprint {
		isNewDevice = true
	}

	return s.completeLogin(ctx, user, req.DeviceFingerprint, ipAddress, userAgent, isNewDevice)
}

// verifyTOTP is a helper to verify TOTP code
func (s *AuthService) verifyTOTP(secret, code string) bool {
	// Remove spaces from secret and convert to uppercase
	secret = strings.ReplaceAll(secret, " ", "")
	secret = strings.ToUpper(secret)

	// Decode base32 secret
	secretBytes, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return false
	}

	// Get current time step
	currentStep := time.Now().Unix() / 30

	// Check current step and adjacent steps for tolerance
	for _, step := range []int64{currentStep, currentStep - 1, currentStep + 1} {
		expectedCode := s.generateHOTP(secretBytes, step)
		if subtle.ConstantTimeCompare([]byte(expectedCode), []byte(code)) == 1 {
			return true
		}
	}

	return false
}

// generateHOTP generates an HOTP code (Internal use)
func (s *AuthService) generateHOTP(secret []byte, counter int64) string {
	// Convert counter to bytes
	buf := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		buf[i] = byte(counter & 0xff)
		counter >>= 8
	}

	// Calculate HMAC-SHA1
	h := hmac.New(sha1.New, secret)
	h.Write(buf)
	hash := h.Sum(nil)

	// Dynamic truncation
	offset := hash[len(hash)-1] & 0x0f
	truncated := (int(hash[offset])&0x7f)<<24 |
		(int(hash[offset+1])&0xff)<<16 |
		(int(hash[offset+2])&0xff)<<8 |
		(int(hash[offset+3]) & 0xff)

	code := truncated % 1000000
	return fmt.Sprintf("%06d", code)
}

// Register creates a new user
func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest, ipAddress string) error {
	// Check if username exists
	if _, err := s.userRepo.FindByUsername(req.Username); err == nil {
		return errors.Wrap(errors.ErrDuplicateEntry.Code, "Username already exists", errors.ErrDuplicateEntry.StatusCode, nil)
	}

	// Check if email exists
	if _, err := s.userRepo.FindByEmail(req.Email); err == nil {
		return errors.Wrap(errors.ErrDuplicateEntry.Code, "Email already exists", errors.ErrDuplicateEntry.StatusCode, nil)
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		s.LogError("Failed to hash password", utils.LogField("error", err))
		return errors.ErrInternalServer
	}

	// Create user
	user := &models.User{
		Username:   req.Username,
		Email:      req.Email,
		PasswordHash:   hashedPassword,
		FullName:   req.FullName,
		Phone:      req.Phone,
		Department: req.Department,
		IsActive:   true,
	}

	if err := s.userRepo.Create(user); err != nil {
		s.LogError("Failed to create user", utils.LogField("error", err))
		return errors.ErrInternalServer
	}

	s.LogInfo("User registered successfully", utils.LogField("user_id", user.ID))

	// Send to Kafka for centralized logging
	if s.kafkaProducer != nil && s.kafkaProducer.IsEnabled() {
		s.kafkaProducer.Info(ctx, "User registered", map[string]interface{}{
			"user_id":    user.ID.String(),
			"username":  user.Username,
			"email":     user.Email,
			"ip_address": ipAddress,
		})
	}

	return nil
}

// Logout invalidates a token
func (s *AuthService) Logout(ctx context.Context, token, userID, username, ipAddress, userAgent string) error {
	key := "token:" + token
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		s.LogError("Failed to delete token from redis", utils.LogField("error", err))
		return errors.ErrInternalServer
	}
	s.LogInfo("User logged out successfully")

	// Send to Kafka for centralized logging
	if s.kafkaProducer != nil && s.kafkaProducer.IsEnabled() {
		s.kafkaProducer.Info(ctx, "User logged out", map[string]interface{}{
			"user_id":    userID,
			"username":   username,
			"ip_address": ipAddress,
			"user_agent": userAgent,
		})
	}

	return nil
}

// ValidateToken validates a JWT token
func (s *AuthService) ValidateToken(token string) (*utils.Claims, error) {
	// Check if token exists in Redis
	ctx := context.Background()
	key := "token:" + token
	exists, err := s.redis.Exists(ctx, key).Result()
	if err != nil || exists == 0 {
		return nil, errors.ErrInvalidToken
	}

	// Parse and validate token
	claims, err := utils.ParseToken(token, s.config.JWT.Secret)
	if err != nil {
		return nil, errors.ErrInvalidToken
	}

	return claims, nil
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, req *models.ChangePasswordRequest, ipAddress string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	// Verify old password
	if !utils.CheckPassword(req.OldPassword, user.PasswordHash) {
		return errors.ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		s.LogError("Failed to hash password", utils.LogField("error", err))
		return errors.ErrInternalServer
	}

	user.PasswordHash = hashedPassword
	if err := s.userRepo.Update(user); err != nil {
		s.LogError("Failed to update password", utils.LogField("error", err))
		return errors.ErrInternalServer
	}

	s.LogInfo("Password changed successfully", utils.LogField("user_id", user.ID))

	// Send to Kafka for centralized logging
	if s.kafkaProducer != nil && s.kafkaProducer.IsEnabled() {
		s.kafkaProducer.Info(ctx, "Password changed successfully", map[string]interface{}{
			"user_id":    user.ID.String(),
			"username":  user.Username,
			"ip_address": ipAddress,
		})
	}

	return nil
}
