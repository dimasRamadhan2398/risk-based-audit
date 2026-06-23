package impact
import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/impact"
	"github.com/google/uuid"
)
type ImpactServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.Impact, error)
	FindById(ctx *base.BaseService, id string) (*models.Impact, error)
	Create(ctx *base.BaseService, i *models.Impact) (*models.Impact, error)
	Update(ctx *base.BaseService, id string, i *models.Impact) (*models.Impact, error)
	Delete(ctx *base.BaseService, id string) error
}
type ImpactService struct { iRepo repo.IImpactRepository }
func NewImpactService(iRepo repo.IImpactRepository) ImpactServiceInterface { return &ImpactService{iRepo: iRepo} }
func (s *ImpactService) Create(ctx *base.BaseService, i *models.Impact) (*models.Impact, error) {
	if _, err := s.iRepo.FindByCode(i.Code); err == nil {
		return nil, apperrors.Wrap("IMPACT_CODE_ALREADY_EXISTS", "Impact code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate impact code", 500, err)
	}
	if err := s.iRepo.Create(i); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create impact", 500, err) }
	return i, nil
}
func (s *ImpactService) Delete(ctx *base.BaseService, id string) error {
	iID, err := uuid.Parse(id)
	if err != nil { return apperrors.Wrap("INVALID_IMPACT_ID", "Invalid impact ID format", 400, err) }
	if _, err := s.iRepo.FindByID(iID); err != nil {
		if err == apperrors.ErrNotFound { return err }
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find impact", 500, err)
	}
	if err := s.iRepo.Delete(iID); err != nil { return apperrors.Wrap("DATABASE_ERROR", "Failed to delete impact", 500, err) }
	return nil
}
func (s *ImpactService) FindAll(ctx *base.BaseService) (*[]models.Impact, error) {
	is, err := s.iRepo.FindAll()
	if err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch impacts", 500, err) }
	result := make([]models.Impact, 0, len(is))
	for _, i := range is { if i != nil { result = append(result, *i) } }
	return &result, nil
}
func (s *ImpactService) FindById(ctx *base.BaseService, id string) (*models.Impact, error) {
	iID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_IMPACT_ID", "Invalid impact ID format", 400, err) }
	i, err := s.iRepo.FindByID(iID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch impact", 500, err)
	}
	return i, nil
}
func (s *ImpactService) Update(ctx *base.BaseService, id string, i *models.Impact) (*models.Impact, error) {
	iID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_IMPACT_ID", "Invalid impact ID format", 400, err) }
	existingI, err := s.iRepo.FindByID(iID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch impact", 500, err)
	}
	if existingI.Code != i.Code {
		if _, err := s.iRepo.FindByCode(i.Code); err == nil {
			return nil, apperrors.Wrap("IMPACT_CODE_ALREADY_EXISTS", "Impact code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate impact code", 500, err)
		}
	}
	existingI.Code = i.Code
	existingI.Name = i.Name
	existingI.Score = i.Score
	existingI.Description = i.Description
	existingI.IsActive = i.IsActive
	if err := s.iRepo.Update(existingI); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update impact", 500, err) }
	return existingI, nil
}
