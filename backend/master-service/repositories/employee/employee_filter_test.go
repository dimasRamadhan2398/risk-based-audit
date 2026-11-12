package employee

import (
	"reflect"
	"strings"
	"testing"

	"master-service/models"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func insertFiltered(t *testing.T, db *gorm.DB, code string, dept uuid.UUID, active bool) {
	t.Helper()
	e := insertEmployee(t, db, code, "Name "+code, strings.ToLower(code)+"@example.com")
	// GORM skips false on Create (default:true), so set both columns explicitly.
	if err := db.Model(&models.Employee{}).Where("id = ?", e.ID).
		Updates(map[string]any{"department_id": dept, "is_active": active}).Error; err != nil {
		t.Fatalf("update %s: %v", code, err)
	}
}

func TestFindPage_DepartmentAndActiveFilters(t *testing.T) {
	repo, db := newTestRepo(t)
	deptA, deptB := uuid.New(), uuid.New()
	insertFiltered(t, db, "EMP-0001", deptA, true)
	insertFiltered(t, db, "EMP-0002", deptA, false)
	insertFiltered(t, db, "EMP-0003", deptB, true)
	insertFiltered(t, db, "EMP-0004", deptA, true)
	gone := insertEmployee(t, db, "EMP-0005", "Gone", "gone@example.com")
	if err := db.Model(&models.Employee{}).Where("id = ?", gone.ID).Update("department_id", deptA).Error; err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(gone.ID); err != nil {
		t.Fatal(err)
	}

	yes, no := true, false
	cases := []struct {
		name   string
		filter EmployeeListFilter
		want   []string
	}{
		{"no filter", EmployeeListFilter{}, []string{"EMP-0001", "EMP-0002", "EMP-0003", "EMP-0004"}},
		{"department", EmployeeListFilter{DepartmentID: &deptA}, []string{"EMP-0001", "EMP-0002", "EMP-0004"}},
		{"active", EmployeeListFilter{IsActive: &yes}, []string{"EMP-0001", "EMP-0003", "EMP-0004"}},
		{"inactive", EmployeeListFilter{IsActive: &no}, []string{"EMP-0002"}},
		{"department and active", EmployeeListFilter{DepartmentID: &deptA, IsActive: &yes}, []string{"EMP-0001", "EMP-0004"}},
		{"department, active and search", EmployeeListFilter{Search: "0004", DepartmentID: &deptA, IsActive: &yes}, []string{"EMP-0004"}},
		{"unknown department", EmployeeListFilter{DepartmentID: ptr(uuid.New())}, []string{}},
	}
	for _, tc := range cases {
		rows, err := repo.FindPage(0, 10, tc.filter)
		if err != nil {
			t.Fatalf("%s: FindPage: %v", tc.name, err)
		}
		if got := codesOf(rows); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: FindPage = %v, want %v", tc.name, got, tc.want)
		}
		total, err := repo.Count(tc.filter)
		if err != nil || total != int64(len(tc.want)) {
			t.Errorf("%s: Count = %d, %v; want %d", tc.name, total, err, len(tc.want))
		}
	}
}

func ptr[T any](v T) *T { return &v }

// The filter columns are fixed identifiers and the values are bound, so the
// SQL shape must not change with the input.
func TestEmployeeListFilter_SQLShapeIsParameterised(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=127.0.0.1 dbname=none"}), &gorm.Config{
		DryRun: true, DisableAutomaticPing: true, Logger: logger.Discard,
	})
	if err != nil {
		t.Fatal(err)
	}
	dept := uuid.New()
	active := false
	stmt := applyEmployeeListFilter(db.Model(&models.Employee{}), EmployeeListFilter{DepartmentID: &dept, IsActive: &active}).
		Find(&[]models.Employee{}).Statement
	sql := stmt.SQL.String()
	if !strings.Contains(sql, `"department_id" = $1`) || !strings.Contains(sql, `"is_active" = $2`) {
		t.Fatalf("SQL = %s, want quoted columns with bound values", sql)
	}
	if len(stmt.Vars) != 2 || stmt.Vars[0] != dept || stmt.Vars[1] != false {
		t.Fatalf("vars = %v", stmt.Vars)
	}
}
