package department

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/pkg/validations"
	departmentSvc "master-service/services/department"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// fakeRepo implements the department repository; references always resolve.
type fakeRepo struct {
	findByIDErr error
	createErr   error
}

func (f *fakeRepo) Create(*models.Department) error { return f.createErr }
func (f *fakeRepo) Update(*models.Department) error { return nil }
func (f *fakeRepo) Delete(uuid.UUID) error          { return nil }
func (f *fakeRepo) FindByID(id uuid.UUID) (*models.Department, error) {
	if f.findByIDErr != nil {
		return nil, f.findByIDErr
	}
	return &models.Department{ID: id}, nil
}
func (f *fakeRepo) FindByCode(string) (*models.Department, error) { return nil, apperrors.ErrNotFound }
func (f *fakeRepo) FindByName(string) (*models.Department, error) { return nil, apperrors.ErrNotFound }
func (f *fakeRepo) FindAll() ([]*models.Department, error)        { return nil, nil }
func (f *fakeRepo) FindMany(int, int, string) ([]*models.Department, error) {
	return nil, nil
}
func (f *fakeRepo) Count(string) (int64, error)                { return 0, nil }
func (f *fakeRepo) CompanyExists(uuid.UUID) (bool, error)      { return true, nil }
func (f *fakeRepo) BusinessUnitExists(uuid.UUID) (bool, error) { return true, nil }
func (f *fakeRepo) BusinessUnitBelongsToCompany(uuid.UUID, uuid.UUID) (bool, error) {
	return true, nil
}
func (f *fakeRepo) EmployeeExists(uuid.UUID) (bool, error) { return true, nil }
func (f *fakeRepo) EmployeeBelongsToCompany(uuid.UUID, uuid.UUID) (bool, error) {
	return true, nil
}

func router(repo *fakeRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	ctrl := NewDepartmentController(departmentSvc.NewDepartmentService(repo), validations.New())
	r := gin.New()
	r.GET("/departments/:id", ctrl.FindById)
	r.POST("/departments", ctrl.Create)
	return r
}

type body struct {
	Code  string `json:"code"`
	Error struct {
		Code    string            `json:"code"`
		Message string            `json:"message"`
		Fields  map[string]string `json:"fields"`
	} `json:"error"`
}

func call(t *testing.T, r *gin.Engine, method, path, payload string) (int, body, string) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var b body
	if err := json.Unmarshal(w.Body.Bytes(), &b); err != nil {
		t.Fatalf("invalid json %s", w.Body.String())
	}
	return w.Code, b, w.Body.String()
}

func validPayload(companyID string) string {
	return fmt.Sprintf(`{"department_code":"FIN","department_name":"Finance","pic_id":"%s","level":1,"company_id":"%s","business_unit_id":"%s"}`,
		uuid.NewString(), companyID, uuid.NewString())
}

func TestCreateDepartment_BindingValidationUsesJSONFieldNames(t *testing.T) {
	status, b, raw := call(t, router(&fakeRepo{}), http.MethodPost, "/departments", `{"department_name":"Finance"}`)
	if status != http.StatusBadRequest || b.Code != "VALIDATION_FAILED" || b.Error.Code != "VALIDATION_FAILED" {
		t.Fatalf("got %d %s", status, raw)
	}
	for _, f := range []string{"department_code", "pic_id", "level", "company_id", "business_unit_id"} {
		if b.Error.Fields[f] != "REQUIRED" {
			t.Fatalf("fields[%s] = %q (all %v)", f, b.Error.Fields[f], b.Error.Fields)
		}
	}
	if _, ok := b.Error.Fields["department_name"]; ok {
		t.Fatalf("department_name was provided: %v", b.Error.Fields)
	}
	if strings.Contains(raw, "Key:") || strings.Contains(raw, "CreateDepartmentRequest") {
		t.Fatalf("leaks validator text: %s", raw)
	}
}

func TestCreateDepartment_InvalidUUIDFormat(t *testing.T) {
	status, b, raw := call(t, router(&fakeRepo{}), http.MethodPost, "/departments", validPayload("not-a-uuid"))
	if status != http.StatusBadRequest || b.Error.Code != "VALIDATION_FAILED" || b.Error.Fields["company_id"] != "INVALID_FORMAT" {
		t.Fatalf("got %d %s", status, raw)
	}
}

func TestCreateDepartment_UniqueViolationFromDB(t *testing.T) {
	pg := &pgconn.PgError{Severity: "ERROR", Code: "23505", ConstraintName: "idx_departments_department_name",
		Message: `duplicate key value violates unique constraint "idx_departments_department_name"`}
	repo := &fakeRepo{createErr: fmt.Errorf("%w: %w", apperrors.ErrDatabase, pg)}

	status, b, raw := call(t, router(repo), http.MethodPost, "/departments", validPayload(uuid.NewString()))
	if status != http.StatusConflict || b.Error.Code != "DEPARTMENT_NAME_ALREADY_EXISTS" || b.Error.Fields["department_name"] != "ALREADY_EXISTS" {
		t.Fatalf("got %d %s", status, raw)
	}
	if strings.Contains(raw, "duplicate key") || strings.Contains(raw, "SQLSTATE") {
		t.Fatalf("leak: %s", raw)
	}
}

func TestFindDepartment_NotFound(t *testing.T) {
	status, b, raw := call(t, router(&fakeRepo{findByIDErr: apperrors.ErrNotFound}), http.MethodGet, "/departments/"+uuid.NewString(), "")
	if status != http.StatusNotFound || b.Code != "DEPARTMENT_NOT_FOUND" || b.Error.Code != "DEPARTMENT_NOT_FOUND" {
		t.Fatalf("got %d %s", status, raw)
	}
}
