package risk_cause
import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	causeSvc "master-service/services/risk_cause"
	"github.com/gin-gonic/gin"
)
type RiskCauseControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}
type RiskCauseController struct {
	*base.BaseController
	causeSvc causeSvc.RiskCauseServiceInterface
}
func NewRiskCauseController(causeSvc causeSvc.RiskCauseServiceInterface, validator *validations.Validator) RiskCauseControllerInterface {
	return &RiskCauseController{BaseController: base.NewBaseController(validator), causeSvc: causeSvc}
}
func (d *RiskCauseController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	causes, err := d.causeSvc.FindAll(baseService)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risk causes fetched successfully", causes)
}
func (d *RiskCauseController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	cause, err := d.causeSvc.FindById(baseService, id)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risk cause fetched successfully", cause)
}
func (d *RiskCauseController) Create(ctx *gin.Context) {
	var req models.RiskCause
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	cause, err := d.causeSvc.Create(baseService, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.Created(ctx, "Risk cause created successfully", cause)
}
func (d *RiskCauseController) Update(ctx *gin.Context) {
	var req models.RiskCause
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	cause, err := d.causeSvc.Update(baseService, id, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risk cause updated successfully", cause)
}
func (d *RiskCauseController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.causeSvc.Delete(baseService, id); err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risk cause deleted successfully", nil)
}
