package routes

import (
	"audit-service/pkg/middleware"
	"fmt"

	"github.com/gin-gonic/gin"
)

// RouteHandler handles all route registration
type RouteHandler struct {
	engine         *gin.Engine
	registry       *RouteRegistry
	authMiddleware *middleware.AuthMiddleware
}

// NewRouteHandler creates a new route handler
func NewRouteHandler(
	engine *gin.Engine,
	authMiddleware *middleware.AuthMiddleware,
) *RouteHandler {
	return &RouteHandler{
		engine:         engine,
		registry:       NewRouteRegistry(),
		authMiddleware: authMiddleware,
	}
}

// SetRegistry sets the route registry with controllers
func (h *RouteHandler) SetRegistry(registry *RouteRegistry) {
	h.registry = registry
}

// RegisterRoutes registers all API routes
func (h *RouteHandler) RegisterRoutes() {
	// Debug route to list all registered routes
	h.engine.GET("/debug/routes", func(c *gin.Context) {
		routes := []gin.H{}
		for _, route := range h.engine.Routes() {
			routes = append(routes, gin.H{
				"method": route.Method,
				"path":   route.Path,
			})
		}
		c.JSON(200, gin.H{
			"routes": routes,
		})
	})

	apiV1 := h.engine.Group("/api/v1")

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

	// Print registered routes for debugging
	fmt.Println("\n=== Registered Routes ===")
	for _, route := range h.engine.Routes() {
		fmt.Printf("%s %s\n", route.Method, route.Path)
	}
	fmt.Println("========================\n")
}
