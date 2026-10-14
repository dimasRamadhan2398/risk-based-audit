package audit_execution_sync

import (
	"encoding/json"
	"testing"

	"audit-service/models"
)

// full is a Signals value that has every stage's evidence, so the expected
// progress comes from the stage under test and not from missing data.
func full() Signals {
	return Signals{
		AnnualPlanLinked:           true,
		AssignmentLetterIssued:     true,
		EntryAlignment:             entryItems,
		FieldworkRecords:           1,
		TestControls:               4,
		TestControlsTested:         4,
		TestControlsWithFinding:    2,
		TestControlsWithMitigation: 1,
		WorkingPaperCauses:         2,
		WorkingPaperPlansComplete:  2,
	}
}

func TestDeriveProgress(t *testing.T) {
	cases := []struct {
		name string
		s    Signals
		want int
	}{
		// Planning, 0
		{"no sources", Signals{}, 0},
		{"annual plan only", Signals{AnnualPlanLinked: true}, 0},

		// Entry meeting, 1..25
		{"letter issued, no entry items", Signals{AssignmentLetterIssued: true}, 1},
		{"letter issued, one entry item", Signals{AssignmentLetterIssued: true, EntryAlignment: 1}, 9},
		{"letter issued, two entry items", Signals{AssignmentLetterIssued: true, EntryAlignment: 2}, 17},
		{"letter issued, all entry items (boundary 25)", Signals{AssignmentLetterIssued: true, EntryAlignment: entryItems}, 25},

		// Fieldwork, 26..50
		{"boundary 26: fieldwork evidence, no test controls", Signals{AssignmentLetterIssued: true, FieldworkRecords: 3}, 26},
		{"fieldwork, none tested", Signals{AssignmentLetterIssued: true, FieldworkRecords: 1, TestControls: 4}, 26},
		{"fieldwork, half tested", Signals{AssignmentLetterIssued: true, FieldworkRecords: 1, TestControls: 4, TestControlsTested: 2}, 38},
		{"boundary 50: fieldwork, all tested", Signals{AssignmentLetterIssued: true, FieldworkRecords: 1, TestControls: 4, TestControlsTested: 4}, 50},
		{"tested controls without fieldwork records count as evidence", Signals{AssignmentLetterIssued: true, TestControls: 4, TestControlsTested: 1}, 32},

		// Working paper header, risk and sample rows are not fieldwork evidence
		{"WP header only, letter issued, stays at entry", Signals{AssignmentLetterIssued: true, WorkingPaperRecords: 5}, 1},
		{"WP header only, no letter, planning", Signals{WorkingPaperRecords: 5}, 0},
		{"WP header only does not pass the 30 mark", Signals{AssignmentLetterIssued: true, WorkingPaperRecords: 9, EntryAlignment: entryItems}, 25},

		// Draft findings, 55..75
		{"causes without plans stay in fieldwork (<=50)", Signals{AssignmentLetterIssued: true, FieldworkRecords: 1, TestControls: 4, TestControlsTested: 4, WorkingPaperCauses: 1, TestControlsWithMitigation: 1}, 50},
		{"causes and plans without mitigation stay in fieldwork", Signals{AssignmentLetterIssued: true, FieldworkRecords: 1, TestControls: 4, TestControlsTested: 4, WorkingPaperCauses: 2, WorkingPaperPlansComplete: 2}, 50},
		{"boundary 55: one plan against many causes", Signals{AssignmentLetterIssued: true, FieldworkRecords: 1, TestControls: 4, TestControlsTested: 4, WorkingPaperCauses: 40, WorkingPaperPlansComplete: 1, TestControlsWithMitigation: 1}, 55},
		{"55 + ratio: one of four causes planned", Signals{AssignmentLetterIssued: true, FieldworkRecords: 1, WorkingPaperCauses: 4, WorkingPaperPlansComplete: 1, TestControlsWithMitigation: 1}, 60},
		{"half of causes planned", Signals{AssignmentLetterIssued: true, FieldworkRecords: 1, WorkingPaperCauses: 4, WorkingPaperPlansComplete: 2, TestControlsWithMitigation: 1}, 65},
		{"boundary 75: every cause planned", Signals{AssignmentLetterIssued: true, FieldworkRecords: 1, WorkingPaperCauses: 4, WorkingPaperPlansComplete: 4, TestControlsWithMitigation: 1}, 75},
		{"plans beyond causes clamp to 75", Signals{AssignmentLetterIssued: true, FieldworkRecords: 1, WorkingPaperCauses: 1, WorkingPaperPlansComplete: 3, TestControlsWithMitigation: 1}, 75},
		{"no causes: findings on controls are the base", Signals{AssignmentLetterIssued: true, FieldworkRecords: 1, TestControls: 4, TestControlsTested: 4, TestControlsWithFinding: 4, WorkingPaperPlansComplete: 4, TestControlsWithMitigation: 1}, 75},

		// Plans without causes or findings never reach the draft band
		{"plans without causes or findings", Signals{AssignmentLetterIssued: true, FieldworkRecords: 1, WorkingPaperPlansComplete: 2, TestControlsWithMitigation: 1}, 26},

		// Mitigation gate (draftGate): findings + complete plans + mitigation, all three.
		// Without mitigation the cap is the fieldwork top, 50.
		{"gate: no mitigation, causes and plans, all controls tested -> cap 50", Signals{AssignmentLetterIssued: true, FieldworkRecords: 1, TestControls: 4, TestControlsTested: 4, WorkingPaperCauses: 4, WorkingPaperPlansComplete: 4}, 50},
		{"gate: no mitigation, fieldwork only -> 26 not 50", Signals{AssignmentLetterIssued: true, FieldworkRecords: 1, TestControls: 4, WorkingPaperCauses: 1, WorkingPaperPlansComplete: 1}, 26},
		{"gate: no mitigation, no fieldwork, letter issued -> entry band", Signals{AssignmentLetterIssued: true, EntryAlignment: entryItems, WorkingPaperCauses: 1, WorkingPaperPlansComplete: 1}, 25},
		{"gate: mitigation without plans -> cap 50", Signals{AssignmentLetterIssued: true, FieldworkRecords: 1, TestControls: 4, TestControlsTested: 4, WorkingPaperCauses: 2, TestControlsWithMitigation: 1}, 50},
		{"gate: mitigation and plans without findings -> cap 50", Signals{AssignmentLetterIssued: true, FieldworkRecords: 1, TestControls: 4, TestControlsTested: 4, WorkingPaperPlansComplete: 2, TestControlsWithMitigation: 1}, 50},
		{"gate: mitigation and plans, findings from test controls, 1 of 2 planned -> 65", Signals{AssignmentLetterIssued: true, FieldworkRecords: 1, TestControls: 4, TestControlsTested: 4, TestControlsWithFinding: 2, WorkingPaperPlansComplete: 1, TestControlsWithMitigation: 1}, 65},
		{"gate: mitigation on one control is enough", Signals{AssignmentLetterIssued: true, FieldworkRecords: 1, WorkingPaperCauses: 2, WorkingPaperPlansComplete: 2, TestControlsWithMitigation: 1}, 75},

		// Reporting, 76 / 88 / 90 / 100
		{"boundary 76: LHA drafted", Signals{AssignmentLetterIssued: true, FieldworkRecords: 1, LHAExists: true}, ProgressLHADraft},
		{"boundary 88: LHA in review", Signals{AssignmentLetterIssued: true, FieldworkRecords: 1, LHAExists: true, LHAInReview: true}, ProgressLHAReview},
		{"boundary 90: LHA awaiting CAE", Signals{AssignmentLetterIssued: true, FieldworkRecords: 1, LHAExists: true, LHAInReview: true, LHAAwaitingCAE: true}, ProgressLHACAEPending},
		{"boundary 100: LHA approved", Signals{AssignmentLetterIssued: true, FieldworkRecords: 1, LHAExists: true, LHAApproved: true}, ProgressCompleted},
		{"approved LHA wins over unfinished test controls", Signals{AssignmentLetterIssued: true, TestControls: 4, LHAExists: true, LHAApproved: true}, ProgressCompleted},
		{"approved LHA wins over in-review flag", Signals{AssignmentLetterIssued: true, LHAExists: true, LHAInReview: true, LHAApproved: true}, ProgressCompleted},
		{"approved LHA wins over awaiting-CAE flag", Signals{AssignmentLetterIssued: true, LHAExists: true, LHAInReview: true, LHAAwaitingCAE: true, LHAApproved: true}, ProgressCompleted},
		{"awaiting CAE is not final: never 100", Signals{AssignmentLetterIssued: true, FieldworkRecords: 1, TestControls: 4, TestControlsTested: 4, WorkingPaperCauses: 1, WorkingPaperPlansComplete: 1, TestControlsWithMitigation: 1, LHAExists: true, LHAInReview: true, LHAAwaitingCAE: true}, ProgressLHACAEPending},

		// Draft letter rule: 0 unless real fieldwork evidence
		{"draft or missing letter with LHA and WP, no fieldwork: 0", Signals{LHAExists: true, LHAApproved: true, WorkingPaperCauses: 1, WorkingPaperPlansComplete: 1, TestControlsWithMitigation: 1, TestControls: 1}, 0},
		{"draft letter with WP header and plans: 0", Signals{WorkingPaperRecords: 3, WorkingPaperCauses: 1, WorkingPaperPlansComplete: 1, TestControlsWithMitigation: 1}, 0},
		{"draft letter with fieldwork evidence is trusted", Signals{FieldworkRecords: 1, TestControls: 4, TestControlsTested: 1}, 32},
		{"draft letter with approved LHA and fieldwork evidence", Signals{FieldworkRecords: 1, LHAExists: true, LHAApproved: true}, ProgressCompleted},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := DeriveProgress(tc.s); got != tc.want {
				t.Fatalf("DeriveProgress() = %d, want %d", got, tc.want)
			}
		})
	}
}

