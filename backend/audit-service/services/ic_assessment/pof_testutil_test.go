package ic_assessment

import (
	"fmt"
	"math"
	"testing"
	"time"

	"audit-service/models"
	"audit-service/pkg/riskclient"

	"github.com/google/uuid"
)

// testCatalogue builds a catalogue with the real structure and weights (17
// principles, the component weights 25/25/20/17.5/12.5, the principle weights
// and PoF counts of the built-in catalogue) and two elements per PoF, split in
// thousandths so every invariant holds exactly. The wording is placeholder
// text: these tests cover arithmetic and rules, not the catalogue itself (that
// is tested in risk-service).
func testCatalogue() *riskclient.PofCatalogue {
	principleWeight := map[string]int64{
		"P01": 5500, "P02": 5000, "P03": 4500, "P04": 5000, "P05": 5000,
		"P06": 6000, "P07": 7000, "P08": 6000, "P09": 6000,
		"P10": 7000, "P11": 6500, "P12": 6500,
		"P13": 6000, "P14": 6000, "P15": 5500,
		"P16": 6500, "P17": 6000,
	}
	pofCount := map[string]int{"P01": 4, "P02": 4, "P03": 3, "P04": 4, "P05": 5, "P06": 8, "P07": 5, "P08": 4, "P09": 3,
		"P10": 6, "P11": 4, "P12": 6, "P13": 5, "P14": 4, "P15": 5, "P16": 8, "P17": 3}
	split := func(total int64, n int) []int64 {
		out := make([]int64, n)
		for i := range out {
			out[i] = total / int64(n)
			if int64(i) < total%int64(n) {
				out[i]++
			}
		}
		return out
	}
	cat := &riskclient.PofCatalogue{Framework: "COSO_2013", Version: "test.1", Name: "test", IsCurrent: true}
	for ci, code := range ComponentCodes {
		comp := riskclient.PofComponent{Code: code, Number: ci + 1, NameEN: ComponentNames[code], Weight: ComponentWeights[code]}
		pi := 0
		for _, p := range Principles {
			if p.ComponentCode != code {
				continue
			}
			pi++
			pr := riskclient.PofPrinciple{Code: p.Code, Number: p.Number, Index: pi, NameEN: p.NameEN, Weight: float64(principleWeight[p.Code]) / 1000,
				DefaultOwnerRoles: []string{"DEPARTMENT_HEAD"}}
			for fi, w := range split(principleWeight[p.Code], pofCount[p.Code]) {
				pofCode := fmt.Sprintf("%d.%d.%d", ci+1, pi, fi+1)
				pf := riskclient.Pof{Code: pofCode, Index: fi + 1, NameEN: "PoF " + pofCode + " alpha", NameID: "TKF " + pofCode + " alfa", Weight: float64(w) / 1000,
					DefaultOwnerRoles: pr.DefaultOwnerRoles}
				for ei, ew := range split(w, 2) {
					scale := ScaleDiscrete
					typ := models.CosoElementPresentName
					if ei == 1 {
						typ = models.CosoElementFunctionName
					}
					pf.Elements = append(pf.Elements, riskclient.PofElement{Code: fmt.Sprintf("%s.%d", pofCode, ei+1), Index: ei + 1, Type: typ,
						DescriptionEN: fmt.Sprintf("Element %s.%d is documented and applied", pofCode, ei+1), DescriptionID: fmt.Sprintf("Unsur %s.%d didokumentasikan dan diterapkan", pofCode, ei+1),
						Weight: float64(ew) / 1000, DefaultMethods: []string{"D"}, ScaleType: scale})
				}
				pr.Pofs = append(pr.Pofs, pf)
			}
			comp.Principles = append(comp.Principles, pr)
		}
		cat.Components = append(cat.Components, comp)
	}
	return cat
}

// testRows copies testCatalogue into assessment rows
func testRows() ([]models.ICAssessmentPoF, []models.ICAssessmentElement) {
	return MaterializeRows(uuid.New(), testCatalogue(), nil, time.Now())
}

func fp(v float64) *float64 { return &v }

func near(t *testing.T, label string, got *float64, want float64) {
	t.Helper()
	if got == nil {
		t.Fatalf("%s is nil, want %v", label, want)
	}
	if math.Abs(*got-want) > 0.00051 {
		t.Fatalf("%s = %v, want %v", label, *got, want)
	}
}
