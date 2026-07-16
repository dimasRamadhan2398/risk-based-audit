package department

import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/department"

	"github.com/google/uuid"
)

type DepartmentServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.Department, error)
	FindById(ctx *base.BaseService, id string) (*models.Department, error)
	FindMany(ctx *base.BaseService, offset, limit int, search string) (*[]models.Department, int64, error)
	Create(ctx *base.BaseService, department *models.Department) (*models.Department, error)
	Update(ctx *base.BaseService, id string, department *models.Department) (*models.Department, error)
	Delete(ctx *base.BaseService, id string) error
}

type DepartmentService struct {
	departmentRepo repo.IDepartmentRepository
}

// Create implements DepartmentServiceInterface.
func (d *DepartmentService) Create(ctx *base.BaseService, department *models.Department) (*models.Department, error) {
	if err := d.validateReferences(department); err != nil {
		return nil, err
	}

	if _, err := d.departmentRepo.FindByCode(department.DepartmentCode); err == nil {
		return nil, apperrors.Wrap("DEPARTMENT_CODE_ALREADY_EXISTS", "Department code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate department code", 500, err)
	}

	if _, err := d.departmentRepo.FindByName(department.DepartmentName); err == nil {
		return nil, apperrors.Wrap("DEPARTMENT_NAME_ALREADY_EXISTS", "Department name already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate department name", 500, err)
	}

	if err := d.departmentRepo.Create(department); err != nil {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create department", 500, err)
	}

	return department, nil
}

// Delete implements DepartmentServiceInterface.
func (d *DepartmentService) Delete(ctx *base.BaseService, id string) error {
	departmentID, err := uuid.Parse(id)
	if err != nil {
		return apperrors.Wrap("INVALID_DEPARTMENT_ID", "Invalid department ID format", 400, err)
	}

	if _, err := d.departmentRepo.FindByID(departmentID); err != nil {
		if err == apperrors.ErrNotFound {
			return err
		}
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find department", 500, err)
	}

	if err := d.departmentRepo.Delete(departmentID); err != nil {
		return apperrors.Wrap("DATABASE_ERROR", "Failed to delete department", 500, err)
	}

	return nil
}

// FindAll implements DepartmentServiceInterface.
func (d *DepartmentService) FindAll(ctx *base.BaseService) (*[]models.Department, error) {
	departments, err := d.departmentRepo.FindAll()
	if err != nil {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch departments", 500, err)
	}

	result := make([]models.Department, 0, len(departments))
	for _, department := range departments {
		if department != nil {
			result = append(result, *department)
		}
	}

	return &result, nil
}

// FindById implements DepartmentServiceInterface.
func (d *DepartmentService) FindById(ctx *base.BaseService, id string) (*models.Department, error) {
	departmentID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.Wrap("INVALID_DEPARTMENT_ID", "Invalid department ID format", 400, err)
	}

	department, err := d.departmentRepo.FindByID(departmentID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			return nil, err
		}
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch department", 500, err)
	}

	return department, nil
}

// FindMany finds departments with pagination
func (d *DepartmentService) FindMany(ctx *base.BaseService, offset, limit int, search string) (*[]models.Department, int64, error) {
	departments, err := d.departmentRepo.FindMany(offset, limit, search)
	if err != nil {
		return nil, 0, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch departments", 500, err)
	}

	count, err := d.departmentRepo.Count(search)
	if err != nil {
		return nil, 0, apperrors.Wrap("DATABASE_ERROR", "Failed to count departments", 500, err)
	}

	result := make([]models.Department, 0, len(departments))
	for _, department := range departments {
		if department != nil {
			result = append(result, *department)
		}
	}

	return &result, count, nil
}

