package department

import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	departmentSvc "master-service/services/department"

	"github.com/gin-gonic/gin"
)

type DepartmentControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

type DepartmentController struct {
	*base.BaseController
	departmentSvc departmentSvc.DepartmentServiceInterface
}

func NewDepartmentController(departmentSvc departmentSvc.DepartmentServiceInterface, validator *validations.Validator) DepartmentControllerInterface {
	return &DepartmentController{
		BaseController: base.NewBaseController(validator),
		departmentSvc:  departmentSvc,
	}
}

func (d *DepartmentController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())

	departments, err := d.departmentSvc.FindAll(baseService)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Departments fetched successfully", departments)
}

func (d *DepartmentController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")

	department, err := d.departmentSvc.FindById(baseService, id)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Department fetched successfully", department)
}

func (d *DepartmentController) Create(ctx *gin.Context) {
	var req models.Department
	if !d.ValidateRequest(ctx, &req) {
		return
	}

	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	department, err := d.departmentSvc.Create(baseService, &req)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}

	response.Created(ctx, "Department created successfully", department)
}

func (d *DepartmentController) Update(ctx *gin.Context) {
	var req models.Department
	if !d.ValidateRequest(ctx, &req) {
		return
	}

	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")

	department, err := d.departmentSvc.Update(baseService, id, &req)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Department updated successfully", department)
}

func (d *DepartmentController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")

	if err := d.departmentSvc.Delete(baseService, id); err != nil {
		d.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Department deleted successfully", nil)
}
