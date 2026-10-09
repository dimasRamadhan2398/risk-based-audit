package action_taken_report

import (
	"encoding/json"
	"time"

	"audit-service/models"

	"gorm.io/gorm"
)

// BackfillFindingIDs gives every finding of every audit result report
// (soft-deleted ones included) a stable id, keeping the ids already there.
// Reports whose findings all have valid unique ids are not touched, so running
// it again changes nothing. Returns the number of reports updated.
//
// It writes with UpdateColumns (no hooks, updated_at unchanged): this is a
// data migration, not an edit of the report.
func BackfillFindingIDs(db *gorm.DB) (int, error) {
	updated := 0
	var batch []models.AuditResultReport
	err := db.Unscoped().Model(&models.AuditResultReport{}).
		Select("id", "findings").
		FindInBatches(&batch, 200, func(tx *gorm.DB, _ int) error {
			for _, r := range batch {
				if !models.FindingsNeedIDs(r.Findings) {
					continue
				}
				fixed := models.AssignFindingIDs(r.Findings, nil)
				encoded, err := json.Marshal(fixed)
				if err != nil {
					return err
				}
				if err := db.Session(&gorm.Session{NewDB: true}).Unscoped().
					Model(&models.AuditResultReport{}).
					Where("id = ?", r.ID).
					UpdateColumns(map[string]interface{}{
						"findings":       string(encoded),
						"findings_count": len(fixed),
					}).Error; err != nil {
					return err
				}
				updated++
			}
			return nil
		}).Error
	return updated, err
}

// BackfillResult reports what BackfillApprovedReports did
type BackfillResult struct {
	FindingIDReports int   // reports that got finding ids
	ApprovedReports  int   // approved (non-deleted) reports synced
	CreatedATRs      int64 // ATRs created
	CancelledATRs    int64 // PLANNED ATRs cancelled because their finding is gone
}

// BackfillApprovedReports creates the missing ATRs of every approved,
// non-deleted LHA (one per finding). It first backfills finding ids. Each
// report is synced in its own transaction with SyncFromReport, so it is
// idempotent: a second run creates nothing.
func BackfillApprovedReports(db *gorm.DB, now time.Time) (BackfillResult, error) {
	var res BackfillResult
	n, err := BackfillFindingIDs(db)
	res.FindingIDReports = n
	if err != nil {
		return res, err
	}

	upper := make([]string, len(models.LHAApprovedStatuses))
	copy(upper, models.LHAApprovedStatuses)

	var reports []models.AuditResultReport
	if err := db.Where("UPPER(TRIM(status)) IN ?", upper).
		Order("created_at ASC").Order("id ASC").
		Find(&reports).Error; err != nil {
		return res, err
	}
	for i := range reports {
		err := db.Transaction(func(tx *gorm.DB) error {
			r, err := SyncFromReport(tx, &reports[i], now)
			res.CreatedATRs += r.Created
			res.CancelledATRs += r.Cancelled
			return err
		})
		if err != nil {
			return res, err
		}
		res.ApprovedReports++
	}
	return res, nil
}
