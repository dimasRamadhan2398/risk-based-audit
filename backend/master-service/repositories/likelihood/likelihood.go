package likelihood
import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)
type ILikelihoodRepository interface {
	Create(l *models.Likelihood) error
	Update(l *models.Likelihood) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.Likelihood, error)
	FindByCode(code string) (*models.Likelihood, error)
	FindAll() ([]*models.Likelihood, error)
}
type LikelihoodRepository struct { *repositories.BaseRepository }
func NewLikelihoodRepository(db *gorm.DB) ILikelihoodRepository { return &LikelihoodRepository{BaseRepository: repositories.NewBaseRepository(db)} }
func (r *LikelihoodRepository) Create(l *models.Likelihood) error { return r.BaseRepository.Create(l) }
func (r *LikelihoodRepository) Update(l *models.Likelihood) error { return r.BaseRepository.Update(l) }
func (r *LikelihoodRepository) Delete(id uuid.UUID) error { return r.BaseRepository.Delete(&models.Likelihood{ID: id}) }
func (r *LikelihoodRepository) FindByID(id uuid.UUID) (*models.Likelihood, error) {
	var l models.Likelihood
	if err := r.GetDB().First(&l, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &l, nil
}
func (r *LikelihoodRepository) FindByCode(code string) (*models.Likelihood, error) {
	var l models.Likelihood
	if err := r.GetDB().Where("code = ?", code).First(&l).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &l, nil
}
func (r *LikelihoodRepository) FindAll() ([]*models.Likelihood, error) {
	var ls []*models.Likelihood
	if err := r.GetDB().Find(&ls).Error; err != nil { return nil, err }
	return ls, nil
}
