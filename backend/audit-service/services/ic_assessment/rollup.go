package ic_assessment

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"audit-service/models"
)

// Roll-up rules (plan section 4.3)

// ValidatePrinciple checks one principle's status / severity / rationale
// combination (plan 4.2.3).
func ValidatePrinciple(p *models.ICAssessmentPrinciple) error {
	rationale := strings.TrimSpace(p.Rationale)
	if !IsSeverity(p.Severity) {
		return fmt.Errorf("%s: unknown severity %q", p.PrincipleCode, p.Severity)
	}
	switch p.Status {
	case PrincipleNotAssessed:
		if p.Severity != SeverityNone {
			return fmt.Errorf("%s: a principle that is not assessed has severity NONE", p.PrincipleCode)
		}
	case PrincipleYes:
		if p.Severity != SeverityNone && p.Severity != SeverityDeficiency {
			return fmt.Errorf("%s: status YES allows severity NONE or DEFICIENCY", p.PrincipleCode)
		}
	case PrinciplePartial:
		if p.Severity != SeverityDeficiency && p.Severity != SeveritySignificant {
			return fmt.Errorf("%s: status PARTIAL needs severity DEFICIENCY or SIGNIFICANT", p.PrincipleCode)
		}
	case PrincipleNo:
		if p.Severity != SeverityMajor {
			return fmt.Errorf("%s: status NO is a major deficiency (severity MAJOR)", p.PrincipleCode)
		}
	case PrincipleNotApplicable:
		if p.Severity != SeverityNone {
			return fmt.Errorf("%s: a NOT_APPLICABLE principle has severity NONE", p.PrincipleCode)
		}
	default:
		return fmt.Errorf("%s: unknown status %q", p.PrincipleCode, p.Status)
	}
	needsRationale := p.Status == PrinciplePartial || p.Status == PrincipleNo ||
		p.Status == PrincipleNotApplicable || SeverityRank(p.Severity) > 0
	if needsRationale && rationale == "" {
		return fmt.Errorf("%s: a rationale is required for status %s / severity %s", p.PrincipleCode, p.Status, p.Severity)
	}
	return nil
}

// ComponentDerivation is what the principles of one component imply
type ComponentDerivation struct {
	Status     string // YES, PARTIAL, NO, or "" while incomplete and not NO
	Severity   string
	Incomplete bool // a principle is NOT_ASSESSED
	Applicable int  // principles that are not NOT_APPLICABLE
}

// DeriveComponent rolls the principles of one component up. NOT_APPLICABLE
// principles are ignored.
func DeriveComponent(principles []models.ICAssessmentPrinciple) ComponentDerivation {
	d := ComponentDerivation{Severity: SeverityNone}
	anyNo, allYes := false, true
	for _, p := range principles {
		if p.Status == PrincipleNotApplicable {
			continue
		}
		d.Applicable++
		if p.Status == PrincipleNotAssessed {
			d.Incomplete = true
			allYes = false
			continue
		}
		if SeverityRank(p.Severity) > SeverityRank(d.Severity) {
			d.Severity = p.Severity
		}
		if p.Status == PrincipleNo || p.Severity == SeverityMajor {
			anyNo = true
		}
		if p.Status != PrincipleYes {
			allYes = false
		}
	}
	switch {
	case anyNo:
		d.Status = PrincipleNo
	case d.Incomplete:
		d.Status = ""
	case d.Applicable == 0:
		d.Status = ""
	case allYes:
		d.Status = PrincipleYes
	default:
		d.Status = PrinciplePartial
	}
	return d
}

// PrinciplesOf returns the principles of one component
func PrinciplesOf(component string, all []models.ICAssessmentPrinciple) []models.ICAssessmentPrinciple {
	var out []models.ICAssessmentPrinciple
	for _, p := range all {
		if p.ComponentCode == component {
			out = append(out, p)
		}
	}
	return out
}

