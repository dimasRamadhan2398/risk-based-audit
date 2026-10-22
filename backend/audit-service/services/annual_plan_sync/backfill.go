package annual_plan_sync

import (
	"context"

	"audit-service/models"

	"gorm.io/gorm"
)

// BackfillResult counts what BackfillSnapshots filled.
type BackfillResult struct {
	PlansChanged int
	FieldsFilled int
}

// BackfillSnapshots fills the empty Department, RiskName and RiskLevel of the
// activities of every live annual plan from the lookup. A field that already has
// a value is never changed, so a correct snapshot and a value the event sync has
// refreshed are left alone. Running it again changes nothing.
//
// It does not repair a wrong value: a changed name or level reaches a plan through
// the events (and the save hooks), not through this backfill.
//
// Pass a transaction to rehearse it (roll back afterwards); each plan is written
// with the same optimistic check as the event sync.
func BackfillSnapshots(ctx context.Context, db *gorm.DB) (BackfillResult, error) {
	var refs []models.AnnualSnapshotRef
	if err := db.WithContext(ctx).Find(&refs).Error; err != nil {
		return BackfillResult{}, err
	}
	ix := indexRefs(refs)

	// fills holds the count of the last attempt per plan, so a retried write
	// is counted once.
	fills := make(map[string]int)
	err := New(db).mutatePlans(ctx, func(plan *models.AuditAnnual) bool {
		n := 0
		for i := range plan.Activities {
			row := &plan.Activities[i]
			if row.Department == "" {
				if name, ok := ix.name(models.SnapshotRefDepartment, row.DepartmentID); ok {
					row.Department = name
					n++
				}
			}
			if row.RiskName == "" {
				if name, ok := ix.name(models.SnapshotRefEntity, row.AuditUniverseID); ok {
					row.RiskName = name
					n++
				}
			}
			if row.RiskLevel == "" {
				if level, ok := ix.level(row.AuditUniverseID, plan.Year); ok {
					row.RiskLevel = level
					n++
				}
			}
		}
		fills[plan.ID.String()] = n
		return n > 0
	})
	if err != nil {
		return BackfillResult{}, err
	}

	var res BackfillResult
	for _, n := range fills {
		if n > 0 {
			res.PlansChanged++
			res.FieldsFilled += n
		}
	}
	return res, nil
}
