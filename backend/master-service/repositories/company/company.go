package employee

import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ICompanyRepository = CompanyRepositoryInterface

// CompanyRepositoryInterface defines the company repository interface
type CompanyRepositoryInterface interface {
	Create(company *models.Company) error
	Update(company *models.Company) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.Company, error)
	FindByName(name string) (*models.Company, error)
	FindByResourceAndAction(resource, action string) (*models.Company, error)
	FindAll() ([]*models.Company, error)
	FindMany(offset, limit int, search string) ([]*models.Company, error)
	Count(search string) (int64, error)
	LocationExists(id uuid.UUID) (bool, error)
}

// CompanyRepository handles company data operations
type CompanyRepository struct {
	*repositories.BaseRepository
}

// NewCompanyRepository creates a new company repository
func NewCompanyRepository(db *gorm.DB) ICompanyRepository {
	return &CompanyRepository{
		BaseRepository: repositories.NewBaseRepository(db),
	}
}

// Create creates a new company
func (r *CompanyRepository) Create(company *models.Company) error {
	return r.BaseRepository.Create(company)
}

// Update updates a company
func (r *CompanyRepository) Update(company *models.Company) error {
	return r.BaseRepository.Update(company)
}

// Delete deletes a company
func (r *CompanyRepository) Delete(id uuid.UUID) error {
	return r.BaseRepository.Delete(&models.Company{ID: id})
}

// FindByID finds a company by ID
func (r *CompanyRepository) FindByID(id uuid.UUID) (*models.Company, error) {
	var company models.Company
	if err := r.GetDB().First(&company, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &company, nil
}

// FindByName finds a company by name
func (r *CompanyRepository) FindByName(name string) (*models.Company, error) {
	var company models.Company
	if err := r.GetDB().Where("name = ?", name).First(&company).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &company, nil
}

// FindByResourceAndAction finds a company by resource and action
func (r *CompanyRepository) FindByResourceAndAction(resource, action string) (*models.Company, error) {
	var company models.Company
	if err := r.GetDB().Where("resource = ? AND action = ?", resource, action).First(&company).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &company, nil
}

// FindAll finds all companys
func (r *CompanyRepository) FindAll() ([]*models.Company, error) {
	var companys []*models.Company
	if err := r.GetDB().Find(&companys).Error; err != nil {
		return nil, err
	}
	return companys, nil
}

// FindMany finds multiple companys with pagination
func (r *CompanyRepository) FindMany(offset, limit int, search string) ([]*models.Company, error) {
	var companys []*models.Company
	query := r.GetDB().Model(&models.Company{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name LIKE ? OR resource LIKE ? OR action LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if err := query.Offset(offset).Limit(limit).Find(&companys).Error; err != nil {
		return nil, err
	}

	return companys, nil
}

// Count counts companys with filters
func (r *CompanyRepository) Count(search string) (int64, error) {
	var count int64
	query := r.GetDB().Model(&models.Company{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name LIKE ? OR resource LIKE ? OR action LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// LocationExists checks if a location exists by ID
func (r *CompanyRepository) LocationExists(id uuid.UUID) (bool, error) {
	var count int64
	if err := r.GetDB().Model(&models.Location{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