// ApplyDerivation stores the derivation and recomputes is_override on c
func ApplyDerivation(c *models.ICAssessmentComponent, d ComponentDerivation) {
	c.DerivedStatus = d.Status
	c.DerivedSeverity = d.Severity
	c.IsOverride = IsOverride(c)
}

// IsOverride: a final rating that differs from the proposal, or any final
// rating where there was no proposal
func IsOverride(c *models.ICAssessmentComponent) bool {
	if c.FinalRating == nil {
		return false
	}
	return c.ProposedRating == nil || *c.ProposedRating != *c.FinalRating
}

func isFinalStatus(s string) bool {
	return s == PrincipleYes || s == PrinciplePartial || s == PrincipleNo
}

// HasFinal reports whether all three final values are set
func HasFinal(c *models.ICAssessmentComponent) bool {
	return c.FinalRating != nil && c.FinalStatus != "" && c.FinalSeverity != ""
}

// ValidateComponentFinal checks the final values that are set against the
// derivation. With requireReason the override reason rule is enforced too.
func ValidateComponentFinal(c *models.ICAssessmentComponent, d ComponentDerivation, requireReason bool) error {
	if c.FinalRating != nil && (*c.FinalRating < 1 || *c.FinalRating > 5) {
		return fmt.Errorf("%s: final_rating must be 1-5", c.ComponentCode)
	}
	if c.FinalStatus != "" && !isFinalStatus(c.FinalStatus) {
		return fmt.Errorf("%s: final_status must be YES, PARTIAL or NO", c.ComponentCode)
	}
	if c.FinalSeverity != "" && !IsSeverity(c.FinalSeverity) {
		return fmt.Errorf("%s: unknown final_severity %q", c.ComponentCode, c.FinalSeverity)
	}
	if c.FinalSeverity != "" && SeverityRank(c.FinalSeverity) < SeverityRank(d.Severity) {
		return fmt.Errorf("%s: final_severity may not be lower than the principles' %s", c.ComponentCode, d.Severity)
	}
	if c.FinalStatus == PrincipleYes && d.Status == PrincipleNo {
		return fmt.Errorf("%s: final_status may not be YES when a principle is not present and functioning", c.ComponentCode)
	}
	if c.FinalStatus == PrincipleYes && c.FinalSeverity == SeverityMajor {
		return fmt.Errorf("%s: final_status YES is not compatible with a MAJOR deficiency", c.ComponentCode)
	}
	if c.FinalRating != nil {
		if c.FinalSeverity == SeverityMajor && *c.FinalRating > 2 {
			return fmt.Errorf("%s: a MAJOR deficiency needs final_rating 1 or 2", c.ComponentCode)
		}
		if c.FinalStatus == PrincipleYes && *c.FinalRating < 3 {
			return fmt.Errorf("%s: final_status YES needs final_rating 3 or more", c.ComponentCode)
		}
	}
	if requireReason && IsOverride(c) && utf8.RuneCountInString(strings.TrimSpace(c.OverrideReason)) < MinOverrideReasonLen {
		return fmt.Errorf("%s: the final rating differs from the proposal (or there was none); an override reason of at least %d characters is required",
			c.ComponentCode, MinOverrideReasonLen)
	}
	return nil
}

// DeriveConclusion is the overall conclusion the components imply
func DeriveConclusion(components []models.ICAssessmentComponent, principles []models.ICAssessmentPrinciple) string {
	for _, p := range principles {
		if p.Status == PrincipleNotAssessed {
			return ConclusionIncomplete
		}
	}
	if len(components) != len(ComponentCodes) {
		return ConclusionIncomplete
	}
	for i := range components {
		if !HasFinal(&components[i]) {
			return ConclusionIncomplete
		}
	}
	allYes, maxSev := true, 0
	for _, c := range components {
		if c.FinalStatus == PrincipleNo || c.FinalSeverity == SeverityMajor {
			return ConclusionNotEffective
		}
		if c.FinalStatus != PrincipleYes {
			allYes = false
		}
		if r := SeverityRank(c.FinalSeverity); r > maxSev {
			maxSev = r
		}
	}
	if allYes && maxSev <= SeverityRank(SeverityDeficiency) {
		return ConclusionEffective
	}
	return ConclusionRequiresJudgment
}

