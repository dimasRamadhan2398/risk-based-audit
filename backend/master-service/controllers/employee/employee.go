package employee

import (
	"master-service/models"
	"master-service/pkg/base"
	"master-service/pkg/response"
	"master-service/pkg/validations"
	employeeSvc "master-service/services/employee"
	"strconv"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Request DTOs
type CreateEmployeeRequest struct {
	EmployeeCode      string  `json:"employee_code" binding:"required"`
	FullName          string  `json:"full_name" binding:"required"`
	Email             string  `json:"email" binding:"required,email"`
	Phone             string  `json:"phone"`
	CompanyID         string  `json:"company_id" binding:"required"`
	DepartmentID      string  `json:"department_id" binding:"required"`
	JobRoleID         string  `json:"job_role_id" binding:"required"`
	LevelGrade        int     `json:"level_grade" binding:"required"`
	WorkLocationID    *string `json:"work_location_id"`
	ResidenceAddress  string  `json:"residence_address"`
	ResidenceCity     string  `json:"residence_city"`
	ResidenceProvince string  `json:"residence_province"`
	ResidencePostal   string  `json:"residence_postal_code"`
	ManagerID         *string `json:"manager_id"`
	IsActive          bool    `json:"is_active"`
	JoinDate          string  `json:"join_date" binding:"required"`
}

type UpdateEmployeeRequest struct {
	FullName          *string `json:"full_name"`
	Email             *string `json:"email"`
	Phone             *string `json:"phone"`
	DepartmentID      *string `json:"department_id"`
	JobRoleID         *string `json:"job_role_id"`
	LevelGrade        *int    `json:"level_grade"`
	WorkLocationID    *string `json:"work_location_id"`
	ResidenceAddress  *string `json:"residence_address"`
	ResidenceCity     *string `json:"residence_city"`
	ResidenceProvince *string `json:"residence_province"`
	ResidencePostal   *string `json:"residence_postal_code"`
	ManagerID         *string `json:"manager_id"`
	IsActive          *bool   `json:"is_active"`
	JoinDate          *string `json:"join_date"`
}

type ListEmployeesRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Search   string `form:"search"`
}

type EmployeeControllerInterface interface {
	FindAll(ctx *gin.Context)
	FindById(ctx *gin.Context)
	Create(ctx *gin.Context)
	Update(ctx *gin.Context)
	Delete(ctx *gin.Context)
	List(ctx *gin.Context)
}

type EmployeeController struct {
	*base.BaseController
	employeeSvc employeeSvc.EmployeeServiceInterface
}

func NewEmployeeController(employeeSvc employeeSvc.EmployeeServiceInterface, validator *validations.Validator) EmployeeControllerInterface {
	return &EmployeeController{
		BaseController: base.NewBaseController(validator),
		employeeSvc:    employeeSvc,
	}
}

func (e *EmployeeController) FindAll(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())

	employees, err := e.employeeSvc.FindAll(baseService)
	if err != nil {
		e.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Employees fetched successfully", employees)
}

