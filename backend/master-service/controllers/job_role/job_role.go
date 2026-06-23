package job_role
import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	roleSvc "master-service/services/job_role"
	"github.com/gin-gonic/gin"
)
type JobRoleControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}
type JobRoleController struct {
	*base.BaseController
	roleSvc roleSvc.JobRoleServiceInterface
}
func NewJobRoleController(roleSvc roleSvc.JobRoleServiceInterface, validator *validations.Validator) JobRoleControllerInterface {
	return &JobRoleController{BaseController: base.NewBaseController(validator), roleSvc: roleSvc}
}
func (d *JobRoleController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	roles, err := d.roleSvc.FindAll(baseService)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Job roles fetched successfully", roles)
}
func (d *JobRoleController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	role, err := d.roleSvc.FindById(baseService, id)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Job role fetched successfully", role)
}
func (d *JobRoleController) Create(ctx *gin.Context) {
	var req models.JobRole
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	role, err := d.roleSvc.Create(baseService, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.Created(ctx, "Job role created successfully", role)
}
func (d *JobRoleController) Update(ctx *gin.Context) {
	var req models.JobRole
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	role, err := d.roleSvc.Update(baseService, id, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Job role updated successfully", role)
}
func (d *JobRoleController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.roleSvc.Delete(baseService, id); err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Job role deleted successfully", nil)
}
