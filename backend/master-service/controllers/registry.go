package controllers

import (
	annualAuditPlanCtrl "master-service/controllers/annual_audit_plan"
	auditFindingCtrl "master-service/controllers/audit_finding"
	auditIssueCtrl "master-service/controllers/audit_issue"
	auditPeriodCtrl "master-service/controllers/audit_period"
	auditRecommendationCtrl "master-service/controllers/audit_recommendation"
	auditScopeCtrl "master-service/controllers/audit_scope"
	auditUniverseCtrl "master-service/controllers/audit_universe"
	auditWorkpaperCtrl "master-service/controllers/audit_workpaper"
	businessUnitCtrl "master-service/controllers/business_unit"
	companyCtrl "master-service/controllers/company"
	controlCtrl "master-service/controllers/control"
	controlAssessmentCtrl "master-service/controllers/control_assessment"
	departmentCtrl "master-service/controllers/department"
	employeeCtrl "master-service/controllers/employee"
	impactCtrl "master-service/controllers/impact"
	jobRoleCtrl "master-service/controllers/job_role"
	likelihoodCtrl "master-service/controllers/likelihood"
	locationCtrl "master-service/controllers/location"
	mitigationActionCtrl "master-service/controllers/mitigation_action"
	riskCategoryCtrl "master-service/controllers/risk_category"
	riskCauseCtrl "master-service/controllers/risk_cause"
	riskEffectCtrl "master-service/controllers/risk_effect"
	riskIndicatorCtrl "master-service/controllers/risk_indicator"
	riskLevelCtrl "master-service/controllers/risk_level"
	riskMatrixCellCtrl "master-service/controllers/risk_matrix_cell"
	riskRegisterCtrl "master-service/controllers/risk_register"
	vmgsCtrl "master-service/controllers/vision_mission_goals"
	"master-service/pkg/validations"
	"master-service/services"
)

type Registry struct {
	service   services.IServiceRegistry
	validator *validations.Validator
}

type IControllerRegistry interface {
	GetAnnualAuditPlan() annualAuditPlanCtrl.AnnualAuditPlanControllerInterface
	GetAuditFinding() auditFindingCtrl.AuditFindingControllerInterface
	GetAuditIssue() auditIssueCtrl.AuditIssueControllerInterface
	GetAuditPeriod() auditPeriodCtrl.AuditPeriodControllerInterface
	GetAuditRecommendation() auditRecommendationCtrl.AuditRecommendationControllerInterface
	GetAuditScope() auditScopeCtrl.AuditScopeControllerInterface
	GetAuditUniverse() auditUniverseCtrl.AuditUniverseControllerInterface
	GetAuditWorkpaper() auditWorkpaperCtrl.AuditWorkpaperControllerInterface
	GetBusinessUnit() businessUnitCtrl.BusinessUnitControllerInterface
	GetCompany() companyCtrl.CompanyControllerInterface
	GetControl() controlCtrl.ControlControllerInterface
	GetControlAssessment() controlAssessmentCtrl.ControlAssessmentControllerInterface
	GetDepartment() departmentCtrl.DepartmentControllerInterface
	GetEmployee() employeeCtrl.EmployeeControllerInterface
	GetImpact() impactCtrl.ImpactControllerInterface
	GetJobRole() jobRoleCtrl.JobRoleControllerInterface
	GetLikelihood() likelihoodCtrl.LikelihoodControllerInterface
	GetLocation() locationCtrl.LocationControllerInterface
	GetMitigationAction() mitigationActionCtrl.MitigationActionControllerInterface
	GetRiskCause() riskCauseCtrl.RiskCauseControllerInterface
	GetRiskCategory() riskCategoryCtrl.RiskCategoryControllerInterface
	GetRiskEffect() riskEffectCtrl.RiskEffectControllerInterface
	GetRiskIndicator() riskIndicatorCtrl.RiskIndicatorControllerInterface
	GetRiskLevel() riskLevelCtrl.RiskLevelControllerInterface
	GetRiskMatrixCell() riskMatrixCellCtrl.RiskMatrixCellControllerInterface
	GetRiskRegister() riskRegisterCtrl.RiskRegisterControllerInterface
	GetVisionMissionGoals() vmgsCtrl.VisionMissionGoalsControllerInterface
}

func NewControllerRegistry(service services.IServiceRegistry, validator *validations.Validator) IControllerRegistry {
	return &Registry{
		service:   service,
		validator: validator,
	}
}

