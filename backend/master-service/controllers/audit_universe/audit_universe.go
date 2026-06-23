package audit_universe
import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	uSvc "master-service/services/audit_universe"
	"github.com/gin-gonic/gin"
)
type AuditUniverseControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}
type AuditUniverseController struct {
	*base.BaseController
	uSvc uSvc.AuditUniverseServiceInterface
}
func NewAuditUniverseController(uSvc uSvc.AuditUniverseServiceInterface, validator *validations.Validator) AuditUniverseControllerInterface {
	return &AuditUniverseController{BaseController: base.NewBaseController(validator), uSvc: uSvc}
}
func (d *AuditUniverseController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	us, err := d.uSvc.FindAll(baseService)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit universe entities fetched successfully", us)
}
func (d *AuditUniverseController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	u, err := d.uSvc.FindById(baseService, id)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit universe entity fetched successfully", u)
}
func (d *AuditUniverseController) Create(ctx *gin.Context) {
	var req models.AuditUniverse
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	u, err := d.uSvc.Create(baseService, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.Created(ctx, "Audit universe entity created successfully", u)
}
func (d *AuditUniverseController) Update(ctx *gin.Context) {
	var req models.AuditUniverse
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	u, err := d.uSvc.Update(baseService, id, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit universe entity updated successfully", u)
}
func (d *AuditUniverseController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.uSvc.Delete(baseService, id); err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit universe entity deleted successfully", nil)
}
