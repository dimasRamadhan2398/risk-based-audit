package services

import (
	"regexp"
	"strings"
	"testing"

	"risk-service/models"
	"risk-service/pkg/sqlitetest"
)

func TestBuiltInPofCatalogueInvariants(t *testing.T) {
	cat := BuildCosoPofCatalogue()
	if issues := ValidateCosoPofCatalogue(cat); len(issues) > 0 {
		t.Fatalf("built-in catalogue is invalid:\n%s", strings.Join(issues, "\n"))
	}
	pofs, els := CountPofElements(cat)
	if pofs < 70 || pofs > 90 {
		t.Errorf("expected roughly 80 points of focus, got %d", pofs)
	}
	if els < 150 {
		t.Errorf("expected at least 150 elements, got %d", els)
	}
	var total int64
	for _, c := range cat.Components {
		total += milli(c.Weight)
	}
	if total != 100000 {
		t.Errorf("total weight = %v", fromMilli(total))
	}
}

func TestComponentWeightsAre25_25_20_17_5_12_5(t *testing.T) {
	want := map[string]float64{"CE": 25, "RA": 25, "CA": 20, "IC": 17.5, "MA": 12.5}
	for _, c := range BuildCosoPofCatalogue().Components {
		if c.Weight != want[c.Code] {
			t.Errorf("%s weight %v, want %v", c.Code, c.Weight, want[c.Code])
		}
	}
}

