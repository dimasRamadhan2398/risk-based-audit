package ic_assessment

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"audit-service/models"
)

// QUANTITATIVE_POF scoring (docs/coso-pof-assessment-api.md section 3 and 4).
//
// Everything here is pure: rows in, numbers out. The score is never stored as
// the source of truth; it is computed from the weights and achievements of the
// assessment's own rows every time, then the derived principle and component
// rows are synchronised for the dashboards (SyncDerivedRows).
//
//   element score   = element weight x achievement
//   PoF achievement = sum(element scores) / PoF weight   (elements win once any is filled)
//   PoF score       = PoF achievement x PoF weight        (or weight x direct value)
//   principle       = sum of PoF scores, component = sum of principle scores, total = sum of components
//
// A node that is not completely assessed has score and achievement nil, never
// 0: the partial sums are reported separately as AssessedScore/AssessedWeight.

// Assessment methods
const (
	MethodQualitative     = models.ICMethodQualitative
	MethodQuantitativePoF = models.ICMethodQuantitativePoF
)

// IsMethod reports a known assessment method
func IsMethod(m string) bool { return m == MethodQualitative || m == MethodQuantitativePoF }

// Node statuses besides YES / PARTIAL / NO
const (
	NodeNotAssessed = "NOT_ASSESSED"
	NodeInProgress  = "IN_PROGRESS"
)

// Element scales (mirrors risk-service coso_pof_elements.scale_type)
const (
	ScaleDiscrete      = "DISCRETE"
	ScaleQuestionnaire = "QUESTIONNAIRE"
)

// Value bases of a PoF
const (
	BasisNone     = "NONE"
	BasisPoF      = "POF"
	BasisElements = "ELEMENTS"
)

// Default thresholds (research doc, section 6.3)
const (
	DefaultPofYesThreshold     = 0.80
	DefaultPofPartialThreshold = 0.40
)

// Thresholds map an achievement (0..1) to YES / PARTIAL / NO
type Thresholds struct {
	Yes     float64 `json:"yes"`
	Partial float64 `json:"partial"`
}

// DefaultThresholds returns 0.80 / 0.40
func DefaultThresholds() Thresholds {
	return Thresholds{Yes: DefaultPofYesThreshold, Partial: DefaultPofPartialThreshold}
}

func round(v float64, places int) float64 {
	p := math.Pow(10, float64(places))
	return math.Round(v*p) / p
}

func round4(v float64) float64 { return round(v, 4) }

// hasAtMostDecimals reports whether v has no more than n decimal places
func hasAtMostDecimals(v float64, n int) bool {
	p := math.Pow(10, float64(n))
	return math.Abs(v*p-math.Round(v*p)) < 1e-6
}

// Validate checks 0 < partial < yes <= 1 with at most 3 decimals
func (t Thresholds) Validate() error {
	if !hasAtMostDecimals(t.Yes, 3) || !hasAtMostDecimals(t.Partial, 3) {
		return fmt.Errorf("thresholds have at most 3 decimals")
	}
	if !(t.Partial > 0 && t.Partial < t.Yes && t.Yes <= 1) {
		return fmt.Errorf("thresholds must satisfy 0 < pof_partial_threshold < pof_yes_threshold <= 1")
	}
	return nil
}

// StatusFor maps an achievement to YES, PARTIAL or NO. The comparison is made
// at 6 decimals so that 0.7999999999 from floating point sums counts as 0.8.
func (t Thresholds) StatusFor(achievement float64) string {
	a := round(achievement, 6)
	switch {
	case a >= round(t.Yes, 6):
		return PrincipleYes
	case a >= round(t.Partial, 6):
		return PrinciplePartial
	default:
		return PrincipleNo
	}
}

// Progress counts what has been assessed under a node
type Progress struct {
	PofsTotal        int `json:"pofs_total"`
	PofsAssessed     int `json:"pofs_assessed"`
	ElementsTotal    int `json:"elements_total"`
	ElementsAssessed int `json:"elements_assessed"`
}

