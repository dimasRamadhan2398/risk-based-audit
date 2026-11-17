package ic_assessment

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sort"
	"strings"
)

// Signal keys (plan 5.3)
const (
	SignalCEIntegrityPact      = "ce.integrity_pact"
	SignalCEUnitAccountability = "ce.unit_accountability"
	SignalCEAuditCharter       = "ce.audit_charter"
	SignalRAQuarterlyCoverage  = "ra.quarterly_coverage"
	SignalRAAboveAppetite      = "ra.above_appetite_mitigated"
	SignalCARCMEffectiveness   = "ca.rcm_effectiveness"
	SignalICLHATimeliness      = "ic.lha_timeliness"
	SignalICManagementResponse = "ic.management_response"
	SignalMAATRCompletedOnTime = "ma.atr_completed_on_time"
	SignalMAATRNotOverdue      = "ma.atr_not_overdue"

	// Context evidence: shown, never scored (weight 0)
	ContextExecutiveSummary = "ctx.executive_summary"
	ContextThresholds       = "ctx.thresholds"
)

// SignalDef is a scored signal and its fixed weight (Q10: weights are fixed
// in code)
type SignalDef struct {
	Key           string
	Component     string
	Weight        float64
	PrincipleHint string
}

// SignalDefs are the scored signals, grouped by component
var SignalDefs = []SignalDef{
	{SignalCEIntegrityPact, ComponentCE, 60, "P01"},
	{SignalCEUnitAccountability, ComponentCE, 20, "P03,P05"},
	{SignalCEAuditCharter, ComponentCE, 20, "P02"},
	{SignalRAQuarterlyCoverage, ComponentRA, 50, "P07,P09"},
	{SignalRAAboveAppetite, ComponentRA, 50, "P07"},
	{SignalCARCMEffectiveness, ComponentCA, 100, "P10,P11,P12"},
	{SignalICLHATimeliness, ComponentIC, 50, "P14"},
	{SignalICManagementResponse, ComponentIC, 50, "P14,P17"},
	{SignalMAATRCompletedOnTime, ComponentMA, 50, "P16,P17"},
	{SignalMAATRNotOverdue, ComponentMA, 50, "P17"},
}

// SignalDefByKey returns the definition of a scored signal
func SignalDefByKey(key string) (SignalDef, bool) {
	for _, d := range SignalDefs {
		if d.Key == key {
			return d, true
		}
	}
	return SignalDef{}, false
}

// MinCoveragePct: below this share of the component weight with a value, the
// component gets no score (NULL, "Insufficient data"), never 0
const MinCoveragePct = 50.0

// Signal is one collected value (an evidence row before it is stored)
type Signal struct {
	Key           string
	Component     string
	PrincipleHint string
	SourceService string
	SourceRef     string
	State         string
	ValuePct      *float64
	Numerator     *int
	Denominator   *int
	Weight        float64
	ScopeNote     string
	Detail        map[string]interface{}
}

// HasValue reports whether the signal contributes to the score
func (s Signal) HasValue() bool { return s.State == StateOK && s.ValuePct != nil }

// ComponentScore is the proposal for one component
type ComponentScore struct {
	ScorePct    *float64
	Rating      *int
	Status      string
	Severity    string
	CoveragePct float64
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

func pct(num, den int) *float64 {
	if den <= 0 {
		return nil
	}
	v := round2(float64(num) / float64(den) * 100)
	return &v
}

func intPtr(v int) *int { return &v }

// ScoreComponent combines the signals of one component: the weighted mean of
// the signals with a value, or NULL when they carry less than MinCoveragePct
// of the component's total (fixed) weight.
func ScoreComponent(component string, signals []Signal) ComponentScore {
	var total, covered, sum float64
	for _, d := range SignalDefs {
		if d.Component == component {
			total += d.Weight
		}
	}
	for _, s := range signals {
		if s.Component != component || s.Weight <= 0 || !s.HasValue() {
			continue
		}
		covered += s.Weight
		sum += s.Weight * *s.ValuePct
	}
	out := ComponentScore{}
	if total > 0 {
		out.CoveragePct = round2(covered / total * 100)
	}
	if total == 0 || covered == 0 || out.CoveragePct < MinCoveragePct {
		return out
	}
	score := round2(sum / covered)
	rating := RatingForScore(score)
	out.ScorePct = &score
	out.Rating = &rating
	out.Status, out.Severity = SuggestedStatus(rating)
	return out
}

// RatingForScore maps a 0-100 score to the RCM bands
// (GetEffectivenessRating): >=90 5, >=80 4, >=70 3, >=60 2, else 1
func RatingForScore(score float64) int {
	switch {
	case score >= 90:
		return 5
	case score >= 80:
		return 4
	case score >= 70:
		return 3
	case score >= 60:
		return 2
	}
	return 1
}

// SuggestedStatus: rating 4-5 YES, 3 PARTIAL, 1-2 NO; severity 5 NONE, 3-4
// DEFICIENCY, 1-2 SIGNIFICANT. The system never proposes MAJOR.
func SuggestedStatus(rating int) (status, severity string) {
	switch {
	case rating >= 4:
		status = PrincipleYes
	case rating == 3:
		status = PrinciplePartial
	default:
		status = PrincipleNo
	}
	switch {
	case rating == 5:
		severity = SeverityNone
	case rating >= 3:
		severity = SeverityDeficiency
	default:
		severity = SeveritySignificant
	}
	return status, severity
}

// SignalFingerprint is the comparable part of a scored signal
type SignalFingerprint struct {
	Key         string   `json:"signal_key"`
	State       string   `json:"state"`
	ValuePct    *float64 `json:"value_pct"`
	Numerator   *int     `json:"numerator"`
	Denominator *int     `json:"denominator"`
}

func fingerprintOf(s Signal) SignalFingerprint {
	return SignalFingerprint{Key: s.Key, State: s.State, ValuePct: s.ValuePct, Numerator: s.Numerator, Denominator: s.Denominator}
}

func (f SignalFingerprint) String() string {
	v, n, d := "-", "-", "-"
	if f.ValuePct != nil {
		v = fmt.Sprintf("%.2f", *f.ValuePct)
	}
	if f.Numerator != nil {
		n = fmt.Sprint(*f.Numerator)
	}
	if f.Denominator != nil {
		d = fmt.Sprint(*f.Denominator)
	}
	return strings.Join([]string{f.Key, f.State, v, n, d}, "|")
}

// HashSignals is the SHA-256 of the scored signals' inputs (key, state,
// value, numerator, denominator), independent of order. Context rows are
// not part of it.
func HashSignals(signals []Signal) string {
	var lines []string
	for _, s := range signals {
		if _, scored := SignalDefByKey(s.Key); !scored {
			continue
		}
		lines = append(lines, fingerprintOf(s).String())
	}
	sort.Strings(lines)
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return hex.EncodeToString(sum[:])
}
