package routes

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"audit-service/controllers"
	ctrlATR "audit-service/controllers/action_taken_report"
	ctrlActivityPlan "audit-service/controllers/activity_plan"
	ctrlAnnualPlan "audit-service/controllers/annual_plan"
	ctrlAuditExecution "audit-service/controllers/audit_execution"
	ctrlAssignmentLetter "audit-service/controllers/assignment_letter"
	"audit-service/controllers/crud"
	ctrlExecutiveSummary "audit-service/controllers/executive_summary"
	ctrlFieldwork "audit-service/controllers/fieldwork"
	ctrlFindings "audit-service/controllers/findings"
	ctrlStrategicPlan "audit-service/controllers/strategic_plan"
	ctrlWorkingPaper "audit-service/controllers/working_paper"
	"audit-service/models"
	"audit-service/pkg/docxbuilder"
	"audit-service/pkg/masterclient"
	"audit-service/pkg/middleware"
	"audit-service/pkg/redis"
	atrsvc "audit-service/services/action_taken_report"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RouteHandler handles all route registration
type RouteHandler struct {
	engine         *gin.Engine
	registry       *RouteRegistry
	authMiddleware *middleware.AuthMiddleware
	db             *gorm.DB
	redisClient    *redis.Client // may be nil
}

// NewRouteHandler creates a new route handler
func NewRouteHandler(
	engine *gin.Engine,
	authMiddleware *middleware.AuthMiddleware,
	db *gorm.DB,
	redisClient *redis.Client,
) *RouteHandler {
	return &RouteHandler{
		engine:         engine,
		registry:       NewRouteRegistry(),
		authMiddleware: authMiddleware,
		db:             db,
		redisClient:    redisClient,
	}
}

// SetRegistry sets the route registry with controllers
func (h *RouteHandler) SetRegistry(registry *RouteRegistry) {
	h.registry = registry
}

