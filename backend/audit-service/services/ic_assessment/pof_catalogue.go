package ic_assessment

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"audit-service/models"
	"audit-service/pkg/riskclient"

	"github.com/google/uuid"
)

// The PoF catalogue lives in risk-service. When a QUANTITATIVE_POF assessment
// is created the catalogue is validated and copied into the assessment's own
// rows (ic_assessment_pofs / ic_assessment_elements), so scoring and import
// need no cross-service call and an assessment never changes when the
// catalogue is revised.

// ComponentWeights are the fixed component weights (points of 100)
var ComponentWeights = map[string]float64{"CE": 25, "RA": 25, "CA": 20, "IC": 17.5, "MA": 12.5}

var (
	pofCodeRe     = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)
	elementCodeRe = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)\.(\d+)$`)
	// OwnerLabelRe is a free owner / role label
	OwnerLabelRe = regexp.MustCompile(`^[A-Z][A-Z0-9_]{1,39}$`)
)

func micro(w float64) int64 { return int64(math.Round(w * 1e6)) }

// ValidateCatalogue checks the weight invariants (element sum = PoF, PoF sum =
// principle, principle sum = component, components 25/25/20/17.5/12.5, total
// 100) and the structure. A catalogue that fails is never copied.
func ValidateCatalogue(c *riskclient.PofCatalogue) []string {
	var issues []string
	add := func(f string, a ...interface{}) { issues = append(issues, fmt.Sprintf(f, a...)) }
	if c == nil || len(c.Components) == 0 {
		return []string{"catalogue is empty"}
	}
	if strings.TrimSpace(c.Version) == "" || len(c.Version) > 20 {
		add("catalogue version is empty or longer than 20 characters")
	}
	if len(c.Components) != len(ComponentCodes) {
		add("expected %d components, got %d", len(ComponentCodes), len(c.Components))
	}
	seen := map[string]bool{}
	var total int64
	principles := 0
	for ci, comp := range c.Components {
		want, ok := ComponentWeights[comp.Code]
		if !ok {
			add("unknown component %q", comp.Code)
		} else if micro(comp.Weight) != micro(want) {
			add("component %s weight %v, expected %v", comp.Code, comp.Weight, want)
		}
		var sumP int64
		for pi, pr := range comp.Principles {
			principles++
			meta, ok := PrincipleByCode(pr.Code)
			if !ok || meta.ComponentCode != comp.Code {
				add("principle %q is not a principle of component %s", pr.Code, comp.Code)
			}
			var sumF int64
			for fi, p := range pr.Pofs {
				if seen[p.Code] {
					add("duplicate code %s", p.Code)
				}
				seen[p.Code] = true
				if m := pofCodeRe.FindStringSubmatch(p.Code); m == nil || m[1] != fmt.Sprint(ci+1) || m[2] != fmt.Sprint(pi+1) || m[3] != fmt.Sprint(fi+1) {
					add("PoF code %q does not match its position %d.%d.%d", p.Code, ci+1, pi+1, fi+1)
				}
				if len(p.Elements) == 0 {
					add("PoF %s has no elements", p.Code)
				}
				var sumE int64
				for ei, e := range p.Elements {
					if seen[e.Code] {
						add("duplicate code %s", e.Code)
					}
					seen[e.Code] = true
					if e.Code != fmt.Sprintf("%s.%d", p.Code, ei+1) || !elementCodeRe.MatchString(e.Code) {
						add("element code %q does not match its position under %s", e.Code, p.Code)
					}
					if e.Type != models.CosoElementPresentName && e.Type != models.CosoElementFunctionName {
						add("element %s has unknown type %q", e.Code, e.Type)
					}
					if e.ScaleType != ScaleDiscrete && e.ScaleType != ScaleQuestionnaire {
						add("element %s has unknown scale %q", e.Code, e.ScaleType)
					}
					if _, err := NormalizeMethods(e.DefaultMethods); err != nil {
						add("element %s: %v", e.Code, err)
					}
					if e.Weight <= 0 {
						add("element %s weight must be positive", e.Code)
					}
					sumE += micro(e.Weight)
				}
				if sumE != micro(p.Weight) {
					add("PoF %s: element weights add up to %v, PoF weight is %v", p.Code, float64(sumE)/1e6, p.Weight)
				}
				sumF += micro(p.Weight)
			}
			if sumF != micro(pr.Weight) {
				add("principle %s: PoF weights add up to %v, principle weight is %v", pr.Code, float64(sumF)/1e6, pr.Weight)
			}
			sumP += micro(pr.Weight)
		}
		if sumP != micro(comp.Weight) {
			add("component %s: principle weights add up to %v, component weight is %v", comp.Code, float64(sumP)/1e6, comp.Weight)
		}
		total += micro(comp.Weight)
	}
	if principles != len(Principles) {
		add("expected %d principles, got %d", len(Principles), principles)
	}
	if total != 100*1e6 {
		add("component weights add up to %v, expected 100", float64(total)/1e6)
	}
	return issues
}

// MaterializeRows copies a validated catalogue into assessment rows.
// ownerOverrides replaces the default owner labels per principle code.
func MaterializeRows(assessmentID uuid.UUID, c *riskclient.PofCatalogue, ownerOverrides map[string]string, now time.Time) ([]models.ICAssessmentPoF, []models.ICAssessmentElement) {
	var pofs []models.ICAssessmentPoF
	var els []models.ICAssessmentElement
	order := 0
	for _, comp := range c.Components {
		for _, pr := range comp.Principles {
			owners := strings.Join(pr.DefaultOwnerRoles, ",")
			if o, ok := ownerOverrides[pr.Code]; ok && o != "" {
				owners = o
			}
			for _, p := range pr.Pofs {
				order++
				pofs = append(pofs, models.ICAssessmentPoF{
					AssessmentID: assessmentID, PofCode: p.Code, PrincipleCode: pr.Code, ComponentCode: comp.Code, SortOrder: order,
					NameEn: p.NameEN, NameID: p.NameID, Weight: p.Weight, DefaultOwnerRole: owners, CreatedAt: now, UpdatedAt: now,
				})
				for _, e := range p.Elements {
					m, _ := NormalizeMethods(e.DefaultMethods)
					els = append(els, models.ICAssessmentElement{
						AssessmentID: assessmentID, ElementCode: e.Code, PofCode: p.Code, SortOrder: order*100 + e.Index,
						Type: e.Type, DescriptionEn: e.DescriptionEN, DescriptionID: e.DescriptionID, Weight: e.Weight,
						DefaultMethods: m, ScaleType: e.ScaleType, RequiresIndependentAssessor: e.RequiresIndependentAssessor,
						CreatedAt: now, UpdatedAt: now,
					})
				}
			}
		}
	}
	return pofs, els
}
