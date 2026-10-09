package company

import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	companySvc "master-service/services/company"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateCompanyRequest struct {
	CompanyCode string  `json:"company_code"`
	Code        string  `json:"code"`
	CompanyName string  `json:"company_name"`
	Name        string  `json:"name"`
	LegalName   string  `json:"legal_name"`
	TaxID       string  `json:"tax_id"`
	CompanyType string  `json:"company_type"`
	ParentID    *string `json:"parent_id"`
	LocationID  *string `json:"location_id"`
	Phone       string  `json:"phone"`
	Email       string  `json:"email"`
	Website     string  `json:"website"`
	IsActive    *bool   `json:"is_active"`
}

type UpdateCompanyRequest struct {
	CompanyCode *string `json:"company_code"`
	Code        *string `json:"code"`
	CompanyName *string `json:"company_name"`
	Name        *string `json:"name"`
	LegalName   *string `json:"legal_name"`
	TaxID       *string `json:"tax_id"`
	CompanyType *string `json:"company_type"`
	ParentID    *string `json:"parent_id"`
	LocationID  *string `json:"location_id"`
	Phone       *string `json:"phone"`
	Email       *string `json:"email"`
	Website     *string `json:"website"`
	IsActive    *bool   `json:"is_active"`
}

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
	if err != nil {
		d.RespondError(ctx, err)
		return
	}
	response.OK(ctx, "Companies fetched successfully", companies)
}
func (d *CompanyController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	company, err := d.companySvc.FindById(baseService, id)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}
	response.OK(ctx, "Company fetched successfully", company)
}
func (d *CompanyController) Create(ctx *gin.Context) {
	var req CreateCompanyRequest
	if !d.ValidateRequest(ctx, &req) {
		return
	}

	code := strings.TrimSpace(req.CompanyCode)
	if code == "" {
		code = strings.TrimSpace(req.Code)
	}
	if code == "" {
		response.BadRequest(ctx, "Company code is required")
		return
	}

	name := strings.TrimSpace(req.CompanyName)
	if name == "" {
		name = strings.TrimSpace(req.Name)
	}
	if name == "" {
		response.BadRequest(ctx, "Company name is required")
		return
	}

	compType := models.CompanyType(strings.ToUpper(strings.TrimSpace(req.CompanyType)))
	if compType == "" {
		compType = models.CompanyTypeSubsidiary
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	company := &models.Company{
		CompanyCode: code,
		CompanyName: name,
		LegalName:   strings.TrimSpace(req.LegalName),
		TaxID:       strings.TrimSpace(req.TaxID),
		CompanyType: compType,
		Phone:       strings.TrimSpace(req.Phone),
		Email:       strings.TrimSpace(req.Email),
		Website:     strings.TrimSpace(req.Website),
		IsActive:    isActive,
	}

	if req.ParentID != nil && strings.TrimSpace(*req.ParentID) != "" {
		if pid, err := uuid.Parse(strings.TrimSpace(*req.ParentID)); err == nil {
			company.ParentID = &pid
		}
	}
	if req.LocationID != nil && strings.TrimSpace(*req.LocationID) != "" {
		if lid, err := uuid.Parse(strings.TrimSpace(*req.LocationID)); err == nil {
			company.LocationID = &lid
		}
	}

	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	result, err := d.companySvc.Create(baseService, company)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}
	response.Created(ctx, "Company created successfully", result)
}
func (d *CompanyController) Update(ctx *gin.Context) {
	var req UpdateCompanyRequest
	if !d.ValidateRequest(ctx, &req) {
		return
	}
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")

	existingCompany, err := d.companySvc.FindById(baseService, id)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}

	if req.CompanyCode != nil && strings.TrimSpace(*req.CompanyCode) != "" {
		existingCompany.CompanyCode = strings.TrimSpace(*req.CompanyCode)
	} else if req.Code != nil && strings.TrimSpace(*req.Code) != "" {
		existingCompany.CompanyCode = strings.TrimSpace(*req.Code)
	}

	if req.CompanyName != nil && strings.TrimSpace(*req.CompanyName) != "" {
		existingCompany.CompanyName = strings.TrimSpace(*req.CompanyName)
	} else if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
		existingCompany.CompanyName = strings.TrimSpace(*req.Name)
	}

	if req.LegalName != nil {
		existingCompany.LegalName = strings.TrimSpace(*req.LegalName)
	}
	if req.TaxID != nil {
		existingCompany.TaxID = strings.TrimSpace(*req.TaxID)
	}
	if req.CompanyType != nil && strings.TrimSpace(*req.CompanyType) != "" {
		existingCompany.CompanyType = models.CompanyType(strings.ToUpper(strings.TrimSpace(*req.CompanyType)))
	}
	if req.Phone != nil {
		existingCompany.Phone = strings.TrimSpace(*req.Phone)
	}
	if req.Email != nil {
		existingCompany.Email = strings.TrimSpace(*req.Email)
	}
	if req.Website != nil {
		existingCompany.Website = strings.TrimSpace(*req.Website)
	}
	if req.IsActive != nil {
		existingCompany.IsActive = *req.IsActive
	}

	if req.ParentID != nil {
		if strings.TrimSpace(*req.ParentID) == "" {
			existingCompany.ParentID = nil
		} else if pid, err := uuid.Parse(strings.TrimSpace(*req.ParentID)); err == nil {
			existingCompany.ParentID = &pid
		}
	}
	if req.LocationID != nil {
		if strings.TrimSpace(*req.LocationID) == "" {
			existingCompany.LocationID = nil
		} else if lid, err := uuid.Parse(strings.TrimSpace(*req.LocationID)); err == nil {
			existingCompany.LocationID = &lid
		}
	}

	result, err := d.companySvc.Update(baseService, id, existingCompany)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}
	response.OK(ctx, "Company updated successfully", result)
}
func (d *CompanyController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")
	if err := d.companySvc.Delete(baseService, id); err != nil {
		d.RespondError(ctx, err)
		return
	}
	response.OK(ctx, "Company deleted successfully", nil)
}
