package audit_workpaper
import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	wpSvc "master-service/services/audit_workpaper"
	"github.com/gin-gonic/gin"
)
type AuditWorkpaperControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}
type AuditWorkpaperController struct {
	*base.BaseController
	wpSvc wpSvc.AuditWorkpaperServiceInterface
}
func NewAuditWorkpaperController(wpSvc wpSvc.AuditWorkpaperServiceInterface, validator *validations.Validator) AuditWorkpaperControllerInterface {
	return &AuditWorkpaperController{BaseController: base.NewBaseController(validator), wpSvc: wpSvc}
}
func (d *AuditWorkpaperController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	wps, err := d.wpSvc.FindAll(baseService)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit workpapers fetched successfully", wps)
}
func (d *AuditWorkpaperController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	wp, err := d.wpSvc.FindById(baseService, id)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit workpaper fetched successfully", wp)
}
func (d *AuditWorkpaperController) Create(ctx *gin.Context) {
	var req models.AuditWorkpaper
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	wp, err := d.wpSvc.Create(baseService, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.Created(ctx, "Audit workpaper created successfully", wp)
}
func (d *AuditWorkpaperController) Update(ctx *gin.Context) {
	var req models.AuditWorkpaper
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	wp, err := d.wpSvc.Update(baseService, id, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit workpaper updated successfully", wp)
}
func (d *AuditWorkpaperController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.wpSvc.Delete(baseService, id); err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Audit workpaper deleted successfully", nil)
}