// RegisterRoutes registers all API routes
func (h *RouteHandler) RegisterRoutes() {
	// Debug route to list all registered routes. Off unless ENABLE_DEBUG_ROUTES=true.
	registerDebugRoutes(h.engine)

	// Every API route requires a valid JWT. Kong does not check tokens for
	// this service, so without this the whole API was reachable anonymously.
	apiV1 := h.engine.Group("/api/v1")
	apiV1.Use(h.authMiddleware.Authenticate())

	// Audit Charter routes
	auditCharters := apiV1.Group("/audit-charters")
	{
		auditCharters.GET("", h.registry.AuditCharter.ListCharters)
		auditCharters.POST("", h.registry.AuditCharter.CreateCharter)
		auditCharters.GET("/active", h.registry.AuditCharter.GetActiveCharter)
		auditCharters.GET("/version", h.registry.AuditCharter.GetCharterByVersion)
		// IMPORTANT: specific routes must come BEFORE /:id to avoid conflicts
		auditCharters.GET("/:id/download", h.registry.AuditCharter.DownloadCharter)
		auditCharters.POST("/:id/activate", h.registry.AuditCharter.SetActiveCharter)
		auditCharters.GET("/:id", h.registry.AuditCharter.GetCharter)
		auditCharters.PUT("/:id", h.registry.AuditCharter.UpdateCharter)
		auditCharters.DELETE("/:id", h.registry.AuditCharter.DeleteCharter)
	}

	// Audit Guidelines routes
	auditGuidelines := apiV1.Group("/audit-guidelines")
	{
		auditGuidelines.GET("", crud.List(h.db, "AuditGuideline", func() interface{} { return &[]models.AuditGuideline{} }))
		auditGuidelines.GET("/:id", crud.GetByID(h.db, "AuditGuideline", func() interface{} { return &models.AuditGuideline{} }))
		auditGuidelines.POST("", crud.Create(h.db, "AuditGuideline", func() interface{} { return &models.AuditGuideline{} }))
		auditGuidelines.PUT("/:id", crud.Update(h.db, "AuditGuideline", func() interface{} { return &models.AuditGuideline{} }))
		auditGuidelines.DELETE("/:id", h.deleteAuditGuideline)
	}

	// Audit SOPs routes
	auditSops := apiV1.Group("/audit-sops")
	{
		auditSops.GET("", crud.List(h.db, "AuditSop", func() interface{} { return &[]models.AuditSop{} }, "Guideline"))
		auditSops.GET("/:id", crud.GetByID(h.db, "AuditSop", func() interface{} { return &models.AuditSop{} }, "Guideline"))
		auditSops.POST("", crud.Create(h.db, "AuditSop", func() interface{} { return &models.AuditSop{} }))
		auditSops.PUT("/:id", crud.Update(h.db, "AuditSop", func() interface{} { return &models.AuditSop{} }))
		auditSops.DELETE("/:id", crud.Delete(h.db, "AuditSop", func() interface{} { return &models.AuditSop{} }))
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
		// Reads add isUsed and activities[].activityPlanRef (the Audit Activity
		// Plan holding each row), see controllers/annual_plan/annual_plan_usage.go.
		annualPlans.GET("", crud.ListWithItems(h.db, "AuditAnnual", func() interface{} { return &[]models.AuditAnnual{} }, ctrlAnnualPlan.DecorateList(h.db)))
		annualPlans.GET("/:id", ctrlAnnualPlan.GetByID(h.db))
		// The server assigns the plan code (PKAT-{year}-{seq}) and checks status
		// transitions, see controllers/annual_plan.
		annualPlans.POST("", crud.CreateWithHook(h.db, "AuditAnnual", func() interface{} { return &models.AuditAnnual{} }, ctrlAnnualPlan.PrepareOnCreate))
		// Rows held by an Audit Activity Plan cannot be removed or retyped; the
		// after hook refreshes the plan rows' PKAT snapshots in the same transaction.
		annualPlans.PUT("/:id", crud.UpdateWithHooks(h.db, "AuditAnnual", func() interface{} { return &models.AuditAnnual{} }, ctrlAnnualPlan.PrepareOnUpdate, ctrlAnnualPlan.RefreshSnapshotsAfterUpdate))
		// 409 ANNUAL_PLAN_IN_USE when an activity plan or audit activity uses it.
		annualPlans.DELETE("/:id", ctrlAnnualPlan.Delete(h.db))
	}

	// 2. Activity Plan: the Formulir Rencana Kegiatan Audit (header + one
	// audit_activities row per assignment), see services/activity_plan_form.
	// /annual-activity-options is registered before /:id.
	ctrlActivityPlan.NewFormController(h.db).Register(apiV1.Group("/activity-plans"))

	// 3. Strategic Audit Plan
	strategicPlans := apiV1.Group("/strategic-plans")
	{
		strategicPlans.GET("", crud.List(h.db, "StrategicPlan", func() interface{} { return &[]models.StrategicPlan{} }))
		strategicPlans.GET("/:id", crud.GetByID(h.db, "StrategicPlan", func() interface{} { return &models.StrategicPlan{} }))
		// category must be empty or one of models.StrategicPlanCategories (400 otherwise)
		// the server assigns the code (SO-IA{n}); yearEnd >= yearStart; targets/realizations are numbers (400 otherwise)
		strategicPlans.POST("", ctrlStrategicPlan.ValidateCategory(), crud.CreateWithHook(h.db, "StrategicPlan", func() interface{} { return &models.StrategicPlan{} }, ctrlStrategicPlan.PrepareOnCreate))
		strategicPlans.PUT("/:id", ctrlStrategicPlan.ValidateCategory(), crud.UpdateWithHook(h.db, "StrategicPlan", func() interface{} { return &models.StrategicPlan{} }, ctrlStrategicPlan.PrepareOnUpdate))
		strategicPlans.DELETE("/:id", crud.Delete(h.db, "StrategicPlan", func() interface{} { return &models.StrategicPlan{} }))
	}

	// 4. Assignment Letter
	assignmentLetters := apiV1.Group("/assignment-letters")
	{
		assignmentLetters.GET("", crud.List(h.db, "AssignmentLetter", func() interface{} { return &[]models.AssignmentLetter{} }))
		assignmentLetters.GET("/:id", crud.GetByID(h.db, "AssignmentLetter", func() interface{} { return &models.AssignmentLetter{} }))
		// Only the roles allowed by the RBAC matrix ("Create / Publish Letter")
		// may publish or otherwise change a letter's status
		letterStatusGuard := middleware.RequireRolesForStatusChange(h.db, "assignment_letters", "Draft",
			"ADMIN", "AUDIT_MANAGER", "CHIEF_AUDIT_EXECUTIVE")
		assignmentLetters.POST("", letterStatusGuard, crud.CreateWithHooks(h.db, "AssignmentLetter", func() interface{} { return &models.AssignmentLetter{} }, nil, ctrlAuditExecution.SyncAfterCreate(ctrlAuditExecution.LetterKeysOf)))
		// auditPurpose echoed back from purposeList is not frozen into the column, see controllers/assignment_letter
		assignmentLetters.PUT("/:id", letterStatusGuard, ctrlAuditExecution.SyncOnWrite(h.db, func() interface{} { return &models.AssignmentLetter{} }, ctrlAuditExecution.LetterKeysOf, crud.UpdateWithHook(h.db, "AssignmentLetter", func() interface{} { return &models.AssignmentLetter{} }, ctrlAssignmentLetter.NormalizePurposeOnUpdate)))
		assignmentLetters.DELETE("/:id", ctrlAuditExecution.SyncOnWrite(h.db, func() interface{} { return &models.AssignmentLetter{} }, ctrlAuditExecution.LetterKeysOf, crud.Delete(h.db, "AssignmentLetter", func() interface{} { return &models.AssignmentLetter{} })))
	}

	// 5. Audit Execution
	auditExecutions := apiV1.Group("/audit-executions")
	{
		auditExecutions.GET("", ctrlAuditExecution.ListWithSources(h.db))
		auditExecutions.POST("/sync", ctrlAuditExecution.NewSyncController(h.db).SyncFromSources)
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
		fieldworkInterviews.POST("", crud.CreateWithHooks(h.db, "FieldworkInterview", func() interface{} { return &models.FieldworkInterview{} }, nil, ctrlAuditExecution.SyncAfterCreate(ctrlAuditExecution.LetterKeysOf)))
		fieldworkInterviews.PUT("/:id", ctrlAuditExecution.SyncOnWrite(h.db, func() interface{} { return &models.FieldworkInterview{} }, ctrlAuditExecution.LetterKeysOf, crud.Update(h.db, "FieldworkInterview", func() interface{} { return &models.FieldworkInterview{} })))
		fieldworkInterviews.DELETE("/:id", ctrlAuditExecution.SyncOnWrite(h.db, func() interface{} { return &models.FieldworkInterview{} }, ctrlAuditExecution.LetterKeysOf, crud.Delete(h.db, "FieldworkInterview", func() interface{} { return &models.FieldworkInterview{} })))
	}

	// 7. Fieldwork Observations
	fieldworkObservations := apiV1.Group("/fieldwork/observations")
	{
		fieldworkObservations.GET("", crud.List(h.db, "FieldworkObservation", func() interface{} { return &[]models.FieldworkObservation{} }))
		fieldworkObservations.GET("/:id", crud.GetByID(h.db, "FieldworkObservation", func() interface{} { return &models.FieldworkObservation{} }))
		fieldworkObservations.POST("", crud.CreateWithHooks(h.db, "FieldworkObservation", func() interface{} { return &models.FieldworkObservation{} }, nil, ctrlAuditExecution.SyncAfterCreate(ctrlAuditExecution.LetterKeysOf)))
		fieldworkObservations.PUT("/:id", ctrlAuditExecution.SyncOnWrite(h.db, func() interface{} { return &models.FieldworkObservation{} }, ctrlAuditExecution.LetterKeysOf, crud.Update(h.db, "FieldworkObservation", func() interface{} { return &models.FieldworkObservation{} })))
		fieldworkObservations.DELETE("/:id", ctrlAuditExecution.SyncOnWrite(h.db, func() interface{} { return &models.FieldworkObservation{} }, ctrlAuditExecution.LetterKeysOf, crud.Delete(h.db, "FieldworkObservation", func() interface{} { return &models.FieldworkObservation{} })))
	}

	// 8. Fieldwork Documents
	fieldworkDocuments := apiV1.Group("/fieldwork/documents")
	{
		fieldworkDocuments.GET("", crud.List(h.db, "FieldworkDocument", func() interface{} { return &[]models.FieldworkDocument{} }))
		fieldworkDocuments.GET("/:id", crud.GetByID(h.db, "FieldworkDocument", func() interface{} { return &models.FieldworkDocument{} }))
		fieldworkDocuments.POST("", crud.CreateWithHooks(h.db, "FieldworkDocument", func() interface{} { return &models.FieldworkDocument{} }, nil, ctrlAuditExecution.SyncAfterCreate(ctrlAuditExecution.LetterKeysOf)))
		fieldworkDocuments.PUT("/:id", ctrlAuditExecution.SyncOnWrite(h.db, func() interface{} { return &models.FieldworkDocument{} }, ctrlAuditExecution.LetterKeysOf, crud.Update(h.db, "FieldworkDocument", func() interface{} { return &models.FieldworkDocument{} })))
		fieldworkDocuments.DELETE("/:id", ctrlAuditExecution.SyncOnWrite(h.db, func() interface{} { return &models.FieldworkDocument{} }, ctrlAuditExecution.LetterKeysOf, crud.Delete(h.db, "FieldworkDocument", func() interface{} { return &models.FieldworkDocument{} })))
	}

	// 9. Fieldwork Samples
	fieldworkSamples := apiV1.Group("/fieldwork/samples")
	{
		fieldworkSamples.GET("", crud.List(h.db, "FieldworkSample", func() interface{} { return &[]models.FieldworkSample{} }))
		fieldworkSamples.GET("/:id", crud.GetByID(h.db, "FieldworkSample", func() interface{} { return &models.FieldworkSample{} }))
		fieldworkSamples.POST("", crud.CreateWithHooks(h.db, "FieldworkSample", func() interface{} { return &models.FieldworkSample{} }, nil, ctrlAuditExecution.SyncAfterCreate(ctrlAuditExecution.LetterKeysOf)))
		fieldworkSamples.PUT("/:id", ctrlAuditExecution.SyncOnWrite(h.db, func() interface{} { return &models.FieldworkSample{} }, ctrlAuditExecution.LetterKeysOf, crud.Update(h.db, "FieldworkSample", func() interface{} { return &models.FieldworkSample{} })))
		fieldworkSamples.DELETE("/:id", ctrlAuditExecution.SyncOnWrite(h.db, func() interface{} { return &models.FieldworkSample{} }, ctrlAuditExecution.LetterKeysOf, crud.Delete(h.db, "FieldworkSample", func() interface{} { return &models.FieldworkSample{} })))
	}

	// 10. Fieldwork Test Controls
	fieldworkTestControls := apiV1.Group("/fieldwork/test-controls")
	{
		fieldworkTestControls.GET("", crud.List(h.db, "FieldworkTestControl", func() interface{} { return &[]models.FieldworkTestControl{} }))
		// Recommendation count per LHA (Executive Summary compilation), one query set for all ids
		fieldworkTestControls.GET("/recommendations/summary", ctrlFieldwork.RecommendationSummaryHandler(h.db))
		fieldworkTestControls.GET("/:id", crud.GetByID(h.db, "FieldworkTestControl", func() interface{} { return &models.FieldworkTestControl{} }))
		fieldworkTestControls.POST("", crud.CreateWithHooks(h.db, "FieldworkTestControl", func() interface{} { return &models.FieldworkTestControl{} }, nil, ctrlAuditExecution.SyncAfterCreate(ctrlAuditExecution.LetterKeysOf)))
		fieldworkTestControls.PUT("/:id", ctrlAuditExecution.SyncOnWrite(h.db, func() interface{} { return &models.FieldworkTestControl{} }, ctrlAuditExecution.LetterKeysOf, crud.Update(h.db, "FieldworkTestControl", func() interface{} { return &models.FieldworkTestControl{} })))
		fieldworkTestControls.DELETE("/:id", ctrlAuditExecution.SyncOnWrite(h.db, func() interface{} { return &models.FieldworkTestControl{} }, ctrlAuditExecution.LetterKeysOf, crud.Delete(h.db, "FieldworkTestControl", func() interface{} { return &models.FieldworkTestControl{} })))
	}

	// 11. Working Paper Headers
	workingPaperHeaders := apiV1.Group("/working-papers/headers")
	{
		workingPaperHeaders.GET("", crud.List(h.db, "WorkingPaperHeader", func() interface{} { return &[]models.WorkingPaperHeader{} }))
		workingPaperHeaders.GET("/:id", crud.GetByID(h.db, "WorkingPaperHeader", func() interface{} { return &models.WorkingPaperHeader{} }))
		workingPaperHeaders.POST("", crud.CreateWithHooks(h.db, "WorkingPaperHeader", func() interface{} { return &models.WorkingPaperHeader{} }, nil, ctrlAuditExecution.SyncAfterCreate(ctrlAuditExecution.LetterKeysOf)))
		// auditPurpose follows the linked assignment letter, see controllers/working_paper
		workingPaperHeaders.PUT("/:id", ctrlAuditExecution.SyncOnWrite(h.db, func() interface{} { return &models.WorkingPaperHeader{} }, ctrlAuditExecution.LetterKeysOf, crud.UpdateWithHook(h.db, "WorkingPaperHeader", func() interface{} { return &models.WorkingPaperHeader{} }, ctrlWorkingPaper.SyncHeaderOnUpdate)))
		workingPaperHeaders.DELETE("/:id", ctrlAuditExecution.SyncOnWrite(h.db, func() interface{} { return &models.WorkingPaperHeader{} }, ctrlAuditExecution.LetterKeysOf, crud.Delete(h.db, "WorkingPaperHeader", func() interface{} { return &models.WorkingPaperHeader{} })))
	}

	// 12. Working Paper Risks
	workingPaperRisks := apiV1.Group("/working-papers/risks")
	{
		workingPaperRisks.GET("", crud.List(h.db, "WorkingPaperRisk", func() interface{} { return &[]models.WorkingPaperRisk{} }))
		workingPaperRisks.GET("/:id", crud.GetByID(h.db, "WorkingPaperRisk", func() interface{} { return &models.WorkingPaperRisk{} }))
		workingPaperRisks.POST("", crud.CreateWithHooks(h.db, "WorkingPaperRisk", func() interface{} { return &models.WorkingPaperRisk{} }, nil, ctrlAuditExecution.SyncAfterCreate(ctrlAuditExecution.LetterKeysOf)))
		workingPaperRisks.PUT("/:id", ctrlAuditExecution.SyncOnWrite(h.db, func() interface{} { return &models.WorkingPaperRisk{} }, ctrlAuditExecution.LetterKeysOf, crud.Update(h.db, "WorkingPaperRisk", func() interface{} { return &models.WorkingPaperRisk{} })))
		workingPaperRisks.DELETE("/:id", ctrlAuditExecution.SyncOnWrite(h.db, func() interface{} { return &models.WorkingPaperRisk{} }, ctrlAuditExecution.LetterKeysOf, crud.Delete(h.db, "WorkingPaperRisk", func() interface{} { return &models.WorkingPaperRisk{} })))
	}

	// 13. Working Paper Samples
	workingPaperSamples := apiV1.Group("/working-papers/samples")
	{
		workingPaperSamples.GET("", crud.List(h.db, "WorkingPaperSample", func() interface{} { return &[]models.WorkingPaperSample{} }))
		workingPaperSamples.GET("/:id", crud.GetByID(h.db, "WorkingPaperSample", func() interface{} { return &models.WorkingPaperSample{} }))
		workingPaperSamples.POST("", crud.CreateWithHooks(h.db, "WorkingPaperSample", func() interface{} { return &models.WorkingPaperSample{} }, nil, ctrlAuditExecution.SyncAfterCreate(ctrlAuditExecution.LetterKeysOf)))
		workingPaperSamples.PUT("/:id", ctrlAuditExecution.SyncOnWrite(h.db, func() interface{} { return &models.WorkingPaperSample{} }, ctrlAuditExecution.LetterKeysOf, crud.Update(h.db, "WorkingPaperSample", func() interface{} { return &models.WorkingPaperSample{} })))
		workingPaperSamples.DELETE("/:id", ctrlAuditExecution.SyncOnWrite(h.db, func() interface{} { return &models.WorkingPaperSample{} }, ctrlAuditExecution.LetterKeysOf, crud.Delete(h.db, "WorkingPaperSample", func() interface{} { return &models.WorkingPaperSample{} })))
	}

	// 14. Working Paper Causes
	workingPaperCauses := apiV1.Group("/working-papers/causes")
	{
		workingPaperCauses.GET("", crud.List(h.db, "WorkingPaperCause", func() interface{} { return &[]models.WorkingPaperCause{} }))
		workingPaperCauses.GET("/:id", crud.GetByID(h.db, "WorkingPaperCause", func() interface{} { return &models.WorkingPaperCause{} }))
		workingPaperCauses.POST("", crud.CreateWithHooks(h.db, "WorkingPaperCause", func() interface{} { return &models.WorkingPaperCause{} }, nil, ctrlAuditExecution.SyncAfterCreate(ctrlAuditExecution.LetterKeysOf)))
		workingPaperCauses.PUT("/:id", ctrlAuditExecution.SyncOnWrite(h.db, func() interface{} { return &models.WorkingPaperCause{} }, ctrlAuditExecution.LetterKeysOf, crud.Update(h.db, "WorkingPaperCause", func() interface{} { return &models.WorkingPaperCause{} })))
		workingPaperCauses.DELETE("/:id", ctrlAuditExecution.SyncOnWrite(h.db, func() interface{} { return &models.WorkingPaperCause{} }, ctrlAuditExecution.LetterKeysOf, crud.Delete(h.db, "WorkingPaperCause", func() interface{} { return &models.WorkingPaperCause{} })))
	}

	// 15. Working Paper Plans
	workingPaperPlans := apiV1.Group("/working-papers/plans")
	{
		workingPaperPlans.GET("", crud.List(h.db, "WorkingPaperPlan", func() interface{} { return &[]models.WorkingPaperPlan{} }))
		workingPaperPlans.GET("/:id", crud.GetByID(h.db, "WorkingPaperPlan", func() interface{} { return &models.WorkingPaperPlan{} }))
		workingPaperPlans.POST("", crud.CreateWithHooks(h.db, "WorkingPaperPlan", func() interface{} { return &models.WorkingPaperPlan{} }, nil, ctrlAuditExecution.SyncAfterCreate(ctrlAuditExecution.LetterKeysOf)))
		workingPaperPlans.PUT("/:id", ctrlAuditExecution.SyncOnWrite(h.db, func() interface{} { return &models.WorkingPaperPlan{} }, ctrlAuditExecution.LetterKeysOf, crud.Update(h.db, "WorkingPaperPlan", func() interface{} { return &models.WorkingPaperPlan{} })))
		workingPaperPlans.DELETE("/:id", ctrlAuditExecution.SyncOnWrite(h.db, func() interface{} { return &models.WorkingPaperPlan{} }, ctrlAuditExecution.LetterKeysOf, crud.Delete(h.db, "WorkingPaperPlan", func() interface{} { return &models.WorkingPaperPlan{} })))
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
		importedWorkingPapers.GET("/:id/view", importedWorkingPaperCtrl.View)
		importedWorkingPapers.GET("/:id/view/:filename", importedWorkingPaperCtrl.View)
	}

	// 16c. Uploaded Plan Documents (Audit Activity Plan)
	uploadedPlanDocumentCtrl := controllers.NewUploadedPlanDocumentController(h.db)
	uploadedPlanDocuments := apiV1.Group("/uploaded-plan-documents")
	{
		uploadedPlanDocuments.GET("", uploadedPlanDocumentCtrl.List)
		uploadedPlanDocuments.POST("", uploadedPlanDocumentCtrl.Upload)
		uploadedPlanDocuments.DELETE("/:id", uploadedPlanDocumentCtrl.Delete)
		uploadedPlanDocuments.GET("/:id/download", uploadedPlanDocumentCtrl.Download)
		uploadedPlanDocuments.GET("/:id/view", uploadedPlanDocumentCtrl.View)
		uploadedPlanDocuments.GET("/:id/view/:filename", uploadedPlanDocumentCtrl.View)
	}

	// 16c-bis. Uploaded Strategic Plan Documents (Strategic Audit Plan)
	uploadedStrategicPlanDocumentCtrl := controllers.NewUploadedStrategicPlanDocumentController(h.db)
	uploadedStrategicPlanDocuments := apiV1.Group("/uploaded-strategic-plan-documents")
	{
		uploadedStrategicPlanDocuments.GET("", uploadedStrategicPlanDocumentCtrl.List)
		uploadedStrategicPlanDocuments.POST("", uploadedStrategicPlanDocumentCtrl.Upload)
		uploadedStrategicPlanDocuments.DELETE("/:id", uploadedStrategicPlanDocumentCtrl.Delete)
		uploadedStrategicPlanDocuments.GET("/:id/download", uploadedStrategicPlanDocumentCtrl.Download)
		uploadedStrategicPlanDocuments.GET("/:id/view", uploadedStrategicPlanDocumentCtrl.View)
		uploadedStrategicPlanDocuments.GET("/:id/view/:filename", uploadedStrategicPlanDocumentCtrl.View)
	}

	// 16d. Uploaded Annual Plans
	uploadedAnnualPlanCtrl := controllers.NewUploadedAnnualPlanController(h.db)
	uploadedAnnualPlans := apiV1.Group("/uploaded-annual-plans")
	{
		uploadedAnnualPlans.GET("", uploadedAnnualPlanCtrl.List)
		uploadedAnnualPlans.POST("", uploadedAnnualPlanCtrl.Upload)
		uploadedAnnualPlans.DELETE("/:id", uploadedAnnualPlanCtrl.Delete)
		uploadedAnnualPlans.GET("/:id/download", uploadedAnnualPlanCtrl.Download)
		uploadedAnnualPlans.GET("/:id/view", uploadedAnnualPlanCtrl.View)
		uploadedAnnualPlans.GET("/:id/view/:filename", uploadedAnnualPlanCtrl.View)
	}

	// 16e. Uploaded Assignment Letters
	uploadedAssignmentLetterCtrl := controllers.NewUploadedAssignmentLetterController(h.db)
	uploadedAssignmentLetters := apiV1.Group("/uploaded-assignment-letters")
	{
		uploadedAssignmentLetters.GET("", uploadedAssignmentLetterCtrl.List)
		uploadedAssignmentLetters.POST("", uploadedAssignmentLetterCtrl.Upload)
		uploadedAssignmentLetters.DELETE("/:id", uploadedAssignmentLetterCtrl.Delete)
		uploadedAssignmentLetters.GET("/:id/download", uploadedAssignmentLetterCtrl.Download)
		uploadedAssignmentLetters.GET("/:id/view", uploadedAssignmentLetterCtrl.View)
		uploadedAssignmentLetters.GET("/:id/view/:filename", uploadedAssignmentLetterCtrl.View)
	}

	// 16f. Uploaded Audit Result Reports
	uploadedAuditResultReportCtrl := controllers.NewUploadedAuditResultReportController(h.db)
	uploadedAuditResultReports := apiV1.Group("/uploaded-audit-result-reports")
	{
		uploadedAuditResultReports.GET("", uploadedAuditResultReportCtrl.List)
		uploadedAuditResultReports.POST("", uploadedAuditResultReportCtrl.Upload)
		uploadedAuditResultReports.DELETE("/:id", uploadedAuditResultReportCtrl.Delete)
		uploadedAuditResultReports.GET("/:id/download", uploadedAuditResultReportCtrl.Download)
		uploadedAuditResultReports.GET("/:id/view", uploadedAuditResultReportCtrl.View)
		uploadedAuditResultReports.GET("/:id/view/:filename", uploadedAuditResultReportCtrl.View)
	}

	// 16g. Uploaded Executive Summaries
	uploadedExecutiveSummaryCtrl := controllers.NewUploadedExecutiveSummaryController(h.db)
	uploadedExecutiveSummaries := apiV1.Group("/uploaded-executive-summaries")
	{
		uploadedExecutiveSummaries.GET("", uploadedExecutiveSummaryCtrl.List)
		uploadedExecutiveSummaries.POST("", uploadedExecutiveSummaryCtrl.Upload)
		uploadedExecutiveSummaries.DELETE("/:id", uploadedExecutiveSummaryCtrl.Delete)
		uploadedExecutiveSummaries.GET("/:id/download", uploadedExecutiveSummaryCtrl.Download)
		uploadedExecutiveSummaries.GET("/:id/view", uploadedExecutiveSummaryCtrl.View)
		uploadedExecutiveSummaries.GET("/:id/view/:filename", uploadedExecutiveSummaryCtrl.View)
	}

	// 16h. Uploaded Executive Summary Reports
	uploadedExecutiveSummaryReportCtrl := controllers.NewUploadedExecutiveSummaryReportController(h.db)
	uploadedExecutiveSummaryReports := apiV1.Group("/uploaded-executive-summary-reports")
	{
		uploadedExecutiveSummaryReports.GET("", uploadedExecutiveSummaryReportCtrl.List)
		uploadedExecutiveSummaryReports.POST("", uploadedExecutiveSummaryReportCtrl.Upload)
		uploadedExecutiveSummaryReports.DELETE("/:id", uploadedExecutiveSummaryReportCtrl.Delete)
		uploadedExecutiveSummaryReports.GET("/:id/download", uploadedExecutiveSummaryReportCtrl.Download)
		uploadedExecutiveSummaryReports.GET("/:id/view", uploadedExecutiveSummaryReportCtrl.View)
		uploadedExecutiveSummaryReports.GET("/:id/view/:filename", uploadedExecutiveSummaryReportCtrl.View)
	}

	// 16i. Uploaded Consulting Documents
	uploadedConsultingDocumentCtrl := controllers.NewUploadedConsultingDocumentController(h.db)
	uploadedConsultingDocuments := apiV1.Group("/uploaded-consulting-documents")
	{
		uploadedConsultingDocuments.GET("", uploadedConsultingDocumentCtrl.List)
		uploadedConsultingDocuments.POST("", uploadedConsultingDocumentCtrl.Upload)
		uploadedConsultingDocuments.DELETE("/:id", uploadedConsultingDocumentCtrl.Delete)
		uploadedConsultingDocuments.GET("/:id/download", uploadedConsultingDocumentCtrl.Download)
		uploadedConsultingDocuments.GET("/:id/view", uploadedConsultingDocumentCtrl.View)
		uploadedConsultingDocuments.GET("/:id/view/:filename", uploadedConsultingDocumentCtrl.View)
	}

	// 16j. Uploaded Performance Reports (Q1, Q2, Q3, Q4, Tahunan)
	uploadedPerformanceReportCtrl := controllers.NewUploadedPerformanceReportController(h.db)
	uploadedPerformanceReports := apiV1.Group("/uploaded-performance-reports")
	{
		uploadedPerformanceReports.GET("", uploadedPerformanceReportCtrl.List)
		uploadedPerformanceReports.POST("", uploadedPerformanceReportCtrl.Upload)
		uploadedPerformanceReports.DELETE("/:id", uploadedPerformanceReportCtrl.Delete)
		uploadedPerformanceReports.GET("/:id/download", uploadedPerformanceReportCtrl.Download)
		uploadedPerformanceReports.GET("/:id/view", uploadedPerformanceReportCtrl.View)
		uploadedPerformanceReports.GET("/:id/view/:filename", uploadedPerformanceReportCtrl.View)
	}

	// 17. Audit Result Reports
	auditResultReports := apiV1.Group("/audit-result-reports")
	{
		auditResultReports.GET("", crud.List(h.db, "AuditResultReport", func() interface{} { return &[]models.AuditResultReport{} }))
		auditResultReports.GET("/auto-findings", h.getAutoFindings)
		auditResultReports.GET("/recent-findings", ctrlFindings.Recent(h.db))
		auditResultReports.GET("/:id/download-docx", h.downloadAuditResultReportDocx)
		auditResultReports.GET("/:id", crud.GetByID(h.db, "AuditResultReport", func() interface{} { return &models.AuditResultReport{} }))
		// Saving a report keeps its action taken reports in step in the same
		// transaction: approving it ("Final"/"Approved"/"Published") creates one
		// ATR per finding (see services/action_taken_report.SyncFromReport)
		auditResultReports.POST("", crud.CreateWithHooks(h.db, "AuditResultReport", func() interface{} { return &models.AuditResultReport{} }, nil, syncReportAndExecutionsAfterCreate))
		auditResultReports.PUT("/:id", ctrlAuditExecution.SyncOnWrite(h.db, func() interface{} { return &models.AuditResultReport{} }, ctrlAuditExecution.LetterKeysOf, crud.UpdateWithHooks(h.db, "AuditResultReport", func() interface{} { return &models.AuditResultReport{} }, nil, SyncReportATRs)))
		auditResultReports.DELETE("/:id", ctrlAuditExecution.SyncOnWrite(h.db, func() interface{} { return &models.AuditResultReport{} }, ctrlAuditExecution.LetterKeysOf, crud.Delete(h.db, "AuditResultReport", func() interface{} { return &models.AuditResultReport{} })))
	}

	// 17c. Auditee Surveys
	auditeeSurveys := apiV1.Group("/auditee-surveys")
	{
		auditeeSurveys.GET("", crud.List(h.db, "AuditeeSurvey", func() interface{} { return &[]models.AuditeeSurvey{} }))
		auditeeSurveys.GET("/:id", crud.GetByID(h.db, "AuditeeSurvey", func() interface{} { return &models.AuditeeSurvey{} }))
		auditeeSurveys.POST("", crud.Create(h.db, "AuditeeSurvey", func() interface{} { return &models.AuditeeSurvey{} }))
		auditeeSurveys.PUT("/:id", crud.Update(h.db, "AuditeeSurvey", func() interface{} { return &models.AuditeeSurvey{} }))
		auditeeSurveys.DELETE("/:id", crud.Delete(h.db, "AuditeeSurvey", func() interface{} { return &models.AuditeeSurvey{} }))
	}

	// 17b. Executive Summary Reports
	executiveSummaries := apiV1.Group("/executive-summaries")
	{
		// List rows leave out the approval signature image (ListDecorator)
		executiveSummaries.GET("", crud.ListWithItems(h.db, "ExecutiveSummary", func() interface{} { return &[]models.ExecutiveSummary{} }, ctrlExecutiveSummary.ListDecorator))
		// Read-only aggregates for the compilation form. Lives under /executive-summaries
		// because that prefix is already routed by every Kong config (dev, prod, tenant
		// template); /executive-summary/* is not, and Kong answers it with 404
		// "no Route matched with those values". Registered before /:id for readability;
		// gin matches the static segment first either way.
		executiveSummaries.GET("/compilation-stats", ctrlExecutiveSummary.CompilationStatsHandler(h.db))
		executiveSummaries.GET("/:id", crud.GetByID(h.db, "ExecutiveSummary", func() interface{} { return &models.ExecutiveSummary{} }))
		// A summary cannot be created as Approved, see GuardCreateHook
		executiveSummaries.POST("", crud.CreateWithHook(h.db, "ExecutiveSummary", func() interface{} { return &models.ExecutiveSummary{} }, ctrlExecutiveSummary.GuardCreateHook))
		// Draft -> Approved only through /:id/approve (signature required). Approved is locked,
		// and the status guard and risk count sync run in UpdateHook
		executiveSummaries.PUT("/:id", crud.UpdateWithHook(h.db, "ExecutiveSummary", func() interface{} { return &models.ExecutiveSummary{} }, ctrlExecutiveSummary.UpdateHook))
		// Approval with signature; ADMIN or EXECUTIVE only (see ApproveHandler)
		executiveSummaries.POST("/:id/approve", ctrlExecutiveSummary.ApproveHandler(h.db))
		executiveSummaries.DELETE("/:id", crud.Delete(h.db, "ExecutiveSummary", func() interface{} { return &models.ExecutiveSummary{} }))
	}

	// Read-only aggregates for the compilation form ("Statistik Kompilasi" and
	// "Status Tindak Lanjut"), counted from the source tables for one year.
	// Legacy path: not routed by Kong, so only reachable when calling audit-service
	// directly. The frontend uses /executive-summaries/compilation-stats (above).
	executiveSummaryStats := apiV1.Group("/executive-summary")
	{
		executiveSummaryStats.GET("/compilation-stats", ctrlExecutiveSummary.CompilationStatsHandler(h.db))
	}

	// 18. Action Taken Reports: one per finding of an approved LHA, created by
	// the server (no generic create/update). Role checks live in the
	// controller (services/action_taken_report.RolePermissions).
	ctrlATR.NewController(h.db).Register(apiV1.Group("/action-taken-reports"))

	// 19. Media routes
	media := apiV1.Group("/media")
	{
		media.POST("/upload", h.registry.Media.Upload)
		// Under apiV1, so it requires a valid token like the rest of the API.
		media.GET("/download/:id", h.registry.Media.Download)
	}

	// 20. Performance routes
	perfStatsCtrl := controllers.NewPerformanceStatsController(h.db, h.redisClient)
	performance := apiV1.Group("/performance")
	{
		performance.GET("/dashboard-summary", perfStatsCtrl.GetDashboardSummary)
		performance.GET("/monthly-trends", perfStatsCtrl.GetMonthlyTrends)
		performance.GET("/kpi-breakdown", perfStatsCtrl.GetKpiBreakdown)
		performance.GET("/completion-analysis", perfStatsCtrl.GetCompletionAnalysis)
		performance.GET("/completion-history", perfStatsCtrl.GetCompletionHistory)
		performance.GET("/export-pdf", perfStatsCtrl.ExportPdfReport)

		performance.GET("/kpi", crud.List(h.db, "KPIAchievement", func() interface{} { return &[]models.KPIAchievement{} }))
		performance.GET("/kpi/:id", crud.GetByID(h.db, "KPIAchievement", func() interface{} { return &models.KPIAchievement{} }))
		performance.POST("/kpi", crud.Create(h.db, "KPIAchievement", func() interface{} { return &models.KPIAchievement{} }))
		performance.PUT("/kpi/:id", crud.Update(h.db, "KPIAchievement", func() interface{} { return &models.KPIAchievement{} }))
		performance.DELETE("/kpi/:id", crud.Delete(h.db, "KPIAchievement", func() interface{} { return &models.KPIAchievement{} }))

		performance.GET("/realization", crud.List(h.db, "WorkPlanRealization", func() interface{} { return &[]models.WorkPlanRealization{} }))
		performance.GET("/realization/:id", crud.GetByID(h.db, "WorkPlanRealization", func() interface{} { return &models.WorkPlanRealization{} }))
		performance.POST("/realization", crud.Create(h.db, "WorkPlanRealization", func() interface{} { return &models.WorkPlanRealization{} }))
		performance.PUT("/realization/:id", crud.Update(h.db, "WorkPlanRealization", func() interface{} { return &models.WorkPlanRealization{} }))
		performance.DELETE("/realization/:id", crud.Delete(h.db, "WorkPlanRealization", func() interface{} { return &models.WorkPlanRealization{} }))
	}
}

