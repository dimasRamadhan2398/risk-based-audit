package control_assessment
import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	assessmentSvc "master-service/services/control_assessment"
	"github.com/gin-gonic/gin"
)
type ControlAssessmentControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}
type ControlAssessmentController struct {
	*base.BaseController
	assessmentSvc assessmentSvc.ControlAssessmentServiceInterface
}
func NewControlAssessmentController(assessmentSvc assessmentSvc.ControlAssessmentServiceInterface, validator *validations.Validator) ControlAssessmentControllerInterface {
	return &ControlAssessmentController{BaseController: base.NewBaseController(validator), assessmentSvc: assessmentSvc}
}
func (d *ControlAssessmentController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	assessments, err := d.assessmentSvc.FindAll(baseService)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Control assessments fetched successfully", assessments)
}
func (d *ControlAssessmentController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	assessment, err := d.assessmentSvc.FindById(baseService, id)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Control assessment fetched successfully", assessment)
}
func (d *ControlAssessmentController) Create(ctx *gin.Context) {
	var req models.ControlAssessment
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	assessment, err := d.assessmentSvc.Create(baseService, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.Created(ctx, "Control assessment created successfully", assessment)
}
func (d *ControlAssessmentController) Update(ctx *gin.Context) {
	var req models.ControlAssessment
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	assessment, err := d.assessmentSvc.Update(baseService, id, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Control assessment updated successfully", assessment)
}
func (d *ControlAssessmentController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.assessmentSvc.Delete(baseService, id); err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Control assessment deleted successfully", nil)
}
