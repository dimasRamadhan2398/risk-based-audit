package employee

import (
	"io"
	"strconv"
	"strings"

	"master-service/models"
	"master-service/pkg/authclient"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	employeeRepo "master-service/repositories/employee"
	employeeSvc "master-service/services/employee"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EmployeeControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	FindByCode(ctx *gin.Context)
	Create(ctx *gin.Context)
	NextCode(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
	ProvisionLoginAccount(ctx *gin.Context)
}
type EmployeeController struct {
	*base.BaseController
	employeeSvc employeeSvc.EmployeeServiceInterface
}

func NewEmployeeController(employeeSvc employeeSvc.EmployeeServiceInterface, validator *validations.Validator) EmployeeControllerInterface {
	if validator == nil {
		validator = validations.New()
	}
	return &EmployeeController{BaseController: base.NewBaseController(validator), employeeSvc: employeeSvc}
}

// FindAll returns one page of employees (GET /employees). Query params: page
// (default 1), page_size (default 10, max 100), search (employee_code,
// full_name or email, case-insensitive), department_id (uuid) and is_active
// (true/false). Omitted or empty filters do not narrow the list; a malformed
// department_id or is_active is a 400 VALIDATION_FAILED. The envelope matches
// the department list: {"employees": [...], "pagination": {page, page_size,
// total, total_pages}}.
func (d *EmployeeController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())

	// A missing or non-numeric value parses to 0, which the service turns into
	// the default.
	page, _ := strconv.Atoi(ctx.Query("page"))
	pageSize, _ := strconv.Atoi(ctx.Query("page_size"))

	filter, err := parseEmployeeListFilter(ctx)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}

	result, err := d.employeeSvc.FindPage(baseService, page, pageSize, filter)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Employees fetched successfully", gin.H{
		"employees": result.Employees,
		"pagination": gin.H{
			"page":        result.Page,
			"page_size":   result.PageSize,
			"total":       result.Total,
			"total_pages": result.TotalPages,
		},
	})
}

// parseEmployeeListFilter reads the typed list filters of GET /employees.
func parseEmployeeListFilter(ctx *gin.Context) (employeeRepo.EmployeeListFilter, error) {
	filter := employeeRepo.EmployeeListFilter{Search: ctx.Query("search")}
	fields := map[string]string{}
	if raw := strings.TrimSpace(ctx.Query("department_id")); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			fields["department_id"] = apperrors.FieldInvalidFormat
		} else {
			filter.DepartmentID = &id
		}
	}
	if raw := strings.TrimSpace(ctx.Query("is_active")); raw != "" {
		active, err := strconv.ParseBool(raw)
		if err != nil {
			fields["is_active"] = apperrors.FieldInvalidFormat
		} else {
			filter.IsActive = &active
		}
	}
	if len(fields) > 0 {
		return filter, apperrors.ValidationFailed("", fields)
	}
	return filter, nil
}

func (d *EmployeeController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	employee, err := d.employeeSvc.FindById(baseService, id)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}
	response.OK(ctx, "Employee fetched successfully", employee)
}

// FindByCode returns one employee by employee_code (GET /employees/by-code/:code).
func (d *EmployeeController) FindByCode(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())

	employee, err := d.employeeSvc.FindByCode(baseService, ctx.Param("code"))
	if err != nil {
		d.RespondError(ctx, err)
		return
	}
	response.OK(ctx, "Employee fetched successfully", employee)
}

// Create adds an employee (POST /employees) and its login account; the admin's
// Authorization header is forwarded to auth-service for that. If the account
// cannot be created the employee is rolled back and an error is returned. The
// response carries "account": status "created" (temporary password, shown
// once), "linked" (existing account with that email, password unchanged).
func (d *EmployeeController) Create(ctx *gin.Context) {
	var req models.Employee
	if !d.ValidateRequest(ctx, &req) {
		return
	}
	reqCtx := authclient.WithAuthorization(ctx.Request.Context(), ctx.GetHeader("Authorization"))
	baseService := base.NewBaseService().WithContext(reqCtx)
	employee, err := d.employeeSvc.Create(baseService, &req)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}
	response.Created(ctx, "Employee created successfully", employee)
}
func (d *EmployeeController) NextCode(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	code, err := d.employeeSvc.NextEmployeeCode(baseService)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}
	response.OK(ctx, "Next employee code fetched successfully", gin.H{"employee_code": code})
}
// Update edits an employee (PUT /employees/:id). An is_active change is applied
// to the linked login account as well (Authorization is forwarded); if
// auth-service refuses (409 SEAT_LIMIT_REACHED or LAST_ADMIN) the employee is
// left unchanged and that 409 is returned with its data.
func (d *EmployeeController) Update(ctx *gin.Context) {
	var req models.Employee
	if !d.ValidateRequest(ctx, &req) {
		return
	}
	reqCtx := authclient.WithAuthorization(ctx.Request.Context(), ctx.GetHeader("Authorization"))
	baseService := base.NewBaseService().WithContext(reqCtx)
	id := ctx.Param("id")
	employee, err := d.employeeSvc.Update(baseService, id, &req)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}
	response.OK(ctx, "Employee updated successfully", employee)
}

// Delete removes an employee (DELETE /employees/:id). By default it is a soft
// delete that also deactivates the login account. ?hard=true deletes the
// employee and its login account permanently; 409 EMPLOYEE_IN_USE when other
// records still reference the employee. 409 LAST_ADMIN blocks both kinds.
func (d *EmployeeController) Delete(ctx *gin.Context) {
	hard := false
	if raw := strings.TrimSpace(ctx.Query("hard")); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			d.RespondError(ctx, apperrors.ValidationFailed("hard must be true or false.",
				map[string]string{"hard": apperrors.FieldInvalidFormat}))
			return
		}
		hard = parsed
	}
	reqCtx := authclient.WithAuthorization(ctx.Request.Context(), ctx.GetHeader("Authorization"))
	baseService := base.NewBaseService().WithContext(reqCtx)
	if err := d.employeeSvc.Delete(baseService, ctx.Param("id"), hard); err != nil {
		d.RespondError(ctx, err)
		return
	}
	message := "Employee deleted successfully"
	if hard {
		message = "Employee deleted permanently"
	}
	response.OK(ctx, message, nil)
}

// ProvisionLoginAccount creates or links the login account of an existing
// employee (POST /employees/:id/login-account, ADMIN only). Optional body
// {"access_role": "AUDITOR"}. Idempotent: status "exists" when already linked.
func (d *EmployeeController) ProvisionLoginAccount(ctx *gin.Context) {
	var req struct {
		AccessRole string `json:"access_role"`
	}
	if ctx.Request.ContentLength != 0 {
		if err := ctx.ShouldBindJSON(&req); err != nil && err != io.EOF {
			response.BadRequest(ctx, "Invalid request body")
			return
		}
	}
	reqCtx := authclient.WithAuthorization(ctx.Request.Context(), ctx.GetHeader("Authorization"))
	baseService := base.NewBaseService().WithContext(reqCtx)
	account, err := d.employeeSvc.ProvisionLoginAccount(baseService, ctx.Param("id"), req.AccessRole)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}
	response.OK(ctx, "Login account ready", gin.H{"account": account})
}
