package services

import (
	"gorm.io/gorm"

	annualAuditPlanRepo "master-service/repositories/annual_audit_plan"
	annualAuditPlanSvc "master-service/services/annual_audit_plan"
	auditFindingRepo "master-service/repositories/audit_finding"
	auditFindingSvc "master-service/services/audit_finding"
	auditIssueRepo "master-service/repositories/audit_issue"
	auditIssueSvc "master-service/services/audit_issue"
	auditPeriodRepo "master-service/repositories/audit_period"
	auditPeriodSvc "master-service/services/audit_period"
	auditRecommendationRepo "master-service/repositories/audit_recommendation"
	auditRecommendationSvc "master-service/services/audit_recommendation"
	auditScopeRepo "master-service/repositories/audit_scope"
	auditScopeSvc "master-service/services/audit_scope"
	auditUniverseRepo "master-service/repositories/audit_universe"
	auditUniverseSvc "master-service/services/audit_universe"
	auditWorkpaperRepo "master-service/repositories/audit_workpaper"
	auditWorkpaperSvc "master-service/services/audit_workpaper"
	businessUnitRepo "master-service/repositories/business_unit"
	businessUnitSvc "master-service/services/business_unit"
	companyRepo "master-service/repositories/company"
	companySvc "master-service/services/company"
	controlRepo "master-service/repositories/control"
	controlSvc "master-service/services/control"
	controlAssessmentRepo "master-service/repositories/control_assessment"
	controlAssessmentSvc "master-service/services/control_assessment"
	departmentRepo "master-service/repositories/department"
	departmentSvc "master-service/services/department"
	employeeRepo "master-service/repositories/employee"
	employeeSvc "master-service/services/employee"
	impactRepo "master-service/repositories/impact"
	impactSvc "master-service/services/impact"
	jobRoleRepo "master-service/repositories/job_role"
	jobRoleSvc "master-service/services/job_role"
	likelihoodRepo "master-service/repositories/likelihood"
	likelihoodSvc "master-service/services/likelihood"
	locationRepo "master-service/repositories/location"
	locationSvc "master-service/services/location"
	mitigationActionRepo "master-service/repositories/mitigation_action"
	mitigationActionSvc "master-service/services/mitigation_action"
	riskCauseRepo "master-service/repositories/risk_cause"
	riskCauseSvc "master-service/services/risk_cause"
	riskCategoryRepo "master-service/repositories/risk_category"
	riskCategorySvc "master-service/services/risk_category"
	riskEffectRepo "master-service/repositories/risk_effect"
	riskEffectSvc "master-service/services/risk_effect"
	riskIndicatorRepo "master-service/repositories/risk_indicator"
	riskIndicatorSvc "master-service/services/risk_indicator"
	riskLevelRepo "master-service/repositories/risk_level"
	riskLevelSvc "master-service/services/risk_level"
	riskMatrixCellRepo "master-service/repositories/risk_matrix_cell"
	riskMatrixCellSvc "master-service/services/risk_matrix_cell"
	riskRegisterRepo "master-service/repositories/risk_register"
	riskRegisterSvc "master-service/services/risk_register"
)

type IServiceRegistry interface {
	GetAnnualAuditPlan() annualAuditPlanSvc.AnnualAuditPlanServiceInterface
	GetAuditFinding() auditFindingSvc.AuditFindingServiceInterface
	GetAuditIssue() auditIssueSvc.AuditIssueServiceInterface
	GetAuditPeriod() auditPeriodSvc.AuditPeriodServiceInterface
	GetAuditRecommendation() auditRecommendationSvc.AuditRecommendationServiceInterface
	GetAuditScope() auditScopeSvc.AuditScopeServiceInterface
	GetAuditUniverse() auditUniverseSvc.AuditUniverseServiceInterface
	GetAuditWorkpaper() auditWorkpaperSvc.AuditWorkpaperServiceInterface
	GetBusinessUnit() businessUnitSvc.BusinessUnitServiceInterface
	GetCompany() companySvc.CompanyServiceInterface
	GetControl() controlSvc.ControlServiceInterface
	GetControlAssessment() controlAssessmentSvc.ControlAssessmentServiceInterface
	GetDepartment() departmentSvc.DepartmentServiceInterface
	GetEmployee() employeeSvc.EmployeeServiceInterface
	GetImpact() impactSvc.ImpactServiceInterface
	GetJobRole() jobRoleSvc.JobRoleServiceInterface
	GetLikelihood() likelihoodSvc.LikelihoodServiceInterface
	GetLocation() locationSvc.LocationServiceInterface
	GetMitigationAction() mitigationActionSvc.MitigationActionServiceInterface
	GetRiskCause() riskCauseSvc.RiskCauseServiceInterface
	GetRiskCategory() riskCategorySvc.RiskCategoryServiceInterface
	GetRiskEffect() riskEffectSvc.RiskEffectServiceInterface
	GetRiskIndicator() riskIndicatorSvc.RiskIndicatorServiceInterface
	GetRiskLevel() riskLevelSvc.RiskLevelServiceInterface
	GetRiskMatrixCell() riskMatrixCellSvc.RiskMatrixCellServiceInterface
	GetRiskRegister() riskRegisterSvc.RiskRegisterServiceInterface
}

