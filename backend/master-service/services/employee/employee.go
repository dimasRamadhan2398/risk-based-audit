package employee

import (
	stderrors "errors"
	"fmt"
	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	repo "master-service/repositories/employee"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type EmployeeServiceInterface interface {
	FindAll(ctx *base.BaseService) (*[]models.Employee, error)
	FindPage(ctx *base.BaseService, page, pageSize int, search string) (*EmployeePage, error)
	FindById(ctx *base.BaseService, id string) (*models.Employee, error)
	FindByCode(ctx *base.BaseService, code string) (*models.Employee, error)
	Create(ctx *base.BaseService, employee *models.Employee) (*models.Employee, error)
	NextEmployeeCode(ctx *base.BaseService) (string, error)
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

// Pagination limits for the employee list. An out-of-range page_size is clamped
// rather than rejected, so a caller asking for "everything" still gets a page.
const (
	defaultEmployeePageSize = 10
	maxEmployeePageSize     = 100
)

// EmployeePage is one page of the employee list. Page and PageSize are the
// normalized values actually used, not the raw request values.
type EmployeePage struct {
	Employees  []models.Employee
	Page       int
	PageSize   int
	Total      int64
	TotalPages int
}

var (
	errEmployeeCodeExists = apperrors.New(apperrors.CodeEmployeeCodeAlreadyExists,
		"An employee with this employee code already exists.", http.StatusConflict).
		WithFields(map[string]string{"employee_code": apperrors.FieldAlreadyExists})
	errEmployeeEmailExists = apperrors.New(apperrors.CodeEmployeeEmailAlreadyExists,
		"An employee with this email already exists.", http.StatusConflict).
		WithFields(map[string]string{"email": apperrors.FieldAlreadyExists})
)

const (
	// employeeCodePrefix and employeeCodeDigits define the generated format,
	// e.g. "EMP-0001". Numbers past 9999 simply gain digits ("EMP-10000").
	employeeCodePrefix = "EMP-"
	employeeCodeDigits = 4
	// legacyEmployeeCodePrefix is the pre-canonical format, e.g. "EMP001". It is
	// read when computing the next number so that existing legacy codes are not
	// reused, but it is never generated.
	legacyEmployeeCodePrefix = "EMP"
	// employeeCodeMaxAttempts bounds the regenerate-on-conflict loop used when
	// two creates race for the same next number.
	employeeCodeMaxAttempts = 5
	// defaultLevelGrade is stored when a create omits level_grade.
	defaultLevelGrade = 1
)

// employeeCodeConstraints are the unique indexes on employee_code. GORM's
// uniqueIndex creates the first name; the bare column UNIQUE from the SQL
// migration creates the second.
var employeeCodeConstraints = map[string]bool{
	"idx_employees_employee_code": true,
	"employees_employee_code_key": true,
}

// nextEmployeeCode returns the canonical code after the highest employee number
// in existing. Both "EMP-<n>" (canonical) and "EMP<n>" (legacy) codes count
// towards the maximum; the result is always "EMP-%04d". Gaps are not filled and
// codes in neither format are ignored. With no matching codes it returns "EMP-0001".
func nextEmployeeCode(existing []string) string {
	highest := 0
	for _, code := range existing {
		n, ok := employeeCodeNumber(code)
		if ok && n > highest {
			highest = n
		}
	}
	return fmt.Sprintf("%s%0*d", employeeCodePrefix, employeeCodeDigits, highest+1)
}

// employeeCodeNumber parses the number out of a canonical or legacy employee
// code. It reports false for anything else, including lower-case prefixes,
// signed or non-digit suffixes and numbers that overflow int.
func employeeCodeNumber(code string) (int, bool) {
	suffix, ok := strings.CutPrefix(code, employeeCodePrefix)
	if !ok {
		suffix, ok = strings.CutPrefix(code, legacyEmployeeCodePrefix)
		if !ok {
			return 0, false
		}
	}
	if !isDigits(suffix) {
		return 0, false
	}
	n, err := strconv.Atoi(suffix)
	if err != nil {
		return 0, false // overflow: not a code we would have generated
	}
	return n, true
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// isEmployeeCodeConflict reports whether err is the unique-index violation on
// employee_code (as opposed to email or any other constraint).
func isEmployeeCodeConflict(err error) bool {
	var pgErr *pgconn.PgError
	return stderrors.As(err, &pgErr) && pgErr.Code == "23505" && employeeCodeConstraints[pgErr.ConstraintName]
}

// normalizeAndValidate trims input and reports all missing required fields at once.
// employee_code is only required when requireCode is true (update).
func normalizeAndValidate(employee *models.Employee, requireCode bool) error {
	employee.EmployeeCode = strings.TrimSpace(employee.EmployeeCode)
	employee.Email = strings.TrimSpace(strings.ToLower(employee.Email))

	fields := map[string]string{}
	if requireCode && employee.EmployeeCode == "" {
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
	if err := normalizeAndValidate(employee, false); err != nil {
		return nil, err
	}
	if employee.LevelGrade == 0 {
		employee.LevelGrade = defaultLevelGrade
	}

	generateCode := employee.EmployeeCode == ""
	if !generateCode {
		if _, err := s.employeeRepo.FindByCode(employee.EmployeeCode); err == nil {
			return nil, errEmployeeCodeExists
		} else if err != apperrors.ErrNotFound {
			return nil, apperrors.DB(err, apperrors.OpRead, "Failed to validate the employee code.")
		}
	}

	if _, err := s.employeeRepo.FindByEmail(employee.Email); err == nil {
		return nil, errEmployeeEmailExists
	} else if err != apperrors.ErrNotFound {
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to validate the employee email.")
	}

	if generateCode {
		return s.createWithGeneratedCode(employee)
	}

	if err := s.employeeRepo.Create(employee); err != nil {
		return nil, apperrors.DB(err, apperrors.OpCreate, "Failed to create the employee.")
	}

	return employee, nil
}

// createWithGeneratedCode assigns the next EMP-NNNN code and inserts. If a
// concurrent create takes the same number, the unique index rejects this
// insert and the code is recomputed from fresh data, up to employeeCodeMaxAttempts.
func (s *EmployeeService) createWithGeneratedCode(employee *models.Employee) (*models.Employee, error) {
	var lastErr error
	for attempt := 0; attempt < employeeCodeMaxAttempts; attempt++ {
		codes, err := s.employeeRepo.FindCodesWithPrefix(employeeCodePrefix)
		if err != nil {
			return nil, apperrors.DB(err, apperrors.OpRead, "Failed to generate the employee code.")
		}
		employee.EmployeeCode = nextEmployeeCode(codes)

		err = s.employeeRepo.Create(employee)
		if err == nil {
			return employee, nil
		}
		if !isEmployeeCodeConflict(err) {
			return nil, apperrors.DB(err, apperrors.OpCreate, "Failed to create the employee.")
		}
		lastErr = err
	}

	return nil, apperrors.Internal("Failed to generate a unique employee code. Please try again.", lastErr)
}

// NextEmployeeCode previews the code the next generated create would assign.
// It only reads; nothing is reserved, so a concurrent create may take the
// number first. Create still resolves that race on insert.
func (s *EmployeeService) NextEmployeeCode(ctx *base.BaseService) (string, error) {
	if s.employeeRepo == nil {
		return "", apperrors.Internal("", nil)
	}
	codes, err := s.employeeRepo.FindCodesWithPrefix(employeeCodePrefix)
	if err != nil {
		return "", apperrors.DB(err, apperrors.OpRead, "Failed to generate the employee code.")
	}
	return nextEmployeeCode(codes), nil
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

// FindPage returns one page of employees matching search. A page below 1 becomes
// 1, a page_size below 1 becomes the default (10) and above 100 becomes 100.
// TotalPages is computed from the filtered total, so it is 0 when nothing matches.
func (s *EmployeeService) FindPage(ctx *base.BaseService, page, pageSize int, search string) (*EmployeePage, error) {
	if s.employeeRepo == nil {
		return nil, apperrors.Internal("", nil)
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultEmployeePageSize
	}
	if pageSize > maxEmployeePageSize {
		pageSize = maxEmployeePageSize
	}
	search = strings.TrimSpace(search)

	total, err := s.employeeRepo.Count(search)
	if err != nil {
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to load employees.")
	}
	rows, err := s.employeeRepo.FindPage((page-1)*pageSize, pageSize, search)
	if err != nil {
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to load employees.")
	}

	employees := make([]models.Employee, 0, len(rows))
	for _, employee := range rows {
		if employee != nil {
			employees = append(employees, *employee)
		}
	}

	return &EmployeePage{
		Employees:  employees,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	}, nil
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

// FindByCode returns the employee with the given employee_code. Auth-service
// uses it to resolve a login account's employee_id link (users.employee_id).
func (s *EmployeeService) FindByCode(ctx *base.BaseService, code string) (*models.Employee, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, apperrors.ErrEmployeeNotFound
	}

	employee, err := s.employeeRepo.FindByCode(code)
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
	if err := normalizeAndValidate(employee, true); err != nil {
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
	if employee.LevelGrade != 0 {
		existingEmployee.LevelGrade = employee.LevelGrade // omitted level_grade keeps the stored value
	}
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