func (p *Progress) add(o Progress) {
	p.PofsTotal += o.PofsTotal
	p.PofsAssessed += o.PofsAssessed
	p.ElementsTotal += o.ElementsTotal
	p.ElementsAssessed += o.ElementsAssessed
}

// ScoreBlock is the score of one node (total, component, principle)
type ScoreBlock struct {
	Weight         float64  `json:"weight"`
	Score          *float64 `json:"score"`
	Achievement    *float64 `json:"achievement"`
	Status         string   `json:"status"`
	AssessedWeight float64  `json:"assessed_weight"`
	AssessedScore  float64  `json:"assessed_score"`
	Progress       Progress `json:"progress"`
}

// PoFCalc is the computed PoF with its element rows
type PoFCalc struct {
	Row      *models.ICAssessmentPoF
	Elements []*models.ICAssessmentElement
	Basis    string
	// Complete: Score and Achievement are set
	Complete       bool
	Score          *float64
	Achievement    *float64
	AssessedWeight float64
	AssessedScore  float64
	Status         string
	ElementsDone   int
}

// PrincipleCalc is the computed principle
type PrincipleCalc struct {
	Principle Principle
	Pofs      []*PoFCalc
	ScoreBlock
}

// ComponentCalc is the computed component
type ComponentCalc struct {
	Code       string
	Principles []*PrincipleCalc
	ScoreBlock
}

// TreeCalc is the whole computed assessment
type TreeCalc struct {
	Components []*ComponentCalc
	Total      ScoreBlock
	Thresholds Thresholds
	byPoF      map[string]*PoFCalc
	byElement  map[string]*models.ICAssessmentElement
}

// PoF returns the computed PoF with code
func (t *TreeCalc) PoF(code string) *PoFCalc { return t.byPoF[code] }

// Element returns the element row with code
func (t *TreeCalc) Element(code string) *models.ICAssessmentElement { return t.byElement[code] }

// Principle returns the computed principle with code
func (t *TreeCalc) Principle(code string) *PrincipleCalc {
	for _, c := range t.Components {
		for _, p := range c.Principles {
			if p.Principle.Code == code {
				return p
			}
		}
	}
	return nil
}

// Component returns the computed component with code
func (t *TreeCalc) Component(code string) *ComponentCalc {
	for _, c := range t.Components {
		if c.Code == code {
			return c
		}
	}
	return nil
}

// AllPoFs lists the PoFs in catalogue order
func (t *TreeCalc) AllPoFs() []*PoFCalc {
	var out []*PoFCalc
	for _, c := range t.Components {
		for _, p := range c.Principles {
			out = append(out, p.Pofs...)
		}
	}
	return out
}

func statusOf(complete bool, assessedAny bool, achievement float64, th Thresholds) string {
	switch {
	case complete:
		return th.StatusFor(achievement)
	case assessedAny:
		return NodeInProgress
	default:
		return NodeNotAssessed
	}
}

func finish(b *ScoreBlock, th Thresholds) {
	complete := b.Progress.PofsTotal > 0 && b.Progress.PofsAssessed == b.Progress.PofsTotal
	if complete && b.Weight > 0 {
		s := round4(b.AssessedScore)
		a := round4(b.AssessedScore / b.Weight)
		b.Score, b.Achievement = &s, &a
		b.Status = th.StatusFor(b.AssessedScore / b.Weight)
	} else {
		b.Score, b.Achievement = nil, nil
		b.Status = statusOf(false, b.Progress.PofsAssessed > 0 || b.AssessedWeight > 0, 0, th)
	}
	b.AssessedScore = round4(b.AssessedScore)
	b.AssessedWeight = round4(b.AssessedWeight)
}

