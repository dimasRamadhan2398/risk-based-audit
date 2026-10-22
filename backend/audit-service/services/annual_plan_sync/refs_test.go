package annual_plan_sync

import (
	"context"
	"errors"
	"testing"

	"audit-service/models"

	"gorm.io/gorm"
)

func seedRefs(t *testing.T, db *gorm.DB, refs ...models.AnnualSnapshotRef) {
	t.Helper()
	for _, r := range refs {
		if err := upsertRef(db, r); err != nil {
			t.Fatal(err)
		}
	}
}

func dept(id, name string) models.AnnualSnapshotRef {
	return models.AnnualSnapshotRef{Kind: models.SnapshotRefDepartment, RefID: id, Name: name}
}

func entity(id, name string) models.AnnualSnapshotRef {
	return models.AnnualSnapshotRef{Kind: models.SnapshotRefEntity, RefID: id, Name: name}
}

func yearScore(id string, year int, level string) models.AnnualSnapshotRef {
	return models.AnnualSnapshotRef{Kind: models.SnapshotRefYearScore, RefID: id, Year: year, RiskLevel: level}
}

// Events write the lookup even when no plan uses the id yet: a plan saved later
// must still find the name.
func TestDepartmentEventWritesLookupWithoutPlans(t *testing.T) {
	db := testDB(t)
	s := New(db)

	if err := s.ApplyDepartment(context.Background(), DepartmentData{ID: "DEP-1", DepartmentName: " Finance "}); err != nil {
		t.Fatal(err)
	}
	var got models.AnnualSnapshotRef
	if err := db.First(&got, "kind = ? AND ref_id = ? AND year = ?", models.SnapshotRefDepartment, "dep-1", 0).Error; err != nil {
		t.Fatalf("lookup row not written: %v", err)
	}
	if got.Name != "Finance" {
		t.Fatalf("Name = %q, want trimmed Finance", got.Name)
	}
}

func TestEmptyNameEventDoesNotOverwriteLookup(t *testing.T) {
	db := testDB(t)
	s := New(db)
	if err := s.ApplyEntity(context.Background(), EntityData{ID: "node-1", Name: "Treasury"}); err != nil {
		t.Fatal(err)
	}
	if err := s.ApplyEntity(context.Background(), EntityData{ID: "node-1", Name: "  "}); err != nil {
		t.Fatal(err)
	}
	ix := loadAll(t, db)
	if name, ok := ix.name(models.SnapshotRefEntity, "node-1"); !ok || name != "Treasury" {
		t.Fatalf("name = %q, %v; want Treasury kept", name, ok)
	}
}

func TestYearScoreEventWritesLevelPerYear(t *testing.T) {
	db := testDB(t)
	s := New(db)
	if err := s.ApplyYearScore(context.Background(), YearScoreData{CorporateAuditUniverseID: "node-1", Year: 2026, RiskLevel: "High"}); err != nil {
		t.Fatal(err)
	}
	if err := s.ApplyYearScore(context.Background(), YearScoreData{CorporateAuditUniverseID: "node-1", Year: 2027, RiskLevel: "Low"}); err != nil {
		t.Fatal(err)
	}
	ix := loadAll(t, db)
	if lvl, ok := ix.level("node-1", 2026); !ok || lvl != "High" {
		t.Fatalf("2026 level = %q, %v; want High", lvl, ok)
	}
	if lvl, ok := ix.level("node-1", 2027); !ok || lvl != "Low" {
		t.Fatalf("2027 level = %q, %v; want Low", lvl, ok)
	}
}

func loadAll(t *testing.T, db *gorm.DB) refIndex {
	t.Helper()
	var refs []models.AnnualSnapshotRef
	if err := db.Find(&refs).Error; err != nil {
		t.Fatal(err)
	}
	return indexRefs(refs)
}

func TestFillSnapshotsResolvesNewRowFromLookup(t *testing.T) {
	db := testDB(t)
	seedRefs(t, db, dept("DEP-A", "Information Technology"), entity("UNI-A", "Pengadaan"), yearScore("UNI-A", 2026, "High"))

	rows := []models.AnnualAuditActivity{{DepartmentID: "dep-a", AuditUniverseID: "UNI-A"}}
	if err := FillSnapshots(db, 2026, rows, nil); err != nil {
		t.Fatal(err)
	}
	got := rows[0]
	if got.Department != "Information Technology" || got.RiskName != "Pengadaan" || got.RiskLevel != "High" {
		t.Fatalf("snapshots = %q / %q / %q", got.Department, got.RiskName, got.RiskLevel)
	}
}

