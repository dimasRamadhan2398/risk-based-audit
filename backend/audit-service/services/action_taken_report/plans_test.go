package action_taken_report

import (
	"testing"

	"audit-service/models"
)

func TestATRIsDerivedFromFindingActionPlans(t *testing.T) {
	db := newDB(t)
	r := createReport(t, db, "Final", "A")
	r.Findings[0].ActionPlans = []models.FindingActionPlan{
		{Action: "Revisi SOP", TargetDate: "2026-12-31", PIC: "Budi", FinalApprover: "Kepala SKAI", Evidence: "SOP v2"},
		{Action: "Training", TargetDate: "2026-11-15", PIC: "Sari"},
	}
	if err := db.Save(r).Error; err != nil {
		t.Fatal(err)
	}
	sync(t, db, r.ID)
	atrs := atrsOf(t, db, r.ID)
	if len(atrs) != 1 {
		t.Fatalf("want 1 atr, got %d", len(atrs))
	}
	a := atrs[0]
	if len(a.PlannedActions) != 2 || a.PlannedActions[0].Evidence != "SOP v2" {
		t.Fatalf("planned actions: %+v", a.PlannedActions)
	}
	if a.FinalApprover != "Kepala SKAI" || a.PICName != "Budi" {
		t.Fatalf("approver/pic: %q %q", a.FinalApprover, a.PICName)
	}
	if a.DueDate == nil || a.DueDate.Format("2006-01-02") != "2026-11-15" {
		t.Fatalf("due date: %v", a.DueDate)
	}
	if a.Recommendation != "Revisi SOP\nTraining" {
		t.Fatalf("recommendation: %q", a.Recommendation)
	}

	// editing the LHA refreshes a PLANNED ATR the PIC has not submitted
	var stored models.AuditResultReport
	if err := db.First(&stored, "id = ?", r.ID).Error; err != nil {
		t.Fatal(err)
	}
	stored.Findings[0].ActionPlans = stored.Findings[0].ActionPlans[:1]
	stored.Findings[0].ActionPlans[0].FinalApprover = "CAE"
	if err := db.Save(&stored).Error; err != nil {
		t.Fatal(err)
	}
	sync(t, db, r.ID)
	a = atrsOf(t, db, r.ID)[0]
	if len(a.PlannedActions) != 1 || a.FinalApprover != "CAE" || a.DueDate == nil || a.DueDate.Format("2006-01-02") != "2026-12-31" {
		t.Fatalf("not refreshed: %+v due=%v", a.PlannedActions, a.DueDate)
	}
}

func TestATRLegacyActionStillWorks(t *testing.T) {
	db := newDB(t)
	r := createReport(t, db, "Final", "A")
	sync(t, db, r.ID)
	a := atrsOf(t, db, r.ID)[0]
	if a.Recommendation != "Fix A" || len(a.PlannedActions) != 0 || a.DueDate != nil || a.PICName != "" {
		t.Fatalf("legacy atr: %+v", a)
	}
}
