package annualplan

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"audit-service/models"
	apperrors "audit-service/pkg/errors"
	"audit-service/pkg/logger"
	form "audit-service/services/activity_plan_form"
	"audit-service/services/annual_plan_sync"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// errCodeActivityInvalid is returned for a plan header or activity row that
// fails validation.
const errCodeActivityInvalid = "ANNUAL_PLAN_ACTIVITY_INVALID"

// PrepareOnCreate is the create hook for /annual-audit-plans. It runs the code
// and status hook, then checks the plan header and gives every activity row a
// stable id.
func PrepareOnCreate(tx *gorm.DB, entity interface{}) error {
	if err := AssignCodeAndValidateStatusOnCreate(tx, entity); err != nil {
		return err
	}
	plan, ok := entity.(*models.AuditAnnual)
	if !ok {
		return fmt.Errorf("annualplan: unexpected entity %T", entity)
	}
	if err := validateActivityType(plan.ActivityType); err != nil {
		return err
	}
	if err := validateActivities(plan.Activities); err != nil {
		return err
	}
	assignActivityIDs(plan.Activities)
	fillSnapshots(tx, plan.Year, plan.Activities, nil)
	return nil
}

// PrepareOnUpdate is the update hook for /annual-audit-plans. It runs the status
// check, then validates the header columns and rewrites the activities column:
// rows are decoded into AnnualAuditActivity (so only documented fields are
// stored) and each row keeps its id, or gets a new one when it has none or a
// duplicate. Then the RKA rules apply (form.PrepareAnnualPlanRows): the
// deprecated activityCode/justification/auditObjective/auditFocus of a kept
// row are preserved when sent empty, and removing a row an Audit Activity Plan
// holds (409 ANNUAL_PLAN_ACTIVITY_IN_USE) or changing its category so that it
// no longer fits that plan (409 ANNUAL_PLAN_ACTIVITY_TYPE_CONFLICT) is refused.
// The header activityType is deprecated; it is still validated for old clients.
func PrepareOnUpdate(tx *gorm.DB, existing interface{}, columns map[string]interface{}) error {
	if err := ValidateStatusOnUpdate(tx, existing, columns); err != nil {
		return err
	}
	if raw, ok := columns["activity_type"]; ok {
		if err := validateActivityType(raw); err != nil {
			return err
		}
	}
	raw, ok := columns["activities"]
	if !ok || raw == nil {
		return nil
	}
	body, ok := raw.([]byte)
	if !ok {
		return apperrors.New(errCodeActivityInvalid, "activities must be a list", http.StatusBadRequest)
	}
	var rows []models.AnnualAuditActivity
	if err := json.Unmarshal(body, &rows); err != nil {
		return apperrors.New(errCodeActivityInvalid, "activities: "+err.Error(), http.StatusBadRequest)
	}
	if err := validateActivities(rows); err != nil {
		return err
	}
	assignActivityIDs(rows)
	// Keep the deprecated RKA fields the PKAT form no longer sends, and refuse
	// to drop or retype a row an Audit Activity Plan holds.
	plan, ok := existing.(*models.AuditAnnual)
	if !ok {
		return fmt.Errorf("annualplan: unexpected entity %T", existing)
	}
	if err := form.PrepareAnnualPlanRows(tx, plan.ID, rows); err != nil {
		return err
	}
	// The client does not send the snapshots (see the PKAT payload), so they are
	// resolved here, and a row keeps what it had when its department or entity did not change.
	year := plan.Year
	if y, ok := yearFromColumn(columns["year"]); ok {
		year = y
	}
	fillSnapshots(tx, year, rows, &annual_plan_sync.Prior{Year: plan.Year, Rows: priorRows(plan)})
	out, err := json.Marshal(rows)
	if err != nil {
		return err
	}
	columns["activities"] = out
	return nil
}

// fillSnapshots resolves the PKAT display snapshots (department, risk name, risk
// level) of rows. It is best effort: the save does not fail when the lookup
// cannot be read. The lookup runs in a savepoint, so a failed read does not abort
// the surrounding transaction. The event sync and `audit migrate
// backfill-annual-snapshots` fill what is left.
func fillSnapshots(tx *gorm.DB, year int, rows []models.AnnualAuditActivity, prior *annual_plan_sync.Prior) {
	err := tx.Transaction(func(sp *gorm.DB) error {
		return annual_plan_sync.FillSnapshots(sp, year, rows, prior)
	})
	if err != nil {
		logger.Warn("annual plan: snapshot lookup failed, saving without it", logger.LogField("error", err))
	}
}

// priorRows returns the stored rows of plan with the row ids the PKAT form was
// given by GET (form.EnsureRowIDs, which assigns ids to legacy rows). The
// plan itself is not changed.
func priorRows(plan *models.AuditAnnual) []models.AnnualAuditActivity {
	cp := &models.AuditAnnual{ID: plan.ID, Activities: append([]models.AnnualAuditActivity(nil), plan.Activities...)}
	form.EnsureRowIDs(cp)
	return cp.Activities
}

// yearFromColumn reads a year sent in the update body (a JSON number or string).
func yearFromColumn(raw interface{}) (int, bool) {
	switch v := raw.(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(v))
		return n, err == nil
	}
	return 0, false
}

// validateActivityType accepts an empty value (not yet chosen) or one of the
// template choices. raw is the value as sent (string or nil).
func validateActivityType(raw interface{}) error {
	var s string
	switch v := raw.(type) {
	case nil:
		return nil
	case string:
		s = strings.TrimSpace(v)
	default:
		return apperrors.New(errCodeActivityInvalid, "activityType must be a string", http.StatusBadRequest)
	}
	if s == "" {
		return nil
	}
	for _, allowed := range models.AnnualPlanActivityTypes {
		if s == allowed {
			return nil
		}
	}
	return apperrors.New(errCodeActivityInvalid,
		fmt.Sprintf("activityType %q is not one of %v", s, models.AnnualPlanActivityTypes),
		http.StatusBadRequest)
}

func validateActivities(rows []models.AnnualAuditActivity) error {
	for i, row := range rows {
		if row.AuditorCount < 0 || row.TotalMandays < 0 {
			return apperrors.New(errCodeActivityInvalid,
				fmt.Sprintf("activity %d: auditorCount and totalMandays cannot be negative", i+1),
				http.StatusBadRequest)
		}
	}
	return nil
}

// assignActivityIDs sets a canonical UUID on every row. A row keeps the id the
// client sent when it is a valid UUID not used by an earlier row; otherwise it
// gets a new one.
func assignActivityIDs(rows []models.AnnualAuditActivity) {
	seen := make(map[string]bool, len(rows))
	for i := range rows {
		id, err := uuid.Parse(strings.TrimSpace(rows[i].ID))
		canonical := ""
		if err == nil {
			canonical = id.String()
		}
		if canonical == "" || seen[canonical] {
			canonical = uuid.NewString()
		}
		seen[canonical] = true
		rows[i].ID = canonical
	}
}
