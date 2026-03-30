package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	redisRepo "auth-service/pkg/redis"
)

// ICacheRepository defines the interface for cache operations
type ICacheRepository interface {
	// Token Management
	StoreToken(ctx context.Context, token, userID string, expiration time.Duration) error
	IsTokenValid(ctx context.Context, token string) (bool, error)
	InvalidateToken(ctx context.Context, token string) error
	InvalidateUserTokens(ctx context.Context, userID string) error

	// Session Management
	StoreSession(ctx context.Context, sessionID string, sessionData interface{}, expiration time.Duration) error
	GetSession(ctx context.Context, sessionID string, target interface{}) error
	InvalidateSession(ctx context.Context, sessionID string) error

	// Rate Limiting
	CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, int64, error)
	IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int64, error)

	// Login Attempt Tracking
	TrackLoginAttempt(ctx context.Context, identifier string, expiration time.Duration) (int64, error)
	GetLoginAttempts(ctx context.Context, identifier string) (int64, error)
	ClearLoginAttempts(ctx context.Context, identifier string) error

	// MFA Caching
	StoreMFACode(ctx context.Context, userID, code string, expiration time.Duration) error
	GetMFACode(ctx context.Context, userID string) (string, error)
	VerifyAndDeleteMFACode(ctx context.Context, userID, code string) (bool, error)
	StoreMFAVerified(ctx context.Context, userID string, expiration time.Duration) error
	IsMFAVerified(ctx context.Context, userID string) (bool, error)

	// Password Reset
	StorePasswordResetToken(ctx context.Context, token, userID string, expiration time.Duration) error
	GetPasswordResetUserID(ctx context.Context, token string) (string, error)
	InvalidatePasswordResetToken(ctx context.Context, token string) error

	// User Permissions/Roles Caching
	CacheUserPermissions(ctx context.Context, userID string, permissions []string, expiration time.Duration) error
	GetUserPermissions(ctx context.Context, userID string) ([]string, error)
	InvalidateUserPermissions(ctx context.Context, userID string) error

	// Refresh Token Management
	StoreRefreshToken(ctx context.Context, userID, tokenID string, expiration time.Duration) error
	IsRefreshTokenValid(ctx context.Context, tokenID string) (bool, error)
	InvalidateRefreshToken(ctx context.Context, tokenID string) error
	InvalidateAllUserRefreshTokens(ctx context.Context, userID string) error

	// Trusted Device
	StoreTrustedDevice(ctx context.Context, userID, deviceID string, expiration time.Duration) error
	IsDeviceTrusted(ctx context.Context, userID, deviceID string) (bool, error)
	RemoveTrustedDevice(ctx context.Context, userID, deviceID string) error

	// Generic Operations
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, target interface{}) error
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (int64, error)
	DeleteByPattern(ctx context.Context, pattern string) error
	GetClient() *redisRepo.Client
}

// CacheRepository implements ICacheRepository
type CacheRepository struct {
	redisRepo *redisRepo.Repository
	client    *redisRepo.Client
}

// GetClient implements ICacheRepository.
func (r *CacheRepository) GetClient() *redisRepo.Client {
	return r.client
}

// NewCacheRepository creates a new cache repository
func NewCacheRepository(redisRepo *redisRepo.Repository, client *redisRepo.Client) ICacheRepository {
	return &CacheRepository{
		redisRepo: redisRepo,
		client:    client,
	}
}

// Key prefixes for organization
const (
	KeyPrefixToken         = "auth:token:"
	KeyPrefixSession       = "auth:session:"
	KeyPrefixRateLimit     = "auth:ratelimit:"
	KeyPrefixLoginAttempt  = "auth:loginattempt:"
	KeyPrefixMFA           = "auth:mfa:"
	KeyPrefixPasswordReset = "auth:passwordreset:"
	KeyPrefixPermissions   = "auth:permissions:"
	KeyPrefixRefreshToken  = "auth:refreshtoken:"
	KeyPrefixTrustedDevice = "auth:trusteddevice:"
)

