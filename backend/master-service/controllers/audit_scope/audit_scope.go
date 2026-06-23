package audit_scope
import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	scopeSvc "master-service/services/audit_scope"
	"github.com/gin-gonic/gin"
)
type AuditScopeControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}
type AuditScopeController struct {
	*base.BaseController
	scopeSvc scopeSvc.AuditScopeServiceInterface
}
func NewAuditScopeController(scopeSvc scopeSvc.AuditScopeServiceInterface, validator *validations.Validator) AuditScopeControllerInterface {
	return &AuditScopeController{BaseController: base.NewBaseController(validator), scopeSvc: scopeSvc}
}
func (d *AuditScopeController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	scopes, err := d.scopeSvc.FindAll(baseService)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit scopes fetched successfully", scopes)
}
func (d *AuditScopeController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	scope, err := d.scopeSvc.FindById(baseService, id)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit scope fetched successfully", scope)
}
func (d *AuditScopeController) Create(ctx *gin.Context) {
	var req models.AuditScope
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	scope, err := d.scopeSvc.Create(baseService, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.Created(ctx, "Audit scope created successfully", scope)
}
func (d *AuditScopeController) Update(ctx *gin.Context) {
	var req models.AuditScope
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	scope, err := d.scopeSvc.Update(baseService, id, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit scope updated successfully", scope)
}
func (d *AuditScopeController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.scopeSvc.Delete(baseService, id); err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit scope deleted successfully", nil)
}
