package department_link

import (
	"testing"

	"audit-service/models"
	"audit-service/pkg/masterclient"

	"github.com/google/uuid"
)

func TestPlan(t *testing.T) {
	fin, it, dup1, dup2 := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	depts := []masterclient.Department{
		{ID: fin.String(), DepartmentCode: "FIN", DepartmentName: "Finance & Accounting"},
		{ID: it.String(), DepartmentCode: "IT", DepartmentName: "Information Technology"},
		{ID: dup1.String(), DepartmentCode: "OPS1", DepartmentName: "Operations"},
		{ID: dup2.String(), DepartmentCode: "OPS2", DepartmentName: "Operations"},
	}
	already := uuid.New()
	letters := []models.AssignmentLetter{
		{ID: uuid.New(), LetterNumber: "ST-1", WorkingUnit: " finance  &  accounting (fin) "},
		{ID: uuid.New(), LetterNumber: "ST-2", WorkingUnit: "IT"},
		{ID: uuid.New(), LetterNumber: "ST-3", WorkingUnit: "Operations"},
		{ID: uuid.New(), LetterNumber: "ST-4", WorkingUnit: "Production"},
		{ID: uuid.New(), LetterNumber: "ST-5", WorkingUnit: "whatever", WorkingUnitDepartmentID: &already},
	}
	reports := []models.AuditResultReport{
		{ID: uuid.New(), ReportNumber: "LHA-1", AssignmentLetterID: "ST-1"},
		{ID: uuid.New(), ReportNumber: "LHA-5", AssignmentLetterID: letters[4].ID.String()},
		{ID: uuid.New(), ReportNumber: "LHA-4", AssignmentLetterID: "ST-4", Department: "Information Technology"},
		{ID: uuid.New(), ReportNumber: "LHA-X", AssignmentLetterID: "ST-4", Department: "Production"},
		{ID: uuid.New(), ReportNumber: "LHA-DONE", AssignmentLetterID: "ST-1", DepartmentID: &it},
	}

	links, misses := Plan(letters, reports, depts)
	got := map[string]uuid.UUID{}
	for _, l := range links {
		got[l.Ref] = l.DepartmentID
	}
	want := map[string]uuid.UUID{"ST-1": fin, "ST-2": it, "LHA-1": fin, "LHA-5": already, "LHA-4": it}
	if len(got) != len(want) {
		t.Fatalf("links = %v", links)
	}
	for ref, id := range want {
		if got[ref] != id {
			t.Errorf("%s -> %v, want %v", ref, got[ref], id)
		}
	}
	missed := map[string]bool{}
	for _, m := range misses {
		missed[m.Ref] = true
	}
	for _, ref := range []string{"ST-3", "ST-4", "LHA-X"} {
		if !missed[ref] {
			t.Errorf("%s should be reported as a miss: %v", ref, misses)
		}
	}
}
