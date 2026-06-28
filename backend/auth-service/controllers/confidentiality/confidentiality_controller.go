package confidentialityCtrl

import (
	"auth-service/models"
	"auth-service/pkg/base"
	"auth-service/pkg/response"
	"auth-service/pkg/validations"
	confidentialitySvc "auth-service/services/confidentiality"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ConfidentialityControllerInterface defines the confidentiality controller interface
type ConfidentialityControllerInterface interface {
	GetStatus(c *gin.Context)
	Accept(c *gin.Context)
}

// ConfidentialityController handles confidentiality HTTP requests
type ConfidentialityController struct {
	*base.BaseController
	service confidentialitySvc.ConfidentialityServiceInterface
}

// NewConfidentialityController creates a new confidentiality controller
func NewConfidentialityController(
	validator *validations.Validator,
	service confidentialitySvc.ConfidentialityServiceInterface,
) ConfidentialityControllerInterface {
	return &ConfidentialityController{
		BaseController: base.NewBaseController(validator),
		service:        service,
	}
}

// getUserID extracts the authenticated user's UUID from gin context (set by auth middleware)
func getUserID(c *gin.Context) (uuid.UUID, bool) {
	raw, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	idStr, ok := raw.(string)
	if !ok {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

// GetStatus returns whether the authenticated user has accepted the latest confidentiality agreement
func (ctrl *ConfidentialityController) GetStatus(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	hasAccepted, acceptedAt, err := ctrl.service.GetStatus(userID)
	if err != nil {
		response.Error(c, 500, "INTERNAL_ERROR", "Failed to check status", "")
		return
	}

	result := models.ConfidentialityStatusResponse{
		HasAccepted: hasAccepted,
	}
	if acceptedAt != nil {
		result.AcceptedAt = acceptedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	response.Success(c, 200, "Status retrieved", result)
}

// Accept records the user's acceptance of the confidentiality agreement
func (ctrl *ConfidentialityController) Accept(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "Unauthorized")
		return
	}

	var req models.AcceptConfidentialityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	if err := ctrl.service.Accept(userID, &req, ipAddress, userAgent); err != nil {
		response.Error(c, 500, "INTERNAL_ERROR", "Failed to record agreement", "")
		return
	}

	response.Success(c, 200, "Agreement accepted", nil)
}
