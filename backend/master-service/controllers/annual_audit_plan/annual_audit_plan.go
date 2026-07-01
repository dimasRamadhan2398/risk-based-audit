package annual_audit_plan
import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	planSvc "master-service/services/annual_audit_plan"
	"github.com/gin-gonic/gin"
)
type AnnualAuditPlanControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}
type AnnualAuditPlanController struct {
	*base.BaseController
	planSvc planSvc.AnnualAuditPlanServiceInterface
}
func NewAnnualAuditPlanController(planSvc planSvc.AnnualAuditPlanServiceInterface, validator *validations.Validator) AnnualAuditPlanControllerInterface {
	return &AnnualAuditPlanController{BaseController: base.NewBaseController(validator), planSvc: planSvc}
}
func (d *AnnualAuditPlanController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	plans, err := d.planSvc.FindAll(baseService)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit plans fetched successfully", plans)
}
func (d *AnnualAuditPlanController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	plan, err := d.planSvc.FindById(baseService, id)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit plan fetched successfully", plan)
}
func (d *AnnualAuditPlanController) Create(ctx *gin.Context) {
	var req models.AnnualAuditPlan
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	plan, err := d.planSvc.Create(baseService, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.Created(ctx, "Audit plan created successfully", plan)
}
func (d *AnnualAuditPlanController) Update(ctx *gin.Context) {
	var req models.AnnualAuditPlan
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	plan, err := d.planSvc.Update(baseService, id, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit plan updated successfully", plan)
}
func (d *AnnualAuditPlanController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.planSvc.Delete(baseService, id); err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit plan deleted successfully", nil)
}
