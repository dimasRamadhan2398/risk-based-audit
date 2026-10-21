package employee

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	"master-service/models"
	"master-service/pkg/base"
)

// Count and FindPage on fakeRepo filter by search over code, name and email,
// sort by code, then slice. That is enough to check the service's page math;
// the SQL itself is covered in repositories/employee.

var errTestList = errors.New("list failed")

func (f *fakeRepo) matching(search string) []*models.Employee {
	q := strings.ToLower(search)
	var out []*models.Employee
	for _, e := range f.employees {
		if q == "" ||
			strings.Contains(strings.ToLower(e.EmployeeCode), q) ||
			strings.Contains(strings.ToLower(e.FullName), q) ||
			strings.Contains(strings.ToLower(e.Email), q) {
			out = append(out, e)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].EmployeeCode < out[j].EmployeeCode })
	return out
}

func (f *fakeRepo) Count(search string) (int64, error) {
	if f.listErr != nil {
		return 0, f.listErr
	}
	return int64(len(f.matching(search))), nil
}

func (f *fakeRepo) FindPage(offset, limit int, search string) ([]*models.Employee, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	all := f.matching(search)
	if offset >= len(all) {
		return []*models.Employee{}, nil
	}
	end := min(offset+limit, len(all))
	return all[offset:end], nil
}

// seededEmployees returns n employees: EMP-0001..EMP-000n, named "Person i",
// email personi@example.com.
func seededEmployees(n int) []*models.Employee {
	out := make([]*models.Employee, 0, n)
	for i := 1; i <= n; i++ {
		out = append(out, &models.Employee{
			EmployeeCode: fmt.Sprintf("EMP-%04d", i),
			FullName:     fmt.Sprintf("Person %d", i),
			Email:        fmt.Sprintf("person%d@example.com", i),
		})
	}
	return out
}

func TestFindPage_PageBoundariesAndTotals(t *testing.T) {
	repo := newFakeRepo()
	repo.employees = seededEmployees(25)
	svc := &EmployeeService{employeeRepo: repo}

	cases := []struct {
		name      string
		page      int
		wantRows  int
		wantFirst string
		wantLast  string
	}{
		{"first page is full", 1, 10, "EMP-0001", "EMP-0010"},
		{"middle page", 2, 10, "EMP-0011", "EMP-0020"},
		{"last page is partial", 3, 5, "EMP-0021", "EMP-0025"},
		{"page past the end is empty", 4, 0, "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := svc.FindPage(base.NewBaseService(), tc.page, 10, "")
			if err != nil {
				t.Fatalf("FindPage returned error: %v", err)
			}
			if len(got.Employees) != tc.wantRows {
				t.Fatalf("rows = %d, want %d", len(got.Employees), tc.wantRows)
			}
			if tc.wantRows > 0 {
				first, last := got.Employees[0].EmployeeCode, got.Employees[tc.wantRows-1].EmployeeCode
				if first != tc.wantFirst || last != tc.wantLast {
					t.Fatalf("range = %s..%s, want %s..%s", first, last, tc.wantFirst, tc.wantLast)
				}
			}
			if got.Total != 25 || got.TotalPages != 3 {
				t.Fatalf("total=%d total_pages=%d, want 25 and 3", got.Total, got.TotalPages)
			}
			if got.Page != tc.page || got.PageSize != 10 {
				t.Fatalf("page=%d page_size=%d, want %d and 10", got.Page, got.PageSize, tc.page)
			}
		})
	}
}

func TestFindPage_DefaultsWhenParamsMissingOrInvalid(t *testing.T) {
	repo := newFakeRepo()
	repo.employees = seededEmployees(15)
	svc := &EmployeeService{employeeRepo: repo}

	for _, in := range [][2]int{{0, 0}, {-3, -1}} {
		got, err := svc.FindPage(base.NewBaseService(), in[0], in[1], "")
		if err != nil {
			t.Fatalf("FindPage(%v) returned error: %v", in, err)
		}
		if got.Page != 1 || got.PageSize != 10 {
			t.Fatalf("FindPage(%v): page=%d page_size=%d, want 1 and 10", in, got.Page, got.PageSize)
		}
		if len(got.Employees) != 10 || got.TotalPages != 2 {
			t.Fatalf("FindPage(%v): rows=%d total_pages=%d, want 10 and 2", in, len(got.Employees), got.TotalPages)
		}
	}
}

func TestFindPage_PageSizeIsCappedAt100(t *testing.T) {
	repo := newFakeRepo()
	repo.employees = seededEmployees(150)
	svc := &EmployeeService{employeeRepo: repo}

	got, err := svc.FindPage(base.NewBaseService(), 1, 1000, "")
	if err != nil {
		t.Fatalf("FindPage returned error: %v", err)
	}
	if got.PageSize != 100 {
		t.Fatalf("page_size = %d, want cap 100", got.PageSize)
	}
	if len(got.Employees) != 100 || got.TotalPages != 2 {
		t.Fatalf("rows=%d total_pages=%d, want 100 and 2", len(got.Employees), got.TotalPages)
	}
}

func TestFindPage_SearchFiltersTotalAndPages(t *testing.T) {
	repo := newFakeRepo()
	repo.employees = seededEmployees(25)
	svc := &EmployeeService{employeeRepo: repo}

	// "  PERSON 1  " is trimmed and lowered: it matches Person 1 and Person 10..19.
	got, err := svc.FindPage(base.NewBaseService(), 1, 5, "  PERSON 1  ")
	if err != nil {
		t.Fatalf("FindPage returned error: %v", err)
	}
	if got.Total != 11 || got.TotalPages != 3 || len(got.Employees) != 5 {
		t.Fatalf("total=%d total_pages=%d rows=%d, want 11, 3, 5", got.Total, got.TotalPages, len(got.Employees))
	}

	// EMP-001 matches EMP-0010..EMP-0019: 10 rows, page 2 of size 5 starts at EMP-0015.
	got, err = svc.FindPage(base.NewBaseService(), 2, 5, "EMP-001")
	if err != nil {
		t.Fatalf("FindPage returned error: %v", err)
	}
	if got.Total != 10 || got.TotalPages != 2 || len(got.Employees) != 5 {
		t.Fatalf("total=%d total_pages=%d rows=%d, want 10, 2, 5", got.Total, got.TotalPages, len(got.Employees))
	}
	if got.Employees[0].EmployeeCode != "EMP-0015" {
		t.Fatalf("page 2 first = %s, want EMP-0015", got.Employees[0].EmployeeCode)
	}

	got, err = svc.FindPage(base.NewBaseService(), 1, 10, "nothing-matches")
	if err != nil {
		t.Fatalf("FindPage returned error: %v", err)
	}
	if got.Total != 0 || got.TotalPages != 0 || len(got.Employees) != 0 {
		t.Fatalf("no match: total=%d total_pages=%d rows=%d, want all zero", got.Total, got.TotalPages, len(got.Employees))
	}
}

func TestFindPage_RepoErrorIsReported(t *testing.T) {
	repo := newFakeRepo()
	repo.employees = seededEmployees(3)
	repo.listErr = errTestList
	svc := &EmployeeService{employeeRepo: repo}

	if _, err := svc.FindPage(base.NewBaseService(), 1, 10, ""); err == nil {
		t.Fatal("FindPage returned nil error when the repository failed")
	}
}

func TestFindPage_NilRepoIsInternal(t *testing.T) {
	svc := &EmployeeService{}
	if _, err := svc.FindPage(base.NewBaseService(), 1, 10, ""); err == nil {
		t.Fatal("FindPage with nil repo returned nil error")
	}
}
