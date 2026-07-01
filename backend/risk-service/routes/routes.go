package routes

import (
	"log"
	"net/http"
	"risk-service/controllers"

	"github.com/gin-gonic/gin"
)

type RiskRoute struct {
	riskCtrl       *controllers.RiskController
	mitigationCtrl *controllers.MitigationController
	group          *gin.RouterGroup
}

type IRiskRoute interface {
	Run()
}

func NewRiskRoute(
	riskCtrl *controllers.RiskController,
	mitigationCtrl *controllers.MitigationController,
	group *gin.RouterGroup,
) IRiskRoute {
	return &RiskRoute{
		riskCtrl:       riskCtrl,
		mitigationCtrl: mitigationCtrl,
		group:          group,
	}
}

func (r *RiskRoute) Run() {
	// Health endpoint registered under the main group (root)
	r.group.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "risk-service",
		})
	})

	apiV1 := r.group.Group("/api/v1")
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
	}
}
