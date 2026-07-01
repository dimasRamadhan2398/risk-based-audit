package audit_recommendation
import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)
type IAuditRecommendationRepository interface {
	Create(rec *models.AuditRecommendation) error
	Update(rec *models.AuditRecommendation) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.AuditRecommendation, error)
	FindByCode(code string) (*models.AuditRecommendation, error)
	FindAll() ([]*models.AuditRecommendation, error)
}
type AuditRecommendationRepository struct { *repositories.BaseRepository }
func NewAuditRecommendationRepository(db *gorm.DB) IAuditRecommendationRepository { return &AuditRecommendationRepository{BaseRepository: repositories.NewBaseRepository(db)} }
func (r *AuditRecommendationRepository) Create(rec *models.AuditRecommendation) error { return r.BaseRepository.Create(rec) }
func (r *AuditRecommendationRepository) Update(rec *models.AuditRecommendation) error { return r.BaseRepository.Update(rec) }
func (r *AuditRecommendationRepository) Delete(id uuid.UUID) error { return r.BaseRepository.Delete(&models.AuditRecommendation{ID: id}) }
func (r *AuditRecommendationRepository) FindByID(id uuid.UUID) (*models.AuditRecommendation, error) {
	var rec models.AuditRecommendation
	if err := r.GetDB().Preload("Finding").Preload("Responsible").Preload("VerifiedBy").First(&rec, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &rec, nil
}
func (r *AuditRecommendationRepository) FindByCode(code string) (*models.AuditRecommendation, error) {
	var rec models.AuditRecommendation
	if err := r.GetDB().Where("recommendation_code = ?", code).First(&rec).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &rec, nil
}
func (r *AuditRecommendationRepository) FindAll() ([]*models.AuditRecommendation, error) {
	var recs []*models.AuditRecommendation
	if err := r.GetDB().Preload("Finding").Preload("Responsible").Preload("VerifiedBy").Find(&recs).Error; err != nil { return nil, err }
	return recs, nil
}
