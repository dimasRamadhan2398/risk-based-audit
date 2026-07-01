package employee
import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)
type IEmployeeRepository interface {
	Create(employee *models.Employee) error
	Update(employee *models.Employee) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.Employee, error)
	FindByEmail(email string) (*models.Employee, error)
	FindByCode(code string) (*models.Employee, error)
	FindAll() ([]*models.Employee, error)
}
type EmployeeRepository struct { *repositories.BaseRepository }
func NewEmployeeRepository(db *gorm.DB) IEmployeeRepository { return &EmployeeRepository{BaseRepository: repositories.NewBaseRepository(db)} }
func (r *EmployeeRepository) Create(employee *models.Employee) error { return r.BaseRepository.Create(employee) }
func (r *EmployeeRepository) Update(employee *models.Employee) error { return r.BaseRepository.Update(employee) }
func (r *EmployeeRepository) Delete(id uuid.UUID) error { return r.BaseRepository.Delete(&models.Employee{ID: id}) }
func (r *EmployeeRepository) FindByID(id uuid.UUID) (*models.Employee, error) {
	var employee models.Employee
	if err := r.GetDB().First(&employee, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &employee, nil
}
func (r *EmployeeRepository) FindByEmail(email string) (*models.Employee, error) {
	var employee models.Employee
	if err := r.GetDB().Where("email = ?", email).First(&employee).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &employee, nil
}
func (r *EmployeeRepository) FindByCode(code string) (*models.Employee, error) {
	var employee models.Employee
	if err := r.GetDB().Where("employee_code = ?", code).First(&employee).Error; err != nil {
		if err == gorm.ErrRecordNotFound { return nil, apperrors.ErrNotFound }
		return nil, err
	}
	return &employee, nil
}
func (r *EmployeeRepository) FindAll() ([]*models.Employee, error) {
	var employees []*models.Employee
	if err := r.GetDB().Find(&employees).Error; err != nil { return nil, err }
	return employees, nil
}