// calcPoF computes one PoF from its rows
func calcPoF(row *models.ICAssessmentPoF, els []*models.ICAssessmentElement, th Thresholds) *PoFCalc {
	c := &PoFCalc{Row: row, Elements: els, Basis: BasisNone}
	var doneW, doneS float64
	for _, e := range els {
		if e.Achievement != nil {
			c.ElementsDone++
			doneW += e.Weight
			doneS += e.Weight * *e.Achievement
		}
	}
	switch {
	case c.ElementsDone > 0:
		c.Basis = BasisElements
		c.AssessedWeight, c.AssessedScore = doneW, doneS
		if c.ElementsDone == len(els) {
			c.Complete = true
			c.Score = &doneS
			a := 0.0
			if row.Weight > 0 {
				a = doneS / row.Weight
			}
			c.Achievement = &a
		}
	case row.Achievement != nil:
		c.Basis = BasisPoF
		c.Complete = true
		s := row.Weight * *row.Achievement
		a := *row.Achievement
		c.Score, c.Achievement = &s, &a
		c.AssessedWeight, c.AssessedScore = row.Weight, s
	}
	if c.Complete {
		c.Status = th.StatusFor(*c.Achievement)
	} else {
		c.Status = statusOf(false, c.AssessedWeight > 0, 0, th)
	}
	return c
}

// Calculate computes the whole tree from the assessment's PoF and element rows
func Calculate(pofs []models.ICAssessmentPoF, elements []models.ICAssessmentElement, th Thresholds) *TreeCalc {
	tree := &TreeCalc{Thresholds: th, byPoF: map[string]*PoFCalc{}, byElement: map[string]*models.ICAssessmentElement{}}
	elsByPoF := map[string][]*models.ICAssessmentElement{}
	for i := range elements {
		e := &elements[i]
		elsByPoF[e.PofCode] = append(elsByPoF[e.PofCode], e)
		tree.byElement[e.ElementCode] = e
	}
	for _, list := range elsByPoF {
		sort.SliceStable(list, func(i, j int) bool {
			if list[i].SortOrder != list[j].SortOrder {
				return list[i].SortOrder < list[j].SortOrder
			}
			return list[i].ElementCode < list[j].ElementCode
		})
	}
	pofsByPrinciple := map[string][]*PoFCalc{}
	sorted := make([]*models.ICAssessmentPoF, len(pofs))
	for i := range pofs {
		sorted[i] = &pofs[i]
	}
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].SortOrder != sorted[j].SortOrder {
			return sorted[i].SortOrder < sorted[j].SortOrder
		}
		return sorted[i].PofCode < sorted[j].PofCode
	})
	for _, row := range sorted {
		pc := calcPoF(row, elsByPoF[row.PofCode], th)
		tree.byPoF[row.PofCode] = pc
		pofsByPrinciple[row.PrincipleCode] = append(pofsByPrinciple[row.PrincipleCode], pc)
	}

	for _, code := range ComponentCodes {
		comp := &ComponentCalc{Code: code}
		for _, p := range Principles {
			if p.ComponentCode != code {
				continue
			}
			pr := &PrincipleCalc{Principle: p, Pofs: pofsByPrinciple[p.Code]}
			for _, pc := range pr.Pofs {
				pr.Weight += pc.Row.Weight
				pr.AssessedWeight += pc.AssessedWeight
				pr.AssessedScore += pc.AssessedScore
				pr.Progress.PofsTotal++
				pr.Progress.ElementsTotal += len(pc.Elements)
				pr.Progress.ElementsAssessed += pc.ElementsDone
				if pc.Complete {
					pr.Progress.PofsAssessed++
				}
			}
			pr.Weight = round(pr.Weight, 6)
			finish(&pr.ScoreBlock, th)
			comp.Principles = append(comp.Principles, pr)
			comp.Weight += pr.Weight
			comp.AssessedWeight += pr.AssessedWeight
			comp.AssessedScore += pr.AssessedScore
			comp.Progress.add(pr.Progress)
		}
		comp.Weight = round(comp.Weight, 6)
		finish(&comp.ScoreBlock, th)
		tree.Components = append(tree.Components, comp)
		tree.Total.Weight += comp.Weight
		tree.Total.AssessedWeight += comp.AssessedWeight
		tree.Total.AssessedScore += comp.AssessedScore
		tree.Total.Progress.add(comp.Progress)
	}
	tree.Total.Weight = round(tree.Total.Weight, 6)
	finish(&tree.Total, th)
	return tree
}

