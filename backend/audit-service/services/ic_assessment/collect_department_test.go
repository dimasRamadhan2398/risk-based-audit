package ic_assessment

import (
	"testing"

	"audit-service/models"

	"github.com/google/uuid"
)

func TestLHADepartmentMatchesByID(t *testing.T) {
	fin, it := uuid.New(), uuid.New()
	sc := Scope{DepartmentID: fin.String(), DepartmentName: "Finance", DepartmentCode: "FIN"}
	letterFin := &models.AssignmentLetter{WorkingUnit: "Finance (FIN)", WorkingUnitDepartmentID: &fin}
	letterIT := &models.AssignmentLetter{WorkingUnit: "Finance (FIN)", WorkingUnitDepartmentID: &it}
	legacyLetter := &models.AssignmentLetter{WorkingUnit: "Finance (FIN)"}

	cases := []struct {
		name   string
		r      models.AuditResultReport
		letter *models.AssignmentLetter
		want   string
	}{
		{"report id wins over text", models.AuditResultReport{DepartmentID: &fin, Department: "Renamed"}, nil, deptMatchID},
		{"report id of another department", models.AuditResultReport{DepartmentID: &it, Department: "Finance"}, letterFin, deptMatchNone},
		{"letter id", models.AuditResultReport{}, letterFin, deptMatchID},
		{"letter id beats matching text", models.AuditResultReport{Department: "Finance"}, letterIT, deptMatchNone},
		{"legacy report text", models.AuditResultReport{Department: " finance "}, nil, deptMatchLegacy},
		{"legacy letter label", models.AuditResultReport{}, legacyLetter, deptMatchLegacy},
		{"legacy report text decides first", models.AuditResultReport{Department: "IT"}, legacyLetter, deptMatchNone},
		{"nothing", models.AuditResultReport{}, nil, deptMatchNone},
	}
	for _, c := range cases {
		if got := lhaDepartment(c.r, c.letter, sc); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}
