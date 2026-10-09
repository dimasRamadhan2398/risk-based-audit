package action_taken_report

import (
	"testing"

	"audit-service/models"

	"github.com/google/uuid"
)

func TestMigrateLegacyColumns(t *testing.T) {
	db := newDB(t)
	// The old table: the new layout (AutoMigrate adds the new columns to the
	// existing table) plus the old assignment-letter columns
	for _, col := range LegacyColumns {
		if err := db.Exec("ALTER TABLE action_taken_reports ADD COLUMN `" + col + "` TEXT").Error; err != nil {
			t.Fatal(err)
		}
	}

	// An LHA of letter ST-001 with a finding titled like legacy row 1
	lha := &models.AuditResultReport{
		ID: uuid.New(), ReportNumber: "LHA-021/SKAI/2026", ReportTitle: "Kas", Status: "Draft", AssignmentLetterID: "ST-001/SKAI/2026",
		Findings: []models.AuditReportFinding{{Title: "Rekonsiliasi Kas Harian", Category: "Major", Action: "Rekonsiliasi harian"}},
	}
	if err := db.Create(lha).Error; err != nil {
		t.Fatal(err)
	}
	var stored models.AuditResultReport
	db.First(&stored, "id = ?", lha.ID)

	linked, unlinked := uuid.New(), uuid.New()
	ins := `INSERT INTO action_taken_reports (id, status, audit_ref, title, pic, deadline, progress_description, progress, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 0, ?, ?)`
	if err := db.Exec(ins, linked.String(), "In Progress", "ST-001/SKAI/2026", "rekonsiliasi  kas harian", "Departemen Finance", "2026-04-15", "Sedang berjalan", now, now).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(ins, unlinked.String(), "Completed", "ST-999/SKAI/2026", "Opname stok", "Logistik", "15 Apr 2026", "", now, now).Error; err != nil {
		t.Fatal(err)
	}

	res, err := MigrateLegacyColumns(db)
	if err != nil {
		t.Fatal(err)
	}
	if res.Rows != 2 || res.Linked != 1 || res.Unlinked != 1 || len(res.DroppedColumns) != len(LegacyColumns) {
		t.Fatalf("result %+v", res)
	}
	for _, col := range LegacyColumns {
		if db.Migrator().HasColumn(&models.ActionTakenReport{}, col) {
			t.Fatalf("legacy column %s not dropped", col)
		}
	}

	var a models.ActionTakenReport
	if err := db.First(&a, "id = ?", linked).Error; err != nil {
		t.Fatal(err)
	}
	if a.AuditResultReportID == nil || *a.AuditResultReportID != lha.ID || a.FindingID != stored.Findings[0].ID {
		t.Fatalf("legacy row not linked to the LHA finding: %+v", a)
	}
	if a.Status != models.ATRStatusInProgress || a.PICName != "Departemen Finance" || a.ActionPlan != "Sedang berjalan" ||
		a.DueDate == nil || a.DueDate.Format("2006-01-02") != "2026-04-15" || a.FindingTitle != "rekonsiliasi  kas harian" ||
		a.ReportNumber != "LHA-021/SKAI/2026" || a.Recommendation != "Rekonsiliasi harian" {
		t.Fatalf("legacy row converted wrong: %+v", a)
	}

	var u models.ActionTakenReport
	if err := db.First(&u, "id = ?", unlinked).Error; err != nil {
		t.Fatal(err)
	}
	if u.AuditResultReportID != nil || u.FindingID != "" || u.Status != models.ATRStatusCompleted || u.DueDate != nil || u.FindingTitle != "Opname stok" {
		t.Fatalf("unlinked row: %+v", u)
	}

	// Approving the LHA does not create a second ATR for the linked finding
	db.Model(&models.AuditResultReport{}).Where("id = ?", lha.ID).UpdateColumn("status", "Final")
	if r := sync(t, db, lha.ID); r.Created != 0 {
		t.Fatalf("approval after migration created %d ATRs", r.Created)
	}

	// Second run is a no-op
	res, err = MigrateLegacyColumns(db)
	if err != nil || res.Rows != 0 || len(res.DroppedColumns) != 0 {
		t.Fatalf("second run: %+v %v", res, err)
	}
}
