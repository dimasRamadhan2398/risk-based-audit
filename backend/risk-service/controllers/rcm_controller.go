package controllers

import (
	"math"
	"net/http"
	"strconv"

	"risk-service/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RCMController struct {
	db *gorm.DB
}

func NewRCMController(db *gorm.DB) *RCMController {
	return &RCMController{db: db}
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
	ID                           *string  `json:"id,omitempty"`
	RiskID                       *string  `json:"risk_id,omitempty"`
	RiskCode                     string   `json:"risk_code"`
	RiskEvent                    string   `json:"risk_event"`
	ControlCode                  string   `json:"control_code"`
	ControlDescription           string   `json:"control_description"`
	ControlType                  string   `json:"control_type"`
	ControlOwner                 string   `json:"control_owner"`
	Department                   string   `json:"department"`
	Year                         int      `json:"year"`
	DesignEffectivenessWeight    float64  `json:"design_effectiveness_weight"`
	DesignEffectivenessRating    int      `json:"design_effectiveness_rating"`
	OperatingEffectivenessWeight float64  `json:"operating_effectiveness_weight"`
	OperatingEffectivenessRating int      `json:"operating_effectiveness_rating"`
	CoverageCompletenessWeight   float64  `json:"coverage_completeness_weight"`
	CoverageCompletenessRating   int      `json:"coverage_completeness_rating"`
	TimelinessWeight             float64  `json:"timeliness_weight"`
	TimelinessRating             int      `json:"timeliness_rating"`
	AutomationMonitoringWeight   float64  `json:"automation_monitoring_weight"`
	AutomationMonitoringRating   int      `json:"automation_monitoring_rating"`
	TotalWeightedScore           float64  `json:"total_weighted_score"`
	InherentRisk                 int      `json:"inherent_risk"`
	ResidualRisk                 int      `json:"residual_risk"`
	Notes                        string   `json:"notes"`
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
		target.Year = 2026
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
	yearStr := c.Query("year")
	dept := c.Query("department")

	query := ctrl.db.Model(&models.RiskControlMatrix{}).Order("created_at desc")

	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			query = query.Where("year = ?", y)
		}
	}
	if dept != "" && dept != "All Departments" && dept != "All Branches" {
		query = query.Where("department = ?", dept)
	}

	var list []models.RiskControlMatrix
	if err := query.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "DB_ERROR",
				"message": "Failed to list RCM items: " + err.Error(),
			},
		})
		return
	}

	if list == nil {
		list = []models.RiskControlMatrix{}
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

	parseRCMRequest(&req, &item)

	if err := ctrl.db.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "DB_ERROR",
				"message": "Failed to create RCM item: " + err.Error(),
			},
		})
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

	var existing models.RiskControlMatrix
	if err := ctrl.db.First(&existing, "id = ?", itemUUID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "RCM item not found",
			},
		})
		return
	}

	parseRCMRequest(&req, &existing)

	if err := ctrl.db.Save(&existing).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "DB_ERROR",
				"message": "Failed to update RCM item: " + err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "RCM item updated successfully",
		"data":    existing,
	})
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

func (ctrl *RCMController) GetRCMSummary(c *gin.Context) {
	yearStr := c.Query("year")
	dept := c.Query("department")

	query := ctrl.db.Model(&models.RiskControlMatrix{})
	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			query = query.Where("year = ?", y)
		}
	}
	if dept != "" && dept != "All Departments" && dept != "All Branches" {
		query = query.Where("department = ?", dept)
	}

	var list []models.RiskControlMatrix
	if err := query.Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "DB_ERROR",
				"message": "Failed to query RCM summary: " + err.Error(),
			},
		})
		return
	}

	if len(list) == 0 {
		// Return default fallback structure
		rating, interpretation := GetEffectivenessRating(60.0)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"total_controls":               0,
				"avg_total_weighted_score":    0.0,
				"avg_design_effectiveness":    0.0,
				"avg_operating_effectiveness": 0.0,
				"avg_coverage_completeness":   0.0,
				"avg_timeliness":              0.0,
				"avg_automation_monitoring":   0.0,
				"total_inherent_risk":          20,
				"total_residual_risk":          8,
				"internal_control_effectiveness": 60.0,
				"rating":                      rating,
				"interpretation":              interpretation,
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
	var effectiveness float64 = 0.0
	if totalInherent > 0 {
		effectiveness = (1.0 - (float64(totalResidual) / float64(totalInherent))) * 100.0
	}
	effectiveness = math.Round(effectiveness*100) / 100

	rating, interpretation := GetEffectivenessRating(effectiveness)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_controls":                 len(list),
			"avg_total_weighted_score":      avgTotalWeighted,
			"avg_design_effectiveness":      avgDesign,
			"avg_operating_effectiveness":   avgOperating,
			"avg_coverage_completeness":     avgCoverage,
			"avg_timeliness":                avgTimeliness,
			"avg_automation_monitoring":     avgAutomation,
			"total_inherent_risk":            totalInherent,
			"total_residual_risk":            totalResidual,
			"internal_control_effectiveness": effectiveness,
			"rating":                        rating,
			"interpretation":                interpretation,
		},
	})
}
