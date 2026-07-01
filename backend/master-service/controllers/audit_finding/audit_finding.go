package audit_finding
import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	findingSvc "master-service/services/audit_finding"
	"github.com/gin-gonic/gin"
)
type AuditFindingControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}
type AuditFindingController struct {
	*base.BaseController
	findingSvc findingSvc.AuditFindingServiceInterface
}
func NewAuditFindingController(findingSvc findingSvc.AuditFindingServiceInterface, validator *validations.Validator) AuditFindingControllerInterface {
	return &AuditFindingController{BaseController: base.NewBaseController(validator), findingSvc: findingSvc}
}
func (d *AuditFindingController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	findings, err := d.findingSvc.FindAll(baseService)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit findings fetched successfully", findings)
}
func (d *AuditFindingController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	finding, err := d.findingSvc.FindById(baseService, id)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit finding fetched successfully", finding)
}
func (d *AuditFindingController) Create(ctx *gin.Context) {
	var req models.AuditFinding
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	finding, err := d.findingSvc.Create(baseService, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.Created(ctx, "Audit finding created successfully", finding)
}
func (d *AuditFindingController) Update(ctx *gin.Context) {
	var req models.AuditFinding
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	finding, err := d.findingSvc.Update(baseService, id, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit finding updated successfully", finding)
}
func (d *AuditFindingController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.findingSvc.Delete(baseService, id); err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit finding deleted successfully", nil)
}
