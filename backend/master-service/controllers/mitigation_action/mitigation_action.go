package mitigation_action
import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	actionSvc "master-service/services/mitigation_action"
	"github.com/gin-gonic/gin"
)
type MitigationActionControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}
type MitigationActionController struct {
	*base.BaseController
	actionSvc actionSvc.MitigationActionServiceInterface
}
func NewMitigationActionController(actionSvc actionSvc.MitigationActionServiceInterface, validator *validations.Validator) MitigationActionControllerInterface {
	return &MitigationActionController{BaseController: base.NewBaseController(validator), actionSvc: actionSvc}
}
func (d *MitigationActionController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	actions, err := d.actionSvc.FindAll(baseService)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Mitigation actions fetched successfully", actions)
}
func (d *MitigationActionController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	action, err := d.actionSvc.FindById(baseService, id)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Mitigation action fetched successfully", action)
}
func (d *MitigationActionController) Create(ctx *gin.Context) {
	var req models.MitigationAction
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	action, err := d.actionSvc.Create(baseService, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.Created(ctx, "Mitigation action created successfully", action)
}
func (d *MitigationActionController) Update(ctx *gin.Context) {
	var req models.MitigationAction
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	action, err := d.actionSvc.Update(baseService, id, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Mitigation action updated successfully", action)
}
func (d *MitigationActionController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.actionSvc.Delete(baseService, id); err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Mitigation action deleted successfully", nil)
}
