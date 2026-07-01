package audit_issue
import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)
type IAuditIssueRepository interface {
	Create(issue *models.AuditIssue) error
	Update(issue *models.AuditIssue) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.AuditIssue, error)
	FindByCode(code string) (*models.AuditIssue, error)
	FindAll() ([]*models.AuditIssue, error)
}
type AuditIssueRepository struct { *repositories.BaseRepository }
func NewAuditIssueRepository(db *gorm.DB) IAuditIssueRepository { return &AuditIssueRepository{BaseRepository: repositories.NewBaseRepository(db)} }
func (r *AuditIssueRepository) Create(issue *models.AuditIssue) error { return r.BaseRepository.Create(issue) }
func (r *AuditIssueRepository) Update(issue *models.AuditIssue) error { return r.BaseRepository.Update(issue) }
func (r *AuditIssueRepository) Delete(id uuid.UUID) error { return r.BaseRepository.Delete(&models.AuditIssue{ID: id}) }
func (r *AuditIssueRepository) FindByID(id uuid.UUID) (*models.AuditIssue, error) {
	var issue models.AuditIssue
	if err := r.GetDB().Preload("Finding").Preload("AuditPlan").Preload("RiskRegister").Preload("Control").Preload("Department").Preload("Owner").Preload("Auditor").First(&issue, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &issue, nil
}
func (r *AuditIssueRepository) FindByCode(code string) (*models.AuditIssue, error) {
	var issue models.AuditIssue
	if err := r.GetDB().Where("issue_code = ?", code).First(&issue).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &issue, nil
}
func (r *AuditIssueRepository) FindAll() ([]*models.AuditIssue, error) {
	var issues []*models.AuditIssue
	if err := r.GetDB().Preload("Finding").Preload("AuditPlan").Preload("RiskRegister").Preload("Control").Preload("Department").Preload("Owner").Preload("Auditor").Find(&issues).Error; err != nil { return nil, err }
	return issues, nil
}
