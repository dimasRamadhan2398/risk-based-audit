package routes

import (
	"analytics-service/controllers"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	analyticsController := controllers.NewAnalyticsController()

	api := r.Group("/api/analytics")
	{
		api.GET("/report", analyticsController.GetReport)
		api.GET("/predict", analyticsController.GetPredictiveTrends)
		api.GET("/risk-score", analyticsController.GetRiskScore)
		api.GET("/anomaly", analyticsController.GetAnomaly)
		api.POST("/text-analysis", analyticsController.GetTextAnalysis)
		api.POST("/performance-trend", analyticsController.GetPerformanceTrend)
	}
}