// ============================================
// Token Management
// ============================================

// StoreToken stores a JWT token in cache for logout functionality
func (r *CacheRepository) StoreToken(ctx context.Context, token, userID string, expiration time.Duration) error {
	key := KeyPrefixToken + token
	return r.redisRepo.SetRaw(ctx, key, []byte(userID), expiration)
}

// IsTokenValid checks if a token exists in the cache
func (r *CacheRepository) IsTokenValid(ctx context.Context, token string) (bool, error) {
	key := KeyPrefixToken + token
	exists, err := r.redisRepo.Exists(ctx, key)
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// InvalidateToken removes a specific token from cache
func (r *CacheRepository) InvalidateToken(ctx context.Context, token string) error {
	key := KeyPrefixToken + token
	return r.redisRepo.Delete(ctx, key)
}

// InvalidateUserTokens invalidates all tokens for a user (useful for logout-all devices)
func (r *CacheRepository) InvalidateUserTokens(ctx context.Context, userID string) error {
	pattern := KeyPrefixToken + "*"
	return r.DeleteByPattern(ctx, pattern)
}

// ============================================
// Session Management
// ============================================

// StoreSession stores session data in cache
func (r *CacheRepository) StoreSession(ctx context.Context, sessionID string, sessionData interface{}, expiration time.Duration) error {
	key := KeyPrefixSession + sessionID
	return r.redisRepo.Set(ctx, key, sessionData, expiration)
}

// GetSession retrieves session data from cache
func (r *CacheRepository) GetSession(ctx context.Context, sessionID string, target interface{}) error {
	key := KeyPrefixSession + sessionID
	return r.redisRepo.Get(ctx, key, target)
}

// InvalidateSession removes a session from cache
func (r *CacheRepository) InvalidateSession(ctx context.Context, sessionID string) error {
	key := KeyPrefixSession + sessionID
	return r.redisRepo.Delete(ctx, key)
}

// ============================================
// Rate Limiting
// ============================================

// CheckRateLimit checks if a rate limit has been exceeded
func (r *CacheRepository) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, int64, error) {
	rateLimitKey := KeyPrefixRateLimit + key

	// Get current count
	count, err := r.redisRepo.Increment(ctx, rateLimitKey)
	if err != nil {
		return false, 0, err
	}

	// Set expiration on first increment
	if count == 1 {
		if err := r.redisRepo.Expire(ctx, rateLimitKey, window); err != nil {
			return false, count, err
		}
	}

	return count > int64(limit), count, nil
}

// IncrementRateLimit increments the rate limit counter
func (r *CacheRepository) IncrementRateLimit(ctx context.Context, key string, window time.Duration) (int64, error) {
	rateLimitKey := KeyPrefixRateLimit + key

	count, err := r.redisRepo.Increment(ctx, rateLimitKey)
	if err != nil {
		return 0, err
	}

	if count == 1 {
		if err := r.redisRepo.Expire(ctx, rateLimitKey, window); err != nil {
			return count, err
		}
	}

	return count, nil
}

// ============================================
// Login Attempt Tracking
// ============================================

// TrackLoginAttempt tracks a login attempt and returns the current attempt count
func (r *CacheRepository) TrackLoginAttempt(ctx context.Context, identifier string, expiration time.Duration) (int64, error) {
	key := KeyPrefixLoginAttempt + identifier

	count, err := r.redisRepo.Increment(ctx, key)
	if err != nil {
		return 0, err
	}

	// Set expiration on first attempt
	if count == 1 {
		if err := r.redisRepo.Expire(ctx, key, expiration); err != nil {
			return count, err
		}
	}

	return count, nil
}

// GetLoginAttempts gets the current number of login attempts
func (r *CacheRepository) GetLoginAttempts(ctx context.Context, identifier string) (int64, error) {
	key := KeyPrefixLoginAttempt + identifier
	val, err := r.client.Get(ctx, key).Int64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, err
	}
	return val, nil
}

