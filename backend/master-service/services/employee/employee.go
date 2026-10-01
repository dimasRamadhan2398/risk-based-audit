package employee

import (
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/employee"
	"net/http"
	"strings"

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

var (
	errEmployeeCodeExists = apperrors.New(apperrors.CodeEmployeeCodeAlreadyExists,
		"An employee with this employee code already exists.", http.StatusConflict).
		WithFields(map[string]string{"employee_code": apperrors.FieldAlreadyExists})
	errEmployeeEmailExists = apperrors.New(apperrors.CodeEmployeeEmailAlreadyExists,
		"An employee with this email already exists.", http.StatusConflict).
		WithFields(map[string]string{"email": apperrors.FieldAlreadyExists})
)

// normalizeAndValidate trims input and reports all missing required fields at once.
func normalizeAndValidate(employee *models.Employee) error {
	employee.EmployeeCode = strings.TrimSpace(employee.EmployeeCode)
	employee.Email = strings.TrimSpace(strings.ToLower(employee.Email))

	fields := map[string]string{}
	if employee.EmployeeCode == "" {
		fields["employee_code"] = apperrors.FieldRequired
	}
	if employee.Email == "" {
		fields["email"] = apperrors.FieldRequired
	}
	if len(fields) > 0 {
		return apperrors.ValidationFailed("", fields)
	}
	return nil
}

func (s *EmployeeService) Create(ctx *base.BaseService, employee *models.Employee) (*models.Employee, error) {
	if s.employeeRepo == nil {
		return nil, apperrors.Internal("", nil)
	}
	if employee == nil {
		return nil, apperrors.New(apperrors.CodeInvalidRequestBody, apperrors.MsgInvalidRequestBody, http.StatusBadRequest)
	}
	if err := normalizeAndValidate(employee); err != nil {
		return nil, err
	}

	if _, err := s.employeeRepo.FindByCode(employee.EmployeeCode); err == nil {
		return nil, errEmployeeCodeExists
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to validate the employee code.")
	}

	if _, err := s.employeeRepo.FindByEmail(employee.Email); err == nil {
		return nil, errEmployeeEmailExists
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to validate the employee email.")
	}

	if err := s.employeeRepo.Create(employee); err != nil {
		return nil, apperrors.DB(err, apperrors.OpCreate, "Failed to create the employee.")
	}

	return employee, nil
}

func (s *EmployeeService) Delete(ctx *base.BaseService, id string) error {
	employeeID, err := uuid.Parse(id)
	if err != nil {
		return apperrors.InvalidID(err)
	}

	if _, err := s.employeeRepo.FindByID(employeeID); err != nil {
		if err == apperrors.ErrNotFound {
			return apperrors.ErrEmployeeNotFound
		}
		return apperrors.DB(err, apperrors.OpRead, "Failed to load the employee.")
	}

	if err := s.employeeRepo.Delete(employeeID); err != nil {
		return apperrors.DB(err, apperrors.OpDelete, "Failed to delete the employee.")
	}

	return nil
}

func (s *EmployeeService) FindAll(ctx *base.BaseService) (*[]models.Employee, error) {
	employees, err := s.employeeRepo.FindAll()
	if err != nil {
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to load employees.")
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
		return nil, apperrors.InvalidID(err)
	}

	employee, err := s.employeeRepo.FindByID(employeeID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			return nil, apperrors.ErrEmployeeNotFound
		}
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to load the employee.")
	}

	return employee, nil
}

func (s *EmployeeService) Update(ctx *base.BaseService, id string, employee *models.Employee) (*models.Employee, error) {
	if s.employeeRepo == nil {
		return nil, apperrors.Internal("", nil)
	}
	if employee == nil {
		return nil, apperrors.New(apperrors.CodeInvalidRequestBody, apperrors.MsgInvalidRequestBody, http.StatusBadRequest)
	}
	if err := normalizeAndValidate(employee); err != nil {
		return nil, err
	}

	employeeID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.InvalidID(err)
	}

	existingEmployee, err := s.employeeRepo.FindByID(employeeID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			return nil, apperrors.ErrEmployeeNotFound
		}
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to load the employee.")
	}

	if existingEmployee.EmployeeCode != employee.EmployeeCode {
		if _, err := s.employeeRepo.FindByCode(employee.EmployeeCode); err == nil {
			return nil, errEmployeeCodeExists
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.DB(err, apperrors.OpRead, "Failed to validate the employee code.")
		}
	}

	if existingEmployee.Email != employee.Email {
		if _, err := s.employeeRepo.FindByEmail(employee.Email); err == nil {
			return nil, errEmployeeEmailExists
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.DB(err, apperrors.OpRead, "Failed to validate the employee email.")
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
		return nil, apperrors.DB(err, apperrors.OpUpdate, "Failed to update the employee.")
	}

	return existingEmployee, nil
}
