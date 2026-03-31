package controllers

import (
	"strconv"

	"auth-service/models"
	"auth-service/pkg/base"
	apperrors "auth-service/pkg/errors"
	"auth-service/pkg/response"
	"auth-service/pkg/validations"
	userService "auth-service/services/user"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserControllerInterface defines the user controller interface
type UserControllerInterface interface {
	CreateUser(c *gin.Context)
	UpdateUser(c *gin.Context)
	DeleteUser(c *gin.Context)
	GetUser(c *gin.Context)
	ListUsers(c *gin.Context)
}

// UserController handles user HTTP requests
type UserController struct {
	*base.BaseController
	userService userService.UserServiceInterface
}

// NewUserController creates a new user controller
func NewUserController(
	validator *validations.Validator,
	userService userService.UserServiceInterface,
) UserControllerInterface {
	return &UserController{
		BaseController: base.NewBaseController(validator),
		userService:    userService,
	}
}

// CreateUser creates a new user
// @Summary Create User
// @Description Create a new user
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body models.CreateUserRequest true "Create user request"
// @Success 201 {object} response.Response
// @Router /api/v1/users [post]
func (ctrl *UserController) CreateUser(c *gin.Context) {
	var req models.CreateUserRequest
	if !ctrl.ValidateRequest(c, &req) {
		return
	}

	if err := ctrl.userService.CreateUser(c.Request.Context(), &req); err != nil {
		appErr, ok := err.(*apperrors.AppError)
		if ok {
			response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, "")
		} else {
			response.InternalServerError(c, err.Error())
		}
		return
	}

	response.Created(c, "User created successfully", nil)
}

// UpdateUser updates a user
// @Summary Update User
// @Description Update a user
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "User ID"
// @Param request body models.UpdateUserRequest true "Update user request"
// @Success 200 {object} response.Response
// @Router /api/v1/users/{id} [put]
func (ctrl *UserController) UpdateUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req models.UpdateUserRequest
	if !ctrl.ValidateRequest(c, &req) {
		return
	}

	if err := ctrl.userService.UpdateUser(c.Request.Context(), id, &req); err != nil {
		appErr, ok := err.(*apperrors.AppError)
		if ok {
			response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, "")
		} else {
			response.InternalServerError(c, err.Error())
		}
		return
	}

	response.OK(c, "User updated successfully", nil)
}

// DeleteUser deletes a user
// @Summary Delete User
// @Description Delete a user
// @Tags users
// @Produce json
// @Security Bearer
// @Param id path string true "User ID"
// @Success 200 {object} response.Response
// @Router /api/v1/users/{id} [delete]
func (ctrl *UserController) DeleteUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	if err := ctrl.userService.DeleteUser(c.Request.Context(), id); err != nil {
		appErr, ok := err.(*apperrors.AppError)
		if ok {
			response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, "")
		} else {
			response.InternalServerError(c, err.Error())
		}
		return
	}

	response.OK(c, "User deleted successfully", nil)
}

// GetUser retrieves a user by ID
// @Summary Get User
// @Description Get a user by ID
// @Tags users
// @Produce json
// @Security Bearer
// @Param id path string true "User ID"
// @Success 200 {object} response.Response{data=models.UserResponse}
// @Router /api/v1/users/{id} [get]
func (ctrl *UserController) GetUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	user, err := ctrl.userService.GetUser(c.Request.Context(), id)
	if err != nil {
		appErr, ok := err.(*apperrors.AppError)
		if ok {
			response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, "")
		} else {
			response.InternalServerError(c, err.Error())
		}
		return
	}

	response.OK(c, "User retrieved successfully", user)
}

// ListUsers retrieves a list of users
// @Summary List Users
// @Description Get a list of users with pagination
// @Tags users
// @Produce json
// @Security Bearer
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param search query string false "Search query"
// @Param department query string false "Department filter"
// @Param is_active query bool false "Active status filter"
// @Success 200 {object} response.Response{data=[]models.UserResponse}
// @Router /api/v1/users [get]
func (ctrl *UserController) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	req := &models.ListUsersRequest{
		Page:       page,
		PageSize:   pageSize,
		Search:     c.Query("search"),
		Department: c.Query("department"),
	}

	users, pagination, err := ctrl.userService.ListUsers(c.Request.Context(), req)
	if err != nil {
		appErr, ok := err.(*apperrors.AppError)
		if ok {
			response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, "")
		} else {
			response.InternalServerError(c, err.Error())
		}
		return
	}

	response.OK(c, "Users retrieved successfully", gin.H{
		"users":      users,
		"pagination": pagination,
	})
}