// SyncReportATRs is the after-write hook of the audit result report create and
// update routes (see atrsvc.AfterReportWrite): approving a report creates its
// action taken reports in the same transaction.
var SyncReportATRs crud.AfterHook = atrsvc.AfterReportWrite

// syncReportAndExecutionsAfterCreate is the create hook of the LHA route: the
// ATR sync, then the sync of the report's letter executions, in one transaction.
var syncReportAndExecutionsAfterCreate crud.AfterHook = func(tx *gorm.DB, entity interface{}) error {
	if err := SyncReportATRs(tx, entity); err != nil {
		return err
	}
	return ctrlAuditExecution.SyncAfterCreate(ctrlAuditExecution.LetterKeysOf)(tx, entity)
}

func (h *RouteHandler) deleteAuditGuideline(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid AuditGuideline ID",
		})
		return
	}

	var guideline models.AuditGuideline

	if err := h.db.First(&guideline, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "Audit guideline not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch AuditGuideline",
		})
		return
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		// Soft-delete seluruh SOP/Juknis yang menggunakan Pedoman ini.
		if err := tx.
			Where("guideline_id = ?", id).
			Delete(&models.AuditSop{}).Error; err != nil {
			return err
		}

		// Setelah child record berhasil dihapus, soft-delete Pedoman.
		if err := tx.Delete(&guideline).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to delete AuditGuideline",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "AuditGuideline deleted successfully",
	})
}

