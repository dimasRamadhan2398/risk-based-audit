package confidentialityCtrl

import (
	"fmt"
	"strings"

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
	AcceptanceSummary(c *gin.Context)
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

// AcceptanceSummary answers GET /confidentiality/acceptance-summary?employee_codes=a,b,c
// for the COSO integrity pact signal. employee_codes is comma-separated (the
// parameter may also be repeated); codes are trimmed, blanks dropped and
// duplicates collapsed, keeping first-seen order. 1 to
// models.MaxAcceptanceSummaryCodes codes of at most
// models.MaxEmployeeCodeLength characters, else 400. The role check is on the
// route (routes.AcceptanceSummaryRoles).
func (ctrl *ConfidentialityController) AcceptanceSummary(c *gin.Context) {
	codes, msg := parseEmployeeCodes(c.QueryArray("employee_codes"))
	if msg != "" {
		response.Error(c, 400, "VALIDATION_ERROR", msg, "")
		return
	}

	items, err := ctrl.service.AcceptanceSummary(codes)
	if err != nil {
		response.Error(c, 500, "INTERNAL_ERROR", "Failed to load acceptance summary", "")
		return
	}
	response.Success(c, 200, "Acceptance summary retrieved", items)
}

// parseEmployeeCodes splits, trims and de-duplicates the employee_codes values.
// It returns a non-empty message when the input is not acceptable.
func parseEmployeeCodes(values []string) ([]string, string) {
	seen := map[string]bool{}
	var codes []string
	for _, v := range values {
		for _, code := range strings.Split(v, ",") {
			code = strings.TrimSpace(code)
			if code == "" || seen[code] {
				continue
			}
			if len(code) > models.MaxEmployeeCodeLength {
				return nil, fmt.Sprintf("employee code longer than %d characters", models.MaxEmployeeCodeLength)
			}
			seen[code] = true
			codes = append(codes, code)
			if len(codes) > models.MaxAcceptanceSummaryCodes {
				return nil, fmt.Sprintf("at most %d employee_codes per request", models.MaxAcceptanceSummaryCodes)
			}
		}
	}
	if len(codes) == 0 {
		return nil, "employee_codes is required"
	}
	return codes, ""
}
