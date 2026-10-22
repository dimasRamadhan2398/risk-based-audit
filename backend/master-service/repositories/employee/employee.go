package employee

import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IEmployeeRepository interface {
	Create(employee *models.Employee) error
	Update(employee *models.Employee) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.Employee, error)
	FindByEmail(email string) (*models.Employee, error)
	FindByCode(code string) (*models.Employee, error)
	FindCodesWithPrefix(prefix string) ([]string, error)
	FindAll() ([]*models.Employee, error)
	FindPage(offset, limit int, search string) ([]*models.Employee, error)
	Count(search string) (int64, error)
}
type EmployeeRepository struct{ *repositories.BaseRepository }

func NewEmployeeRepository(db *gorm.DB) IEmployeeRepository {
	return &EmployeeRepository{BaseRepository: repositories.NewBaseRepository(db)}
}
func (r *EmployeeRepository) Create(employee *models.Employee) error {
	return r.BaseRepository.Create(employee)
}
func (r *EmployeeRepository) Update(employee *models.Employee) error {
	return r.BaseRepository.Update(employee)
}
func (r *EmployeeRepository) Delete(id uuid.UUID) error {
	return r.BaseRepository.Delete(&models.Employee{ID: id})
}
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

// FindCodesWithPrefix returns every employee_code starting with prefix,
// including soft-deleted rows: the unique index on employee_code still counts
// them, so a generated code must not reuse one.
func (r *EmployeeRepository) FindCodesWithPrefix(prefix string) ([]string, error) {
	var codes []string
	if err := r.GetDB().Unscoped().Model(&models.Employee{}).
		Where("employee_code LIKE ?", prefix+"%").
		Pluck("employee_code", &codes).Error; err != nil {
		return nil, err
	}
	return codes, nil
}
func (r *EmployeeRepository) FindAll() ([]*models.Employee, error) {
	var employees []*models.Employee
	if err := r.GetDB().Find(&employees).Error; err != nil {
		return nil, err
	}
	return employees, nil
}

// employeeSearchEscape is the LIKE escape character used by applyEmployeeSearch.
// It is "!" rather than the usual backslash so the literal needs no escaping.
const employeeSearchEscape = "!"

// likeEscaper escapes the LIKE wildcards in user input so "%" or "_" typed in
// the search box match literally.
var likeEscaper = strings.NewReplacer(
	employeeSearchEscape, employeeSearchEscape+employeeSearchEscape,
	"%", employeeSearchEscape+"%",
	"_", employeeSearchEscape+"_",
)

// applyEmployeeSearch narrows q to rows whose employee_code, full_name or email
// contains search, case-insensitively. The search text is bound as a parameter;
// the column names are fixed here and never come from the request.
func applyEmployeeSearch(q *gorm.DB, search string) *gorm.DB {
	if search == "" {
		return q
	}
	pattern := "%" + likeEscaper.Replace(strings.ToLower(search)) + "%"
	return q.Where(
		"LOWER(employee_code) LIKE ? ESCAPE '"+employeeSearchEscape+"' OR "+
			"LOWER(full_name) LIKE ? ESCAPE '"+employeeSearchEscape+"' OR "+
			"LOWER(email) LIKE ? ESCAPE '"+employeeSearchEscape+"'",
		pattern, pattern, pattern,
	)
}

// FindPage returns one page of non-deleted employees matching search, ordered
// by employee_code (unique) with id as a tie-breaker so pages are stable.
func (r *EmployeeRepository) FindPage(offset, limit int, search string) ([]*models.Employee, error) {
	var employees []*models.Employee
	q := applyEmployeeSearch(r.GetDB().Model(&models.Employee{}), search)
	if err := q.
		Order(clause.OrderByColumn{Column: clause.Column{Name: "employee_code"}}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).
		Offset(offset).Limit(limit).
		Find(&employees).Error; err != nil {
		return nil, err
	}
	return employees, nil
}

// Count returns the number of non-deleted employees matching search.
func (r *EmployeeRepository) Count(search string) (int64, error) {
	var count int64
	q := applyEmployeeSearch(r.GetDB().Model(&models.Employee{}), search)
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
