package risk_category

import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/risk_category"

	"github.com/google/uuid"
)

type RiskCategoryServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.RiskCategory, error)
	FindById(ctx *base.BaseService, id string) (*models.RiskCategory, error)
	Create(ctx *base.BaseService, cat *models.RiskCategory) (*models.RiskCategory, error)
	Update(ctx *base.BaseService, id string, cat *models.RiskCategory) (*models.RiskCategory, error)
	Delete(ctx *base.BaseService, id string) error
}

type RiskCategoryService struct {
	catRepo repo.IRiskCategoryRepository
}

func NewRiskCategoryService(catRepo repo.IRiskCategoryRepository) RiskCategoryServiceInterface {
	return &RiskCategoryService{
		catRepo: catRepo,
	}
}

func (s *RiskCategoryService) Create(ctx *base.BaseService, cat *models.RiskCategory) (*models.RiskCategory, error) {
	if _, err := s.catRepo.FindByCode(cat.Code); err == nil {
		return nil, apperrors.Wrap("RISK_CATEGORY_CODE_ALREADY_EXISTS", "Risk category code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate risk category code", 500, err)
	}

	if err := s.catRepo.Create(cat); err != nil {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create risk category", 500, err)
	}

	return cat, nil
}

func (s *RiskCategoryService) Delete(ctx *base.BaseService, id string) error {
	catID, err := uuid.Parse(id)
	if err != nil {
		return apperrors.Wrap("INVALID_RISK_CATEGORY_ID", "Invalid risk category ID format", 400, err)
	}

	if _, err := s.catRepo.FindByID(catID); err != nil {
		if err == apperrors.ErrNotFound {
			return err
		}
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find risk category", 500, err)
	}

	if err := s.catRepo.Delete(catID); err != nil {
		return apperrors.Wrap("DATABASE_ERROR", "Failed to delete risk category", 500, err)
	}

	return nil
}

func (s *RiskCategoryService) FindAll(ctx *base.BaseService) (*[]models.RiskCategory, error) {
	cats, err := s.catRepo.FindAll()
	if err != nil {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch risk categories", 500, err)
	}

	result := make([]models.RiskCategory, 0, len(cats))
	for _, cat := range cats {
		if cat != nil {
			result = append(result, *cat)
		}
	}

	return &result, nil
}

func (s *RiskCategoryService) FindById(ctx *base.BaseService, id string) (*models.RiskCategory, error) {
	catID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.Wrap("INVALID_RISK_CATEGORY_ID", "Invalid risk category ID format", 400, err)
	}

	cat, err := s.catRepo.FindByID(catID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			return nil, err
		}
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch risk category", 500, err)
	}

	return cat, nil
}

func (s *RiskCategoryService) Update(ctx *base.BaseService, id string, cat *models.RiskCategory) (*models.RiskCategory, error) {
	catID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.Wrap("INVALID_RISK_CATEGORY_ID", "Invalid risk category ID format", 400, err)
	}

	existingCat, err := s.catRepo.FindByID(catID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			return nil, err
		}
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch risk category", 500, err)
	}

	if existingCat.Code != cat.Code {
		if _, err := s.catRepo.FindByCode(cat.Code); err == nil {
			return nil, apperrors.Wrap("RISK_CATEGORY_CODE_ALREADY_EXISTS", "Risk category code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate risk category code", 500, err)
		}
	}

	existingCat.Code = cat.Code
	existingCat.Name = cat.Name
	existingCat.Description = cat.Description
	existingCat.IsActive = cat.IsActive

	if err := s.catRepo.Update(existingCat); err != nil {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update risk category", 500, err)
	}

	return existingCat, nil
}
