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
	HardDelete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.Employee, error)
	FindByIDUnscoped(id uuid.UUID) (*models.Employee, error)
	FindReferences(id uuid.UUID) ([]models.EmployeeReference, error)
	FindByEmail(email string) (*models.Employee, error)
	FindByCode(code string) (*models.Employee, error)
	FindCodesWithPrefix(prefix string) ([]string, error)
	FindAll() ([]*models.Employee, error)
	FindPage(offset, limit int, filter EmployeeListFilter) ([]*models.Employee, error)
	Count(filter EmployeeListFilter) (int64, error)
}

// EmployeeListFilter narrows GET /employees. The zero value lists every
// non-deleted employee. Search matches employee_code, full_name or email
// (case-insensitive); DepartmentID and IsActive are exact matches when set.
type EmployeeListFilter struct {
	Search       string
	DepartmentID *uuid.UUID
	IsActive     *bool
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

// HardDelete removes the row permanently. Used to roll back an employee whose
// login account could not be created: a soft-deleted row would still hold the
// unique employee_code and email.
func (r *EmployeeRepository) HardDelete(id uuid.UUID) error {
	return r.GetDB().Unscoped().Delete(&models.Employee{ID: id}).Error
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

// FindByIDUnscoped is FindByID including soft-deleted rows, so a soft-deleted
// employee can still be hard-deleted.
func (r *EmployeeRepository) FindByIDUnscoped(id uuid.UUID) (*models.Employee, error) {
	var employee models.Employee
	if err := r.GetDB().Unscoped().First(&employee, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &employee, nil
}

// employeeReference is one column in master-service that holds an employee id.
// withDeleted counts soft-deleted rows too: those columns have a foreign key
// to employees (GORM belongs-to), which soft-deleted rows still hold.
type employeeReference struct {
	model       any
	table       string
	column      string
	label       string
	withDeleted bool
}

// employeeReferences lists every master-service column that stores an
// employee id. Keep it in step with the models: a column missing here is left
// dangling (or, with a foreign key, fails the delete with a generic 409).
var employeeReferences = []employeeReference{
	{&models.Employee{}, "employees", "manager_id", "Employees (direct manager)", true},
	{&models.Department{}, "departments", "pic_id", "Departments (head / PIC)", false},
	{&models.RiskRegister{}, "risk_registers", "risk_owner_id", "Risk registers (risk owner)", true},
	{&models.AnnualAuditPlan{}, "annual_audit_plans", "requested_by_id", "Annual audit plans (requested by)", true},
	{&models.AnnualAuditPlan{}, "annual_audit_plans", "approved_by_id", "Annual audit plans (approved by)", true},
	{&models.AuditFinding{}, "audit_findings", "owner_id", "Audit findings (owner)", true},
	{&models.AuditFinding{}, "audit_findings", "auditor_id", "Audit findings (auditor)", true},
	{&models.AuditIssue{}, "audit_issues", "owner_id", "Audit issues (owner)", true},
	{&models.AuditIssue{}, "audit_issues", "auditor_id", "Audit issues (auditor)", true},
	{&models.AuditRecommendation{}, "audit_recommendations", "responsible_id", "Audit recommendations (responsible)", true},
	{&models.AuditRecommendation{}, "audit_recommendations", "verified_by_id", "Audit recommendations (verified by)", true},
	{&models.AuditWorkpaper{}, "audit_workpapers", "auditor_id", "Audit workpapers (auditor)", true},
	{&models.AuditWorkpaper{}, "audit_workpapers", "reviewer_id", "Audit workpapers (reviewer)", true},
	{&models.Control{}, "controls", "owner_id", "Controls (owner)", true},
	{&models.ControlAssessment{}, "control_assessments", "auditor_id", "Control assessments (auditor)", true},
	{&models.MitigationAction{}, "mitigation_actions", "owner_id", "Mitigation actions (owner)", true},
	{&models.MitigationAction{}, "mitigation_actions", "verified_by_id", "Mitigation actions (verified by)", true},
	{&models.RiskIndicator{}, "risk_indicators", "owner_id", "Risk indicators (owner)", true},
}

// FindReferences returns the places in master-service that still point at the
// employee, one entry per column with a non-zero count, in a fixed order.
// Column names are fixed above and passed as clause.Column; only id is bound.
func (r *EmployeeRepository) FindReferences(id uuid.UUID) ([]models.EmployeeReference, error) {
	var refs []models.EmployeeReference
	for _, ref := range employeeReferences {
		q := r.GetDB().Model(ref.model)
		if ref.withDeleted {
			q = q.Unscoped()
		}
		q = q.Where(clause.Eq{Column: clause.Column{Name: ref.column}, Value: id})
		if ref.table == "employees" {
			// An employee is not "in use" because it is its own row.
			q = q.Where(clause.Neq{Column: clause.Column{Name: "id"}, Value: id})
		}
		var count int64
		if err := q.Count(&count).Error; err != nil {
			return nil, err
		}
		if count > 0 {
			refs = append(refs, models.EmployeeReference{Table: ref.table, Column: ref.column, Label: ref.label, Count: count})
		}
	}
	return refs, nil
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

// applyEmployeeListFilter adds the search and the typed department/active
// filters. Column names are fixed here; values are bound as parameters.
func applyEmployeeListFilter(q *gorm.DB, f EmployeeListFilter) *gorm.DB {
	q = applyEmployeeSearch(q, f.Search)
	if f.DepartmentID != nil {
		q = q.Where(clause.Eq{Column: clause.Column{Name: "department_id"}, Value: *f.DepartmentID})
	}
	if f.IsActive != nil {
		q = q.Where(clause.Eq{Column: clause.Column{Name: "is_active"}, Value: *f.IsActive})
	}
	return q
}

// FindPage returns one page of non-deleted employees matching filter, ordered
// by employee_code (unique) with id as a tie-breaker so pages are stable.
func (r *EmployeeRepository) FindPage(offset, limit int, filter EmployeeListFilter) ([]*models.Employee, error) {
	var employees []*models.Employee
	q := applyEmployeeListFilter(r.GetDB().Model(&models.Employee{}), filter)
	if err := q.
		Order(clause.OrderByColumn{Column: clause.Column{Name: "employee_code"}}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}}).
		Offset(offset).Limit(limit).
		Find(&employees).Error; err != nil {
		return nil, err
	}
	return employees, nil
}

// Count returns the number of non-deleted employees matching filter.
func (r *EmployeeRepository) Count(filter EmployeeListFilter) (int64, error) {
	var count int64
	q := applyEmployeeListFilter(r.GetDB().Model(&models.Employee{}), filter)
	if err := q.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
