package controllers

import (
	"io"
	"net/http"

	"auth-service/pkg/base"
	"auth-service/pkg/validations"
	resendService "auth-service/services/resend"

	"github.com/gin-gonic/gin"
)

// ResendControllerInterface defines the Resend webhook HTTP handler
type ResendControllerInterface interface {
	HandleWebhook(c *gin.Context)
}

// ResendController handles Resend webhook requests
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

// HandleWebhook handles inbound Resend webhook events (delivery, bounces, spam complaints)
// @Summary      Handle Resend webhook event
// @Description  Receives and processes webhook events from Resend with Svix signature verification
// @Tags         resend
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Router       /api/v1/resend/webhook [post]
func (ctrl *ResendController) HandleWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	if !ctrl.resendService.VerifyWebhookSignature(body, c.Request.Header) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid webhook signature or secret"})
		return
	}

	if err := ctrl.resendService.HandleWebhookEvent(c.Request.Context(), body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"received": true,
		"status":   "processed",
	})
}
