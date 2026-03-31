package repositories

import (
	"auth-service/models"

	apperrors "auth-service/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IMFASetupRepository is an alias for the MFA setup repository interface
type IMFASetupRepository = MFASetupRepositoryInterface

// MFASetupRepositoryInterface defines the MFA setup repository interface
type MFASetupRepositoryInterface interface {
	Create(mfa *models.MFASetup) error
	Update(mfa *models.MFASetup) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.MFASetup, error)
	FindByUserID(userID uuid.UUID) (*models.MFASetup, error)
	FindByUserIDAndType(userID uuid.UUID, mfaType models.MFAType) (*models.MFASetup, error)
	Enable(userID uuid.UUID) error
	Disable(userID uuid.UUID) error
}

// MFASetupRepository handles MFA setup data operations
type MFASetupRepository struct {
	*BaseRepository
}

// NewMFASetupRepository creates a new MFA setup repository
func NewMFASetupRepository(db *gorm.DB) IMFASetupRepository {
	return &MFASetupRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new MFA setup
func (r *MFASetupRepository) Create(mfa *models.MFASetup) error {
	return r.BaseRepository.Create(mfa)
}

// Update updates an MFA setup
func (r *MFASetupRepository) Update(mfa *models.MFASetup) error {
	return r.BaseRepository.Update(mfa)
}

// Delete deletes an MFA setup
func (r *MFASetupRepository) Delete(id uuid.UUID) error {
	return r.BaseRepository.Delete(&models.MFASetup{ID: id})
}

// FindByID finds an MFA setup by ID
func (r *MFASetupRepository) FindByID(id uuid.UUID) (*models.MFASetup, error) {
	var mfa models.MFASetup
	if err := r.GetDB().First(&mfa, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &mfa, nil
}

// FindByUserID finds an MFA setup by user ID
func (r *MFASetupRepository) FindByUserID(userID uuid.UUID) (*models.MFASetup, error) {
	var mfa models.MFASetup
	if err := r.GetDB().Where("user_id = ?", userID).First(&mfa).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &mfa, nil
}

// FindByUserIDAndType finds an MFA setup by user ID and MFA type
func (r *MFASetupRepository) FindByUserIDAndType(userID uuid.UUID, mfaType models.MFAType) (*models.MFASetup, error) {
	var mfa models.MFASetup
	if err := r.GetDB().Where("user_id = ? AND mfa_type = ?", userID, mfaType).First(&mfa).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &mfa, nil
}

// Enable enables MFA for a user
func (r *MFASetupRepository) Enable(userID uuid.UUID) error {
	return r.GetDB().Model(&models.MFASetup{}).Where("user_id = ?", userID).Update("is_enabled", true).Error
}

// Disable disables MFA for a user
func (r *MFASetupRepository) Disable(userID uuid.UUID) error {
	return r.GetDB().Model(&models.MFASetup{}).Where("user_id = ?", userID).Update("is_enabled", false).Error
}