// ---------------------------------------------------------------------------
// Value validation
// ---------------------------------------------------------------------------

// ValidateAchievement checks one achievement. Elements of scale DISCRETE take
// 0, 0.25, 0.5, 0.75 or 1; QUESTIONNAIRE elements and PoF-level values take any
// value in [0, 1] with at most 3 decimals.
func ValidateAchievement(v float64, scale string, pofLevel bool) error {
	if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
		return fmt.Errorf("must be a number between 0 and 1")
	}
	if !pofLevel && scale == ScaleDiscrete {
		if math.Abs(v*4-math.Round(v*4)) > 1e-9 {
			return fmt.Errorf("must be one of 0, 0.25, 0.5, 0.75, 1")
		}
		return nil
	}
	if !hasAtMostDecimals(v, 3) {
		return fmt.Errorf("must have at most 3 decimals")
	}
	return nil
}

// ValidMethodString checks a methods list (subset of D K W O, no duplicates)
// and returns the stored form ("D,W", in D K W O order)
func NormalizeMethods(in []string) (string, error) {
	seen := map[string]bool{}
	for _, m := range in {
		m = strings.ToUpper(strings.TrimSpace(m))
		if m != "D" && m != "K" && m != "W" && m != "O" {
			return "", fmt.Errorf("unknown method %q, use D, K, W or O", m)
		}
		if seen[m] {
			return "", fmt.Errorf("duplicate method %q", m)
		}
		seen[m] = true
	}
	var out []string
	for _, m := range []string{"D", "K", "W", "O"} {
		if seen[m] {
			out = append(out, m)
		}
	}
	return strings.Join(out, ","), nil
}