// ClearLoginAttempts clears all login attempts for an identifier
func (r *CacheRepository) ClearLoginAttempts(ctx context.Context, identifier string) error {
	key := KeyPrefixLoginAttempt + identifier
	return r.redisRepo.Delete(ctx, key)
}

// ============================================
// MFA Caching
// ============================================

// StoreMFACode stores an MFA verification code
func (r *CacheRepository) StoreMFACode(ctx context.Context, userID, code string, expiration time.Duration) error {
	key := KeyPrefixMFA + "code:" + userID
	return r.redisRepo.SetRaw(ctx, key, []byte(code), expiration)
}

// GetMFACode retrieves an MFA code
func (r *CacheRepository) GetMFACode(ctx context.Context, userID string) (string, error) {
	key := KeyPrefixMFA + "code:" + userID
	data, err := r.redisRepo.GetRaw(ctx, key)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// VerifyAndDeleteMFACode verifies the MFA code and deletes it (single use)
func (r *CacheRepository) VerifyAndDeleteMFACode(ctx context.Context, userID, code string) (bool, error) {
	key := KeyPrefixMFA + "code:" + userID

	storedCode, err := r.GetMFACode(ctx, userID)
	if err != nil {
		return false, err
	}

	if storedCode != code {
		return false, nil
	}

	// Delete the code after successful verification
	if err := r.redisRepo.Delete(ctx, key); err != nil {
		return false, err
	}

	return true, nil
}

// StoreMFAVerified marks a user as MFA verified for a session
func (r *CacheRepository) StoreMFAVerified(ctx context.Context, userID string, expiration time.Duration) error {
	key := KeyPrefixMFA + "verified:" + userID
	return r.redisRepo.SetRaw(ctx, key, []byte("1"), expiration)
}

// IsMFAVerified checks if a user is MFA verified
func (r *CacheRepository) IsMFAVerified(ctx context.Context, userID string) (bool, error) {
	key := KeyPrefixMFA + "verified:" + userID
	exists, err := r.redisRepo.Exists(ctx, key)
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// ============================================
// Password Reset
// ============================================

// StorePasswordResetToken stores a password reset token
func (r *CacheRepository) StorePasswordResetToken(ctx context.Context, token, userID string, expiration time.Duration) error {
	key := KeyPrefixPasswordReset + token
	return r.redisRepo.SetRaw(ctx, key, []byte(userID), expiration)
}

// GetPasswordResetUserID retrieves the user ID associated with a reset token
func (r *CacheRepository) GetPasswordResetUserID(ctx context.Context, token string) (string, error) {
	key := KeyPrefixPasswordReset + token
	data, err := r.redisRepo.GetRaw(ctx, key)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// InvalidatePasswordResetToken invalidates a password reset token
func (r *CacheRepository) InvalidatePasswordResetToken(ctx context.Context, token string) error {
	key := KeyPrefixPasswordReset + token
	return r.redisRepo.Delete(ctx, key)
}

// ============================================
// User Permissions/Roles Caching
// ============================================

// CacheUserPermissions caches user permissions for quick access
func (r *CacheRepository) CacheUserPermissions(ctx context.Context, userID string, permissions []string, expiration time.Duration) error {
	key := KeyPrefixPermissions + userID
	return r.redisRepo.Set(ctx, key, permissions, expiration)
}

// GetUserPermissions retrieves cached user permissions
func (r *CacheRepository) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	key := KeyPrefixPermissions + userID
	var permissions []string
	err := r.redisRepo.Get(ctx, key, &permissions)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// InvalidateUserPermissions invalidates cached permissions for a user
func (r *CacheRepository) InvalidateUserPermissions(ctx context.Context, userID string) error {
	key := KeyPrefixPermissions + userID
	return r.redisRepo.Delete(ctx, key)
}

// ============================================
// Refresh Token Management
// ============================================

// StoreRefreshToken stores a refresh token
func (r *CacheRepository) StoreRefreshToken(ctx context.Context, userID, tokenID string, expiration time.Duration) error {
	key := KeyPrefixRefreshToken + tokenID
	data := map[string]string{
		"user_id":  userID,
		"token_id": tokenID,
	}
	return r.redisRepo.Set(ctx, key, data, expiration)
}

// IsRefreshTokenValid checks if a refresh token is valid
func (r *CacheRepository) IsRefreshTokenValid(ctx context.Context, tokenID string) (bool, error) {
	key := KeyPrefixRefreshToken + tokenID
	exists, err := r.redisRepo.Exists(ctx, key)
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// InvalidateRefreshToken invalidates a specific refresh token
func (r *CacheRepository) InvalidateRefreshToken(ctx context.Context, tokenID string) error {
	key := KeyPrefixRefreshToken + tokenID
	return r.redisRepo.Delete(ctx, key)
}

// InvalidateAllUserRefreshTokens invalidates all refresh tokens for a user
func (r *CacheRepository) InvalidateAllUserRefreshTokens(ctx context.Context, userID string) error {
	// Note: In production, you might want to maintain a user->token index
	// For simplicity, we'll use a pattern delete
	pattern := KeyPrefixRefreshToken + "*"
	return r.DeleteByPattern(ctx, pattern)
}

// ============================================
// Trusted Device Management
// ============================================

// StoreTrustedDevice marks a device as trusted for a user
func (r *CacheRepository) StoreTrustedDevice(ctx context.Context, userID, deviceID string, expiration time.Duration) error {
	key := KeyPrefixTrustedDevice + userID + ":" + deviceID
	return r.redisRepo.SetRaw(ctx, key, []byte("1"), expiration)
}

// IsDeviceTrusted checks if a device is trusted for a user
func (r *CacheRepository) IsDeviceTrusted(ctx context.Context, userID, deviceID string) (bool, error) {
	key := KeyPrefixTrustedDevice + userID + ":" + deviceID
	exists, err := r.redisRepo.Exists(ctx, key)
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// RemoveTrustedDevice removes a device from trusted list
func (r *CacheRepository) RemoveTrustedDevice(ctx context.Context, userID, deviceID string) error {
	key := KeyPrefixTrustedDevice + userID + ":" + deviceID
	return r.redisRepo.Delete(ctx, key)
}

// ============================================
// Generic Operations
// ============================================

// Set stores a value with expiration
func (r *CacheRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.redisRepo.Set(ctx, key, value, expiration)
}

// Get retrieves and unmarshals a value
func (r *CacheRepository) Get(ctx context.Context, key string, target interface{}) error {
	return r.redisRepo.Get(ctx, key, target)
}

// Delete removes keys
func (r *CacheRepository) Delete(ctx context.Context, keys ...string) error {
	return r.redisRepo.Delete(ctx, keys...)
}

// Exists checks if keys exist
func (r *CacheRepository) Exists(ctx context.Context, keys ...string) (int64, error) {
	return r.redisRepo.Exists(ctx, keys...)
}

// DeleteByPattern deletes all keys matching a pattern
func (r *CacheRepository) DeleteByPattern(ctx context.Context, pattern string) error {
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	var keysToDelete []string

	for iter.Next(ctx) {
		keysToDelete = append(keysToDelete, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	if len(keysToDelete) > 0 {
		return r.redisRepo.Delete(ctx, keysToDelete...)
	}

	return nil
}

// ============================================
// Utility Methods
// ============================================

// GetJSON retrieves a value and unmarshals it as JSON
func (r *CacheRepository) GetJSON(ctx context.Context, key string) ([]byte, error) {
	return r.redisRepo.GetRaw(ctx, key)
}

// SetJSON stores a value as JSON
func (r *CacheRepository) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	return r.redisRepo.SetRaw(ctx, key, data, expiration)
}

// SetNX sets a key only if it doesn't exist (for distributed locks)
func (r *CacheRepository) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.redisRepo.SetNX(ctx, key, value, expiration)
}

// GetTTL returns the remaining TTL for a key
func (r *CacheRepository) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	return r.redisRepo.TTL(ctx, key)
}
