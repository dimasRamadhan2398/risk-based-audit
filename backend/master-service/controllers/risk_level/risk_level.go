package risk_level
import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	rlSvc "master-service/services/risk_level"
	"github.com/gin-gonic/gin"
)
type RiskLevelControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}
type RiskLevelController struct {
	*base.BaseController
	rlSvc rlSvc.RiskLevelServiceInterface
}
func NewRiskLevelController(rlSvc rlSvc.RiskLevelServiceInterface, validator *validations.Validator) RiskLevelControllerInterface {
	return &RiskLevelController{BaseController: base.NewBaseController(validator), rlSvc: rlSvc}
}
func (d *RiskLevelController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	rls, err := d.rlSvc.FindAll(baseService)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risk levels fetched successfully", rls)
}
func (d *RiskLevelController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	rl, err := d.rlSvc.FindById(baseService, id)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risk level fetched successfully", rl)
}
func (d *RiskLevelController) Create(ctx *gin.Context) {
	var req models.RiskLevel
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	rl, err := d.rlSvc.Create(baseService, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.Created(ctx, "Risk level created successfully", rl)
}
func (d *RiskLevelController) Update(ctx *gin.Context) {
	var req models.RiskLevel
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	rl, err := d.rlSvc.Update(baseService, id, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risk level updated successfully", rl)
}
func (d *RiskLevelController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.rlSvc.Delete(baseService, id); err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risk level deleted successfully", nil)
}
