package risk_effect
import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	effectSvc "master-service/services/risk_effect"
	"github.com/gin-gonic/gin"
)
type RiskEffectControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}
type RiskEffectController struct {
	*base.BaseController
	effectSvc effectSvc.RiskEffectServiceInterface
}
func NewRiskEffectController(effectSvc effectSvc.RiskEffectServiceInterface, validator *validations.Validator) RiskEffectControllerInterface {
	return &RiskEffectController{BaseController: base.NewBaseController(validator), effectSvc: effectSvc}
}
func (d *RiskEffectController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	effects, err := d.effectSvc.FindAll(baseService)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risk effects fetched successfully", effects)
}
func (d *RiskEffectController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	effect, err := d.effectSvc.FindById(baseService, id)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risk effect fetched successfully", effect)
}
func (d *RiskEffectController) Create(ctx *gin.Context) {
	var req models.RiskEffect
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	effect, err := d.effectSvc.Create(baseService, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.Created(ctx, "Risk effect created successfully", effect)
}
func (d *RiskEffectController) Update(ctx *gin.Context) {
	var req models.RiskEffect
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	effect, err := d.effectSvc.Update(baseService, id, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risk effect updated successfully", effect)
}
func (d *RiskEffectController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.effectSvc.Delete(baseService, id); err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risk effect deleted successfully", nil)
}
