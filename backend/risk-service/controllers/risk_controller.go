package controllers

import (
	"context"
	"errors"
	"net/http"

	"risk-service/pkg/masterclient"
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
	data, err := ctrl.service.GetAll(requestContext(c))
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

	res, err := ctrl.service.Create(requestContext(c), &req)
	if err != nil {
		respondWriteError(c, err)
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

	res, err := ctrl.service.Update(requestContext(c), regID, &req)
	if err != nil {
		respondWriteError(c, err)
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

// requestContext forwards the caller's Authorization header to master-service
// lookups made while serving this request.
func requestContext(c *gin.Context) context.Context {
	return masterclient.WithAuthorization(c.Request.Context(), c.GetHeader("Authorization"))
}

// respondWriteError maps create/update failures: an unregistered location is
// the client's fault (400), an unreachable Location master is 503, anything
// else stays a DB error.
func respondWriteError(c *gin.Context, err error) {
	var locErr *services.LocationError
	switch {
	case errors.As(err, &locErr):
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    locErr.Code,
				"message": locErr.Message,
			},
		})
	case errors.Is(err, services.ErrLocationsUnavailable):
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "LOCATION_MASTER_UNAVAILABLE",
				"message": "Location master data is unavailable; try again shortly",
			},
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "DB_ERROR",
				"message": err.Error(),
			},
		})
	}
}
