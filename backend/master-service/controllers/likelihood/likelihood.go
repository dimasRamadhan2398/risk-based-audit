package likelihood
import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	lSvc "master-service/services/likelihood"
	"github.com/gin-gonic/gin"
)
type LikelihoodControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}
type LikelihoodController struct {
	*base.BaseController
	lSvc lSvc.LikelihoodServiceInterface
}
func NewLikelihoodController(lSvc lSvc.LikelihoodServiceInterface, validator *validations.Validator) LikelihoodControllerInterface {
	return &LikelihoodController{BaseController: base.NewBaseController(validator), lSvc: lSvc}
}
func (d *LikelihoodController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	ls, err := d.lSvc.FindAll(baseService)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Likelihoods fetched successfully", ls)
}
func (d *LikelihoodController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	l, err := d.lSvc.FindById(baseService, id)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Likelihood fetched successfully", l)
}
func (d *LikelihoodController) Create(ctx *gin.Context) {
	var req models.Likelihood
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	l, err := d.lSvc.Create(baseService, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.Created(ctx, "Likelihood created successfully", l)
}
func (d *LikelihoodController) Update(ctx *gin.Context) {
	var req models.Likelihood
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	l, err := d.lSvc.Update(baseService, id, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Likelihood updated successfully", l)
}
func (d *LikelihoodController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.lSvc.Delete(baseService, id); err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Likelihood deleted successfully", nil)
}
