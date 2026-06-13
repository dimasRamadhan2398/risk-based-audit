package controllers

import (
	"strconv"

	"audit-service/models"
	"audit-service/pkg/response"
	svcCharter "audit-service/services/audit_charter"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuditCharterControllerInterface defines the audit charter controller interface
type AuditCharterControllerInterface interface {
	CreateCharter(c *gin.Context)
	UpdateCharter(c *gin.Context)
	DeleteCharter(c *gin.Context)
	GetCharter(c *gin.Context)
	GetCharterByVersion(c *gin.Context)
	GetActiveCharter(c *gin.Context)
	ListCharters(c *gin.Context)
	SetActiveCharter(c *gin.Context)
}

// AuditCharterController handles audit charter HTTP requests
type AuditCharterController struct {
	service svcCharter.AuditCharterServiceInterface
}

// NewAuditCharterController creates a new audit charter controller
func NewAuditCharterController(service svcCharter.AuditCharterServiceInterface) AuditCharterControllerInterface {
	return &AuditCharterController{
		service: service,
	}
}

// CreateCharter creates a new audit charter
// @Summary Create Audit Charter
// @Description Create a new audit charter
// @Tags audit-charters
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.CreateAuditCharterRequest true "Create audit charter request"
// @Success 201 {object} response.Response
// @Router /api/v1/audit-charters [post]
func (ctrl *AuditCharterController) CreateCharter(c *gin.Context) {
	var req models.CreateAuditCharterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := ctrl.service.CreateCharter(c.Request.Context(), &req)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.Created(c, "Audit charter created successfully", result)
}

// UpdateCharter updates an audit charter
// @Summary Update Audit Charter
// @Description Update an audit charter
// @Tags audit-charters
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Audit Charter ID"
// @Param request body models.UpdateAuditCharterRequest true "Update audit charter request"
// @Success 200 {object} response.Response
// @Router /api/v1/audit-charters/{id} [put]
func (ctrl *AuditCharterController) UpdateCharter(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid audit charter ID")
		return
	}

	var req models.UpdateAuditCharterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := ctrl.service.UpdateCharter(c.Request.Context(), id, &req)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Audit charter updated successfully", result)
}

// DeleteCharter deletes an audit charter
// @Summary Delete Audit Charter
// @Description Delete an audit charter
// @Tags audit-charters
// @Produce json
// @Security Bearer
// @Param id path string true "Audit Charter ID"
// @Success 200 {object} response.Response
// @Router /api/v1/audit-charters/{id} [delete]
func (ctrl *AuditCharterController) DeleteCharter(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid audit charter ID")
		return
	}

	if err := ctrl.service.DeleteCharter(c.Request.Context(), id); err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Audit charter deleted successfully", nil)
}

// GetCharter retrieves an audit charter by ID
// @Summary Get Audit Charter
// @Description Get an audit charter by ID
// @Tags audit-charters
// @Produce json
// @Security Bearer
// @Param id path string true "Audit Charter ID"
// @Success 200 {object} response.Response{data=models.AuditCharterResponse}
// @Router /api/v1/audit-charters/{id} [get]
func (ctrl *AuditCharterController) GetCharter(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid audit charter ID")
		return
	}

	result, err := ctrl.service.GetCharter(c.Request.Context(), id)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Audit charter retrieved successfully", result)
}

// GetCharterByVersion retrieves an audit charter by version
// @Summary Get Audit Charter By Version
// @Description Get an audit charter by version
// @Tags audit-charters
// @Produce json
// @Security Bearer
// @Param version query string true "Version"
// @Success 200 {object} response.Response{data=models.AuditCharterResponse}
// @Router /api/v1/audit-charters/version [get]
func (ctrl *AuditCharterController) GetCharterByVersion(c *gin.Context) {
	version := c.Query("version")
	if version == "" {
		response.BadRequest(c, "Version is required")
		return
	}

	result, err := ctrl.service.GetCharterByVersion(c.Request.Context(), version)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Audit charter retrieved successfully", result)
}

// GetActiveCharter retrieves the active audit charter
// @Summary Get Active Audit Charter
// @Description Get the active audit charter
// @Tags audit-charters
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=models.AuditCharterResponse}
// @Router /api/v1/audit-charters/active [get]
func (ctrl *AuditCharterController) GetActiveCharter(c *gin.Context) {
	result, err := ctrl.service.GetActiveCharter(c.Request.Context())
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Active audit charter retrieved successfully", result)
}

// ListCharters retrieves a list of audit charters with pagination
// @Summary List Audit Charters
// @Description Get a list of audit charters with pagination
// @Tags audit-charters
// @Produce json
// @Security Bearer
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param search query string false "Search query"
// @Param is_active query bool false "Active status filter"
// @Success 200 {object} response.Response
// @Router /api/v1/audit-charters [get]
func (ctrl *AuditCharterController) ListCharters(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	req := &models.ListAuditChartersRequest{
		Page:     page,
		PageSize: pageSize,
		Search:   c.Query("search"),
	}

	if isActive := c.Query("is_active"); isActive != "" {
		val := isActive == "true"
		req.IsActive = &val
	}

	results, pagination, err := ctrl.service.ListCharters(c.Request.Context(), req)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Audit charters retrieved successfully", gin.H{
		"charters":   results,
		"pagination": pagination,
	})
}

// SetActiveCharter sets an audit charter as active
// @Summary Set Active Audit Charter
// @Description Set an audit charter as the active charter
// @Tags audit-charters
// @Produce json
// @Security Bearer
// @Param id path string true "Audit Charter ID"
// @Success 200 {object} response.Response
// @Router /api/v1/audit-charters/{id}/activate [post]
func (ctrl *AuditCharterController) SetActiveCharter(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid audit charter ID")
		return
	}

	if err := ctrl.service.SetActiveCharter(c.Request.Context(), id); err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Audit charter set as active successfully", nil)
}

// handleError handles service errors
func (ctrl *AuditCharterController) handleError(c *gin.Context, err error) {
	if appErr, ok := err.(*response.AppError); ok {
		response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, "")
		return
	}
	response.InternalServerError(c, "An unexpected error occurred")
}