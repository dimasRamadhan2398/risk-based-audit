package audit_recommendation
import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	recSvc "master-service/services/audit_recommendation"
	"github.com/gin-gonic/gin"
)
type AuditRecommendationControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}
type AuditRecommendationController struct {
	*base.BaseController
	recSvc recSvc.AuditRecommendationServiceInterface
}
func NewAuditRecommendationController(recSvc recSvc.AuditRecommendationServiceInterface, validator *validations.Validator) AuditRecommendationControllerInterface {
	return &AuditRecommendationController{BaseController: base.NewBaseController(validator), recSvc: recSvc}
}
func (d *AuditRecommendationController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	recs, err := d.recSvc.FindAll(baseService)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit recommendations fetched successfully", recs)
}
func (d *AuditRecommendationController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	rec, err := d.recSvc.FindById(baseService, id)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit recommendation fetched successfully", rec)
}
func (d *AuditRecommendationController) Create(ctx *gin.Context) {
	var req models.AuditRecommendation
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	rec, err := d.recSvc.Create(baseService, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.Created(ctx, "Audit recommendation created successfully", rec)
}
func (d *AuditRecommendationController) Update(ctx *gin.Context) {
	var req models.AuditRecommendation
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	rec, err := d.recSvc.Update(baseService, id, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit recommendation updated successfully", rec)
}
func (d *AuditRecommendationController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.recSvc.Delete(baseService, id); err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit recommendation deleted successfully", nil)
}
