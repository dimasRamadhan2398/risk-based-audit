package models

import (
	"strings"
	"time"
)

// Follow-up buckets of a finding of an approved LHA, as reported by the
// executive summary (GET /executive-summaries/follow-up-status, the
// compilation statistics and the snapshot frozen on approval).
const (
	FollowUpBucketClosed     = "CLOSED"
	FollowUpBucketInProgress = "IN_PROGRESS"
	FollowUpBucketOverdue    = "OVERDUE"
	FollowUpBucketNoAction   = "NO_ACTION"
)

// Reasons a finding is in FollowUpBucketNoAction
const (
	// NoActionReasonNoPIC: no PIC assigned, or the finding has no ATR row
	NoActionReasonNoPIC = "NO_PIC"
	// NoActionReasonNoPlanEscalated: a PIC is assigned but no action plan was
	// saved FollowUpNoPlanEscalationDays or more after the LHA was approved
	NoActionReasonNoPlanEscalated = "NO_PLAN_14D"
	// NoActionReasonNoPlan: a PIC is assigned, no action plan saved yet
	NoActionReasonNoPlan = "NO_PLAN"
)

// FollowUpNoPlanEscalationDays is the number of calendar days (Asia/Jakarta)
// after the LHA approval from which a PLANNED ATR with a PIC is reported as
// NO_PLAN_14D instead of NO_PLAN
const FollowUpNoPlanEscalationDays = 14

// FollowUpInput is what the bucketing needs to know about one finding
type FollowUpInput struct {
	// HasATR is false when the finding has no (live) ATR row
	HasATR    bool
	Status    string
	DueDate   *time.Time
	PICUserID string
	PICName   string
	// LHAApprovedAt is when the finding's LHA was approved; nil when unknown
	LHAApprovedAt *time.Time
}

// FollowUpResult is the bucket of one finding
type FollowUpResult struct {
	// Excluded findings (CANCELLED ATR) are neither counted nor listed
	Excluded bool
	Bucket   string
	// NoActionReason is set only for FollowUpBucketNoAction
	NoActionReason string
	// OverdueDays is the number of whole days past the due date (0 unless OVERDUE)
	OverdueDays int
}

// ClassifyFollowUp puts one finding of an approved LHA into exactly one
// bucket at asOf. The checks run in this order:
//
//  1. CANCELLED ATR: excluded.
//  2. COMPLETED: CLOSED.
//  3. Open (PLANNED, IN_PROGRESS, PENDING_REVIEW or an unrecognised legacy
//     status) with a due date before the asOf date in Asia/Jakarta: OVERDUE.
//     Overdue wins over no action.
//  4. PLANNED (no action plan saved yet), or no ATR row: NO_ACTION, with
//     reason NO_PIC (no PIC or no ATR row), NO_PLAN_14D (LHA approved
//     FollowUpNoPlanEscalationDays or more calendar days before asOf) or
//     NO_PLAN.
//  5. IN_PROGRESS, PENDING_REVIEW or an unrecognised status: IN_PROGRESS.
//
// A blank status is read as PLANNED (the column default).
func ClassifyFollowUp(in FollowUpInput, asOf time.Time) FollowUpResult {
	if !in.HasATR {
		return FollowUpResult{Bucket: FollowUpBucketNoAction, NoActionReason: NoActionReasonNoPIC}
	}

	status, known := NormalizeATRStatus(in.Status)
	if strings.TrimSpace(in.Status) == "" {
		status, known = ATRStatusPlanned, true
	}
	switch status {
	case ATRStatusCancelled:
		return FollowUpResult{Excluded: true}
	case ATRStatusCompleted:
		return FollowUpResult{Bucket: FollowUpBucketClosed}
	}

	// Open from here on (ComputeATROverdue only skips COMPLETED/CANCELLED)
	if overdue, days := ComputeATROverdue(in.DueDate, status, asOf); overdue {
		return FollowUpResult{Bucket: FollowUpBucketOverdue, OverdueDays: days}
	}

	if known && status == ATRStatusPlanned {
		return FollowUpResult{Bucket: FollowUpBucketNoAction, NoActionReason: noActionReason(in, asOf)}
	}
	return FollowUpResult{Bucket: FollowUpBucketInProgress}
}

func noActionReason(in FollowUpInput, asOf time.Time) string {
	if strings.TrimSpace(in.PICUserID) == "" && strings.TrimSpace(in.PICName) == "" {
		return NoActionReasonNoPIC
	}
	if in.LHAApprovedAt != nil && !in.LHAApprovedAt.IsZero() {
		approvedDay := ATRToday(*in.LHAApprovedAt)
		if !ATRToday(asOf).Before(approvedDay.AddDate(0, 0, FollowUpNoPlanEscalationDays)) {
			return NoActionReasonNoPlanEscalated
		}
	}
	return NoActionReasonNoPlan
}
