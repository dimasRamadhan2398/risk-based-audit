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

// GetRiskScore handles POST & GET /api/analytics/risk-score
func (c *AnalyticsController) GetRiskScore(ctx *gin.Context) {
	var req services.DepartmentRiskRequest
	if ctx.Request.Method == http.MethodPost {
		_ = ctx.ShouldBindJSON(&req)
	}

	res, err := c.service.GetRiskScore(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   res,
	})
}

func (c *AnalyticsController) GetRiskScoreBatch(ctx *gin.Context) {
	res, err := c.service.GetRiskScoreBatch()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   res,
	})
}

// GetAnomaly handles POST & GET /api/analytics/anomaly
func (c *AnalyticsController) GetAnomaly(ctx *gin.Context) {
	var req services.AnomalyRequest
	if ctx.Request.Method == http.MethodPost {
		_ = ctx.ShouldBindJSON(&req)
	}

	res, err := c.service.GetAnomaly(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   res,
	})
}

func (c *AnalyticsController) GetAnomalyBatch(ctx *gin.Context) {
	res, err := c.service.GetAnomalyBatch()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   res,
	})
}

// GetTextAnalysis handles POST & GET /api/analytics/text-analysis
func (c *AnalyticsController) GetTextAnalysis(ctx *gin.Context) {
	var req struct {
		Text string `json:"text"`
	}
	if ctx.Request.Method == http.MethodPost {
		_ = ctx.ShouldBindJSON(&req)
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

func (c *AnalyticsController) GetTextAnalysisBatch(ctx *gin.Context) {
	res, err := c.service.GetTextAnalysisBatch()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   res,
	})
}

// GetPerformanceTrend handles POST & GET /api/analytics/performance-trend
func (c *AnalyticsController) GetPerformanceTrend(ctx *gin.Context) {
	var req services.PerformanceTrendRequest
	if ctx.Request.Method == http.MethodPost {
		_ = ctx.ShouldBindJSON(&req)
	}
	res, err := c.service.GetPerformanceTrend(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   res,
	})
}

func (c *AnalyticsController) GetPerformanceTrendBatch(ctx *gin.Context) {
	res, err := c.service.GetPerformanceTrendBatch()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   res,
	})
}

func (c *AnalyticsController) TriggerAutoRetrain(ctx *gin.Context) {
	res, err := c.service.TriggerAutoRetrain()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   res,
	})
}
