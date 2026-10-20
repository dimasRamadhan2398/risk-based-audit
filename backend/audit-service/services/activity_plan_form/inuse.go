package activityplanform

import (
	"audit-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// inUseCondition is the "in use" rule (contract 7.4): the row's status is no
// longer PLANNED, or a live audit execution points at it. Fixed SQL, no input.
const inUseCondition = `(UPPER(COALESCE(audit_activities.status, '')) NOT IN ('PLANNED', '')
 OR EXISTS (SELECT 1 FROM audit_executions e
            WHERE e.activity_id = audit_activities.id AND e.deleted_at IS NULL))`

// InUse returns the subset of the given audit_activities ids that are in use,
// in one query. In-use rows cannot be removed from an RKA, their RKA cannot be
// deleted and their Kode Penugasan is locked.
func InUse(db *gorm.DB, ids []uuid.UUID) (map[uuid.UUID]bool, error) {
	out := map[uuid.UUID]bool{}
	if len(ids) == 0 {
		return out, nil
	}
	var used []uuid.UUID
	if err := db.Model(&models.AuditActivity{}).
		Where("audit_activities.id IN ?", ids).
		Where(inUseCondition).
		Pluck("audit_activities.id", &used).Error; err != nil {
		return nil, err
	}
	for _, id := range used {
		out[id] = true
	}
	return out, nil
}
