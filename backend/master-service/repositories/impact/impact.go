package impact
import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)
type IImpactRepository interface {
	Create(i *models.Impact) error
	Update(i *models.Impact) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.Impact, error)
	FindByCode(code string) (*models.Impact, error)
	FindAll() ([]*models.Impact, error)
}
type ImpactRepository struct { *repositories.BaseRepository }
func NewImpactRepository(db *gorm.DB) IImpactRepository { return &ImpactRepository{BaseRepository: repositories.NewBaseRepository(db)} }
func (r *ImpactRepository) Create(i *models.Impact) error { return r.BaseRepository.Create(i) }
func (r *ImpactRepository) Update(i *models.Impact) error { return r.BaseRepository.Update(i) }
func (r *ImpactRepository) Delete(id uuid.UUID) error { return r.BaseRepository.Delete(&models.Impact{ID: id}) }
func (r *ImpactRepository) FindByID(id uuid.UUID) (*models.Impact, error) {
	var i models.Impact
	if err := r.GetDB().First(&i, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &i, nil
}
func (r *ImpactRepository) FindByCode(code string) (*models.Impact, error) {
	var i models.Impact
	if err := r.GetDB().Where("code = ?", code).First(&i).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &i, nil
}
func (r *ImpactRepository) FindAll() ([]*models.Impact, error) {
	var is []*models.Impact
	if err := r.GetDB().Find(&is).Error; err != nil { return nil, err }
	return is, nil
}
