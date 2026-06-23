package risk_category

import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	catSvc "master-service/services/risk_category"

	"github.com/gin-gonic/gin"
)

type RiskCategoryControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

type RiskCategoryController struct {
	*base.BaseController
	catSvc catSvc.RiskCategoryServiceInterface
}

func NewRiskCategoryController(catSvc catSvc.RiskCategoryServiceInterface, validator *validations.Validator) RiskCategoryControllerInterface {
	return &RiskCategoryController{
		BaseController: base.NewBaseController(validator),
		catSvc:         catSvc,
	}
}

func (d *RiskCategoryController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())

	cats, err := d.catSvc.FindAll(baseService)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Risk categories fetched successfully", cats)
}

func (d *RiskCategoryController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")

	cat, err := d.catSvc.FindById(baseService, id)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Risk category fetched successfully", cat)
}

func (d *RiskCategoryController) Create(ctx *gin.Context) {
	var req models.RiskCategory
	if !d.ValidateRequest(ctx, &req) {
		return
	}

	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	cat, err := d.catSvc.Create(baseService, &req)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}

	response.Created(ctx, "Risk category created successfully", cat)
}

func (d *RiskCategoryController) Update(ctx *gin.Context) {
	var req models.RiskCategory
	if !d.ValidateRequest(ctx, &req) {
		return
	}

	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")

	cat, err := d.catSvc.Update(baseService, id, &req)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Risk category updated successfully", cat)
}

func (d *RiskCategoryController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")

	if err := d.catSvc.Delete(baseService, id); err != nil {
		d.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Risk category deleted successfully", nil)
}
