package business_unit
import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)
type IBusinessUnitRepository interface {
	Create(bu *models.BusinessUnit) error
	Update(bu *models.BusinessUnit) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.BusinessUnit, error)
	FindByCode(code string) (*models.BusinessUnit, error)
	FindAll() ([]*models.BusinessUnit, error)
}
type BusinessUnitRepository struct { *repositories.BaseRepository }
func NewBusinessUnitRepository(db *gorm.DB) IBusinessUnitRepository { return &BusinessUnitRepository{BaseRepository: repositories.NewBaseRepository(db)} }
func (r *BusinessUnitRepository) Create(bu *models.BusinessUnit) error { return r.BaseRepository.Create(bu) }
func (r *BusinessUnitRepository) Update(bu *models.BusinessUnit) error { return r.BaseRepository.Update(bu) }
func (r *BusinessUnitRepository) Delete(id uuid.UUID) error { return r.BaseRepository.Delete(&models.BusinessUnit{ID: id}) }
func (r *BusinessUnitRepository) FindByID(id uuid.UUID) (*models.BusinessUnit, error) {
	var bu models.BusinessUnit
	if err := r.GetDB().First(&bu, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &bu, nil
}
func (r *BusinessUnitRepository) FindByCode(code string) (*models.BusinessUnit, error) {
	var bu models.BusinessUnit
	if err := r.GetDB().Where("business_unit_code = ?", code).First(&bu).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &bu, nil
}
func (r *BusinessUnitRepository) FindAll() ([]*models.BusinessUnit, error) {
	var bus []*models.BusinessUnit
	if err := r.GetDB().Find(&bus).Error; err != nil { return nil, err }
	return bus, nil
}
