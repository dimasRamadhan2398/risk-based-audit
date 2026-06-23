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
}

type EmployeeService struct {
	employeeRepo repo.IEmployeeRepository
}

func NewEmployeeService(employeeRepo repo.IEmployeeRepository) EmployeeServiceInterface {
	return &EmployeeService{
		employeeRepo: employeeRepo,
	}
}

func (s *EmployeeService) Create(ctx *base.BaseService, employee *models.Employee) (*models.Employee, error) {
	if _, err := s.employeeRepo.FindByCode(employee.EmployeeCode); err == nil {
		return nil, apperrors.Wrap("EMPLOYEE_CODE_ALREADY_EXISTS", "Employee code already exists", 409, nil)
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate employee code", 500, err)
	}

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

func (s *EmployeeService) Update(ctx *base.BaseService, id string, employee *models.Employee) (*models.Employee, error) {
	employeeID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.Wrap("INVALID_EMPLOYEE_ID", "Invalid employee ID format", 400, err)
	}

	existingEmployee, err := s.employeeRepo.FindByID(employeeID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			return nil, err
		}
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to fetch employee", 500, err)
	}

	if existingEmployee.EmployeeCode != employee.EmployeeCode {
		if _, err := s.employeeRepo.FindByCode(employee.EmployeeCode); err == nil {
			return nil, apperrors.Wrap("EMPLOYEE_CODE_ALREADY_EXISTS", "Employee code already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate employee code", 500, err)
		}
	}

	if existingEmployee.Email != employee.Email {
		if _, err := s.employeeRepo.FindByEmail(employee.Email); err == nil {
			return nil, apperrors.Wrap("EMPLOYEE_EMAIL_ALREADY_EXISTS", "Employee email already exists", 409, nil)
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to validate employee email", 500, err)
		}
	}

	existingEmployee.EmployeeCode = employee.EmployeeCode
	existingEmployee.FullName = employee.FullName
	existingEmployee.Email = employee.Email
	existingEmployee.Phone = employee.Phone
	existingEmployee.CompanyID = employee.CompanyID
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
	existingEmployee.JoinDate = employee.JoinDate

	if err := s.employeeRepo.Update(existingEmployee); err != nil {
		return nil, apperrors.Wrap("DATABASE_ERROR", "Failed to update employee", 500, err)
	}

	return existingEmployee, nil
}
