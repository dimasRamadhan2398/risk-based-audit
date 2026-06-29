package risk_register
import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/risk_register"
	"github.com/google/uuid"
)
type RiskRegisterServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.RiskRegister, error)
	FindById(ctx *base.BaseService, id string) (*models.RiskRegister, error)
	Create(ctx *base.BaseService, risk *models.RiskRegister) (*models.RiskRegister, error)
	Update(ctx *base.BaseService, id string, risk *models.RiskRegister) (*models.RiskRegister, error)
	Delete(ctx *base.BaseService, id string) error
}
type RiskRegisterService struct { riskRepo repo.IRiskRegisterRepository }
func NewRiskRegisterService(riskRepo repo.IRiskRegisterRepository) RiskRegisterServiceInterface { return &RiskRegisterService{riskRepo: riskRepo} }
func (s *RiskRegisterService) Create(ctx *base.BaseService, risk *models.RiskRegister) (*models.RiskRegister, error) {
	if _, err := s.riskRepo.FindByCode(risk.Code); err == nil {
		return nil, apperrors.Wrap("RISK_CODE_ALREADY_EXISTS", "Risk code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate risk code", 500, err)
	}
	if err := s.riskRepo.Create(risk); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create risk", 500, err) }
	return s.riskRepo.FindByID(risk.ID)
}
func (s *RiskRegisterService) Delete(ctx *base.BaseService, id string) error {
	riskID, err := uuid.Parse(id)
	if err != nil { return apperrors.Wrap("INVALID_RISK_ID", "Invalid risk ID format", 400, err) }
	if _, err := s.riskRepo.FindByID(riskID); err != nil {
		if err == apperrors.ErrNotFound { return err }
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find risk", 500, err)
	}
	if err := s.riskRepo.Delete(riskID); err != nil { return apperrors.Wrap("DATABASE_ERROR", "Failed to delete risk", 500, err) }
	return nil
}
func (s *RiskRegisterService) FindAll(ctx *base.BaseService) (*[]models.RiskRegister, error) {
	risks, err := s.riskRepo.FindAll()
	if err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch risks", 500, err) }
	result := make([]models.RiskRegister, 0, len(risks))
	for _, risk := range risks { if risk != nil { result = append(result, *risk) } }
	return &result, nil
}
func (s *RiskRegisterService) FindById(ctx *base.BaseService, id string) (*models.RiskRegister, error) {
	riskID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_RISK_ID", "Invalid risk ID format", 400, err) }
	risk, err := s.riskRepo.FindByID(riskID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch risk", 500, err)
	}
	return risk, nil
}
func (s *RiskRegisterService) Update(ctx *base.BaseService, id string, risk *models.RiskRegister) (*models.RiskRegister, error) {
	riskID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_RISK_ID", "Invalid risk ID format", 400, err) }
	existingRisk, err := s.riskRepo.FindByID(riskID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch risk", 500, err)
	}
	if existingRisk.Code != risk.Code {
		if _, err := s.riskRepo.FindByCode(risk.Code); err == nil {
			return nil, apperrors.Wrap("RISK_CODE_ALREADY_EXISTS", "Risk code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate risk code", 500, err)
		}
	}
	existingRisk.Code = risk.Code
	existingRisk.Name = risk.Name
	existingRisk.Description = risk.Description
	existingRisk.DepartmentID = risk.DepartmentID
	existingRisk.RiskCategoryID = risk.RiskCategoryID
	existingRisk.RiskOwnerID = risk.RiskOwnerID
	existingRisk.InherentLikelihoodID = risk.InherentLikelihoodID
	existingRisk.InherentImpactID = risk.InherentImpactID
	existingRisk.InherentScore = risk.InherentScore
	existingRisk.InherentRiskLevelID = risk.InherentRiskLevelID
	existingRisk.ResidualLikelihoodID = risk.ResidualLikelihoodID
	existingRisk.ResidualImpactID = risk.ResidualImpactID
	existingRisk.ResidualScore = risk.ResidualScore
	existingRisk.ResidualRiskLevelID = risk.ResidualRiskLevelID
	existingRisk.AssessmentDate = risk.AssessmentDate
	existingRisk.NextReviewDate = risk.NextReviewDate
	existingRisk.IsActive = risk.IsActive
	if err := s.riskRepo.Update(existingRisk); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update risk", 500, err) }
	return s.riskRepo.FindByID(riskID)
}
