package base

import (
	apperrors "master-service/pkg/errors"
	"master-service/pkg/logger"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// BaseController provides common controller operations
type BaseController struct {
	validator *validations.Validator
}

// NewBaseController creates a new base controller
func NewBaseController(validator *validations.Validator) *BaseController {
	return &BaseController{
		validator: validator,
	}
}

// ValidateRequest binds and validates a JSON body. On failure it responds with
// INVALID_REQUEST_BODY or VALIDATION_FAILED (+ fields); raw parser/validator
// text is logged, never returned.
func (c *BaseController) ValidateRequest(ctx *gin.Context, req interface{}) bool {
	if err := ctx.ShouldBindJSON(req); err != nil {
		c.RespondError(ctx, apperrors.FromBinding(err, req))
		return false
	}

	if err := c.validator.Validate(req); err != nil {
		c.RespondError(ctx, apperrors.FromBinding(err, req))
		return false
	}

	return true
}

// ValidateQuery binds and validates query parameters.
func (c *BaseController) ValidateQuery(ctx *gin.Context, req interface{}) bool {
	if err := ctx.ShouldBindQuery(req); err != nil {
		c.RespondError(ctx, apperrors.FromBinding(err, req))
		return false
	}

	if err := c.validator.Validate(req); err != nil {
		c.RespondError(ctx, apperrors.FromBinding(err, req))
		return false
	}

	return true
}

// GetUserID retrieves user_id from context
func (c *BaseController) GetUserID(ctx *gin.Context) (string, error) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		return "", apperrors.Wrap("USER_NOT_AUTHENTICATED", "User not authenticated", 401, nil)
	}

	id, ok := userID.(string)
	if !ok {
		return "", apperrors.Wrap("INVALID_USER_ID", "Invalid user ID in context", 500, nil)
	}

	return id, nil
}

// GetUsername retrieves username from context
func (c *BaseController) GetUsername(ctx *gin.Context) string {
	if username, exists := ctx.Get("username"); exists {
		if name, ok := username.(string); ok {
			return name
		}
	}
	return ""
}

// GetRoles retrieves roles from context
func (c *BaseController) GetRoles(ctx *gin.Context) []string {
	if roles, exists := ctx.Get("roles"); exists {
		if r, ok := roles.([]string); ok {
			return r
		}
	}
	return []string{}
}

// RespondError writes a client-safe error response. The error is normalised
// to a stable code and human-readable message (DB/driver errors are mapped via
// pgconn codes); the underlying technical error is logged, never returned.
func (c *BaseController) RespondError(ctx *gin.Context, err error) {
	RespondError(ctx, err)
}

// RespondError is the package-level form of BaseController.RespondError.
func RespondError(ctx *gin.Context, err error) {
	appErr := apperrors.Normalize(err, apperrors.OpFromHTTPMethod(ctx.Request.Method))
	if appErr == nil {
		appErr = apperrors.Internal("", nil)
	}
	status := appErr.StatusCode
	if status == 0 {
		status = http.StatusInternalServerError
	}

	logError(ctx, status, appErr)
	response.ErrorWithFields(ctx, status, appErr.Code, appErr.Message, "", appErr.Fields)
}

func logError(ctx *gin.Context, status int, appErr *apperrors.AppError) {
	log := logger.GetLogger()
	if log == nil {
		return
	}
	fields := []zap.Field{
		zap.String("code", appErr.Code),
		zap.Int("status", status),
		zap.String("method", ctx.Request.Method),
		zap.String("path", ctx.FullPath()),
	}
	if appErr.Err != nil {
		fields = append(fields, zap.Error(appErr.Err))
	}
	switch {
	case status >= http.StatusInternalServerError:
		log.Error("request failed", fields...)
	case appErr.Err != nil:
		log.Warn("request rejected", fields...)
	}
}
