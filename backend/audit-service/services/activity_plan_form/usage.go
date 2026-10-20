package activityplanform

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"audit-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ActivityPlanRef is the RKA that holds a PKAT row (annual plan responses,
// activities[].activityPlanRef).
type ActivityPlanRef struct {
	ActivityPlanID    string `json:"activityPlanId"`
	ActivityPlanTitle string `json:"activityPlanTitle"`
	AuditActivityID   string `json:"auditActivityId"`
	AssignmentCode    string `json:"assignmentCode"`
}

// Usage says which annual plans are in use and which PKAT rows an RKA holds.
type Usage struct {
	// Used is true for a plan with a live form-version-2 activity plan or a
	// live audit_activities row (seed and legacy rows included).
	Used map[uuid.UUID]bool
	// Refs maps a PKAT row id to the live RKA row holding it.
	Refs map[string]*ActivityPlanRef
}

// PlanUsage computes Usage for the given annual plans in two queries.
func PlanUsage(db *gorm.DB, planIDs []uuid.UUID) (*Usage, error) {
	u := &Usage{Used: map[uuid.UUID]bool{}, Refs: map[string]*ActivityPlanRef{}}
	if len(planIDs) == 0 {
		return u, nil
	}
	var rows []models.AuditActivity
	if err := db.Select("id", "annual_plan_id", "annual_activity_id", "activity_plan_id", "project_code").
		Where("annual_plan_id IN ?", planIDs).Find(&rows).Error; err != nil {
		return nil, err
	}
	var holderIDs []uuid.UUID
	for _, r := range rows {
		u.Used[r.AnnualPlanID] = true
		if r.ActivityPlanID != nil {
			holderIDs = append(holderIDs, *r.ActivityPlanID)
		}
	}

	q := db.Select("id", "plan_title", "annual_plan_id", "form_version").Model(&models.ActivityPlan{})
	if len(holderIDs) > 0 {
		q = q.Where("(annual_plan_id IN ? AND form_version = ?) OR id IN ?", planIDs, models.ActivityPlanFormRKA, holderIDs)
	} else {
		q = q.Where("annual_plan_id IN ? AND form_version = ?", planIDs, models.ActivityPlanFormRKA)
	}
	var plans []models.ActivityPlan
	if err := q.Find(&plans).Error; err != nil {
		return nil, err
	}
	holders := map[uuid.UUID]models.ActivityPlan{}
	for _, p := range plans {
		holders[p.ID] = p
		if p.FormVersion == models.ActivityPlanFormRKA && p.AnnualPlanID != nil {
			u.Used[*p.AnnualPlanID] = true
		}
	}
	for _, r := range rows {
		key := RowKey(strValue(r.AnnualActivityID))
		if key == "" || r.ActivityPlanID == nil {
			continue
		}
		h, ok := holders[*r.ActivityPlanID]
		if !ok {
			continue
		}
		u.Refs[key] = &ActivityPlanRef{
			ActivityPlanID:    h.ID.String(),
			ActivityPlanTitle: h.PlanTitle,
			AuditActivityID:   r.ID.String(),
			AssignmentCode:    r.ProjectCode,
		}
	}
	return u, nil
}

// heldRow is a live RKA row of a PKAT with its live RKA header.
type heldRow struct {
	row  models.AuditActivity
	plan models.ActivityPlan
}

// heldRows returns the PKAT rows of pkatID that a live RKA holds, by row id.
func heldRows(tx *gorm.DB, pkatID uuid.UUID) (map[string]heldRow, error) {
	var rows []models.AuditActivity
	if err := tx.Select("id", "annual_activity_id", "activity_plan_id", "project_code").
		Where("annual_plan_id = ? AND activity_plan_id IS NOT NULL AND annual_activity_id IS NOT NULL", pkatID).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := map[string]heldRow{}
	if len(rows) == 0 {
		return out, nil
	}
	ids := make([]uuid.UUID, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, *r.ActivityPlanID)
	}
	var plans []models.ActivityPlan
	if err := tx.Select("id", "plan_title", "activity_type").Where("id IN ?", ids).Find(&plans).Error; err != nil {
		return nil, err
	}
	byID := map[uuid.UUID]models.ActivityPlan{}
	for _, p := range plans {
		byID[p.ID] = p
	}
	for _, r := range rows {
		if p, ok := byID[*r.ActivityPlanID]; ok {
			out[RowKey(strValue(r.AnnualActivityID))] = heldRow{row: r, plan: p}
		}
	}
	return out, nil
}

