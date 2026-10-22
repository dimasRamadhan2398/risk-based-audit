package employee

import (
	"errors"
	"testing"

	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// fakeRepo is an in-memory IEmployeeRepository. codes feeds FindCodesWithPrefix;
// createErrs (if set) is consumed one per Create call before inserts succeed.
type fakeRepo struct {
	codes       []string
	codesErr    error
	byCode      map[string]bool
	byEmail     map[string]bool
	createErrs  []error
	created     []*models.Employee
	codeLookups int
	employees   []*models.Employee // seeded rows for FindPage/Count (employee_page_test.go)
	listErr     error              // returned by Count and FindPage when set
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{byCode: map[string]bool{}, byEmail: map[string]bool{}}
}

func (f *fakeRepo) Create(e *models.Employee) error {
	if len(f.createErrs) > 0 {
		err := f.createErrs[0]
		f.createErrs = f.createErrs[1:]
		if err != nil {
			return err
		}
	}
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	f.codes = append(f.codes, e.EmployeeCode)
	f.byCode[e.EmployeeCode] = true
	f.byEmail[e.Email] = true
	cp := *e
	f.created = append(f.created, &cp)
	return nil
}
func (f *fakeRepo) Update(*models.Employee) error { return nil }
func (f *fakeRepo) Delete(uuid.UUID) error        { return nil }
func (f *fakeRepo) FindByID(uuid.UUID) (*models.Employee, error) {
	return nil, apperrors.ErrNotFound
}
func (f *fakeRepo) FindByEmail(email string) (*models.Employee, error) {
	if f.byEmail[email] {
		return &models.Employee{Email: email}, nil
	}
	return nil, apperrors.ErrNotFound
}
func (f *fakeRepo) FindByCode(code string) (*models.Employee, error) {
	if f.byCode[code] {
		return &models.Employee{EmployeeCode: code}, nil
	}
	return nil, apperrors.ErrNotFound
}
func (f *fakeRepo) FindCodesWithPrefix(prefix string) ([]string, error) {
	f.codeLookups++
	if f.codesErr != nil {
		return nil, f.codesErr
	}
	var out []string
	for _, c := range f.codes {
		if len(c) >= len(prefix) && c[:len(prefix)] == prefix {
			out = append(out, c)
		}
	}
	return out, nil
}
func (f *fakeRepo) FindAll() ([]*models.Employee, error) { return nil, nil }

func codeConflict(constraint string) error {
	return &pgconn.PgError{Code: "23505", ConstraintName: constraint, Message: "duplicate key value violates unique constraint"}
}

func TestNextEmployeeCode(t *testing.T) {
	tests := []struct {
		name     string
		existing []string
		want     string
	}{
		{"no rows", nil, "EMP-0001"},
		{"empty slice", []string{}, "EMP-0001"},
		{"first code after EMP-0001", []string{"EMP-0001"}, "EMP-0002"},
		{"increments from highest", []string{"EMP-0001", "EMP-0002", "EMP-0009"}, "EMP-0010"},
		{"input order does not matter", []string{"EMP-0009", "EMP-0002", "EMP-0001"}, "EMP-0010"},
		{"gap is not filled", []string{"EMP-0001", "EMP-0005"}, "EMP-0006"},
		{"leading zeros parsed as numbers", []string{"EMP-0010"}, "EMP-0011"},
		{"crosses into five digits", []string{"EMP-9999"}, "EMP-10000"},
		{"six digits keep growing", []string{"EMP-99999"}, "EMP-100000"},
		{"ignores other prefixes", []string{"X-0500", "emp-0100", "AUD-0900"}, "EMP-0001"},
		{"legacy only", []string{"EMP001", "EMP002", "EMP005"}, "EMP-0006"},
		{"legacy only, no padding", []string{"EMP7", "EMP12"}, "EMP-0013"},
		{"canonical only", []string{"EMP-0001", "EMP-0007"}, "EMP-0008"},
		{"mixed legacy and canonical takes the higher", []string{"EMP001", "EMP-0009", "EMP005"}, "EMP-0010"},
		{"mixed, legacy higher", []string{"EMP-0003", "EMP0012"}, "EMP-0013"},
		{"legacy code with no canonical is still counted", []string{"EMP-0002", "EMP0100"}, "EMP-0101"},
		{"legacy zero counts as zero", []string{"EMP000"}, "EMP-0001"},
		{"legacy ignores non-digit suffix", []string{"EMP00A", "EMP-", "EMP", "EMPX001"}, "EMP-0001"},
		{"ignores non-numeric suffix", []string{"EMP-ABC", "EMP-0007A", "EMP-", "EMP-0003-OLD"}, "EMP-0001"},
		{"ignores signed suffix", []string{"EMP--1", "EMP-+5"}, "EMP-0001"},
		{"ignores suffix that overflows int", []string{"EMP-99999999999999999999999"}, "EMP-0001"},
		{"non-matching codes do not affect max", []string{"EMP-0004", "EMP-ABC", "EMP-0002"}, "EMP-0005"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nextEmployeeCode(tt.existing); got != tt.want {
				t.Fatalf("nextEmployeeCode(%v) = %q, want %q", tt.existing, got, tt.want)
			}
		})
	}
}