// Update implements DepartmentServiceInterface.
func (d *DepartmentService) Update(ctx *base.BaseService, id string, department *models.Department) (*models.Department, error) {
	departmentID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.Wrap("INVALID_DEPARTMENT_ID", "Invalid department ID format", 400, err)
	}

	existingDepartment, err := d.departmentRepo.FindByID(departmentID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			return nil, err
		}
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch department", 500, err)
	}

	if err := d.validateReferences(department); err != nil {
		return nil, err
	}

	if existingDepartment.DepartmentCode != department.DepartmentCode {
		if _, err := d.departmentRepo.FindByCode(department.DepartmentCode); err == nil {
			return nil, apperrors.Wrap("DEPARTMENT_CODE_ALREADY_EXISTS", "Department code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate department code", 500, err)
		}
	}

	if existingDepartment.DepartmentName != department.DepartmentName {
		if _, err := d.departmentRepo.FindByName(department.DepartmentName); err == nil {
			return nil, apperrors.Wrap("DEPARTMENT_NAME_ALREADY_EXISTS", "Department name already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate department name", 500, err)
		}
	}

	existingDepartment.DepartmentCode = department.DepartmentCode
	existingDepartment.DepartmentName = department.DepartmentName
	existingDepartment.DepartmentDescription = department.DepartmentDescription
	existingDepartment.PicID = department.PicID
	existingDepartment.Level = department.Level
	existingDepartment.IsActive = department.IsActive
	existingDepartment.CompanyID = department.CompanyID
	existingDepartment.BusinessUnitID = department.BusinessUnitID

	if err := d.departmentRepo.Update(existingDepartment); err != nil {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update department", 500, err)
	}

	return existingDepartment, nil
}

func (d *DepartmentService) validateReferences(department *models.Department) error {
	if department.CompanyID == uuid.Nil {
		return apperrors.Wrap("COMPANY_ID_REQUIRED", "company_id is required", 400, nil)
	}
	if department.BusinessUnitID == uuid.Nil {
		return apperrors.Wrap("BUSINESS_UNIT_ID_REQUIRED", "business_unit_id is required", 400, nil)
	}
	if department.PicID == uuid.Nil {
		return apperrors.Wrap("PIC_ID_REQUIRED", "pic_id is required", 400, nil)
	}

	companyExists, err := d.departmentRepo.CompanyExists(department.CompanyID)
	if err != nil {
		return apperrors.Wrap("DATABASE_ERROR", "Failed to validate company", 500, err)
	}
	if !companyExists {
		return apperrors.Wrap("COMPANY_NOT_FOUND", "Company not found", 404, nil)
	}

	businessUnitExists, err := d.departmentRepo.BusinessUnitExists(department.BusinessUnitID)
	if err != nil {
		return apperrors.Wrap("DATABASE_ERROR", "Failed to validate business unit", 500, err)
	}
	if !businessUnitExists {
		return apperrors.Wrap("BUSINESS_UNIT_NOT_FOUND", "Business unit not found", 404, nil)
	}

	businessUnitBelongs, err := d.departmentRepo.BusinessUnitBelongsToCompany(department.BusinessUnitID, department.CompanyID)
	if err != nil {
		return apperrors.Wrap("DATABASE_ERROR", "Failed to validate business unit company", 500, err)
	}
	if !businessUnitBelongs {
		return apperrors.Wrap("BUSINESS_UNIT_COMPANY_MISMATCH", "Business unit does not belong to the selected company", 409, nil)
	}

	picExists, err := d.departmentRepo.EmployeeExists(department.PicID)
	if err != nil {
		return apperrors.Wrap("DATABASE_ERROR", "Failed to validate PIC", 500, err)
	}
	if !picExists {
		return apperrors.Wrap("PIC_NOT_FOUND", "PIC employee not found", 404, nil)
	}

	picBelongs, err := d.departmentRepo.EmployeeBelongsToCompany(department.PicID, department.CompanyID)
	if err != nil {
		return apperrors.Wrap("DATABASE_ERROR", "Failed to validate PIC company", 500, err)
	}
	if !picBelongs {
		return apperrors.Wrap("PIC_COMPANY_MISMATCH", "PIC employee does not belong to the selected company", 409, nil)
	}

	return nil
}

func NewDepartmentService(departmentRepo repo.IDepartmentRepository) DepartmentServiceInterface {
	return &DepartmentService{
		departmentRepo: departmentRepo,
	}
}
