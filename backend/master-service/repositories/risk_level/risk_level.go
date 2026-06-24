package risk_level
import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)
type IRiskLevelRepository interface {
	Create(rl *models.RiskLevel) error
	Update(rl *models.RiskLevel) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.RiskLevel, error)
	FindByCode(code string) (*models.RiskLevel, error)
	FindAll() ([]*models.RiskLevel, error)
}
type RiskLevelRepository struct { *repositories.BaseRepository }
func NewRiskLevelRepository(db *gorm.DB) IRiskLevelRepository { return &RiskLevelRepository{BaseRepository: repositories.NewBaseRepository(db)} }
func (r *RiskLevelRepository) Create(rl *models.RiskLevel) error { return r.BaseRepository.Create(rl) }
func (r *RiskLevelRepository) Update(rl *models.RiskLevel) error { return r.BaseRepository.Update(rl) }
func (r *RiskLevelRepository) Delete(id uuid.UUID) error { return r.BaseRepository.Delete(&models.RiskLevel{ID: id}) }
func (r *RiskLevelRepository) FindByID(id uuid.UUID) (*models.RiskLevel, error) {
	var rl models.RiskLevel
	if err := r.GetDB().First(&rl, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &rl, nil
}
func (r *RiskLevelRepository) FindByCode(code string) (*models.RiskLevel, error) {
	var rl models.RiskLevel
	if err := r.GetDB().Where("risk_code = ?", code).First(&rl).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &rl, nil
}
func (r *RiskLevelRepository) FindAll() ([]*models.RiskLevel, error) {
	var rls []*models.RiskLevel
	if err := r.GetDB().Find(&rls).Error; err != nil { return nil, err }
	return rls, nil
}
