package risk_effect
import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)
type IRiskEffectRepository interface {
	Create(effect *models.RiskEffect) error
	Update(effect *models.RiskEffect) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.RiskEffect, error)
	FindByCode(code string) (*models.RiskEffect, error)
	FindAll() ([]*models.RiskEffect, error)
}
type RiskEffectRepository struct { *repositories.BaseRepository }
func NewRiskEffectRepository(db *gorm.DB) IRiskEffectRepository { return &RiskEffectRepository{BaseRepository: repositories.NewBaseRepository(db)} }
func (r *RiskEffectRepository) Create(effect *models.RiskEffect) error { return r.BaseRepository.Create(effect) }
func (r *RiskEffectRepository) Update(effect *models.RiskEffect) error { return r.BaseRepository.Update(effect) }
func (r *RiskEffectRepository) Delete(id uuid.UUID) error { return r.BaseRepository.Delete(&models.RiskEffect{ID: id}) }
func (r *RiskEffectRepository) FindByID(id uuid.UUID) (*models.RiskEffect, error) {
	var effect models.RiskEffect
	if err := r.GetDB().Preload("RiskRegister").First(&effect, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &effect, nil
}
func (r *RiskEffectRepository) FindByCode(code string) (*models.RiskEffect, error) {
	var effect models.RiskEffect
	if err := r.GetDB().Where("effect_code = ?", code).First(&effect).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &effect, nil
}
func (r *RiskEffectRepository) FindAll() ([]*models.RiskEffect, error) {
	var effects []*models.RiskEffect
	if err := r.GetDB().Preload("RiskRegister").Find(&effects).Error; err != nil { return nil, err }
	return effects, nil
}
