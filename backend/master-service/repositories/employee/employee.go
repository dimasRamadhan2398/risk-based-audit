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
	FindByName(name string) (*models.Employee, error)
	FindByResourceAndAction(resource, action string) (*models.Employee, error)
	FindAll() ([]*models.Employee, error)
	FindMany(offset, limit int, search string) ([]*models.Employee, error)
	Count(search string) (int64, error)
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

// FindByName finds a employee by name
func (r *EmployeeRepository) FindByName(name string) (*models.Employee, error) {
	var employee models.Employee
	if err := r.GetDB().Where("name = ?", name).First(&employee).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &employee, nil
}

// FindByResourceAndAction finds a employee by resource and action
func (r *EmployeeRepository) FindByResourceAndAction(resource, action string) (*models.Employee, error) {
	var employee models.Employee
	if err := r.GetDB().Where("resource = ? AND action = ?", resource, action).First(&employee).Error; err != nil {
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
		query = query.Where("name LIKE ? OR resource LIKE ? OR action LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if err := query.Offset(offset).Limit(limit).Find(&employees).Error; err != nil {
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
		query = query.Where("name LIKE ? OR resource LIKE ? OR action LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}
