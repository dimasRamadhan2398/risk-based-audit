package audit_period
import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	pSvc "master-service/services/audit_period"
	"github.com/gin-gonic/gin"
)
type AuditPeriodControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}
type AuditPeriodController struct {
	*base.BaseController
	pSvc pSvc.AuditPeriodServiceInterface
}
func NewAuditPeriodController(pSvc pSvc.AuditPeriodServiceInterface, validator *validations.Validator) AuditPeriodControllerInterface {
	return &AuditPeriodController{BaseController: base.NewBaseController(validator), pSvc: pSvc}
}
func (d *AuditPeriodController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	ps, err := d.pSvc.FindAll(baseService)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit periods fetched successfully", ps)
}
func (d *AuditPeriodController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	p, err := d.pSvc.FindById(baseService, id)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit period fetched successfully", p)
}
func (d *AuditPeriodController) Create(ctx *gin.Context) {
	var req models.AuditPeriod
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	p, err := d.pSvc.Create(baseService, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.Created(ctx, "Audit period created successfully", p)
}
func (d *AuditPeriodController) Update(ctx *gin.Context) {
	var req models.AuditPeriod
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	p, err := d.pSvc.Update(baseService, id, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit period updated successfully", p)
}
func (d *AuditPeriodController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.pSvc.Delete(baseService, id); err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit period deleted successfully", nil)
}
