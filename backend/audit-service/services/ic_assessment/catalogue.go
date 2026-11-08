// Package ic_assessment implements the COSO 2013 internal control component
// assessment of a department (optionally a branch) for a year (optionally a
// quarter): the principle and component ratings, the proposed component
// scores computed from existing data, the roll-up to an overall conclusion,
// the DRAFT -> SUBMITTED -> APPROVED workflow with segregation of duties,
// period locks and an append-only event log.
//
// Spec: docs/coso-2013-component-assessment-plan.md (sections 4.2, 4.3, 5, 7)
// with the decisions of 2026-11-06 in section 13.
package ic_assessment

// Framework is the only framework assessed in phase 1
const Framework = "COSO_2013"

// Component codes, in display order
const (
	ComponentCE = "CE" // Control Environment
	ComponentRA = "RA" // Risk Assessment
	ComponentCA = "CA" // Control Activities
	ComponentIC = "IC" // Information & Communication
	ComponentMA = "MA" // Monitoring Activities
)

// ComponentCodes lists the 5 COSO 2013 components in order
var ComponentCodes = []string{ComponentCE, ComponentRA, ComponentCA, ComponentIC, ComponentMA}

// Principle is one of the 17 COSO 2013 principles. It mirrors the
// risk-service seed (coso_principles); the 17 principles are fixed by the
// standard, so audit-service keeps them as constants.
type Principle struct {
	Code          string `json:"code"`
	Number        int    `json:"number"`
	ComponentCode string `json:"component_code"`
	NameEN        string `json:"name_en"`
}

// Principles lists the 17 principles in order
var Principles = []Principle{
	{"P01", 1, ComponentCE, "Demonstrates commitment to integrity and ethical values"},
	{"P02", 2, ComponentCE, "Exercises oversight responsibility"},
	{"P03", 3, ComponentCE, "Establishes structure, authority and responsibility"},
	{"P04", 4, ComponentCE, "Demonstrates commitment to competence"},
	{"P05", 5, ComponentCE, "Enforces accountability"},
	{"P06", 6, ComponentRA, "Specifies suitable objectives"},
	{"P07", 7, ComponentRA, "Identifies and analyzes risk"},
	{"P08", 8, ComponentRA, "Assesses fraud risk"},
	{"P09", 9, ComponentRA, "Identifies and analyzes significant change"},
	{"P10", 10, ComponentCA, "Selects and develops control activities"},
	{"P11", 11, ComponentCA, "Selects and develops general controls over technology"},
	{"P12", 12, ComponentCA, "Deploys through policies and procedures"},
	{"P13", 13, ComponentIC, "Uses relevant information"},
	{"P14", 14, ComponentIC, "Communicates internally"},
	{"P15", 15, ComponentIC, "Communicates externally"},
	{"P16", 16, ComponentMA, "Conducts ongoing and/or separate evaluations"},
	{"P17", 17, ComponentMA, "Evaluates and communicates deficiencies"},
}

var principleByCode = func() map[string]Principle {
	m := make(map[string]Principle, len(Principles))
	for _, p := range Principles {
		m[p.Code] = p
	}
	return m
}()

// PrincipleByCode returns the principle with code
func PrincipleByCode(code string) (Principle, bool) {
	p, ok := principleByCode[code]
	return p, ok
}

// IsComponentCode reports whether code is one of the 5 components
func IsComponentCode(code string) bool {
	for _, c := range ComponentCodes {
		if c == code {
			return true
		}
	}
	return false
}

// Assessment statuses
const (
	StatusDraft     = "DRAFT"
	StatusSubmitted = "SUBMITTED"
	StatusApproved  = "APPROVED"
	StatusReopened  = "REOPENED"
)

// AssessmentStatuses are the valid header statuses (list filter)
var AssessmentStatuses = []string{StatusDraft, StatusSubmitted, StatusApproved, StatusReopened}

// Principle statuses; component final statuses are YES, PARTIAL, NO
const (
	PrincipleNotAssessed   = "NOT_ASSESSED"
	PrincipleYes           = "YES"
	PrinciplePartial       = "PARTIAL"
	PrincipleNo            = "NO"
	PrincipleNotApplicable = "NOT_APPLICABLE"
)

// Severities, in increasing order
const (
	SeverityNone        = "NONE"
	SeverityDeficiency  = "DEFICIENCY"
	SeveritySignificant = "SIGNIFICANT"
	SeverityMajor       = "MAJOR"
)

var severityRank = map[string]int{SeverityNone: 0, SeverityDeficiency: 1, SeveritySignificant: 2, SeverityMajor: 3}

// IsSeverity reports whether s is a known severity
func IsSeverity(s string) bool { _, ok := severityRank[s]; return ok }

// SeverityRank orders severities (unknown = -1)
func SeverityRank(s string) int {
	if r, ok := severityRank[s]; ok {
		return r
	}
	return -1
}

// Conclusions
const (
	ConclusionEffective        = "EFFECTIVE"
	ConclusionNotEffective     = "NOT_EFFECTIVE"
	ConclusionRequiresJudgment = "REQUIRES_JUDGMENT"
	ConclusionIncomplete       = "INCOMPLETE"
)

// Evidence states
const (
	StateOK                = "OK"
	StateNoData            = "NO_DATA"
	StateNotApplicable     = "NOT_APPLICABLE"
	StateSourceUnavailable = "SOURCE_UNAVAILABLE"
)

// Event actions
const (
	ActionCreate         = "CREATE"
	ActionPropose        = "PROPOSE"
	ActionUpdate         = "UPDATE"
	ActionOverride       = "OVERRIDE"
	ActionSubmit         = "SUBMIT"
	ActionReturn         = "RETURN"
	ActionApprove        = "APPROVE"
	ActionReopen         = "REOPEN"
	ActionDelete         = "DELETE"
	ActionPeriodLock     = "PERIOD_LOCK"
	ActionPeriodUnlock   = "PERIOD_UNLOCK"
	ActionSettingsUpdate = "SETTINGS_UPDATE"
)

// Period lock statuses
const (
	PeriodOpen   = "OPEN"
	PeriodLocked = "LOCKED"
)

// MinOverrideReasonLen is the shortest accepted override reason
const MinOverrideReasonLen = 20
