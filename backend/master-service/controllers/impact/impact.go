package impact
import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	iSvc "master-service/services/impact"
	"github.com/gin-gonic/gin"
)
type ImpactControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}
type ImpactController struct {
	*base.BaseController
	iSvc iSvc.ImpactServiceInterface
}
func NewImpactController(iSvc iSvc.ImpactServiceInterface, validator *validations.Validator) ImpactControllerInterface {
	return &ImpactController{BaseController: base.NewBaseController(validator), iSvc: iSvc}
}
func (d *ImpactController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	is, err := d.iSvc.FindAll(baseService)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Impacts fetched successfully", is)
}
func (d *ImpactController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	i, err := d.iSvc.FindById(baseService, id)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Impact fetched successfully", i)
}
func (d *ImpactController) Create(ctx *gin.Context) {
	var req models.Impact
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	i, err := d.iSvc.Create(baseService, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.Created(ctx, "Impact created successfully", i)
}
func (d *ImpactController) Update(ctx *gin.Context) {
	var req models.Impact
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	i, err := d.iSvc.Update(baseService, id, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Impact updated successfully", i)
}
func (d *ImpactController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.iSvc.Delete(baseService, id); err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Impact deleted successfully", nil)
}
