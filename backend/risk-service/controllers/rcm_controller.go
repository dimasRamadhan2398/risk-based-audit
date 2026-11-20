package controllers

import (
	"errors"
	"math"
	"net/http"
	"time"

	"risk-service/models"
	"risk-service/pkg/masterclient"
	"risk-service/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RCMController struct {
	db *gorm.DB
	// departments validates department_id on write (nil: department_id
	// writes are refused as unavailable).
	departments masterclient.DepartmentLookup
	// locations narrows /rcm/coso-signals by branch (nil: legacy branch names
	// cannot be matched; stored location ids still are).
	locations masterclient.LocationSource
	now       func() time.Time
}

func NewRCMController(db *gorm.DB, departments masterclient.DepartmentLookup, locations masterclient.LocationSource) *RCMController {
	return &RCMController{db: db, departments: departments, locations: locations, now: time.Now}
}

// Get Effectiveness Rating & Interpretation based on score percentage
func GetEffectivenessRating(score float64) (string, string) {
	switch {
	case score >= 90.0:
		return "Highly Effective", "Controls reliably mitigate risk and require only routine monitoring."
	case score >= 80.0:
		return "Effective", "Controls function well; only minor improvements are recommended."
	case score >= 70.0:
		return "Moderately Effective", "Some weaknesses exist; corrective actions should be planned."
	case score >= 60.0:
		return "Weak", "Significant improvements are needed to reduce risk adequately."
	default:
		return "Ineffective", "Controls do not provide sufficient risk mitigation and require immediate attention."
	}
}

type RCMRequest struct {
	ID                           *string `json:"id,omitempty"`
	RiskID                       *string `json:"risk_id,omitempty"`
	RiskCode                     string  `json:"risk_code"`
	RiskEvent                    string  `json:"risk_event"`
	ControlCode                  string  `json:"control_code"`
	ControlDescription           string  `json:"control_description"`
	ControlType                  string  `json:"control_type"`
	ControlOwner                 string  `json:"control_owner"`
	Department                   string  `json:"department"`
	Year                         int     `json:"year"`
	DesignEffectivenessWeight    float64 `json:"design_effectiveness_weight"`
	DesignEffectivenessRating    int     `json:"design_effectiveness_rating"`
	OperatingEffectivenessWeight float64 `json:"operating_effectiveness_weight"`
	OperatingEffectivenessRating int     `json:"operating_effectiveness_rating"`
	CoverageCompletenessWeight   float64 `json:"coverage_completeness_weight"`
	CoverageCompletenessRating   int     `json:"coverage_completeness_rating"`
	TimelinessWeight             float64 `json:"timeliness_weight"`
	TimelinessRating             int     `json:"timeliness_rating"`
	AutomationMonitoringWeight   float64 `json:"automation_monitoring_weight"`
	AutomationMonitoringRating   int     `json:"automation_monitoring_rating"`
	TotalWeightedScore           float64 `json:"total_weighted_score"`
	InherentRisk                 int     `json:"inherent_risk"`
	ResidualRisk                 int     `json:"residual_risk"`
	Notes                        string  `json:"notes"`

	// COSO 2013 mapping and master department. Omitted = unchanged on PUT;
	// department_id / coso_component_code null = clear; coso_principle_codes
	// [] = clear (null is treated as omitted).
	DepartmentID optionalString `json:"department_id"`
	// location_id: the row's own branch (Location master). Omitted =
	// unchanged on PUT; null/"" = follow the linked risk's branch.
	LocationID         optionalString `json:"location_id"`
	CosoComponentCode  optionalString `json:"coso_component_code"`
	CosoPrincipleCodes *[]string      `json:"coso_principle_codes"`
}

