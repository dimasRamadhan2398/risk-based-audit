package risk_register
import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	riskSvc "master-service/services/risk_register"
	"github.com/gin-gonic/gin"
)
type RiskRegisterControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}
type RiskRegisterController struct {
	*base.BaseController
	riskSvc riskSvc.RiskRegisterServiceInterface
}
func NewRiskRegisterController(riskSvc riskSvc.RiskRegisterServiceInterface, validator *validations.Validator) RiskRegisterControllerInterface {
	return &RiskRegisterController{BaseController: base.NewBaseController(validator), riskSvc: riskSvc}
}
func (d *RiskRegisterController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	risks, err := d.riskSvc.FindAll(baseService)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risks fetched successfully", risks)
}
func (d *RiskRegisterController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	risk, err := d.riskSvc.FindById(baseService, id)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risk fetched successfully", risk)
}
func (d *RiskRegisterController) Create(ctx *gin.Context) {
	var req models.RiskRegister
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	risk, err := d.riskSvc.Create(baseService, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.Created(ctx, "Risk created successfully", risk)
}
func (d *RiskRegisterController) Update(ctx *gin.Context) {
	var req models.RiskRegister
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	risk, err := d.riskSvc.Update(baseService, id, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risk updated successfully", risk)
}
func (d *RiskRegisterController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.riskSvc.Delete(baseService, id); err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risk deleted successfully", nil)
}
