package controllers

import (
	"strconv"

	"audit-service/models"
	pkgErrors "audit-service/pkg/errors"
	"audit-service/pkg/response"
	svcMandate "audit-service/services/audit_mandate"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuditMandateControllerInterface defines the audit mandate controller interface
type AuditMandateControllerInterface interface {
	CreateMandate(c *gin.Context)
	UpdateMandate(c *gin.Context)
	DeleteMandate(c *gin.Context)
	GetMandate(c *gin.Context)
	GetMandateByReference(c *gin.Context)
	GetActiveMandate(c *gin.Context)
	ListMandates(c *gin.Context)
}

// AuditMandateController handles audit mandate HTTP requests
type AuditMandateController struct {
	service svcMandate.AuditMandateServiceInterface
}

// NewAuditMandateController creates a new audit mandate controller
func NewAuditMandateController(service svcMandate.AuditMandateServiceInterface) AuditMandateControllerInterface {
	return &AuditMandateController{
		service: service,
	}
}

// CreateMandate creates a new audit mandate
// @Summary Create Audit Mandate
// @Description Create a new audit mandate
// @Tags audit-mandates
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.CreateAuditMandateRequest true "Create audit mandate request"
// @Success 201 {object} response.Response
// @Router /api/v1/audit-mandates [post]
func (ctrl *AuditMandateController) CreateMandate(c *gin.Context) {
	var req models.CreateAuditMandateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := ctrl.service.CreateMandate(c.Request.Context(), &req)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.Created(c, "Audit mandate created successfully", result)
}

// UpdateMandate updates an audit mandate
// @Summary Update Audit Mandate
// @Description Update an audit mandate
// @Tags audit-mandates
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Audit Mandate ID"
// @Param request body models.UpdateAuditMandateRequest true "Update audit mandate request"
// @Success 200 {object} response.Response
// @Router /api/v1/audit-mandates/{id} [put]
func (ctrl *AuditMandateController) UpdateMandate(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid audit mandate ID")
		return
	}

	var req models.UpdateAuditMandateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := ctrl.service.UpdateMandate(c.Request.Context(), id, &req)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Audit mandate updated successfully", result)
}

// DeleteMandate deletes an audit mandate
// @Summary Delete Audit Mandate
// @Description Delete an audit mandate
// @Tags audit-mandates
// @Produce json
// @Security Bearer
// @Param id path string true "Audit Mandate ID"
// @Success 200 {object} response.Response
// @Router /api/v1/audit-mandates/{id} [delete]
func (ctrl *AuditMandateController) DeleteMandate(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid audit mandate ID")
		return
	}

	if err := ctrl.service.DeleteMandate(c.Request.Context(), id); err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Audit mandate deleted successfully", nil)
}

// GetMandate retrieves an audit mandate by ID
// @Summary Get Audit Mandate
// @Description Get an audit mandate by ID
// @Tags audit-mandates
// @Produce json
// @Security Bearer
// @Param id path string true "Audit Mandate ID"
// @Success 200 {object} response.Response{data=models.AuditMandateResponse}
// @Router /api/v1/audit-mandates/{id} [get]
func (ctrl *AuditMandateController) GetMandate(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid audit mandate ID")
		return
	}

	result, err := ctrl.service.GetMandate(c.Request.Context(), id)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Audit mandate retrieved successfully", result)
}

// GetMandateByReference retrieves an audit mandate by reference number
// @Summary Get Audit Mandate By Reference
// @Description Get an audit mandate by reference number
// @Tags audit-mandates
// @Produce json
// @Security Bearer
// @Param reference query string true "Reference Number"
// @Success 200 {object} response.Response{data=models.AuditMandateResponse}
// @Router /api/v1/audit-mandates/reference [get]
func (ctrl *AuditMandateController) GetMandateByReference(c *gin.Context) {
	refNumber := c.Query("reference")
	if refNumber == "" {
		response.BadRequest(c, "Reference number is required")
		return
	}

	result, err := ctrl.service.GetMandateByReference(c.Request.Context(), refNumber)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Audit mandate retrieved successfully", result)
}

// GetActiveMandate retrieves the active audit mandate
// @Summary Get Active Audit Mandate
// @Description Get the active audit mandate
// @Tags audit-mandates
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=models.AuditMandateResponse}
// @Router /api/v1/audit-mandates/active [get]
func (ctrl *AuditMandateController) GetActiveMandate(c *gin.Context) {
	result, err := ctrl.service.GetActiveMandate(c.Request.Context())
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Active audit mandate retrieved successfully", result)
}

// ListMandates retrieves a list of audit mandates with pagination
// @Summary List Audit Mandates
// @Description Get a list of audit mandates with pagination
// @Tags audit-mandates
// @Produce json
// @Security Bearer
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param search query string false "Search query"
// @Param is_active query bool false "Active status filter"
// @Success 200 {object} response.Response
// @Router /api/v1/audit-mandates [get]
func (ctrl *AuditMandateController) ListMandates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	req := &models.ListAuditMandatesRequest{
		Page:     page,
		PageSize: pageSize,
		Search:   c.Query("search"),
	}

	if isActive := c.Query("is_active"); isActive != "" {
		val := isActive == "true"
		req.IsActive = &val
	}

	results, pagination, err := ctrl.service.ListMandates(c.Request.Context(), req)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Audit mandates retrieved successfully", gin.H{
		"mandates":  results,
		"pagination": pagination,
	})
}

// handleError handles service errors
func (ctrl *AuditMandateController) handleError(c *gin.Context, err error) {
	if appErr, ok := err.(*pkgErrors.AppError); ok {
		response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, "")
		return
	}
	response.InternalServerError(c, "An unexpected error occurred")
}