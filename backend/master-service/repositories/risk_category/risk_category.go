package risk_category

import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IRiskCategoryRepository = RiskCategoryRepositoryInterface

type RiskCategoryRepositoryInterface interface {
	Create(cat *models.RiskCategory) error
	Update(cat *models.RiskCategory) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.RiskCategory, error)
	FindByCode(code string) (*models.RiskCategory, error)
	FindAll() ([]*models.RiskCategory, error)
}

type RiskCategoryRepository struct {
	*repositories.BaseRepository
}

func NewRiskCategoryRepository(db *gorm.DB) IRiskCategoryRepository {
	return &RiskCategoryRepository{
		BaseRepository: repositories.NewBaseRepository(db),
	}
}

func (r *RiskCategoryRepository) Create(cat *models.RiskCategory) error {
	return r.BaseRepository.Create(cat)
}

func (r *RiskCategoryRepository) Update(cat *models.RiskCategory) error {
	return r.BaseRepository.Update(cat)
}

func (r *RiskCategoryRepository) Delete(id uuid.UUID) error {
	return r.BaseRepository.Delete(&models.RiskCategory{ID: id})
}

func (r *RiskCategoryRepository) FindByID(id uuid.UUID) (*models.RiskCategory, error) {
	var cat models.RiskCategory
	if err := r.GetDB().First(&cat, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &cat, nil
}

func (r *RiskCategoryRepository) FindByCode(code string) (*models.RiskCategory, error) {
	var cat models.RiskCategory
	if err := r.GetDB().Where("code = ?", code).First(&cat).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &cat, nil
}

func (r *RiskCategoryRepository) FindAll() ([]*models.RiskCategory, error) {
	var cats []*models.RiskCategory
	if err := r.GetDB().Find(&cats).Error; err != nil {
		return nil, err
	}
	return cats, nil
}
