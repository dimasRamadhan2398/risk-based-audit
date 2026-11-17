package repositories

import (
	"time"

	"auth-service/models"
	apperrors "auth-service/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IConfidentialityAgreementRepository is an alias for the confidentiality agreement repository interface
type IConfidentialityAgreementRepository = ConfidentialityAgreementRepositoryInterface

// ConfidentialityAgreementRepositoryInterface defines the confidentiality agreement repository interface
type ConfidentialityAgreementRepositoryInterface interface {
	Create(agreement *models.ConfidentialityAgreement) error
	Update(agreement *models.ConfidentialityAgreement) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.ConfidentialityAgreement, error)
	FindLatestByUserID(userID uuid.UUID, agreementType string) (*models.ConfidentialityAgreement, error)
	FindAllByUserID(userID uuid.UUID, offset, limit int) ([]*models.ConfidentialityAgreement, error)
	HasAcceptedLatest(userID uuid.UUID) (bool, error)
	CountByUserID(userID uuid.UUID) (int64, error)
	DeleteExpired() error
	FindAcceptanceByEmployeeCodes(employeeCodes []string, agreementType models.AgreementType, version string, now time.Time) (*EmployeeAcceptance, error)
}

// EmployeeAcceptance is the result of FindAcceptanceByEmployeeCodes, keyed by
// users.employee_id (= master employees.employee_code).
type EmployeeAcceptance struct {
	// Linked holds the codes that have an active, non-deleted account.
	Linked map[string]bool
	// AcceptedAt holds, for linked codes with a valid acceptance, the latest
	// accepted_at of that agreement type and version.
	AcceptedAt map[string]time.Time
}

// ConfidentialityAgreementRepository handles confidentiality agreement data operations
type ConfidentialityAgreementRepository struct {
	*BaseRepository
}

// NewConfidentialityAgreementRepository creates a new confidentiality agreement repository
func NewConfidentialityAgreementRepository(db *gorm.DB) IConfidentialityAgreementRepository {
	return &ConfidentialityAgreementRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new confidentiality agreement
func (r *ConfidentialityAgreementRepository) Create(agreement *models.ConfidentialityAgreement) error {
	return r.BaseRepository.Create(agreement)
}

// Update updates a confidentiality agreement
func (r *ConfidentialityAgreementRepository) Update(agreement *models.ConfidentialityAgreement) error {
	return r.BaseRepository.Update(agreement)
}

// Delete deletes a confidentiality agreement
func (r *ConfidentialityAgreementRepository) Delete(id uuid.UUID) error {
	return r.BaseRepository.Delete(&models.ConfidentialityAgreement{ID: id})
}

// FindByID finds a confidentiality agreement by ID
func (r *ConfidentialityAgreementRepository) FindByID(id uuid.UUID) (*models.ConfidentialityAgreement, error) {
	var agreement models.ConfidentialityAgreement
	if err := r.GetDB().Preload("User").First(&agreement, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &agreement, nil
}

// FindLatestByUserID finds the latest confidentiality agreement for a user
func (r *ConfidentialityAgreementRepository) FindLatestByUserID(userID uuid.UUID, agreementType string) (*models.ConfidentialityAgreement, error) {
	var agreement models.ConfidentialityAgreement
	query := r.GetDB().Where("user_id = ?", userID).Where("is_accepted = ?", true)

	if agreementType != "" {
		query = query.Where("agreement_type = ?", agreementType)
	}

	if err := query.Order("accepted_at DESC").First(&agreement).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &agreement, nil
}

// FindAllByUserID finds all confidentiality agreements for a user
func (r *ConfidentialityAgreementRepository) FindAllByUserID(userID uuid.UUID, offset, limit int) ([]*models.ConfidentialityAgreement, error) {
	var agreements []*models.ConfidentialityAgreement
	if err := r.GetDB().Preload("User").Where("user_id = ?", userID).Offset(offset).Limit(limit).Order("created_at DESC").Find(&agreements).Error; err != nil {
		return nil, err
	}
	return agreements, nil
}

// HasAcceptedLatest checks if user has accepted the latest version of the agreement
func (r *ConfidentialityAgreementRepository) HasAcceptedLatest(userID uuid.UUID) (bool, error) {
	var agreement models.ConfidentialityAgreement
	err := r.GetDB().
		Where("user_id = ?", userID).
		Where("is_accepted = ?", true).
		Order("created_at DESC").
		First(&agreement).Error

	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	// Check if agreement is expired
	if agreement.ExpiresAt != nil && agreement.ExpiresAt.Before(time.Now()) {
		return false, nil
	}

	return true, nil
}

// CountByUserID counts confidentiality agreements by user ID
func (r *ConfidentialityAgreementRepository) CountByUserID(userID uuid.UUID) (int64, error) {
	var count int64
	if err := r.GetDB().Model(&models.ConfidentialityAgreement{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// DeleteExpired deletes expired confidentiality agreements
func (r *ConfidentialityAgreementRepository) DeleteExpired() error {
	result := r.GetDB().Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).Delete(&models.ConfidentialityAgreement{})
	return result.Error
}

// FindAcceptanceByEmployeeCodes looks up, for each employee code, the active
// account linked through users.employee_id and its accepted, unexpired
// agreements of agreementType at exactly version. Soft-deleted and inactive
// accounts and soft-deleted agreements do not count. Codes are bound as
// parameters; the caller caps their number.
func (r *ConfidentialityAgreementRepository) FindAcceptanceByEmployeeCodes(employeeCodes []string, agreementType models.AgreementType, version string, now time.Time) (*EmployeeAcceptance, error) {
	out := &EmployeeAcceptance{Linked: map[string]bool{}, AcceptedAt: map[string]time.Time{}}
	if len(employeeCodes) == 0 {
		return out, nil
	}

	var linked []string
	if err := r.GetDB().Model(&models.User{}).
		Where("employee_id IN ?", employeeCodes).
		Where("is_active = ?", true).
		Pluck("employee_id", &linked).Error; err != nil {
		return nil, err
	}
	for _, code := range linked {
		out.Linked[code] = true
	}
	if len(linked) == 0 {
		return out, nil
	}

	var rows []struct {
		EmployeeID string
		AcceptedAt *time.Time
	}
	if err := r.GetDB().Model(&models.ConfidentialityAgreement{}).
		Select("users.employee_id AS employee_id, confidentiality_agreements.accepted_at AS accepted_at").
		Joins("JOIN users ON users.id = confidentiality_agreements.user_id AND users.deleted_at IS NULL").
		Where("users.employee_id IN ?", linked).
		Where("users.is_active = ?", true).
		Where("confidentiality_agreements.agreement_type = ?", string(agreementType)).
		Where("confidentiality_agreements.version = ?", version).
		Where("confidentiality_agreements.is_accepted = ?", true).
		Where("confidentiality_agreements.expires_at IS NULL OR confidentiality_agreements.expires_at > ?", now).
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row.AcceptedAt == nil {
			// Accepted without a timestamp: count it, report the zero time as null.
			if _, seen := out.AcceptedAt[row.EmployeeID]; !seen {
				out.AcceptedAt[row.EmployeeID] = time.Time{}
			}
			continue
		}
		if prev, seen := out.AcceptedAt[row.EmployeeID]; !seen || row.AcceptedAt.After(prev) {
			out.AcceptedAt[row.EmployeeID] = *row.AcceptedAt
		}
	}
	return out, nil
}
