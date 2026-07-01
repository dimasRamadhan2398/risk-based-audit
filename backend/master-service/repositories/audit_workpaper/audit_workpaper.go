package audit_workpaper
import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)
type IAuditWorkpaperRepository interface {
	Create(wp *models.AuditWorkpaper) error
	Update(wp *models.AuditWorkpaper) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.AuditWorkpaper, error)
	FindByCode(code string) (*models.AuditWorkpaper, error)
	FindAll() ([]*models.AuditWorkpaper, error)
}
type AuditWorkpaperRepository struct { *repositories.BaseRepository }
func NewAuditWorkpaperRepository(db *gorm.DB) IAuditWorkpaperRepository { return &AuditWorkpaperRepository{BaseRepository: repositories.NewBaseRepository(db)} }
func (r *AuditWorkpaperRepository) Create(wp *models.AuditWorkpaper) error { return r.BaseRepository.Create(wp) }
func (r *AuditWorkpaperRepository) Update(wp *models.AuditWorkpaper) error { return r.BaseRepository.Update(wp) }
func (r *AuditWorkpaperRepository) Delete(id uuid.UUID) error { return r.BaseRepository.Delete(&models.AuditWorkpaper{ID: id}) }
func (r *AuditWorkpaperRepository) FindByID(id uuid.UUID) (*models.AuditWorkpaper, error) {
	var wp models.AuditWorkpaper
	if err := r.GetDB().Preload("AuditPlan").Preload("AuditScope").Preload("ControlAssessment").Preload("AuditFinding").Preload("Auditor").Preload("Reviewer").First(&wp, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &wp, nil
}
func (r *AuditWorkpaperRepository) FindByCode(code string) (*models.AuditWorkpaper, error) {
	var wp models.AuditWorkpaper
	if err := r.GetDB().Where("workpaper_code = ?", code).First(&wp).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &wp, nil
}
func (r *AuditWorkpaperRepository) FindAll() ([]*models.AuditWorkpaper, error) {
	var wps []*models.AuditWorkpaper
	if err := r.GetDB().Preload("AuditPlan").Preload("AuditScope").Preload("ControlAssessment").Preload("AuditFinding").Preload("Auditor").Preload("Reviewer").Find(&wps).Error; err != nil { return nil, err }
	return wps, nil
}
