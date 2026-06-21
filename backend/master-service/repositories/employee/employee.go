package employee

import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IEmployeeRepository = EmployeeRepositoryInterface

// EmployeeRepositoryInterface defines the employee repository interface
type EmployeeRepositoryInterface interface {
	Create(employee *models.Employee) error
	Update(employee *models.Employee) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.Employee, error)
	FindByCode(code string) (*models.Employee, error)
	FindByEmail(email string) (*models.Employee, error)
	FindAll() ([]*models.Employee, error)
	FindMany(offset, limit int, search string) ([]*models.Employee, error)
	Count(search string) (int64, error)
	CompanyExists(id uuid.UUID) (bool, error)
	DepartmentExists(id uuid.UUID) (bool, error)
	DepartmentBelongsToCompany(departmentID uuid.UUID, companyID uuid.UUID) (bool, error)
	JobRoleExists(id uuid.UUID) (bool, error)
	LocationExists(id uuid.UUID) (bool, error)
	EmployeeExists(id uuid.UUID) (bool, error)
	EmployeeBelongsToCompany(employeeID uuid.UUID, companyID uuid.UUID) (bool, error)
}

// EmployeeRepository handles employee data operations
type EmployeeRepository struct {
	*repositories.BaseRepository
}

// NewEmployeeRepository creates a new employee repository
func NewEmployeeRepository(db *gorm.DB) IEmployeeRepository {
	return &EmployeeRepository{
		BaseRepository: repositories.NewBaseRepository(db),
	}
}

// Create creates a new employee
func (r *EmployeeRepository) Create(employee *models.Employee) error {
	return r.BaseRepository.Create(employee)
}

// Update updates a employee
func (r *EmployeeRepository) Update(employee *models.Employee) error {
	return r.BaseRepository.Update(employee)
}

// Delete deletes a employee
func (r *EmployeeRepository) Delete(id uuid.UUID) error {
	return r.BaseRepository.Delete(&models.Employee{ID: id})
}

// FindByID finds a employee by ID
func (r *EmployeeRepository) FindByID(id uuid.UUID) (*models.Employee, error) {
	var employee models.Employee
	if err := r.GetDB().First(&employee, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &employee, nil
}

// FindByCode finds a employee by employee code
func (r *EmployeeRepository) FindByCode(code string) (*models.Employee, error) {
	var employee models.Employee
	if err := r.GetDB().Where("employee_code = ?", code).First(&employee).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &employee, nil
}

// FindByEmail finds a employee by email
func (r *EmployeeRepository) FindByEmail(email string) (*models.Employee, error) {
	var employee models.Employee
	if err := r.GetDB().Where("email = ?", email).First(&employee).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &employee, nil
}

// FindAll finds all employees
func (r *EmployeeRepository) FindAll() ([]*models.Employee, error) {
	var employees []*models.Employee
	if err := r.GetDB().Find(&employees).Error; err != nil {
		return nil, err
	}
	return employees, nil
}

// FindMany finds multiple employees with pagination
func (r *EmployeeRepository) FindMany(offset, limit int, search string) ([]*models.Employee, error) {
	var employees []*models.Employee
	query := r.GetDB().Model(&models.Employee{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("employee_code LIKE ? OR full_name LIKE ? OR email LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&employees).Error; err != nil {
		return nil, err
	}

	return employees, nil
}

// Count counts employees with filters
func (r *EmployeeRepository) Count(search string) (int64, error) {
	var count int64
	query := r.GetDB().Model(&models.Employee{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("employee_code LIKE ? OR full_name LIKE ? OR email LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *EmployeeRepository) CompanyExists(id uuid.UUID) (bool, error) {
	var count int64
	if err := r.GetDB().Model(&models.Company{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *EmployeeRepository) DepartmentExists(id uuid.UUID) (bool, error) {
	var count int64
	if err := r.GetDB().Model(&models.Department{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *EmployeeRepository) DepartmentBelongsToCompany(departmentID uuid.UUID, companyID uuid.UUID) (bool, error) {
	var count int64
	if err := r.GetDB().Model(&models.Department{}).
		Where("id = ? AND company_id = ?", departmentID, companyID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *EmployeeRepository) JobRoleExists(id uuid.UUID) (bool, error) {
	var count int64
	if err := r.GetDB().Model(&models.JobRole{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *EmployeeRepository) LocationExists(id uuid.UUID) (bool, error) {
	var count int64
	if err := r.GetDB().Model(&models.Location{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *EmployeeRepository) EmployeeExists(id uuid.UUID) (bool, error) {
	var count int64
	if err := r.GetDB().Model(&models.Employee{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *EmployeeRepository) EmployeeBelongsToCompany(employeeID uuid.UUID, companyID uuid.UUID) (bool, error) {
	var count int64
	if err := r.GetDB().Model(&models.Employee{}).
		Where("id = ? AND company_id = ?", employeeID, companyID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
