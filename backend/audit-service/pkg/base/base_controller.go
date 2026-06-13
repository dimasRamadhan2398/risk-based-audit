package base

import (
	apperrors "audit-service/pkg/errors"
	"audit-service/pkg/response"
	"audit-service/pkg/validations"
	"net/http"

	"github.com/gin-gonic/gin"
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

// ValidateRequest validates a request struct
func (c *BaseController) ValidateRequest(ctx *gin.Context, req interface{}) bool {
	if err := ctx.ShouldBindJSON(req); err != nil {
		response.BadRequest(ctx, err.Error())
		return false
	}

	if err := c.validator.Validate(req); err != nil {
		response.BadRequest(ctx, err.Error())
		return false
	}

	return true
}

// ValidateQuery validates a query struct
func (c *BaseController) ValidateQuery(ctx *gin.Context, req interface{}) bool {
	if err := ctx.ShouldBindQuery(req); err != nil {
		response.BadRequest(ctx, err.Error())
		return false
	}

	if err := c.validator.Validate(req); err != nil {
		response.BadRequest(ctx, err.Error())
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

func (c *BaseController) RespondError(ctx *gin.Context, err error) {
	if appErr, ok := err.(*apperrors.AppError); ok {
		response.Error(ctx, appErr.StatusCode, appErr.Code, appErr.Message, "")
		return
	}
	response.Error(ctx, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred", "")
}
