package repositories

import (
	"crypto/subtle"
	"fmt"

	"auth-service/models"

	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/secretcrypt"

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
	// EnablePending turns MFA on only if it is still pending with secretKey,
	// so a setup call that replaced the secret in between is not enabled with a
	// secret the user never scanned. It reports whether a row was switched on.
	EnablePending(userID uuid.UUID, secretKey string) (bool, error)
	Disable(userID uuid.UUID) error
}

// MFASetupRepository handles MFA setup data operations.
//
// SecretKey is encrypted at rest: callers always pass and receive plaintext,
// the repository encrypts on write and decrypts on read (see pkg/secretcrypt).
// Rows without the "enc:" prefix are legacy plaintext and are read as is.
type MFASetupRepository struct {
	*BaseRepository
	cipher *secretcrypt.Cipher
}

// ErrMFASecretUnreadable is returned when a stored secret cannot be decrypted
// (missing or wrong MFA_ENCRYPTION_KEY, or a tampered value). Callers must
// treat it as a failure, never as "MFA not set up".
var ErrMFASecretUnreadable = apperrors.New("MFA_SECRET_UNREADABLE", "MFA secret cannot be read", 500)

// NewMFASetupRepository creates an MFA setup repository without encryption:
// secrets are written as plaintext and encrypted rows cannot be read.
func NewMFASetupRepository(db *gorm.DB) IMFASetupRepository {
	return NewMFASetupRepositoryWithCipher(db, nil)
}

// NewMFASetupRepositoryWithCipher creates an MFA setup repository that
// encrypts secrets with c. A nil c behaves like NewMFASetupRepository.
func NewMFASetupRepositoryWithCipher(db *gorm.DB, c *secretcrypt.Cipher) IMFASetupRepository {
	return &MFASetupRepository{
		BaseRepository: NewBaseRepository(db),
		cipher:         c,
	}
}

// mfaSecretAAD binds a ciphertext to its owner, so a value copied into
// another user's row does not decrypt.
func mfaSecretAAD(userID uuid.UUID) []byte {
	return []byte("mfa_setups.secret_key:" + userID.String())
}

// sealed returns a copy of mfa with SecretKey encrypted for storage.
func (r *MFASetupRepository) sealed(mfa *models.MFASetup) (*models.MFASetup, error) {
	row := *mfa
	row.User = nil
	enc, err := r.cipher.Encrypt(mfa.SecretKey, mfaSecretAAD(mfa.UserID))
	if err != nil {
		return nil, apperrors.Wrap("MFA_SECRET_ENCRYPT_FAILED", "MFA secret cannot be encrypted", 500, err)
	}
	row.SecretKey = enc
	return &row, nil
}

// open decrypts mfa.SecretKey in place.
func (r *MFASetupRepository) open(mfa *models.MFASetup) error {
	plain, err := r.cipher.Decrypt(mfa.SecretKey, mfaSecretAAD(mfa.UserID))
	if err != nil {
		return apperrors.Wrap(ErrMFASecretUnreadable.Code, ErrMFASecretUnreadable.Message, ErrMFASecretUnreadable.StatusCode,
			fmt.Errorf("mfa_setups %s: %w", mfa.ID, err))
	}
	mfa.SecretKey = plain
	return nil
}

// Create creates a new MFA setup. mfa.SecretKey stays plaintext for the caller.
func (r *MFASetupRepository) Create(mfa *models.MFASetup) error {
	row, err := r.sealed(mfa)
	if err != nil {
		return err
	}
	if err := r.BaseRepository.Create(row); err != nil {
		return err
	}
	mfa.ID, mfa.CreatedAt, mfa.UpdatedAt = row.ID, row.CreatedAt, row.UpdatedAt
	return nil
}

// Update updates an MFA setup. mfa.SecretKey stays plaintext for the caller.
func (r *MFASetupRepository) Update(mfa *models.MFASetup) error {
	row, err := r.sealed(mfa)
	if err != nil {
		return err
	}
	if err := r.BaseRepository.Update(row); err != nil {
		return err
	}
	mfa.ID, mfa.CreatedAt, mfa.UpdatedAt = row.ID, row.CreatedAt, row.UpdatedAt
	return nil
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
	if err := r.open(&mfa); err != nil {
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
	if err := r.open(&mfa); err != nil {
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
	if err := r.open(&mfa); err != nil {
		return nil, err
	}
	return &mfa, nil
}

// Enable enables MFA for a user
func (r *MFASetupRepository) Enable(userID uuid.UUID) error {
	return r.GetDB().Model(&models.MFASetup{}).Where("user_id = ?", userID).Update("is_enabled", true).Error
}

// EnablePending implements MFASetupRepositoryInterface.
// secretKey is plaintext. Ciphertexts differ per write (random nonce), so the
// stored values are decrypted and compared here; the UPDATE is then guarded by
// the exact stored value, which keeps it atomic against a concurrent setup.
func (r *MFASetupRepository) EnablePending(userID uuid.UUID, secretKey string) (bool, error) {
	var pending []models.MFASetup
	if err := r.GetDB().Where("user_id = ? AND is_enabled = ?", userID, false).Find(&pending).Error; err != nil {
		return false, err
	}
	for _, row := range pending {
		stored := row.SecretKey
		if err := r.open(&row); err != nil {
			return false, err
		}
		if subtle.ConstantTimeCompare([]byte(row.SecretKey), []byte(secretKey)) != 1 {
			continue
		}
		res := r.GetDB().Model(&models.MFASetup{}).
			Where("id = ? AND secret_key = ? AND is_enabled = ?", row.ID, stored, false).
			Update("is_enabled", true)
		if res.Error != nil {
			return false, res.Error
		}
		if res.RowsAffected > 0 {
			return true, nil
		}
	}
	return false, nil
}

// Disable disables MFA for a user
func (r *MFASetupRepository) Disable(userID uuid.UUID) error {
	return r.GetDB().Model(&models.MFASetup{}).Where("user_id = ?", userID).Update("is_enabled", false).Error
}
