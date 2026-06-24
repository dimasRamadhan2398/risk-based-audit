package risk_cause
import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)
type IRiskCauseRepository interface {
	Create(cause *models.RiskCause) error
	Update(cause *models.RiskCause) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.RiskCause, error)
	FindByCode(code string) (*models.RiskCause, error)
	FindAll() ([]*models.RiskCause, error)
}
type RiskCauseRepository struct { *repositories.BaseRepository }
func NewRiskCauseRepository(db *gorm.DB) IRiskCauseRepository { return &RiskCauseRepository{BaseRepository: repositories.NewBaseRepository(db)} }
func (r *RiskCauseRepository) Create(cause *models.RiskCause) error { return r.BaseRepository.Create(cause) }
func (r *RiskCauseRepository) Update(cause *models.RiskCause) error { return r.BaseRepository.Update(cause) }
func (r *RiskCauseRepository) Delete(id uuid.UUID) error { return r.BaseRepository.Delete(&models.RiskCause{ID: id}) }
func (r *RiskCauseRepository) FindByID(id uuid.UUID) (*models.RiskCause, error) {
	var cause models.RiskCause
	if err := r.GetDB().Preload("RiskRegister").First(&cause, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &cause, nil
}
func (r *RiskCauseRepository) FindByCode(code string) (*models.RiskCause, error) {
	var cause models.RiskCause
	if err := r.GetDB().Where("cause_code = ?", code).First(&cause).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &cause, nil
}
func (r *RiskCauseRepository) FindAll() ([]*models.RiskCause, error) {
	var causes []*models.RiskCause
	if err := r.GetDB().Preload("RiskRegister").Find(&causes).Error; err != nil { return nil, err }
	return causes, nil
}
