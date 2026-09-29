package controllers

import (
	"encoding/json"
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

type StandardRiskFactorReq struct {
	Name            string `json:"name" binding:"required"`
	Description     string `json:"description"`
	ScoreGuidelines string `json:"score_guidelines"`
}

// ListStandardRiskFactors returns all standard risk factors from library
func (ctrl *RiskFactorController) ListStandardRiskFactors(c *gin.Context) {
	var factors []models.StandardRiskFactor
	if err := ctrl.db.Order("name asc").Find(&factors).Error; err != nil {
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

// CreateStandardRiskFactor adds a new standard risk factor to the library
func (ctrl *RiskFactorController) CreateStandardRiskFactor(c *gin.Context) {
	var req StandardRiskFactorReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Name is required",
		})
		return
	}

	guidelines := req.ScoreGuidelines
	if guidelines == "" {
		defaultGuidelines := []struct {
			Score int    `json:"score"`
			Desc  string `json:"desc"`
		}{
			{Score: 5, Desc: "High – Major contributor to enterprise risk"},
			{Score: 4, Desc: "Medium to High – Significant risk affecting key operations"},
			{Score: 3, Desc: "Medium – Moderate risk exposure with limited enterprise impact"},
			{Score: 2, Desc: "Low to Medium – Low risk operations with minimal impact"},
			{Score: 1, Desc: "Low – Administrative or routine activity with minimal risk"},
		}
		b, _ := json.Marshal(defaultGuidelines)
		guidelines = string(b)
	}

	now := time.Now()
	factor := models.StandardRiskFactor{
		ID:              uuid.New(),
		Name:            req.Name,
		Description:     req.Description,
		ScoreGuidelines: guidelines,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := ctrl.db.Create(&factor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create standard risk factor: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Standard risk factor created successfully",
		"data":    factor,
	})
}

// UpdateStandardRiskFactor updates an existing standard risk factor
func (ctrl *RiskFactorController) UpdateStandardRiskFactor(c *gin.Context) {
	idStr := c.Param("id")
	factorID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid factor ID format",
		})
		return
	}

	var req StandardRiskFactorReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	var factor models.StandardRiskFactor
	if err := ctrl.db.First(&factor, "id = ?", factorID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Standard risk factor not found",
		})
		return
	}

	factor.Name = req.Name
	factor.Description = req.Description
	if req.ScoreGuidelines != "" {
		factor.ScoreGuidelines = req.ScoreGuidelines
	}
	factor.UpdatedAt = time.Now()

	if err := ctrl.db.Save(&factor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update standard risk factor: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Standard risk factor updated successfully",
		"data":    factor,
	})
}

// DeleteStandardRiskFactor deletes a standard risk factor and cleans up associated corporate weights
func (ctrl *RiskFactorController) DeleteStandardRiskFactor(c *gin.Context) {
	idStr := c.Param("id")
	factorID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid factor ID format",
		})
		return
	}

	var factor models.StandardRiskFactor
	if err := ctrl.db.First(&factor, "id = ?", factorID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Standard risk factor not found",
		})
		return
	}

	err = ctrl.db.Transaction(func(tx *gorm.DB) error {
		// 1. Delete dependent scores for any corporate risk factors linked to this standard factor
		var corpFactors []models.CorporateRiskFactor
		if err := tx.Where("standard_risk_factor_id = ?", factorID).Find(&corpFactors).Error; err != nil {
			return err
		}

		for _, cf := range corpFactors {
			if err := tx.Where("corporate_risk_factor_id = ?", cf.ID).Delete(&models.AuditUniverseRiskScore{}).Error; err != nil {
				return err
			}
		}

		// 2. Delete corporate risk factors linked to this factor
		if err := tx.Where("standard_risk_factor_id = ?", factorID).Delete(&models.CorporateRiskFactor{}).Error; err != nil {
			return err
		}

		// 3. Delete the standard risk factor
		if err := tx.Delete(&factor).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete standard risk factor: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Standard risk factor deleted successfully",
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
