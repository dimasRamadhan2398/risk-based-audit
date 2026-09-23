package role

import (
	"strconv"

	"auth-service/models"
	"auth-service/pkg/base"
	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/response"
	"auth-service/pkg/validations"
	roleService "auth-service/services/role"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RoleControllerInterface defines the role controller interface
type RoleControllerInterface interface {
	CreateRole(c *gin.Context)
	UpdateRole(c *gin.Context)
	DeleteRole(c *gin.Context)
	GetRoleByID(c *gin.Context)
	GetAllRoles(c *gin.Context)
	GetRoles(c *gin.Context)
	AssignPermissions(c *gin.Context)
	RemovePermissions(c *gin.Context)
	ListPermissions(c *gin.Context)
}

// RoleController handles role HTTP requests
type RoleController struct {
	*base.BaseController
	roleService roleService.IRoleService
}

// NewRoleController creates a new role controller
func NewRoleController(
	validator *validations.Validator,
	roleService roleService.IRoleService,
) RoleControllerInterface {
	return &RoleController{
		BaseController: base.NewBaseController(validator),
		roleService:    roleService,
	}
}

// CreateRole creates a new role
func (ctrl *RoleController) CreateRole(c *gin.Context) {
	var req models.CreateRoleRequest
	if !ctrl.ValidateRequest(c, &req) {
		return
	}

	role, err := ctrl.roleService.CreateRole(c.Request.Context(), req)
	if err != nil {
		appErr, ok := err.(*apperrors.AppError)
		if ok {
			response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, "")
		} else {
			response.InternalServerError(c, err.Error())
		}
		return
	}

	response.Created(c, "Role created successfully", role)
}

// UpdateRole updates an existing role
func (ctrl *RoleController) UpdateRole(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid role ID")
		return
	}

	var req models.UpdateRoleRequest
	if !ctrl.ValidateRequest(c, &req) {
		return
	}

	role, err := ctrl.roleService.UpdateRole(c.Request.Context(), id, req)
	if err != nil {
		appErr, ok := err.(*apperrors.AppError)
		if ok {
			response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, "")
		} else {
			response.InternalServerError(c, err.Error())
		}
		return
	}

	response.OK(c, "Role updated successfully", role)
}

// DeleteRole deletes a role
func (ctrl *RoleController) DeleteRole(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid role ID")
		return
	}

	if err := ctrl.roleService.DeleteRole(c.Request.Context(), id); err != nil {
		appErr, ok := err.(*apperrors.AppError)
		if ok {
			response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, "")
		} else {
			response.InternalServerError(c, err.Error())
		}
		return
	}

	response.OK(c, "Role deleted successfully", nil)
}

// GetRoleByID gets a role by ID
func (ctrl *RoleController) GetRoleByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid role ID")
		return
	}

	role, err := ctrl.roleService.GetRoleByID(c.Request.Context(), id)
	if err != nil {
		appErr, ok := err.(*apperrors.AppError)
		if ok {
			response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, "")
		} else {
			response.InternalServerError(c, err.Error())
		}
		return
	}

	response.OK(c, "Role retrieved successfully", role)
}

// GetAllRoles retrieves all roles without pagination
func (ctrl *RoleController) GetAllRoles(c *gin.Context) {
	roles, err := ctrl.roleService.GetAllRoles(c.Request.Context())
	if err != nil {
		appErr, ok := err.(*apperrors.AppError)
		if ok {
			response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, "")
		} else {
			response.InternalServerError(c, err.Error())
		}
		return
	}

	response.OK(c, "Roles retrieved successfully", roles)
}

// GetRoles retrieves paginated roles
func (ctrl *RoleController) GetRoles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")
	offset := (page - 1) * pageSize

	resp, err := ctrl.roleService.GetRoles(c.Request.Context(), offset, pageSize, search)
	if err != nil {
		appErr, ok := err.(*apperrors.AppError)
		if ok {
			response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, "")
		} else {
			response.InternalServerError(c, err.Error())
		}
		return
	}

	response.OK(c, "Roles retrieved successfully", resp)
}

// AssignPermissions assigns permissions to a role
func (ctrl *RoleController) AssignPermissions(c *gin.Context) {
	idParam := c.Param("id")
	roleID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid role ID")
		return
	}

	var req models.AssignPermissionsRequest
	if !ctrl.ValidateRequest(c, &req) {
		return
	}

	if err := ctrl.roleService.AssignPermissions(c.Request.Context(), roleID, req.PermissionIDs); err != nil {
		appErr, ok := err.(*apperrors.AppError)
		if ok {
			response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, "")
		} else {
			response.InternalServerError(c, err.Error())
		}
		return
	}

	response.OK(c, "Permissions assigned successfully", nil)
}

// RemovePermissions removes permissions from a role
func (ctrl *RoleController) RemovePermissions(c *gin.Context) {
	idParam := c.Param("id")
	roleID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid role ID")
		return
	}

	var req models.AssignPermissionsRequest
	if !ctrl.ValidateRequest(c, &req) {
		return
	}

	if err := ctrl.roleService.RemovePermissions(c.Request.Context(), roleID, req.PermissionIDs); err != nil {
		appErr, ok := err.(*apperrors.AppError)
		if ok {
			response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, "")
		} else {
			response.InternalServerError(c, err.Error())
		}
		return
	}

	response.OK(c, "Permissions removed successfully", nil)
}

// ListPermissions returns all system permissions
func (ctrl *RoleController) ListPermissions(c *gin.Context) {
	permissions, err := ctrl.roleService.ListPermissions(c.Request.Context())
	if err != nil {
		appErr, ok := err.(*apperrors.AppError)
		if ok {
			response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, "")
		} else {
			response.InternalServerError(c, err.Error())
		}
		return
	}

	response.OK(c, "Permissions retrieved successfully", permissions)
}
