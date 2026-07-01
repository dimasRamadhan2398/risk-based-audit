package employee
import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	employeeSvc "master-service/services/employee"
	"github.com/gin-gonic/gin"
)
type EmployeeControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
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
func (d *EmployeeController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	employees, err := d.employeeSvc.FindAll(baseService)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Employees fetched successfully", employees)
}
func (d *EmployeeController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	employee, err := d.employeeSvc.FindById(baseService, id)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Employee fetched successfully", employee)
}
func (d *EmployeeController) Create(ctx *gin.Context) {
	var req models.Employee
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	employee, err := d.employeeSvc.Create(baseService, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.Created(ctx, "Employee created successfully", employee)
}
func (d *EmployeeController) Update(ctx *gin.Context) {
	var req models.Employee
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	employee, err := d.employeeSvc.Update(baseService, id, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Employee updated successfully", employee)
}
func (d *EmployeeController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.employeeSvc.Delete(baseService, id); err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Employee deleted successfully", nil)
}