func (r *Registry) GetAnnualAuditPlan() annualAuditPlanCtrl.AnnualAuditPlanControllerInterface {
	return annualAuditPlanCtrl.NewAnnualAuditPlanController(r.service.GetAnnualAuditPlan(), r.validator)
}
func (r *Registry) GetAuditFinding() auditFindingCtrl.AuditFindingControllerInterface {
	return auditFindingCtrl.NewAuditFindingController(r.service.GetAuditFinding(), r.validator)
}
func (r *Registry) GetAuditIssue() auditIssueCtrl.AuditIssueControllerInterface {
	return auditIssueCtrl.NewAuditIssueController(r.service.GetAuditIssue(), r.validator)
}
func (r *Registry) GetAuditPeriod() auditPeriodCtrl.AuditPeriodControllerInterface {
	return auditPeriodCtrl.NewAuditPeriodController(r.service.GetAuditPeriod(), r.validator)
}
func (r *Registry) GetAuditRecommendation() auditRecommendationCtrl.AuditRecommendationControllerInterface {
	return auditRecommendationCtrl.NewAuditRecommendationController(r.service.GetAuditRecommendation(), r.validator)
}
func (r *Registry) GetAuditScope() auditScopeCtrl.AuditScopeControllerInterface {
	return auditScopeCtrl.NewAuditScopeController(r.service.GetAuditScope(), r.validator)
}
func (r *Registry) GetAuditUniverse() auditUniverseCtrl.AuditUniverseControllerInterface {
	return auditUniverseCtrl.NewAuditUniverseController(r.service.GetAuditUniverse(), r.validator)
}
func (r *Registry) GetAuditWorkpaper() auditWorkpaperCtrl.AuditWorkpaperControllerInterface {
	return auditWorkpaperCtrl.NewAuditWorkpaperController(r.service.GetAuditWorkpaper(), r.validator)
}
func (r *Registry) GetBusinessUnit() businessUnitCtrl.BusinessUnitControllerInterface {
	return businessUnitCtrl.NewBusinessUnitController(r.service.GetBusinessUnit(), r.validator)
}
func (r *Registry) GetCompany() companyCtrl.CompanyControllerInterface {
	return companyCtrl.NewCompanyController(r.service.GetCompany(), r.validator)
}
func (r *Registry) GetControl() controlCtrl.ControlControllerInterface {
	return controlCtrl.NewControlController(r.service.GetControl(), r.validator)
}
func (r *Registry) GetControlAssessment() controlAssessmentCtrl.ControlAssessmentControllerInterface {
	return controlAssessmentCtrl.NewControlAssessmentController(r.service.GetControlAssessment(), r.validator)
}
func (r *Registry) GetDepartment() departmentCtrl.DepartmentControllerInterface {
	return departmentCtrl.NewDepartmentController(r.service.GetDepartment(), r.validator)
}
func (r *Registry) GetEmployee() employeeCtrl.EmployeeControllerInterface {
	return employeeCtrl.NewEmployeeController(r.service.GetEmployee(), r.validator)
}
func (r *Registry) GetImpact() impactCtrl.ImpactControllerInterface {
	return impactCtrl.NewImpactController(r.service.GetImpact(), r.validator)
}
func (r *Registry) GetJobRole() jobRoleCtrl.JobRoleControllerInterface {
	return jobRoleCtrl.NewJobRoleController(r.service.GetJobRole(), r.validator)
}
func (r *Registry) GetLikelihood() likelihoodCtrl.LikelihoodControllerInterface {
	return likelihoodCtrl.NewLikelihoodController(r.service.GetLikelihood(), r.validator)
}
func (r *Registry) GetLocation() locationCtrl.LocationControllerInterface {
	return locationCtrl.NewLocationController(r.service.GetLocation(), r.validator)
}
func (r *Registry) GetMitigationAction() mitigationActionCtrl.MitigationActionControllerInterface {
	return mitigationActionCtrl.NewMitigationActionController(r.service.GetMitigationAction(), r.validator)
}
func (r *Registry) GetRiskCause() riskCauseCtrl.RiskCauseControllerInterface {
	return riskCauseCtrl.NewRiskCauseController(r.service.GetRiskCause(), r.validator)
}
func (r *Registry) GetRiskCategory() riskCategoryCtrl.RiskCategoryControllerInterface {
	return riskCategoryCtrl.NewRiskCategoryController(r.service.GetRiskCategory(), r.validator)
}
func (r *Registry) GetRiskEffect() riskEffectCtrl.RiskEffectControllerInterface {
	return riskEffectCtrl.NewRiskEffectController(r.service.GetRiskEffect(), r.validator)
}
func (r *Registry) GetRiskIndicator() riskIndicatorCtrl.RiskIndicatorControllerInterface {
	return riskIndicatorCtrl.NewRiskIndicatorController(r.service.GetRiskIndicator(), r.validator)
}
func (r *Registry) GetRiskLevel() riskLevelCtrl.RiskLevelControllerInterface {
	return riskLevelCtrl.NewRiskLevelController(r.service.GetRiskLevel(), r.validator)
}
func (r *Registry) GetRiskMatrixCell() riskMatrixCellCtrl.RiskMatrixCellControllerInterface {
	return riskMatrixCellCtrl.NewRiskMatrixCellController(r.service.GetRiskMatrixCell(), r.validator)
}
func (r *Registry) GetRiskRegister() riskRegisterCtrl.RiskRegisterControllerInterface {
	return riskRegisterCtrl.NewRiskRegisterController(r.service.GetRiskRegister(), r.validator)
}

func (r *Registry) GetVisionMissionGoals() vmgsCtrl.VisionMissionGoalsControllerInterface {
	return vmgsCtrl.NewVisionMissionGoalsController(r.service.GetVisionMissionGoals(), r.validator)
}
