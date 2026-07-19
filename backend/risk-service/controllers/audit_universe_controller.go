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

type AuditUniverseController struct {
	db *gorm.DB
}

func NewAuditUniverseController(db *gorm.DB) *AuditUniverseController {
	return &AuditUniverseController{db: db}
}

// ListStandardAuditUniverse returns the standard audit universe library
func (ctrl *AuditUniverseController) ListStandardAuditUniverse(c *gin.Context) {
	var standard []models.StandardAuditUniverse
	// Fetch root nodes and preload children recursive
	if err := ctrl.db.Preload("Children.Children").Where("parent_id IS NULL").Find(&standard).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch standard library: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    standard,
	})
}

// ListCorporateAuditUniverse returns the corporate audit universe library
func (ctrl *AuditUniverseController) ListCorporateAuditUniverse(c *gin.Context) {
	var corporate []models.CorporateAuditUniverse
	if err := ctrl.db.Preload("Children.Children").Where("parent_id IS NULL").Find(&corporate).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch corporate library: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    corporate,
	})
}

type CorporateNodeReq struct {
	ID                      string `json:"id,omitempty"` // empty for new nodes
	StandardAuditUniverseID string `json:"standard_audit_universe_id,omitempty"`
	Name                    string `json:"name" binding:"required"`
	ParentID                string `json:"parent_id,omitempty"`
}

// CreateOrUpdateCorporateNode creates a new custom corporate node or updates an existing one
func (ctrl *AuditUniverseController) CreateOrUpdateCorporateNode(c *gin.Context) {
	var req CorporateNodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	var node models.CorporateAuditUniverse
	now := time.Now()

	if req.ID != "" {
		nodeID, err := uuid.Parse(req.ID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID format"})
			return
		}
		if err := ctrl.db.First(&node, "id = ?", nodeID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Corporate node not found"})
			return
		}
		node.Name = req.Name
		node.UpdatedAt = now
	} else {
		node.ID = uuid.New()
		node.Name = req.Name
		node.CreatedAt = now
		node.UpdatedAt = now

		if req.StandardAuditUniverseID != "" {
			stdID, err := uuid.Parse(req.StandardAuditUniverseID)
			if err == nil {
				node.StandardAuditUniverseID = &stdID
			}
		}

		if req.ParentID != "" {
			parentID, err := uuid.Parse(req.ParentID)
			if err == nil {
				node.ParentID = &parentID
			}
		}
	}

	// Perform saving inside a transaction to copy standard children if it is a parent
	err := ctrl.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&node).Error; err != nil {
			return err
		}

		// If it's a new root node (parent) with standard ID, auto-copy its standard child sub-entities
		if req.ID == "" && node.StandardAuditUniverseID != nil && node.ParentID == nil {
			var stdChildren []models.StandardAuditUniverse
			if err := tx.Where("parent_id = ?", *node.StandardAuditUniverseID).Find(&stdChildren).Error; err == nil {
				for _, sc := range stdChildren {
					childStdID := sc.ID
					corpChild := models.CorporateAuditUniverse{
						ID:                      uuid.New(),
						StandardAuditUniverseID: &childStdID,
						Name:                    sc.Name,
						ParentID:                &node.ID,
						CreatedAt:               now,
						UpdatedAt:               now,
					}
					if err := tx.Create(&corpChild).Error; err != nil {
						return err
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to save corporate node: " + err.Error(),
		})
		return
	}

	// Fetch with standard preloads if it has one
	ctrl.db.Preload("StandardAuditUniverse").First(&node, "id = ?", node.ID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Corporate node saved successfully",
		"data":    node,
	})
}

// DeleteCorporateNode deletes a corporate node and its children
func (ctrl *AuditUniverseController) DeleteCorporateNode(c *gin.Context) {
	idStr := c.Param("id")
	nodeID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID format"})
		return
	}

	// Fetch node to ensure it exists
	var node models.CorporateAuditUniverse
	if err := ctrl.db.First(&node, "id = ?", nodeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Corporate node not found"})
		return
	}

	// Delete recursively (in transaction)
	err = ctrl.db.Transaction(func(tx *gorm.DB) error {
		// Delete child scores first, if any
		var childYears []models.AuditUniverseYear
		tx.Where("corporate_audit_universe_id = ? OR corporate_audit_universe_id IN (SELECT id FROM corporate_audit_universe WHERE parent_id = ?)", nodeID, nodeID).Find(&childYears)
		for _, cy := range childYears {
			tx.Where("audit_universe_year_id = ?", cy.ID).Delete(&models.AuditUniverseRiskScore{})
		}
		tx.Where("corporate_audit_universe_id = ? OR corporate_audit_universe_id IN (SELECT id FROM corporate_audit_universe WHERE parent_id = ?)", nodeID, nodeID).Delete(&models.AuditUniverseYear{})

		// Delete corporate nodes
		tx.Where("parent_id = ?", nodeID).Delete(&models.CorporateAuditUniverse{})
		tx.Delete(&node)
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete corporate node: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Corporate node and associated data deleted successfully",
	})
}

