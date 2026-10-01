package department

import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/department"
	"net/http"

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

var (
	errDepartmentCodeExists = apperrors.New(apperrors.CodeDepartmentCodeAlreadyExists,
		"A department with this code already exists.", http.StatusConflict).
		WithFields(map[string]string{"department_code": apperrors.FieldAlreadyExists})
	errDepartmentNameExists = apperrors.New(apperrors.CodeDepartmentNameAlreadyExists,
		"A department with this name already exists.", http.StatusConflict).
		WithFields(map[string]string{"department_name": apperrors.FieldAlreadyExists})
)

func fieldErr(code, message string, status int, field, fieldCode string) *apperrors.AppError {
	return apperrors.New(code, message, status).WithFields(map[string]string{field: fieldCode})
}

// Create implements DepartmentServiceInterface.
func (d *DepartmentService) Create(ctx *base.BaseService, department *models.Department) (*models.Department, error) {
	if err := d.validateReferences(department); err != nil {
		return nil, err
	}

	if _, err := d.departmentRepo.FindByCode(department.DepartmentCode); err == nil {
		return nil, errDepartmentCodeExists
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to validate the department code.")
	}

	if _, err := d.departmentRepo.FindByName(department.DepartmentName); err == nil {
		return nil, errDepartmentNameExists
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to validate the department name.")
	}

	if err := d.departmentRepo.Create(department); err != nil {
		return nil, apperrors.DB(err, apperrors.OpCreate, "Failed to create the department.")
	}

	return department, nil
}

// Delete implements DepartmentServiceInterface.
func (d *DepartmentService) Delete(ctx *base.BaseService, id string) error {
	departmentID, err := uuid.Parse(id)
	if err != nil {
		return apperrors.InvalidID(err)
	}

	if _, err := d.departmentRepo.FindByID(departmentID); err != nil {
		if err == apperrors.ErrNotFound {
			return apperrors.ErrDepartmentNotFound
		}
		return apperrors.DB(err, apperrors.OpRead, "Failed to load the department.")
	}

	if err := d.departmentRepo.Delete(departmentID); err != nil {
		return apperrors.DB(err, apperrors.OpDelete, "Failed to delete the department.")
	}

	return nil
}

// FindAll implements DepartmentServiceInterface.
func (d *DepartmentService) FindAll(ctx *base.BaseService) (*[]models.Department, error) {
	departments, err := d.departmentRepo.FindAll()
	if err != nil {
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to load departments.")
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
		return nil, apperrors.InvalidID(err)
	}

	department, err := d.departmentRepo.FindByID(departmentID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			return nil, apperrors.ErrDepartmentNotFound
		}
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to load the department.")
	}

	return department, nil
}

// FindMany finds departments with pagination
func (d *DepartmentService) FindMany(ctx *base.BaseService, offset, limit int, search string) (*[]models.Department, int64, error) {
	departments, err := d.departmentRepo.FindMany(offset, limit, search)
	if err != nil {
		return nil, 0, apperrors.DB(err, apperrors.OpRead, "Failed to load departments.")
	}

	count, err := d.departmentRepo.Count(search)
	if err != nil {
		return nil, 0, apperrors.DB(err, apperrors.OpRead, "Failed to load departments.")
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
		return nil, apperrors.InvalidID(err)
	}

	existingDepartment, err := d.departmentRepo.FindByID(departmentID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			return nil, apperrors.ErrDepartmentNotFound
		}
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to load the department.")
	}

	if err := d.validateReferences(department); err != nil {
		return nil, err
	}

	if existingDepartment.DepartmentCode != department.DepartmentCode {
		if _, err := d.departmentRepo.FindByCode(department.DepartmentCode); err == nil {
			return nil, errDepartmentCodeExists
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.DB(err, apperrors.OpRead, "Failed to validate the department code.")
		}
	}

	if existingDepartment.DepartmentName != department.DepartmentName {
		if _, err := d.departmentRepo.FindByName(department.DepartmentName); err == nil {
			return nil, errDepartmentNameExists
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.DB(err, apperrors.OpRead, "Failed to validate the department name.")
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
		return nil, apperrors.DB(err, apperrors.OpUpdate, "Failed to update the department.")
	}

	return existingDepartment, nil
}

func (d *DepartmentService) validateReferences(department *models.Department) error {
	required := map[string]string{}
	if department.CompanyID == uuid.Nil {
		required["company_id"] = apperrors.FieldRequired
	}
	if department.BusinessUnitID == uuid.Nil {
		required["business_unit_id"] = apperrors.FieldRequired
	}
	if department.PicID == uuid.Nil {
		required["pic_id"] = apperrors.FieldRequired
	}
	if len(required) > 0 {
		return apperrors.ValidationFailed("", required)
	}

	companyExists, err := d.departmentRepo.CompanyExists(department.CompanyID)
	if err != nil {
		return apperrors.DB(err, apperrors.OpRead, "Failed to validate the company.")
	}
	if !companyExists {
		return fieldErr(apperrors.CodeCompanyNotFound, "The selected company does not exist.", http.StatusNotFound, "company_id", apperrors.FieldNotFound)
	}

	businessUnitExists, err := d.departmentRepo.BusinessUnitExists(department.BusinessUnitID)
	if err != nil {
		return apperrors.DB(err, apperrors.OpRead, "Failed to validate the business unit.")
	}
	if !businessUnitExists {
		return fieldErr(apperrors.CodeBusinessUnitNotFound, "The selected business unit does not exist.", http.StatusNotFound, "business_unit_id", apperrors.FieldNotFound)
	}

	businessUnitBelongs, err := d.departmentRepo.BusinessUnitBelongsToCompany(department.BusinessUnitID, department.CompanyID)
	if err != nil {
		return apperrors.DB(err, apperrors.OpRead, "Failed to validate the business unit.")
	}
	if !businessUnitBelongs {
		return fieldErr(apperrors.CodeBusinessUnitCompanyMismatch, "The selected business unit does not belong to the selected company.", http.StatusConflict, "business_unit_id", apperrors.FieldMismatch)
	}

	picExists, err := d.departmentRepo.EmployeeExists(department.PicID)
	if err != nil {
		return apperrors.DB(err, apperrors.OpRead, "Failed to validate the person in charge.")
	}
	if !picExists {
		return fieldErr(apperrors.CodePICNotFound, "The selected person in charge does not exist.", http.StatusNotFound, "pic_id", apperrors.FieldNotFound)
	}

	picBelongs, err := d.departmentRepo.EmployeeBelongsToCompany(department.PicID, department.CompanyID)
	if err != nil {
		return apperrors.DB(err, apperrors.OpRead, "Failed to validate the person in charge.")
	}
	if !picBelongs {
		return fieldErr(apperrors.CodePICCompanyMismatch, "The selected person in charge does not belong to the selected company.", http.StatusConflict, "pic_id", apperrors.FieldMismatch)
	}

	return nil
}

func NewDepartmentService(departmentRepo repo.IDepartmentRepository) DepartmentServiceInterface {
	return &DepartmentService{
		departmentRepo: departmentRepo,
	}
}
