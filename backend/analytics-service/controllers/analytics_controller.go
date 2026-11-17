package controllers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
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

func (c *AnalyticsController) SyncEntityPredictionProfile(ctx *gin.Context) {
	var req interface{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}
	res, err := c.service.SyncEntityPredictionProfile(req)
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
	var req services.TextRequest
	if ctx.Request.Method == http.MethodPost {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			// Check if multipart form data was sent (direct file upload)
			if file, header, formErr := ctx.Request.FormFile("file"); formErr == nil {
				defer file.Close()
				buf := new(bytes.Buffer)
				if _, copyErr := io.Copy(buf, file); copyErr == nil {
					req.FileName = header.Filename
					req.FileContentBase64 = base64.StdEncoding.EncodeToString(buf.Bytes())
				}
				req.Title = ctx.PostForm("title")
				req.Source = ctx.PostForm("source")
				req.SourceType = ctx.PostForm("sourceType")
				req.Text = ctx.PostForm("text")
				req.RiskCategory = ctx.PostForm("riskCategory")
			}
		}
	}
	res, err := c.service.AnalyzeDocumentOrText(req)
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

func (c *AnalyticsController) GetNLPTaxonomy(ctx *gin.Context) {
	res, err := c.service.GetNLPTaxonomy()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   res,
	})
}

func (c *AnalyticsController) UpdateNLPTaxonomy(ctx *gin.Context) {
	var payload map[string]interface{}
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := c.service.UpdateNLPTaxonomy(payload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   res,
	})
}

func (c *AnalyticsController) SyncNLPMultiSources(ctx *gin.Context) {
	res, err := c.service.SyncNLPMultiSources()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   res,
	})
}

func (c *AnalyticsController) DeleteNLPAnalyzeDocument(ctx *gin.Context) {
	docID := ctx.Param("docId")
	if docID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "docId is required"})
		return
	}
	res, err := c.service.DeleteNLPAnalyzeDocument(docID)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"status": "deleted_locally", "docId": docID})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   res,
	})
}

func (c *AnalyticsController) ResetNLPAnalyzeDocuments(ctx *gin.Context) {
	res, err := c.service.ResetNLPAnalyzeDocuments()
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"status": "reset_locally"})
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

// PredictAuditorKPIForecast handles POST /api/analytics/performance-trend/auditor-forecast.
func (c *AnalyticsController) PredictAuditorKPIForecast(ctx *gin.Context) {
	var payload map[string]interface{}
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	res, err := c.service.PredictAuditorKPIForecast(payload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   res,
	})
}

// GetRetrainStatus handles GET /api/analytics/retrain/status.
func (c *AnalyticsController) GetRetrainStatus(ctx *gin.Context) {
	res, err := c.service.GetRetrainStatus(ctx.Request.URL.RawQuery)
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
	force := ctx.Query("force") == "true" || ctx.Query("force") == "1"
	var payload map[string]interface{}
	if ctx.Request.Body != nil && ctx.Request.ContentLength > 0 {
		if err := ctx.ShouldBindJSON(&payload); err == nil {
			if bodyForce, ok := payload["force"].(bool); ok && bodyForce {
				force = true
			}
		}
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}
	if force {
		payload["force"] = true
		payload["wait"] = true
	}

	res, err := c.service.TriggerAutoRetrainWithOptions(force, payload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   res,
	})
}

// TriggerModelRetrain handles POST /api/analytics/retrain/model/:model.
func (c *AnalyticsController) TriggerModelRetrain(ctx *gin.Context) {
	model := ctx.Param("model")
	if !services.IsValidRetrainModel(model) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Unknown model '" + model + "'. Valid models: entity, anomaly, nlp, kpi, kpi_auditee, kpi_auditor.",
		})
		return
	}
	var payload map[string]interface{}
	if ctx.Request.Body != nil && ctx.Request.ContentLength > 0 {
		_ = ctx.ShouldBindJSON(&payload)
	}
	if payload == nil {
		payload = map[string]interface{}{}
	}
	if ctx.Query("force") == "true" || ctx.Query("force") == "1" {
		payload["force"] = true
	}
	if ctx.Query("wait") == "true" || ctx.Query("wait") == "1" {
		payload["wait"] = true
	}

	res, err := c.service.TriggerModelRetrain(model, payload)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   res,
	})
}

