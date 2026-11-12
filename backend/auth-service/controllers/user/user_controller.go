package controllers

import (
	"strconv"
	"strings"

	"auth-service/models"
	"auth-service/pkg/base"
	"auth-service/pkg/middleware"
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
	LookupByEmployee(c *gin.Context)
	ResetPassword(c *gin.Context)
	LinkEmployee(c *gin.Context)
	ListAssignableUsers(c *gin.Context)
	ListUserContacts(c *gin.Context)
	GetSeats(c *gin.Context)
	UpdateUserStatus(c *gin.Context)
	ReplaceUserRoles(c *gin.Context)
}

// ListAssignableUsers lists active users that can be assigned as the PIC of
// an action taken report. The route restricts it to the assigner/reviewer
// roles (routes.AssignableUserRoles).
// @Summary List Assignable Users
// @Description Active users with id, full_name, department and position only. Search is case-insensitive on full_name, department and position; ordered by full_name, id.
// @Tags users
// @Produce json
// @Security Bearer
// @Param search query string false "Search on full name, department or position"
// @Param id query string false "Restrict to one user id"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size (max 100)" default(20)
// @Success 200 {object} response.Response
// @Router /api/v1/users/assignable [get]
func (ctrl *UserController) ListAssignableUsers(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", strconv.Itoa(userService.AssignableDefaultPageSize)))
	if err != nil {
		pageSize = userService.AssignableDefaultPageSize
	}
	req := &models.ListAssignableUsersRequest{
		Page:     page,
		PageSize: pageSize,
		Search:   c.Query("search"),
	}
	if raw := c.Query("id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			response.BadRequest(c, "Invalid user ID")
			return
		}
		req.ID = &id
	}

	users, pagination, err := ctrl.userService.ListAssignableUsers(c.Request.Context(), req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, "")
		} else {
			response.InternalServerError(c, err.Error())
		}
		return
	}

	response.OK(c, "Assignable users retrieved successfully", gin.H{
		"users":      users,
		"pagination": pagination,
	})
}

// ListUserContacts returns the email and name of active users, for the
// reminders audit-service sends. Filters: ids (comma-separated user ids, max
// 20) and department_head_of (users holding DEPARTMENT_HEAD in that
// department). Same roles as the assignable list (routes.AssignableUserRoles).
// @Summary List User Contacts
// @Description Active users with id, full_name, email and department, filtered by ids and/or department_head_of. At least one filter is required.
// @Tags users
// @Produce json
// @Security Bearer
// @Param ids query string false "Comma-separated user ids (max 20)"
// @Param department_head_of query string false "Department whose DEPARTMENT_HEAD users to list"
// @Success 200 {object} response.Response
// @Router /api/v1/users/contacts [get]
func (ctrl *UserController) ListUserContacts(c *gin.Context) {
	req := &models.ListUserContactsRequest{DepartmentHeadOf: strings.TrimSpace(c.Query("department_head_of"))}
	for _, raw := range strings.Split(c.Query("ids"), ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		id, err := uuid.Parse(raw)
		if err != nil {
			response.BadRequest(c, "Invalid user ID in ids")
			return
		}
		req.IDs = append(req.IDs, id)
	}

	users, err := ctrl.userService.ListUserContacts(c.Request.Context(), req)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			response.Error(c, appErr.StatusCode, appErr.Code, appErr.Message, "")
		} else {
			response.InternalServerError(c, err.Error())
		}
		return
	}
	response.OK(c, "User contacts retrieved successfully", gin.H{"users": users})
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

	result, err := ctrl.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		respondError(c, err)
		return
	}

	response.Created(c, "User created successfully", result)
}

// LinkEmployee links an existing account to a master-data employee (ADMIN only)
// @Summary Link User To Employee
// @Description Set users.employee_id on an existing account (ADMIN only)
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "User ID"
// @Param request body models.LinkEmployeeRequest true "Employee code"
// @Success 200 {object} response.Response{data=models.UserResponse}
// @Router /api/v1/users/{id}/link-employee [post]
func (ctrl *UserController) LinkEmployee(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req models.LinkEmployeeRequest
	if !ctrl.ValidateRequest(c, &req) {
		return
	}

	user, err := ctrl.userService.LinkEmployee(c.Request.Context(), id, req.EmployeeCode)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OK(c, "Account linked to employee", user)
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

	// Users can edit their own profile, but only an admin can (de)activate accounts
	if req.IsActive != nil && !middleware.HasRole(c, "ADMIN") {
		response.Forbidden(c, "Only an admin can change account status")
		return
	}

	updated, err := ctrl.userService.UpdateUser(c.Request.Context(), id, &req)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OK(c, "User updated successfully", updated)
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

	hard := false
	if raw := c.Query("hard"); raw != "" {
		v, err := strconv.ParseBool(raw)
		if err != nil {
			response.BadRequest(c, "hard must be true or false")
			return
		}
		hard = v
	}

	if err := ctrl.userService.DeleteUser(c.Request.Context(), id, hard); err != nil {
		respondError(c, err)
		return
	}

	if hard {
		response.OK(c, "User permanently deleted", nil)
		return
	}
	response.OK(c, "User deleted successfully", nil)
}

