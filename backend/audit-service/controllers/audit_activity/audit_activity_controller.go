package controllers

import (
	"strconv"

	"audit-service/models"
	pkgErrors "audit-service/pkg/errors"
	"audit-service/pkg/response"
	svcActivity "audit-service/services/audit_activity"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuditActivityControllerInterface defines the audit activity controller interface
type AuditActivityControllerInterface interface {
	CreateActivity(c *gin.Context)
	UpdateActivity(c *gin.Context)
	DeleteActivity(c *gin.Context)
	GetActivity(c *gin.Context)
	GetActivityByProjectCode(c *gin.Context)
	ListActivities(c *gin.Context)
}

// AuditActivityController handles audit activity HTTP requests
type AuditActivityController struct {
	service svcActivity.AuditActivityServiceInterface
}

// NewAuditActivityController creates a new audit activity controller
func NewAuditActivityController(service svcActivity.AuditActivityServiceInterface) AuditActivityControllerInterface {
	return &AuditActivityController{
		service: service,
	}
}

// CreateActivity creates a new audit activity
// @Summary Create Audit Activity
// @Description Create a new audit activity
// @Tags audit-activities
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.CreateActivityPlanRequest true "Create audit activity request"
// @Success 201 {object} response.Response
// @Router /api/v1/audit-activities [post]
func (ctrl *AuditActivityController) CreateActivity(c *gin.Context) {
	var req models.CreateActivityPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := ctrl.service.CreateActivity(c.Request.Context(), &req)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.Created(c, "Audit activity created successfully", result)
}

// UpdateActivity updates an audit activity
// @Summary Update Audit Activity
// @Description Update an audit activity
// @Tags audit-activities
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Audit Activity ID"
// @Param request body models.UpdateActivityPlanRequest true "Update audit activity request"
// @Success 200 {object} response.Response
// @Router /api/v1/audit-activities/{id} [put]
func (ctrl *AuditActivityController) UpdateActivity(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid audit activity ID")
		return
	}

	var req models.UpdateActivityPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := ctrl.service.UpdateActivity(c.Request.Context(), id, &req)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Audit activity updated successfully", result)
}

// DeleteActivity deletes an audit activity
// @Summary Delete Audit Activity
// @Description Delete an audit activity
// @Tags audit-activities
// @Produce json
// @Security Bearer
// @Param id path string true "Audit Activity ID"
// @Success 200 {object} response.Response
// @Router /api/v1/audit-activities/{id} [delete]
func (ctrl *AuditActivityController) DeleteActivity(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid audit activity ID")
		return
	}

	if err := ctrl.service.DeleteActivity(c.Request.Context(), id); err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Audit activity deleted successfully", nil)
}

// GetActivity retrieves an audit activity by ID
// @Summary Get Audit Activity
// @Description Get an audit activity by ID
// @Tags audit-activities
// @Produce json
// @Security Bearer
// @Param id path string true "Audit Activity ID"
// @Success 200 {object} response.Response{data=models.ActivityPlanResponse}
// @Router /api/v1/audit-activities/{id} [get]
func (ctrl *AuditActivityController) GetActivity(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid audit activity ID")
		return
	}

	result, err := ctrl.service.GetActivity(c.Request.Context(), id)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Audit activity retrieved successfully", result)
}

// GetActivityByProjectCode retrieves an audit activity by project code
// @Summary Get Audit Activity By Project Code
// @Description Get an audit activity by project code
// @Tags audit-activities
// @Produce json
// @Security Bearer
// @Param project_code query string true "Project Code"
// @Success 200 {object} response.Response{data=models.ActivityPlanResponse}
// @Router /api/v1/audit-activities/project [get]
func (ctrl *AuditActivityController) GetActivityByProjectCode(c *gin.Context) {
	projectCode := c.Query("project_code")
	if projectCode == "" {
		response.BadRequest(c, "Project code is required")
		return
	}

	result, err := ctrl.service.GetActivityByProjectCode(c.Request.Context(), projectCode)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Audit activity retrieved successfully", result)
}

// ListActivities retrieves a list of audit activities with pagination
// @Summary List Audit Activities
// @Description Get a list of audit activities with pagination
// @Tags audit-activities
// @Produce json
// @Security Bearer
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param search query string false "Search query"
// @Param annual_plan_id query string false "Annual Plan ID filter"
// @Param target_unit_id query string false "Target Unit ID filter"
// @Param status query string false "Status filter"
// @Success 200 {object} response.Response
// @Router /api/v1/audit-activities [get]
func (ctrl *AuditActivityController) ListActivities(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	req := &models.ListActivityPlansRequest{
		Page:     page,
		PageSize: pageSize,
		Search:   c.Query("search"),
	}

	if annualPlanID := c.Query("annual_plan_id"); annualPlanID != "" {
		if id, err := uuid.Parse(annualPlanID); err == nil {
			req.AnnualPlanID = &id
		}
	}

	if targetUnitID := c.Query("target_unit_id"); targetUnitID != "" {
		if id, err := uuid.Parse(targetUnitID); err == nil {
			req.TargetUnitID = &id
		}
	}

	if status := c.Query("status"); status != "" {
		req.Status = &status
	}

	results, pagination, err := ctrl.service.ListActivities(c.Request.Context(), req)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Audit activities retrieved successfully", gin.H{
		"activities": results,
		"pagination": pagination,
	})
}

// handleError handles service errors
func (ctrl *AuditActivityController) handleError(c *gin.Context, err error) {
	if appErr, ok := err.(*pkgErrors.AppError); ok {
		response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, "")
		return
	}
	response.InternalServerError(c, "An unexpected error occurred")
}
