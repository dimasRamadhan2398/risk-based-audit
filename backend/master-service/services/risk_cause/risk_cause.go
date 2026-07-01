package risk_cause
import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/risk_cause"
	"github.com/google/uuid"
)
type RiskCauseServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.RiskCause, error)
	FindById(ctx *base.BaseService, id string) (*models.RiskCause, error)
	Create(ctx *base.BaseService, cause *models.RiskCause) (*models.RiskCause, error)
	Update(ctx *base.BaseService, id string, cause *models.RiskCause) (*models.RiskCause, error)
	Delete(ctx *base.BaseService, id string) error
}
type RiskCauseService struct { causeRepo repo.IRiskCauseRepository }
func NewRiskCauseService(causeRepo repo.IRiskCauseRepository) RiskCauseServiceInterface { return &RiskCauseService{causeRepo: causeRepo} }
func (s *RiskCauseService) Create(ctx *base.BaseService, cause *models.RiskCause) (*models.RiskCause, error) {
	if _, err := s.causeRepo.FindByCode(cause.CauseCode); err == nil {
		return nil, apperrors.Wrap("CAUSE_CODE_ALREADY_EXISTS", "Cause code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate cause code", 500, err)
	}
	if err := s.causeRepo.Create(cause); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create risk cause", 500, err) }
	return s.causeRepo.FindByID(cause.ID)
}
func (s *RiskCauseService) Delete(ctx *base.BaseService, id string) error {
	causeID, err := uuid.Parse(id)
	if err != nil { return apperrors.Wrap("INVALID_CAUSE_ID", "Invalid cause ID format", 400, err) }
	if _, err := s.causeRepo.FindByID(causeID); err != nil {
		if err == apperrors.ErrNotFound { return err }
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find cause", 500, err)
	}
	if err := s.causeRepo.Delete(causeID); err != nil { return apperrors.Wrap("DATABASE_ERROR", "Failed to delete cause", 500, err) }
	return nil
}
func (s *RiskCauseService) FindAll(ctx *base.BaseService) (*[]models.RiskCause, error) {
	causes, err := s.causeRepo.FindAll()
	if err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch risk causes", 500, err) }
	result := make([]models.RiskCause, 0, len(causes))
	for _, cause := range causes { if cause != nil { result = append(result, *cause) } }
	return &result, nil
}
func (s *RiskCauseService) FindById(ctx *base.BaseService, id string) (*models.RiskCause, error) {
	causeID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_CAUSE_ID", "Invalid cause ID format", 400, err) }
	cause, err := s.causeRepo.FindByID(causeID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch cause", 500, err)
	}
	return cause, nil
}
func (s *RiskCauseService) Update(ctx *base.BaseService, id string, cause *models.RiskCause) (*models.RiskCause, error) {
	causeID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_CAUSE_ID", "Invalid cause ID format", 400, err) }
	existingCause, err := s.causeRepo.FindByID(causeID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch cause", 500, err)
	}
	if existingCause.CauseCode != cause.CauseCode {
		if _, err := s.causeRepo.FindByCode(cause.CauseCode); err == nil {
			return nil, apperrors.Wrap("CAUSE_CODE_ALREADY_EXISTS", "Cause code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate cause code", 500, err)
		}
	}
	existingCause.CauseCode = cause.CauseCode
	existingCause.CauseName = cause.CauseName
	existingCause.Description = cause.Description
	existingCause.Category = cause.Category
	existingCause.RiskRegisterID = cause.RiskRegisterID
	existingCause.IsActive = cause.IsActive
	if err := s.causeRepo.Update(existingCause); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update cause", 500, err) }
	return s.causeRepo.FindByID(causeID)
}
