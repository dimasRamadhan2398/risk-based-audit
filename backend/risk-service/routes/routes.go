package routes

import (
	"log"
	"net/http"
	"risk-service/controllers"
	"risk-service/middleware"

	"github.com/gin-gonic/gin"
)

type RiskRoute struct {
	riskCtrl          *controllers.RiskController
	mitigationCtrl    *controllers.MitigationController
	riskFactorCtrl    *controllers.RiskFactorController
	auditUniverseCtrl *controllers.AuditUniverseController
	rcmCtrl           *controllers.RCMController
	group             *gin.RouterGroup
	authMiddleware    *middleware.AuthMiddleware
}

type IRiskRoute interface {
	Run()
}

func NewRiskRoute(
	riskCtrl *controllers.RiskController,
	mitigationCtrl *controllers.MitigationController,
	riskFactorCtrl *controllers.RiskFactorController,
	auditUniverseCtrl *controllers.AuditUniverseController,
	rcmCtrl *controllers.RCMController,
	group *gin.RouterGroup,
	authMiddleware *middleware.AuthMiddleware,
) IRiskRoute {
	return &RiskRoute{
		riskCtrl:          riskCtrl,
		mitigationCtrl:    mitigationCtrl,
		riskFactorCtrl:    riskFactorCtrl,
		auditUniverseCtrl: auditUniverseCtrl,
		rcmCtrl:           rcmCtrl,
		group:             group,
		authMiddleware:    authMiddleware,
	}
}

func (r *RiskRoute) Run() {
	// Health endpoint registered under the main group (root), deliberately
	// before the authenticated group: Kong's active healthcheck sends no token.
	r.group.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "risk-service",
		})
	})

	apiV1 := r.group.Group("/api/v1")
	// Every risk route requires a valid JWT. Kong does not verify tokens for
	// this service, so without this the risk register and RCM were readable and
	// writable by anyone who could reach the gateway.
	apiV1.Use(r.authMiddleware.Authenticate())
	// VIEWER is read-only: every POST/PUT/PATCH/DELETE is refused by default.
	// No risk-service POST is a read, so there is no exemption.
	apiV1.Use(middleware.DenyViewerWrites())
	{
		log.Printf("[RiskRoute] Registered routes under /api/v1/risks:")
		log.Printf("  GET    /api/v1/risks")
		log.Printf("  POST   /api/v1/risks")
		log.Printf("  PUT    /api/v1/risks/:id")
		log.Printf("  DELETE /api/v1/risks/:id")

		apiV1.GET("/risks", r.riskCtrl.ListRisks)
		apiV1.POST("/risks", r.riskCtrl.CreateRisk)
		apiV1.PUT("/risks/:id", r.riskCtrl.UpdateRisk)
		apiV1.DELETE("/risks/:id", r.riskCtrl.DeleteRisk)

		log.Printf("[RiskRoute] Registered routes under /api/v1/mitigations:")
		log.Printf("  GET    /api/v1/mitigations")
		log.Printf("  POST   /api/v1/mitigations")
		log.Printf("  PUT    /api/v1/mitigations/:id")
		log.Printf("  DELETE /api/v1/mitigations/:id")

		apiV1.GET("/mitigations", r.mitigationCtrl.ListMitigations)
		apiV1.POST("/mitigations", r.mitigationCtrl.CreateMitigation)
		apiV1.PUT("/mitigations/:id", r.mitigationCtrl.UpdateMitigation)
		apiV1.DELETE("/mitigations/:id", r.mitigationCtrl.DeleteMitigation)

		// New routes for Risk Factors
		// Standard risk factor library: writes limited to auditUniverseWriters.
		stdFactors := apiV1.Group("/risk-factors/standard", middleware.RequireAnyRoleOnWrites(auditUniverseWriters...))
		stdFactors.GET("", r.riskFactorCtrl.ListStandardRiskFactors)
		stdFactors.POST("", r.riskFactorCtrl.CreateStandardRiskFactor)
		stdFactors.PUT("/:id", r.riskFactorCtrl.UpdateStandardRiskFactor)
		stdFactors.DELETE("/:id", r.riskFactorCtrl.DeleteStandardRiskFactor)
		apiV1.GET("/risk-factors/corporate", r.riskFactorCtrl.ListCorporateRiskFactors)
		apiV1.POST("/risk-factors/corporate", r.riskFactorCtrl.UpdateCorporateRiskFactors)

		// Audit Universe: writes limited to auditUniverseWriters (non-GET only).
		universe := apiV1.Group("/audit-universe", middleware.RequireAnyRoleOnWrites(auditUniverseWriters...))
		universe.GET("/standard", r.auditUniverseCtrl.ListStandardAuditUniverse)
		universe.GET("/corporate", r.auditUniverseCtrl.ListCorporateAuditUniverse)
		universe.POST("/corporate/sync-master", r.auditUniverseCtrl.SyncMasterData)
		// Corporate entries are the only audit-universe routes DEPARTMENT_HEAD
		// reaches; the controller scopes it to its own department. A sibling group
		// with the same prefix keeps every other (and future) universe route on the
		// full-writer list by default.
		corporate := apiV1.Group("/audit-universe/corporate", middleware.RequireAnyRoleOnWrites(corporateWriters...))
		corporate.POST("", r.auditUniverseCtrl.CreateOrUpdateCorporateNode)
		corporate.DELETE("/:id", r.auditUniverseCtrl.DeleteCorporateNode)

		universe.GET("/year/:year", r.auditUniverseCtrl.ListYearlyAuditUniverse)
		universe.POST("/year/:year", r.auditUniverseCtrl.EstablishYearlyUniverse)
		universe.POST("/year/:year/score", r.auditUniverseCtrl.ScoreYearlyEntity)

		// Re-emits the full universe state to Kafka (see pkg/kafka/events.go).
		universe.POST("/events/resync", r.auditUniverseCtrl.ResyncEvents)

		// Routes for Risk Control Matrix (RCM)
		apiV1.GET("/rcm", r.rcmCtrl.ListRCM)
		apiV1.POST("/rcm", r.rcmCtrl.CreateRCM)
		apiV1.PUT("/rcm/:id", r.rcmCtrl.UpdateRCM)
		apiV1.DELETE("/rcm/:id", r.rcmCtrl.DeleteRCM)
		apiV1.GET("/rcm/summary", r.rcmCtrl.GetRCMSummary)
	}
}
