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
	}
}
