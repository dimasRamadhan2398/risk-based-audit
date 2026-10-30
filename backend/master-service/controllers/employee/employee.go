package employee

import (
	"strconv"

	"master-service/models"
	"master-service/pkg/authclient"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	employeeSvc "master-service/services/employee"

	"github.com/gin-gonic/gin"
)

type EmployeeControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	FindByCode(ctx *gin.Context)
	Create(ctx *gin.Context)
	NextCode(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
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
// (default 1), page_size (default 10, max 100) and search (employee_code,
// full_name or email, case-insensitive). The envelope matches the department
// list: {"employees": [...], "pagination": {page, page_size, total, total_pages}}.
func (d *EmployeeController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())

	// A missing or non-numeric value parses to 0, which the service turns into
	// the default.
	page, _ := strconv.Atoi(ctx.Query("page"))
	pageSize, _ := strconv.Atoi(ctx.Query("page_size"))

	result, err := d.employeeSvc.FindPage(baseService, page, pageSize, ctx.Query("search"))
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

// Create adds an employee (POST /employees). An @aifl.co.id email also gets a
// login account; the admin's Authorization header is forwarded to auth-service
// for that. The response carries "account" then, with the temporary password
// once (status "created") or just the linked username (status "linked").
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
func (d *EmployeeController) Update(ctx *gin.Context) {
	var req models.Employee
	if !d.ValidateRequest(ctx, &req) {
		return
	}
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	employee, err := d.employeeSvc.Update(baseService, id, &req)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}
	response.OK(ctx, "Employee updated successfully", employee)
}
func (d *EmployeeController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.employeeSvc.Delete(baseService, id); err != nil {
		d.RespondError(ctx, err)
		return
	}
	response.OK(ctx, "Employee deleted successfully", nil)
}