// respondError writes err in the standard error envelope. AppError data (for
// example the seat counts of SEAT_LIMIT_REACHED) is returned in "data".
func respondError(c *gin.Context, err error) {
	if appErr, ok := err.(*apperrors.AppError); ok {
		response.ErrorWithData(c, appErr.StatusCode, appErr.Code, appErr.Message, "", appErr.Data)
		return
	}
	response.InternalServerError(c, err.Error())
}

// GetSeats reports the full user seats in use (ADMIN only)
// @Summary Get User Seats
// @Description Full seats (active users holding any role other than VIEWER) against the fixed limit of 50
// @Tags users
// @Produce json
// @Security Bearer
// @Success 200 {object} response.Response{data=models.SeatUsage}
// @Router /api/v1/users/seats [get]
func (ctrl *UserController) GetSeats(c *gin.Context) {
	usage, err := ctrl.userService.GetSeatUsage(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	response.OK(c, "Seat usage retrieved successfully", usage)
}

// UpdateUserStatus activates or deactivates a user (ADMIN only)
// @Summary Update User Status
// @Description Set is_active. Reactivating a user holding a full-seat role needs a free seat (409 SEAT_LIMIT_REACHED); deactivating the last active admin is refused (409 LAST_ADMIN).
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "User ID"
// @Param request body models.UpdateUserStatusRequest true "New status"
// @Success 200 {object} response.Response{data=models.UserResponse}
// @Router /api/v1/users/{id}/status [patch]
func (ctrl *UserController) UpdateUserStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}
	var req models.UpdateUserStatusRequest
	if !ctrl.ValidateRequest(c, &req) {
		return
	}
	user, err := ctrl.userService.SetUserStatus(c.Request.Context(), id, *req.IsActive)
	if err != nil {
		respondError(c, err)
		return
	}
	response.OK(c, "User status updated successfully", user)
}

// ReplaceUserRoles replaces all roles of a user (ADMIN only)
// @Summary Replace User Roles
// @Description Replace the user's roles. Moving an active user from no full-seat role to one needs a free seat (409 SEAT_LIMIT_REACHED); removing ADMIN from the last active admin is refused (409 LAST_ADMIN).
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path string true "User ID"
// @Param request body models.ReplaceUserRolesRequest true "Roles"
// @Success 200 {object} response.Response{data=models.UserResponse}
// @Router /api/v1/users/{id}/roles [put]
func (ctrl *UserController) ReplaceUserRoles(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}
	var req models.ReplaceUserRolesRequest
	if !ctrl.ValidateRequest(c, &req) {
		return
	}
	user, err := ctrl.userService.ReplaceUserRoles(c.Request.Context(), id, req.Roles)
	if err != nil {
		respondError(c, err)
		return
	}
	response.OK(c, "User roles updated successfully", user)
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
		respondError(c, err)
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
		respondError(c, err)
		return
	}

	response.OK(c, "Users retrieved successfully", gin.H{
		"users":      users,
		"pagination": pagination,
	})
}

// LookupByEmployee finds the login account linked to a master-data employee
// @Summary Lookup User By Employee
// @Description Find the login account of an employee by employee code, falling back to email (ADMIN only)
// @Tags users
// @Produce json
// @Security Bearer
// @Param employee_code query string false "Employee code"
// @Param email query string false "Employee email"
// @Success 200 {object} response.Response{data=models.UserResponse}
// @Router /api/v1/users/lookup [get]
func (ctrl *UserController) LookupByEmployee(c *gin.Context) {
	employeeCode := c.Query("employee_code")
	email := c.Query("email")
	if employeeCode == "" && email == "" {
		response.BadRequest(c, "employee_code or email is required")
		return
	}

	user, err := ctrl.userService.FindUserByEmployee(c.Request.Context(), employeeCode, email)
	if err != nil {
		respondError(c, err)
		return
	}

	response.OK(c, "User retrieved successfully", user)
}

// ResetPassword resets a user's password to a temporary one (ADMIN only)
// @Summary Reset User Password
// @Description Generate a temporary password for a user and force a password change at next login (ADMIN only)
// @Tags users
// @Produce json
// @Security Bearer
// @Param id path string true "User ID"
// @Success 200 {object} response.Response{data=models.AdminResetPasswordResponse}
// @Router /api/v1/users/{id}/reset-password [post]
func (ctrl *UserController) ResetPassword(c *gin.Context) {
	targetID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	actorIDStr, _ := ctrl.GetUserID(c)
	actorID, err := uuid.Parse(actorIDStr)
	if err != nil {
		response.Unauthorized(c, "Invalid user session")
		return
	}

	result, err := ctrl.userService.AdminResetPassword(c.Request.Context(), actorID, targetID, c.ClientIP())
	if err != nil {
		respondError(c, err)
		return
	}

	response.OK(c, "Password reset successfully", result)
}