func parseRCMRequest(req *RCMRequest, target *models.RiskControlMatrix) {
	if req.RiskID != nil && *req.RiskID != "" {
		if parsed, err := uuid.Parse(*req.RiskID); err == nil {
			target.RiskID = &parsed
		} else {
			target.RiskID = nil
		}
	} else {
		target.RiskID = nil
	}

	target.RiskCode = req.RiskCode
	target.RiskEvent = req.RiskEvent
	target.ControlCode = req.ControlCode
	target.ControlDescription = req.ControlDescription
	if req.ControlType != "" {
		target.ControlType = req.ControlType
	} else if target.ControlType == "" {
		target.ControlType = "Preventive"
	}
	target.ControlOwner = req.ControlOwner
	target.Department = req.Department
	if req.Year != 0 {
		target.Year = req.Year
	} else if target.Year == 0 {
		target.Year = time.Now().Year()
	}

	if req.DesignEffectivenessWeight > 0 {
		target.DesignEffectivenessWeight = req.DesignEffectivenessWeight
	} else if target.DesignEffectivenessWeight <= 0 {
		target.DesignEffectivenessWeight = 20.0
	}
	if req.DesignEffectivenessRating > 0 {
		target.DesignEffectivenessRating = req.DesignEffectivenessRating
	} else if target.DesignEffectivenessRating <= 0 {
		target.DesignEffectivenessRating = 3
	}

	if req.OperatingEffectivenessWeight > 0 {
		target.OperatingEffectivenessWeight = req.OperatingEffectivenessWeight
	} else if target.OperatingEffectivenessWeight <= 0 {
		target.OperatingEffectivenessWeight = 20.0
	}
	if req.OperatingEffectivenessRating > 0 {
		target.OperatingEffectivenessRating = req.OperatingEffectivenessRating
	} else if target.OperatingEffectivenessRating <= 0 {
		target.OperatingEffectivenessRating = 3
	}

	if req.CoverageCompletenessWeight > 0 {
		target.CoverageCompletenessWeight = req.CoverageCompletenessWeight
	} else if target.CoverageCompletenessWeight <= 0 {
		target.CoverageCompletenessWeight = 20.0
	}
	if req.CoverageCompletenessRating > 0 {
		target.CoverageCompletenessRating = req.CoverageCompletenessRating
	} else if target.CoverageCompletenessRating <= 0 {
		target.CoverageCompletenessRating = 3
	}

	if req.TimelinessWeight > 0 {
		target.TimelinessWeight = req.TimelinessWeight
	} else if target.TimelinessWeight <= 0 {
		target.TimelinessWeight = 20.0
	}
	if req.TimelinessRating > 0 {
		target.TimelinessRating = req.TimelinessRating
	} else if target.TimelinessRating <= 0 {
		target.TimelinessRating = 3
	}

	if req.AutomationMonitoringWeight > 0 {
		target.AutomationMonitoringWeight = req.AutomationMonitoringWeight
	} else if target.AutomationMonitoringWeight <= 0 {
		target.AutomationMonitoringWeight = 20.0
	}
	if req.AutomationMonitoringRating > 0 {
		target.AutomationMonitoringRating = req.AutomationMonitoringRating
	} else if target.AutomationMonitoringRating <= 0 {
		target.AutomationMonitoringRating = 3
	}

	if req.InherentRisk > 0 {
		target.InherentRisk = req.InherentRisk
	} else if target.InherentRisk <= 0 {
		target.InherentRisk = 20
	}
	if req.ResidualRisk > 0 {
		target.ResidualRisk = req.ResidualRisk
	} else if target.ResidualRisk <= 0 {
		target.ResidualRisk = 8
	}

	target.Notes = req.Notes
	target.CalculateWeightedScore()
}

