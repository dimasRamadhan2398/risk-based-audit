package department

import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IDepartmentRepository = DepartmentRepositoryInterface

// DepartmentRepositoryInterface defines the company repository interface
type DepartmentRepositoryInterface interface {
	Create(company *models.Department) error
	Update(company *models.Department) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.Department, error)
	FindByCode(code string) (*models.Department, error)
	FindByName(name string) (*models.Department, error)
	FindAll() ([]*models.Department, error)
	FindMany(offset, limit int, search string) ([]*models.Department, error)
	Count(search string) (int64, error)
	CompanyExists(id uuid.UUID) (bool, error)
	BusinessUnitExists(id uuid.UUID) (bool, error)
	BusinessUnitBelongsToCompany(businessUnitID uuid.UUID, companyID uuid.UUID) (bool, error)
	EmployeeExists(id uuid.UUID) (bool, error)
	EmployeeBelongsToCompany(employeeID uuid.UUID, companyID uuid.UUID) (bool, error)
}

// DepartmentRepository handles company data operations
type DepartmentRepository struct {
	*repositories.BaseRepository
}

// NewDepartmentRepository creates a new company repository
func NewDepartmentRepository(db *gorm.DB) IDepartmentRepository {
	return &DepartmentRepository{
		BaseRepository: repositories.NewBaseRepository(db),
	}
}

// Create creates a new company
func (r *DepartmentRepository) Create(company *models.Department) error {
	return r.BaseRepository.Create(company)
}

// Update updates a company
func (r *DepartmentRepository) Update(company *models.Department) error {
	return r.BaseRepository.Update(company)
}

// Delete deletes a company
func (r *DepartmentRepository) Delete(id uuid.UUID) error {
	return r.BaseRepository.Delete(&models.Department{ID: id})
}

// FindByID finds a company by ID
func (r *DepartmentRepository) FindByID(id uuid.UUID) (*models.Department, error) {
	var company models.Department
	if err := r.GetDB().First(&company, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &company, nil
}

// FindByCode finds a company by code
func (r *DepartmentRepository) FindByCode(code string) (*models.Department, error) {
	var company models.Department
	if err := r.GetDB().Where("department_code = ?", code).First(&company).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &company, nil
}

// FindByName finds a company by name
func (r *DepartmentRepository) FindByName(name string) (*models.Department, error) {
	var company models.Department
	if err := r.GetDB().Where("department_name = ?", name).First(&company).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &company, nil
}

// FindAll finds all companys
func (r *DepartmentRepository) FindAll() ([]*models.Department, error) {
	var companys []*models.Department
	if err := r.GetDB().Find(&companys).Error; err != nil {
		return nil, err
	}
	return companys, nil
}

// FindMany finds multiple companys with pagination
func (r *DepartmentRepository) FindMany(offset, limit int, search string) ([]*models.Department, error) {
	var companys []*models.Department
	query := r.GetDB().Model(&models.Department{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			"department_code LIKE ? OR department_name LIKE ? OR department_description LIKE ?",
			searchPattern, searchPattern, searchPattern,
		)
	}

	if err := query.Offset(offset).Limit(limit).Find(&companys).Error; err != nil {
		return nil, err
	}

	return companys, nil
}

// Count counts companys with filters
func (r *DepartmentRepository) Count(search string) (int64, error) {
	var count int64
	query := r.GetDB().Model(&models.Department{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			"department_code LIKE ? OR department_name LIKE ? OR department_description LIKE ?",
			searchPattern, searchPattern, searchPattern,
		)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *DepartmentRepository) CompanyExists(id uuid.UUID) (bool, error) {
	var count int64
	if err := r.GetDB().Model(&models.Company{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *DepartmentRepository) BusinessUnitExists(id uuid.UUID) (bool, error) {
	var count int64
	if err := r.GetDB().Model(&models.BusinessUnit{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *DepartmentRepository) BusinessUnitBelongsToCompany(businessUnitID uuid.UUID, companyID uuid.UUID) (bool, error) {
	var count int64
	if err := r.GetDB().Model(&models.BusinessUnit{}).
		Where("id = ? AND company_id = ?", businessUnitID, companyID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *DepartmentRepository) EmployeeExists(id uuid.UUID) (bool, error) {
	var count int64
	if err := r.GetDB().Model(&models.Employee{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *DepartmentRepository) EmployeeBelongsToCompany(employeeID uuid.UUID, companyID uuid.UUID) (bool, error) {
	var count int64
	if err := r.GetDB().Model(&models.Employee{}).
		Where("id = ? AND company_id = ?", employeeID, companyID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
