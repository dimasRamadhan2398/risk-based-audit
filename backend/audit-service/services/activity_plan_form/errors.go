// Package activityplanform implements the Formulir Rencana Kegiatan Audit
// (RKA, the Audit Activity Plan form, docs/audit-activity-plan-rework.md).
//
// An RKA is one activity_plans header (form_version 2) plus one
// audit_activities row per assignment (activity_plan_id). Each row links one
// PKAT row (audit_annuals.activities[].id) through annual_activity_id. The PKAT
// owns the object, category, universe, department, schedule, team size,
// mandays and supervisor; those are always read live from the PKAT and only
// copied to the row's snapshot columns for downstream SQL. The RKA owns the
// header (Jenis Kegiatan Audit, Judul) and the row's Kode Penugasan, Jenis
// Audit, Justifikasi, Tujuan, Fokus and Ketua Tim.
//
// Invariants enforced here and by two partial unique indexes (EnsureIndexes):
//
//	I1  at most one live audit_activities row per annual_activity_id
//	I2  at most one live RKA per (annual_plan_id, activity_type)
//	I3  every row of an RKA belongs to the RKA's PKAT and its PKAT category
//	    maps to the RKA's activity_type
//	I4  snapshot columns are written by the server only
package activityplanform

import (
	"encoding/json"
	"net/http"

	apperrors "audit-service/pkg/errors"
)

// Error codes of the activity plan API (contract section 6.7).
const (
	CodeValidation                 = "VALIDATION_ERROR"
	CodeNotFound                   = "NOT_FOUND"
	CodeAnnualPlanNotFound         = "ANNUAL_PLAN_NOT_FOUND"
	CodeAnnualPlanInactive         = "ANNUAL_PLAN_INACTIVE"
	CodeAnnualActivityNotFound     = "ANNUAL_ACTIVITY_NOT_FOUND"
	CodeAnnualActivityTypeMismatch = "ANNUAL_ACTIVITY_TYPE_MISMATCH"
	CodeAnnualActivityTaken        = "ANNUAL_ACTIVITY_TAKEN"
	CodeAssignmentCodeTaken        = "ASSIGNMENT_CODE_TAKEN"
	CodeAssignmentCodeLocked       = "ASSIGNMENT_CODE_LOCKED"
	CodeActivityRowInUse           = "ACTIVITY_ROW_IN_USE"
	CodeActivityPlanInUse          = "ACTIVITY_PLAN_IN_USE"
	CodeActivityPlanDuplicate      = "ACTIVITY_PLAN_DUPLICATE"
	CodeActivityPlanStale          = "ACTIVITY_PLAN_STALE"
	CodeActivityPlanLegacyReadonly = "ACTIVITY_PLAN_LEGACY_READONLY"
	CodeAnnualPlanInUse            = "ANNUAL_PLAN_IN_USE"
	CodeAnnualPlanActivityInUse    = "ANNUAL_PLAN_ACTIVITY_IN_USE"
	CodeAnnualPlanActivityConflict = "ANNUAL_PLAN_ACTIVITY_TYPE_CONFLICT"
	// CodeActivityRowManagedByPlan is returned by PUT /audit-activities/:id for
	// a PKAT-owned change to a row that belongs to an RKA.
	CodeActivityRowManagedByPlan = "ACTIVITY_ROW_MANAGED_BY_PLAN"
)

// Details is error.details, sent as a JSON-encoded string. Row is 1-based.
type Details struct {
	Field             string `json:"field,omitempty"`
	Row               int    `json:"row,omitempty"`
	ActivityPlanID    string `json:"activityPlanId,omitempty"`
	ActivityPlanTitle string `json:"activityPlanTitle,omitempty"`
	AssignmentCode    string `json:"assignmentCode,omitempty"`
}

// Fail builds an AppError whose Details carry d as JSON ("" when d is empty).
func Fail(code string, status int, message string, d Details) *apperrors.AppError {
	e := apperrors.New(code, message, status)
	if d != (Details{}) {
		if b, err := json.Marshal(d); err == nil {
			e.Details = string(b)
		}
	}
	return e
}

func invalid(field, message string) *apperrors.AppError {
	return Fail(CodeValidation, http.StatusBadRequest, message, Details{Field: field})
}

func invalidRow(row int, field, message string) *apperrors.AppError {
	return Fail(CodeValidation, http.StatusBadRequest, message, Details{Field: field, Row: row})
}