// Every derived value must be inside the band of its status.
func TestDeriveProgressMatchesStatusBuckets(t *testing.T) {
	cases := []struct {
		name     string
		progress int
		status   string
	}{
		{"planned", ProgressPlanned, models.AuditStatusPlanning},
		{"entry min", ProgressEntryMin, models.AuditStatusEntryMeeting},
		{"entry max", ProgressEntryMin + entrySpan, models.AuditStatusEntryMeeting},
		{"fieldwork min", ProgressFieldworkMin, models.AuditStatusFieldwork},
		{"fieldwork max", ProgressFieldworkMin + fieldworkSpan, models.AuditStatusFieldwork},
		{"draft min", ProgressDraftMin, models.AuditStatusDraftFindings},
		{"draft max", ProgressDraftMin + draftSpan, models.AuditStatusDraftFindings},
		{"LHA draft", ProgressLHADraft, models.AuditStatusReporting},
		{"LHA review", ProgressLHAReview, models.AuditStatusReporting},
		{"LHA CAE pending", ProgressLHACAEPending, models.AuditStatusReporting},
		{"completed", ProgressCompleted, models.AuditStatusCompleted},
	}
	for _, tc := range cases {
		got, _ := models.MapProgressToStatus(tc.progress)
		if got != tc.status {
			t.Errorf("%s: progress %d maps to %q, want %q", tc.name, tc.progress, got, tc.status)
		}
	}
}

