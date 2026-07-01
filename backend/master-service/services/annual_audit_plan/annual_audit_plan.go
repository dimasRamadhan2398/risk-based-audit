package annual_audit_plan
import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/annual_audit_plan"
	"github.com/google/uuid"
)
type AnnualAuditPlanServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.AnnualAuditPlan, error)
	FindById(ctx *base.BaseService, id string) (*models.AnnualAuditPlan, error)
	Create(ctx *base.BaseService, plan *models.AnnualAuditPlan) (*models.AnnualAuditPlan, error)
	Update(ctx *base.BaseService, id string, plan *models.AnnualAuditPlan) (*models.AnnualAuditPlan, error)
	Delete(ctx *base.BaseService, id string) error
}
type AnnualAuditPlanService struct { planRepo repo.IAnnualAuditPlanRepository }
func NewAnnualAuditPlanService(planRepo repo.IAnnualAuditPlanRepository) AnnualAuditPlanServiceInterface { return &AnnualAuditPlanService{planRepo: planRepo} }
func (s *AnnualAuditPlanService) Create(ctx *base.BaseService, plan *models.AnnualAuditPlan) (*models.AnnualAuditPlan, error) {
	if _, err := s.planRepo.FindByCode(plan.PlanCode); err == nil {
		return nil, apperrors.Wrap("AUDIT_PLAN_CODE_ALREADY_EXISTS", "Audit plan code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate audit plan code", 500, err)
	}
	if err := s.planRepo.Create(plan); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create audit plan", 500, err) }
	return s.planRepo.FindByID(plan.ID)
}
func (s *AnnualAuditPlanService) Delete(ctx *base.BaseService, id string) error {
	planID, err := uuid.Parse(id)
	if err != nil { return apperrors.Wrap("INVALID_AUDIT_PLAN_ID", "Invalid audit plan ID format", 400, err) }
	if _, err := s.planRepo.FindByID(planID); err != nil {
		if err == apperrors.ErrNotFound { return err }
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find audit plan", 500, err)
	}
	if err := s.planRepo.Delete(planID); err != nil { return apperrors.Wrap("DATABASE_ERROR", "Failed to delete audit plan", 500, err) }
	return nil
}
func (s *AnnualAuditPlanService) FindAll(ctx *base.BaseService) (*[]models.AnnualAuditPlan, error) {
	plans, err := s.planRepo.FindAll()
	if err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch audit plans", 500, err) }
	result := make([]models.AnnualAuditPlan, 0, len(plans))
	for _, plan := range plans { if plan != nil { result = append(result, *plan) } }
	return &result, nil
}
func (s *AnnualAuditPlanService) FindById(ctx *base.BaseService, id string) (*models.AnnualAuditPlan, error) {
	planID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_AUDIT_PLAN_ID", "Invalid audit plan ID format", 400, err) }
	plan, err := s.planRepo.FindByID(planID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch audit plan", 500, err)
	}
	return plan, nil
}
func (s *AnnualAuditPlanService) Update(ctx *base.BaseService, id string, plan *models.AnnualAuditPlan) (*models.AnnualAuditPlan, error) {
	planID, err := uuid.Parse(id)
	if err != nil { return nil, apperrors.Wrap("INVALID_AUDIT_PLAN_ID", "Invalid audit plan ID format", 400, err) }
	existingPlan, err := s.planRepo.FindByID(planID)
	if err != nil {
		if err == apperrors.ErrNotFound { return nil, err }
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch audit plan", 500, err)
	}
	if existingPlan.PlanCode != plan.PlanCode {
		if _, err := s.planRepo.FindByCode(plan.PlanCode); err == nil {
			return nil, apperrors.Wrap("AUDIT_PLAN_CODE_ALREADY_EXISTS", "Audit plan code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate audit plan code", 500, err)
		}
	}
	existingPlan.PlanCode = plan.PlanCode
	existingPlan.PlanTitle = plan.PlanTitle
	existingPlan.Description = plan.Description
	existingPlan.AuditPeriodID = plan.AuditPeriodID
	existingPlan.RiskRegisterID = plan.RiskRegisterID
	existingPlan.DepartmentID = plan.DepartmentID
	existingPlan.Priority = plan.Priority
	existingPlan.Status = plan.Status
	existingPlan.EstimatedDays = plan.EstimatedDays
	existingPlan.PlannedStartDate = plan.PlannedStartDate
	existingPlan.PlannedEndDate = plan.PlannedEndDate
	existingPlan.ApprovedByID = plan.ApprovedByID
	existingPlan.ApprovedAt = plan.ApprovedAt
	existingPlan.ApprovalNotes = plan.ApprovalNotes
	existingPlan.RequestedByID = plan.RequestedByID
	existingPlan.RevisionNumber = plan.RevisionNumber
	existingPlan.ParentPlanID = plan.ParentPlanID
	if err := s.planRepo.Update(existingPlan); err != nil { return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update audit plan", 500, err) }
	return s.planRepo.FindByID(planID)
}
