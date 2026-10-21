package controllers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"risk-service/models"
	"risk-service/pkg/kafka"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditUniverseController struct {
	db *gorm.DB
	// events may be nil (Kafka disabled or unreachable); *kafka.Publisher
	// treats a nil receiver as a no-op.
	events *kafka.Publisher
}

func NewAuditUniverseController(db *gorm.DB, events *kafka.Publisher) *AuditUniverseController {
	return &AuditUniverseController{db: db, events: events}
}

// publishEntity emits an entity event after the database write has committed.
// Publishing is best-effort: a broker error is logged, not returned, so the
// HTTP write still succeeds. POST /audit-universe/events/resync re-emits the
// full state to repair any gap.
func (ctrl *AuditUniverseController) publishEntity(eventType string, node models.CorporateAuditUniverse) {
	if err := ctrl.events.PublishEntity(context.Background(), eventType, entityData(node)); err != nil {
		log.Printf("audit universe event %s for %s not published: %v", eventType, node.ID, err)
	}
}

func entityData(node models.CorporateAuditUniverse) kafka.EntityData {
	return kafka.EntityData{
		ID:                      node.ID.String(),
		Name:                    node.Name,
		ParentID:                uuidPtrToString(node.ParentID),
		StandardAuditUniverseID: uuidPtrToString(node.StandardAuditUniverseID),
	}
}

func uuidPtrToString(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
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

type SyncMasterReq struct {
	Companies []string `json:"companies,omitempty"`
	Locations []string `json:"locations,omitempty"`
}

// SyncMasterData synchronizes master data companies and locations into Corporate Audit Universe
func (ctrl *AuditUniverseController) SyncMasterData(c *gin.Context) {
	var req SyncMasterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		// Proceed even if body is empty
	}

	now := time.Now()
	var createdNodes []models.CorporateAuditUniverse

	err := ctrl.db.Transaction(func(tx *gorm.DB) error {
		// 1. Find or create root node Subsidiaries/Business Unit
		var subParent models.CorporateAuditUniverse
		err := tx.Where("LOWER(name) = ? AND parent_id IS NULL", "subsidiaries/business unit").First(&subParent).Error
		if err != nil {
			err = tx.Where("name LIKE ? AND parent_id IS NULL", "%Subsidiaries%").First(&subParent).Error
		}
		if err == gorm.ErrRecordNotFound {
			subParent = models.CorporateAuditUniverse{
				ID:        uuid.New(),
				Name:      "Subsidiaries/Business Unit",
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := tx.Create(&subParent).Error; err != nil {
				return err
			}
			createdNodes = append(createdNodes, subParent)
		}

		// 2. Find or create root node Branches/Regional Offices
		var branchParent models.CorporateAuditUniverse
		err = tx.Where("LOWER(name) = ? AND parent_id IS NULL", "branches/regional offices").First(&branchParent).Error
		if err != nil {
			err = tx.Where("name LIKE ? AND parent_id IS NULL", "%Branches%").First(&branchParent).Error
		}
		if err == gorm.ErrRecordNotFound {
			branchParent = models.CorporateAuditUniverse{
				ID:        uuid.New(),
				Name:      "Branches/Regional Offices",
				CreatedAt: now,
				UpdatedAt: now,
			}
			if err := tx.Create(&branchParent).Error; err != nil {
				return err
			}
			createdNodes = append(createdNodes, branchParent)
		}

		// 3. Sync Companies under Subsidiaries/Business Unit
		for _, compName := range req.Companies {
			trimmed := strings.TrimSpace(compName)
			if trimmed == "" {
				continue
			}
			var existing models.CorporateAuditUniverse
			err := tx.Where("parent_id = ? AND LOWER(TRIM(name)) = ?", subParent.ID, strings.ToLower(trimmed)).First(&existing).Error
			if err == gorm.ErrRecordNotFound {
				newChild := models.CorporateAuditUniverse{
					ID:        uuid.New(),
					Name:      trimmed,
					ParentID:  &subParent.ID,
					CreatedAt: now,
					UpdatedAt: now,
				}
				if err := tx.Create(&newChild).Error; err != nil {
					return err
				}
				createdNodes = append(createdNodes, newChild)
			}
		}

		// 4. Sync Locations under Branches/Regional Offices
		for _, locName := range req.Locations {
			trimmed := strings.TrimSpace(locName)
			if trimmed == "" {
				continue
			}
			var existing models.CorporateAuditUniverse
			err := tx.Where("parent_id = ? AND LOWER(TRIM(name)) = ?", branchParent.ID, strings.ToLower(trimmed)).First(&existing).Error
			if err == gorm.ErrRecordNotFound {
				newChild := models.CorporateAuditUniverse{
					ID:        uuid.New(),
					Name:      trimmed,
					ParentID:  &branchParent.ID,
					CreatedAt: now,
					UpdatedAt: now,
				}
				if err := tx.Create(&newChild).Error; err != nil {
					return err
				}
				createdNodes = append(createdNodes, newChild)
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to synchronize master data: " + err.Error(),
		})
		return
	}

	for _, node := range createdNodes {
		ctrl.publishEntity(kafka.EventEntityCreated, node)
	}

	// Fetch updated corporate audit universe
	var corporate []models.CorporateAuditUniverse
	ctrl.db.Preload("Children.Children").Where("parent_id IS NULL").Find(&corporate)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Master data synchronized successfully with Corporate Audit Universe",
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
	var copiedChildren []models.CorporateAuditUniverse
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
					copiedChildren = append(copiedChildren, corpChild)
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

	if req.ID == "" {
		ctrl.publishEntity(kafka.EventEntityCreated, node)
		for _, child := range copiedChildren {
			ctrl.publishEntity(kafka.EventEntityCreated, child)
		}
	} else {
		ctrl.publishEntity(kafka.EventEntityUpdated, node)
	}

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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Corporate node not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to load corporate node: " + err.Error()})
		return
	}

	// Delete the node and ALL its descendants, with their yearly rows and
	// scores, in one transaction. Any error rolls everything back.
	// CorporateAuditUniverse and AuditUniverseYear have no DeletedAt, so this
	// is a hard delete: the yearly score history of the subtree goes with it.
	var deleted []models.CorporateAuditUniverse // deepest level first
	err = ctrl.db.Transaction(func(tx *gorm.DB) error {
		var derr error
		deleted, derr = deleteCorporateSubtree(tx, node)
		return derr
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to delete corporate node: " + err.Error(),
		})
		return
	}

	for _, n := range deleted {
		ctrl.publishEntity(kafka.EventEntityDeleted, n)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Corporate node and associated data deleted successfully",
	})
}

