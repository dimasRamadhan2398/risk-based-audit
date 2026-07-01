package risk_register
import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)
type IRiskRegisterRepository interface {
	Create(risk *models.RiskRegister) error
	Update(risk *models.RiskRegister) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.RiskRegister, error)
	FindByCode(code string) (*models.RiskRegister, error)
	FindAll() ([]*models.RiskRegister, error)
}
type RiskRegisterRepository struct { *repositories.BaseRepository }
func NewRiskRegisterRepository(db *gorm.DB) IRiskRegisterRepository { return &RiskRegisterRepository{BaseRepository: repositories.NewBaseRepository(db)} }
func (r *RiskRegisterRepository) Create(risk *models.RiskRegister) error { return r.BaseRepository.Create(risk) }
func (r *RiskRegisterRepository) Update(risk *models.RiskRegister) error { return r.BaseRepository.Update(risk) }
func (r *RiskRegisterRepository) Delete(id uuid.UUID) error { return r.BaseRepository.Delete(&models.RiskRegister{ID: id}) }
func (r *RiskRegisterRepository) FindByID(id uuid.UUID) (*models.RiskRegister, error) {
	var risk models.RiskRegister
	if err := r.GetDB().Preload("Department").Preload("RiskCategory").Preload("RiskOwner").Preload("InherentLikelihood").Preload("InherentImpact").Preload("InherentRiskLevel").Preload("ResidualLikelihood").Preload("ResidualImpact").Preload("ResidualRiskLevel").First(&risk, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &risk, nil
}
func (r *RiskRegisterRepository) FindByCode(code string) (*models.RiskRegister, error) {
	var risk models.RiskRegister
	if err := r.GetDB().Where("code = ?", code).First(&risk).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &risk, nil
}
func (r *RiskRegisterRepository) FindAll() ([]*models.RiskRegister, error) {
	var risks []*models.RiskRegister
	if err := r.GetDB().Preload("Department").Preload("RiskCategory").Preload("RiskOwner").Preload("InherentLikelihood").Preload("InherentImpact").Preload("InherentRiskLevel").Preload("ResidualLikelihood").Preload("ResidualImpact").Preload("ResidualRiskLevel").Find(&risks).Error; err != nil { return nil, err }
	return risks, nil
}