func (ctrl *RCMController) ListRCM(c *gin.Context) {
	db := ctrl.db.WithContext(c.Request.Context())
	query, fields, err := applyRCMFilters(c, db, db.Model(&models.RiskControlMatrix{}).Order("created_at desc"))
	if err != nil {
		respondDB(c, "Failed to list RCM items", err)
		return
	}
	if fields != nil {
		respondValidation(c, fields)
		return
	}

	var list []models.RiskControlMatrix
	if err := query.Find(&list).Error; err != nil {
		respondDB(c, "Failed to list RCM items", err)
		return
	}
	if list == nil {
		list = []models.RiskControlMatrix{}
	}
	if err := attachTags(db, list); err != nil {
		respondDB(c, "Failed to list RCM items", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "RCM items retrieved successfully",
		"data":    list,
	})
}

func (ctrl *RCMController) CreateRCM(c *gin.Context) {
	var req RCMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid request body: " + err.Error(),
			},
		})
		return
	}

	var item models.RiskControlMatrix
	if req.ID != nil && *req.ID != "" {
		if parsedID, err := uuid.Parse(*req.ID); err == nil {
			item.ID = parsedID
		} else {
			item.ID = uuid.New()
		}
	} else {
		item.ID = uuid.New()
	}

	ctx := masterclient.WithAuthorization(c.Request.Context(), c.GetHeader("Authorization"))
	w, fields, err := ctrl.resolveCosoWrite(ctx, &req, &item, nil)
	if !ctrl.handleCosoResolveError(c, fields, err) {
		return
	}

	parseRCMRequest(&req, &item)
	w.apply(&item, nil, "")

	err = ctrl.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit(clause.Associations).Create(&item).Error; err != nil {
			return err
		}
		return writeTags(tx, item.ID, w.principleIDs, false)
	})
	if err != nil {
		respondDB(c, "Failed to create RCM item", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "RCM item created successfully",
		"data":    item,
	})
}

func (ctrl *RCMController) UpdateRCM(c *gin.Context) {
	idStr := c.Param("id")
	itemUUID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid RCM ID format",
			},
		})
		return
	}

	var req RCMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "BAD_REQUEST",
				"message": "Invalid request body: " + err.Error(),
			},
		})
		return
	}

	db := ctrl.db.WithContext(c.Request.Context())
	var existing models.RiskControlMatrix
	if err := db.First(&existing, "id = ?", itemUUID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "RCM item not found",
			},
		})
		return
	}
	tags, err := services.LoadRCMPrincipleCodes(db, []uuid.UUID{existing.ID})
	if err != nil {
		respondDB(c, "Failed to update RCM item", err)
		return
	}

	ctx := masterclient.WithAuthorization(c.Request.Context(), c.GetHeader("Authorization"))
	w, fields, err := ctrl.resolveCosoWrite(ctx, &req, &existing, tags[existing.ID])
	if !ctrl.handleCosoResolveError(c, fields, err) {
		return
	}

	prevDeptID, prevDept := existing.DepartmentID, existing.Department
	parseRCMRequest(&req, &existing)
	w.apply(&existing, prevDeptID, prevDept)

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit(clause.Associations).Save(&existing).Error; err != nil {
			return err
		}
		if !w.tagsChanged {
			return nil
		}
		return writeTags(tx, existing.ID, w.principleIDs, true)
	})
	if err != nil {
		respondDB(c, "Failed to update RCM item", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "RCM item updated successfully",
		"data":    existing,
	})
}

// handleCosoResolveError writes the response for a failed resolveCosoWrite and
// reports whether the handler may continue.
func (ctrl *RCMController) handleCosoResolveError(c *gin.Context, fields fieldErrors, err error) bool {
	switch {
	case errors.Is(err, errDepartmentsUnavailable):
		respondDepartmentsUnavailable(c)
		return false
	case errors.Is(err, errLocationsUnavailable):
		respondLocationsUnavailable(c)
		return false
	case err != nil:
		respondDB(c, "Failed to validate RCM item", err)
		return false
	case fields != nil:
		respondValidation(c, fields)
		return false
	}
	return true
}

