package mitigation_action
import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)
type IMitigationActionRepository interface {
	Create(action *models.MitigationAction) error
	Update(action *models.MitigationAction) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.MitigationAction, error)
	FindByCode(code string) (*models.MitigationAction, error)
	FindAll() ([]*models.MitigationAction, error)
}
type MitigationActionRepository struct { *repositories.BaseRepository }
func NewMitigationActionRepository(db *gorm.DB) IMitigationActionRepository { return &MitigationActionRepository{BaseRepository: repositories.NewBaseRepository(db)} }
func (r *MitigationActionRepository) Create(action *models.MitigationAction) error { return r.BaseRepository.Create(action) }
func (r *MitigationActionRepository) Update(action *models.MitigationAction) error { return r.BaseRepository.Update(action) }
func (r *MitigationActionRepository) Delete(id uuid.UUID) error { return r.BaseRepository.Delete(&models.MitigationAction{ID: id}) }
func (r *MitigationActionRepository) FindByID(id uuid.UUID) (*models.MitigationAction, error) {
	var action models.MitigationAction
	if err := r.GetDB().Preload("RiskRegister").Preload("Control").Preload("Issue").Preload("Owner").Preload("Department").Preload("VerifiedBy").First(&action, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &action, nil
}
func (r *MitigationActionRepository) FindByCode(code string) (*models.MitigationAction, error) {
	var action models.MitigationAction
	if err := r.GetDB().Where("action_code = ?", code).First(&action).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &action, nil
}
func (r *MitigationActionRepository) FindAll() ([]*models.MitigationAction, error) {
	var actions []*models.MitigationAction
	if err := r.GetDB().Preload("RiskRegister").Preload("Control").Preload("Issue").Preload("Owner").Preload("Department").Preload("VerifiedBy").Find(&actions).Error; err != nil { return nil, err }
	return actions, nil
}
