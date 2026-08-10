package controllers

import (
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

func (c *ReportTimelinessController) CreateOrUpdate(ctx *gin.Context) {
	var req services.ReportTimelinessReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := c.service.CreateOrUpdate(&req)
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
		ctx.JSON(http.StatusNotFound, gin.H{"message": "Not found"})
		return
	}

	ctx.JSON(http.StatusOK, res)
}
