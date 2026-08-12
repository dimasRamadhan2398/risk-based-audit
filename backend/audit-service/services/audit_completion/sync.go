package audit_completion

import (
	"audit-service/models"
	"audit-service/pkg/logger"
	"context"
	"strings"

	"gorm.io/gorm"
)

// SyncExecutionStatusToActivity maps an execution's status and progress to its parent AuditActivity.
func SyncExecutionStatusToActivity(ctx context.Context, db *gorm.DB, exec *models.AuditExecution) error {
	if exec == nil {
		return nil
	}

	var activity models.AuditActivity

	// 1. Try finding parent AuditActivity by explicit ActivityID link first
	if exec.ActivityID != nil {
		err := db.WithContext(ctx).First(&activity, "id = ?", *exec.ActivityID).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
	}

	// 2. Fall back to matching by activity title / ref / name if ActivityID link is absent
	if activity.ID == (models.AuditActivity{}).ID {
		err := db.WithContext(ctx).
			Where("LOWER(title) = LOWER(?) OR LOWER(code) = LOWER(?)", strings.TrimSpace(exec.Name), strings.TrimSpace(exec.Ref)).
			First(&activity).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil // No matching activity plan item found
			}
			return err
		}
	}

	// Determine new status for AuditActivity
	var targetStatus string
	statusUpper := strings.ToUpper(strings.TrimSpace(exec.Status))

	switch {
	case exec.Progress >= 100 || statusUpper == "COMPLETED" || statusUpper == "FINISHED" || statusUpper == "DONE":
		targetStatus = "COMPLETED"
	case statusUpper == "REPORTING" || statusUpper == "FINAL_REVIEW":
		targetStatus = "REPORTING"
	case exec.Progress > 0 || statusUpper == "IN_PROGRESS" || statusUpper == "IN PROGRESS" || statusUpper == "STARTED":
		targetStatus = "IN_PROGRESS"
	case statusUpper == "CANCELLED" || statusUpper == "CANCELED":
		targetStatus = "CANCELLED"
	default:
		targetStatus = "PLANNED"
	}

	// Update Activity status if changed
	if activity.Status != targetStatus {
		err := db.WithContext(ctx).Model(&activity).Update("status", targetStatus).Error
		if err != nil {
			logger.Error("[SyncExecutionStatusToActivity] Failed to sync activity status",
				logger.LogField("activity_id", activity.ID),
				logger.LogField("new_status", targetStatus),
				logger.LogField("error", err),
			)
			return err
		}
		logger.Info("[SyncExecutionStatusToActivity] Synced activity status",
			logger.LogField("activity_id", activity.ID),
			logger.LogField("activity_title", activity.Title),
			logger.LogField("old_status", activity.Status),
			logger.LogField("new_status", targetStatus),
		)
	}

	return nil
}

// SyncAllExecutionsForYear fetches all executions created in the given year and syncs their status to audit_activities.
func SyncAllExecutionsForYear(ctx context.Context, db *gorm.DB, year int) error {
	var executions []models.AuditExecution
	err := db.WithContext(ctx).
		Where("EXTRACT(YEAR FROM created_at) = ?", year).
		Find(&executions).Error
	if err != nil {
		return err
	}

	for i := range executions {
		if err := SyncExecutionStatusToActivity(ctx, db, &executions[i]); err != nil {
			logger.Warn("[SyncAllExecutionsForYear] Execution sync warning",
				logger.LogField("execution_id", executions[i].ID),
				logger.LogField("error", err),
			)
		}
	}
	return nil
}
