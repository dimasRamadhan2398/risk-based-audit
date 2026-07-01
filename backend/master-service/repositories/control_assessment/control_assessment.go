package control_assessment
import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)
type IControlAssessmentRepository interface {
	Create(assessment *models.ControlAssessment) error
	Update(assessment *models.ControlAssessment) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.ControlAssessment, error)
	FindByCode(code string) (*models.ControlAssessment, error)
	FindAll() ([]*models.ControlAssessment, error)
}
type ControlAssessmentRepository struct { *repositories.BaseRepository }
func NewControlAssessmentRepository(db *gorm.DB) IControlAssessmentRepository { return &ControlAssessmentRepository{BaseRepository: repositories.NewBaseRepository(db)} }
func (r *ControlAssessmentRepository) Create(assessment *models.ControlAssessment) error { return r.BaseRepository.Create(assessment) }
func (r *ControlAssessmentRepository) Update(assessment *models.ControlAssessment) error { return r.BaseRepository.Update(assessment) }
func (r *ControlAssessmentRepository) Delete(id uuid.UUID) error { return r.BaseRepository.Delete(&models.ControlAssessment{ID: id}) }
func (r *ControlAssessmentRepository) FindByID(id uuid.UUID) (*models.ControlAssessment, error) {
	var assessment models.ControlAssessment
	if err := r.GetDB().Preload("Control").Preload("AuditPlan").Preload("AuditScope").Preload("Auditor").First(&assessment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &assessment, nil
}
func (r *ControlAssessmentRepository) FindByCode(code string) (*models.ControlAssessment, error) {
	var assessment models.ControlAssessment
	if err := r.GetDB().Where("assessment_code = ?", code).First(&assessment).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &assessment, nil
}
func (r *ControlAssessmentRepository) FindAll() ([]*models.ControlAssessment, error) {
	var assessments []*models.ControlAssessment
	if err := r.GetDB().Preload("Control").Preload("AuditPlan").Preload("AuditScope").Preload("Auditor").Find(&assessments).Error; err != nil { return nil, err }
	return assessments, nil
}
