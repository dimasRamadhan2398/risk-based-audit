package annualplan

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"audit-service/models"
	apperrors "audit-service/pkg/errors"

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
	return nil
}

// PrepareOnUpdate is the update hook for /annual-audit-plans. It runs the status
// check, then validates the header columns and rewrites the activities column:
// rows are decoded into AnnualAuditActivity (so only documented fields are
// stored) and each row keeps its id, or gets a new one when it has none or a
// duplicate.
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
	out, err := json.Marshal(rows)
	if err != nil {
		return err
	}
	columns["activities"] = out
	return nil
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
