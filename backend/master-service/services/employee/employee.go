package employee

import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/employee"

	"github.com/google/uuid"
)

type EmployeeServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.Employee, error)
	FindById(ctx *base.BaseService, id string) (*models.Employee, error)
	Create(ctx *base.BaseService, employee *models.Employee) (*models.Employee, error)
	Update(ctx *base.BaseService, id string, employee *models.Employee) (*models.Employee, error)
	Delete(ctx *base.BaseService, id string) error
	FindMany(ctx *base.BaseService, offset, limit int, search string) (*[]models.Employee, int64, error)
}

type EmployeeService struct {
	employeeRepo repo.IEmployeeRepository
}

// FindMany finds employees with pagination
func (s *EmployeeService) FindMany(ctx *base.BaseService, offset, limit int, search string) (*[]models.Employee, int64, error) {
	employees, err := s.employeeRepo.FindMany(offset, limit, search)
	if err != nil {
		return nil, 0, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch employees", 500, err)
	}

	count, err := s.employeeRepo.Count(search)
	if err != nil {
		return nil, 0, apperrors.Wrap("DATABASE_ERROR", "Failed to count employees", 500, err)
	}

	result := make([]models.Employee, 0, len(employees))
	for _, employee := range employees {
		if employee != nil {
			result = append(result, *employee)
		}
	}

	return &result, count, nil
}

// FindAll implements EmployeeServiceInterface.
func (s *EmployeeService) FindAll(ctx *base.BaseService) (*[]models.Employee, error) {
	employees, err := s.employeeRepo.FindAll()
	if err != nil {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch employees", 500, err)
	}

	result := make([]models.Employee, 0, len(employees))
	for _, employee := range employees {
		if employee != nil {
			result = append(result, *employee)
		}
	}

	return &result, nil
}

// FindById implements EmployeeServiceInterface.
func (s *EmployeeService) FindById(ctx *base.BaseService, id string) (*models.Employee, error) {
	employeeID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.Wrap("INVALID_EMPLOYEE_ID", "Invalid employee ID format", 400, err)
	}

	employee, err := s.employeeRepo.FindByID(employeeID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			return nil, err
		}
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch employee", 500, err)
	}

	return employee, nil
}

// Create implements EmployeeServiceInterface.
func (s *EmployeeService) Create(ctx *base.BaseService, employee *models.Employee) (*models.Employee, error) {
	if err := s.validateReferences(uuid.Nil, employee); err != nil {
		return nil, err
	}
	// Check if employee code already exists
	if _, err := s.employeeRepo.FindByCode(employee.EmployeeCode); err == nil {
		return nil, apperrors.Wrap("EMPLOYEE_CODE_ALREADY_EXISTS", "Employee code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate employee code", 500, err)
	}

	// Check if email already exists
	if _, err := s.employeeRepo.FindByEmail(employee.Email); err == nil {
		return nil, apperrors.Wrap("EMPLOYEE_EMAIL_ALREADY_EXISTS", "Employee email already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate employee email", 500, err)
	}

	if err := s.employeeRepo.Create(employee); err != nil {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to create employee", 500, err)
	}

	return employee, nil
}

// Update implements EmployeeServiceInterface.
func (s *EmployeeService) Update(ctx *base.BaseService, id string, employee *models.Employee) (*models.Employee, error) {
	employeeID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.Wrap("INVALID_EMPLOYEE_ID", "Invalid employee ID format", 400, err)
	}

	if err := s.validateReferences(employeeID, employee); err != nil {
		return nil, err
	}

	existingEmployee, err := s.employeeRepo.FindByID(employeeID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			return nil, err
		}
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch employee", 500, err)
	}

	// Check if new employee code conflicts with another employee
	if existingEmployee.EmployeeCode != employee.EmployeeCode {
		if _, err := s.employeeRepo.FindByCode(employee.EmployeeCode); err == nil {
			return nil, apperrors.Wrap("EMPLOYEE_CODE_ALREADY_EXISTS", "Employee code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate employee code", 500, err)
		}
	}

	// Check if new email conflicts with another employee
	if existingEmployee.Email != employee.Email {
		if _, err := s.employeeRepo.FindByEmail(employee.Email); err == nil {
			return nil, apperrors.Wrap("EMPLOYEE_EMAIL_ALREADY_EXISTS", "Employee email already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate employee email", 500, err)
		}
	}

	// Update fields
	existingEmployee.FullName = employee.FullName
	existingEmployee.Email = employee.Email
	existingEmployee.Phone = employee.Phone
	existingEmployee.DepartmentID = employee.DepartmentID
	existingEmployee.JobRoleID = employee.JobRoleID
	existingEmployee.LevelGrade = employee.LevelGrade
	existingEmployee.WorkLocationID = employee.WorkLocationID
	existingEmployee.ResidenceAddress = employee.ResidenceAddress
	existingEmployee.ResidenceCity = employee.ResidenceCity
	existingEmployee.ResidenceProvince = employee.ResidenceProvince
	existingEmployee.ResidencePostal = employee.ResidencePostal
	existingEmployee.ManagerID = employee.ManagerID
	existingEmployee.IsActive = employee.IsActive

	if err := s.employeeRepo.Update(existingEmployee); err != nil {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update employee", 500, err)
	}

	return existingEmployee, nil
}

