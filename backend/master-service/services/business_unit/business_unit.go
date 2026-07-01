package business_unit
import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/business_unit"
	"github.com/google/uuid"
)
type BusinessUnitServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.BusinessUnit, error)
	FindById(ctx *base.BaseService, id string) (*models.BusinessUnit, error)
	Create(ctx *base.BaseService, bu *models.BusinessUnit) (*models.BusinessUnit, error)
	Update(ctx *base.BaseService, id string, bu *models.BusinessUnit) (*models.BusinessUnit, error)
	Delete(ctx *base.BaseService, id string) error
}
type BusinessUnitService struct { buRepo repo.IBusinessUnitRepository }
func NewBusinessUnitService(buRepo repo.IBusinessUnitRepository) BusinessUnitServiceInterface { return &BusinessUnitService{buRepo: buRepo} }
func (s *BusinessUnitService) Create(ctx *base.BaseService, bu *models.BusinessUnit) (*models.BusinessUnit, error) {
	if _, err := s.buRepo.FindByCode(bu.BusinessUnitCode); err == nil {
		return nil, apperrors.Wrap("BUSINESS_UNIT_CODE_ALREADY_EXISTS", "Business unit code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate business unit code", 500, err)
	}
	if err := s.buRepo.Create(bu); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create business unit", 500, err) }
	return bu, nil
}
func (s *BusinessUnitService) Delete(ctx *base.BaseService, id string) error {
	buID, err := uuid.Parse(id)
	if err != nil { return apperrors.Wrap("INVALID_BUSINESS_UNIT_ID", "Invalid business unit ID format", 400, err) }
	if _, err := s.buRepo.FindByID(buID); err != nil {
		if err == apperrors.ErrNotFound { return err }
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find business unit", 500, err)
	}
	if err := s.buRepo.Delete(buID); err != nil { return apperrors.Wrap("DATABASE_ERROR", "Failed to delete business unit", 500, err) }
	return nil
}
func (s *BusinessUnitService) FindAll(ctx *base.BaseService) (*[]models.BusinessUnit, error) {
	bus, err := s.buRepo.FindAll()
	if err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch business units", 500, err) }
	result := make([]models.BusinessUnit, 0, len(bus))
	for _, bu := range bus { if bu != nil { result = append(result, *bu) } }
	return &result, nil
}
func (s *BusinessUnitService) FindById(ctx *base.BaseService, id string) (*models.BusinessUnit, error) {
	buID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_BUSINESS_UNIT_ID", "Invalid business unit ID format", 400, err) }
	bu, err := s.buRepo.FindByID(buID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch business unit", 500, err)
	}
	return bu, nil
}
func (s *BusinessUnitService) Update(ctx *base.BaseService, id string, bu *models.BusinessUnit) (*models.BusinessUnit, error) {
	buID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_BUSINESS_UNIT_ID", "Invalid business unit ID format", 400, err) }
	existingBU, err := s.buRepo.FindByID(buID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch business unit", 500, err)
	}
	if existingBU.BusinessUnitCode != bu.BusinessUnitCode {
		if _, err := s.buRepo.FindByCode(bu.BusinessUnitCode); err == nil {
			return nil, apperrors.Wrap("BUSINESS_UNIT_CODE_ALREADY_EXISTS", "Business unit code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate business unit code", 500, err) }
	}
	existingBU.BusinessUnitCode = bu.BusinessUnitCode
	existingBU.BusinessUnitName = bu.BusinessUnitName
	existingBU.BusinessUnitDescription = bu.BusinessUnitDescription
	existingBU.CostCenterCode = bu.CostCenterCode
	existingBU.ParentBusinessUnitID = bu.ParentBusinessUnitID
	existingBU.IsActive = bu.IsActive
	existingBU.CompanyID = bu.CompanyID
	if err := s.buRepo.Update(existingBU); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update business unit", 500, err) }
	return existingBU, nil
}
