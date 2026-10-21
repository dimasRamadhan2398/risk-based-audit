package activityplanform

import (
	"fmt"
	"net/http"

	"audit-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Delete soft-deletes an activity plan (contract 6.5). For an RKA, every row
// goes with it: rows the RKA created are soft-deleted, rows it adopted (they
// existed before the RKA) are detached. 409 ACTIVITY_PLAN_IN_USE when any row
// is in use. Legacy plans can always be deleted.
func Delete(db *gorm.DB, id uuid.UUID) error {
	return db.Transaction(func(tx *gorm.DB) error {
		plan, err := loadPlan(tx, id, true)
		if err != nil {
			return err
		}
		if isRKA(plan) {
			rows, err := planRows(tx, plan.ID)
			if err != nil {
				return err
			}
			ids := make([]uuid.UUID, len(rows))
			for i := range rows {
				ids[i] = rows[i].ID
			}
			inUse, err := InUse(tx, ids)
			if err != nil {
				return err
			}
			for i, a := range rows {
				if inUse[a.ID] {
					return Fail(CodeActivityPlanInUse, http.StatusConflict,
						fmt.Sprintf("assignment %d (%s) has started; the activity plan cannot be deleted", i+1, a.ProjectCode),
						Details{Row: i + 1, AssignmentCode: a.ProjectCode, ActivityPlanID: plan.ID.String()})
				}
			}
			for _, a := range rows {
				if err := releaseRow(tx, a, plan.CreatedAt); err != nil {
					return err
				}
			}
		}
		return tx.Delete(&models.ActivityPlan{}, "id = ?", plan.ID).Error
	})
}