type AutoFindingItem struct {
	Title    string `json:"title"`
	Category string `json:"category"`
	Action   string `json:"action"`
	Source   string `json:"source"`
	Impact   string `json:"impact,omitempty"`
	Criteria string `json:"criteria,omitempty"`
}

func (h *RouteHandler) getAutoFindings(c *gin.Context) {
	assignmentLetterId := strings.TrimSpace(c.Query("assignmentLetterId"))
	if assignmentLetterId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "assignmentLetterId query parameter is required",
		})
		return
	}

	// Live findings: KKA (F04 causes / F05 plans / F02 risks) and fieldwork
	// test controls of this assignment letter (see controllers/findings). ATRs
	// are not a source: they are created from the LHA's findings.
	var src ctrlFindings.LiveSources
	h.db.Where("working_paper_id = ?", assignmentLetterId).Order("created_at ASC").Order("id ASC").Find(&src.Causes)
	h.db.Where("working_paper_id = ?", assignmentLetterId).Order("created_at ASC").Order("id ASC").Find(&src.Plans)
	h.db.Where("working_paper_id = ?", assignmentLetterId).Find(&src.Risks)
	h.db.Where("assignment_letter_id = ?", assignmentLetterId).Order("created_at ASC").Order("id ASC").Find(&src.TestControls)

	var findings []AutoFindingItem
	for _, f := range ctrlFindings.Derive(src) {
		findings = append(findings, AutoFindingItem{
			Title:    f.Title,
			Category: f.Category,
			Action:   f.Action,
			Source:   f.Source,
			Impact:   f.Impact,
			Criteria: f.Criteria,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"assignmentLetterId": assignmentLetterId,
			"total":              len(findings),
			"findings":           findings,
		},
	})
}

