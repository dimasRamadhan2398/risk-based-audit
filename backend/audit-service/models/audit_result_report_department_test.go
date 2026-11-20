package models

import (
	"testing"

	"audit-service/pkg/testdb"

	"github.com/google/uuid"
)

func TestApplyLetterDepartment(t *testing.T) {
	db := testdb.New(t, &AssignmentLetter{})
	fin := uuid.New()
	linked := AssignmentLetter{ID: uuid.New(), LetterNumber: "ST-1", WorkingUnit: "Finance (FIN)", WorkingUnitDepartmentID: &fin}
	legacy := AssignmentLetter{ID: uuid.New(), LetterNumber: "ST-2", WorkingUnit: "Finance"}
	for _, l := range []*AssignmentLetter{&linked, &legacy} {
		if err := db.Create(l).Error; err != nil {
			t.Fatal(err)
		}
	}

	spoofed := uuid.New()
	r := &AuditResultReport{AssignmentLetterID: "ST-1", Department: "typed", DepartmentID: &spoofed}
	if err := ApplyLetterDepartmentOnCreate(db, r); err != nil {
		t.Fatal(err)
	}
	if r.DepartmentID == nil || *r.DepartmentID != fin || r.Department != "Finance (FIN)" {
		t.Errorf("create linked = %v %q", r.DepartmentID, r.Department)
	}
	r = &AuditResultReport{AssignmentLetterID: "ST-2", Department: "typed", DepartmentID: &spoofed}
	_ = ApplyLetterDepartmentOnCreate(db, r)
	if r.DepartmentID != nil || r.Department != "typed" {
		t.Errorf("create legacy = %v %q", r.DepartmentID, r.Department)
	}

	// Update: same letter, client tries to set department_id -> server value
	existing := &AuditResultReport{AssignmentLetterID: "ST-1"}
	cols := map[string]interface{}{"department_id": spoofed.String()}
	_ = ApplyLetterDepartmentOnUpdate(db, existing, cols)
	if cols["department_id"] != fin.String() || cols["department"] != "Finance (FIN)" {
		t.Errorf("update linked = %v", cols)
	}
	// Moving to a letter without a department clears the id
	cols = map[string]interface{}{"assignment_letter_id": "ST-2"}
	_ = ApplyLetterDepartmentOnUpdate(db, &AuditResultReport{AssignmentLetterID: "ST-1", DepartmentID: &fin}, cols)
	if v, ok := cols["department_id"]; !ok || v != nil {
		t.Errorf("moved = %v", cols)
	}
	// Staying on a legacy letter leaves the column alone
	cols = map[string]interface{}{"department_id": spoofed.String()}
	_ = ApplyLetterDepartmentOnUpdate(db, &AuditResultReport{AssignmentLetterID: "ST-2"}, cols)
	if _, ok := cols["department_id"]; ok {
		t.Errorf("legacy update touched department_id: %v", cols)
	}
}
