package annual_audit_plan
import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)
type IAnnualAuditPlanRepository interface {
	Create(plan *models.AnnualAuditPlan) error
	Update(plan *models.AnnualAuditPlan) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.AnnualAuditPlan, error)
	FindByCode(code string) (*models.AnnualAuditPlan, error)
	FindAll() ([]*models.AnnualAuditPlan, error)
}
type AnnualAuditPlanRepository struct { *repositories.BaseRepository }
func NewAnnualAuditPlanRepository(db *gorm.DB) IAnnualAuditPlanRepository { return &AnnualAuditPlanRepository{BaseRepository: repositories.NewBaseRepository(db)} }
func (r *AnnualAuditPlanRepository) Create(plan *models.AnnualAuditPlan) error { return r.BaseRepository.Create(plan) }
func (r *AnnualAuditPlanRepository) Update(plan *models.AnnualAuditPlan) error { return r.BaseRepository.Update(plan) }
func (r *AnnualAuditPlanRepository) Delete(id uuid.UUID) error { return r.BaseRepository.Delete(&models.AnnualAuditPlan{ID: id}) }
func (r *AnnualAuditPlanRepository) FindByID(id uuid.UUID) (*models.AnnualAuditPlan, error) {
	var plan models.AnnualAuditPlan
	if err := r.GetDB().Preload("AuditPeriod").Preload("RiskRegister").Preload("Department").Preload("ApprovedBy").Preload("RequestedBy").Preload("Activities.InvolvedDepartments").First(&plan, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &plan, nil
}
func (r *AnnualAuditPlanRepository) FindByCode(code string) (*models.AnnualAuditPlan, error) {
	var plan models.AnnualAuditPlan
	if err := r.GetDB().Preload("Activities.InvolvedDepartments").Where("plan_code = ?", code).First(&plan).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &plan, nil
}
func (r *AnnualAuditPlanRepository) FindAll() ([]*models.AnnualAuditPlan, error) {
	var plans []*models.AnnualAuditPlan
	if err := r.GetDB().Preload("AuditPeriod").Preload("RiskRegister").Preload("Department").Preload("ApprovedBy").Preload("RequestedBy").Preload("Activities.InvolvedDepartments").Find(&plans).Error; err != nil { return nil, err }
	return plans, nil
}

