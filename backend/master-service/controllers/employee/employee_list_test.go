package employee

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"master-service/models"
	apperrors "master-service/pkg/errors"
	employeeSvc "master-service/services/employee"

	"github.com/gin-gonic/gin"
)

// Count and FindPage on memRepo filter and sort like the SQL does, so the
// controller tests see the same rows the service would get from Postgres.
func (m *memRepo) matching(search string) []*models.Employee {
	q := strings.ToLower(search)
	var out []*models.Employee
	for _, e := range m.employees {
		if q == "" || strings.Contains(strings.ToLower(e.EmployeeCode), q) ||
			strings.Contains(strings.ToLower(e.FullName), q) || strings.Contains(strings.ToLower(e.Email), q) {
			out = append(out, e)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].EmployeeCode < out[j].EmployeeCode })
	return out
}

func (m *memRepo) Count(search string) (int64, error) {
	return int64(len(m.matching(search))), nil
}

func (m *memRepo) FindPage(offset, limit int, search string) ([]*models.Employee, error) {
	all := m.matching(search)
	if offset >= len(all) {
		return []*models.Employee{}, nil
	}
	return all[offset:min(offset+limit, len(all))], nil
}

type listEnvelope struct {
	Data struct {
		Employees  []models.Employee `json:"employees"`
		Pagination struct {
			Page       int   `json:"page"`
			PageSize   int   `json:"page_size"`
			Total      int64 `json:"total"`
			TotalPages int   `json:"total_pages"`
		} `json:"pagination"`
	} `json:"data"`
}

func listRouter(repo *memRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	ctrl := NewEmployeeController(employeeSvc.NewEmployeeService(repo), nil)
	r := gin.New()
	r.GET("/employees", ctrl.FindAll)
	return r
}

func getList(t *testing.T, r *gin.Engine, query string) listEnvelope {
	t.Helper()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/employees?"+query, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("GET /employees?%s status = %d, body = %s", query, w.Code, w.Body.String())
	}
	var env listEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode: %v; body = %s", err, w.Body.String())
	}
	return env
}

func seedList(n int) *memRepo {
	repo := &memRepo{}
	for i := 1; i <= n; i++ {
		repo.employees = append(repo.employees, &models.Employee{
			EmployeeCode: fmt.Sprintf("EMP-%04d", i),
			FullName:     fmt.Sprintf("Person %d", i),
			Email:        fmt.Sprintf("person%d@example.com", i),
		})
	}
	return repo
}

func TestListEmployees_EnvelopeMatchesDepartmentShape(t *testing.T) {
	env := getList(t, listRouter(seedList(25)), "page=2&page_size=10")

	if len(env.Data.Employees) != 10 || env.Data.Employees[0].EmployeeCode != "EMP-0011" {
		t.Fatalf("page 2 rows = %d first = %s", len(env.Data.Employees), env.Data.Employees[0].EmployeeCode)
	}
	p := env.Data.Pagination
	if p.Page != 2 || p.PageSize != 10 || p.Total != 25 || p.TotalPages != 3 {
		t.Fatalf("pagination = %+v, want page 2, size 10, total 25, pages 3", p)
	}
}

func TestListEmployees_DefaultsAndInvalidParams(t *testing.T) {
	r := listRouter(seedList(15))

	for _, q := range []string{"", "page=abc&page_size=xyz", "page=0&page_size=-5"} {
		env := getList(t, r, q)
		if env.Data.Pagination.Page != 1 || env.Data.Pagination.PageSize != 10 || len(env.Data.Employees) != 10 {
			t.Fatalf("query %q: pagination = %+v rows = %d, want page 1 size 10 rows 10",
				q, env.Data.Pagination, len(env.Data.Employees))
		}
	}
}

func TestListEmployees_PageSizeCapAndLastPage(t *testing.T) {
	r := listRouter(seedList(150))

	env := getList(t, r, "page_size=1000")
	if env.Data.Pagination.PageSize != 100 || len(env.Data.Employees) != 100 || env.Data.Pagination.TotalPages != 2 {
		t.Fatalf("cap: pagination = %+v rows = %d", env.Data.Pagination, len(env.Data.Employees))
	}

	env = getList(t, r, "page=2&page_size=100")
	if len(env.Data.Employees) != 50 || env.Data.Employees[49].EmployeeCode != "EMP-0150" {
		t.Fatalf("last page rows = %d", len(env.Data.Employees))
	}
}

func TestListEmployees_SearchFiltersBeforePaging(t *testing.T) {
	env := getList(t, listRouter(seedList(25)), "search=EMP-001&page_size=5&page=2")

	if env.Data.Pagination.Total != 10 || env.Data.Pagination.TotalPages != 2 {
		t.Fatalf("pagination = %+v, want total 10 and pages 2", env.Data.Pagination)
	}
	if len(env.Data.Employees) != 5 || env.Data.Employees[0].EmployeeCode != "EMP-0015" {
		t.Fatalf("search page 2 rows = %d", len(env.Data.Employees))
	}
}

func TestListEmployees_RepoErrorIsNotOK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewEmployeeController(employeeSvc.NewEmployeeService(&failingListRepo{memRepo: &memRepo{}}), nil)
	r := gin.New()
	r.GET("/employees", ctrl.FindAll)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/employees", nil))
	if w.Code < 500 {
		t.Fatalf("status = %d, want a 5xx for a database failure (apperrors.DB)", w.Code)
	}
	_ = apperrors.ErrNotFound
}

// failingListRepo makes Count fail, to check that a DB error surfaces as a 5xx.
type failingListRepo struct{ *memRepo }

func (failingListRepo) Count(string) (int64, error) { return 0, fmt.Errorf("db down") }
