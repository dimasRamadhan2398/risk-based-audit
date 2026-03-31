package repositories

import (
	"time"

	"auth-service/models"
	apperrors "auth-service/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IRefreshTokenRepository is an alias for the refresh token repository interface
type IRefreshTokenRepository = RefreshTokenRepositoryInterface

// RefreshTokenRepositoryInterface defines the refresh token repository interface
type RefreshTokenRepositoryInterface interface {
	Create(token *models.RefreshToken) error
	FindByToken(token string) (*models.RefreshToken, error)
	FindByUserID(userID uuid.UUID) ([]*models.RefreshToken, error)
	Revoke(token string) error
	RevokeAllByUserID(userID uuid.UUID) error
	DeleteExpired() error
}

// RefreshTokenRepository handles refresh token data operations
type RefreshTokenRepository struct {
	*BaseRepository
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *gorm.DB) IRefreshTokenRepository {
	return &RefreshTokenRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new refresh token
func (r *RefreshTokenRepository) Create(token *models.RefreshToken) error {
	return r.BaseRepository.Create(token)
}

// FindByToken finds a refresh token by token string
func (r *RefreshTokenRepository) FindByToken(token string) (*models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	if err := r.GetDB().Where("token = ?", token).First(&refreshToken).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &refreshToken, nil
}

// FindByUserID finds all refresh tokens for a user
func (r *RefreshTokenRepository) FindByUserID(userID uuid.UUID) ([]*models.RefreshToken, error) {
	var tokens []*models.RefreshToken
	if err := r.GetDB().Where("user_id = ?", userID).Order("created_at DESC").Find(&tokens).Error; err != nil {
		return nil, err
	}
	return tokens, nil
}

// Revoke revokes a refresh token
func (r *RefreshTokenRepository) Revoke(token string) error {
	return r.GetDB().Where("token = ?", token).Update("revoked", true).Error
}

// RevokeAllByUserID revokes all refresh tokens for a user
func (r *RefreshTokenRepository) RevokeAllByUserID(userID uuid.UUID) error {
	return r.GetDB().Where("user_id = ?", userID).Update("revoked", true).Error
}

// DeleteExpired deletes all expired refresh tokens
func (r *RefreshTokenRepository) DeleteExpired() error {
	result := r.GetDB().Where("expires_at < ?", time.Now()).Delete(&models.RefreshToken{})
	return result.Error
}