func baseEmployee(code string) *models.Employee {
	return &models.Employee{
		EmployeeCode: code,
		FullName:     "Budi",
		Email:        "Budi@Example.com",
	}
}

func TestCreate_OmittedCodeIsGeneratedFromHighest(t *testing.T) {
	repo := newFakeRepo()
	repo.codes = []string{"EMP-0001", "EMP-0007", "EMP-0003"}
	svc := NewEmployeeService(repo)

	got, err := svc.Create(base.NewBaseService(), baseEmployee(""))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.EmployeeCode != "EMP-0008" {
		t.Fatalf("EmployeeCode = %q, want EMP-0008", got.EmployeeCode)
	}
	if got.Email != "budi@example.com" {
		t.Fatalf("email not normalized: %q", got.Email)
	}
}

func TestCreate_OmittedCodeWithNoRowsIsFirstCode(t *testing.T) {
	repo := newFakeRepo()
	svc := NewEmployeeService(repo)

	got, err := svc.Create(base.NewBaseService(), baseEmployee("   "))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.EmployeeCode != "EMP-0001" {
		t.Fatalf("EmployeeCode = %q, want EMP-0001", got.EmployeeCode)
	}
}

func TestCreate_OmittedLevelGradeDefaultsToOne(t *testing.T) {
	repo := newFakeRepo()
	svc := NewEmployeeService(repo)

	got, err := svc.Create(base.NewBaseService(), baseEmployee(""))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.LevelGrade != 1 {
		t.Fatalf("LevelGrade = %d, want 1", got.LevelGrade)
	}
}

func TestCreate_ExplicitLevelGradeIsKept(t *testing.T) {
	repo := newFakeRepo()
	svc := NewEmployeeService(repo)
	e := baseEmployee("")
	e.LevelGrade = 7

	got, err := svc.Create(base.NewBaseService(), e)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.LevelGrade != 7 {
		t.Fatalf("LevelGrade = %d, want 7", got.LevelGrade)
	}
}

func TestCreate_ExplicitCodeIsUsedAsIs(t *testing.T) {
	repo := newFakeRepo()
	repo.codes = []string{"EMP-0001"}
	svc := NewEmployeeService(repo)

	got, err := svc.Create(base.NewBaseService(), baseEmployee("  CUSTOM-42 "))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.EmployeeCode != "CUSTOM-42" {
		t.Fatalf("EmployeeCode = %q, want CUSTOM-42", got.EmployeeCode)
	}
	if repo.codeLookups != 0 {
		t.Fatalf("generator ran for an explicit code (%d lookups)", repo.codeLookups)
	}
}

func TestCreate_ExplicitDuplicateCodeStillConflicts(t *testing.T) {
	repo := newFakeRepo()
	repo.byCode["EMP-0001"] = true
	svc := NewEmployeeService(repo)

	_, err := svc.Create(base.NewBaseService(), baseEmployee("EMP-0001"))
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeEmployeeCodeAlreadyExists {
		t.Fatalf("err = %v, want EMPLOYEE_CODE_ALREADY_EXISTS", err)
	}
	if len(repo.created) != 0 {
		t.Fatalf("employee was inserted despite duplicate code")
	}
}

func TestCreate_GeneratedCodeRetriesOnCodeConflict(t *testing.T) {
	repo := newFakeRepo()
	repo.codes = []string{"EMP-0004"}
	// First insert loses a race for EMP-0005; by the retry, another create
	// has taken EMP-0005 so the generator must move on to EMP-0006.
	repo.createErrs = []error{codeConflict("idx_employees_employee_code")}
	svc := NewEmployeeService(&raceRepo{fakeRepo: repo, onRetry: "EMP-0005"})

	got, err := svc.Create(base.NewBaseService(), baseEmployee(""))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.EmployeeCode != "EMP-0006" {
		t.Fatalf("EmployeeCode = %q, want EMP-0006", got.EmployeeCode)
	}
	if repo.codeLookups != 2 {
		t.Fatalf("codeLookups = %d, want 2", repo.codeLookups)
	}
}

func TestCreate_GeneratedCodeRetryUsesBareColumnConstraintName(t *testing.T) {
	repo := newFakeRepo()
	repo.createErrs = []error{codeConflict("employees_employee_code_key")}
	svc := NewEmployeeService(repo)

	got, err := svc.Create(base.NewBaseService(), baseEmployee(""))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.EmployeeCode != "EMP-0001" {
		t.Fatalf("EmployeeCode = %q, want EMP-0001", got.EmployeeCode)
	}
}

func TestCreate_EmailConflictDuringInsertIsNotRetried(t *testing.T) {
	repo := newFakeRepo()
	repo.createErrs = []error{codeConflict("idx_employees_email")}
	svc := NewEmployeeService(repo)

	_, err := svc.Create(base.NewBaseService(), baseEmployee(""))
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeEmployeeEmailAlreadyExists {
		t.Fatalf("err = %v, want EMPLOYEE_EMAIL_ALREADY_EXISTS", err)
	}
	if repo.codeLookups != 1 {
		t.Fatalf("codeLookups = %d, want 1 (no retry)", repo.codeLookups)
	}
}

