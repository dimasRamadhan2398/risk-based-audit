// Package activityplan holds request handling specific to /activity-plans that
// the generic crud handlers do not cover: assigning the Activity ID
// (PlannedActivity.ActivityCode) of each planned activity.
package activityplan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"audit-service/models"
	"audit-service/pkg/activitycode"
	apperrors "audit-service/pkg/errors"

	"gorm.io/gorm"
)

const codeKey = "activityCode"

// AssignCodesOnCreate is a crud.CreateHook for ActivityPlan. Every planned
// activity gets a new Activity ID for its category and the plan's year;
// activityCode values sent by the client are ignored.
func AssignCodesOnCreate(tx *gorm.DB, entity interface{}) error {
	plan, ok := entity.(*models.ActivityPlan)
	if !ok {
		return fmt.Errorf("activityplan: unexpected entity %T", entity)
	}
	if len(plan.PlannedActivities) == 0 {
		return nil
	}
	year := activitycode.PlanYear(plan.PlanYear, plan.PlanPeriodStart, plan.CreationDate)
	reqs := make([]activitycode.Request, len(plan.PlannedActivities))
	for i, a := range plan.PlannedActivities {
		reqs[i] = activitycode.Request{AuditType: a.Category}
	}
	codes, err := activitycode.NextMany(tx, year, reqs)
	if err != nil {
		return err
	}
	for i := range plan.PlannedActivities {
		plan.PlannedActivities[i].ActivityCode = codes[i]
	}
	return nil
}

// AssignCodesOnUpdate is a crud.UpdateHook for ActivityPlan. It only acts when
// the update replaces planned_activities. An activity that was already stored
// in this plan keeps its Activity ID, even if its category or the plan year
// changed; it is matched by "id" first, then by the "activityCode" it sends.
// Other activities (new ones, or ones stored before codes existed) get a new
// code for their category and the plan's year as it is after this update.
// A client cannot claim a code that was not stored in this plan.
func AssignCodesOnUpdate(tx *gorm.DB, existing interface{}, columns map[string]interface{}) error {
	plan, ok := existing.(*models.ActivityPlan)
	if !ok {
		return fmt.Errorf("activityplan: unexpected entity %T", existing)
	}
	raw, present := columns["planned_activities"]
	if !present || raw == nil {
		return nil
	}
	var data []byte
	switch v := raw.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return nil
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber() // keep numbers exactly as sent
	var incoming []map[string]interface{}
	if err := dec.Decode(&incoming); err != nil {
		return apperrors.New("VALIDATION_ERROR", "plannedActivities must be an array of objects", http.StatusBadRequest)
	}

	kept := KeepCodes(plan.PlannedActivities, incoming)

	year := activitycode.PlanYear(
		stringColumn(columns, "plan_year", plan.PlanYear),
		stringColumn(columns, "plan_period_start", plan.PlanPeriodStart),
		stringColumn(columns, "creation_date", plan.CreationDate),
	)
	var need []int
	var reqs []activitycode.Request
	for i, code := range kept {
		if code == "" {
			need = append(need, i)
			reqs = append(reqs, activitycode.Request{AuditType: str(incoming[i]["category"])})
		}
	}
	codes, err := activitycode.NextMany(tx, year, reqs)
	if err != nil {
		return err
	}
	for k, i := range need {
		kept[i] = codes[k]
	}
	for i := range incoming {
		if incoming[i] == nil {
			incoming[i] = map[string]interface{}{}
		}
		incoming[i][codeKey] = kept[i]
	}

	out, err := json.Marshal(incoming)
	if err != nil {
		return err
	}
	columns["planned_activities"] = out
	return nil
}

// KeepCodes returns, for each incoming activity, the stored code it keeps, or
// "" when it needs a new one. Each stored code is kept at most once, so a
// duplicated row gets a code of its own.
func KeepCodes(stored []models.PlannedActivity, incoming []map[string]interface{}) []string {
	byID := map[string]string{}
	storedCodes := map[string]bool{}
	for _, a := range stored {
		if a.ActivityCode == "" {
			continue
		}
		storedCodes[a.ActivityCode] = true
		if a.ID != "" {
			if _, dup := byID[a.ID]; !dup {
				byID[a.ID] = a.ActivityCode
			}
		}
	}
	used := map[string]bool{}
	out := make([]string, len(incoming))
	for i, a := range incoming {
		if c, ok := byID[str(a["id"])]; ok && str(a["id"]) != "" && !used[c] {
			out[i], used[c] = c, true
			continue
		}
		if c := str(a[codeKey]); c != "" && storedCodes[c] && !used[c] {
			out[i], used[c] = c, true
		}
	}
	return out
}

func str(v interface{}) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case json.Number:
		return t.String()
	case nil:
		return ""
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

// stringColumn is the value being written to col, else the stored value.
func stringColumn(columns map[string]interface{}, col, stored string) string {
	if v, ok := columns[col]; ok && v != nil {
		return str(v)
	}
	return stored
}