func TestPofCatalogueNumberingAndOwners(t *testing.T) {
	cat := BuildCosoPofCatalogue()
	re := regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+$`)
	principles := 0
	independent := 0
	for _, c := range cat.Components {
		for _, p := range c.Principles {
			principles++
			if len(p.DefaultOwnerRoles) == 0 {
				t.Errorf("%s has no default owner", p.Code)
			}
			for _, f := range p.Pofs {
				for _, e := range f.Elements {
					if !re.MatchString(e.Code) || !strings.HasPrefix(e.Code, f.Code+".") {
						t.Errorf("bad element code %s under %s", e.Code, f.Code)
					}
					if e.RequiresIndependentAssessor {
						independent++
						if c.Code != "MA" {
							t.Errorf("%s is flagged independent but is not monitoring", e.Code)
						}
					}
				}
			}
		}
	}
	if principles != 17 {
		t.Errorf("principles = %d", principles)
	}
	if independent == 0 {
		t.Error("the internal audit function elements must be flagged requires_independent_assessor")
	}
	// Default owners from the decision list
	owners := map[string]string{}
	for _, c := range cat.Components {
		for _, p := range c.Principles {
			owners[p.Code] = strings.Join(p.DefaultOwnerRoles, ",")
		}
	}
	for code, want := range map[string]string{
		"P01": "EXECUTIVE", "P02": "AUDIT_COMMITTEE,CORPORATE_SECRETARY", "P03": "EXECUTIVE", "P04": "HEAD_OF_HR", "P05": "EXECUTIVE",
		"P06": "RISK_MANAGEMENT,DEPARTMENT_HEAD", "P09": "RISK_MANAGEMENT,DEPARTMENT_HEAD",
		"P10": "DEPARTMENT_HEAD", "P11": "HEAD_OF_IT", "P12": "DEPARTMENT_HEAD",
		"P13": "DEPARTMENT_HEAD,HEAD_OF_IT", "P14": "DEPARTMENT_HEAD,HEAD_OF_IT", "P15": "CORPORATE_COMMUNICATION,CORPORATE_SECRETARY",
		"P16": "MANAGEMENT", "P17": "MANAGEMENT",
	} {
		if owners[code] != want {
			t.Errorf("%s owners = %q, want %q", code, owners[code], want)
		}
	}
}

// The catalogue must not carry third-party identifiers (regulation numbers,
// company names) in its wording.
func TestPofCatalogueHasNoThirdPartyMarkers(t *testing.T) {
	banned := regexp.MustCompile(`(?i)perdir|\bsk[- ]?\d|icon\s*plus|\bpln\b|kke\b|\bno\.\s*\d`)
	for _, c := range BuildCosoPofCatalogue().Components {
		for _, p := range c.Principles {
			for _, f := range p.Pofs {
				texts := []string{f.NameEN, f.NameID}
				for _, e := range f.Elements {
					texts = append(texts, e.DescriptionEN, e.DescriptionID)
				}
				for _, s := range texts {
					if banned.MatchString(s) {
						t.Errorf("%s contains a third-party marker: %q", f.Code, s)
					}
				}
			}
		}
	}
}

func TestValidateRejectsBrokenSums(t *testing.T) {
	cases := map[string]func(c *PofCatalogue){
		"element sum":   func(c *PofCatalogue) { c.Components[0].Principles[0].Pofs[0].Elements[0].Weight += 0.001 },
		"pof sum":       func(c *PofCatalogue) { c.Components[1].Principles[0].Pofs[0].Weight += 0.5 },
		"principle sum": func(c *PofCatalogue) { c.Components[2].Principles[1].Weight -= 1 },
		"component":     func(c *PofCatalogue) { c.Components[3].Weight = 20 },
		"bad method": func(c *PofCatalogue) {
			c.Components[0].Principles[0].Pofs[0].Elements[0].DefaultMethods = []string{"X"}
		},
		"bad code":  func(c *PofCatalogue) { c.Components[0].Principles[0].Pofs[0].Elements[0].Code = "9.9.9.9" },
		"bad type":  func(c *PofCatalogue) { c.Components[0].Principles[0].Pofs[0].Elements[0].Type = "OTHER" },
		"empty pof": func(c *PofCatalogue) { c.Components[4].Principles[1].Pofs[0].Elements = nil },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			c := BuildCosoPofCatalogue()
			mutate(c)
			if len(ValidateCosoPofCatalogue(c)) == 0 {
				t.Fatal("expected an issue")
			}
		})
	}
}

func TestSplitMilliExact(t *testing.T) {
	for _, tc := range []struct {
		total int64
		n     int
	}{{5500, 4}, {4500, 3}, {1375, 3}, {7, 5}, {100000, 7}} {
		var sum int64
		for _, p := range splitMilli(tc.total, tc.n) {
			sum += p
		}
		if sum != tc.total {
			t.Errorf("split %d/%d sums to %d", tc.total, tc.n, sum)
		}
	}
}

func TestSeedPofCatalogueIsIdempotentAndNeverOverwrites(t *testing.T) {
	db := sqlitetest.Open(t, models.CosoPofModels()...)
	if err := SeedCosoPofCatalogue(db); err != nil {
		t.Fatal(err)
	}
	count := func(model interface{}) int64 {
		var n int64
		db.Model(model).Count(&n)
		return n
	}
	pofs, els := count(&models.CosoPointOfFocus{}), count(&models.CosoPofElement{})
	w := count(&models.CosoCatalogWeight{})
	if pofs == 0 || els == 0 || w != 22 {
		t.Fatalf("seed counts pofs=%d els=%d weights=%d", pofs, els, w)
	}

	// An edited row stays edited; a deleted row comes back; nothing doubles
	db.Model(&models.CosoPointOfFocus{}).Where("code = ?", "1.1.1").Update("name_en", "edited by hand")
	db.Where("code = ?", "1.1.1.1").Delete(&models.CosoPofElement{})
	if err := SeedCosoPofCatalogue(db); err != nil {
		t.Fatal(err)
	}
	var p models.CosoPointOfFocus
	db.Where("code = ?", "1.1.1").Take(&p)
	if p.NameEn != "edited by hand" {
		t.Errorf("seed overwrote an existing row: %q", p.NameEn)
	}
	if count(&models.CosoPointOfFocus{}) != pofs || count(&models.CosoPofElement{}) != els || count(&models.CosoCatalogWeight{}) != w {
		t.Error("second seed changed the row counts")
	}
	var v []models.CosoCatalogVersion
	db.Find(&v)
	if len(v) != 1 || !v[0].IsCurrent || v[0].Version != CosoPofCatalogVersion {
		t.Errorf("versions = %+v", v)
	}
}

func TestLoadPofCatalogueRoundTripKeepsInvariants(t *testing.T) {
	db := sqlitetest.Open(t, models.CosoPofModels()...)
	if err := SeedCosoPofCatalogue(db); err != nil {
		t.Fatal(err)
	}
	cat, err := LoadCosoPofCatalogue(db, models.CosoPofFramework, "")
	if err != nil {
		t.Fatal(err)
	}
	if issues := ValidateCosoPofCatalogue(cat); len(issues) > 0 {
		t.Fatalf("stored catalogue invalid:\n%s", strings.Join(issues, "\n"))
	}
	built := BuildCosoPofCatalogue()
	bp, be := CountPofElements(built)
	lp, le := CountPofElements(cat)
	if bp != lp || be != le {
		t.Errorf("loaded %d/%d, built %d/%d", lp, le, bp, be)
	}
	if _, err := LoadCosoPofCatalogue(db, models.CosoPofFramework, "9999.9"); err != ErrCatalogNotFound {
		t.Errorf("unknown version err = %v", err)
	}
}
