package activityplanform

import (
	"fmt"
	"strings"

	"audit-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Index names of the two invariants. They also appear in unique violation
// messages on Postgres (see classifyUniqueViolation).
const (
	IndexAnnualActivityLive = "ux_audit_activities_annual_activity_live"
	IndexAnnualTypeLive     = "ux_activity_plans_annual_type_live"
)

// IndexStatements create the partial unique indexes for I1 and I2. Idempotent,
// and valid on both Postgres and sqlite.
var IndexStatements = []string{
	`CREATE UNIQUE INDEX IF NOT EXISTS ` + IndexAnnualActivityLive + `
  ON audit_activities (annual_activity_id)
  WHERE deleted_at IS NULL AND annual_activity_id IS NOT NULL AND annual_activity_id <> ''`,
	`CREATE UNIQUE INDEX IF NOT EXISTS ` + IndexAnnualTypeLive + `
  ON activity_plans (annual_plan_id, activity_type)
  WHERE deleted_at IS NULL AND annual_plan_id IS NOT NULL AND form_version = 2`,
}

// EnsureIndexes runs IndexStatements and returns the first error.
func EnsureIndexes(db *gorm.DB) error {
	for _, s := range IndexStatements {
		if err := db.Exec(s).Error; err != nil {
			return err
		}
	}
	return nil
}

// LinkResult counts what LinkLegacyActivities did.
type LinkResult struct {
	Linked    int // annual_activity_id set
	Ambiguous int // several PKAT rows (or several legacy rows) share the name
	Unmatched int // no PKAT row with that name, or the PKAT row is already linked
}

// LinkLegacyActivities links live audit_activities rows that have no
// annual_activity_id to the PKAT row of their annual plan whose name equals
// the row title (trimmed, case-insensitive), when the match is unique: exactly
// one PKAT row has that name, no other legacy row claims the same PKAT row,
// and no live row already holds its id. Only the new annual_activity_id column
// is written (without touching updated_at); activity_plan_id is left alone.
// Idempotent: linked rows are skipped on the next run.
//
// Linking lets an RKA adopt the existing row (with its code, status and
// execution link) when its PKAT row is added, instead of creating a duplicate.
func LinkLegacyActivities(db *gorm.DB) (LinkResult, error) {
	var res LinkResult
	var rows []models.AuditActivity
	if err := db.Select("id", "annual_plan_id", "title").
		Where("annual_activity_id IS NULL OR annual_activity_id = ''").
		Find(&rows).Error; err != nil {
		return res, fmt.Errorf("read unlinked audit activities: %w", err)
	}
	if len(rows) == 0 {
		return res, nil
	}

	planIDs := map[uuid.UUID]bool{}
	for _, r := range rows {
		planIDs[r.AnnualPlanID] = true
	}
	ids := make([]uuid.UUID, 0, len(planIDs))
	for id := range planIDs {
		ids = append(ids, id)
	}
	var plans []models.AuditAnnual
	if err := db.Select("id", "activities").Where("id IN ?", ids).Find(&plans).Error; err != nil {
		return res, fmt.Errorf("read annual plans: %w", err)
	}
	// plan -> normalised name -> PKAT row ids with that name
	byName := map[uuid.UUID]map[string][]string{}
	for i := range plans {
		p := &plans[i]
		EnsureRowIDs(p)
		m := map[string][]string{}
		for _, a := range p.Activities {
			name := strings.ToLower(strings.TrimSpace(a.Name))
			m[name] = append(m[name], a.ID)
		}
		byName[p.ID] = m
	}

	var held []string
	if err := db.Model(&models.AuditActivity{}).
		Where("annual_activity_id IS NOT NULL AND annual_activity_id <> ''").
		Pluck("annual_activity_id", &held).Error; err != nil {
		return res, fmt.Errorf("read linked audit activities: %w", err)
	}
	taken := map[string]bool{}
	for _, h := range held {
		taken[RowKey(h)] = true
	}

	// First pass: the candidate PKAT row of every legacy row.
	candidate := make([]string, len(rows))
	claims := map[string]int{}
	for i, r := range rows {
		name := strings.ToLower(strings.TrimSpace(r.Title))
		matches := byName[r.AnnualPlanID][name]
		switch {
		case name == "" || len(matches) == 0:
			res.Unmatched++
		case len(matches) > 1:
			res.Ambiguous++
		case taken[matches[0]]:
			res.Unmatched++
		default:
			candidate[i] = matches[0]
			claims[matches[0]]++
		}
	}
	for i, r := range rows {
		c := candidate[i]
		if c == "" {
			continue
		}
		if claims[c] > 1 {
			res.Ambiguous++
			continue
		}
		if err := db.Model(&models.AuditActivity{}).
			Where("id = ? AND (annual_activity_id IS NULL OR annual_activity_id = '')", r.ID).
			UpdateColumn("annual_activity_id", c).Error; err != nil {
			return res, fmt.Errorf("link audit activity %s: %w", r.ID, err)
		}
		res.Linked++
	}
	return res, nil
}
