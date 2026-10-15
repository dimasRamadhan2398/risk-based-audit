package routes

import (
	"analytics-service/controllers"
	"analytics-service/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, authMiddleware *middleware.AuthMiddleware) {
	analyticsController := controllers.NewAnalyticsController()

	api := r.Group("/api/analytics")
	// Every analytics route requires a valid JWT. Kong does not verify tokens
	// for this service, so without this the dashboard aggregates, risk scores
	// and CAATT results were readable by anyone who could reach the gateway -
	// and /retrain/auto was anonymously triggerable.
	api.Use(authMiddleware.Authenticate())
	{
		api.GET("/report", analyticsController.GetReport)
		api.GET("/predict", analyticsController.GetPredictiveTrends)
		
		api.GET("/risk-score", analyticsController.GetRiskScore)
		api.POST("/risk-score", analyticsController.GetRiskScore)
		api.GET("/risk-score/batch", analyticsController.GetRiskScoreBatch)

		api.GET("/anomaly", analyticsController.GetAnomaly)
		api.POST("/anomaly", analyticsController.GetAnomaly)
		api.GET("/anomaly/batch", analyticsController.GetAnomalyBatch)

		api.GET("/text-analysis", analyticsController.GetTextAnalysis)
		api.POST("/text-analysis", analyticsController.GetTextAnalysis)
		api.GET("/text-analysis/batch", analyticsController.GetTextAnalysisBatch)

		api.GET("/performance-trend", analyticsController.GetPerformanceTrend)
		api.POST("/performance-trend", analyticsController.GetPerformanceTrend)
		api.GET("/performance-trend/batch", analyticsController.GetPerformanceTrendBatch)

		api.POST("/retrain/auto", analyticsController.TriggerAutoRetrain)
		
		// CAATT Analytics routes
		caatt := api.Group("/caatt")
		{
			caatt.GET("/full-population", analyticsController.GetFullPopulation)
			caatt.GET("/duplicate-gap", analyticsController.GetDuplicateGap)
			caatt.GET("/benford", analyticsController.GetBenfordAnalysis)
			caatt.GET("/stratification", analyticsController.GetStratification)
			caatt.GET("/reconciliation", analyticsController.GetReconciliation)
			caatt.GET("/policy-violations", analyticsController.GetPolicyViolations)
			caatt.GET("/data-quality", analyticsController.GetDataQualityMetrics)
		}
	}
}
