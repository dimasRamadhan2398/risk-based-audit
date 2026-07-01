package risk_level
import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/risk_level"
	"github.com/google/uuid"
)
type RiskLevelServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.RiskLevel, error)
	FindById(ctx *base.BaseService, id string) (*models.RiskLevel, error)
	Create(ctx *base.BaseService, rl *models.RiskLevel) (*models.RiskLevel, error)
	Update(ctx *base.BaseService, id string, rl *models.RiskLevel) (*models.RiskLevel, error)
	Delete(ctx *base.BaseService, id string) error
}
type RiskLevelService struct { rlRepo repo.IRiskLevelRepository }
func NewRiskLevelService(rlRepo repo.IRiskLevelRepository) RiskLevelServiceInterface { return &RiskLevelService{rlRepo: rlRepo} }
func (s *RiskLevelService) Create(ctx *base.BaseService, rl *models.RiskLevel) (*models.RiskLevel, error) {
	if _, err := s.rlRepo.FindByCode(rl.RiskCode); err == nil {
		return nil, apperrors.Wrap("RISK_LEVEL_CODE_ALREADY_EXISTS", "Risk level code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate risk level code", 500, err)
	}
	if err := s.rlRepo.Create(rl); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create risk level", 500, err) }
	return rl, nil
}
func (s *RiskLevelService) Delete(ctx *base.BaseService, id string) error {
	rlID, err := uuid.Parse(id)
	if err != nil { return apperrors.Wrap("INVALID_RISK_LEVEL_ID", "Invalid risk level ID format", 400, err) }
	if _, err := s.rlRepo.FindByID(rlID); err != nil {
		if err == apperrors.ErrNotFound { return err }
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find risk level", 500, err)
	}
	if err := s.rlRepo.Delete(rlID); err != nil { return apperrors.Wrap("DATABASE_ERROR", "Failed to delete risk level", 500, err) }
	return nil
}
func (s *RiskLevelService) FindAll(ctx *base.BaseService) (*[]models.RiskLevel, error) {
	rls, err := s.rlRepo.FindAll()
	if err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch risk levels", 500, err) }
	result := make([]models.RiskLevel, 0, len(rls))
	for _, rl := range rls { if rl != nil { result = append(result, *rl) } }
	return &result, nil
}
func (s *RiskLevelService) FindById(ctx *base.BaseService, id string) (*models.RiskLevel, error) {
	rlID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_RISK_LEVEL_ID", "Invalid risk level ID format", 400, err) }
	rl, err := s.rlRepo.FindByID(rlID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch risk level", 500, err)
	}
	return rl, nil
}
func (s *RiskLevelService) Update(ctx *base.BaseService, id string, rl *models.RiskLevel) (*models.RiskLevel, error) {
	rlID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_RISK_LEVEL_ID", "Invalid risk level ID format", 400, err) }
	existingRL, err := s.rlRepo.FindByID(rlID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch risk level", 500, err)
	}
	if existingRL.RiskCode != rl.RiskCode {
		if _, err := s.rlRepo.FindByCode(rl.RiskCode); err == nil {
			return nil, apperrors.Wrap("RISK_LEVEL_CODE_ALREADY_EXISTS", "Risk level code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate risk level code", 500, err)
		}
	}
	existingRL.RiskCode = rl.RiskCode
	existingRL.RiskName = rl.RiskName
	existingRL.RiskDescription = rl.RiskDescription
	existingRL.MinScore = rl.MinScore
	existingRL.MaxScore = rl.MaxScore
	existingRL.Color = rl.Color
	existingRL.IsActive = rl.IsActive
	if err := s.rlRepo.Update(existingRL); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update risk level", 500, err) }
	return existingRL, nil
}
