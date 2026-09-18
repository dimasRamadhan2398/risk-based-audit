package controllers

import (
	"auth-service/pkg/base"
	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/response"
	"auth-service/pkg/validations"
	resendService "auth-service/services/resend"

	"github.com/gin-gonic/gin"
)

// ResendControllerInterface defines the Resend provisioning HTTP handler
type ResendControllerInterface interface {
	ProvisionClientDomain(c *gin.Context)
}

// ResendController handles Resend provisioning requests
type ResendController struct {
	*base.BaseController
	resendService resendService.ResendServiceInterface
}

// NewResendController creates a new ResendController
func NewResendController(
	validator *validations.Validator,
	resendSvc resendService.ResendServiceInterface,
) ResendControllerInterface {
	return &ResendController{
		BaseController: base.NewBaseController(validator),
		resendService:  resendSvc,
	}
}

// ProvisionClientDomain registers a new Resend domain and creates a scoped API key
// @Summary      Provision Resend sending domain
// @Description  Creates a sending domain on Resend and returns a scoped API key + DNS records for the tenant.
//
//	This endpoint is protected by ADMIN role. The returned `client_api_key` must be
//	stored securely and injected into the tenant's auth-service SMTP/Resend config.
//
// @Tags         resend
// @Accept       json
// @Produce      json
// @Security     Bearer
// @Param        request body  resendService.ProvisionDomainRequest  true  "Provision request"
// @Success      200  {object}  response.Response{data=resendService.ProvisionDomainResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Failure      502  {object}  response.Response
// @Router       /api/v1/resend/provision [post]
func (ctrl *ResendController) ProvisionClientDomain(c *gin.Context) {
	var req resendService.ProvisionDomainRequest
	if !ctrl.ValidateRequest(c, &req) {
		return
	}

	result, err := ctrl.resendService.ProvisionClientDomain(c.Request.Context(), &req)
	if err != nil {
		appErr, ok := err.(*apperrors.AppError)
		if ok {
			response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, "")
		} else {
			response.InternalServerError(c, err.Error())
		}
		return
	}

	response.OK(c, "Client sending domain provisioned successfully", result)
}