// The bands must read as the product spec: 0, 1..25, 26..50, 51..75, 76..99, 100.
func TestBandBoundaries(t *testing.T) {
	cases := []struct {
		progress int
		status   string
	}{
		{0, models.AuditStatusPlanning},
		{1, models.AuditStatusEntryMeeting},
		{25, models.AuditStatusEntryMeeting},
		{26, models.AuditStatusFieldwork},
		{50, models.AuditStatusFieldwork},
		{51, models.AuditStatusDraftFindings},
		{75, models.AuditStatusDraftFindings},
		{76, models.AuditStatusReporting},
		{88, models.AuditStatusReporting},
		{90, models.AuditStatusReporting},
		{99, models.AuditStatusReporting},
		{100, models.AuditStatusCompleted},
	}
	for _, tc := range cases {
		got, _ := models.MapProgressToStatus(tc.progress)
		if got != tc.status {
			t.Errorf("progress %d maps to %q, want %q", tc.progress, got, tc.status)
		}
	}
}

// Derived progress never leaves the spec's bands.
func TestDerivedProgressStaysInBands(t *testing.T) {
	inBand := func(p int) bool {
		return p == 0 || (p >= 1 && p <= 25) || (p >= 26 && p <= 50) || (p >= 55 && p <= 75) || p == 76 || p == 88 || p == 90 || p == 100
	}
	if !inBand(DeriveProgress(full())) {
		t.Fatalf("full signals out of band: %d", DeriveProgress(full()))
	}
	for _, s := range []Signals{
		{AssignmentLetterIssued: true, EntryAlignment: entryItems},
		{AssignmentLetterIssued: true, FieldworkRecords: 1, TestControls: 7, TestControlsTested: 3},
		{AssignmentLetterIssued: true, FieldworkRecords: 1, WorkingPaperCauses: 3, WorkingPaperPlansComplete: 2, TestControlsWithMitigation: 1},
		full(),
	} {
		if p := DeriveProgress(s); !inBand(p) {
			t.Errorf("progress %d for %+v is outside the spec bands", p, s)
		}
	}
}

