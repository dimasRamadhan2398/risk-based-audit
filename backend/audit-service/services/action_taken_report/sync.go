// Package action_taken_report keeps action taken reports (ATRs) in step with
// the audit result reports (LHAs) they follow up, and holds the ATR role
// matrix.
//
// One ATR exists per finding of an approved LHA. They are created by
// SyncFromReport, which the LHA create/update handlers run in the same
// transaction as the write, and by BackfillApprovedReports for LHAs approved
// before this existed.
package action_taken_report

import (
	"strings"
	"time"

	"audit-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RemovedFindingNote is the review note set on an untouched (PLANNED) ATR
// whose finding was removed from its LHA
const RemovedFindingNote = "Cancelled automatically: the finding was removed from the audit result report"

// SystemActor is recorded as reviewed_by for changes the server makes itself
const SystemActor = "system"

// SyncResult reports what SyncFromReport changed
type SyncResult struct {
	Created   int64
	Cancelled int64
}

// SyncFromReport brings the ATRs of report in line with it. Run it inside the
// transaction that saved the report, with the report as stored (findings
// carrying their ids).
//
//   - The report number/title copied on its ATRs are refreshed.
//   - ATRs still PLANNED (nobody worked on them) whose finding is no longer in
//     the report are cancelled with RemovedFindingNote. ATRs that were worked
//     on are left alone for a manager to decide.
//   - When the report is approved (models.IsLHAApproved), an ATR is created for
//     every finding that has none yet. The unique index on
//     (audit_result_report_id, finding_id) makes this idempotent: re-approving
//     or re-saving an approved report only adds ATRs for new findings. A
//     soft-deleted ATR still occupies its slot, so an ATR an admin deleted is
//     not recreated.
//
// Draft (non-approved) reports never get ATRs.
func SyncFromReport(tx *gorm.DB, report *models.AuditResultReport, now time.Time) (SyncResult, error) {
	var res SyncResult
	if report == nil || report.ID == uuid.Nil {
		return res, nil
	}
	reportTitle := reportTitleOf(report)

	if err := tx.Model(&models.ActionTakenReport{}).
		Where("audit_result_report_id = ?", report.ID).
		UpdateColumns(map[string]interface{}{
			"report_number": report.ReportNumber,
			"report_title":  reportTitle,
		}).Error; err != nil {
		return res, err
	}

	ids := make([]string, 0, len(report.Findings))
	for _, f := range report.Findings {
		if f.ID != "" {
			ids = append(ids, f.ID)
		}
	}

	removed := tx.Model(&models.ActionTakenReport{}).
		Where("audit_result_report_id = ? AND status = ?", report.ID, models.ATRStatusPlanned)
	if len(ids) > 0 {
		removed = removed.Where("finding_id NOT IN ?", ids)
	}
	cancel := removed.UpdateColumns(map[string]interface{}{
		"status":      models.ATRStatusCancelled,
		"review_note": RemovedFindingNote,
		"reviewed_by": SystemActor,
		"reviewed_at": now,
		"updated_at":  now,
	})
	if cancel.Error != nil {
		return res, cancel.Error
	}
	res.Cancelled = cancel.RowsAffected

	if !models.IsLHAApproved(report.Status) {
		return res, nil
	}

	reportID := report.ID
	for _, f := range report.Findings {
		if f.ID == "" {
			continue
		}
		atr := models.ActionTakenReport{
			ID:                  uuid.New(),
			AuditResultReportID: &reportID,
			FindingID:           f.ID,
			ReportNumber:        report.ReportNumber,
			ReportTitle:         reportTitle,
			FindingTitle:        strings.TrimSpace(f.Title),
			FindingCategory:     strings.TrimSpace(f.Category),
			Recommendation:      strings.TrimSpace(f.Action),
			Status:              models.ATRStatusPlanned,
			CreatedAt:           now,
			UpdatedAt:           now,
		}
		insert := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "audit_result_report_id"}, {Name: "finding_id"}},
			DoNothing: true,
		}).Omit(clause.Associations).Create(&atr)
		if insert.Error != nil {
			return res, insert.Error
		}
		res.Created += insert.RowsAffected
	}
	return res, nil
}

// SyncReportByID loads report id inside tx and syncs its ATRs
func SyncReportByID(tx *gorm.DB, id uuid.UUID, now time.Time) (SyncResult, error) {
	var report models.AuditResultReport
	if err := tx.First(&report, "id = ?", id).Error; err != nil {
		return SyncResult{}, err
	}
	return SyncFromReport(tx, &report, now)
}

func reportTitleOf(r *models.AuditResultReport) string {
	if t := strings.TrimSpace(r.ReportTitle); t != "" {
		return t
	}
	return strings.TrimSpace(r.Title)
}