// ─── CAATT Analytics Handlers ────────────────────────────────────────────────

// TriggerCAATTRun handles POST /api/analytics/caatt/run
func (c *AnalyticsController) TriggerCAATTRun(ctx *gin.Context) {
	sourceID := ctx.Query("source_id")
	if sourceID == "" {
		var body struct {
			SourceID string `json:"source_id"`
		}
		if err := ctx.ShouldBindJSON(&body); err == nil && body.SourceID != "" {
			sourceID = body.SourceID
		}
	}

	res, err := c.service.TriggerCAATTRun(sourceID)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, res)
}

// GetFullPopulation handles GET /api/analytics/caatt/full-population
func (c *AnalyticsController) GetFullPopulation(ctx *gin.Context) {
	sourceID := ctx.Query("source_id")
	limitStr := ctx.DefaultQuery("limit", "100")
	limit := 100
	fmt.Sscanf(limitStr, "%d", &limit)

	res, err := c.service.GetCAATTFullPopulation(sourceID, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    res.Data,
		"summary": res.Summary,
	})
}

// GetDuplicateGap handles GET /api/analytics/caatt/duplicate-gap
func (c *AnalyticsController) GetDuplicateGap(ctx *gin.Context) {
	sourceID := ctx.Query("source_id")
	resultType := ctx.Query("result_type")
	limitStr := ctx.DefaultQuery("limit", "100")
	limit := 100
	fmt.Sscanf(limitStr, "%d", &limit)

	res, err := c.service.GetCAATTDuplicateGap(sourceID, resultType, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    res.Data,
		"summary": res.Summary,
	})
}

// GetBenfordAnalysis handles GET /api/analytics/caatt/benford
func (c *AnalyticsController) GetBenfordAnalysis(ctx *gin.Context) {
	sourceID := ctx.Query("source_id")
	digit := ctx.Query("digit")
	findingsLimit := ctx.Query("findings_limit")
	res, err := c.service.GetCAATTBenfordWithParams(sourceID, digit, findingsLimit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"data":     res.Data,
		"findings": res.Findings,
		"summary":  res.Summary,
	})
}

// GetStratification handles GET /api/analytics/caatt/stratification
func (c *AnalyticsController) GetStratification(ctx *gin.Context) {
	sourceID := ctx.Query("source_id")
	category := ctx.Query("category")

	res, err := c.service.GetCAATTStratification(sourceID, category)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    res.Data,
		"summary": res.Summary,
	})
}

// GetReconciliation handles GET /api/analytics/caatt/reconciliation
func (c *AnalyticsController) GetReconciliation(ctx *gin.Context) {
	sourceID := ctx.Query("source_id")
	res, err := c.service.GetCAATTReconciliation(sourceID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    res.Data,
		"summary": res.Summary,
	})
}

// GetPolicyViolations handles GET /api/analytics/caatt/policy-violations
func (c *AnalyticsController) GetPolicyViolations(ctx *gin.Context) {
	sourceID := ctx.Query("source_id")
	severity := ctx.Query("severity")
	limitStr := ctx.DefaultQuery("limit", "100")
	limit := 100
	fmt.Sscanf(limitStr, "%d", &limit)

	res, err := c.service.GetCAATTPolicyViolations(sourceID, severity, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    res.Data,
		"summary": res.Summary,
	})
}

// GetDataQualityMetrics handles GET /api/analytics/caatt/data-quality
func (c *AnalyticsController) GetDataQualityMetrics(ctx *gin.Context) {
	sourceID := ctx.Query("source_id")
	res, err := c.service.GetDataQualityMetrics(sourceID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status":                "success",
		"data":                  res.Data,
		"summary":               res.Summary,
		"overall_quality_score": res.OverallQualityScore,
	})
}
