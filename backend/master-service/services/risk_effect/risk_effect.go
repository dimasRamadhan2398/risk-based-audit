package risk_effect
import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/risk_effect"
	"github.com/google/uuid"
)
type RiskEffectServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.RiskEffect, error)
	FindById(ctx *base.BaseService, id string) (*models.RiskEffect, error)
	Create(ctx *base.BaseService, effect *models.RiskEffect) (*models.RiskEffect, error)
	Update(ctx *base.BaseService, id string, effect *models.RiskEffect) (*models.RiskEffect, error)
	Delete(ctx *base.BaseService, id string) error
}
type RiskEffectService struct { effectRepo repo.IRiskEffectRepository }
func NewRiskEffectService(effectRepo repo.IRiskEffectRepository) RiskEffectServiceInterface { return &RiskEffectService{effectRepo: effectRepo} }
func (s *RiskEffectService) Create(ctx *base.BaseService, effect *models.RiskEffect) (*models.RiskEffect, error) {
	if _, err := s.effectRepo.FindByCode(effect.EffectCode); err == nil {
		return nil, apperrors.Wrap("EFFECT_CODE_ALREADY_EXISTS", "Effect code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate effect code", 500, err)
	}
	if err := s.effectRepo.Create(effect); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create risk effect", 500, err) }
	return s.effectRepo.FindByID(effect.ID)
}
func (s *RiskEffectService) Delete(ctx *base.BaseService, id string) error {
	effectID, err := uuid.Parse(id)
	if err != nil { return apperrors.Wrap("INVALID_EFFECT_ID", "Invalid effect ID format", 400, err) }
	if _, err := s.effectRepo.FindByID(effectID); err != nil {
		if err == apperrors.ErrNotFound { return err }
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find effect", 500, err)
	}
	if err := s.effectRepo.Delete(effectID); err != nil { return apperrors.Wrap("DATABASE_ERROR", "Failed to delete effect", 500, err) }
	return nil
}
func (s *RiskEffectService) FindAll(ctx *base.BaseService) (*[]models.RiskEffect, error) {
	effects, err := s.effectRepo.FindAll()
	if err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch risk effects", 500, err) }
	result := make([]models.RiskEffect, 0, len(effects))
	for _, effect := range effects { if effect != nil { result = append(result, *effect) } }
	return &result, nil
}
func (s *RiskEffectService) FindById(ctx *base.BaseService, id string) (*models.RiskEffect, error) {
	effectID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_EFFECT_ID", "Invalid effect ID format", 400, err) }
	effect, err := s.effectRepo.FindByID(effectID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch effect", 500, err)
	}
	return effect, nil
}
func (s *RiskEffectService) Update(ctx *base.BaseService, id string, effect *models.RiskEffect) (*models.RiskEffect, error) {
	effectID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_EFFECT_ID", "Invalid effect ID format", 400, err) }
	existingEffect, err := s.effectRepo.FindByID(effectID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch effect", 500, err)
	}
	if existingEffect.EffectCode != effect.EffectCode {
		if _, err := s.effectRepo.FindByCode(effect.EffectCode); err == nil {
			return nil, apperrors.Wrap("EFFECT_CODE_ALREADY_EXISTS", "Effect code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate effect code", 500, err)
		}
	}
	existingEffect.EffectCode = effect.EffectCode
	existingEffect.EffectName = effect.EffectName
	existingEffect.Description = effect.Description
	existingEffect.Category = effect.Category
	existingEffect.RiskRegisterID = effect.RiskRegisterID
	existingEffect.IsActive = effect.IsActive
	if err := s.effectRepo.Update(existingEffect); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update effect", 500, err) }
	return s.effectRepo.FindByID(effectID)
}
