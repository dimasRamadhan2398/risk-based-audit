package audit_universe
import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)
type IAuditUniverseRepository interface {
	Create(u *models.AuditUniverse) error
	Update(u *models.AuditUniverse) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.AuditUniverse, error)
	FindByCode(code string) (*models.AuditUniverse, error)
	FindAll() ([]*models.AuditUniverse, error)
}
type AuditUniverseRepository struct { *repositories.BaseRepository }
func NewAuditUniverseRepository(db *gorm.DB) IAuditUniverseRepository { return &AuditUniverseRepository{BaseRepository: repositories.NewBaseRepository(db)} }
func (r *AuditUniverseRepository) Create(u *models.AuditUniverse) error { return r.BaseRepository.Create(u) }
func (r *AuditUniverseRepository) Update(u *models.AuditUniverse) error { return r.BaseRepository.Update(u) }
func (r *AuditUniverseRepository) Delete(id uuid.UUID) error { return r.BaseRepository.Delete(&models.AuditUniverse{ID: id}) }
func (r *AuditUniverseRepository) FindByID(id uuid.UUID) (*models.AuditUniverse, error) {
	var u models.AuditUniverse
	if err := r.GetDB().Preload("Parent").Preload("Department").First(&u, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &u, nil
}
func (r *AuditUniverseRepository) FindByCode(code string) (*models.AuditUniverse, error) {
	var u models.AuditUniverse
	if err := r.GetDB().Where("entity_code = ?", code).First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &u, nil
}
func (r *AuditUniverseRepository) FindAll() ([]*models.AuditUniverse, error) {
	var us []*models.AuditUniverse
	if err := r.GetDB().Preload("Parent").Preload("Department").Find(&us).Error; err != nil { return nil, err }
	return us, nil
}
