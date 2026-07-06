package controllers

import (
	"strconv"

	"audit-service/models"
	pkgErrors "audit-service/pkg/errors"
	"audit-service/pkg/response"
	svcAssignment "audit-service/services/audit_assignment"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuditAssignmentControllerInterface defines the audit assignment controller interface
type AuditAssignmentControllerInterface interface {
	CreateAssignment(c *gin.Context)
	UpdateAssignment(c *gin.Context)
	DeleteAssignment(c *gin.Context)
	GetAssignment(c *gin.Context)
	GetAssignmentsByAuditor(c *gin.Context)
	GetAssignmentsByAuditPlan(c *gin.Context)
	ListAssignments(c *gin.Context)
	UpdateStatus(c *gin.Context)
}

// AuditAssignmentController handles audit assignment HTTP requests
type AuditAssignmentController struct {
	service svcAssignment.AuditAssignmentServiceInterface
}

// NewAuditAssignmentController creates a new audit assignment controller
func NewAuditAssignmentController(service svcAssignment.AuditAssignmentServiceInterface) AuditAssignmentControllerInterface {
	return &AuditAssignmentController{
		service: service,
	}
}

// CreateAssignment creates a new audit assignment
// @Summary Create Audit Assignment
// @Description Create a new audit assignment
// @Tags audit-assignments
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.CreateAuditAssignmentRequest true "Create audit assignment request"
// @Success 201 {object} response.Response
// @Router /api/v1/audit-assignments [post]
func (ctrl *AuditAssignmentController) CreateAssignment(c *gin.Context) {
	var req models.CreateAuditAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := ctrl.service.CreateAssignment(c.Request.Context(), &req)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.Created(c, "Audit assignment created successfully", result)
}

// UpdateAssignment updates an audit assignment
// @Summary Update Audit Assignment
// @Description Update an audit assignment
// @Tags audit-assignments
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Audit Assignment ID"
// @Param request body models.UpdateAuditAssignmentRequest true "Update audit assignment request"
// @Success 200 {object} response.Response
// @Router /api/v1/audit-assignments/{id} [put]
func (ctrl *AuditAssignmentController) UpdateAssignment(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid audit assignment ID")
		return
	}

	var req models.UpdateAuditAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := ctrl.service.UpdateAssignment(c.Request.Context(), id, &req)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Audit assignment updated successfully", result)
}

// DeleteAssignment deletes an audit assignment
// @Summary Delete Audit Assignment
// @Description Delete an audit assignment
// @Tags audit-assignments
// @Produce json
// @Security Bearer
// @Param id path string true "Audit Assignment ID"
// @Success 200 {object} response.Response
// @Router /api/v1/audit-assignments/{id} [delete]
func (ctrl *AuditAssignmentController) DeleteAssignment(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid audit assignment ID")
		return
	}

	if err := ctrl.service.DeleteAssignment(c.Request.Context(), id); err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Audit assignment deleted successfully", nil)
}

// GetAssignment retrieves an audit assignment by ID
// @Summary Get Audit Assignment
// @Description Get an audit assignment by ID
// @Tags audit-assignments
// @Produce json
// @Security Bearer
// @Param id path string true "Audit Assignment ID"
// @Success 200 {object} response.Response{data=models.AuditAssignmentResponse}
// @Router /api/v1/audit-assignments/{id} [get]
func (ctrl *AuditAssignmentController) GetAssignment(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid audit assignment ID")
		return
	}

	result, err := ctrl.service.GetAssignment(c.Request.Context(), id)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Audit assignment retrieved successfully", result)
}

// GetAssignmentsByAuditor retrieves audit assignments by auditor ID
// @Summary Get Assignments By Auditor
// @Description Get audit assignments by auditor ID
// @Tags audit-assignments
// @Produce json
// @Security Bearer
// @Param auditor_id path string true "Auditor ID"
// @Success 200 {object} response.Response
// @Router /api/v1/audit-assignments/auditor/{auditor_id} [get]
func (ctrl *AuditAssignmentController) GetAssignmentsByAuditor(c *gin.Context) {
	auditorIDParam := c.Param("auditor_id")
	auditorID, err := uuid.Parse(auditorIDParam)
	if err != nil {
		response.BadRequest(c, "Invalid auditor ID")
		return
	}

	results, err := ctrl.service.GetAssignmentsByAuditor(c.Request.Context(), auditorID)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Audit assignments retrieved successfully", results)
}

// GetAssignmentsByAuditPlan retrieves audit assignments by audit plan ID
// @Summary Get Assignments By Audit Plan
// @Description Get audit assignments by audit plan ID
// @Tags audit-assignments
// @Produce json
// @Security Bearer
// @Param audit_plan_id path string true "Audit Plan ID"
// @Success 200 {object} response.Response
// @Router /api/v1/audit-assignments/audit-plan/{audit_plan_id} [get]
func (ctrl *AuditAssignmentController) GetAssignmentsByAuditPlan(c *gin.Context) {
	auditPlanIDParam := c.Param("audit_plan_id")
	auditPlanID, err := uuid.Parse(auditPlanIDParam)
	if err != nil {
		response.BadRequest(c, "Invalid audit plan ID")
		return
	}

	results, err := ctrl.service.GetAssignmentsByAuditPlan(c.Request.Context(), auditPlanID)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Audit assignments retrieved successfully", results)
}

// ListAssignments retrieves a list of audit assignments with pagination
// @Summary List Audit Assignments
// @Description Get a list of audit assignments with pagination
// @Tags audit-assignments
// @Produce json
// @Security Bearer
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param search query string false "Search query"
// @Param auditor_id query string false "Auditor ID filter"
// @Param audit_plan_id query string false "Audit Plan ID filter"
// @Param status query string false "Status filter"
// @Success 200 {object} response.Response
// @Router /api/v1/audit-assignments [get]
func (ctrl *AuditAssignmentController) ListAssignments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	req := &models.ListAuditAssignmentsRequest{
		Page:     page,
		PageSize: pageSize,
		Search:   c.Query("search"),
	}

	if auditorID := c.Query("auditor_id"); auditorID != "" {
		if id, err := uuid.Parse(auditorID); err == nil {
			req.AuditorID = &id
		}
	}

	if auditPlanID := c.Query("audit_plan_id"); auditPlanID != "" {
		if id, err := uuid.Parse(auditPlanID); err == nil {
			req.AuditPlanID = &id
		}
	}

	if status := c.Query("status"); status != "" {
		req.Status = &status
	}

	results, pagination, err := ctrl.service.ListAssignments(c.Request.Context(), req)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Audit assignments retrieved successfully", gin.H{
		"assignments": results,
		"pagination":  pagination,
	})
}

// UpdateStatus updates the status of an audit assignment
// @Summary Update Assignment Status
// @Description Update the status of an audit assignment
// @Tags audit-assignments
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "Audit Assignment ID"
// @Param request body models.UpdateAssignmentStatusRequest true "Update status request"
// @Success 200 {object} response.Response
// @Router /api/v1/audit-assignments/{id}/status [put]
func (ctrl *AuditAssignmentController) UpdateStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid audit assignment ID")
		return
	}

	var req models.UpdateAssignmentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := ctrl.service.UpdateStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		ctrl.handleError(c, err)
		return
	}

	response.OK(c, "Audit assignment status updated successfully", result)
}

// handleError handles service errors
func (ctrl *AuditAssignmentController) handleError(c *gin.Context, err error) {
	if appErr, ok := err.(*pkgErrors.AppError); ok {
		response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, "")
		return
	}
	response.InternalServerError(c, "An unexpected error occurred")
}