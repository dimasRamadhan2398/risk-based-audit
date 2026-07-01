package audit_period
import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)
type IAuditPeriodRepository interface {
	Create(p *models.AuditPeriod) error
	Update(p *models.AuditPeriod) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.AuditPeriod, error)
	FindByCode(code string) (*models.AuditPeriod, error)
	FindAll() ([]*models.AuditPeriod, error)
}
type AuditPeriodRepository struct { *repositories.BaseRepository }
func NewAuditPeriodRepository(db *gorm.DB) IAuditPeriodRepository { return &AuditPeriodRepository{BaseRepository: repositories.NewBaseRepository(db)} }
func (r *AuditPeriodRepository) Create(p *models.AuditPeriod) error { return r.BaseRepository.Create(p) }
func (r *AuditPeriodRepository) Update(p *models.AuditPeriod) error { return r.BaseRepository.Update(p) }
func (r *AuditPeriodRepository) Delete(id uuid.UUID) error { return r.BaseRepository.Delete(&models.AuditPeriod{ID: id}) }
func (r *AuditPeriodRepository) FindByID(id uuid.UUID) (*models.AuditPeriod, error) {
	var p models.AuditPeriod
	if err := r.GetDB().First(&p, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &p, nil
}
func (r *AuditPeriodRepository) FindByCode(code string) (*models.AuditPeriod, error) {
	var p models.AuditPeriod
	if err := r.GetDB().Where("period_code = ?", code).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &p, nil
}
func (r *AuditPeriodRepository) FindAll() ([]*models.AuditPeriod, error) {
	var ps []*models.AuditPeriod
	if err := r.GetDB().Find(&ps).Error; err != nil { return nil, err }
	return ps, nil
}
