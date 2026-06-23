package control
import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	controlSvc "master-service/services/control"
	"github.com/gin-gonic/gin"
)
type ControlControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}
type ControlController struct {
	*base.BaseController
	controlSvc controlSvc.ControlServiceInterface
}
func NewControlController(controlSvc controlSvc.ControlServiceInterface, validator *validations.Validator) ControlControllerInterface {
	return &ControlController{BaseController: base.NewBaseController(validator), controlSvc: controlSvc}
}
func (d *ControlController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	controls, err := d.controlSvc.FindAll(baseService)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Controls fetched successfully", controls)
}
func (d *ControlController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	control, err := d.controlSvc.FindById(baseService, id)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Control fetched successfully", control)
}
func (d *ControlController) Create(ctx *gin.Context) {
	var req models.Control
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	control, err := d.controlSvc.Create(baseService, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.Created(ctx, "Control created successfully", control)
}
func (d *ControlController) Update(ctx *gin.Context) {
	var req models.Control
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	control, err := d.controlSvc.Update(baseService, id, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Control updated successfully", control)
}
func (d *ControlController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.controlSvc.Delete(baseService, id); err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Control deleted successfully", nil)
}