// ListYearlyAuditUniverse returns the audit universe established for a specific year
func (ctrl *AuditUniverseController) ListYearlyAuditUniverse(c *gin.Context) {
	yearStr := c.Param("year")
	var year int
	if _, err := fmt.Sscanf(yearStr, "%d", &year); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid year format"})
		return
	}

	var universe []models.AuditUniverseYear
	if err := ctrl.db.Preload("CorporateAuditUniverse.StandardAuditUniverse").Preload("RiskScores.CorporateRiskFactor.StandardRiskFactor").Where("year = ?", year).Find(&universe).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch yearly audit universe: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    universe,
	})
}

type EstablishUniverseReq struct {
	CorporateAuditUniverseIDs []string `json:"corporate_audit_universe_ids" binding:"required"`
}

// EstablishYearlyUniverse sets which corporate auditable entities are active for a specific year
func (ctrl *AuditUniverseController) EstablishYearlyUniverse(c *gin.Context) {
	yearStr := c.Param("year")
	var year int
	if _, err := fmt.Sscanf(yearStr, "%d", &year); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid year format"})
		return
	}

	var req EstablishUniverseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	// Perform updates in transaction
	err := ctrl.db.Transaction(func(tx *gorm.DB) error {
		// 1. Fetch current established entities for this year
		var existing []models.AuditUniverseYear
		tx.Where("year = ?", year).Find(&existing)

		existingMap := make(map[string]models.AuditUniverseYear)
		for _, e := range existing {
			existingMap[e.CorporateAuditUniverseID.String()] = e
		}

		// 2. Insert new ones, keep existing ones
		keepIDs := make(map[string]bool)
		now := time.Now()

		// Get corporate risk factors to seed default scores (3) for newly established entities
		var corporateRiskFactors []models.CorporateRiskFactor
		tx.Find(&corporateRiskFactors)

		for _, corpIDStr := range req.CorporateAuditUniverseIDs {
			corpID, err := uuid.Parse(corpIDStr)
			if err != nil {
				continue
			}

			keepIDs[corpID.String()] = true

			if _, exists := existingMap[corpID.String()]; !exists {
				// Create new AuditUniverseYear
				auYear := models.AuditUniverseYear{
					ID:                       uuid.New(),
					CorporateAuditUniverseID: corpID,
					Year:                     year,
					RiskIndex:                60.0, // Default risk index matching default score 3 (TotalWeightedScore=3.0, 3.0/5.0*100 = 60.0)
					RiskLevel:                "Medium to High",
					AuditPriority:            true,
					CreatedAt:                now,
					UpdatedAt:                now,
				}
				if err := tx.Create(&auYear).Error; err != nil {
					return err
				}

				// Create default scores of 3
				for _, rf := range corporateRiskFactors {
					sc := models.AuditUniverseRiskScore{
						ID:                    uuid.New(),
						AuditUniverseYearID:   auYear.ID,
						CorporateRiskFactorID: rf.ID,
						Score:                 3,
						WeightedScore:         rf.Weight * 3.0,
						CreatedAt:             now,
						UpdatedAt:             now,
					}
					if err := tx.Create(&sc).Error; err != nil {
						return err
					}
				}
			}
		}

		// 3. Delete ones that are no longer selected
		for corpIDStr, existingYear := range existingMap {
			if !keepIDs[corpIDStr] {
				// Delete scores first
				if err := tx.Where("audit_universe_year_id = ?", existingYear.ID).Delete(&models.AuditUniverseRiskScore{}).Error; err != nil {
					return err
				}
				if err := tx.Delete(&existingYear).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to establish yearly universe: " + err.Error(),
		})
		return
	}

	// Fetch and return the updated yearly universe
	var universe []models.AuditUniverseYear
	ctrl.db.Preload("CorporateAuditUniverse.StandardAuditUniverse").Preload("RiskScores.CorporateRiskFactor.StandardRiskFactor").Where("year = ?", year).Find(&universe)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Audit universe for year " + yearStr + " established successfully",
		"data":    universe,
	})
}

