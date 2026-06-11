package repositories

import (
	"audit-service/models"
	apperrors "audit-service/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuditMandateRepositoryInterface defines the audit mandate repository interface
type AuditMandateRepositoryInterface interface {
	Create(mandate *models.AuditMandate) error
	Update(mandate *models.AuditMandate) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.AuditMandate, error)
	FindByReferenceNumber(refNumber string) (*models.AuditMandate, error)
	FindActive() (*models.AuditMandate, error)
	FindMany(offset, limit int, search string, isActive *bool) ([]*models.AuditMandate, error)
	Count(search string, isActive *bool) (int64, error)
}

// AuditMandateRepository handles audit mandate data operations
type AuditMandateRepository struct {
	*BaseRepository
}

// NewAuditMandateRepository creates a new audit mandate repository
func NewAuditMandateRepository(baseRepo *BaseRepository) AuditMandateRepositoryInterface {
	return &AuditMandateRepository{
		BaseRepository: baseRepo,
	}
}

// Create creates a new audit mandate
func (r *AuditMandateRepository) Create(mandate *models.AuditMandate) error {
	if err := r.DB.Create(mandate).Error; err != nil {
		return apperrors.ErrDatabase
	}
	return nil
}

// Update updates an audit mandate
func (r *AuditMandateRepository) Update(mandate *models.AuditMandate) error {
	if err := r.DB.Save(mandate).Error; err != nil {
		return apperrors.ErrDatabase
	}
	return nil
}

// Delete deletes an audit mandate (soft delete)
func (r *AuditMandateRepository) Delete(id uuid.UUID) error {
	result := r.DB.Delete(&models.AuditMandate{}, "id = ?", id)
	if result.Error != nil {
		return apperrors.ErrDatabase
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// FindByID finds an audit mandate by ID
func (r *AuditMandateRepository) FindByID(id uuid.UUID) (*models.AuditMandate, error) {
	var mandate models.AuditMandate
	if err := r.DB.First(&mandate, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &mandate, nil
}

// FindByReferenceNumber finds an audit mandate by reference number
func (r *AuditMandateRepository) FindByReferenceNumber(refNumber string) (*models.AuditMandate, error) {
	var mandate models.AuditMandate
	if err := r.DB.Where("reference_number = ?", refNumber).First(&mandate).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &mandate, nil
}

// FindActive finds the active audit mandate
func (r *AuditMandateRepository) FindActive() (*models.AuditMandate, error) {
	var mandate models.AuditMandate
	if err := r.DB.Where("is_active = ?", true).Order("effective_date DESC").First(&mandate).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &mandate, nil
}

// FindMany finds multiple audit mandates with filters
func (r *AuditMandateRepository) FindMany(offset, limit int, search string, isActive *bool) ([]*models.AuditMandate, error) {
	var mandates []*models.AuditMandate
	query := r.DB.Model(&models.AuditMandate{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("title ILIKE ? OR reference_number ILIKE ? OR mandate_source ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&mandates).Error; err != nil {
		return nil, err
	}

	return mandates, nil
}

// Count counts audit mandates with filters
func (r *AuditMandateRepository) Count(search string, isActive *bool) (int64, error) {
	var count int64
	query := r.DB.Model(&models.AuditMandate{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("title ILIKE ? OR reference_number ILIKE ? OR mandate_source ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}