func TestCreate_GeneratedCodeGivesUpAfterMaxAttempts(t *testing.T) {
	repo := newFakeRepo()
	for i := 0; i < employeeCodeMaxAttempts; i++ {
		repo.createErrs = append(repo.createErrs, codeConflict("idx_employees_employee_code"))
	}
	svc := NewEmployeeService(repo)

	_, err := svc.Create(base.NewBaseService(), baseEmployee(""))
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("err = %v, want AppError", err)
	}
	if appErr.Code == apperrors.CodeEmployeeCodeAlreadyExists || appErr.StatusCode != 500 {
		t.Fatalf("code %q status %d; want non-conflict 500", appErr.Code, appErr.StatusCode)
	}
	if repo.codeLookups != employeeCodeMaxAttempts {
		t.Fatalf("codeLookups = %d, want %d", repo.codeLookups, employeeCodeMaxAttempts)
	}
}

func TestCreate_CodeLookupFailureIsReported(t *testing.T) {
	repo := newFakeRepo()
	repo.codesErr = errors.New("connection reset")
	svc := NewEmployeeService(repo)

	_, err := svc.Create(base.NewBaseService(), baseEmployee(""))
	if err == nil || len(repo.created) != 0 {
		t.Fatalf("err = %v, created = %d; want failure and no insert", err, len(repo.created))
	}
}

func TestUpdate_CodeStillRequired(t *testing.T) {
	svc := NewEmployeeService(newFakeRepo())

	_, err := svc.Update(base.NewBaseService(), uuid.New().String(), baseEmployee(""))
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperrors.CodeValidationFailed {
		t.Fatalf("err = %v, want VALIDATION_FAILED", err)
	}
	if appErr.Fields["employee_code"] != apperrors.FieldRequired {
		t.Fatalf("fields = %v", appErr.Fields)
	}
}

// raceRepo simulates a concurrent create that takes onRetry between the
// generator's first and second read.
type raceRepo struct {
	*fakeRepo
	onRetry string
}

func (r *raceRepo) FindCodesWithPrefix(prefix string) ([]string, error) {
	if r.codeLookups == 1 {
		r.codes = append(r.codes, r.onRetry)
	}
	return r.fakeRepo.FindCodesWithPrefix(prefix)
}

func TestNextEmployeeCode_NoRowsIsFirstCode(t *testing.T) {
	repo := newFakeRepo()
	svc := NewEmployeeService(repo)

	got, err := svc.NextEmployeeCode(base.NewBaseService())
	if err != nil {
		t.Fatalf("NextEmployeeCode: %v", err)
	}
	if got != "EMP-0001" {
		t.Fatalf("NextEmployeeCode = %q, want EMP-0001", got)
	}
}

func TestNextEmployeeCode_IncrementsFromHighestAndDoesNotReserve(t *testing.T) {
	repo := newFakeRepo()
	repo.codes = []string{"EMP-0001", "EMP-0009"}
	svc := NewEmployeeService(repo)

	got, err := svc.NextEmployeeCode(base.NewBaseService())
	if err != nil {
		t.Fatalf("NextEmployeeCode: %v", err)
	}
	if got != "EMP-0010" {
		t.Fatalf("NextEmployeeCode = %q, want EMP-0010", got)
	}
	if len(repo.created) != 0 || len(repo.codes) != 2 {
		t.Fatalf("preview must not insert: created=%d codes=%v", len(repo.created), repo.codes)
	}
	// Repeated previews return the same value until a row is actually created.
	again, _ := svc.NextEmployeeCode(base.NewBaseService())
	if again != "EMP-0010" {
		t.Fatalf("second preview = %q, want EMP-0010", again)
	}
}

func TestNextEmployeeCode_MatchesCreate(t *testing.T) {
	repo := newFakeRepo()
	repo.codes = []string{"EMP-0041", "EMP-0042"}
	svc := NewEmployeeService(repo)

	preview, err := svc.NextEmployeeCode(base.NewBaseService())
	if err != nil {
		t.Fatalf("NextEmployeeCode: %v", err)
	}
	created, err := svc.Create(base.NewBaseService(), baseEmployee(""))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if preview != created.EmployeeCode {
		t.Fatalf("preview %q != created %q", preview, created.EmployeeCode)
	}
}

func TestNextEmployeeCode_LookupFailureIsDBRead(t *testing.T) {
	repo := newFakeRepo()
	lookupErr := errors.New("connection refused")
	repo.codesErr = lookupErr
	svc := NewEmployeeService(repo)

	_, err := svc.NextEmployeeCode(base.NewBaseService())
	var appErr *apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("err = %v, want AppError", err)
	}
	if appErr.Message != "Failed to generate the employee code." {
		t.Fatalf("message = %q", appErr.Message)
	}
	if !errors.Is(appErr, lookupErr) && appErr.Err != lookupErr {
		t.Fatalf("underlying error not kept: %v", appErr.Err)
	}
}
