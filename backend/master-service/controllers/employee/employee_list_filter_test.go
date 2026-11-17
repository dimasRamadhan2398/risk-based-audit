package employee

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"master-service/models"

	"github.com/google/uuid"
)

func filterRepo(deptA, deptB uuid.UUID) *memRepo {
	return &memRepo{employees: []*models.Employee{
		{EmployeeCode: "EMP-0001", DepartmentID: deptA, IsActive: true},
		{EmployeeCode: "EMP-0002", DepartmentID: deptA, IsActive: false},
		{EmployeeCode: "EMP-0003", DepartmentID: deptB, IsActive: true},
	}}
}

func listCodes(env listEnvelope) string {
	codes := make([]string, 0, len(env.Data.Employees))
	for _, e := range env.Data.Employees {
		codes = append(codes, e.EmployeeCode)
	}
	return strings.Join(codes, ",")
}

func TestListEmployees_DepartmentAndActiveFilters(t *testing.T) {
	deptA, deptB := uuid.New(), uuid.New()
	r := listRouter(filterRepo(deptA, deptB))

	cases := []struct{ query, want string }{
		// No filter params: unchanged behaviour, every row.
		{"", "EMP-0001,EMP-0002,EMP-0003"},
		{"department_id=&is_active=", "EMP-0001,EMP-0002,EMP-0003"},
		{"department_id=" + deptA.String(), "EMP-0001,EMP-0002"},
		{"department_id=" + strings.ToUpper(deptA.String()), "EMP-0001,EMP-0002"},
		{"is_active=true", "EMP-0001,EMP-0003"},
		{"is_active=false", "EMP-0002"},
		{"is_active=1", "EMP-0001,EMP-0003"},
		{"department_id=" + deptA.String() + "&is_active=true", "EMP-0001"},
		{"department_id=" + uuid.NewString(), ""},
	}
	for _, tc := range cases {
		env := getList(t, r, tc.query)
		if got := listCodes(env); got != tc.want {
			t.Errorf("?%s rows = %q, want %q", tc.query, got, tc.want)
		}
		if want := int64(len(env.Data.Employees)); env.Data.Pagination.Total != want {
			t.Errorf("?%s total = %d, want %d", tc.query, env.Data.Pagination.Total, want)
		}
	}
}

func TestListEmployees_InvalidFiltersAre400(t *testing.T) {
	r := listRouter(filterRepo(uuid.New(), uuid.New()))
	for _, q := range []string{
		"department_id=not-a-uuid",
		"department_id=1%27%20OR%201%3D1",
		"is_active=maybe",
		"department_id=x&is_active=y",
	} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/employees?"+q, nil))
		if w.Code != http.StatusBadRequest {
			t.Errorf("?%s status = %d, want 400; body = %s", q, w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "VALIDATION_FAILED") {
			t.Errorf("?%s body = %s, want VALIDATION_FAILED", q, w.Body.String())
		}
	}
}