// Delete implements EmployeeServiceInterface.
func (s *EmployeeService) Delete(ctx *base.BaseService, id string) error {
	employeeID, err := uuid.Parse(id)
	if err != nil {
		return apperrors.Wrap("INVALID_EMPLOYEE_ID", "Invalid employee ID format", 400, err)
	}

	if _, err := s.employeeRepo.FindByID(employeeID); err != nil {
		if err == apperrors.ErrNotFound {
			return err
		}
		return apperrors.Wrap("DATABASE_ERROR", "Failed to find employee", 500, err)
	}

	if err := s.employeeRepo.Delete(employeeID); err != nil {
		return apperrors.Wrap("DATABASE_ERROR", "Failed to delete employee", 500, err)
	}

	return nil
}

func (s *EmployeeService) validateReferences(currentEmployeeID uuid.UUID, employee *models.Employee) error {
	if employee.CompanyID == uuid.Nil {
		return apperrors.Wrap("COMPANY_ID_REQUIRED", "company_id is required", 400, nil)
	}
	if employee.DepartmentID == uuid.Nil {
		return apperrors.Wrap("DEPARTMENT_ID_REQUIRED", "department_id is required", 400, nil)
	}
	if employee.JobRoleID == uuid.Nil {
		return apperrors.Wrap("JOB_ROLE_ID_REQUIRED", "job_role_id is required", 400, nil)
	}

	companyExists, err := s.employeeRepo.CompanyExists(employee.CompanyID)
	if err != nil {
		return apperrors.Wrap("DATABASE_ERROR", "Failed to validate company", 500, err)
	}
	if !companyExists {
		return apperrors.Wrap("COMPANY_NOT_FOUND", "Company not found", 404, nil)
	}

	departmentExists, err := s.employeeRepo.DepartmentExists(employee.DepartmentID)
	if err != nil {
		return apperrors.Wrap("DATABASE_ERROR", "Failed to validate department", 500, err)
	}
	if !departmentExists {
		return apperrors.Wrap("DEPARTMENT_NOT_FOUND", "Department not found", 404, nil)
	}

	departmentBelongs, err := s.employeeRepo.DepartmentBelongsToCompany(employee.DepartmentID, employee.CompanyID)
	if err != nil {
		return apperrors.Wrap("DATABASE_ERROR", "Failed to validate department company", 500, err)
	}
	if !departmentBelongs {
		return apperrors.Wrap("DEPARTMENT_COMPANY_MISMATCH", "Department does not belong to the selected company", 409, nil)
	}

	jobRoleExists, err := s.employeeRepo.JobRoleExists(employee.JobRoleID)
	if err != nil {
		return apperrors.Wrap("DATABASE_ERROR", "Failed to validate job role", 500, err)
	}
	if !jobRoleExists {
		return apperrors.Wrap("JOB_ROLE_NOT_FOUND", "Job role not found", 404, nil)
	}

	if employee.WorkLocationID != nil {
		locationExists, err := s.employeeRepo.LocationExists(*employee.WorkLocationID)
		if err != nil {
			return apperrors.Wrap("DATABASE_ERROR", "Failed to validate work location", 500, err)
		}
		if !locationExists {
			return apperrors.Wrap("WORK_LOCATION_NOT_FOUND", "Work location not found", 404, nil)
		}
	}

	if employee.ManagerID != nil {
		if currentEmployeeID != uuid.Nil && *employee.ManagerID == currentEmployeeID {
			return apperrors.Wrap("INVALID_MANAGER", "Employee cannot be their own manager", 400, nil)
		}

		managerExists, err := s.employeeRepo.EmployeeExists(*employee.ManagerID)
		if err != nil {
			return apperrors.Wrap("DATABASE_ERROR", "Failed to validate manager", 500, err)
		}
		if !managerExists {
			return apperrors.Wrap("MANAGER_NOT_FOUND", "Manager employee not found", 404, nil)
		}

		managerBelongs, err := s.employeeRepo.EmployeeBelongsToCompany(*employee.ManagerID, employee.CompanyID)
		if err != nil {
			return apperrors.Wrap("DATABASE_ERROR", "Failed to validate manager company", 500, err)
		}
		if !managerBelongs {
			return apperrors.Wrap("MANAGER_COMPANY_MISMATCH", "Manager does not belong to the selected company", 409, nil)
		}
	}

	return nil
}

func NewEmployeeService(employeeRepo repo.IEmployeeRepository) EmployeeServiceInterface {
	return &EmployeeService{
		employeeRepo: employeeRepo,
	}
}
