package repositories

import (
	"time"

	"auth-service/models"
	apperrors "auth-service/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ITrustedDeviceRepository is an alias for the trusted device repository interface
type ITrustedDeviceRepository = TrustedDeviceRepositoryInterface

// TrustedDeviceRepositoryInterface defines the trusted device repository interface
type TrustedDeviceRepositoryInterface interface {
	Create(device *models.TrustedDevice) error
	Delete(id uuid.UUID) error
	DeleteByFingerprint(userID uuid.UUID, fingerprint string) error
	FindByID(id uuid.UUID) (*models.TrustedDevice, error)
	FindByUserID(userID uuid.UUID) ([]*models.TrustedDevice, error)
	FindByFingerprint(userID uuid.UUID, fingerprint string) (*models.TrustedDevice, error)
	IsExpired(device *models.TrustedDevice) bool
	DeleteExpiredDevices(userID uuid.UUID) error
}

// TrustedDeviceRepository handles trusted device data operations
type TrustedDeviceRepository struct {
	*BaseRepository
}

// NewTrustedDeviceRepository creates a new trusted device repository
func NewTrustedDeviceRepository(db *gorm.DB) ITrustedDeviceRepository {
	return &TrustedDeviceRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new trusted device
func (r *TrustedDeviceRepository) Create(device *models.TrustedDevice) error {
	return r.BaseRepository.Create(device)
}

// Delete deletes a trusted device
func (r *TrustedDeviceRepository) Delete(id uuid.UUID) error {
	return r.BaseRepository.Delete(&models.TrustedDevice{ID: id})
}

// DeleteByFingerprint deletes a trusted device by fingerprint
func (r *TrustedDeviceRepository) DeleteByFingerprint(userID uuid.UUID, fingerprint string) error {
	return r.GetDB().Where("user_id = ? AND device_fingerprint = ?", userID, fingerprint).Delete(&models.TrustedDevice{}).Error
}

// FindByID finds a trusted device by ID
func (r *TrustedDeviceRepository) FindByID(id uuid.UUID) (*models.TrustedDevice, error) {
	var device models.TrustedDevice
	if err := r.GetDB().First(&device, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &device, nil
}

// FindByUserID finds all trusted devices for a user
func (r *TrustedDeviceRepository) FindByUserID(userID uuid.UUID) ([]*models.TrustedDevice, error) {
	var devices []*models.TrustedDevice
	if err := r.GetDB().Where("user_id = ?", userID).Order("created_at DESC").Find(&devices).Error; err != nil {
		return nil, err
	}
	return devices, nil
}

// FindByFingerprint finds a trusted device by fingerprint
func (r *TrustedDeviceRepository) FindByFingerprint(userID uuid.UUID, fingerprint string) (*models.TrustedDevice, error) {
	var device models.TrustedDevice
	if err := r.GetDB().Where("user_id = ? AND device_fingerprint = ?", userID, fingerprint).First(&device).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &device, nil
}

// IsExpired checks if a trusted device is expired
func (r *TrustedDeviceRepository) IsExpired(device *models.TrustedDevice) bool {
	return device.ExpiresAt != nil && device.ExpiresAt.Before(time.Now())
}

// DeleteExpiredDevices deletes all expired trusted devices for a user
func (r *TrustedDeviceRepository) DeleteExpiredDevices(userID uuid.UUID) error {
	return r.GetDB().Where("user_id = ? AND expires_at < ?", userID, time.Now()).Delete(&models.TrustedDevice{}).Error
}
