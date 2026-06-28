package mfa

import (
	"auth-service/models"
	"auth-service/pkg/base"
	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/response"
	"auth-service/pkg/validations"
	mfaService "auth-service/services/mfa"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetMFAStatus godoc
// @Summary Get MFA Status
// @Description Get the current user's MFA status
// @Tags MFA
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=models.MFAStatus}
// @Failure 401 {object} response.Response
// @Router /api/v1/mfa/status [get]

// EnrollMFA godoc
// @Summary Enroll MFA
// @Description Initiate MFA enrollment
// @Tags MFA
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.MfaSetupRequest true "MFA enrollment request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/mfa/enroll [post]

// VerifyMFA godoc
// @Summary Verify MFA
// @Description Verify MFA code and enable MFA
// @Tags MFA
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.VerifyMfaRequest true "MFA verification request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/mfa/verify [post]

// UnenrollMFA godoc
// @Summary Disable MFA
// @Description Disable MFA for the current user
// @Tags MFA
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.DisableMfaRequest true "Disable MFA request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/mfa/disable [post]

// SendEmailCode godoc
// @Summary Send Email MFA Code
// @Description Send MFA verification code via email
// @Tags MFA
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.EmailCodeRequest true "Email MFA code request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/mfa/email-code [post]

// VerifyEmailCode godoc
// @Summary Verify Email MFA Code
// @Description Verify MFA email code and enable MFA
// @Tags MFA
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.EnableMfaRequest true "Email MFA verification request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/mfa/verify-email-code [post]

// getClientIP extracts the client IP address from the request
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header first (for reverse proxies)
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		return xff
	}
	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}
	// Fall back to RemoteAddr
	return c.ClientIP()
}

type MfaControllerInterface interface {
	GetMFAStatus(ctx *gin.Context)
	EnrollMFA(ctx *gin.Context)
	VerifyMFA(ctx *gin.Context)
	UnenrollMFA(ctx *gin.Context)
	SendEmailCode(ctx *gin.Context)
	VerifyEmailCode(ctx *gin.Context)
}

type MfaController struct {
	*base.BaseController
	mfaService mfaService.MfaServiceInterface
}

// EnrollMFA implements MfaControllerInterface.
func (m *MfaController) EnrollMFA(ctx *gin.Context) {
	userID, err := m.GetUserID(ctx)
	if err != nil {
		m.RespondError(ctx, err)
		return
	}

	var req models.MfaSetupRequest

	if !m.ValidateRequest(ctx, &req) {
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		m.RespondError(ctx, err)
		return
	}

	ipAddress := getClientIP(ctx)

	result, err := m.mfaService.SetupMFA(ctx.Request.Context(), id, req.MFAType, ipAddress)
	if err != nil {
		m.RespondError(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, "MFA enrollment initiated successfully", result)
}

// GetMFAStatus implements MfaControllerInterface.
func (m *MfaController) GetMFAStatus(ctx *gin.Context) {
	userID, err := m.GetUserID(ctx)
	if err != nil {
		m.RespondError(ctx, err)
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		m.RespondError(ctx, err)
		return
	}

	status, err := m.mfaService.GetMFAStatus(ctx.Request.Context(), id)
	if err != nil {
		m.RespondError(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, "Success", gin.H{
		"data": status,	
	})
}

// SendEmailCode implements MfaControllerInterface.
func (m *MfaController) SendEmailCode(ctx *gin.Context) {
	userID, err := m.GetUserID(ctx)
	if err != nil {
		m.RespondError(ctx, err)
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		m.RespondError(ctx, err)
		return
	}

	var req models.EmailCodeRequest

	if !m.ValidateRequest(ctx, &req) {
		return
	}

	ipAddress := getClientIP(ctx)

	if err := m.mfaService.GenerateEmailCode(ctx.Request.Context(), id, req.Email, ipAddress); err != nil {
		m.RespondError(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, "Success", gin.H{
		"message": "Email code sent successfully",
	})
}

// UnenrollMFA implements MfaControllerInterface.
func (m *MfaController) UnenrollMFA(ctx *gin.Context) {
	userID, err := m.GetUserID(ctx)
	if err != nil {
		m.RespondError(ctx, err)
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		m.RespondError(ctx, err)
		return
	}

	var req models.DisableMfaRequest

	if !m.ValidateRequest(ctx, &req) {
		return
	}

	ipAddress := getClientIP(ctx)

	if err := m.mfaService.DisableMFA(ctx.Request.Context(), id, req.Password, ipAddress); err != nil {
		m.RespondError(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, "Success", gin.H{
		"message": "MFA Disabled Successfully",
	})
}

// VerifyEmailCode implements MfaControllerInterface.
func (m *MfaController) VerifyEmailCode(ctx *gin.Context) {
	userID, err := m.GetUserID(ctx)
	if err != nil {
		m.RespondError(ctx, err)
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		m.RespondError(ctx, err)
		return
	}

	var req models.EnableMfaRequest

	if !m.ValidateRequest(ctx, &req) {
		return
	}

	valid, err := m.mfaService.VerifyEmailCode(ctx.Request.Context(), id, req.Code)
	if err != nil {
		m.RespondError(ctx, err)
		return
	}

	if !valid {
		m.RespondError(ctx, apperrors.Wrap("INVALID_CODE", "Invalid verification code", 400, nil))
		return
	}

	response.Success(ctx, http.StatusOK, "Success", gin.H{
		"message": "Email code verified successfully",	
	})
}

// VerifyMFA implements MfaControllerInterface.
func (m *MfaController) VerifyMFA(ctx *gin.Context) {
	userID, err := m.GetUserID(ctx)
	if err != nil {
		m.RespondError(ctx, err)
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		m.RespondError(ctx, err)
		return
	}

	var req models.VerifyMfaRequest

	if !m.ValidateRequest(ctx, &req) {
		return
	}

	ipAddress := getClientIP(ctx)

	if err := m.mfaService.EnableMFA(ctx.Request.Context(), id, req.Code, ipAddress); err != nil {
		respondError(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, "Success", gin.H{
		"message": "MFA enabled successfully",
	})
}

func NewMfaController(
	validator *validations.Validator,
	mfaService mfaService.MfaServiceInterface,
) MfaControllerInterface {
	return &MfaController{
		BaseController: base.NewBaseController(validator),
		mfaService:     mfaService,
	}
}

//helpers
func respondError(c *gin.Context, err error) {
	if appErr, ok := err.(*apperrors.AppError); ok {
		c.JSON(appErr.StatusCode, gin.H{
			"code":    appErr.Code,
			"message": appErr.Message,
		})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    "INTERNAL_ERROR",
		"message": "An unexpected error occurred",
	})
}