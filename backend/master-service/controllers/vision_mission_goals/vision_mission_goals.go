package vision_mission_goals

import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	vmgsvc "master-service/services/vision_mission_goals"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Request DTOs
type CreateVmgRequest struct {
	CompanyID      string   `json:"company_id" binding:"required"`
	Period        string   `json:"period" binding:"required"`
	EffectiveDate string   `json:"effective_date"`
	Vision        string   `json:"vision"`
	Mission       string   `json:"mission"`
	Version       string   `json:"version"`
	Status        string   `json:"status"`
	Notes         string   `json:"notes"`
	Goals         []GoalDTO `json:"goals"`
}

type UpdateVmgRequest struct {
	CompanyID      *string  `json:"company_id"`
	Period        *string  `json:"period"`
	EffectiveDate *string  `json:"effective_date"`
	Vision        *string  `json:"vision"`
	Mission       *string  `json:"mission"`
	Version       *string  `json:"version"`
	Status        *string  `json:"status"`
	Notes         *string  `json:"notes"`
	Goals         []GoalDTO `json:"goals"`
}

type GoalDTO struct {
	ID                 string `json:"id,omitempty"`
	GoalCode           string `json:"goal_code" binding:"required"`
	GoalName           string `json:"goal_name" binding:"required"`
	GoalDescription    string `json:"goal_description"`
	StrategicObjective string `json:"strategic_objective"`
	KPI                string `json:"kpi"`
	Target             string `json:"target"`
	Unit               string `json:"unit"`
	BaselineYear       string `json:"baseline_year"`
	BaselineValue      string `json:"baseline_value"`
}

type VisionMissionGoalsControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	List(ctx *gin.Context)
	FindByCompany(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

type VisionMissionGoalsController struct {
	*base.BaseController
	vmgsvc vmgsvc.VisionMissionGoalsServiceInterface
}

func NewVisionMissionGoalsController(vmgsvc vmgsvc.VisionMissionGoalsServiceInterface, validator *validations.Validator) VisionMissionGoalsControllerInterface {
	return &VisionMissionGoalsController{
		BaseController: base.NewBaseController(validator),
		vmgsvc:        vmgsvc,
	}
}

func (c *VisionMissionGoalsController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())

	vmgs, err := c.vmgsvc.FindAll(baseService)
	if err != nil {
		c.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Vision, Mission & Goals fetched successfully", vmgs)
}

func (c *VisionMissionGoalsController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")

	vmg, err := c.vmgsvc.FindById(baseService, id)
	if err != nil {
		c.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Vision, Mission & Goals fetched successfully", vmg)
}

func (c *VisionMissionGoalsController) List(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())

	// Parse query params
	page := 1
	pageSize := 10
	if p := ctx.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := ctx.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}
	search := ctx.Query("search")

	offset := (page - 1) * pageSize

	vmgs, total, err := c.vmgsvc.FindMany(baseService, offset, pageSize, search)
	if err != nil {
		c.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Vision, Mission & Goals fetched successfully", gin.H{
		"items": vmgs,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

func (c *VisionMissionGoalsController) FindByCompany(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	companyId := ctx.Param("companyId")

	vmgs, err := c.vmgsvc.FindByCompanyId(baseService, companyId)
	if err != nil {
		c.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Vision, Mission & Goals fetched successfully", vmgs)
}

func (c *VisionMissionGoalsController) Create(ctx *gin.Context) {
	var req CreateVmgRequest
	if !c.ValidateRequest(ctx, &req) {
		return
	}

	baseService := base.NewBaseService().WithContext(ctx.Request.Context())

	// Parse company_id
	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		response.BadRequest(ctx, "Invalid company_id format")
		return
	}

	// Parse effective_date if provided
	var effectiveDate *time.Time
	if req.EffectiveDate != "" {
		parsedDate, err := time.Parse("2006-01-02", req.EffectiveDate)
		if err != nil {
			response.BadRequest(ctx, "Invalid effective_date format. Use YYYY-MM-DD")
			return
		}
		effectiveDate = &parsedDate
	}

	// Convert goals
	goals := make([]models.VmgGoal, 0, len(req.Goals))
	for _, g := range req.Goals {
		goals = append(goals, models.VmgGoal{
			GoalCode:           g.GoalCode,
			GoalName:           g.GoalName,
			GoalDescription:    g.GoalDescription,
			StrategicObjective:  g.StrategicObjective,
			KPI:                g.KPI,
			Target:             g.Target,
			Unit:               g.Unit,
			BaselineYear:       g.BaselineYear,
			BaselineValue:      g.BaselineValue,
		})
	}

	// Determine status
	status := models.VmgStatusDraft
	if req.Status != "" {
		status = models.VisionMissionGoalsStatus(req.Status)
	}

	version := "v1.0"
	if req.Version != "" {
		version = req.Version
	}

	vmg := &models.VisionMissionGoals{
		CompanyID:      companyID,
		Period:         req.Period,
		EffectiveDate:  effectiveDate,
		Vision:         req.Vision,
		Mission:        req.Mission,
		Version:        version,
		Status:         status,
		Notes:          req.Notes,
		Goals:          goals,
	}

	// Get created_by from context (auth middleware should set this)
	createdBy := c.GetUsername(ctx)

	result, err := c.vmgsvc.Create(baseService, vmg, createdBy)
	if err != nil {
		c.RespondError(ctx, err)
		return
	}

	response.Created(ctx, "Vision, Mission & Goals created successfully", result)
}

func (c *VisionMissionGoalsController) Update(ctx *gin.Context) {
	var req UpdateVmgRequest
	if !c.ValidateRequest(ctx, &req) {
		return
	}

	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")

	// First get existing VMG
	existing, err := c.vmgsvc.FindById(baseService, id)
	if err != nil {
		c.RespondError(ctx, err)
		return
	}

	// Update fields if provided
	if req.Period != nil {
		existing.Period = *req.Period
	}
	if req.EffectiveDate != nil {
		if *req.EffectiveDate == "" {
			existing.EffectiveDate = nil
		} else {
			parsedDate, err := time.Parse("2006-01-02", *req.EffectiveDate)
			if err != nil {
				response.BadRequest(ctx, "Invalid effective_date format. Use YYYY-MM-DD")
				return
			}
			existing.EffectiveDate = &parsedDate
		}
	}
	if req.Vision != nil {
		existing.Vision = *req.Vision
	}
	if req.Mission != nil {
		existing.Mission = *req.Mission
	}
	if req.Version != nil {
		existing.Version = *req.Version
	}
	if req.Status != nil {
		existing.Status = models.VisionMissionGoalsStatus(*req.Status)
	}
	if req.Notes != nil {
		existing.Notes = *req.Notes
	}

	// Update goals if provided
	if len(req.Goals) > 0 {
		existing.Goals = make([]models.VmgGoal, 0, len(req.Goals))
		for _, g := range req.Goals {
			existing.Goals = append(existing.Goals, models.VmgGoal{
				GoalCode:           g.GoalCode,
				GoalName:           g.GoalName,
				GoalDescription:    g.GoalDescription,
				StrategicObjective: g.StrategicObjective,
				KPI:                g.KPI,
				Target:             g.Target,
				Unit:               g.Unit,
				BaselineYear:       g.BaselineYear,
				BaselineValue:      g.BaselineValue,
			})
		}
	}

	modifiedBy := c.GetUsername(ctx)
	result, err := c.vmgsvc.Update(baseService, id, existing, modifiedBy)
	if err != nil {
		c.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Vision, Mission & Goals updated successfully", result)
}

func (c *VisionMissionGoalsController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")

	if err := c.vmgsvc.Delete(baseService, id); err != nil {
		c.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Vision, Mission & Goals deleted successfully", nil)
}
