package risk_matrix_cell
import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	cellSvc "master-service/services/risk_matrix_cell"
	"github.com/gin-gonic/gin"
)
type RiskMatrixCellControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}
type RiskMatrixCellController struct {
	*base.BaseController
	cellSvc cellSvc.RiskMatrixCellServiceInterface
}
func NewRiskMatrixCellController(cellSvc cellSvc.RiskMatrixCellServiceInterface, validator *validations.Validator) RiskMatrixCellControllerInterface {
	return &RiskMatrixCellController{BaseController: base.NewBaseController(validator), cellSvc: cellSvc}
}
func (d *RiskMatrixCellController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	cells, err := d.cellSvc.FindAll(baseService)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risk matrix cells fetched successfully", cells)
}
func (d *RiskMatrixCellController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	cell, err := d.cellSvc.FindById(baseService, id)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risk matrix cell fetched successfully", cell)
}
func (d *RiskMatrixCellController) Create(ctx *gin.Context) {
	var req models.RiskMatrixCell
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	cell, err := d.cellSvc.Create(baseService, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.Created(ctx, "Risk matrix cell created successfully", cell)
}
func (d *RiskMatrixCellController) Update(ctx *gin.Context) {
	var req models.RiskMatrixCell
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	cell, err := d.cellSvc.Update(baseService, id, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risk matrix cell updated successfully", cell)
}
func (d *RiskMatrixCellController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.cellSvc.Delete(baseService, id); err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Risk matrix cell deleted successfully", nil)
}
