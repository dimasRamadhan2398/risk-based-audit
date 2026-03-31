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
	"time"

	"github.com/google/uuid"
)

// AuthServiceInterface defines the auth service interface
type AuthServiceInterface interface {
	Login(ctx context.Context, req *models.LoginRequest, ipAddress, userAgent string) (*models.LoginResponse, error)
	Register(ctx context.Context, req *models.RegisterRequest, ipAddress string) error
	Logout(ctx context.Context, token, userID, username, ipAddress, userAgent string) error
	ValidateToken(token string) (*utils.Claims, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, req *models.ChangePasswordRequest, ipAddress string) error
}

// AuthService handles authentication logic
type AuthService struct {
	*baseService.BaseService
	userRepo     repositories.UserRepositoryInterface
	redis        *redis.Client
	config       *config.Config
	kafkaProducer *kafka.Producer
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo repositories.UserRepositoryInterface,
	rdb *redis.Client,
	cfg *config.Config,
	kafkaProducer *kafka.Producer,
) AuthServiceInterface {
	return &AuthService{
		BaseService:   baseService.NewBaseService(),
		userRepo:      userRepo,
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
			"user_id":   user.ID.String(),
			"username":  user.Username,
			"ip_address": ipAddress,
			"user_agent": userAgent,
		})
	}

	return &models.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt.Unix(),
		User: models.UserInfo{
			ID:        user.ID.String(),
			Username:  user.Username,
			Email:     user.Email,
			FullName:  user.FullName,
			Phone:     user.Phone,
			Department: user.Department,
			Roles:     roles,
		},
	}, nil
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
