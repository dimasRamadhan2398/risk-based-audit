package employee

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	employeeRepo "master-service/repositories/employee"
	employeeSvc "master-service/services/employee"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// newDB opens a GORM handle with the postgres dialector but no server. Query,
// create, update and delete callbacks are replaced so each test decides what
// the "database" returns, exactly as the pgx driver would surface it.
func newDB(t *testing.T, queryErr, writeErr error) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=127.0.0.1 port=1 user=x dbname=x sslmode=disable"}),
		&gorm.Config{DisableAutomaticPing: true, SkipDefaultTransaction: true})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	fail := func(e error) func(*gorm.DB) {
		return func(db *gorm.DB) {
			if e != nil {
				_ = db.AddError(e)
			}
		}
	}
	if err := db.Callback().Query().Replace("gorm:query", fail(queryErr)); err != nil {
		t.Fatal(err)
	}
	if err := db.Callback().Create().Replace("gorm:create", fail(writeErr)); err != nil {
		t.Fatal(err)
	}
	if err := db.Callback().Update().Replace("gorm:update", fail(writeErr)); err != nil {
		t.Fatal(err)
	}
	if err := db.Callback().Delete().Replace("gorm:delete", fail(writeErr)); err != nil {
		t.Fatal(err)
	}
	return db
}

func newRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	ctrl := NewEmployeeController(employeeSvc.NewEmployeeService(employeeRepo.NewEmployeeRepository(db)), nil)
	r := gin.New()
	r.GET("/employees/:id", ctrl.FindById)
	r.POST("/employees", ctrl.Create)
	r.DELETE("/employees/:id", ctrl.Delete)
	return r
}

type errBody struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Error   struct {
		Code    string            `json:"code"`
		Message string            `json:"message"`
		Details string            `json:"details"`
		Fields  map[string]string `json:"fields"`
	} `json:"error"`
}

func do(t *testing.T, r *gin.Engine, method, path, body string) (int, errBody, string) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var out errBody
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON %q: %v", w.Body.String(), err)
	}
	return w.Code, out, w.Body.String()
}

func assertNoLeak(t *testing.T, raw string) {
	t.Helper()
	for _, bad := range []string{"SQLSTATE", "duplicate key", "violates", "ERROR:", "record not found", "Database error", "Key:", "json:", "uuid"} {
		if strings.Contains(raw, bad) {
			t.Fatalf("response leaks %q: %s", bad, raw)
		}
	}
}

const validEmployee = `{"employee_code":"EMP-0001","full_name":"Budi","email":"Budi@Example.com",
	"company_id":"11111111-1111-1111-1111-111111111111","department_id":"22222222-2222-2222-2222-222222222222",
	"job_role_id":"33333333-3333-3333-3333-333333333333","level_grade":1,"join_date":"2024-01-01T00:00:00Z"}`

func TestCreateEmployee_UniqueViolationFromDB(t *testing.T) {
	// Pre-checks pass (no row found, e.g. the clash is with a soft-deleted
	// row), then the INSERT hits the unique index.
	pg := &pgconn.PgError{Severity: "ERROR", Code: "23505", ConstraintName: "idx_employees_employee_code",
		Message: `duplicate key value violates unique constraint "idx_employees_employee_code"`}
	r := newRouter(newDB(t, gorm.ErrRecordNotFound, pg))

	status, body, raw := do(t, r, http.MethodPost, "/employees", validEmployee)
	if status != http.StatusConflict {
		t.Fatalf("status = %d, body %s", status, raw)
	}
	if body.Success || body.Code != "EMPLOYEE_CODE_ALREADY_EXISTS" || body.Error.Code != "EMPLOYEE_CODE_ALREADY_EXISTS" {
		t.Fatalf("codes = %q / %q", body.Code, body.Error.Code)
	}
	if body.Error.Fields["employee_code"] != "ALREADY_EXISTS" {
		t.Fatalf("fields = %v", body.Error.Fields)
	}
	if body.Error.Message == "" || body.Error.Details != "" {
		t.Fatalf("message %q details %q", body.Error.Message, body.Error.Details)
	}
	assertNoLeak(t, raw)
}

func TestCreateEmployee_ForeignKeyViolationFromDB(t *testing.T) {
	pg := &pgconn.PgError{Severity: "ERROR", Code: "23503", ConstraintName: "fk_employees_department",
		Message: `insert or update on table "employees" violates foreign key constraint "fk_employees_department"`}
	r := newRouter(newDB(t, gorm.ErrRecordNotFound, pg))

	status, body, raw := do(t, r, http.MethodPost, "/employees", validEmployee)
	if status != http.StatusUnprocessableEntity || body.Error.Code != "REFERENCE_NOT_FOUND" {
		t.Fatalf("got %d %s", status, raw)
	}
	if body.Error.Fields["department_id"] != "NOT_FOUND" {
		t.Fatalf("fields = %v", body.Error.Fields)
	}
	assertNoLeak(t, raw)
}

