package controllers

import (
	"fmt"
	"net/http"
	"risk-service/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RiskFactorController struct {
	db *gorm.DB
}

func NewRiskFactorController(db *gorm.DB) *RiskFactorController {
	return &RiskFactorController{db: db}
}

// ListStandardRiskFactors returns all standard risk factors from library
func (ctrl *RiskFactorController) ListStandardRiskFactors(c *gin.Context) {
	var factors []models.StandardRiskFactor
	if err := ctrl.db.Order("id asc").Find(&factors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch standard risk factors: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    factors,
	})
}

// ListCorporateRiskFactors returns all user selected corporate risk factors with weights
func (ctrl *RiskFactorController) ListCorporateRiskFactors(c *gin.Context) {
	var factors []models.CorporateRiskFactor
	if err := ctrl.db.Preload("StandardRiskFactor").Order("weight desc").Find(&factors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch corporate risk factors: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    factors,
	})
}

type CorporateRiskFactorReq struct {
	StandardRiskFactorID string  `json:"standard_risk_factor_id" binding:"required"`
	Weight               float64 `json:"weight" binding:"required"` // weight in percent (e.g. 15.0)
}

// UpdateCorporateRiskFactors replaces all corporate risk factors and weights
func (ctrl *RiskFactorController) UpdateCorporateRiskFactors(c *gin.Context) {
	var reqs []CorporateRiskFactorReq
	if err := c.ShouldBindJSON(&reqs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	// Validate weights sum to 100%
	var totalWeight float64
	for _, req := range reqs {
		totalWeight += req.Weight
	}

	// Allow slight rounding tolerance (e.g., 99.9% to 100.1%)
	if totalWeight < 99.9 || totalWeight > 100.1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Total weight must equal exactly 100% (currently: " + fmt.Sprintf("%.2f", totalWeight) + "%)",
		})
		return
	}

	// Perform database updates inside transaction
	err := ctrl.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		
		// Map requested factors for fast lookup
		reqMap := make(map[uuid.UUID]float64)
		for _, r := range reqs {
			sfID, err := uuid.Parse(r.StandardRiskFactorID)
			if err == nil {
				reqMap[sfID] = r.Weight / 100.0 // convert e.g., 15.0 to 0.15
			}
		}

		// Fetch existing corporate risk factors
		var existing []models.CorporateRiskFactor
		if err := tx.Find(&existing).Error; err != nil {
			return err
		}

		existingMap := make(map[uuid.UUID]models.CorporateRiskFactor)
		for _, e := range existing {
			existingMap[e.StandardRiskFactorID] = e
		}

		// 1. Delete ones that are no longer requested
		for sfID, ext := range existingMap {
			if _, ok := reqMap[sfID]; !ok {
				// Delete dependent risk scores first to prevent FK violation
				if err := tx.Where("corporate_risk_factor_id = ?", ext.ID).Delete(&models.AuditUniverseRiskScore{}).Error; err != nil {
					return err
				}
				if err := tx.Unscoped().Delete(&ext).Error; err != nil {
					return err
				}
			}
		}

		// 2. Insert or Update requested ones
		for sfID, weight := range reqMap {
			if ext, exists := existingMap[sfID]; exists {
				ext.Weight = weight
				ext.UpdatedAt = now
				if err := tx.Save(&ext).Error; err != nil {
					return err
				}
			} else {
				cf := models.CorporateRiskFactor{
					ID:                   uuid.New(),
					StandardRiskFactorID: sfID,
					Weight:               weight,
					CreatedAt:            now,
					UpdatedAt:            now,
				}
				if err := tx.Create(&cf).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update corporate risk factors: " + err.Error(),
		})
		return
	}

	// Fetch updated list to return
	var updatedFactors []models.CorporateRiskFactor
	ctrl.db.Preload("StandardRiskFactor").Order("weight desc").Find(&updatedFactors)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Corporate risk factors updated successfully",
		"data":    updatedFactors,
	})
}
