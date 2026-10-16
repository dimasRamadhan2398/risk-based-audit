package models

import "strings"

// Annual Audit Plan status vocabulary. The same five strings are used by
// master-service (models/annual_audit_plan.go) and by the frontend
// (types/audit.ts). Stored values are exactly these strings.
const (
	AnnualPlanStatusDraft           = "Draft"
	AnnualPlanStatusPendingReview   = "Pending Review"
	AnnualPlanStatusPendingApproval = "Pending Approval"
	AnnualPlanStatusApproved        = "Approved"
	AnnualPlanStatusRevised         = "Revised"
)

// annualPlanStatusLegacy maps old stored and submitted values to the canonical
// status. Keys are compared case-sensitively after trimming, so the upper-case
// entries are listed explicitly.
var annualPlanStatusLegacy = map[string]string{
	"":                 AnnualPlanStatusDraft,
	"Not Available":    AnnualPlanStatusDraft,
	"DRAFT":            AnnualPlanStatusDraft,
	"Work In Progress": AnnualPlanStatusPendingReview,
	"SUBMITTED":        AnnualPlanStatusPendingReview,
	"Done":             AnnualPlanStatusApproved,
	"APPROVED":         AnnualPlanStatusApproved,
	"REJECTED":         AnnualPlanStatusRevised,
	"REVISED":          AnnualPlanStatusRevised,
	"PENDING_APPROVAL": AnnualPlanStatusPendingApproval,
}

// annualPlanTransitions lists the allowed status changes. A change to the same
// status is not a transition and is always accepted.
//
//	Draft            -> Pending Review      (staff submit)
//	Pending Review   -> Pending Approval    (manager approve)
//	Pending Review   -> Revised             (manager reject)
//	Pending Approval -> Approved            (CAE approve)
//	Pending Approval -> Revised             (CAE reject)
//	Revised          -> Pending Review      (resubmit)
//	Approved         -> Revised             (RKAT revision)
var annualPlanTransitions = map[string]map[string]bool{
	AnnualPlanStatusDraft:           {AnnualPlanStatusPendingReview: true},
	AnnualPlanStatusPendingReview:   {AnnualPlanStatusPendingApproval: true, AnnualPlanStatusRevised: true},
	AnnualPlanStatusPendingApproval: {AnnualPlanStatusApproved: true, AnnualPlanStatusRevised: true},
	AnnualPlanStatusRevised:         {AnnualPlanStatusPendingReview: true},
	AnnualPlanStatusApproved:        {AnnualPlanStatusRevised: true},
}

// NormalizeAnnualPlanStatus returns the canonical status for a stored or
// submitted value, including legacy values. ok is false for any other value.
func NormalizeAnnualPlanStatus(raw string) (status string, ok bool) {
	trimmed := strings.TrimSpace(raw)
	if canon, found := annualPlanStatusLegacy[trimmed]; found {
		return canon, true
	}
	switch trimmed {
	case AnnualPlanStatusDraft, AnnualPlanStatusPendingReview, AnnualPlanStatusPendingApproval,
		AnnualPlanStatusApproved, AnnualPlanStatusRevised:
		return trimmed, true
	}
	return "", false
}

// CanTransitionAnnualPlanStatus reports whether a plan may move from one
// canonical status to another. Both arguments must already be canonical.
func CanTransitionAnnualPlanStatus(from, to string) bool {
	if from == to {
		return true
	}
	return annualPlanTransitions[from][to]
}

// AnnualPlanActivityTypes is the closed list for AuditAnnual.ActivityType (the
// Jenis Kegiatan Audit choices of the Formulir Rencana Kegiatan Audit template).
var AnnualPlanActivityTypes = []string{
	"Assurance",
	"Audit Khusus",
	"Audit Tujuan Tertentu",
	"Audit Investigasi",
}
