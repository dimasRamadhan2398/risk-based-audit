package company

import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	companySvc "master-service/services/company"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Request DTOs
type CreateCompanyRequest struct {
	CompanyCode    string  `json:"company_code" binding:"required"`
	CompanyName    string  `json:"company_name" binding:"required"`
	LegalName      string  `json:"legal_name"`
	TaxID          string  `json:"tax_id"`
	CompanyType    string  `json:"company_type" binding:"required"`
	ParentID       *string `json:"parent_id"`
	LocationID     *string `json:"location_id"`
	Phone          string  `json:"phone"`
	Email          string  `json:"email"`
	Website        string  `json:"website"`
	IsActive       bool    `json:"is_active"`
	EstablishedAt  string  `json:"established_at"`
}

type UpdateCompanyRequest struct {
	CompanyName    *string `json:"company_name"`
	LegalName      *string `json:"legal_name"`
	TaxID          *string `json:"tax_id"`
	CompanyType    *string `json:"company_type"`
	ParentID       *string `json:"parent_id"`
	LocationID     *string `json:"location_id"`
	Phone          *string `json:"phone"`
	Email          *string `json:"email"`
	Website        *string `json:"website"`
	IsActive       *bool   `json:"is_active"`
	EstablishedAt  *string `json:"established_at"`
}

type CompanyControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	List(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

type CompanyController struct {
	*base.BaseController
	companySvc companySvc.CompanyServiceInterface
}

func NewCompanyController(companySvc companySvc.CompanyServiceInterface, validator *validations.Validator) CompanyControllerInterface {
	return &CompanyController{
		BaseController: base.NewBaseController(validator),
		companySvc:     companySvc,
	}
}

func (c *CompanyController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())

	companies, err := c.companySvc.FindAll(baseService)
	if err != nil {
		c.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Companies fetched successfully", companies)
}

func (c *CompanyController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")

	company, err := c.companySvc.FindById(baseService, id)
	if err != nil {
		c.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Company fetched successfully", company)
}

func (c *CompanyController) List(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())

	// Parse query params
	page := 1
	pageSize := 10
	if p := ctx.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := ctx.Query("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}
	search := ctx.Query("search")

	offset := (page - 1) * pageSize

	companies, total, err := c.companySvc.FindMany(baseService, offset, pageSize, search)
	if err != nil {
		c.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Companies fetched successfully", gin.H{
		"companies": companies,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

func (c *CompanyController) Create(ctx *gin.Context) {
	var req CreateCompanyRequest
	if !c.ValidateRequest(ctx, &req) {
		return
	}

	baseService := base.NewBaseService().WithContext(ctx.Request.Context())

	// Convert request to model
	company := &models.Company{
		CompanyCode: req.CompanyCode,
		CompanyName: req.CompanyName,
		LegalName:   req.LegalName,
		TaxID:       req.TaxID,
		CompanyType: models.CompanyType(req.CompanyType),
		Phone:       req.Phone,
		Email:       req.Email,
		Website:     req.Website,
		IsActive:    req.IsActive,
	}

	// Parse parent_id if provided
	if req.ParentID != nil && *req.ParentID != "" {
		parentID, err := uuid.Parse(*req.ParentID)
		if err != nil {
			response.BadRequest(ctx, "Invalid parent_id format")
			return
		}
		company.ParentID = &parentID
	}

	// Parse location_id if provided
	if req.LocationID != nil && *req.LocationID != "" {
		locationID, err := uuid.Parse(*req.LocationID)
		if err != nil {
			response.BadRequest(ctx, "Invalid location_id format")
			return
		}
		company.LocationID = &locationID
	}

	// Parse established_at if provided
	if req.EstablishedAt != "" {
		establishedAt, err := time.Parse("2006-01-02", req.EstablishedAt)
		if err != nil {
			response.BadRequest(ctx, "Invalid established_at format. Use YYYY-MM-DD")
			return
		}
		company.EstablishedAt = &establishedAt
	}

	result, err := c.companySvc.Create(baseService, company)
	if err != nil {
		c.RespondError(ctx, err)
		return
	}

	response.Created(ctx, "Company created successfully", result)
}

func (c *CompanyController) Update(ctx *gin.Context) {
	var req UpdateCompanyRequest
	if !c.ValidateRequest(ctx, &req) {
		return
	}

	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")

	// First get existing company
	existingCompany, err := c.companySvc.FindById(baseService, id)
	if err != nil {
		c.RespondError(ctx, err)
		return
	}

	// Update fields if provided
	if req.CompanyName != nil {
		existingCompany.CompanyName = *req.CompanyName
	}
	if req.LegalName != nil {
		existingCompany.LegalName = *req.LegalName
	}
	if req.TaxID != nil {
		existingCompany.TaxID = *req.TaxID
	}
	if req.CompanyType != nil {
		existingCompany.CompanyType = models.CompanyType(*req.CompanyType)
	}
	if req.ParentID != nil {
		if *req.ParentID == "" {
			existingCompany.ParentID = nil
		} else {
			parentID, err := uuid.Parse(*req.ParentID)
			if err != nil {
				response.BadRequest(ctx, "Invalid parent_id format")
				return
			}
			existingCompany.ParentID = &parentID
		}
	}
	if req.LocationID != nil {
		if *req.LocationID == "" {
			existingCompany.LocationID = nil
		} else {
			locationID, err := uuid.Parse(*req.LocationID)
			if err != nil {
				response.BadRequest(ctx, "Invalid location_id format")
				return
			}
			existingCompany.LocationID = &locationID
		}
	}
	if req.Phone != nil {
		existingCompany.Phone = *req.Phone
	}
	if req.Email != nil {
		existingCompany.Email = *req.Email
	}
	if req.Website != nil {
		existingCompany.Website = *req.Website
	}
	if req.IsActive != nil {
		existingCompany.IsActive = *req.IsActive
	}
	if req.EstablishedAt != nil {
		if *req.EstablishedAt == "" {
			existingCompany.EstablishedAt = nil
		} else {
			establishedAt, err := time.Parse("2006-01-02", *req.EstablishedAt)
			if err != nil {
				response.BadRequest(ctx, "Invalid established_at format. Use YYYY-MM-DD")
				return
			}
			existingCompany.EstablishedAt = &establishedAt
		}
	}

	result, err := c.companySvc.Update(baseService, id, existingCompany)
	if err != nil {
		c.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Company updated successfully", result)
}

func (c *CompanyController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")

	if err := c.companySvc.Delete(baseService, id); err != nil {
		c.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Company deleted successfully", nil)
}
