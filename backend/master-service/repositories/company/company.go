package company

import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ICompanyRepository interface {
	Create(company *models.Company) error
	Update(company *models.Company) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.Company, error)
	FindByCode(code string) (*models.Company, error)
	FindAll() ([]*models.Company, error)
}
type CompanyRepository struct{ *repositories.BaseRepository }

func NewCompanyRepository(db *gorm.DB) ICompanyRepository {
	return &CompanyRepository{BaseRepository: repositories.NewBaseRepository(db)}
}
func (r *CompanyRepository) Create(company *models.Company) error {
	return r.BaseRepository.Create(company)
}
func (r *CompanyRepository) Update(company *models.Company) error {
	return r.BaseRepository.Update(company)
}
func (r *CompanyRepository) Delete(id uuid.UUID) error {
	return r.BaseRepository.Delete(&models.Company{ID: id})
}
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
func (r *CompanyRepository) FindByCode(code string) (*models.Company, error) {
	var company models.Company
	if err := r.GetDB().Where("company_code = ?", code).First(&company).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &company, nil
}
func (r *CompanyRepository) FindAll() ([]*models.Company, error) {
	var companies []*models.Company
	if err := r.GetDB().Find(&companies).Error; err != nil {
		return nil, err
	}
	return companies, nil
}
