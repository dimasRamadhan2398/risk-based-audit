package control
import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)
type IControlRepository interface {
	Create(control *models.Control) error
	Update(control *models.Control) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.Control, error)
	FindByCode(code string) (*models.Control, error)
	FindAll() ([]*models.Control, error)
}
type ControlRepository struct { *repositories.BaseRepository }
func NewControlRepository(db *gorm.DB) IControlRepository { return &ControlRepository{BaseRepository: repositories.NewBaseRepository(db)} }
func (r *ControlRepository) Create(control *models.Control) error { return r.BaseRepository.Create(control) }
func (r *ControlRepository) Update(control *models.Control) error { return r.BaseRepository.Update(control) }
func (r *ControlRepository) Delete(id uuid.UUID) error { return r.BaseRepository.Delete(&models.Control{ID: id}) }
func (r *ControlRepository) FindByID(id uuid.UUID) (*models.Control, error) {
	var control models.Control
	if err := r.GetDB().Preload("RiskRegister").Preload("Owner").Preload("Department").First(&control, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &control, nil
}
func (r *ControlRepository) FindByCode(code string) (*models.Control, error) {
	var control models.Control
	if err := r.GetDB().Where("control_code = ?", code).First(&control).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &control, nil
}
func (r *ControlRepository) FindAll() ([]*models.Control, error) {
	var controls []*models.Control
	if err := r.GetDB().Preload("RiskRegister").Preload("Owner").Preload("Department").Find(&controls).Error; err != nil { return nil, err }
	return controls, nil
}
