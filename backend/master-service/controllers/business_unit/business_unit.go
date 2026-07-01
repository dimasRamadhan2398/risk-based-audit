package business_unit
import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	buSvc "master-service/services/business_unit"
	"github.com/gin-gonic/gin"
)
type BusinessUnitControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}
type BusinessUnitController struct {
	*base.BaseController
	buSvc buSvc.BusinessUnitServiceInterface
}
func NewBusinessUnitController(buSvc buSvc.BusinessUnitServiceInterface, validator *validations.Validator) BusinessUnitControllerInterface {
	return &BusinessUnitController{BaseController: base.NewBaseController(validator), buSvc: buSvc}
}
func (d *BusinessUnitController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	bus, err := d.buSvc.FindAll(baseService)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Business units fetched successfully", bus)
}
func (d *BusinessUnitController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	bu, err := d.buSvc.FindById(baseService, id)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Business unit fetched successfully", bu)
}
func (d *BusinessUnitController) Create(ctx *gin.Context) {
	var req models.BusinessUnit
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	bu, err := d.buSvc.Create(baseService, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.Created(ctx, "Business unit created successfully", bu)
}
func (d *BusinessUnitController) Update(ctx *gin.Context) {
	var req models.BusinessUnit
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	bu, err := d.buSvc.Update(baseService, id, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Business unit updated successfully", bu)
}
func (d *BusinessUnitController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.buSvc.Delete(baseService, id); err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Business unit deleted successfully", nil)
}