func (e *EmployeeController) FindById(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")

	employee, err := e.employeeSvc.FindById(baseService, id)
	if err != nil {
		e.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Employee fetched successfully", employee)
}

func (e *EmployeeController) Create(ctx *gin.Context) {
	var req CreateEmployeeRequest
	if !e.ValidateRequest(ctx, &req) {
		return
	}

	baseService := base.NewBaseService().WithContext(ctx.Request.Context())

	// Convert request to model
	employee := &models.Employee{
		EmployeeCode:      req.EmployeeCode,
		FullName:          req.FullName,
		Email:             req.Email,
		Phone:             req.Phone,
		LevelGrade:        req.LevelGrade,
		ResidenceAddress:  req.ResidenceAddress,
		ResidenceCity:     req.ResidenceCity,
		ResidenceProvince: req.ResidenceProvince,
		ResidencePostal:   req.ResidencePostal,
		IsActive:          req.IsActive,
	}

	joinDate, err := time.Parse("2006-01-02", req.JoinDate)
	if err != nil {
		response.BadRequest(ctx, "Invalid join_date format. Use YYYY-MM-DD")
		return
	}
	employee.JoinDate = joinDate

	// Parse UUIDs
	companyID, err := parseUUID(req.CompanyID)
	if err != nil {
		response.BadRequest(ctx, "Invalid company_id format")
		return
	}
	employee.CompanyID = companyID

	departmentID, err := parseUUID(req.DepartmentID)
	if err != nil {
		response.BadRequest(ctx, "Invalid department_id format")
		return
	}
	employee.DepartmentID = departmentID

	jobRoleID, err := parseUUID(req.JobRoleID)
	if err != nil {
		response.BadRequest(ctx, "Invalid job_role_id format")
		return
	}
	employee.JobRoleID = jobRoleID

	if req.WorkLocationID != nil && *req.WorkLocationID != "" {
		workLocationID, err := parseUUID(*req.WorkLocationID)
		if err != nil {
			response.BadRequest(ctx, "Invalid work_location_id format")
			return
		}
		employee.WorkLocationID = &workLocationID
	}

	if req.ManagerID != nil && *req.ManagerID != "" {
		managerID, err := parseUUID(*req.ManagerID)
		if err != nil {
			response.BadRequest(ctx, "Invalid manager_id format")
			return
		}
		employee.ManagerID = &managerID
	}

	result, err := e.employeeSvc.Create(baseService, employee)
	if err != nil {
		e.RespondError(ctx, err)
		return
	}

	response.Created(ctx, "Employee created successfully", result)
}

func (e *EmployeeController) Update(ctx *gin.Context) {
	var req UpdateEmployeeRequest
	if !e.ValidateRequest(ctx, &req) {
		return
	}

	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")

	// First get existing employee
	existingEmployee, err := e.employeeSvc.FindById(baseService, id)
	if err != nil {
		e.RespondError(ctx, err)
		return
	}

	// Update fields if provided
	if req.FullName != nil {
		existingEmployee.FullName = *req.FullName
	}
	if req.Email != nil {
		existingEmployee.Email = *req.Email
	}
	if req.Phone != nil {
		existingEmployee.Phone = *req.Phone
	}
	if req.DepartmentID != nil {
		if deptID, err := parseUUID(*req.DepartmentID); err == nil {
			existingEmployee.DepartmentID = deptID
		}
	}
	if req.JobRoleID != nil {
		if jobRoleID, err := parseUUID(*req.JobRoleID); err == nil {
			existingEmployee.JobRoleID = jobRoleID
		}
	}
	if req.LevelGrade != nil {
		existingEmployee.LevelGrade = *req.LevelGrade
	}
	if req.WorkLocationID != nil {
		if workLocationID, err := parseUUID(*req.WorkLocationID); err == nil {
			existingEmployee.WorkLocationID = &workLocationID
		}
	}
	if req.ResidenceAddress != nil {
		existingEmployee.ResidenceAddress = *req.ResidenceAddress
	}
	if req.ResidenceCity != nil {
		existingEmployee.ResidenceCity = *req.ResidenceCity
	}
	if req.ResidenceProvince != nil {
		existingEmployee.ResidenceProvince = *req.ResidenceProvince
	}
	if req.ResidencePostal != nil {
		existingEmployee.ResidencePostal = *req.ResidencePostal
	}
	if req.ManagerID != nil {
		if *req.ManagerID == "" {
			existingEmployee.ManagerID = nil
		} else {
			managerID, err := parseUUID(*req.ManagerID)
			if err != nil {
				response.BadRequest(ctx, "Invalid manager_id format")
				return
			}
			existingEmployee.ManagerID = &managerID
		}
	}
	if req.IsActive != nil {
		existingEmployee.IsActive = *req.IsActive
	}
	if req.JoinDate != nil {
	joinDate, err := time.Parse("2006-01-02", *req.JoinDate)
	if err != nil {
		response.BadRequest(ctx, "Invalid join_date format. Use YYYY-MM-DD")
		return
	}
	existingEmployee.JoinDate = joinDate
}

	result, err := e.employeeSvc.Update(baseService, id, existingEmployee)
	if err != nil {
		e.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Employee updated successfully", result)
}

func (e *EmployeeController) Delete(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())
	id := ctx.Param("id")

	if err := e.employeeSvc.Delete(baseService, id); err != nil {
		e.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Employee deleted successfully", nil)
}

func (e *EmployeeController) List(ctx *gin.Context) {
	baseService := base.NewBaseService().WithContext(ctx.Request.Context())

	// Parse query params
	page := 1
	pageSize := 10
	if p := ctx.Query("page"); p != "" {
		if parsed, err := parseInt(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := ctx.Query("page_size"); ps != "" {
		if parsed, err := parseInt(ps); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}
	search := ctx.Query("search")

	offset := (page - 1) * pageSize

	employees, total, err := e.employeeSvc.FindMany(baseService, offset, pageSize, search)
	if err != nil {
		e.RespondError(ctx, err)
		return
	}

	response.OK(ctx, "Employees fetched successfully", gin.H{
		"employees": employees,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// Helper functions

func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}
