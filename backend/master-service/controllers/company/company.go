package company
import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	companySvc "master-service/services/company"
	"github.com/gin-gonic/gin"
)
type CompanyControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}
type CompanyController struct {
	*base.BaseController
	companySvc companySvc.CompanyServiceInterface
}
func NewCompanyController(companySvc companySvc.CompanyServiceInterface, validator *validations.Validator) CompanyControllerInterface {
	if validator == nil {
		validator = validations.New()
	}
	return &CompanyController{BaseController: base.NewBaseController(validator), companySvc: companySvc}
}
func (d *CompanyController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	companies, err := d.companySvc.FindAll(baseService)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Companies fetched successfully", companies)
}
func (d *CompanyController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	company, err := d.companySvc.FindById(baseService, id)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Company fetched successfully", company)
}
func (d *CompanyController) Create(ctx *gin.Context) {
	var req models.Company
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	company, err := d.companySvc.Create(baseService, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.Created(ctx, "Company created successfully", company)
}
func (d *CompanyController) Update(ctx *gin.Context) {
	var req models.Company
	if !d.ValidateRequest(ctx, &req) { return }
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	company, err := d.companySvc.Update(baseService, id, &req)
	if err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Company updated successfully", company)
}
func (d *CompanyController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.companySvc.Delete(baseService, id); err != nil { d.RespondError(ctx, err); return }
	response.OK(ctx, "Company deleted successfully", nil)
}