type Registry struct {
	annualAuditPlan    annualAuditPlanSvc.AnnualAuditPlanServiceInterface
	auditFinding       auditFindingSvc.AuditFindingServiceInterface
	auditIssue         auditIssueSvc.AuditIssueServiceInterface
	auditPeriod        auditPeriodSvc.AuditPeriodServiceInterface
	auditRecommendation auditRecommendationSvc.AuditRecommendationServiceInterface
	auditScope         auditScopeSvc.AuditScopeServiceInterface
	auditUniverse      auditUniverseSvc.AuditUniverseServiceInterface
	auditWorkpaper     auditWorkpaperSvc.AuditWorkpaperServiceInterface
	businessUnit       businessUnitSvc.BusinessUnitServiceInterface
	company            companySvc.CompanyServiceInterface
	control            controlSvc.ControlServiceInterface
	controlAssessment  controlAssessmentSvc.ControlAssessmentServiceInterface
	department         departmentSvc.DepartmentServiceInterface
	employee           employeeSvc.EmployeeServiceInterface
	impact             impactSvc.ImpactServiceInterface
	jobRole            jobRoleSvc.JobRoleServiceInterface
	likelihood         likelihoodSvc.LikelihoodServiceInterface
	location           locationSvc.LocationServiceInterface
	mitigationAction   mitigationActionSvc.MitigationActionServiceInterface
	riskCause          riskCauseSvc.RiskCauseServiceInterface
	riskCategory       riskCategorySvc.RiskCategoryServiceInterface
	riskEffect         riskEffectSvc.RiskEffectServiceInterface
	riskIndicator      riskIndicatorSvc.RiskIndicatorServiceInterface
	riskLevel          riskLevelSvc.RiskLevelServiceInterface
	riskMatrixCell     riskMatrixCellSvc.RiskMatrixCellServiceInterface
	riskRegister       riskRegisterSvc.RiskRegisterServiceInterface
}

func NewServiceRegistry(db *gorm.DB) IServiceRegistry {
	return &Registry{
		annualAuditPlan:    annualAuditPlanSvc.NewAnnualAuditPlanService(annualAuditPlanRepo.NewAnnualAuditPlanRepository(db)),
		auditFinding:       auditFindingSvc.NewAuditFindingService(auditFindingRepo.NewAuditFindingRepository(db)),
		auditIssue:         auditIssueSvc.NewAuditIssueService(auditIssueRepo.NewAuditIssueRepository(db)),
		auditPeriod:        auditPeriodSvc.NewAuditPeriodService(auditPeriodRepo.NewAuditPeriodRepository(db)),
		auditRecommendation: auditRecommendationSvc.NewAuditRecommendationService(auditRecommendationRepo.NewAuditRecommendationRepository(db)),
		auditScope:         auditScopeSvc.NewAuditScopeService(auditScopeRepo.NewAuditScopeRepository(db)),
		auditUniverse:      auditUniverseSvc.NewAuditUniverseService(auditUniverseRepo.NewAuditUniverseRepository(db)),
		auditWorkpaper:     auditWorkpaperSvc.NewAuditWorkpaperService(auditWorkpaperRepo.NewAuditWorkpaperRepository(db)),
		businessUnit:       businessUnitSvc.NewBusinessUnitService(businessUnitRepo.NewBusinessUnitRepository(db)),
		company:            companySvc.NewCompanyService(companyRepo.NewCompanyRepository(db)),
		control:            controlSvc.NewControlService(controlRepo.NewControlRepository(db)),
		controlAssessment:  controlAssessmentSvc.NewControlAssessmentService(controlAssessmentRepo.NewControlAssessmentRepository(db)),
		department:         departmentSvc.NewDepartmentService(departmentRepo.NewDepartmentRepository(db)),
		employee:           employeeSvc.NewEmployeeService(employeeRepo.NewEmployeeRepository(db)),
		impact:             impactSvc.NewImpactService(impactRepo.NewImpactRepository(db)),
		jobRole:            jobRoleSvc.NewJobRoleService(jobRoleRepo.NewJobRoleRepository(db)),
		likelihood:         likelihoodSvc.NewLikelihoodService(likelihoodRepo.NewLikelihoodRepository(db)),
		location:           locationSvc.NewLocationService(locationRepo.NewLocationRepository(db)),
		mitigationAction:   mitigationActionSvc.NewMitigationActionService(mitigationActionRepo.NewMitigationActionRepository(db)),
		riskCause:          riskCauseSvc.NewRiskCauseService(riskCauseRepo.NewRiskCauseRepository(db)),
		riskCategory:       riskCategorySvc.NewRiskCategoryService(riskCategoryRepo.NewRiskCategoryRepository(db)),
		riskEffect:         riskEffectSvc.NewRiskEffectService(riskEffectRepo.NewRiskEffectRepository(db)),
		riskIndicator:      riskIndicatorSvc.NewRiskIndicatorService(riskIndicatorRepo.NewRiskIndicatorRepository(db)),
		riskLevel:          riskLevelSvc.NewRiskLevelService(riskLevelRepo.NewRiskLevelRepository(db)),
		riskMatrixCell:     riskMatrixCellSvc.NewRiskMatrixCellService(riskMatrixCellRepo.NewRiskMatrixCellRepository(db)),
		riskRegister:       riskRegisterSvc.NewRiskRegisterService(riskRegisterRepo.NewRiskRegisterRepository(db)),
	}
}