type ScoreInput struct {
	CorporateRiskFactorID string `json:"corporate_risk_factor_id" binding:"required"`
	Score                 int    `json:"score" binding:"required"` // 1 to 5
}

type ScoreUniverseReq struct {
	AuditUniverseYearID string       `json:"audit_universe_year_id" binding:"required"`
	Scores              []ScoreInput `json:"scores" binding:"required"`
}

// ScoreYearlyEntity updates risk scores for a yearly established entity and recalculates its index and level
func (ctrl *AuditUniverseController) ScoreYearlyEntity(c *gin.Context) {
	yearStr := c.Param("year")
	var year int
	if _, err := fmt.Sscanf(yearStr, "%d", &year); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid year format"})
		return
	}

	var req ScoreUniverseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body: " + err.Error(),
		})
		return
	}

	auYearID, err := uuid.Parse(req.AuditUniverseYearID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid audit universe year ID"})
		return
	}

	var auYear models.AuditUniverseYear
	if err := ctrl.db.First(&auYear, "id = ?", auYearID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Yearly entity entry not found"})
		return
	}

	// Perform calculations and updates in transaction
	err = ctrl.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()

		// 1. Process and save each score
		for _, scoreIn := range req.Scores {
			rfID, err := uuid.Parse(scoreIn.CorporateRiskFactorID)
			if err != nil {
				continue
			}

			// Validate score is between 1 and 5
			scoreVal := scoreIn.Score
			if scoreVal < 1 {
				scoreVal = 1
			} else if scoreVal > 5 {
				scoreVal = 5
			}

			// Get corporate risk factor weight
			var rf models.CorporateRiskFactor
			if err := tx.First(&rf, "id = ?", rfID).Error; err != nil {
				return err
			}

			weightedScore := rf.Weight * float64(scoreVal)

			// Find or create Risk Score entry
			var sc models.AuditUniverseRiskScore
			err = tx.First(&sc, "audit_universe_year_id = ? AND corporate_risk_factor_id = ?", auYearID, rfID).Error
			if err == gorm.ErrRecordNotFound {
				sc = models.AuditUniverseRiskScore{
					ID:                    uuid.New(),
					AuditUniverseYearID:   auYearID,
					CorporateRiskFactorID: rfID,
					Score:                 scoreVal,
					WeightedScore:         weightedScore,
					CreatedAt:             now,
					UpdatedAt:             now,
				}
				if err := tx.Create(&sc).Error; err != nil {
					return err
				}
			} else if err == nil {
				sc.Score = scoreVal
				sc.WeightedScore = weightedScore
				sc.UpdatedAt = now
				if err := tx.Save(&sc).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}

		// 2. Fetch all scores for this yearly entity to calculate total weighted score
		var allScores []models.AuditUniverseRiskScore
		if err := tx.Find(&allScores, "audit_universe_year_id = ?", auYearID).Error; err != nil {
			return err
		}

		var totalWeighted float64
		for _, sc := range allScores {
			totalWeighted += sc.WeightedScore
		}

		// Risk Index = (Total Weighted Score / 5.0) * 100%
		riskIndex := (totalWeighted / 5.0) * 100.0
		var riskLevel string
		var priority bool

		if riskIndex >= 80.0 {
			riskLevel = "High"
			priority = true
		} else if riskIndex >= 60.0 {
			riskLevel = "Medium to High"
			priority = true
		} else if riskIndex >= 40.0 {
			riskLevel = "Medium"
			priority = false
		} else if riskIndex >= 20.0 {
			riskLevel = "Low to Medium"
			priority = false
		} else {
			riskLevel = "Low"
			priority = false
		}

		// 3. Update AuditUniverseYear
		auYear.RiskIndex = riskIndex
		auYear.RiskLevel = riskLevel
		auYear.AuditPriority = priority
		auYear.UpdatedAt = now

		if err := tx.Save(&auYear).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update scoring: " + err.Error(),
		})
		return
	}

	// Fetch updated entity to return
	var updated models.AuditUniverseYear
	ctrl.db.Preload("CorporateAuditUniverse.StandardAuditUniverse").Preload("RiskScores.CorporateRiskFactor.StandardRiskFactor").First(&updated, "id = ?", auYearID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Risk scores updated and recalculated successfully",
		"data":    updated,
	})
}
