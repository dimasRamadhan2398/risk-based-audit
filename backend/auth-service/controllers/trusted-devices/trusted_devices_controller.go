package trusteddevices

import (
	"auth-service/models"
	"auth-service/pkg/base"
	"auth-service/pkg/response"
	"auth-service/pkg/validations"
	trustedDevicesService "auth-service/services/trusted-devices"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetTrustedDevices godoc
// @Summary Get Trusted Devices
// @Description Get list of trusted devices for the current user
// @Tags Trusted Devices
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=[]models.TrustedDevice}
// @Failure 401 {object} response.Response
// @Router /api/v1/devices [get]

// GenerateEnrollmentQR godoc
// @Summary Enroll Trusted Device
// @Description Generate enrollment QR code for a new trusted device
// @Tags Trusted Devices
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=models.TrustedDeviceEnrollResponse}
// @Failure 401 {object} response.Response
// @Router /api/v1/devices/enroll [post]

// VerifyEnrollmentToken godoc
// @Summary Verify Device Enrollment
// @Description Verify enrollment token and complete device enrollment
// @Tags Trusted Devices
// @Accept json
// @Produce json
// @Security Bearer
// @Param token path string true "Enrollment verification token"
// @Param request body models.TrustedDeviceRequest true "Trusted device request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/devices/enroll/verify/{token} [post]

// UnenrollTrustedDevice godoc
// @Summary Unenroll Trusted Device
// @Description Unenroll a trusted device
// @Tags Trusted Devices
// @Produce json
// @Security Bearer
// @Param device_id path string true "Device ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/devices/{device_id} [delete]

type TrustedDevicesController struct {
	*base.BaseController
	trustedDevicesService trustedDevicesService.TrustedDevicesServiceInterface
}

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

// getUserAgent extracts the user agent from the request
func getUserAgent(c *gin.Context) string {
	return c.GetHeader("User-Agent")
}

// GenerateEnrollmentQR implements TrustedDevicesControllerInterface.
func (t *TrustedDevicesController) GenerateEnrollmentQR(ctx *gin.Context) {
	userID, err := t.GetUserID(ctx)
	if err != nil {
		t.RespondError(ctx, err)
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		t.RespondError(ctx, err)
		return
	}

	result, err := t.trustedDevicesService.GenerateEnrollmentQR(ctx.Request.Context(), id)
	if err != nil {
		t.RespondError(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, "Success", gin.H{
		"data": result,
	})
}

// VerifyEnrollmentToken implements TrustedDevicesControllerInterface.
func (t *TrustedDevicesController) VerifyEnrollmentToken(ctx *gin.Context) {
	userID, err := t.GetUserID(ctx)
	if err != nil {
		t.RespondError(ctx, err)
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		t.RespondError(ctx, err)
		return
	}

	token := ctx.Param("token")
	if token == "" {
		response.BadRequest(ctx, "Token is required")
		return
	}

	var req models.TrustedDeviceRequest
	if !t.ValidateRequest(ctx, &req) {
		return
	}

	ipAddress := getClientIP(ctx)
	userAgent := getUserAgent(ctx)

	if err := t.trustedDevicesService.VerifyEnrollmentToken(ctx.Request.Context(), id, token, req, ipAddress, userAgent); err != nil {
		t.RespondError(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, "Device enrolled successfully", nil)
}

// GetTrustedDevices implements TrustedDevicesControllerInterface.
func (t *TrustedDevicesController) GetTrustedDevices(ctx *gin.Context) {
	userID, err := t.GetUserID(ctx)
	if err != nil {
		t.RespondError(ctx, err)
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		t.RespondError(ctx, err)
		return
	}

	results, err := t.trustedDevicesService.GetTrustedDevices(ctx, id)
	if err != nil {
		t.RespondError(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, "Success", gin.H{
		"data": results,
	})
}

// UnenrollTrustedDevice implements TrustedDevicesControllerInterface.
func (t *TrustedDevicesController) UnenrollTrustedDevice(ctx *gin.Context) {
	userID, err := t.GetUserID(ctx)
	if err != nil {
		t.RespondError(ctx, err)
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		t.RespondError(ctx, err)
		return
	}

	deviceIDStr := ctx.Param("device_id")
	if deviceIDStr == "" {
		response.BadRequest(ctx, "Device ID is required")
		return
	}

	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		response.BadRequest(ctx, "Invalid device ID")
		return
	}

	ipAddress := getClientIP(ctx)
	userAgent := getUserAgent(ctx)

	if err := t.trustedDevicesService.UnenrollTrustedDevice(ctx.Request.Context(), id, deviceID, ipAddress, userAgent); err != nil {
		t.RespondError(ctx, err)
		return
	}

	response.Success(ctx, http.StatusOK, "Device unenrolled successfully", nil)
}

func NewTrustedDevicesController(
	validator *validations.Validator,
	trustedDevicesService trustedDevicesService.TrustedDevicesServiceInterface,
) TrustedDevicesControllerInterface {
	return &TrustedDevicesController{
		BaseController:        base.NewBaseController(validator),
		trustedDevicesService: trustedDevicesService,
	}
}

type TrustedDevicesControllerInterface interface {
	GetTrustedDevices(ctx *gin.Context)
	GenerateEnrollmentQR(ctx *gin.Context)
	VerifyEnrollmentToken(ctx *gin.Context)
	UnenrollTrustedDevice(ctx *gin.Context)
}
