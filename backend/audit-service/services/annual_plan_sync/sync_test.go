package annual_plan_sync

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"audit-service/models"
	"audit-service/pkg/logger"

	"github.com/glebarez/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const sqliteUUID = `(lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' ||
	substr(lower(hex(randomblob(2))), 2) || '-' || substr('89ab', 1 + (abs(random()) % 4), 1) ||
	substr(lower(hex(randomblob(2))), 2) || '-' || lower(hex(randomblob(6))))`

func init() {
	logger.Log = zap.NewNop()
}

// testDB builds audit_annuals from the GORM schema. AutoMigrate cannot run on
// sqlite because ids default to gen_random_uuid().
func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: gormlogger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)

	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(&models.AuditAnnual{}); err != nil {
		t.Fatal(err)
	}
	var cols []string
	for _, f := range stmt.Schema.Fields {
		if f.DBName == "" {
			continue
		}
		typ := "TEXT"
		switch dt := strings.ToLower(string(f.DataType)); {
		case strings.Contains(dt, "int"):
			typ = "INTEGER"
		case dt == "bool":
			typ = "BOOLEAN"
		case dt == "time":
			typ = "DATETIME"
		}
		col := "`" + f.DBName + "` " + typ
		if f.PrimaryKey {
			col += " PRIMARY KEY DEFAULT " + sqliteUUID
		}
		cols = append(cols, col)
	}
	if err := db.Exec("CREATE TABLE audit_annuals (" + strings.Join(cols, ", ") + ")").Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func seedPlan(t *testing.T, db *gorm.DB, year int, acts ...models.AnnualAuditActivity) models.AuditAnnual {
	t.Helper()
	plan := models.AuditAnnual{Year: year, Status: "DRAFT", Activities: acts}
	if err := db.Create(&plan).Error; err != nil {
		t.Fatal(err)
	}
	return plan
}

func reload(t *testing.T, db *gorm.DB, id string) models.AuditAnnual {
	t.Helper()
	var p models.AuditAnnual
	if err := db.First(&p, "id = ?", id).Error; err != nil {
		t.Fatal(err)
	}
	return p
}

func TestEntityRenameUpdatesLinkedActivity(t *testing.T) {
	db := testDB(t)
	plan := seedPlan(t, db, 2026, models.AnnualAuditActivity{AuditUniverseID: "node-1", RiskName: "Old name"})
	s := New(db)

	if err := s.ApplyEntity(context.Background(), EntityData{ID: "node-1", Name: "New name"}); err != nil {
		t.Fatal(err)
	}
	got := reload(t, db, plan.ID.String())
	if got.Activities[0].RiskName != "New name" {
		t.Fatalf("RiskName = %q, want the renamed entity", got.Activities[0].RiskName)
	}
	if got.Activities[0].AuditUniverseID != "node-1" {
		t.Fatalf("AuditUniverseID lost: %q", got.Activities[0].AuditUniverseID)
	}
}

func TestLegacyActivityLinkedByName(t *testing.T) {
	db := testDB(t)
	plan := seedPlan(t, db, 2026, models.AnnualAuditActivity{RiskName: "Treasury"})
	s := New(db)

	if err := s.ApplyEntity(context.Background(), EntityData{ID: "node-9", Name: "  treasury "}); err != nil {
		t.Fatal(err)
	}
	got := reload(t, db, plan.ID.String())
	if got.Activities[0].AuditUniverseID != "node-9" {
		t.Fatalf("legacy activity not linked, AuditUniverseID = %q", got.Activities[0].AuditUniverseID)
	}
}

func TestYearScoreScopedToPlanYear(t *testing.T) {
	db := testDB(t)
	p2026 := seedPlan(t, db, 2026, models.AnnualAuditActivity{AuditUniverseID: "node-1", RiskName: "A", RiskLevel: "Low"})
	p2027 := seedPlan(t, db, 2027, models.AnnualAuditActivity{AuditUniverseID: "node-1", RiskName: "A", RiskLevel: "Low"})
	s := New(db)

	if err := s.ApplyYearScore(context.Background(), YearScoreData{CorporateAuditUniverseID: "node-1", Year: 2026, RiskLevel: "High"}); err != nil {
		t.Fatal(err)
	}
	if lvl := reload(t, db, p2026.ID.String()).Activities[0].RiskLevel; lvl != "High" {
		t.Fatalf("2026 plan RiskLevel = %q, want High", lvl)
	}
	if lvl := reload(t, db, p2027.ID.String()).Activities[0].RiskLevel; lvl != "Low" {
		t.Fatalf("2027 plan RiskLevel = %q, want it untouched (Low)", lvl)
	}
}

func TestDepartmentRenameAndLegacyLink(t *testing.T) {
	db := testDB(t)
	plan := seedPlan(t, db, 2026,
		models.AnnualAuditActivity{DepartmentID: "dep-1", Department: "Old Finance"},
		models.AnnualAuditActivity{Department: "HR"},
	)
	s := New(db)

	if err := s.ApplyDepartment(context.Background(), DepartmentData{ID: "dep-1", DepartmentName: "Finance"}); err != nil {
		t.Fatal(err)
	}
	if err := s.ApplyDepartment(context.Background(), DepartmentData{ID: "dep-2", DepartmentName: "hr"}); err != nil {
		t.Fatal(err)
	}
	got := reload(t, db, plan.ID.String())
	if got.Activities[0].Department != "Finance" {
		t.Fatalf("department not renamed: %q", got.Activities[0].Department)
	}
	if got.Activities[1].DepartmentID != "dep-2" {
		t.Fatalf("legacy department not linked: %q", got.Activities[1].DepartmentID)
	}
}

func TestRedeliveryIsNoOp(t *testing.T) {
	db := testDB(t)
	plan := seedPlan(t, db, 2026, models.AnnualAuditActivity{AuditUniverseID: "node-1", RiskName: "A"})
	s := New(db)
	ctx := context.Background()

	if err := s.ApplyEntity(ctx, EntityData{ID: "node-1", Name: "B"}); err != nil {
		t.Fatal(err)
	}
	first := reload(t, db, plan.ID.String())

	time.Sleep(5 * time.Millisecond)
	if err := s.ApplyEntity(ctx, EntityData{ID: "node-1", Name: "B"}); err != nil {
		t.Fatal(err)
	}
	second := reload(t, db, plan.ID.String())
	if !second.UpdatedAt.Equal(first.UpdatedAt) {
		t.Fatalf("redelivery wrote the plan again: updated_at %v -> %v", first.UpdatedAt, second.UpdatedAt)
	}
}

func TestOptimisticRetryAfterConcurrentSave(t *testing.T) {
	db := testDB(t)
	plan := seedPlan(t, db, 2026, models.AnnualAuditActivity{AuditUniverseID: "node-1", RiskName: "A"})
	s := New(db)
	calls := 0

	err := s.mutateOne(context.Background(), plan.ID.String(), func(p *models.AuditAnnual) bool {
		calls++
		if calls == 1 {
			// A user saves the plan between our read and our write.
			if err := db.Model(&models.AuditAnnual{}).Where("id = ?", plan.ID).
				Update("updated_at", time.Now().UTC().Add(time.Hour)).Error; err != nil {
				t.Fatal(err)
			}
		}
		p.Activities[0].RiskName = "B"
		return true
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("mutate called %d times, want a retry after the conflict", calls)
	}
	if name := reload(t, db, plan.ID.String()).Activities[0].RiskName; name != "B" {
		t.Fatalf("RiskName = %q, want B after retry", name)
	}
}

func TestSoftDeletedPlanNotTouched(t *testing.T) {
	db := testDB(t)
	plan := seedPlan(t, db, 2026, models.AnnualAuditActivity{AuditUniverseID: "node-1", RiskName: "A"})
	if err := db.Delete(&models.AuditAnnual{}, "id = ?", plan.ID).Error; err != nil {
		t.Fatal(err)
	}
	s := New(db)
	if err := s.ApplyEntity(context.Background(), EntityData{ID: "node-1", Name: "B"}); err != nil {
		t.Fatal(err)
	}
	var raw models.AuditAnnual
	if err := db.Unscoped().First(&raw, "id = ?", plan.ID).Error; err != nil {
		t.Fatal(err)
	}
	if raw.Activities[0].RiskName != "A" {
		t.Fatalf("soft-deleted plan was modified: %q", raw.Activities[0].RiskName)
	}
}

func TestHandlerRoutesAndSkipsSafely(t *testing.T) {
	db := testDB(t)
	plan := seedPlan(t, db, 2026, models.AnnualAuditActivity{AuditUniverseID: "node-1", RiskName: "A"})
	h := NewHandler(New(db))
	ctx := context.Background()

	data, _ := json.Marshal(EntityData{ID: "node-1", Name: "Renamed"})
	body, _ := json.Marshal(Envelope{Version: 1, EventID: "e1", EventType: EventEntityUpdated, Data: data})
	if err := h.Handle(ctx, body); err != nil {
		t.Fatal(err)
	}
	if name := reload(t, db, plan.ID.String()).Activities[0].RiskName; name != "Renamed" {
		t.Fatalf("handler did not apply entity.updated, RiskName = %q", name)
	}

	for _, msg := range []string{
		`not json`,
		`{"version":2,"event_type":"audit_universe.entity.updated","data":{}}`,
		`{"version":1,"event_type":"something.else","data":{}}`,
		`{"version":1,"event_type":"audit_universe.entity.updated","data":"oops"}`,
		`{"version":1,"event_type":"audit_universe.entity.deleted","data":{"id":"node-1"}}`,
	} {
		if err := h.Handle(ctx, []byte(msg)); err != nil {
			t.Fatalf("message %q should be acknowledged, got error %v", msg, err)
		}
	}
	if name := reload(t, db, plan.ID.String()).Activities[0].RiskName; name != "Renamed" {
		t.Fatalf("skipped messages changed the plan: %q", name)
	}
}
