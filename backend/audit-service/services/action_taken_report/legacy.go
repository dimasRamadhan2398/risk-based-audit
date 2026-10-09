package action_taken_report

import (
	"fmt"
	"strings"
	"time"

	"audit-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LegacyColumns are the columns of the old assignment-letter based
// action_taken_reports table that the LHA-bound model no longer has. Columns
// the new model reuses (finding_category, recommendation, status) are not
// listed.
var LegacyColumns = []string{
	"assignment_letter_id",
	"audit_finding_id",
	"audit_ref",
	"title",
	"department",
	"audit_object",
	"condition",
	"criteria",
	"pic",
	"deadline",
	"attachment",
	"progress_description",
}

// LegacyResult reports what MigrateLegacyColumns did
type LegacyResult struct {
	Rows           int      // legacy rows found (soft-deleted included)
	Linked         int      // rows linked to an LHA finding
	Unlinked       int      // rows left without an LHA link (still readable)
	DroppedColumns []string // legacy columns dropped
}

// MigrateLegacyColumns converts the old table layout. It must run after
// AutoMigrate (the new columns exist) and after BackfillFindingIDs (so a
// legacy row can be linked to a finding id). It is a no-op once the legacy
// columns are gone.
//
// For each existing row, in one transaction:
//   - title (or condition) -> finding_title, pic -> pic_name,
//     deadline ("YYYY-MM-DD...") -> due_date, progress_description ->
//     action_plan, status normalised (e.g. "In Progress" -> IN_PROGRESS);
//     values already present in the new columns are kept.
//   - If audit_ref names an assignment letter whose (non-deleted) LHA has a
//     finding with the same title, the row is linked to that LHA/finding
//     (unless that finding already has an ATR).
//   - Rows that cannot be linked keep audit_result_report_id/finding_id NULL;
//     the API lists them like any other ATR (view-all roles only, since they
//     have no PIC user) and they can be assigned, worked on or cancelled.
//
// Then the legacy columns (and with them their indexes and the foreign key to
// assignment_letters) are dropped. The old attachment column held only a file
// name, never a stored file, so it is not carried over.
func MigrateLegacyColumns(db *gorm.DB) (LegacyResult, error) {
	var res LegacyResult
	m := db.Migrator()
	var present []string
	for _, col := range LegacyColumns {
		if m.HasColumn(&models.ActionTakenReport{}, col) {
			present = append(present, col)
		}
	}
	if len(present) == 0 {
		return res, nil
	}
	has := make(map[string]bool, len(present))
	for _, col := range present {
		has[col] = true
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		cols := []string{"id", "status", "finding_title", "pic_name", "due_date", "action_plan",
			"audit_result_report_id", "finding_category", "recommendation"}
		for _, c := range []string{"title", "condition", "pic", "deadline", "progress_description", "audit_ref"} {
			if has[c] {
				cols = append(cols, c)
			}
		}
		var rows []map[string]interface{}
		if err := tx.Table("action_taken_reports").Select(cols).Find(&rows).Error; err != nil {
			return err
		}
		res.Rows = len(rows)

		for _, row := range rows {
			id := asString(row["id"])
			if id == "" {
				continue
			}
			updates := map[string]interface{}{}

			title := firstNonEmpty(asString(row["title"]), asString(row["condition"]))
			if asString(row["finding_title"]) == "" && title != "" {
				updates["finding_title"] = title
			}
			if asString(row["pic_name"]) == "" && asString(row["pic"]) != "" {
				updates["pic_name"] = asString(row["pic"])
			}
			if row["due_date"] == nil {
				if d, ok := models.ParseLegacyATRDeadline(asString(row["deadline"])); ok {
					updates["due_date"] = d
				}
			}
			if asString(row["action_plan"]) == "" && asString(row["progress_description"]) != "" {
				updates["action_plan"] = asString(row["progress_description"])
			}
			if s, ok := models.NormalizeATRStatus(asString(row["status"])); ok && s != asString(row["status"]) {
				updates["status"] = s
			} else if strings.TrimSpace(asString(row["status"])) == "" {
				updates["status"] = models.ATRStatusPlanned
			}

			linked := asString(row["audit_result_report_id"]) != ""
			if !linked {
				if link, ok, err := findLegacyLink(tx, asString(row["audit_ref"]), title); err != nil {
					return err
				} else if ok {
					updates["audit_result_report_id"] = link.reportID
					updates["finding_id"] = link.finding.ID
					updates["report_number"] = link.reportNumber
					updates["report_title"] = link.reportTitle
					if asString(row["finding_category"]) == "" {
						updates["finding_category"] = link.finding.Category
					}
					if asString(row["recommendation"]) == "" {
						updates["recommendation"] = link.finding.Action
					}
					linked = true
				}
			}
			if linked {
				res.Linked++
			} else {
				res.Unlinked++
			}

			if len(updates) == 0 {
				continue
			}
			if err := tx.Table("action_taken_reports").Where("id = ?", id).UpdateColumns(updates).Error; err != nil {
				return fmt.Errorf("migrate action taken report %s: %w", id, err)
			}
		}

		for _, col := range present {
			if err := tx.Migrator().DropColumn(&models.ActionTakenReport{}, col); err != nil {
				return fmt.Errorf("drop column action_taken_reports.%s: %w", col, err)
			}
			res.DroppedColumns = append(res.DroppedColumns, col)
		}
		return nil
	})
	if err != nil {
		return LegacyResult{}, err
	}
	return res, nil
}

type legacyLink struct {
	reportID     uuid.UUID
	reportNumber string
	reportTitle  string
	finding      models.AuditReportFinding
}

// findLegacyLink finds the LHA finding a legacy ATR follows up: an LHA of the
// assignment letter auditRef with a finding titled title (case-insensitive)
// that has no ATR yet
func findLegacyLink(tx *gorm.DB, auditRef, title string) (legacyLink, bool, error) {
	auditRef = strings.TrimSpace(auditRef)
	key := strings.ToLower(strings.Join(strings.Fields(title), " "))
	if auditRef == "" || key == "" {
		return legacyLink{}, false, nil
	}
	var reports []models.AuditResultReport
	if err := tx.Where("assignment_letter_id = ?", auditRef).
		Order("created_at ASC").Order("id ASC").
		Find(&reports).Error; err != nil {
		return legacyLink{}, false, err
	}
	for _, r := range reports {
		for _, f := range r.Findings {
			if f.ID == "" || strings.ToLower(strings.Join(strings.Fields(f.Title), " ")) != key {
				continue
			}
			var taken int64
			if err := tx.Table("action_taken_reports").
				Where("audit_result_report_id = ? AND finding_id = ?", r.ID, f.ID).
				Count(&taken).Error; err != nil {
				return legacyLink{}, false, err
			}
			if taken > 0 {
				continue
			}
			return legacyLink{reportID: r.ID, reportNumber: r.ReportNumber, reportTitle: reportTitleOf(&r), finding: f}, true, nil
		}
	}
	return legacyLink{}, false, nil
}

func asString(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return val
	case []byte:
		return string(val)
	case time.Time:
		return val.Format(time.RFC3339)
	case [16]byte:
		return uuid.UUID(val).String()
	case fmt.Stringer:
		return val.String()
	default:
		return fmt.Sprint(val)
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}
