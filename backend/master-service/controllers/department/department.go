package department

import (
	"encoding/json"
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	departmentRepo "master-service/repositories/department"
	departmentSvc "master-service/services/department"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Request DTOs
type CreateDepartmentRequest struct {
	DepartmentCode        string  `json:"department_code" binding:"required"`
	DepartmentName        string  `json:"department_name" binding:"required"`
	DepartmentDescription string  `json:"department_description"`
	PicID                 *string `json:"pic_id"`
	Level                 int     `json:"level" binding:"required"`
	CompanyID             string  `json:"company_id" binding:"required"`
	BusinessUnitID        *string `json:"business_unit_id"`
	IsActive              bool    `json:"is_active"`
	// Flexible tenants only; a legacy tenant answers 422 ORG_STRUCTURE_NOT_ENABLED.
	ParentDepartmentID *string `json:"parent_department_id"`
}

// OptionalID is a JSON field that distinguishes "absent" (Set=false) from
// null or "" (Set=true, Value=nil): on update, null moves a unit to the root.
type OptionalID struct {
	Set   bool
	Value *string
}

func (o *OptionalID) UnmarshalJSON(b []byte) error {
	o.Set = true
	if string(b) == "null" {
		o.Value = nil
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	o.Value = &s
	return nil
}

// present reports a non-empty value.
func (o OptionalID) present() bool {
	return o.Set && o.Value != nil && strings.TrimSpace(*o.Value) != ""
}

type UpdateDepartmentRequest struct {
	DepartmentName        *string `json:"department_name"`
	DepartmentDescription *string `json:"department_description"`
	PicID                 *string `json:"pic_id"`
	Level                 *int    `json:"level"`
	CompanyID             *string `json:"company_id"`
	BusinessUnitID        *string `json:"business_unit_id"`
	IsActive              *bool   `json:"is_active"`
	// Flexible tenants only. Absent = keep; null or "" = root; id = move.
	ParentDepartmentID OptionalID `json:"parent_department_id"`
}

type DepartmentControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	List(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
	// Tree serves GET /departments/tree (flexible tenants; the route is
	// guarded by middleware.RequireFlexible).
	Tree(ctx *gin.Context)
}

type DepartmentController struct {
	*base.BaseController
	departmentSvc departmentSvc.DepartmentServiceInterface
	// org is set on flexible tenants (the registry then hands out the
	// flexible service); nil means legacy behaviour, unchanged.
	org departmentSvc.OrgDepartmentServiceInterface
}

func NewDepartmentController(svc departmentSvc.DepartmentServiceInterface, validator *validations.Validator) DepartmentControllerInterface {
	org, _ := svc.(departmentSvc.OrgDepartmentServiceInterface)
	return &DepartmentController{
		BaseController: base.NewBaseController(validator),
		departmentSvc:  svc,
		org:            org,
	}
}

// errParentNotEnabled: a parent sent to a legacy tenant is refused, never ignored.
var errParentNotEnabled = apperrors.New(apperrors.CodeOrgStructureNotEnabled,
	"Parent units are not enabled for this tenant.", http.StatusUnprocessableEntity).
	WithFields(map[string]string{"parent_department_id": apperrors.FieldInvalid})

func parseOptionalUUID(raw *string, field string) (*uuid.UUID, error) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil, nil
	}
	id, err := uuid.Parse(strings.TrimSpace(*raw))
	if err != nil {
		return nil, apperrors.ValidationFailed("", map[string]string{field: apperrors.FieldInvalidFormat})
	}
	return &id, nil
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

	if d.org != nil {
		d.listFlexible(ctx, baseService, page, pageSize, search)
		return
	}

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
		d.RespondError(ctx, apperrors.ValidationFailed("", map[string]string{"company_id": apperrors.FieldInvalidFormat}))
		return
	}
	department.CompanyID = companyID

	if req.PicID != nil && *req.PicID != "" {
		picID, err := uuid.Parse(*req.PicID)
		if err != nil {
			response.BadRequest(ctx, "Invalid pic_id format")
			return
		}
		department.PicID = &picID
	}

	if req.BusinessUnitID != nil && *req.BusinessUnitID != "" {
		businessUnitID, err := uuid.Parse(*req.BusinessUnitID)
		if err != nil {
			response.BadRequest(ctx, "Invalid business_unit_id format")
			return
		}
		department.BusinessUnitID = &businessUnitID
	}

	if req.ParentDepartmentID != nil && strings.TrimSpace(*req.ParentDepartmentID) != "" {
		if d.org == nil {
			d.RespondError(ctx, errParentNotEnabled)
			return
		}
		parentID, err := parseOptionalUUID(req.ParentDepartmentID, "parent_department_id")
		if err != nil {
			d.RespondError(ctx, err)
			return
		}
		department.ParentDepartmentID = parentID
	}

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

	if d.org == nil && req.ParentDepartmentID.present() {
		d.RespondError(ctx, errParentNotEnabled)
		return
	}

	// First get existing department
	existingDepartment, err := d.departmentSvc.FindById(baseService, id)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}

	// Update fields if provided
	if req.PicID != nil {
		if *req.PicID == "" {
			existingDepartment.PicID = nil
		} else {
			picID, err := uuid.Parse(*req.PicID)
			if err != nil {
				response.BadRequest(ctx, "Invalid pic_id format")
				return
			}
			existingDepartment.PicID = &picID
		}
	}

	if req.CompanyID != nil {
		companyID, err := uuid.Parse(*req.CompanyID)
		if err != nil {
			d.RespondError(ctx, apperrors.ValidationFailed("", map[string]string{"company_id": apperrors.FieldInvalidFormat}))
			return
		}
		existingDepartment.CompanyID = companyID
	}

	if req.BusinessUnitID != nil {
		if *req.BusinessUnitID == "" {
			existingDepartment.BusinessUnitID = nil
		} else {
			businessUnitID, err := uuid.Parse(*req.BusinessUnitID)
			if err != nil {
				response.BadRequest(ctx, "Invalid business_unit_id format")
				return
			}
			existingDepartment.BusinessUnitID = &businessUnitID
		}
	}

	if d.org != nil {
		// Flexible tenants apply every provided field. (The legacy path above
		// keeps its historical behaviour: it only applies pic, company and
		// business unit.)
		if req.DepartmentName != nil {
			existingDepartment.DepartmentName = *req.DepartmentName
		}
		if req.DepartmentDescription != nil {
			existingDepartment.DepartmentDescription = *req.DepartmentDescription
		}
		if req.Level != nil {
			existingDepartment.Level = *req.Level
		}
		if req.IsActive != nil {
			existingDepartment.IsActive = *req.IsActive
		}
		if req.ParentDepartmentID.Set {
			parentID, err := parseOptionalUUID(req.ParentDepartmentID.Value, "parent_department_id")
			if err != nil {
				d.RespondError(ctx, err)
				return
			}
			existingDepartment.ParentDepartmentID = parentID
		}
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

// listFlexible serves GET /departments on flexible tenants: filters company_id,
// level, parent_id (an id, or "root"), is_active, and order (whitelisted
// columns, see departmentRepo.ParseOrder). The response shape is the legacy one.
func (d *DepartmentController) listFlexible(ctx *gin.Context, baseService *base.BaseService, page, pageSize int, search string) {
	filter := departmentRepo.ListFilter{Offset: (page - 1) * pageSize, Limit: pageSize, Search: search}
	fields := map[string]string{}
	if raw := strings.TrimSpace(ctx.Query("company_id")); raw != "" {
		if id, err := uuid.Parse(raw); err == nil {
			filter.CompanyID = &id
		} else {
			fields["company_id"] = apperrors.FieldInvalidFormat
		}
	}
	if raw := strings.TrimSpace(ctx.Query("level")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			filter.Level = &n
		} else {
			fields["level"] = apperrors.FieldInvalidFormat
		}
	}
	if raw := strings.TrimSpace(ctx.Query("parent_id")); raw != "" {
		if strings.EqualFold(raw, "root") {
			filter.RootsOnly = true
		} else if id, err := uuid.Parse(raw); err == nil {
			filter.ParentID = &id
		} else {
			fields["parent_id"] = apperrors.FieldInvalidFormat
		}
	}
	if raw := strings.TrimSpace(ctx.Query("is_active")); raw != "" {
		if b, err := strconv.ParseBool(raw); err == nil {
			filter.IsActive = &b
		} else {
			fields["is_active"] = apperrors.FieldInvalidFormat
		}
	}
	order, ok := departmentRepo.ParseOrder(ctx.Query("order"))
	if !ok {
		fields["order"] = apperrors.FieldInvalidOption
	}
	filter.Order = order
	if len(fields) > 0 {
		d.RespondError(ctx, apperrors.ValidationFailed("", fields))
		return
	}

	departments, total, err := d.org.List(baseService, filter)
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

// Tree serves GET /departments/tree?company_id= with nested children.
func (d *DepartmentController) Tree(ctx *gin.Context) {
	if d.org == nil {
		d.RespondError(ctx, apperrors.ErrOrgStructureNotEnabled)
		return
	}
	var companyID *uuid.UUID
	if raw := strings.TrimSpace(ctx.Query("company_id")); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			d.RespondError(ctx, apperrors.ValidationFailed("", map[string]string{"company_id": apperrors.FieldInvalidFormat}))
			return
		}
		companyID = &id
	}
	tree, err := d.org.Tree(base.NewBaseService().WithContext(ctx.Request.Context()), companyID)
	if err != nil {
		d.RespondError(ctx, err)
		return
	}
	response.OK(ctx, "Department tree fetched successfully", gin.H{"tree": tree})
}
