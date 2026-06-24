package likelihood
import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/likelihood"
	"github.com/google/uuid"
)
type LikelihoodServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.Likelihood, error)
	FindById(ctx *base.BaseService, id string) (*models.Likelihood, error)
	Create(ctx *base.BaseService, l *models.Likelihood) (*models.Likelihood, error)
	Update(ctx *base.BaseService, id string, l *models.Likelihood) (*models.Likelihood, error)
	Delete(ctx *base.BaseService, id string) error
}
type LikelihoodService struct { lRepo repo.ILikelihoodRepository }
func NewLikelihoodService(lRepo repo.ILikelihoodRepository) LikelihoodServiceInterface { return &LikelihoodService{lRepo: lRepo} }
func (s *LikelihoodService) Create(ctx *base.BaseService, l *models.Likelihood) (*models.Likelihood, error) {
	if _, err := s.lRepo.FindByCode(l.Code); err == nil {
		return nil, apperrors.Wrap("LIKELIHOOD_CODE_ALREADY_EXISTS", "Likelihood code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate likelihood code", 500, err)
	}
	if err := s.lRepo.Create(l); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create likelihood", 500, err) }
	return l, nil
}
func (s *LikelihoodService) Delete(ctx *base.BaseService, id string) error {
	lID, err := uuid.Parse(id)
	if err != nil { return apperrors.Wrap("INVALID_LIKELIHOOD_ID", "Invalid likelihood ID format", 400, err) }
	if _, err := s.lRepo.FindByID(lID); err != nil {
		if err == apperrors.ErrNotFound { return err }
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find likelihood", 500, err)
	}
	if err := s.lRepo.Delete(lID); err != nil { return apperrors.Wrap("DATABASE_ERROR", "Failed to delete likelihood", 500, err) }
	return nil
}
func (s *LikelihoodService) FindAll(ctx *base.BaseService) (*[]models.Likelihood, error) {
	ls, err := s.lRepo.FindAll()
	if err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch likelihoods", 500, err) }
	result := make([]models.Likelihood, 0, len(ls))
	for _, l := range ls { if l != nil { result = append(result, *l) } }
	return &result, nil
}
func (s *LikelihoodService) FindById(ctx *base.BaseService, id string) (*models.Likelihood, error) {
	lID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_LIKELIHOOD_ID", "Invalid likelihood ID format", 400, err) }
	l, err := s.lRepo.FindByID(lID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch likelihood", 500, err)
	}
	return l, nil
}
func (s *LikelihoodService) Update(ctx *base.BaseService, id string, l *models.Likelihood) (*models.Likelihood, error) {
	lID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_LIKELIHOOD_ID", "Invalid likelihood ID format", 400, err) }
	existingL, err := s.lRepo.FindByID(lID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch likelihood", 500, err)
	}
	if existingL.Code != l.Code {
		if _, err := s.lRepo.FindByCode(l.Code); err == nil {
			return nil, apperrors.Wrap("LIKELIHOOD_CODE_ALREADY_EXISTS", "Likelihood code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate likelihood code", 500, err)
		}
	}
	existingL.Code = l.Code
	existingL.Name = l.Name
	existingL.Score = l.Score
	existingL.Description = l.Description
	existingL.IsActive = l.IsActive
	if err := s.lRepo.Update(existingL); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update likelihood", 500, err) }
	return existingL, nil
}