func (h *RouteHandler) downloadAuditResultReportDocx(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var report models.AuditResultReport
	if err := h.db.First(&report, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Audit result report not found"})
		return
	}

	// Fetch related entities by AssignmentLetterID
	var st models.AssignmentLetter
	if report.AssignmentLetterID != "" {
		trimmedLetterID := strings.TrimSpace(report.AssignmentLetterID)
		if err := h.db.Where("letter_number = ?", trimmedLetterID).First(&st).Error; err != nil {
			if _, parseErr := uuid.Parse(trimmedLetterID); parseErr == nil {
				h.db.Where("id = ?", trimmedLetterID).First(&st)
			}
		}
	}

	altIDs := []string{}
	if report.AssignmentLetterID != "" {
		altIDs = append(altIDs, strings.TrimSpace(report.AssignmentLetterID))
	}
	if st.LetterNumber != "" && st.LetterNumber != report.AssignmentLetterID {
		altIDs = append(altIDs, strings.TrimSpace(st.LetterNumber))
	}
	if st.ID != uuid.Nil {
		altIDs = append(altIDs, st.ID.String())
	}

	var interviews []models.FieldworkInterview
	if len(altIDs) > 0 {
		h.db.Where("assignment_letter_id IN ?", altIDs).Find(&interviews)
	}

	var observations []models.FieldworkObservation
	if len(altIDs) > 0 {
		h.db.Where("assignment_letter_id IN ?", altIDs).Find(&observations)
	}

	var fieldworkDocs []models.FieldworkDocument
	if len(altIDs) > 0 {
		h.db.Where("assignment_letter_id IN ?", altIDs).Find(&fieldworkDocs)
	}

	var fieldworkSamples []models.FieldworkSample
	if len(altIDs) > 0 {
		h.db.Where("assignment_letter_id IN ?", altIDs).Find(&fieldworkSamples)
	}

	var wpHeader models.WorkingPaperHeader
	if len(altIDs) > 0 {
		h.db.Where("assignment_letter_id IN ?", altIDs).First(&wpHeader)
	}
	// Fallback/sync header fields from Assignment Letter if empty
	if wpHeader.AuditPurpose == "" && st.AuditPurpose != "" {
		wpHeader.AuditPurpose = st.AuditPurpose
	}
	if len(wpHeader.TeamMembers) == 0 && len(st.MembersList) > 0 {
		members := make([]models.TeamMember, 0, len(st.MembersList))
		for idx, m := range st.MembersList {
			members = append(members, models.TeamMember{
				ID:   time.Now().UnixMilli() + int64(idx),
				Name: m.Name,
				Role: m.Role,
			})
		}
		wpHeader.TeamMembers = members
	}

	var wpRisks []models.WorkingPaperRisk
	if len(altIDs) > 0 {
		h.db.Where("working_paper_id IN ? OR assignment_letter_id IN ?", altIDs, altIDs).Find(&wpRisks)
	}

	var wpSamples []models.WorkingPaperSample
	if len(altIDs) > 0 {
		h.db.Where("working_paper_id IN ?", altIDs).Find(&wpSamples)
	}

	var wpCauses []models.WorkingPaperCause
	if len(altIDs) > 0 {
		h.db.Where("working_paper_id IN ?", altIDs).Find(&wpCauses)
	}

	var wpPlans []models.WorkingPaperPlan
	if len(altIDs) > 0 {
		h.db.Where("working_paper_id IN ?", altIDs).Find(&wpPlans)
	}

	var importedWPs []models.ImportedWorkingPaper
	h.db.Find(&importedWPs)

	// Resolve Company Name dynamically from master-service / st / report
	resolvedComp := masterclient.ResolveCompanyName(c.Request.Context(), c.GetHeader("Authorization"), report.CompanyID, report.CompanyName, st.CompanyID, st.CompanyName, st.WorkingUnit, report.Department)
	if report.CompanyName == "" {
		report.CompanyName = resolvedComp
	}
	if st.CompanyName == "" {
		st.CompanyName = resolvedComp
	}

	docxBytes, err := docxbuilder.GenerateAuditReportDocx(&report, &st, interviews, observations, fieldworkDocs, fieldworkSamples, &wpHeader, wpRisks, wpSamples, wpCauses, wpPlans, importedWPs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Docx file", "details": err.Error()})
		return
	}

	cleanNum := strings.ReplaceAll(report.ReportNumber, "/", "_")
	cleanNum = strings.ReplaceAll(cleanNum, " ", "")
	if cleanNum == "" {
		cleanNum = "LHA"
	}
	filename := fmt.Sprintf("LHA_%s.docx", cleanNum)

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.wordprocessingml.document", docxBytes)
}
