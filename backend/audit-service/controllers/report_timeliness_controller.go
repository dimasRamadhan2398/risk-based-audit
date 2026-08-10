package controllers

import (
	"audit-service/models"
	"audit-service/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ReportTimelinessController struct {
	service services.IReportTimelinessService
}

func NewReportTimelinessController(service services.IReportTimelinessService) *ReportTimelinessController {
	return &ReportTimelinessController{service: service}
}

type SaveMultipleReq struct {
	Year           int                            `json:"year"`
	Period         string                         `json:"period"`
	Questionnaires []services.ReportTimelinessReq `json:"questionnaires"`
}

func (c *ReportTimelinessController) CreateOrUpdate(ctx *gin.Context) {
	var req SaveMultipleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := c.service.SaveMultiple(req.Year, req.Period, req.Questionnaires)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, res)
}

func (c *ReportTimelinessController) GetByYearAndPeriod(ctx *gin.Context) {
	yearStr := ctx.Query("year")
	period := ctx.Query("period")

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year"})
		return
	}

	res, err := c.service.GetByYearAndPeriod(year, period)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if res == nil {
		res = []models.ReportTimeliness{}
	}

	ctx.JSON(http.StatusOK, res)
}
