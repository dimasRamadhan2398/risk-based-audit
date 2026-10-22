package annual_plan_sync

import (
	"context"
	"errors"
	"testing"

	"audit-service/models"

	"gorm.io/gorm"
)

func TestBackfillFillsOnlyEmptyFieldsAndIsIdempotent(t *testing.T) {
	db := testDB(t)
	seedRefs(t, db,
		dept("dep-1", "Finance"),
		entity("uni-1", "Treasury (current)"),
		yearScore("uni-1", 2026, "High"),
	)
	plan := seedPlan(t, db, 2026,
		// Empty snapshots: all three are filled.
		models.AnnualAuditActivity{DepartmentID: "dep-1", AuditUniverseID: "uni-1"},
		// Non-empty snapshots: kept even though the lookup has other values.
		models.AnnualAuditActivity{DepartmentID: "dep-1", Department: "Keep me", AuditUniverseID: "uni-1", RiskName: "Keep me too", RiskLevel: "Low"},
	)

	res, err := BackfillSnapshots(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if res.PlansChanged != 1 || res.FieldsFilled != 3 {
		t.Fatalf("result = %+v, want 1 plan and 3 fields", res)
	}
	got := reload(t, db, plan.ID.String())
	first, second := got.Activities[0], got.Activities[1]
	if first.Department != "Finance" || first.RiskName != "Treasury (current)" || first.RiskLevel != "High" {
		t.Fatalf("first row not filled: %+v", first)
	}
	if second.Department != "Keep me" || second.RiskName != "Keep me too" || second.RiskLevel != "Low" {
		t.Fatalf("second row changed although its fields had values: %+v", second)
	}

	again, err := BackfillSnapshots(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if again.PlansChanged != 0 || again.FieldsFilled != 0 {
		t.Fatalf("second run changed data: %+v", again)
	}
}

func TestBackfillLevelUsesPlanYear(t *testing.T) {
	db := testDB(t)
	seedRefs(t, db, yearScore("uni-1", 2026, "High"))
	plan := seedPlan(t, db, 2027, models.AnnualAuditActivity{AuditUniverseID: "uni-1"})

	res, err := BackfillSnapshots(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if res.PlansChanged != 0 {
		t.Fatalf("2026 score written into a 2027 plan: %+v", res)
	}
	if lvl := reload(t, db, plan.ID.String()).Activities[0].RiskLevel; lvl != "" {
		t.Fatalf("RiskLevel = %q, want empty", lvl)
	}
}

// A dry run is a transaction that is rolled back: the counts are reported and
// nothing is stored.
func TestBackfillDryRunWritesNothing(t *testing.T) {
	db := testDB(t)
	seedRefs(t, db, dept("dep-1", "Finance"))
	plan := seedPlan(t, db, 2026, models.AnnualAuditActivity{DepartmentID: "dep-1"})

	errDryRun := errors.New("dry run")
	var res BackfillResult
	err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		res, err = BackfillSnapshots(context.Background(), tx)
		if err != nil {
			return err
		}
		return errDryRun
	})
	if !errors.Is(err, errDryRun) {
		t.Fatalf("transaction error = %v", err)
	}
	if res.PlansChanged != 1 || res.FieldsFilled != 1 {
		t.Fatalf("dry run counts = %+v", res)
	}
	if got := reload(t, db, plan.ID.String()).Activities[0].Department; got != "" {
		t.Fatalf("dry run wrote Department = %q", got)
	}
}