func (ctrl *RCMController) DeleteRCM(c *gin.Context) {
	idStr := c.Param("id")
	itemUUID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_ID",
				"message": "Invalid RCM ID format",
			},
		})
		return
	}

	if err := ctrl.db.Delete(&models.RiskControlMatrix{}, "id = ?", itemUUID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "DB_ERROR",
				"message": "Failed to delete RCM item: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "RCM item deleted successfully",
	})
}

// GetRCMSummary aggregates the filtered RCM rows. With no rows every score is
// null (never a made-up value); internal_control_effectiveness, rating and
// interpretation are also null when the rows carry no inherent risk.
func (ctrl *RCMController) GetRCMSummary(c *gin.Context) {
	db := ctrl.db.WithContext(c.Request.Context())
	query, fields, err := applyRCMFilters(c, db, db.Model(&models.RiskControlMatrix{}))
	if err != nil {
		respondDB(c, "Failed to query RCM summary", err)
		return
	}
	if fields != nil {
		respondValidation(c, fields)
		return
	}

	var list []models.RiskControlMatrix
	if err := query.Find(&list).Error; err != nil {
		respondDB(c, "Failed to query RCM summary", err)
		return
	}

	if len(list) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"total_controls":                 0,
				"avg_total_weighted_score":       nil,
				"avg_design_effectiveness":       nil,
				"avg_operating_effectiveness":    nil,
				"avg_coverage_completeness":      nil,
				"avg_timeliness":                 nil,
				"avg_automation_monitoring":      nil,
				"total_inherent_risk":            0,
				"total_residual_risk":            0,
				"internal_control_effectiveness": nil,
				"rating":                         nil,
				"interpretation":                 nil,
			},
		})
		return
	}

	var totalDesign, totalOperating, totalCoverage, totalTimeliness, totalAutomation float64
	var totalWeightedSum float64
	totalInherent := 0
	totalResidual := 0

	for _, item := range list {
		totalDesign += float64(item.DesignEffectivenessRating)
		totalOperating += float64(item.OperatingEffectivenessRating)
		totalCoverage += float64(item.CoverageCompletenessRating)
		totalTimeliness += float64(item.TimelinessRating)
		totalAutomation += float64(item.AutomationMonitoringRating)
		totalWeightedSum += item.TotalWeightedScore
		totalInherent += item.InherentRisk
		totalResidual += item.ResidualRisk
	}

	n := float64(len(list))
	avgDesign := math.Round((totalDesign/n)*100) / 100
	avgOperating := math.Round((totalOperating/n)*100) / 100
	avgCoverage := math.Round((totalCoverage/n)*100) / 100
	avgTimeliness := math.Round((totalTimeliness/n)*100) / 100
	avgAutomation := math.Round((totalAutomation/n)*100) / 100
	avgTotalWeighted := math.Round((totalWeightedSum/n)*100) / 100

	// Aggregate inherent and residual risk effectiveness
	// Formula: (1 - Residual Risk / Inherent Risk) * 100%
	// Without inherent risk there is nothing to compute it from: null.
	var effectiveness, rating, interpretation any
	if totalInherent > 0 {
		eff := (1.0 - (float64(totalResidual) / float64(totalInherent))) * 100.0
		eff = math.Round(eff*100) / 100
		r, i := GetEffectivenessRating(eff)
		effectiveness, rating, interpretation = eff, r, i
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_controls":                 len(list),
			"avg_total_weighted_score":       avgTotalWeighted,
			"avg_design_effectiveness":       avgDesign,
			"avg_operating_effectiveness":    avgOperating,
			"avg_coverage_completeness":      avgCoverage,
			"avg_timeliness":                 avgTimeliness,
			"avg_automation_monitoring":      avgAutomation,
			"total_inherent_risk":            totalInherent,
			"total_residual_risk":            totalResidual,
			"internal_control_effectiveness": effectiveness,
			"rating":                         rating,
			"interpretation":                 interpretation,
		},
	})
}
