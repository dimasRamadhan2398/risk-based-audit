package annual_plan_sync

import (
	"strconv"
	"strings"
	"time"

	"audit-service/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// This file holds the local lookup (models.AnnualSnapshotRef) that the PKAT
// display snapshots are read from. The event handlers in sync.go write it; the
// save hooks (controllers/annual_plan) and the backfill read it.

// normRefID is the form every ref id is stored and compared in: trimmed and
// lower-case, so an upper-case id sent by a client still matches.
func normRefID(id string) string {
	return strings.ToLower(strings.TrimSpace(id))
}

// upsertRef writes the name (department, entity) or the risk level (year score)
// of one source id. Callers skip empty values, so a good name is never replaced
// by a blank one.
func upsertRef(tx *gorm.DB, ref models.AnnualSnapshotRef) error {
	ref.RefID = normRefID(ref.RefID)
	ref.UpdatedAt = time.Now().UTC()
	cols := []string{"name", "updated_at"}
	if ref.Kind == models.SnapshotRefYearScore {
		cols = []string{"risk_level", "updated_at"}
	}
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "kind"}, {Name: "ref_id"}, {Name: "year"}},
		DoUpdates: clause.AssignmentColumns(cols),
	}).Create(&ref).Error
}

// refIndex is the lookup of the rows loaded for one resolution, keyed by kind,
// ref id and year (year is 0 for departments and entities).
type refIndex map[string]models.AnnualSnapshotRef

func indexRefs(refs []models.AnnualSnapshotRef) refIndex {
	ix := make(refIndex, len(refs))
	for _, r := range refs {
		ix[refKey(r.Kind, r.RefID, r.Year)] = r
	}
	return ix
}

func refKey(kind, refID string, year int) string {
	return kind + "|" + normRefID(refID) + "|" + strconv.Itoa(year)
}

// name returns the name of a department or an entity. ok is false when the id is
// blank or not known yet.
func (ix refIndex) name(kind, refID string) (string, bool) {
	if normRefID(refID) == "" {
		return "", false
	}
	r, ok := ix[refKey(kind, refID, 0)]
	if !ok || strings.TrimSpace(r.Name) == "" {
		return "", false
	}
	return r.Name, true
}

// level returns the entity's risk level for year. ok is false when it is not known.
func (ix refIndex) level(universeID string, year int) (string, bool) {
	if normRefID(universeID) == "" || year == 0 {
		return "", false
	}
	r, ok := ix[refKey(models.SnapshotRefYearScore, universeID, year)]
	if !ok || strings.TrimSpace(r.RiskLevel) == "" {
		return "", false
	}
	return r.RiskLevel, true
}

// loadRefIndex loads the lookup rows for the given ids (any kind). An empty list
// loads nothing.
func loadRefIndex(tx *gorm.DB, ids []string) (refIndex, error) {
	if len(ids) == 0 {
		return refIndex{}, nil
	}
	var refs []models.AnnualSnapshotRef
	if err := tx.Where("ref_id IN ?", ids).Find(&refs).Error; err != nil {
		return nil, err
	}
	return indexRefs(refs), nil
}

// Prior is the stored plan that a PKAT update replaces. Rows must carry the row
// ids the PKAT form was given (form.EnsureRowIDs). Nil for a create.
type Prior struct {
	Year int
	Rows []models.AnnualAuditActivity
}

// FillSnapshots sets the Department, RiskName and RiskLevel of rows, for a plan
// of the given year, from the lookup. Per field, in order:
//
//  1. the lookup value, when the lookup knows the id;
//  2. the value the row already carries (a value sent by a client);
//  3. the stored value of the same row, when the row still points at the same
//     department or entity (and, for RiskLevel, the same year);
//  4. empty.
//
// So a save never blanks a snapshot the plan already had, a missing lookup never
// fails the save, and a row that points at a new id does not keep the old
// name. Rows are changed in place; the caller stores them.
func FillSnapshots(tx *gorm.DB, year int, rows []models.AnnualAuditActivity, prior *Prior) error {
	ids := make([]string, 0, 2*len(rows))
	for _, r := range rows {
		ids = append(ids, normRefID(r.DepartmentID), normRefID(r.AuditUniverseID))
	}
	ix, err := loadRefIndex(tx, ids)
	if err != nil {
		return err
	}

	var old map[string]models.AnnualAuditActivity
	oldYear := 0
	if prior != nil {
		oldYear = prior.Year
		old = make(map[string]models.AnnualAuditActivity, len(prior.Rows))
		for _, p := range prior.Rows {
			old[normRefID(p.ID)] = p
		}
	}

	for i := range rows {
		row := &rows[i]
		prev, hadPrev := old[normRefID(row.ID)]

		if name, ok := ix.name(models.SnapshotRefDepartment, row.DepartmentID); ok {
			row.Department = name
		} else if row.Department == "" && hadPrev && normRefID(prev.DepartmentID) != "" &&
			normRefID(prev.DepartmentID) == normRefID(row.DepartmentID) {
			row.Department = prev.Department
		}

		if name, ok := ix.name(models.SnapshotRefEntity, row.AuditUniverseID); ok {
			row.RiskName = name
		} else if row.RiskName == "" && hadPrev && normRefID(prev.AuditUniverseID) != "" &&
			normRefID(prev.AuditUniverseID) == normRefID(row.AuditUniverseID) {
			row.RiskName = prev.RiskName
		}

		if level, ok := ix.level(row.AuditUniverseID, year); ok {
			row.RiskLevel = level
		} else if row.RiskLevel == "" && hadPrev && oldYear == year && normRefID(prev.AuditUniverseID) != "" &&
			normRefID(prev.AuditUniverseID) == normRefID(row.AuditUniverseID) {
			row.RiskLevel = prev.RiskLevel
		}
	}
	return nil
}
