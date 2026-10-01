package errors

import (
	stderrors "errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// repoWrap mirrors repositories.dbErr: ErrDatabase with the cause kept in the chain.
func repoWrap(err error) error { return fmt.Errorf("%w: %w", ErrDatabase, err) }

// serviceWrap mirrors legacy service code: Wrap("DATABASE_ERROR", ..., 500, err).
func serviceWrap(err error) error {
	return Wrap("DATABASE_ERROR", "Failed to create employee", http.StatusInternalServerError, repoWrap(err))
}

func pgErr(code, constraint string) *pgconn.PgError {
	return &pgconn.PgError{
		Severity:       "ERROR",
		Code:           code,
		Message:        `duplicate key value violates unique constraint "` + constraint + `"`,
		ConstraintName: constraint,
	}
}

func assertSafe(t *testing.T, e *AppError) {
	t.Helper()
	for _, bad := range []string{"SQLSTATE", "duplicate key", "violates", "ERROR:", "record not found", "*pgconn", "gorm"} {
		if strings.Contains(e.Message, bad) {
			t.Fatalf("message leaks technical text %q: %q", bad, e.Message)
		}
	}
}

func TestFromDB_UniqueViolation(t *testing.T) {
	tests := []struct {
		name       string
		constraint string
		wantCode   string
		wantField  string
	}{
		{"employee code", "idx_employees_employee_code", CodeEmployeeCodeAlreadyExists, "employee_code"},
		{"employee email", "idx_employees_email", CodeEmployeeEmailAlreadyExists, "email"},
		{"department code", "idx_departments_department_code", CodeDepartmentCodeAlreadyExists, "department_code"},
		{"department name", "idx_departments_department_name", CodeDepartmentNameAlreadyExists, "department_name"},
		{"company code", "idx_companies_company_code", CodeCompanyCodeAlreadyExists, "code"},
		{"company tax id", "idx_companies_tax_id", CodeCompanyTaxIDAlreadyExists, "tax_id"},
		{"unknown constraint", "idx_something_else", CodeDuplicateEntry, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, wrapped := range []error{pgErr("23505", tt.constraint), repoWrap(pgErr("23505", tt.constraint)), serviceWrap(pgErr("23505", tt.constraint))} {
				got := FromDB(wrapped, OpCreate)
				if got == nil {
					t.Fatalf("FromDB returned nil for %v", wrapped)
				}
				if got.StatusCode != http.StatusConflict || got.Code != tt.wantCode {
					t.Fatalf("got %d %s, want 409 %s", got.StatusCode, got.Code, tt.wantCode)
				}
				if tt.wantField != "" && got.Fields[tt.wantField] != FieldAlreadyExists {
					t.Fatalf("fields = %v, want %s=ALREADY_EXISTS", got.Fields, tt.wantField)
				}
				if tt.wantField == "" && got.Fields != nil {
					t.Fatalf("unexpected fields %v", got.Fields)
				}
				var cause *pgconn.PgError
				if !stderrors.As(got, &cause) {
					t.Fatalf("original error not kept for logging")
				}
				assertSafe(t, got)
			}
		})
	}
}

func TestFromDB_ForeignKeyViolation(t *testing.T) {
	fk := &pgconn.PgError{Code: "23503", ConstraintName: "fk_employees_department",
		Message: `insert or update on table "employees" violates foreign key constraint "fk_employees_department"`}

	create := FromDB(repoWrap(fk), OpCreate)
	if create.StatusCode != http.StatusUnprocessableEntity || create.Code != CodeReferenceNotFound {
		t.Fatalf("create: got %d %s", create.StatusCode, create.Code)
	}
	if create.Fields["department_id"] != FieldNotFound {
		t.Fatalf("create: fields = %v", create.Fields)
	}
	assertSafe(t, create)

	update := FromDB(repoWrap(fk), OpUpdate)
	if update.Code != CodeReferenceNotFound {
		t.Fatalf("update: got %s", update.Code)
	}

	del := FromDB(repoWrap(&pgconn.PgError{Code: "23503", ConstraintName: "fk_employees_department"}), OpDelete)
	if del.StatusCode != http.StatusConflict || del.Code != CodeForeignKeyViolation {
		t.Fatalf("delete: got %d %s", del.StatusCode, del.Code)
	}
	if del.Fields != nil {
		t.Fatalf("delete: unexpected fields %v", del.Fields)
	}
	assertSafe(t, del)
}

