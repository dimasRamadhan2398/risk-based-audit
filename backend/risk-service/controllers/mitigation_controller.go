package controllers

import (
	"net/http"
	"risk-service/models"
	"risk-service/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MitigationController struct {
	service services.IMitigationService
}

func NewMitigationController(service services.IMitigationService) *MitigationController {
	return &MitigationController{service: service}
}

func (ctrl *MitigationController) ListMitigations(c *gin.Context) {
	riskIdQuery := c.Query("riskId")
	var parsedRiskId uuid.UUID
	var err error

	if riskIdQuery != "" {
		parsedRiskId, err = uuid.Parse(riskIdQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_ID",
					"message": "Invalid riskId query parameter format",
				},
			})
			return
		}
	}

	data, err := ctrl.service.GetAll(parsedRiskId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to query mitigations",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Mitigations fetched successfully",
		"data":    data,
	})
}

func (ctrl *MitigationController) CreateMitigation(c *gin.Context) {
	var req models.RiskMitigation
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
		"message": "Mitigation created successfully",
		"data":    res,
	})
}

func (ctrl *MitigationController) UpdateMitigation(c *gin.Context) {
	idStr := c.Param("id")
	mitID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid mitigation ID format",
			},
		})
		return
	}

	var req models.RiskMitigation
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

	res, err := ctrl.service.Update(mitID, &req)
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
		"message": "Mitigation updated successfully",
		"data":    res,
	})
}

func (ctrl *MitigationController) DeleteMitigation(c *gin.Context) {
	idStr := c.Param("id")
	mitID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid mitigation ID format",
			},
		})
		return
	}

	err = ctrl.service.Delete(mitID)
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
		"message": "Mitigation deleted successfully",
	})
}
