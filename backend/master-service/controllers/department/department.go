package department

import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	departmentSvc "master-service/services/department"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Request DTOs
type CreateDepartmentRequest struct {
	DepartmentCode        string `json:"department_code" binding:"required"`
	DepartmentName        string `json:"department_name" binding:"required"`
	DepartmentDescription string `json:"department_description"`
	PicID                 string `json:"pic_id" binding:"required"`
	Level                 int    `json:"level" binding:"required"`
	CompanyID             string `json:"company_id" binding:"required"`
	BusinessUnitID        string `json:"business_unit_id" binding:"required"`
	IsActive              bool   `json:"is_active"`
}

type UpdateDepartmentRequest struct {
	DepartmentName        *string `json:"department_name"`
	DepartmentDescription *string `json:"department_description"`
	PicID                 *string `json:"pic_id"`
	Level                 *int    `json:"level"`
	CompanyID             *string `json:"company_id"`
	BusinessUnitID        *string `json:"business_unit_id"`
	IsActive              *bool   `json:"is_active"`
}

type DepartmentControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	List(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
}

type DepartmentController struct {
	*base.BaseController
	departmentSvc departmentSvc.DepartmentServiceInterface
}

func NewDepartmentController(departmentSvc departmentSvc.DepartmentServiceInterface, validator *validations.Validator) DepartmentControllerInterface {
	return &DepartmentController{
		BaseController: base.NewBaseController(validator),
		departmentSvc:  departmentSvc,
	}
}

func (d *DepartmentController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())

	departments, err := d.departmentSvc.FindAll(baseService)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Departments fetched successfully", departments)
}

func (d *DepartmentController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")

	department, err := d.departmentSvc.FindById(baseService, id)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Department fetched successfully", department)
}

func (d *DepartmentController) List(ctx *gin.Context) {
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

	departments, total, err := d.departmentSvc.FindMany(baseService, offset, pageSize, search)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Departments fetched successfully", gin.H{
		"departments": departments,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

func (d *DepartmentController) Create(ctx *gin.Context) {
	var req CreateDepartmentRequest
	if !d.ValidateRequest(ctx, &req) {
		return
	}

	baseService := base.NewBaseService().WithContext(ctx.Request.Context())

	// Convert request to model
	department := &models.Department{
		DepartmentCode:        req.DepartmentCode,
		DepartmentName:        req.DepartmentName,
		DepartmentDescription: req.DepartmentDescription,
		Level:                 req.Level,
		IsActive:              req.IsActive,
	}

	// Parse UUIDs
	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		response.BadRequest(ctx, "Invalid company_id format")
		return
	}
	department.CompanyID = companyID

	picID, err := uuid.Parse(req.PicID)
	if err != nil {
		response.BadRequest(ctx, "Invalid pic_id format")
		return
	}
	department.PicID = picID

	businessUnitID, err := uuid.Parse(req.BusinessUnitID)
	if err != nil {
		response.BadRequest(ctx, "Invalid business_unit_id format")
		return
	}
	department.BusinessUnitID = businessUnitID

	result, err := d.departmentSvc.Create(baseService, department)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}

	response.Created(ctx, "Department created successfully", result)
}

func (d *DepartmentController) Update(ctx *gin.Context) {
	var req UpdateDepartmentRequest
	if !d.ValidateRequest(ctx, &req) {
		return
	}

	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")

	// First get existing department
	existingDepartment, err := d.departmentSvc.FindById(baseService, id)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}

	// Update fields if provided
	if req.PicID != nil {
		picID, err := uuid.Parse(*req.PicID)
		if err != nil {
			response.BadRequest(ctx, "Invalid pic_id format")
			return
		}
		existingDepartment.PicID = picID
	}

	if req.CompanyID != nil {
		companyID, err := uuid.Parse(*req.CompanyID)
		if err != nil {
			response.BadRequest(ctx, "Invalid company_id format")
			return
		}
		existingDepartment.CompanyID = companyID
	}

	if req.BusinessUnitID != nil {
		businessUnitID, err := uuid.Parse(*req.BusinessUnitID)
		if err != nil {
			response.BadRequest(ctx, "Invalid business_unit_id format")
			return
		}
		existingDepartment.BusinessUnitID = businessUnitID
	}

	result, err := d.departmentSvc.Update(baseService, id, existingDepartment)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Department updated successfully", result)
}

func (d *DepartmentController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")

	if err := d.departmentSvc.Delete(baseService, id); err != nil {
		d.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Department deleted successfully", nil)
}