func TestFromDB_OtherConditions(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
		wantFields map[string]string
	}{
		{"gorm not found", gorm.ErrRecordNotFound, 404, CodeNotFound, nil},
		{"wrapped gorm not found", repoWrap(gorm.ErrRecordNotFound), 404, CodeNotFound, nil},
		{"not null", &pgconn.PgError{Code: "23502", ColumnName: "full_name"}, 400, CodeValidationFailed, map[string]string{"full_name": FieldRequired}},
		{"too long", &pgconn.PgError{Code: "22001", Message: "value too long for type character varying(20)"}, 400, CodeValueTooLong, nil},
		{"bad uuid text", &pgconn.PgError{Code: "22P02", Message: "invalid input syntax for type uuid"}, 400, CodeValidationFailed, nil},
		{"translated duplicate", repoWrap(gorm.ErrDuplicatedKey), 409, CodeDuplicateEntry, nil},
		{"translated fk", repoWrap(gorm.ErrForeignKeyViolated), 422, CodeReferenceNotFound, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromDB(tt.err, OpCreate)
			if got == nil {
				t.Fatal("nil")
			}
			if got.StatusCode != tt.wantStatus || got.Code != tt.wantCode {
				t.Fatalf("got %d %s, want %d %s", got.StatusCode, got.Code, tt.wantStatus, tt.wantCode)
			}
			if fmt.Sprint(got.Fields) != fmt.Sprint(tt.wantFields) && !(len(got.Fields) == 0 && len(tt.wantFields) == 0) {
				t.Fatalf("fields = %v, want %v", got.Fields, tt.wantFields)
			}
			assertSafe(t, got)
		})
	}

	if FromDB(stderrors.New("connection refused"), OpRead) != nil {
		t.Fatal("unrecognised error should not be mapped")
	}
	if FromDB(&pgconn.PgError{Code: "57P01"}, OpRead) != nil {
		t.Fatal("unmapped SQLSTATE should not be mapped")
	}
}

func TestNormalize(t *testing.T) {
	t.Run("4xx app error passes through", func(t *testing.T) {
		got := Normalize(ErrEmployeeNotFound, OpRead)
		if got.Code != CodeEmployeeNotFound || got.StatusCode != 404 {
			t.Fatalf("got %d %s", got.StatusCode, got.Code)
		}
	})

	t.Run("legacy VALIDATION_ERROR renamed without mutating sentinel", func(t *testing.T) {
		got := Normalize(Wrap("VALIDATION_ERROR", "Employee code is required", 400, nil), OpCreate)
		if got.Code != CodeValidationFailed || got.Message != "Employee code is required" {
			t.Fatalf("got %s %q", got.Code, got.Message)
		}
		_ = Normalize(ErrValidation, OpCreate)
		if ErrValidation.Code != "VALIDATION_ERROR" {
			t.Fatal("sentinel mutated")
		}
	})

	t.Run("500 wrapping a unique violation is mapped", func(t *testing.T) {
		got := Normalize(serviceWrap(pgErr("23505", "idx_employees_employee_code")), OpCreate)
		if got.StatusCode != 409 || got.Code != CodeEmployeeCodeAlreadyExists {
			t.Fatalf("got %d %s", got.StatusCode, got.Code)
		}
	})

	t.Run("generic not found stays NOT_FOUND", func(t *testing.T) {
		got := Normalize(ErrNotFound, OpRead)
		if got.StatusCode != 404 || got.Code != CodeNotFound {
			t.Fatalf("got %d %s", got.StatusCode, got.Code)
		}
	})

	t.Run("unknown DB failure keeps service message, code INTERNAL_ERROR", func(t *testing.T) {
		cause := stderrors.New(`pq: relation "employees" does not exist`)
		got := Normalize(Wrap("DATABASE_ERROR", "Failed to fetch employees", 500, repoWrap(cause)), OpRead)
		if got.StatusCode != 500 || got.Code != CodeInternal || got.Message != "Failed to fetch employees" {
			t.Fatalf("got %d %s %q", got.StatusCode, got.Code, got.Message)
		}
		if !stderrors.Is(got, cause) {
			t.Fatal("cause not kept for logging")
		}
	})

	t.Run("raw repo error gets generic message", func(t *testing.T) {
		got := Normalize(repoWrap(stderrors.New("dial tcp: connection refused")), OpRead)
		if got.Code != CodeInternal || got.Message != MsgInternal {
			t.Fatalf("got %s %q", got.Code, got.Message)
		}
	})

	t.Run("plain Go error", func(t *testing.T) {
		got := Normalize(stderrors.New("runtime error: index out of range"), OpRead)
		if got.StatusCode != 500 || got.Code != CodeInternal || got.Message != MsgInternal {
			t.Fatalf("got %d %s %q", got.StatusCode, got.Code, got.Message)
		}
	})
}

func TestOpFromHTTPMethod(t *testing.T) {
	cases := map[string]Op{"GET": OpRead, "POST": OpCreate, "PUT": OpUpdate, "PATCH": OpUpdate, "DELETE": OpDelete, "delete": OpDelete}
	for m, want := range cases {
		if got := OpFromHTTPMethod(m); got != want {
			t.Fatalf("%s: got %s want %s", m, got, want)
		}
	}
}
