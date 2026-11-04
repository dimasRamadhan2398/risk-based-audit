package controllers

import (
	"auth-service/models"
	"auth-service/pkg/base"
	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/response"
	"auth-service/pkg/validations"
	authService "auth-service/services/auth"

	"github.com/gin-gonic/gin"
)

// StepUpControllerInterface serves POST /auth/verify-password.
type StepUpControllerInterface interface {
	VerifyPassword(c *gin.Context)
}

// StepUpController issues password step-up tokens.
type StepUpController struct {
	*base.BaseController
	stepUp authService.StepUpServiceInterface
}

// NewStepUpController creates the step-up controller.
func NewStepUpController(validator *validations.Validator, stepUp authService.StepUpServiceInterface) StepUpControllerInterface {
	return &StepUpController{BaseController: base.NewBaseController(validator), stepUp: stepUp}
}

// VerifyPassword re-checks the signed-in user's password for a privileged action
// @Summary Verify password (step-up)
// @Description Re-checks the authenticated user's password and returns a one-time step-up token (5 minutes) for the given purpose, sent as X-Step-Up-Token. Only purpose "org_levels.change" is accepted. 5 wrong passwords per 15 minutes lock the user out of this endpoint.
// @Tags auth
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.VerifyPasswordRequest true "Password and purpose"
// @Success 200 {object} response.Response{data=models.VerifyPasswordResponse}
// @Failure 400 {object} response.Response "Invalid body or purpose"
// @Failure 401 {object} response.Response "Not authenticated"
// @Failure 403 {object} response.Response "Wrong password"
// @Failure 429 {object} response.Response "Too many wrong passwords"
// @Failure 503 {object} response.Response "Password check unavailable"
// @Router /api/v1/auth/verify-password [post]
func (ctrl *StepUpController) VerifyPassword(c *gin.Context) {
	var req models.VerifyPasswordRequest
	if !ctrl.ValidateRequest(c, &req) {
		return
	}
	userID, err := ctrl.GetUserID(c)
	if err != nil || userID == "" {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	res, err := ctrl.stepUp.VerifyPassword(c.Request.Context(), authService.VerifyPasswordInput{
		UserID:    userID,
		Username:  ctrl.GetUsername(c),
		Password:  req.Password,
		Purpose:   req.Purpose,
		IP:        getClientIP(c),
		UserAgent: getUserAgent(c),
	})
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, "")
		} else {
			response.InternalServerError(c, "Internal server error")
		}
		return
	}
	response.OK(c, "Password verified", res)
}
