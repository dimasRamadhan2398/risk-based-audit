package main

import (
	"testing"

	"master-service/models"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

func TestNormalizeAnnualPlanStatuses(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	// Minimal table: AutoMigrate of the full model uses Postgres-only types.
	if err := db.Exec(`CREATE TABLE annual_audit_plans (
		id TEXT PRIMARY KEY, plan_code TEXT NOT NULL UNIQUE, plan_title TEXT NOT NULL,
		description TEXT, audit_period_id TEXT, risk_register_id TEXT, department_id TEXT,
		priority TEXT NOT NULL, status TEXT DEFAULT 'Draft', estimated_days INTEGER DEFAULT 0,
		planned_start_date DATETIME, planned_end_date DATETIME, approved_by_id TEXT,
		approved_at DATETIME, approval_notes TEXT, requested_by_id TEXT,
		revision_number INTEGER DEFAULT 1, parent_plan_id TEXT,
		created_at DATETIME, updated_at DATETIME, deleted_at DATETIME)`).Error; err != nil {
		t.Fatalf("create table: %v", err)
	}

	legacy := map[string]models.AuditPlanStatus{
		"DRAFT":            models.AuditPlanStatusDraft,
		"SUBMITTED":        models.AuditPlanStatusPendingReview,
		"APPROVED":         models.AuditPlanStatusApproved,
		"REJECTED":         models.AuditPlanStatusRevised,
		"REVISED":          models.AuditPlanStatusRevised,
		"PENDING_APPROVAL": models.AuditPlanStatusPendingApproval,
		"Done":             models.AuditPlanStatusApproved,
		"Work In Progress": models.AuditPlanStatusPendingReview,
	}
	ids := map[uuid.UUID]models.AuditPlanStatus{}
	n := 0
	for raw, want := range legacy {
		n++
		plan := models.AnnualAuditPlan{
			ID: uuid.New(), PlanCode: "AAP-T-" + uuid.NewString(), PlanTitle: "t",
			Priority: models.AuditPlanPriorityHigh, Status: models.AuditPlanStatus(raw),
		}
		if err := db.Omit(clause.Associations).Create(&plan).Error; err != nil {
			t.Fatalf("seed %q: %v", raw, err)
		}
		ids[plan.ID] = want
	}
	// Canonical row stays as it is.
	kept := models.AnnualAuditPlan{ID: uuid.New(), PlanCode: "AAP-T-KEEP", PlanTitle: "t",
		Priority: models.AuditPlanPriorityHigh, Status: models.AuditPlanStatusPendingApproval}
	if err := db.Omit(clause.Associations).Create(&kept).Error; err != nil {
		t.Fatalf("seed kept: %v", err)
	}
	ids[kept.ID] = models.AuditPlanStatusPendingApproval

	if err := normalizeAnnualPlanStatuses(db); err != nil {
		t.Fatalf("normalize: %v", err)
	}
	for id, want := range ids {
		var got models.AnnualAuditPlan
		if err := db.Unscoped().First(&got, "id = ?", id).Error; err != nil {
			t.Fatalf("reload: %v", err)
		}
		if got.Status != want {
			t.Errorf("plan %s: status %q, want %q", id, got.Status, want)
		}
	}
	if err := normalizeAnnualPlanStatuses(db); err != nil {
		t.Fatalf("second normalize: %v", err)
	}
}