func (r *Registry) GetAnnualAuditPlan() annualAuditPlanSvc.AnnualAuditPlanServiceInterface { return r.annualAuditPlan }
func (r *Registry) GetAuditFinding() auditFindingSvc.AuditFindingServiceInterface { return r.auditFinding }
func (r *Registry) GetAuditIssue() auditIssueSvc.AuditIssueServiceInterface { return r.auditIssue }
func (r *Registry) GetAuditPeriod() auditPeriodSvc.AuditPeriodServiceInterface { return r.auditPeriod }
func (r *Registry) GetAuditRecommendation() auditRecommendationSvc.AuditRecommendationServiceInterface { return r.auditRecommendation }
func (r *Registry) GetAuditScope() auditScopeSvc.AuditScopeServiceInterface { return r.auditScope }
func (r *Registry) GetAuditUniverse() auditUniverseSvc.AuditUniverseServiceInterface { return r.auditUniverse }
func (r *Registry) GetAuditWorkpaper() auditWorkpaperSvc.AuditWorkpaperServiceInterface { return r.auditWorkpaper }
func (r *Registry) GetBusinessUnit() businessUnitSvc.BusinessUnitServiceInterface { return r.businessUnit }
func (r *Registry) GetCompany() companySvc.CompanyServiceInterface { return r.company }
func (r *Registry) GetControl() controlSvc.ControlServiceInterface { return r.control }
func (r *Registry) GetControlAssessment() controlAssessmentSvc.ControlAssessmentServiceInterface { return r.controlAssessment }
func (r *Registry) GetDepartment() departmentSvc.DepartmentServiceInterface { return r.department }
func (r *Registry) GetEmployee() employeeSvc.EmployeeServiceInterface { return r.employee }
func (r *Registry) GetImpact() impactSvc.ImpactServiceInterface { return r.impact }
func (r *Registry) GetJobRole() jobRoleSvc.JobRoleServiceInterface { return r.jobRole }
func (r *Registry) GetLikelihood() likelihoodSvc.LikelihoodServiceInterface { return r.likelihood }
func (r *Registry) GetLocation() locationSvc.LocationServiceInterface { return r.location }
func (r *Registry) GetMitigationAction() mitigationActionSvc.MitigationActionServiceInterface { return r.mitigationAction }
func (r *Registry) GetRiskCause() riskCauseSvc.RiskCauseServiceInterface { return r.riskCause }
func (r *Registry) GetRiskCategory() riskCategorySvc.RiskCategoryServiceInterface { return r.riskCategory }
func (r *Registry) GetRiskEffect() riskEffectSvc.RiskEffectServiceInterface { return r.riskEffect }
func (r *Registry) GetRiskIndicator() riskIndicatorSvc.RiskIndicatorServiceInterface { return r.riskIndicator }
func (r *Registry) GetRiskLevel() riskLevelSvc.RiskLevelServiceInterface { return r.riskLevel }
func (r *Registry) GetRiskMatrixCell() riskMatrixCellSvc.RiskMatrixCellServiceInterface { return r.riskMatrixCell }
func (r *Registry) GetRiskRegister() riskRegisterSvc.RiskRegisterServiceInterface { return r.riskRegister }
