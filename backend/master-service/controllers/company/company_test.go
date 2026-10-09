package company

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"master-service/models"
	apperrors "master-service/pkg/errors"
	companySvc "master-service/services/company"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type fakeRepo struct {
	findByIDErr error
	createErr   error
}

func (f *fakeRepo) Create(*models.Company) error { return f.createErr }
func (f *fakeRepo) Update(*models.Company) error { return nil }
func (f *fakeRepo) Delete(uuid.UUID) error       { return nil }
func (f *fakeRepo) FindByID(id uuid.UUID) (*models.Company, error) {
	if f.findByIDErr != nil {
		return nil, f.findByIDErr
	}
	return &models.Company{ID: id}, nil
}
func (f *fakeRepo) FindByCode(string) (*models.Company, error) { return nil, apperrors.ErrNotFound }
func (f *fakeRepo) FindAll() ([]*models.Company, error)        { return nil, nil }

func router(repo *fakeRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	ctrl := NewCompanyController(companySvc.NewCompanyService(repo), nil)
	r := gin.New()
	r.GET("/companies/:id", ctrl.FindById)
	r.POST("/companies", ctrl.Create)
	return r
}

type body struct {
	Code  string `json:"code"`
	Error struct {
		Code   string            `json:"code"`
		Fields map[string]string `json:"fields"`
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

// tax_id has a unique index but no service pre-check, so the DB error used to
// reach the client verbatim.
func TestCreateCompany_TaxIDUniqueViolationFromDB(t *testing.T) {
	pg := &pgconn.PgError{Severity: "ERROR", Code: "23505", ConstraintName: "idx_companies_tax_id",
		Message: `duplicate key value violates unique constraint "idx_companies_tax_id"`}
	repo := &fakeRepo{createErr: fmt.Errorf("%w: %w", apperrors.ErrDatabase, pg)}

	status, b, raw := call(t, router(repo), http.MethodPost, "/companies", `{"code":"HLD-001","name":"Holding","company_type":"HOLDING","tax_id":""}`)
	if status != http.StatusConflict || b.Code != "COMPANY_TAX_ID_ALREADY_EXISTS" || b.Error.Fields["tax_id"] != "ALREADY_EXISTS" {
		t.Fatalf("got %d %s", status, raw)
	}
	if strings.Contains(raw, "duplicate key") || strings.Contains(raw, "SQLSTATE") {
		t.Fatalf("leak: %s", raw)
	}
}

func TestCreateCompany_MissingCode(t *testing.T) {
	status, b, raw := call(t, router(&fakeRepo{}), http.MethodPost, "/companies", `{"name":"Holding"}`)
	if status != http.StatusBadRequest || b.Error.Code != "VALIDATION_FAILED" || b.Error.Fields["code"] != "REQUIRED" {
		t.Fatalf("got %d %s", status, raw)
	}
}

func TestFindCompany_NotFound(t *testing.T) {
	status, b, raw := call(t, router(&fakeRepo{findByIDErr: apperrors.ErrNotFound}), http.MethodGet, "/companies/"+uuid.NewString(), "")
	if status != http.StatusNotFound || b.Error.Code != "COMPANY_NOT_FOUND" {
		t.Fatalf("got %d %s", status, raw)
	}
}