// SplitMethods turns the stored form into a list
func SplitMethods(s string) []string {
	out := []string{}
	for _, m := range strings.Split(s, ",") {
		if m = strings.TrimSpace(m); m != "" {
			out = append(out, m)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Derived principle / component rows and the conclusion
// ---------------------------------------------------------------------------

func statusRank(s string) int {
	switch s {
	case PrincipleYes:
		return 2
	case PrinciplePartial:
		return 1
	}
	return 0
}

func worseStatus(a, b string) string {
	if statusRank(a) <= statusRank(b) {
		return a
	}
	return b
}

// RatingFromAchievement uses the RCM effectiveness bands: >= 90 percent 5,
// 80 4, 70 3, 60 2, below 1.
func RatingFromAchievement(a float64) int {
	p := round(a*100, 6)
	switch {
	case p >= 90:
		return 5
	case p >= 80:
		return 4
	case p >= 70:
		return 3
	case p >= 60:
		return 2
	}
	return 1
}

func principleSeverity(status string, achievement float64, th Thresholds) string {
	switch status {
	case PrincipleYes:
		return SeverityNone
	case PrinciplePartial:
		if round(achievement, 6) >= round((th.Yes+th.Partial)/2, 6) {
			return SeverityDeficiency
		}
		return SeveritySignificant
	case PrincipleNo:
		return SeverityMajor
	}
	return SeverityNone
}

// AutoRationalePrefix starts the rationale the server writes on a derived
// principle row
const AutoRationalePrefix = "Derived from the quantitative PoF assessment"

// SyncDerivedRows writes the principle statuses and the component final
// values implied by the PoF scores onto a's rows, then the derived conclusion,
// so the dashboards and /summary read a QUANTITATIVE_POF assessment like any
// other. Nodes that are not fully assessed are left NOT_ASSESSED / empty.
func SyncDerivedRows(a *models.ICAssessment, tree *TreeCalc) {
	th := tree.Thresholds
	for i := range a.Principles {
		p := &a.Principles[i]
		pc := tree.Principle(p.PrincipleCode)
		if pc == nil || pc.Achievement == nil {
			p.Status, p.Severity, p.Rationale = PrincipleNotAssessed, SeverityNone, ""
			continue
		}
		p.Status = th.StatusFor(*pc.Achievement)
		p.Severity = principleSeverity(p.Status, *pc.Achievement, th)
		p.Rationale = fmt.Sprintf("%s: %.2f%% of the principle weight (%s).", AutoRationalePrefix, *pc.Achievement*100, p.Status)
	}
	for i := range a.Components {
		c := &a.Components[i]
		d := DeriveComponent(PrinciplesOf(c.ComponentCode, a.Principles))
		ApplyDerivation(c, d)
		cc := tree.Component(c.ComponentCode)
		c.ProposedRating, c.ProposedStatus, c.ProposedSeverity = nil, "", ""
		if cc == nil || cc.Achievement == nil || d.Incomplete || d.Status == "" {
			c.FinalRating, c.FinalStatus, c.FinalSeverity = nil, "", ""
			c.IsOverride = false
			continue
		}
		status := worseStatus(th.StatusFor(*cc.Achievement), d.Status)
		sev := d.Severity
		switch status {
		case PrincipleNo:
			sev = SeverityMajor
		case PrinciplePartial:
			if SeverityRank(sev) < SeverityRank(SeverityDeficiency) {
				sev = SeverityDeficiency
			}
		}
		rating := RatingFromAchievement(*cc.Achievement)
		if status == PrincipleNo && rating > 2 {
			rating = 2
		}
		if status == PrincipleYes && rating < 3 {
			rating = 3
		}
		c.FinalRating, c.FinalStatus, c.FinalSeverity = &rating, status, sev
		c.OverrideReason = ""
		c.IsOverride = false
	}
	a.DerivedConclusion = DeriveConclusion(a.Components, a.Principles)
}

// ValidateQuantForSubmit lists what blocks a submit of a QUANTITATIVE_POF
// assessment (empty when it would pass)
func ValidateQuantForSubmit(a *models.ICAssessment, tree *TreeCalc) []string {
	var issues []string
	if tree.Total.Progress.PofsTotal == 0 {
		issues = append(issues, "the assessment has no points of focus")
	}
	for _, pc := range tree.AllPoFs() {
		code := pc.Row.PofCode
		if !pc.Complete {
			switch pc.Basis {
			case BasisElements:
				issues = append(issues, fmt.Sprintf("%s: %d of %d elements are not assessed", code, len(pc.Elements)-pc.ElementsDone, len(pc.Elements)))
			default:
				issues = append(issues, code+": not assessed")
			}
			continue
		}
		if pc.Basis == BasisPoF && *pc.Achievement < 1 && strings.TrimSpace(pc.Row.Explanation) == "" {
			issues = append(issues, code+": an explanation is required when the achievement is below 1")
		}
		if pc.Basis == BasisElements {
			for _, e := range pc.Elements {
				if e.Achievement != nil && *e.Achievement < 1 && strings.TrimSpace(e.Explanation) == "" {
					issues = append(issues, e.ElementCode+": an explanation is required when the achievement is below 1")
				}
			}
		}
	}
	if a.DerivedConclusion == ConclusionIncomplete && len(issues) == 0 {
		issues = append(issues, "the assessment is incomplete")
	}
	if err := ValidateFinalConclusion(a, a.Components, a.Principles, true); err != nil {
		issues = append(issues, err.Error())
	}
	return issues
}

// ExplanationNeeded reports whether an assessed value below 1 has no explanation
func ExplanationNeeded(achievement *float64, explanation string) bool {
	return achievement != nil && *achievement < 1 && strings.TrimSpace(explanation) == ""
}
