package controllers

import (
	"net/http"

	"analytics-service/services"
	"github.com/gin-gonic/gin"
)

type AnalyticsController struct {
	service *services.AnalyticsService
}

func NewAnalyticsController() *AnalyticsController {
	return &AnalyticsController{
		service: services.NewAnalyticsService(),
	}
}

// GetReport handles GET /api/analytics/report
func (c *AnalyticsController) GetReport(ctx *gin.Context) {
	report := c.service.GenerateReport()
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   report,
	})
}

// GetPredictiveTrends handles GET /api/analytics/predict
func (c *AnalyticsController) GetPredictiveTrends(ctx *gin.Context) {
	prediction := c.service.PredictFutureTrends()
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   prediction,
	})
}

// GetRiskScore handles GET /api/analytics/risk-score
func (c *AnalyticsController) GetRiskScore(ctx *gin.Context) {
	// Dummy input values for now, normally coming from query params
	res, err := c.service.GetRiskScore(0.5, 0.5, 0.5)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   res,
	})
}

// GetAnomaly handles GET /api/analytics/anomaly
func (c *AnalyticsController) GetAnomaly(ctx *gin.Context) {
	// Dummy input values
	res, err := c.service.GetAnomaly(1.0, -1.0)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   res,
	})
}

// GetTextAnalysis handles POST /api/analytics/text-analysis
func (c *AnalyticsController) GetTextAnalysis(ctx *gin.Context) {
	var req struct {
		Text string `json:"text"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		req.Text = "Dummy finding text indicating potential fraud." // fallback
	}
	res, err := c.service.GetTextAnalysis(req.Text)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   res,
	})
}

// GetPerformanceTrend handles POST /api/analytics/performance-trend
func (c *AnalyticsController) GetPerformanceTrend(ctx *gin.Context) {
	var req struct {
		HistoricalData []float64 `json:"historical_data"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil || len(req.HistoricalData) == 0 {
		req.HistoricalData = []float64{0.8, 0.82, 0.85, 0.81, 0.79} // fallback
	}
	res, err := c.service.GetPerformanceTrend(req.HistoricalData)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   res,
	})
}