// maxCorporateDepth bounds the descendant walk so a parent_id cycle in bad data
// cannot loop forever.
const maxCorporateDepth = 64

// deleteCorporateSubtree deletes root, every descendant, their
// audit_universe_years rows and those rows' scores. It returns the deleted
// nodes ordered deepest level first, root last (the order their entity.deleted
// events are published in). Run it inside a transaction.
func deleteCorporateSubtree(tx *gorm.DB, root models.CorporateAuditUniverse) ([]models.CorporateAuditUniverse, error) {
	levels := [][]models.CorporateAuditUniverse{{root}}
	seen := map[uuid.UUID]bool{root.ID: true}
	frontier := []uuid.UUID{root.ID}
	for depth := 0; len(frontier) > 0; depth++ {
		if depth >= maxCorporateDepth {
			return nil, fmt.Errorf("corporate audit universe deeper than %d levels below %s (parent_id cycle?)", maxCorporateDepth, root.ID)
		}
		var children []models.CorporateAuditUniverse
		if err := tx.Where("parent_id IN ?", frontier).Find(&children).Error; err != nil {
			return nil, err
		}
		frontier = frontier[:0:0]
		var level []models.CorporateAuditUniverse
		for _, ch := range children {
			if seen[ch.ID] {
				continue
			}
			seen[ch.ID] = true
			level = append(level, ch)
			frontier = append(frontier, ch.ID)
		}
		if len(level) > 0 {
			levels = append(levels, level)
		}
	}

	ids := make([]uuid.UUID, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}

	yearIDs := tx.Model(&models.AuditUniverseYear{}).Select("id").Where("corporate_audit_universe_id IN ?", ids)
	if err := tx.Where("audit_universe_year_id IN (?)", yearIDs).Delete(&models.AuditUniverseRiskScore{}).Error; err != nil {
		return nil, err
	}
	if err := tx.Where("corporate_audit_universe_id IN ?", ids).Delete(&models.AuditUniverseYear{}).Error; err != nil {
		return nil, err
	}

	// Children before parents, so the parent_id foreign key never points at a
	// row that is already gone.
	ordered := make([]models.CorporateAuditUniverse, 0, len(ids))
	for i := len(levels) - 1; i >= 0; i-- {
		levelIDs := make([]uuid.UUID, 0, len(levels[i]))
		for _, n := range levels[i] {
			levelIDs = append(levelIDs, n.ID)
		}
		if err := tx.Where("id IN ?", levelIDs).Delete(&models.CorporateAuditUniverse{}).Error; err != nil {
			return nil, err
		}
		ordered = append(ordered, levels[i]...)
	}
	return ordered, nil
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
	var added, removed []models.AuditUniverseYear
	err := ctrl.db.Transaction(func(tx *gorm.DB) error {
		// 1. Fetch current established entities for this year
		var existing []models.AuditUniverseYear
		if err := tx.Where("year = ?", year).Find(&existing).Error; err != nil {
			return err
		}

		existingMap := make(map[string]models.AuditUniverseYear)
		for _, e := range existing {
			existingMap[e.CorporateAuditUniverseID.String()] = e
		}

		// 2. Insert new ones, keep existing ones
		keepIDs := make(map[string]bool)
		now := time.Now()

		for _, corpIDStr := range req.CorporateAuditUniverseIDs {
			corpID, err := uuid.Parse(corpIDStr)
			if err != nil {
				continue
			}

			if keepIDs[corpID.String()] {
				continue // duplicate id in the request
			}
			keepIDs[corpID.String()] = true

			if _, exists := existingMap[corpID.String()]; !exists {
				// A newly established entity is NOT scored: no score rows, no
				// index, no level and no priority until someone scores it.
				auYear := models.AuditUniverseYear{
					ID:                       uuid.New(),
					CorporateAuditUniverseID: corpID,
					Year:                     year,
					RiskIndex:                0,
					RiskLevel:                "",
					AuditPriority:            false,
					ScoredAt:                 nil,
					CreatedAt:                now,
					UpdatedAt:                now,
				}
				if err := tx.Create(&auYear).Error; err != nil {
					return err
				}
				added = append(added, auYear)
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
				removed = append(removed, existingYear)
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

	// An added row is announced with its (unscored) state through year.scored,
	// so consumers see risk_level "" for it; a removed row gets year.removed.
	publishYearScores(ctrl.events, added)
	for _, row := range removed {
		data := kafka.YearRemovedData{CorporateAuditUniverseID: row.CorporateAuditUniverseID.String(), Year: row.Year}
		if err := ctrl.events.PublishYearRemoved(context.Background(), data); err != nil {
			log.Printf("audit universe event %s for %s not published: %v", kafka.EventYearRemoved, row.CorporateAuditUniverseID, err)
		}
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

		// 2. Recalculate the row from all of its scores and the current weights
		// (shared with the recompute that follows a weight change).
		var allScores []models.AuditUniverseRiskScore
		if err := tx.Find(&allScores, "audit_universe_year_id = ?", auYearID).Error; err != nil {
			return err
		}
		weights, err := loadFactorWeights(tx)
		if err != nil {
			return err
		}

		auYear.ScoredAt = &now
		if _, err := rescoreYearRow(tx, &auYear, allScores, weights, now); err != nil {
			return err
		}
		auYear.UpdatedAt = now

		// 3. Update AuditUniverseYear
		return saveYearRowScore(tx, &auYear)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update scoring: " + err.Error(),
		})
		return
	}

	publishYearScores(ctrl.events, []models.AuditUniverseYear{auYear})

	// Fetch updated entity to return
	var updated models.AuditUniverseYear
	ctrl.db.Preload("CorporateAuditUniverse.StandardAuditUniverse").Preload("RiskScores.CorporateRiskFactor.StandardRiskFactor").First(&updated, "id = ?", auYearID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Risk scores updated and recalculated successfully",
		"data":    updated,
	})
}

func yearScoreData(auYear models.AuditUniverseYear) kafka.YearScoreData {
	return kafka.YearScoreData{
		CorporateAuditUniverseID: auYear.CorporateAuditUniverseID.String(),
		Year:                     auYear.Year,
		RiskIndex:                auYear.RiskIndex,
		RiskLevel:                auYear.RiskLevel,
		AuditPriority:            auYear.AuditPriority,
	}
}

// ResyncEvents queues the current Audit Universe state for publication: every
// corporate node (as entity.updated) and every yearly score. The relay publishes
// the queued events in the background. Consumers apply events idempotently, so
// a resync is safe to repeat. Use it after the first deploy of the event stream,
// or after a long broker outage, to repair links that were never published.
func (ctrl *AuditUniverseController) ResyncEvents(c *gin.Context) {
	if ctrl.events == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error":   "Kafka event publishing is disabled for this service",
		})
		return
	}

	var nodes []models.CorporateAuditUniverse
	if err := ctrl.db.Order("created_at ASC").Find(&nodes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to load corporate nodes: " + err.Error()})
		return
	}
	var years []models.AuditUniverseYear
	if err := ctrl.db.Order("year ASC").Find(&years).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to load yearly scores: " + err.Error()})
		return
	}

	ctx := c.Request.Context()
	failed := 0
	for _, node := range nodes {
		if err := ctrl.events.PublishEntity(ctx, kafka.EventEntityUpdated, entityData(node)); err != nil {
			failed++
		}
	}
	for _, y := range years {
		if err := ctrl.events.PublishYearScore(ctx, yearScoreData(y)); err != nil {
			failed++
		}
	}

	// 202: the events are queued in kafka_outbox_events, not yet on the broker.
	status := http.StatusAccepted
	if failed > 0 {
		status = http.StatusInternalServerError
	}
	c.JSON(status, gin.H{
		"success":     failed == 0,
		"entities":    len(nodes),
		"year_scores": len(years),
		"failed":      failed,
	})
}
