package audit_issue
import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	issueSvc "master-service/services/audit_issue"
	"github.com/gin-gonic/gin"
)
type AuditIssueControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}
type AuditIssueController struct {
	*base.BaseController
	issueSvc issueSvc.AuditIssueServiceInterface
}
func NewAuditIssueController(issueSvc issueSvc.AuditIssueServiceInterface, validator *validations.Validator) AuditIssueControllerInterface {
	return &AuditIssueController{BaseController: base.NewBaseController(validator), issueSvc: issueSvc}
}
func (d *AuditIssueController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	issues, err := d.issueSvc.FindAll(baseService)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit issues fetched successfully", issues)
}
func (d *AuditIssueController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	issue, err := d.issueSvc.FindById(baseService, id)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit issue fetched successfully", issue)
}
func (d *AuditIssueController) Create(ctx *gin.Context) {
	var req models.AuditIssue
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	issue, err := d.issueSvc.Create(baseService, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.Created(ctx, "Audit issue created successfully", issue)
}
func (d *AuditIssueController) Update(ctx *gin.Context) {
	var req models.AuditIssue
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	issue, err := d.issueSvc.Update(baseService, id, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit issue updated successfully", issue)
}
func (d *AuditIssueController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.issueSvc.Delete(baseService, id); err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit issue deleted successfully", nil)
}