// PrepareAnnualPlanRows applies the RKA rules to a PKAT update that replaces
// the activities (contract 7.3, 7.5, 7.6), inside the update transaction:
//
//   - the PKAT row is locked and re-read, so the checks see committed data;
//   - deprecated fields (activityCode, justification, auditObjective,
//     auditFocus) of a row that keeps its id are copied from the stored row
//     when the incoming value is empty;
//   - removing a row an RKA holds is 409 ANNUAL_PLAN_ACTIVITY_IN_USE;
//   - changing the category of a held row so that it no longer maps to its
//     RKA's type is 409 ANNUAL_PLAN_ACTIVITY_TYPE_CONFLICT.
//
// next must already have canonical row ids.
func PrepareAnnualPlanRows(tx *gorm.DB, pkatID uuid.UUID, next []models.AnnualAuditActivity) error {
	var stored models.AuditAnnual
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id", "activities").First(&stored, "id = ?", pkatID).Error; err != nil {
		return err
	}
	EnsureRowIDs(&stored)
	storedByID := map[string]int{}
	for i, r := range stored.Activities {
		storedByID[r.ID] = i
	}

	nextByID := map[string]int{}
	for i := range next {
		nextByID[next[i].ID] = i
		j, ok := storedByID[next[i].ID]
		if !ok {
			continue
		}
		old := stored.Activities[j]
		if strings.TrimSpace(next[i].ActivityCode) == "" {
			next[i].ActivityCode = old.ActivityCode
		}
		if strings.TrimSpace(next[i].Justification) == "" {
			next[i].Justification = old.Justification
		}
		if strings.TrimSpace(next[i].AuditObjective) == "" {
			next[i].AuditObjective = old.AuditObjective
		}
		if strings.TrimSpace(next[i].AuditFocus) == "" {
			next[i].AuditFocus = old.AuditFocus
		}
	}

	held, err := heldRows(tx, pkatID)
	if err != nil {
		return err
	}
	// Report in PKAT order: removed rows by their stored position, category
	// conflicts by their new position.
	for i, r := range stored.Activities {
		h, ok := held[r.ID]
		if !ok {
			continue
		}
		if _, kept := nextByID[r.ID]; !kept {
			return Fail(CodeAnnualPlanActivityInUse, http.StatusConflict,
				fmt.Sprintf("row %d is used in activity plan %q; remove it there first", i+1, h.plan.PlanTitle),
				Details{Field: "activities[" + strconv.Itoa(i) + "]", Row: i + 1, ActivityPlanID: h.plan.ID.String(),
					ActivityPlanTitle: h.plan.PlanTitle, AssignmentCode: h.row.ProjectCode})
		}
	}
	for i := range next {
		h, ok := held[next[i].ID]
		if !ok {
			continue
		}
		if !CategoryMatches(h.plan.ActivityType, next[i].Category) {
			return Fail(CodeAnnualPlanActivityConflict, http.StatusConflict,
				fmt.Sprintf("row %d: its category cannot change to %q while it is in activity plan %q (%s)",
					i+1, next[i].Category, h.plan.PlanTitle, h.plan.ActivityType),
				Details{Field: "activities[" + strconv.Itoa(i) + "].category", Row: i + 1, ActivityPlanID: h.plan.ID.String(),
					ActivityPlanTitle: h.plan.PlanTitle, AssignmentCode: h.row.ProjectCode})
		}
	}
	return nil
}

// RefreshSnapshots copies the PKAT-owned values of pkatID's rows to the
// snapshot columns of the live audit_activities rows linked to them, and the
// PKAT year to its RKAs' plan_year (contract 7.2, 7.7). Issued codes are never
// changed. Run inside the PKAT update transaction.
func RefreshSnapshots(tx *gorm.DB, pkatID uuid.UUID) error {
	var plan models.AuditAnnual
	if err := tx.First(&plan, "id = ?", pkatID).Error; err != nil {
		return err
	}
	pkat := NewPKAT(&plan)

	var rows []models.AuditActivity
	if err := tx.Where("annual_plan_id = ? AND annual_activity_id IS NOT NULL", pkatID).Find(&rows).Error; err != nil {
		return err
	}
	for i := range rows {
		a := &rows[i]
		r, _, ok := pkat.Row(strValue(a.AnnualActivityID))
		if !ok {
			continue
		}
		if a.Title == r.Name && a.AuditType == r.Category && a.TeamSize == r.AuditorCount &&
			a.TotalMandays == r.TotalMandays && sameUUID(a.AuditUniverseID, r.AuditUniverseID) &&
			sameUUID(a.TargetUnitID, r.DepartmentID) {
			continue
		}
		if err := tx.Model(&models.AuditActivity{}).Where("id = ?", a.ID).Updates(snapshot(r)).Error; err != nil {
			return err
		}
	}

	year := strconv.Itoa(plan.Year)
	return tx.Model(&models.ActivityPlan{}).
		Where("annual_plan_id = ? AND form_version = ? AND plan_year <> ?", pkatID, models.ActivityPlanFormRKA, year).
		Update("plan_year", year).Error
}

func sameUUID(stored *uuid.UUID, s string) bool {
	return uuidText(stored) == uuidText(uuidPtr(s))
}
