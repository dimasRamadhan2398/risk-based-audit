package routes

import (
	"audit-service/controllers"
	"audit-service/controllers/crud"
	"audit-service/models"
	"audit-service/pkg/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RouteHandler handles all route registration
type RouteHandler struct {
	engine         *gin.Engine
	registry       *RouteRegistry
	authMiddleware *middleware.AuthMiddleware
	db             *gorm.DB
}

// NewRouteHandler creates a new route handler
func NewRouteHandler(
	engine *gin.Engine,
	authMiddleware *middleware.AuthMiddleware,
	db *gorm.DB,
) *RouteHandler {
	return &RouteHandler{
		engine:         engine,
		registry:       NewRouteRegistry(),
		authMiddleware: authMiddleware,
		db:             db,
	}
}

// SetRegistry sets the route registry with controllers
func (h *RouteHandler) SetRegistry(registry *RouteRegistry) {
	h.registry = registry
}

// RegisterRoutes registers all API routes
func (h *RouteHandler) RegisterRoutes() {
	apiV1 := h.engine.Group("/api/v1")

	// Audit Charter routes
	auditCharters := apiV1.Group("/audit-charters")
	{
		auditCharters.GET("", h.registry.AuditCharter.ListCharters)
		auditCharters.POST("", h.registry.AuditCharter.CreateCharter)
		auditCharters.GET("/active", h.registry.AuditCharter.GetActiveCharter)
		auditCharters.GET("/version", h.registry.AuditCharter.GetCharterByVersion)
		auditCharters.GET("/:id", h.registry.AuditCharter.GetCharter)
		auditCharters.PUT("/:id", h.registry.AuditCharter.UpdateCharter)
		auditCharters.DELETE("/:id", h.registry.AuditCharter.DeleteCharter)
		auditCharters.POST("/:id/activate", h.registry.AuditCharter.SetActiveCharter)
	}

	// Audit Mandate routes
	auditMandates := apiV1.Group("/audit-mandates")
	{
		auditMandates.GET("", h.registry.AuditMandate.ListMandates)
		auditMandates.POST("", h.registry.AuditMandate.CreateMandate)
		auditMandates.GET("/active", h.registry.AuditMandate.GetActiveMandate)
		auditMandates.GET("/reference", h.registry.AuditMandate.GetMandateByReference)
		auditMandates.GET("/:id", h.registry.AuditMandate.GetMandate)
		auditMandates.PUT("/:id", h.registry.AuditMandate.UpdateMandate)
		auditMandates.DELETE("/:id", h.registry.AuditMandate.DeleteMandate)
	}

	// Audit Assignment routes
	auditAssignments := apiV1.Group("/audit-assignments")
	{
		auditAssignments.GET("", h.registry.AuditAssignment.ListAssignments)
		auditAssignments.POST("", h.registry.AuditAssignment.CreateAssignment)
		auditAssignments.GET("/auditor/:auditor_id", h.registry.AuditAssignment.GetAssignmentsByAuditor)
		auditAssignments.GET("/audit-plan/:audit_plan_id", h.registry.AuditAssignment.GetAssignmentsByAuditPlan)
		auditAssignments.GET("/:id", h.registry.AuditAssignment.GetAssignment)
		auditAssignments.PUT("/:id", h.registry.AuditAssignment.UpdateAssignment)
		auditAssignments.DELETE("/:id", h.registry.AuditAssignment.DeleteAssignment)
		auditAssignments.PUT("/:id/status", h.registry.AuditAssignment.UpdateStatus)
	}

	// Audit Activity routes
	auditActivities := apiV1.Group("/audit-activities")
	{
		auditActivities.GET("", h.registry.AuditActivity.ListActivities)
		auditActivities.POST("", h.registry.AuditActivity.CreateActivity)
		auditActivities.GET("/project", h.registry.AuditActivity.GetActivityByProjectCode)
		auditActivities.GET("/:id", h.registry.AuditActivity.GetActivity)
		auditActivities.PUT("/:id", h.registry.AuditActivity.UpdateActivity)
		auditActivities.DELETE("/:id", h.registry.AuditActivity.DeleteActivity)
	}

	// 1. Annual Audit Plan
	annualPlans := apiV1.Group("/annual-audit-plans")
	{
		annualPlans.GET("", crud.List(h.db, "AuditAnnual", func() interface{} { return &[]models.AuditAnnual{} }))
		annualPlans.GET("/:id", crud.GetByID(h.db, "AuditAnnual", func() interface{} { return &models.AuditAnnual{} }))
		annualPlans.POST("", crud.Create(h.db, "AuditAnnual", func() interface{} { return &models.AuditAnnual{} }))
		annualPlans.PUT("/:id", crud.Update(h.db, "AuditAnnual", func() interface{} { return &models.AuditAnnual{} }))
		annualPlans.DELETE("/:id", crud.Delete(h.db, "AuditAnnual", func() interface{} { return &models.AuditAnnual{} }))
	}

	// 2. Activity Plan
	activityPlans := apiV1.Group("/activity-plans")
	{
		activityPlans.GET("", crud.List(h.db, "ActivityPlan", func() interface{} { return &[]models.ActivityPlan{} }, "AnnualPlan"))
		activityPlans.GET("/:id", crud.GetByID(h.db, "ActivityPlan", func() interface{} { return &models.ActivityPlan{} }, "AnnualPlan"))
		activityPlans.POST("", crud.Create(h.db, "ActivityPlan", func() interface{} { return &models.ActivityPlan{} }))
		activityPlans.PUT("/:id", crud.Update(h.db, "ActivityPlan", func() interface{} { return &models.ActivityPlan{} }))
		activityPlans.DELETE("/:id", crud.Delete(h.db, "ActivityPlan", func() interface{} { return &models.ActivityPlan{} }))
	}

	// 3. Strategic Audit Plan
	strategicPlans := apiV1.Group("/strategic-plans")
	{
		strategicPlans.GET("", crud.List(h.db, "StrategicPlan", func() interface{} { return &[]models.StrategicPlan{} }))
		strategicPlans.GET("/:id", crud.GetByID(h.db, "StrategicPlan", func() interface{} { return &models.StrategicPlan{} }))
		strategicPlans.POST("", crud.Create(h.db, "StrategicPlan", func() interface{} { return &models.StrategicPlan{} }))
		strategicPlans.PUT("/:id", crud.Update(h.db, "StrategicPlan", func() interface{} { return &models.StrategicPlan{} }))
		strategicPlans.DELETE("/:id", crud.Delete(h.db, "StrategicPlan", func() interface{} { return &models.StrategicPlan{} }))
	}

	// 4. Assignment Letter
	assignmentLetters := apiV1.Group("/assignment-letters")
	{
		assignmentLetters.GET("", crud.List(h.db, "AssignmentLetter", func() interface{} { return &[]models.AssignmentLetter{} }))
		assignmentLetters.GET("/:id", crud.GetByID(h.db, "AssignmentLetter", func() interface{} { return &models.AssignmentLetter{} }))
		assignmentLetters.POST("", crud.Create(h.db, "AssignmentLetter", func() interface{} { return &models.AssignmentLetter{} }))
		assignmentLetters.PUT("/:id", crud.Update(h.db, "AssignmentLetter", func() interface{} { return &models.AssignmentLetter{} }))
		assignmentLetters.DELETE("/:id", crud.Delete(h.db, "AssignmentLetter", func() interface{} { return &models.AssignmentLetter{} }))
	}

	// 5. Audit Execution
	auditExecutions := apiV1.Group("/audit-executions")
	{
		auditExecutions.GET("", crud.List(h.db, "AuditExecution", func() interface{} { return &[]models.AuditExecution{} }))
		auditExecutions.GET("/:id", crud.GetByID(h.db, "AuditExecution", func() interface{} { return &models.AuditExecution{} }))
		auditExecutions.POST("", crud.Create(h.db, "AuditExecution", func() interface{} { return &models.AuditExecution{} }))
		auditExecutions.PUT("/:id", crud.Update(h.db, "AuditExecution", func() interface{} { return &models.AuditExecution{} }))
		auditExecutions.DELETE("/:id", crud.Delete(h.db, "AuditExecution", func() interface{} { return &models.AuditExecution{} }))
	}

	// 6. Fieldwork Interviews
	fieldworkInterviews := apiV1.Group("/fieldwork/interviews")
	{
		fieldworkInterviews.GET("", crud.List(h.db, "FieldworkInterview", func() interface{} { return &[]models.FieldworkInterview{} }))
		fieldworkInterviews.GET("/:id", crud.GetByID(h.db, "FieldworkInterview", func() interface{} { return &models.FieldworkInterview{} }))
		fieldworkInterviews.POST("", crud.Create(h.db, "FieldworkInterview", func() interface{} { return &models.FieldworkInterview{} }))
		fieldworkInterviews.PUT("/:id", crud.Update(h.db, "FieldworkInterview", func() interface{} { return &models.FieldworkInterview{} }))
		fieldworkInterviews.DELETE("/:id", crud.Delete(h.db, "FieldworkInterview", func() interface{} { return &models.FieldworkInterview{} }))
	}

	// 7. Fieldwork Observations
	fieldworkObservations := apiV1.Group("/fieldwork/observations")
	{
		fieldworkObservations.GET("", crud.List(h.db, "FieldworkObservation", func() interface{} { return &[]models.FieldworkObservation{} }))
		fieldworkObservations.GET("/:id", crud.GetByID(h.db, "FieldworkObservation", func() interface{} { return &models.FieldworkObservation{} }))
		fieldworkObservations.POST("", crud.Create(h.db, "FieldworkObservation", func() interface{} { return &models.FieldworkObservation{} }))
		fieldworkObservations.PUT("/:id", crud.Update(h.db, "FieldworkObservation", func() interface{} { return &models.FieldworkObservation{} }))
		fieldworkObservations.DELETE("/:id", crud.Delete(h.db, "FieldworkObservation", func() interface{} { return &models.FieldworkObservation{} }))
	}

	// 8. Fieldwork Documents
	fieldworkDocuments := apiV1.Group("/fieldwork/documents")
	{
		fieldworkDocuments.GET("", crud.List(h.db, "FieldworkDocument", func() interface{} { return &[]models.FieldworkDocument{} }))
		fieldworkDocuments.GET("/:id", crud.GetByID(h.db, "FieldworkDocument", func() interface{} { return &models.FieldworkDocument{} }))
		fieldworkDocuments.POST("", crud.Create(h.db, "FieldworkDocument", func() interface{} { return &models.FieldworkDocument{} }))
		fieldworkDocuments.PUT("/:id", crud.Update(h.db, "FieldworkDocument", func() interface{} { return &models.FieldworkDocument{} }))
		fieldworkDocuments.DELETE("/:id", crud.Delete(h.db, "FieldworkDocument", func() interface{} { return &models.FieldworkDocument{} }))
	}

	// 9. Fieldwork Samples
	fieldworkSamples := apiV1.Group("/fieldwork/samples")
	{
		fieldworkSamples.GET("", crud.List(h.db, "FieldworkSample", func() interface{} { return &[]models.FieldworkSample{} }))
		fieldworkSamples.GET("/:id", crud.GetByID(h.db, "FieldworkSample", func() interface{} { return &models.FieldworkSample{} }))
		fieldworkSamples.POST("", crud.Create(h.db, "FieldworkSample", func() interface{} { return &models.FieldworkSample{} }))
		fieldworkSamples.PUT("/:id", crud.Update(h.db, "FieldworkSample", func() interface{} { return &models.FieldworkSample{} }))
		fieldworkSamples.DELETE("/:id", crud.Delete(h.db, "FieldworkSample", func() interface{} { return &models.FieldworkSample{} }))
	}

	// 10. Fieldwork Test Controls
	fieldworkTestControls := apiV1.Group("/fieldwork/test-controls")
	{
		fieldworkTestControls.GET("", crud.List(h.db, "FieldworkTestControl", func() interface{} { return &[]models.FieldworkTestControl{} }))
		fieldworkTestControls.GET("/:id", crud.GetByID(h.db, "FieldworkTestControl", func() interface{} { return &models.FieldworkTestControl{} }))
		fieldworkTestControls.POST("", crud.Create(h.db, "FieldworkTestControl", func() interface{} { return &models.FieldworkTestControl{} }))
		fieldworkTestControls.PUT("/:id", crud.Update(h.db, "FieldworkTestControl", func() interface{} { return &models.FieldworkTestControl{} }))
		fieldworkTestControls.DELETE("/:id", crud.Delete(h.db, "FieldworkTestControl", func() interface{} { return &models.FieldworkTestControl{} }))
	}

	// 11. Working Paper Headers
	workingPaperHeaders := apiV1.Group("/working-papers/headers")
	{
		workingPaperHeaders.GET("", crud.List(h.db, "WorkingPaperHeader", func() interface{} { return &[]models.WorkingPaperHeader{} }))
		workingPaperHeaders.GET("/:id", crud.GetByID(h.db, "WorkingPaperHeader", func() interface{} { return &models.WorkingPaperHeader{} }))
		workingPaperHeaders.POST("", crud.Create(h.db, "WorkingPaperHeader", func() interface{} { return &models.WorkingPaperHeader{} }))
		workingPaperHeaders.PUT("/:id", crud.Update(h.db, "WorkingPaperHeader", func() interface{} { return &models.WorkingPaperHeader{} }))
		workingPaperHeaders.DELETE("/:id", crud.Delete(h.db, "WorkingPaperHeader", func() interface{} { return &models.WorkingPaperHeader{} }))
	}

	// 12. Working Paper Risks
	workingPaperRisks := apiV1.Group("/working-papers/risks")
	{
		workingPaperRisks.GET("", crud.List(h.db, "WorkingPaperRisk", func() interface{} { return &[]models.WorkingPaperRisk{} }))
		workingPaperRisks.GET("/:id", crud.GetByID(h.db, "WorkingPaperRisk", func() interface{} { return &models.WorkingPaperRisk{} }))
		workingPaperRisks.POST("", crud.Create(h.db, "WorkingPaperRisk", func() interface{} { return &models.WorkingPaperRisk{} }))
		workingPaperRisks.PUT("/:id", crud.Update(h.db, "WorkingPaperRisk", func() interface{} { return &models.WorkingPaperRisk{} }))
		workingPaperRisks.DELETE("/:id", crud.Delete(h.db, "WorkingPaperRisk", func() interface{} { return &models.WorkingPaperRisk{} }))
	}

	// 13. Working Paper Samples
	workingPaperSamples := apiV1.Group("/working-papers/samples")
	{
		workingPaperSamples.GET("", crud.List(h.db, "WorkingPaperSample", func() interface{} { return &[]models.WorkingPaperSample{} }))
		workingPaperSamples.GET("/:id", crud.GetByID(h.db, "WorkingPaperSample", func() interface{} { return &models.WorkingPaperSample{} }))
		workingPaperSamples.POST("", crud.Create(h.db, "WorkingPaperSample", func() interface{} { return &models.WorkingPaperSample{} }))
		workingPaperSamples.PUT("/:id", crud.Update(h.db, "WorkingPaperSample", func() interface{} { return &models.WorkingPaperSample{} }))
		workingPaperSamples.DELETE("/:id", crud.Delete(h.db, "WorkingPaperSample", func() interface{} { return &models.WorkingPaperSample{} }))
	}

	// 14. Working Paper Causes
	workingPaperCauses := apiV1.Group("/working-papers/causes")
	{
		workingPaperCauses.GET("", crud.List(h.db, "WorkingPaperCause", func() interface{} { return &[]models.WorkingPaperCause{} }))
		workingPaperCauses.GET("/:id", crud.GetByID(h.db, "WorkingPaperCause", func() interface{} { return &models.WorkingPaperCause{} }))
		workingPaperCauses.POST("", crud.Create(h.db, "WorkingPaperCause", func() interface{} { return &models.WorkingPaperCause{} }))
		workingPaperCauses.PUT("/:id", crud.Update(h.db, "WorkingPaperCause", func() interface{} { return &models.WorkingPaperCause{} }))
		workingPaperCauses.DELETE("/:id", crud.Delete(h.db, "WorkingPaperCause", func() interface{} { return &models.WorkingPaperCause{} }))
	}

	// 15. Working Paper Plans
	workingPaperPlans := apiV1.Group("/working-papers/plans")
	{
		workingPaperPlans.GET("", crud.List(h.db, "WorkingPaperPlan", func() interface{} { return &[]models.WorkingPaperPlan{} }))
		workingPaperPlans.GET("/:id", crud.GetByID(h.db, "WorkingPaperPlan", func() interface{} { return &models.WorkingPaperPlan{} }))
		workingPaperPlans.POST("", crud.Create(h.db, "WorkingPaperPlan", func() interface{} { return &models.WorkingPaperPlan{} }))
		workingPaperPlans.PUT("/:id", crud.Update(h.db, "WorkingPaperPlan", func() interface{} { return &models.WorkingPaperPlan{} }))
		workingPaperPlans.DELETE("/:id", crud.Delete(h.db, "WorkingPaperPlan", func() interface{} { return &models.WorkingPaperPlan{} }))
	}

	// 16. Working Papers (Original)
	workingPapers := apiV1.Group("/working-papers")
	{
		workingPapers.GET("", crud.List(h.db, "WorkingPaper", func() interface{} { return &[]models.WorkingPaper{} }))
		workingPapers.GET("/:id", crud.GetByID(h.db, "WorkingPaper", func() interface{} { return &models.WorkingPaper{} }))
		workingPapers.POST("", crud.Create(h.db, "WorkingPaper", func() interface{} { return &models.WorkingPaper{} }))
		workingPapers.PUT("/:id", crud.Update(h.db, "WorkingPaper", func() interface{} { return &models.WorkingPaper{} }))
		workingPapers.DELETE("/:id", crud.Delete(h.db, "WorkingPaper", func() interface{} { return &models.WorkingPaper{} }))
	}

	// 16b. Imported Working Papers
	importedWorkingPaperCtrl := controllers.NewImportedWorkingPaperController(h.db)
	importedWorkingPapers := apiV1.Group("/working-papers/imports")
	{
		importedWorkingPapers.GET("", importedWorkingPaperCtrl.List)
		importedWorkingPapers.POST("", importedWorkingPaperCtrl.Import)
		importedWorkingPapers.DELETE("/:id", importedWorkingPaperCtrl.Delete)
		importedWorkingPapers.GET("/:id/download", importedWorkingPaperCtrl.Download)
	}

	// 17. Audit Result Reports
	auditResultReports := apiV1.Group("/audit-result-reports")
	{
		auditResultReports.GET("", crud.List(h.db, "AuditResultReport", func() interface{} { return &[]models.AuditResultReport{} }))
		auditResultReports.GET("/:id", crud.GetByID(h.db, "AuditResultReport", func() interface{} { return &models.AuditResultReport{} }))
		auditResultReports.POST("", crud.Create(h.db, "AuditResultReport", func() interface{} { return &models.AuditResultReport{} }))
		auditResultReports.PUT("/:id", crud.Update(h.db, "AuditResultReport", func() interface{} { return &models.AuditResultReport{} }))
		auditResultReports.DELETE("/:id", crud.Delete(h.db, "AuditResultReport", func() interface{} { return &models.AuditResultReport{} }))
	}

	// 18. Action Taken Reports
	actionTakenReports := apiV1.Group("/action-taken-reports")
	{
		actionTakenReports.GET("", crud.List(h.db, "ActionTakenReport", func() interface{} { return &[]models.ActionTakenReport{} }))
		actionTakenReports.GET("/:id", crud.GetByID(h.db, "ActionTakenReport", func() interface{} { return &models.ActionTakenReport{} }))
		actionTakenReports.POST("", crud.Create(h.db, "ActionTakenReport", func() interface{} { return &models.ActionTakenReport{} }))
		actionTakenReports.PUT("/:id", crud.Update(h.db, "ActionTakenReport", func() interface{} { return &models.ActionTakenReport{} }))
		actionTakenReports.DELETE("/:id", crud.Delete(h.db, "ActionTakenReport", func() interface{} { return &models.ActionTakenReport{} }))
	}
}