func TestRatio(t *testing.T) {
	cases := []struct{ part, whole, span, want int }{
		{0, 0, 20, 0},
		{3, 0, 20, 0},
		{0, 4, 20, 0},
		{1, 4, 20, 5},
		{2, 3, 20, 13},
		{9, 4, 20, 20},
	}
	for _, tc := range cases {
		if got := ratio(tc.part, tc.whole, tc.span); got != tc.want {
			t.Errorf("ratio(%d,%d,%d) = %d, want %d", tc.part, tc.whole, tc.span, got, tc.want)
		}
	}
}

func TestLetterIssued(t *testing.T) {
	cases := map[string]bool{
		"Draft":            false,
		"  draft  ":        false,
		"DRAFT":            false,
		"":                 false,
		"Published":        true,
		"PUBLISHED":        true,
		"Waiting Approval": true,
		"Archived":         true,
	}
	for status, want := range cases {
		if got := letterIssued(status); got != want {
			t.Errorf("letterIssued(%q) = %v, want %v", status, got, want)
		}
	}
}

func TestLHAStatusClassification(t *testing.T) {
	approved := []string{"Final", "FINAL", " approved ", "Published"}
	for _, s := range approved {
		if !models.IsLHAApproved(s) {
			t.Errorf("IsLHAApproved(%q) = false, want true", s)
		}
	}
	// Approved statuses are the only ones that give 100
	for _, s := range []string{"DRAFT", "Draft", "IN_REVIEW", "In review", "Waiting Approval", "REJECTED", "CANCELLED"} {
		if models.IsLHAApproved(s) {
			t.Errorf("IsLHAApproved(%q) = true, want false", s)
		}
	}

	for _, s := range []string{"REJECTED", "rejected", "CANCELLED", "Canceled", " cancelled "} {
		if !models.IsLHARejectedOrCancelled(s) {
			t.Errorf("IsLHARejectedOrCancelled(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"DRAFT", "Final", "IN_REVIEW"} {
		if models.IsLHARejectedOrCancelled(s) {
			t.Errorf("IsLHARejectedOrCancelled(%q) = true, want false", s)
		}
	}

	for _, s := range []string{"IN_REVIEW", "In review", "Waiting Approval", "PENDING_CAE", "submitted"} {
		if !models.IsLHAInReview(s) {
			t.Errorf("IsLHAInReview(%q) = false, want true", s)
		}
	}

	// Waiting on CAE (90) is a subset of in review (88).
	for _, s := range []string{"Waiting Approval", "PENDING_CAE", "pending approval", "CAE pending", "waiting-cae"} {
		if !models.IsLHAAwaitingCAE(s) {
			t.Errorf("IsLHAAwaitingCAE(%q) = false, want true", s)
		}
	}
	for _, s := range []string{"IN_REVIEW", "In review", "submitted", "PENDING REVIEW", "DRAFT", "Final", "REJECTED", ""} {
		if models.IsLHAAwaitingCAE(s) {
			t.Errorf("IsLHAAwaitingCAE(%q) = true, want false", s)
		}
	}
	for _, s := range []string{"DRAFT", "Final", "REJECTED", "CANCELLED", ""} {
		if models.IsLHAInReview(s) {
			t.Errorf("IsLHAInReview(%q) = true, want false", s)
		}
	}
}

// The sources JSON is the frontend contract (frontend/types/audit.ts AuditExecutionSources
// and utils/executionSources.ts). Renaming a tag fails here.
func TestSignalsJSONContract(t *testing.T) {
	b, err := json.Marshal(Signals{})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]json.RawMessage
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{
		"annual_plan", "assignment_letter", "fieldwork_records", "test_controls",
		"test_controls_tested", "test_controls_with_finding", "mitigation_plans",
		"working_paper_records", "working_paper_plans", "working_paper_causes",
		"result_report", "result_report_approved",
	} {
		if _, ok := got[key]; !ok {
			t.Errorf("sources JSON missing %q", key)
		}
	}
}
