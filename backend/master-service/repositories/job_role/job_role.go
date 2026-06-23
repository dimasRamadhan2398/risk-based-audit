package job_role
import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)
type IJobRoleRepository interface {
	Create(role *models.JobRole) error
	Update(role *models.JobRole) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.JobRole, error)
	FindByCode(code string) (*models.JobRole, error)
	FindAll() ([]*models.JobRole, error)
}
type JobRoleRepository struct { *repositories.BaseRepository }
func NewJobRoleRepository(db *gorm.DB) IJobRoleRepository { return &JobRoleRepository{BaseRepository: repositories.NewBaseRepository(db)} }
func (r *JobRoleRepository) Create(role *models.JobRole) error { return r.BaseRepository.Create(role) }
func (r *JobRoleRepository) Update(role *models.JobRole) error { return r.BaseRepository.Update(role) }
func (r *JobRoleRepository) Delete(id uuid.UUID) error { return r.BaseRepository.Delete(&models.JobRole{ID: id}) }
func (r *JobRoleRepository) FindByID(id uuid.UUID) (*models.JobRole, error) {
	var role models.JobRole
	if err := r.GetDB().First(&role, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &role, nil
}
func (r *JobRoleRepository) FindByCode(code string) (*models.JobRole, error) {
	var role models.JobRole
	if err := r.GetDB().Where("job_role_code = ?", code).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &role, nil
}
func (r *JobRoleRepository) FindAll() ([]*models.JobRole, error) {
	var roles []*models.JobRole
	if err := r.GetDB().Find(&roles).Error; err != nil { return nil, err }
	return roles, nil
}
