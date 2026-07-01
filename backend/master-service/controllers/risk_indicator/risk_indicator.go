package risk_indicator
import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	indicatorSvc "master-service/services/risk_indicator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)
type RiskIndicatorControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
	AddLog(ctx *gin.Context)
	GetLogs(ctx *gin.Context)
}
type RiskIndicatorController struct {
	*base.BaseController
	indicatorSvc indicatorSvc.RiskIndicatorServiceInterface
}
func NewRiskIndicatorController(indicatorSvc indicatorSvc.RiskIndicatorServiceInterface, validator *validations.Validator) RiskIndicatorControllerInterface {
	return &RiskIndicatorController{BaseController: base.NewBaseController(validator), indicatorSvc: indicatorSvc}
}
func (d *RiskIndicatorController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	indicators, err := d.indicatorSvc.FindAll(baseService)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risk indicators fetched successfully", indicators)
}
func (d *RiskIndicatorController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	indicator, err := d.indicatorSvc.FindById(baseService, id)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risk indicator fetched successfully", indicator)
}
func (d *RiskIndicatorController) Create(ctx *gin.Context) {
	var req models.RiskIndicator
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	indicator, err := d.indicatorSvc.Create(baseService, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.Created(ctx, "Risk indicator created successfully", indicator)
}
func (d *RiskIndicatorController) Update(ctx *gin.Context) {
	var req models.RiskIndicator
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	indicator, err := d.indicatorSvc.Update(baseService, id, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risk indicator updated successfully", indicator)
}
func (d *RiskIndicatorController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.indicatorSvc.Delete(baseService, id); err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risk indicator deleted successfully", nil)
}
func (d *RiskIndicatorController) AddLog(ctx *gin.Context) {
	idStr := ctx.Param("id")
	indicatorID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(ctx, "Invalid indicator ID format")
		return
	}

	var req models.RiskIndicatorLog
	if !d.ValidateRequest(ctx, &req) { return }
	req.IndicatorID = indicatorID

	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	log, err := d.indicatorSvc.AddLog(baseService, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.Created(ctx, "Risk indicator log added successfully", log)
}
func (d *RiskIndicatorController) GetLogs(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	logs, err := d.indicatorSvc.GetLogs(baseService, id)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risk indicator logs fetched successfully", logs)
}
