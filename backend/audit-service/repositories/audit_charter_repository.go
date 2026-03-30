package features

import (
	"github.com/google/uuid"
	"github.com/risk-based-audit/backend/internal/auth/models/features"
	"github.com/risk-based-audit/backend/internal/repository"
	apperrors "github.com/risk-based-audit/backend/pkg/errors"
	"gorm.io/gorm"
)

// AuditCharterRepositoryInterface defines the audit charter repository interface
type AuditCharterRepositoryInterface interface {
	Create(charter *features.AuditCharter) error
	Update(charter *features.AuditCharter) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*features.AuditCharter, error)
	FindByVersion(version string) (*features.AuditCharter, error)
	FindActive() (*features.AuditCharter, error)
	FindMany(offset, limit int, search string, isActive *bool) ([]*features.AuditCharter, error)
	Count(search string, isActive *bool) (int64, error)
}

// AuditCharterRepository handles audit charter data operations
type AuditCharterRepository struct {
	*repository.BaseRepository
}

// NewAuditCharterRepository creates a new audit charter repository
func NewAuditCharterRepository(baseRepo *repository.BaseRepository) AuditCharterRepositoryInterface {
	return &AuditCharterRepository{
		BaseRepository: baseRepo,
	}
}

// Create creates a new audit charter
func (r *AuditCharterRepository) Create(charter *features.AuditCharter) error {
	return r.BaseRepository.Create(charter)
}

// Update updates an audit charter
func (r *AuditCharterRepository) Update(charter *features.AuditCharter) error {
	return r.BaseRepository.Update(charter)
}

// Delete deletes an audit charter
func (r *AuditCharterRepository) Delete(id uuid.UUID) error {
	return r.BaseRepository.Delete(&features.AuditCharter{ID: id})
}

// FindByID finds an audit charter by ID
func (r *AuditCharterRepository) FindByID(id uuid.UUID) (*features.AuditCharter, error) {
	var charter features.AuditCharter
	if err := r.GetDB().First(&charter, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &charter, nil
}

// FindByVersion finds an audit charter by version
func (r *AuditCharterRepository) FindByVersion(version string) (*features.AuditCharter, error) {
	var charter features.AuditCharter
	if err := r.GetDB().Where("version = ?", version).First(&charter).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &charter, nil
}

// FindActive finds the active audit charter
func (r *AuditCharterRepository) FindActive() (*features.AuditCharter, error) {
	var charter features.AuditCharter
	if err := r.GetDB().Where("is_active = ?", true).Order("created_at DESC").First(&charter).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &charter, nil
}

// FindMany finds multiple audit charters with filters
func (r *AuditCharterRepository) FindMany(offset, limit int, search string, isActive *bool) ([]*features.AuditCharter, error) {
	var charters []*features.AuditCharter
	query := r.GetDB().Model(&features.AuditCharter{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("version LIKE ? OR filename LIKE ? OR description LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&charters).Error; err != nil {
		return nil, err
	}

	return charters, nil
}

// Count counts audit charters with filters
func (r *AuditCharterRepository) Count(search string, isActive *bool) (int64, error) {
	var count int64
	query := r.GetDB().Model(&features.AuditCharter{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("version LIKE ? OR filename LIKE ? OR description LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}
