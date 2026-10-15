package annualplan

import (
	"fmt"
	"net/http"

	"audit-service/models"
	apperrors "audit-service/pkg/errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Error codes returned for status problems on /annual-audit-plans.
const (
	errCodeStatusInvalid    = "ANNUAL_PLAN_STATUS_INVALID"
	errCodeStatusTransition = "ANNUAL_PLAN_STATUS_TRANSITION_NOT_ALLOWED"
)

// normalizeSubmittedStatus maps a submitted status (canonical or legacy) to the
// canonical value. A missing or null status counts as "" (Draft).
func normalizeSubmittedStatus(raw interface{}) (string, error) {
	var s string
	switch v := raw.(type) {
	case nil:
		s = ""
	case string:
		s = v
	default:
		return "", apperrors.New(errCodeStatusInvalid, "status must be a string", http.StatusBadRequest)
	}
	canon, ok := models.NormalizeAnnualPlanStatus(s)
	if !ok {
		return "", apperrors.New(errCodeStatusInvalid, fmt.Sprintf("unknown annual plan status %q", s), http.StatusBadRequest)
	}
	return canon, nil
}

// ValidateStatusOnCreate is run inside the create transaction. A new plan starts
// in Draft: a status sent on create is normalized and must be Draft, because
// every other status is reached through a transition.
func ValidateStatusOnCreate(tx *gorm.DB, entity interface{}) error {
	plan, ok := entity.(*models.AuditAnnual)
	if !ok {
		return fmt.Errorf("annualplan: unexpected entity %T", entity)
	}
	status, err := normalizeSubmittedStatus(plan.Status)
	if err != nil {
		return err
	}
	if status != models.AnnualPlanStatusDraft {
		return apperrors.New(errCodeStatusTransition,
			fmt.Sprintf("a new annual plan must start in %q, not %q", models.AnnualPlanStatusDraft, status),
			http.StatusConflict)
	}
	plan.Status = status
	return nil
}

// ValidateStatusOnUpdate is a crud.UpdateHook for AuditAnnual. When the update
// carries a status, it is normalized and checked against the transition rules
// (models/annual_plan_status.go). The stored status is re-read with a row lock
// so two concurrent updates cannot both apply the same transition. Updates
// without a status key are not affected.
func ValidateStatusOnUpdate(tx *gorm.DB, existing interface{}, columns map[string]interface{}) error {
	raw, present := columns["status"]
	if !present {
		return nil
	}
	plan, ok := existing.(*models.AuditAnnual)
	if !ok {
		return fmt.Errorf("annualplan: unexpected entity %T", existing)
	}
	to, err := normalizeSubmittedStatus(raw)
	if err != nil {
		return err
	}

	var stored models.AuditAnnual
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id", "status").
		First(&stored, "id = ?", plan.ID).Error; err != nil {
		return err
	}
	from, known := models.NormalizeAnnualPlanStatus(stored.Status)
	if !known {
		// Only reachable for values outside the vocabulary that the startup
		// normalization did not recognise; keep the plan as it is.
		from = stored.Status
	}

	if !models.CanTransitionAnnualPlanStatus(from, to) {
		return apperrors.New(errCodeStatusTransition,
			fmt.Sprintf("status cannot change from %q to %q", from, to),
			http.StatusConflict)
	}
	columns["status"] = to
	return nil
}

// AssignCodeAndValidateStatusOnCreate is the create hook for /annual-audit-plans.
func AssignCodeAndValidateStatusOnCreate(tx *gorm.DB, entity interface{}) error {
	if err := ValidateStatusOnCreate(tx, entity); err != nil {
		return err
	}
	return AssignCodeOnCreate(tx, entity)
}

// NormalizeStoredStatuses rewrites legacy status values in audit_annuals to the
// canonical vocabulary. It is idempotent and covers soft-deleted rows. Values
// outside the vocabulary are left unchanged and returned so they can be logged.
func NormalizeStoredStatuses(db *gorm.DB) (rewritten int64, unknown []string, err error) {
	// Each group is a canonical target and the legacy values that map to it.
	groups := map[string][]string{
		models.AnnualPlanStatusDraft:           {"", "Not Available", "DRAFT"},
		models.AnnualPlanStatusPendingReview:   {"Work In Progress", "SUBMITTED"},
		models.AnnualPlanStatusPendingApproval: {"PENDING_APPROVAL"},
		models.AnnualPlanStatusApproved:        {"Done", "APPROVED"},
		models.AnnualPlanStatusRevised:         {"REJECTED", "REVISED"},
	}
	for target, legacy := range groups {
		// The condition strings are constants; only the values are bound.
		cond := "status IN ?"
		if target == models.AnnualPlanStatusDraft {
			cond = "status IN ? OR status IS NULL"
		}
		res := db.Unscoped().Model(&models.AuditAnnual{}).
			Where("("+cond+")", legacy).
			Update("status", target)
		if res.Error != nil {
			return rewritten, nil, res.Error
		}
		rewritten += res.RowsAffected
	}

	var rest []string
	canonical := []string{
		models.AnnualPlanStatusDraft, models.AnnualPlanStatusPendingReview,
		models.AnnualPlanStatusPendingApproval, models.AnnualPlanStatusApproved,
		models.AnnualPlanStatusRevised,
	}
	if err := db.Unscoped().Model(&models.AuditAnnual{}).
		Where("status NOT IN ?", canonical).
		Distinct().Pluck("status", &rest).Error; err != nil {
		return rewritten, nil, err
	}
	return rewritten, rest, nil
}
