package controllers

import (
	"net/http"
	"risk-service/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RiskController struct {
	service services.IRiskService
}

func NewRiskController(service services.IRiskService) *RiskController {
	return &RiskController{service: service}
}

func (ctrl *RiskController) ListRisks(c *gin.Context) {
	data, err := ctrl.service.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to query risks",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Risks fetched successfully",
		"data":    data,
	})
}

func (ctrl *RiskController) CreateRisk(c *gin.Context) {
	var req services.RiskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	res, err := ctrl.service.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "DB_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Risk created successfully",
		"data":    res,
	})
}

func (ctrl *RiskController) UpdateRisk(c *gin.Context) {
	idStr := c.Param("id")
	regID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid risk ID format",
			},
		})
		return
	}

	var req services.RiskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid request body",
			},
		})
		return
	}

	res, err := ctrl.service.Update(regID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "DB_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Risk updated successfully",
		"data":    res,
	})
}

func (ctrl *RiskController) DeleteRisk(c *gin.Context) {
	idStr := c.Param("id")
	regID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid risk ID format",
			},
		})
		return
	}

	err = ctrl.service.Delete(regID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "DB_ERROR",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Risk deleted successfully",
	})
}
