// Package audit_execution_sync derives an AuditExecution's progress from the
// data the audit team already keeps in the other modules (annual plan,
// assignment letter, fieldwork, working paper and result report / LHA), so the
// Audit Execution Status no longer depends on hand-typed progress numbers.
package audit_execution_sync

// Progress bands. Each band sits inside the bucket of models.MapProgressToStatus
// and the frontend EXECUTION_PHASES, so the derived status and phase agree with
// the progress number.
//
//	0       Planning             no issued assignment letter and no fieldwork evidence
//	1..25   Entry Meeting        assignment letter issued; spread by entry alignment
//	26..50  Fieldwork            fieldwork evidence; spread by test controls tested
//	55..75  Draft Findings       causes/findings + complete plans + mitigation (draftGate)
//	76      Reporting (draft)    LHA exists in draft
//	88      Reporting (review)   LHA submitted and in review
//	90      Reporting (CAE)      LHA waiting on CAE approval (Waiting Approval, Pending CAE)
//	100     Completed            LHA approved (FINAL, APPROVED, PUBLISHED)
//
// Values 51..54, 77..87, 89 and 91..99 are never produced: the bands are
// discrete stages, and a value inside a stage is only reached through its span.
const (
	ProgressPlanned       = 0   // annual plan item only, or draft letter with no fieldwork evidence
	ProgressEntryMin      = 1   // assignment letter issued, no entry alignment yet
	ProgressFieldworkMin  = 26  // fieldwork evidence, no test controls tested yet
	ProgressDraftMin      = 55  // draft gate met: findings, complete plans and mitigation
	ProgressLHADraft      = 76  // LHA drafted
	ProgressLHAReview     = 88  // LHA submitted and in review
	ProgressLHACAEPending = 90  // LHA waiting on CAE approval, not yet published
	ProgressCompleted     = 100 // LHA approved

	entrySpan     = 24 // 1 + 24 = 25, top of the entry band
	entryItems    = 3  // entry alignment items on the letter: scope, purpose, team
	fieldworkSpan = 24 // 26 + 24 = 50, top of the fieldwork band
	draftSpan     = 20 // 55 + 20 = 75, top of the draft band
)

// Column names the derivation reads. They are constants so a rename in a model
// fails TestColumnNames instead of silently producing zero counts.
const (
	colAssignmentLetterID = "assignment_letter_id" // fieldwork, working paper header, LHA
	colWorkingPaperID     = "working_paper_id"     // working paper risks, samples, causes, plans
	colMitigationPlan     = "mitigation_plan"      // fieldwork_test_controls: the "mitigation" of the product spec
)

// Signals is what each source module reported for one execution.
type Signals struct {
	AnnualPlanLinked bool `json:"annual_plan"`

	// AssignmentLetterIssued is true when a letter for the execution exists and its
	// status is not draft. A draft or missing letter gates progress to 0 unless
	// there is real fieldwork evidence (see DeriveProgress).
	AssignmentLetterIssued bool `json:"assignment_letter"`
	// EntryAlignment counts the entry items recorded on the issued letter
	// (scope, audit purpose, team), 0..entryItems. There is no entry-meeting model,
	// so this is the only scope signal available.
	EntryAlignment int `json:"entry_alignment"`

	// Audit Fieldwork: interviews, observations, documents, samples and test controls
	FieldworkRecords        int `json:"fieldwork_records"`
	TestControls            int `json:"test_controls"`
	TestControlsTested      int `json:"test_controls_tested"`
	TestControlsWithFinding int `json:"test_controls_with_finding"`
	// TestControlsWithMitigation counts test controls with a non-empty mitigation
	// plan, i.e. the auditee has given a follow-up response. JSON: mitigation_plans.
	TestControlsWithMitigation int `json:"mitigation_plans"`

	// Working Paper. WorkingPaperRecords (header, risks, samples) is informational
	// only: it never moves progress on its own.
	WorkingPaperRecords int `json:"working_paper_records"`
	WorkingPaperCauses  int `json:"working_paper_causes"`
	// WorkingPaperPlansComplete counts plans with action_description AND
	// recommendation. JSON: working_paper_plans.
	WorkingPaperPlansComplete int `json:"working_paper_plans"`

	// Result Report (LHA). Rejected and cancelled reports are not counted.
	LHAExists   bool `json:"result_report"`
	LHAInReview bool `json:"result_report_in_review"`
	// LHAAwaitingCAE is set with LHAInReview when the in-review report waits on CAE.
	LHAAwaitingCAE bool `json:"result_report_awaiting_cae"`
	LHAApproved    bool `json:"result_report_approved"`
}

// hasFieldworkEvidence is real fieldwork work: evidence records or a tested test
// control. Working paper header, risk and sample rows do not count.
func hasFieldworkEvidence(s Signals) bool {
	return s.FieldworkRecords > 0 || s.TestControlsTested > 0
}

// draftBase is the number of findings the plans should cover: the working paper
// causes, or the test controls with a finding when no causes are recorded.
func draftBase(s Signals) int {
	if s.WorkingPaperCauses > 0 {
		return s.WorkingPaperCauses
	}
	return s.TestControlsWithFinding
}

// draftGate is the entry condition of the 55..75 band. All three parts are required:
//
//  1. findings: working paper causes, or test controls with a finding when no causes are recorded
//  2. action plans: at least one working paper plan with both an action and a recommendation
//  3. mitigation: at least one test control with a mitigation plan (the auditee's follow-up response)
//
// Without part 3 the execution stays in the fieldwork band, so the cap without
// mitigation is 50 (ProgressFieldworkMin + fieldworkSpan). Without parts 1 or 2 the
// cap is the same. Plans beyond the findings do not lift the cap.
func draftGate(s Signals) bool {
	return draftBase(s) > 0 && s.WorkingPaperPlansComplete > 0 && s.TestControlsWithMitigation > 0
}

// DeriveProgress returns the progress of the furthest stage the signals reach.
//
// Draft letter rule: until an assignment letter is issued (status not draft, and
// a letter exists), progress is 0 even if other data exists. The exception is real
// fieldwork evidence (FieldworkRecords > 0 or TestControlsTested > 0), which is
// trusted as proof that work started before the letter was updated. A missing
// letter is treated as not issued.
//
// Within a band, progress moves in proportion to test controls tested (fieldwork),
// plans complete over findings (draft), and entry items recorded (entry meeting).
func DeriveProgress(s Signals) int {
	if !s.AssignmentLetterIssued && !hasFieldworkEvidence(s) {
		return ProgressPlanned
	}
	switch {
	case s.LHAApproved:
		return ProgressCompleted
	case s.LHAAwaitingCAE:
		return ProgressLHACAEPending
	case s.LHAInReview:
		return ProgressLHAReview
	case s.LHAExists:
		return ProgressLHADraft
	case draftGate(s):
		return ProgressDraftMin + ratio(s.WorkingPaperPlansComplete, draftBase(s), draftSpan)
	case hasFieldworkEvidence(s):
		return ProgressFieldworkMin + ratio(s.TestControlsTested, s.TestControls, fieldworkSpan)
	case s.AssignmentLetterIssued:
		return ProgressEntryMin + ratio(s.EntryAlignment, entryItems, entrySpan)
	default:
		return ProgressPlanned
	}
}

// ratio returns span * part / whole, clamped to [0, span]. It is 0 when whole is 0.
func ratio(part, whole, span int) int {
	if whole <= 0 || part <= 0 {
		return 0
	}
	if part >= whole {
		return span
	}
	return part * span / whole
}