func TestDeleteEmployee_ForeignKeyViolationFromDB(t *testing.T) {
	pg := &pgconn.PgError{Severity: "ERROR", Code: "23503", ConstraintName: "fk_departments_pic",
		Message: `update or delete on table "employees" violates foreign key constraint`}
	r := newRouter(newDB(t, nil, pg))

	status, body, raw := do(t, r, http.MethodDelete, "/employees/44444444-4444-4444-4444-444444444444", "")
	if status != http.StatusConflict || body.Error.Code != "FOREIGN_KEY_VIOLATION" {
		t.Fatalf("got %d %s", status, raw)
	}
	assertNoLeak(t, raw)
}

func TestFindEmployee_NotFound(t *testing.T) {
	r := newRouter(newDB(t, gorm.ErrRecordNotFound, nil))

	status, body, raw := do(t, r, http.MethodGet, "/employees/44444444-4444-4444-4444-444444444444", "")
	if status != http.StatusNotFound || body.Code != "EMPLOYEE_NOT_FOUND" || body.Error.Code != "EMPLOYEE_NOT_FOUND" {
		t.Fatalf("got %d %s", status, raw)
	}
	assertNoLeak(t, raw)
}

func TestFindEmployee_InvalidID(t *testing.T) {
	r := newRouter(newDB(t, nil, nil))

	status, body, raw := do(t, r, http.MethodGet, "/employees/not-a-uuid", "")
	if status != http.StatusBadRequest || body.Error.Code != "INVALID_ID" {
		t.Fatalf("got %d %s", status, raw)
	}
	assertNoLeak(t, raw)
}

func TestFindEmployee_UnknownDBFailureIsInternal(t *testing.T) {
	r := newRouter(newDB(t, &pgconn.PgError{Severity: "ERROR", Code: "42P01", Message: `relation "employees" does not exist`}, nil))

	status, body, raw := do(t, r, http.MethodGet, "/employees/44444444-4444-4444-4444-444444444444", "")
	if status != http.StatusInternalServerError || body.Error.Code != "INTERNAL_ERROR" {
		t.Fatalf("got %d %s", status, raw)
	}
	if strings.Contains(raw, "relation") {
		t.Fatalf("leak: %s", raw)
	}
	assertNoLeak(t, raw)
}

func TestCreateEmployee_Validation(t *testing.T) {
	r := newRouter(newDB(t, gorm.ErrRecordNotFound, nil))

	t.Run("missing required fields", func(t *testing.T) {
		status, body, raw := do(t, r, http.MethodPost, "/employees", `{"full_name":"Budi"}`)
		if status != http.StatusBadRequest || body.Error.Code != "VALIDATION_FAILED" {
			t.Fatalf("got %d %s", status, raw)
		}
		if body.Error.Fields["employee_code"] != "REQUIRED" || body.Error.Fields["email"] != "REQUIRED" {
			t.Fatalf("fields = %v", body.Error.Fields)
		}
	})

	t.Run("empty uuid string (TextUnmarshaler error)", func(t *testing.T) {
		status, body, raw := do(t, r, http.MethodPost, "/employees", `{"employee_code":"E1","email":"a@b.co","company_id":""}`)
		if status != http.StatusBadRequest || body.Error.Code != "VALIDATION_FAILED" {
			t.Fatalf("got %d %s", status, raw)
		}
		assertNoLeak(t, raw)
	})

	t.Run("wrong JSON type", func(t *testing.T) {
		status, body, raw := do(t, r, http.MethodPost, "/employees", `{"employee_code":"E1","email":"a@b.co","level_grade":"senior"}`)
		if status != http.StatusBadRequest || body.Error.Code != "VALIDATION_FAILED" || body.Error.Fields["level_grade"] != "INVALID_TYPE" {
			t.Fatalf("got %d %s", status, raw)
		}
		assertNoLeak(t, raw)
	})

	t.Run("malformed body", func(t *testing.T) {
		status, body, raw := do(t, r, http.MethodPost, "/employees", `{"employee_code":`)
		if status != http.StatusBadRequest || body.Error.Code != "INVALID_REQUEST_BODY" {
			t.Fatalf("got %d %s", status, raw)
		}
		assertNoLeak(t, raw)
	})
}