// hasMajorOrNo reports any component NO, any MAJOR (component or principle)
func hasMajorOrNo(components []models.ICAssessmentComponent, principles []models.ICAssessmentPrinciple) bool {
	for _, c := range components {
		if c.FinalStatus == PrincipleNo || c.FinalSeverity == SeverityMajor {
			return true
		}
	}
	for _, p := range principles {
		if p.Severity == SeverityMajor || p.Status == PrincipleNo {
			return true
		}
	}
	return false
}

// ValidateFinalConclusion checks the auditor's final conclusion. With
// requireText the justification rule is enforced too.
func ValidateFinalConclusion(a *models.ICAssessment, components []models.ICAssessmentComponent, principles []models.ICAssessmentPrinciple, requireText bool) error {
	switch a.FinalConclusion {
	case "":
		if requireText {
			return fmt.Errorf("final_conclusion is required")
		}
		return nil
	case ConclusionEffective, ConclusionNotEffective:
	default:
		return fmt.Errorf("final_conclusion must be EFFECTIVE or NOT_EFFECTIVE")
	}
	if a.FinalConclusion == ConclusionEffective && hasMajorOrNo(components, principles) {
		return fmt.Errorf("final_conclusion EFFECTIVE is not allowed with a MAJOR deficiency or a component not present and functioning")
	}
	if requireText && NeedsJustification(a) && strings.TrimSpace(a.ConclusionJustification) == "" {
		return fmt.Errorf("conclusion_justification is required when the final conclusion differs from the derived one or the derived one requires judgment")
	}
	return nil
}

// NeedsJustification: final differs from derived, or derived requires judgment
func NeedsJustification(a *models.ICAssessment) bool {
	return a.DerivedConclusion == ConclusionRequiresJudgment || a.FinalConclusion != a.DerivedConclusion
}

// Recompute refreshes every component's derivation and is_override and the
// header's derived conclusion, in place
func Recompute(a *models.ICAssessment) {
	for i := range a.Components {
		c := &a.Components[i]
		ApplyDerivation(c, DeriveComponent(PrinciplesOf(c.ComponentCode, a.Principles)))
	}
	a.DerivedConclusion = DeriveConclusion(a.Components, a.Principles)
}

// ValidateForSubmit runs every rule; the first failure is returned
func ValidateForSubmit(a *models.ICAssessment) []string {
	var issues []string
	if a.ProposalRun < 1 {
		issues = append(issues, "run a proposal before submitting")
	}
	for i := range a.Principles {
		p := &a.Principles[i]
		if p.Status == PrincipleNotAssessed {
			issues = append(issues, p.PrincipleCode+": not assessed")
			continue
		}
		if err := ValidatePrinciple(p); err != nil {
			issues = append(issues, err.Error())
		}
	}
	for i := range a.Components {
		c := &a.Components[i]
		d := DeriveComponent(PrinciplesOf(c.ComponentCode, a.Principles))
		if d.Applicable == 0 {
			issues = append(issues, c.ComponentCode+": at least one principle must be applicable")
		}
		if !HasFinal(c) {
			issues = append(issues, c.ComponentCode+": final rating, status and severity are required")
			continue
		}
		if err := ValidateComponentFinal(c, d, true); err != nil {
			issues = append(issues, err.Error())
		}
	}
	if a.DerivedConclusion == ConclusionIncomplete {
		issues = append(issues, "the assessment is incomplete")
	}
	if err := ValidateFinalConclusion(a, a.Components, a.Principles, true); err != nil {
		issues = append(issues, err.Error())
	}
	return issues
}
