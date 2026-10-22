package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"risk-service/models"
	"risk-service/pkg/kafka"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RiskFactorController struct {
	db *gorm.DB
	// events may be nil (Kafka disabled); *kafka.Publisher treats a nil
	// receiver as a no-op. Weight changes re-score year rows and announce them
	// with year.scored, like ScoreYearlyEntity.
	events *kafka.Publisher
}

func NewRiskFactorController(db *gorm.DB, events *kafka.Publisher) *RiskFactorController {
	return &RiskFactorController{db: db, events: events}
}

// errFactorHasWeight is returned from the delete transaction when the standard
// factor is still weighted in the corporate risk factors.
var errFactorHasWeight = errors.New("standard risk factor still has a corporate weight")

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

	var rescored []models.AuditUniverseYear
	var weight float64
	err = ctrl.db.Transaction(func(tx *gorm.DB) error {
		var corpFactors []models.CorporateRiskFactor
		if err := tx.Where("standard_risk_factor_id = ?", factorID).Find(&corpFactors).Error; err != nil {
			return err
		}

		// Deleting a weighted factor would leave the corporate weights under
		// 100%. The user must first save the corporate weights without it.
		for _, cf := range corpFactors {
			if cf.Weight > 0 {
				weight = cf.Weight
				return errFactorHasWeight
			}
		}

		// Rows that lose scores here must be re-derived even if none remain.
		cfIDs := make([]uuid.UUID, 0, len(corpFactors))
		for _, cf := range corpFactors {
			cfIDs = append(cfIDs, cf.ID)
		}
		lostScores, err := yearRowsScoredOn(tx, cfIDs)
		if err != nil {
			return err
		}

		// 1. Delete dependent scores for any (zero-weight) corporate risk factors linked to this standard factor
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

		// 4. Scores were removed: re-derive the scored year rows.
		if len(corpFactors) > 0 {
			var err error
			rescored, err = recomputeScoredYears(tx, time.Now(), lostScores)
			return err
		}
		return nil
	})

	if errors.Is(err, errFactorHasWeight) {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"code":    "FACTOR_HAS_CORPORATE_WEIGHT",
			"error": fmt.Sprintf("Risk factor %q still has a corporate weight of %s%%. Remove it from the corporate risk factors (and save the weights) before deleting it.",
				factor.Name, strconv.FormatFloat(math.Round(weight*10000)/100, 'f', -1, 64)),
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete standard risk factor: " + err.Error(),
		})
		return
	}

	publishYearScores(ctrl.events, rescored)

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
	var rescored []models.AuditUniverseYear
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

		// Rows scored on a factor that is being removed must be re-derived
		// even if they lose their last score (see recomputeScoredYears).
		var removedIDs []uuid.UUID
		for sfID, ext := range existingMap {
			if _, ok := reqMap[sfID]; !ok {
				removedIDs = append(removedIDs, ext.ID)
			}
		}
		lostScores, err := yearRowsScoredOn(tx, removedIDs)
		if err != nil {
			return err
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

		// 3. Weights or the factor set changed: re-derive weighted scores and
		// index/level/priority of every scored year row in this transaction.
		// A newly added factor has no score yet and is "not scored" for each
		// entity (see models.ComputeRiskIndex).
		rescored, err = recomputeScoredYears(tx, now, lostScores)
		return err
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update corporate risk factors: " + err.Error(),
		})
		return
	}

	publishYearScores(ctrl.events, rescored)

	// Fetch updated list to return
	var updatedFactors []models.CorporateRiskFactor
	ctrl.db.Preload("StandardRiskFactor").Order("weight desc").Find(&updatedFactors)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Corporate risk factors updated successfully",
		"data":    updatedFactors,
	})
}
