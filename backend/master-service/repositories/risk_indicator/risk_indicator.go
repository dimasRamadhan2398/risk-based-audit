package risk_indicator
import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)
type IRiskIndicatorRepository interface {
	Create(indicator *models.RiskIndicator) error
	Update(indicator *models.RiskIndicator) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.RiskIndicator, error)
	FindByCode(code string) (*models.RiskIndicator, error)
	FindAll() ([]*models.RiskIndicator, error)
	CreateLog(log *models.RiskIndicatorLog) error
	FindLogsByIndicatorID(indicatorID uuid.UUID) ([]*models.RiskIndicatorLog, error)
}
type RiskIndicatorRepository struct { *repositories.BaseRepository }
func NewRiskIndicatorRepository(db *gorm.DB) IRiskIndicatorRepository { return &RiskIndicatorRepository{BaseRepository: repositories.NewBaseRepository(db)} }
func (r *RiskIndicatorRepository) Create(indicator *models.RiskIndicator) error { return r.BaseRepository.Create(indicator) }
func (r *RiskIndicatorRepository) Update(indicator *models.RiskIndicator) error { return r.BaseRepository.Update(indicator) }
func (r *RiskIndicatorRepository) Delete(id uuid.UUID) error { return r.BaseRepository.Delete(&models.RiskIndicator{ID: id}) }
func (r *RiskIndicatorRepository) FindByID(id uuid.UUID) (*models.RiskIndicator, error) {
	var indicator models.RiskIndicator
	if err := r.GetDB().Preload("RiskRegister").Preload("Owner").First(&indicator, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &indicator, nil
}
func (r *RiskIndicatorRepository) FindByCode(code string) (*models.RiskIndicator, error) {
	var indicator models.RiskIndicator
	if err := r.GetDB().Where("indicator_code = ?", code).First(&indicator).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &indicator, nil
}
func (r *RiskIndicatorRepository) FindAll() ([]*models.RiskIndicator, error) {
	var indicators []*models.RiskIndicator
	if err := r.GetDB().Preload("RiskRegister").Preload("Owner").Find(&indicators).Error; err != nil { return nil, err }
	return indicators, nil
}
func (r *RiskIndicatorRepository) CreateLog(log *models.RiskIndicatorLog) error {
	if err := r.GetDB().Create(log).Error; err != nil { return apperrors.ErrDatabase }
	return nil
}
func (r *RiskIndicatorRepository) FindLogsByIndicatorID(indicatorID uuid.UUID) ([]*models.RiskIndicatorLog, error) {
	var logs []*models.RiskIndicatorLog
	if err := r.GetDB().Where("indicator_id = ?", indicatorID).Order("recorded_at DESC").Find(&logs).Error; err != nil { return nil, err }
	return logs, nil
}