func TestFillSnapshotsLookupBeatsClientValue(t *testing.T) {
	db := testDB(t)
	seedRefs(t, db, dept("dep-a", "Finance"))
	rows := []models.AnnualAuditActivity{{DepartmentID: "dep-a", Department: "Old name typed by client"}}
	if err := FillSnapshots(db, 2026, rows, nil); err != nil {
		t.Fatal(err)
	}
	if rows[0].Department != "Finance" {
		t.Fatalf("Department = %q, want the lookup value", rows[0].Department)
	}
}

func TestFillSnapshotsMissingLookupLeavesBlankWithoutError(t *testing.T) {
	db := testDB(t)
	rows := []models.AnnualAuditActivity{{DepartmentID: "dep-x", AuditUniverseID: "uni-x"}}
	if err := FillSnapshots(db, 2026, rows, nil); err != nil {
		t.Fatalf("missing lookup must not be an error: %v", err)
	}
	if rows[0].Department != "" || rows[0].RiskName != "" || rows[0].RiskLevel != "" {
		t.Fatalf("expected blank snapshots, got %+v", rows[0])
	}
}

func TestFillSnapshotsKeepsStoredValueWhenIDUnchanged(t *testing.T) {
	db := testDB(t)
	prior := &Prior{Year: 2026, Rows: []models.AnnualAuditActivity{{
		ID: "11111111-1111-4111-8111-111111111111", DepartmentID: "dep-a", Department: "Finance",
		AuditUniverseID: "uni-a", RiskName: "Treasury", RiskLevel: "Medium",
	}}}
	// The client sends the row back without snapshots, as the PKAT form does.
	rows := []models.AnnualAuditActivity{{ID: "11111111-1111-4111-8111-111111111111", DepartmentID: "dep-a", AuditUniverseID: "uni-a"}}
	if err := FillSnapshots(db, 2026, rows, prior); err != nil {
		t.Fatal(err)
	}
	if rows[0].Department != "Finance" || rows[0].RiskName != "Treasury" || rows[0].RiskLevel != "Medium" {
		t.Fatalf("stored snapshots not kept: %+v", rows[0])
	}
}

func TestFillSnapshotsDoesNotKeepStoredValueForAnotherID(t *testing.T) {
	db := testDB(t)
	prior := &Prior{Year: 2026, Rows: []models.AnnualAuditActivity{{
		ID: "11111111-1111-4111-8111-111111111111", DepartmentID: "dep-a", Department: "Finance",
		AuditUniverseID: "uni-a", RiskName: "Treasury", RiskLevel: "Medium",
	}}}
	rows := []models.AnnualAuditActivity{{ID: "11111111-1111-4111-8111-111111111111", DepartmentID: "dep-b", AuditUniverseID: "uni-b"}}
	if err := FillSnapshots(db, 2026, rows, prior); err != nil {
		t.Fatal(err)
	}
	if rows[0].Department != "" || rows[0].RiskName != "" || rows[0].RiskLevel != "" {
		t.Fatalf("snapshots of the old id leaked to the new id: %+v", rows[0])
	}
}

func TestFillSnapshotsRiskLevelNotKeptAcrossYears(t *testing.T) {
	db := testDB(t)
	prior := &Prior{Year: 2026, Rows: []models.AnnualAuditActivity{{
		ID: "11111111-1111-4111-8111-111111111111", AuditUniverseID: "uni-a", RiskName: "Treasury", RiskLevel: "Medium",
	}}}
	rows := []models.AnnualAuditActivity{{ID: "11111111-1111-4111-8111-111111111111", AuditUniverseID: "uni-a"}}
	if err := FillSnapshots(db, 2027, rows, prior); err != nil {
		t.Fatal(err)
	}
	if rows[0].RiskLevel != "" {
		t.Fatalf("2026 level kept for a 2027 plan: %q", rows[0].RiskLevel)
	}
	if rows[0].RiskName != "Treasury" {
		t.Fatalf("name is year-independent and should be kept, got %q", rows[0].RiskName)
	}
}

func TestFillSnapshotsFailureIsReturned(t *testing.T) {
	db := testDB(t)
	if err := db.Migrator().DropTable(&models.AnnualSnapshotRef{}); err != nil {
		t.Fatal(err)
	}
	rows := []models.AnnualAuditActivity{{DepartmentID: "dep-a"}}
	err := FillSnapshots(db, 2026, rows, nil)
	if err == nil || errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("want a read error when the lookup table is missing, got %v", err)
	}
}
