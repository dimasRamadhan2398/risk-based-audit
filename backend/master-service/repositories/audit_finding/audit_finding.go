package audit_finding
import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)
type IAuditFindingRepository interface {
	Create(finding *models.AuditFinding) error
	Update(finding *models.AuditFinding) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.AuditFinding, error)
	FindByCode(code string) (*models.AuditFinding, error)
	FindAll() ([]*models.AuditFinding, error)
}
type AuditFindingRepository struct { *repositories.BaseRepository }
func NewAuditFindingRepository(db *gorm.DB) IAuditFindingRepository { return &AuditFindingRepository{BaseRepository: repositories.NewBaseRepository(db)} }
func (r *AuditFindingRepository) Create(finding *models.AuditFinding) error { return r.BaseRepository.Create(finding) }
func (r *AuditFindingRepository) Update(finding *models.AuditFinding) error { return r.BaseRepository.Update(finding) }
func (r *AuditFindingRepository) Delete(id uuid.UUID) error { return r.BaseRepository.Delete(&models.AuditFinding{ID: id}) }
func (r *AuditFindingRepository) FindByID(id uuid.UUID) (*models.AuditFinding, error) {
	var finding models.AuditFinding
	if err := r.GetDB().Preload("AuditPlan").Preload("AuditScope").Preload("ControlAssessment").Preload("RiskRegister").Preload("Department").Preload("Owner").Preload("Auditor").First(&finding, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &finding, nil
}
func (r *AuditFindingRepository) FindByCode(code string) (*models.AuditFinding, error) {
	var finding models.AuditFinding
	if err := r.GetDB().Where("finding_code = ?", code).First(&finding).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &finding, nil
}
func (r *AuditFindingRepository) FindAll() ([]*models.AuditFinding, error) {
	var findings []*models.AuditFinding
	if err := r.GetDB().Preload("AuditPlan").Preload("AuditScope").Preload("ControlAssessment").Preload("RiskRegister").Preload("Department").Preload("Owner").Preload("Auditor").Find(&findings).Error; err != nil { return nil, err }
	return findings, nil
}
