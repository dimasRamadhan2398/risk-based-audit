package services

import (
	"errors"
	"fmt"
	"log"
	"math"
	"regexp"
	"sort"
	"strings"

	"risk-service/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// COSO 2013 PoF catalogue: structure, builder, invariants, seed and loader
// (docs/coso-pof-assessment-api.md). The wording lives in
// coso_pof_catalogue_data.go.

// CosoPofCatalogVersion is the version of the built-in catalogue.
const CosoPofCatalogVersion = "2013.1"

// CosoPofCatalogName names the built-in catalogue.
const CosoPofCatalogName = "AuditSphere COSO 2013 PoF catalogue (original wording)"

// CosoComponentWeights are the fixed component weights of the catalogue
// (thousandths of a point); they add up to 100 points.
var CosoComponentWeights = map[string]int64{"CE": 25000, "RA": 25000, "CA": 20000, "IC": 17500, "MA": 12500}

// ValidMethods are the evaluation methods: Document, Questionnaire (Kuesioner),
// Interview (Wawancara), Observation.
var ValidMethods = map[string]bool{"D": true, "K": true, "W": true, "O": true}

// PofElement is one fulfilment element (unsur).
type PofElement struct {
	Code                        string   `json:"code"`
	Index                       int      `json:"index"`
	Type                        string   `json:"type"`
	DescriptionEN               string   `json:"description_en"`
	DescriptionID               string   `json:"description_id"`
	Weight                      float64  `json:"weight"`
	DefaultMethods              []string `json:"default_methods"`
	ScaleType                   string   `json:"scale_type"`
	RequiresIndependentAssessor bool     `json:"requires_independent_assessor"`
}

// Pof is one point of focus.
type Pof struct {
	Code              string       `json:"code"`
	Index             int          `json:"index"`
	NameEN            string       `json:"name_en"`
	NameID            string       `json:"name_id"`
	Weight            float64      `json:"weight"`
	DefaultOwnerRoles []string     `json:"default_owner_roles"`
	Elements          []PofElement `json:"elements"`
}

// PofPrinciple is one of the 17 principles with its weight and PoFs.
type PofPrinciple struct {
	Code              string   `json:"code"`
	Number            int      `json:"number"`
	Index             int      `json:"index"`
	NameEN            string   `json:"name_en"`
	NameID            string   `json:"name_id"`
	Weight            float64  `json:"weight"`
	DefaultOwnerRoles []string `json:"default_owner_roles"`
	Pofs              []Pof    `json:"pofs"`
}

// PofComponent is one of the 5 components.
type PofComponent struct {
	Code       string         `json:"code"`
	Number     int            `json:"number"`
	NameEN     string         `json:"name_en"`
	NameID     string         `json:"name_id"`
	Weight     float64        `json:"weight"`
	Principles []PofPrinciple `json:"principles"`
}

// PofCatalogue is a whole catalogue version.
type PofCatalogue struct {
	Framework  string         `json:"framework"`
	Version    string         `json:"version"`
	Name       string         `json:"name"`
	IsCurrent  bool           `json:"is_current"`
	Components []PofComponent `json:"components"`
}

func milli(w float64) int64     { return int64(math.Round(w * 1000)) }
func fromMilli(m int64) float64 { return float64(m) / 1000 }

// splitMilli splits total over n parts as evenly as possible; the first
// total%n parts get one extra thousandth, so the parts add up exactly.
func splitMilli(total int64, n int) []int64 {
	out := make([]int64, n)
	if n == 0 {
		return out
	}
	base, rem := total/int64(n), total%int64(n)
	for i := range out {
		out[i] = base
		if int64(i) < rem {
			out[i]++
		}
	}
	return out
}

func splitMethods(s string) []string {
	var out []string
	for _, m := range strings.Split(s, ",") {
		if m = strings.TrimSpace(m); m != "" {
			out = append(out, m)
		}
	}
	return out
}

func splitOwners(s string) []string { return splitMethods(s) }

// BuildCosoPofCatalogue builds the built-in catalogue with derived weights.
func BuildCosoPofCatalogue() *PofCatalogue {
	cat := &PofCatalogue{Framework: models.CosoPofFramework, Version: CosoPofCatalogVersion, Name: CosoPofCatalogName}
	names := map[string]string{}
	numbers := map[string]int{}
	for _, comp := range models.Coso2013Catalogue {
		for _, p := range comp.Principles {
			names[p.Code] = p.NameEn
			numbers[p.Code] = p.Number
		}
	}
	compNames := map[string]string{}
	for _, comp := range models.Coso2013Catalogue {
		compNames[comp.Code] = comp.NameEn
	}
	for ci, cd := range cosoPofComponentDefs {
		comp := PofComponent{Code: cd.Code, Number: ci + 1, NameEN: compNames[cd.Code], NameID: cd.NameID,
			Weight: fromMilli(CosoComponentWeights[cd.Code])}
		for pi, pd := range cd.Principles {
			prin := PofPrinciple{Code: pd.Code, Number: numbers[pd.Code], Index: pi + 1, NameEN: names[pd.Code], NameID: pd.NameID,
				Weight: fromMilli(pd.Weight), DefaultOwnerRoles: splitOwners(pd.Owners)}
			pofW := splitMilli(pd.Weight, len(pd.Pofs))
			for fi, fd := range pd.Pofs {
				pofCode := fmt.Sprintf("%d.%d.%d", ci+1, pi+1, fi+1)
				p := Pof{Code: pofCode, Index: fi + 1, NameEN: fd.NameEN, NameID: fd.NameID, Weight: fromMilli(pofW[fi]),
					DefaultOwnerRoles: prin.DefaultOwnerRoles}
				elW := splitMilli(pofW[fi], len(fd.Elements))
				for ei, ed := range fd.Elements {
					scale := models.CosoScaleDiscrete
					if ed.Methods == "K" {
						scale = models.CosoScaleQuestionnaire
					}
					p.Elements = append(p.Elements, PofElement{
						Code: fmt.Sprintf("%s.%d", pofCode, ei+1), Index: ei + 1, Type: ed.Type,
						DescriptionEN: ed.EN, DescriptionID: ed.ID, Weight: fromMilli(elW[ei]),
						DefaultMethods: splitMethods(ed.Methods), ScaleType: scale, RequiresIndependentAssessor: ed.Indep,
					})
				}
				prin.Pofs = append(prin.Pofs, p)
			}
			comp.Principles = append(comp.Principles, prin)
		}
		cat.Components = append(cat.Components, comp)
	}
	return cat
}

var (
	pofCodeRe     = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)$`)
	elementCodeRe = regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)\.(\d+)$`)
)

// ValidateCosoPofCatalogue checks the weight invariants (sum of element
// weights = PoF weight, sum of PoF = principle, sum of principle = component,
// component weights 25/25/20/17.5/12.5, total 100) and the structure (codes,
// types, methods, scales, no empty PoF). It returns every problem found.
func ValidateCosoPofCatalogue(cat *PofCatalogue) []string {
	var issues []string
	add := func(f string, a ...interface{}) { issues = append(issues, fmt.Sprintf(f, a...)) }
	if cat == nil {
		return []string{"catalogue is nil"}
	}
	if strings.TrimSpace(cat.Version) == "" {
		add("catalogue version is empty")
	}
	if len(cat.Components) != len(CosoComponentWeights) {
		add("expected %d components, got %d", len(CosoComponentWeights), len(cat.Components))
	}
	seen := map[string]bool{}
	dup := func(code string) {
		if seen[code] {
			add("duplicate code %s", code)
		}
		seen[code] = true
	}
	var total int64
	principles := 0
	for ci, comp := range cat.Components {
		want, known := CosoComponentWeights[comp.Code]
		if !known {
			add("unknown component %q", comp.Code)
		} else if milli(comp.Weight) != want {
			add("component %s weight %v, expected %v", comp.Code, comp.Weight, fromMilli(want))
		}
		if comp.Number != ci+1 {
			add("component %s number %d, expected %d", comp.Code, comp.Number, ci+1)
		}
		var sumP int64
		for pi, prin := range comp.Principles {
			principles++
			dup(prin.Code)
			if prin.Index != pi+1 {
				add("principle %s index %d, expected %d", prin.Code, prin.Index, pi+1)
			}
			if len(prin.DefaultOwnerRoles) == 0 {
				add("principle %s has no default owner role", prin.Code)
			}
			var sumF int64
			for fi, p := range prin.Pofs {
				dup(p.Code)
				if m := pofCodeRe.FindStringSubmatch(p.Code); m == nil || m[1] != fmt.Sprint(ci+1) || m[2] != fmt.Sprint(pi+1) || m[3] != fmt.Sprint(fi+1) {
					add("PoF code %q does not match its position %d.%d.%d", p.Code, ci+1, pi+1, fi+1)
				}
				if strings.TrimSpace(p.NameEN) == "" || strings.TrimSpace(p.NameID) == "" {
					add("PoF %s needs an English and an Indonesian name", p.Code)
				}
				if len(p.Elements) == 0 {
					add("PoF %s has no elements", p.Code)
				}
				var sumE int64
				for ei, e := range p.Elements {
					dup(e.Code)
					if m := elementCodeRe.FindStringSubmatch(e.Code); m == nil || e.Code != fmt.Sprintf("%s.%d", p.Code, ei+1) {
						add("element code %q does not match its position under %s", e.Code, p.Code)
					}
					if e.Type != models.CosoElementPresent && e.Type != models.CosoElementFunction {
						add("element %s has unknown type %q", e.Code, e.Type)
					}
					if e.ScaleType != models.CosoScaleDiscrete && e.ScaleType != models.CosoScaleQuestionnaire {
						add("element %s has unknown scale %q", e.Code, e.ScaleType)
					}
					if len(e.DefaultMethods) == 0 {
						add("element %s has no default evaluation method", e.Code)
					}
					for _, m := range e.DefaultMethods {
						if !ValidMethods[m] {
							add("element %s has unknown method %q", e.Code, m)
						}
					}
					if strings.TrimSpace(e.DescriptionEN) == "" || strings.TrimSpace(e.DescriptionID) == "" {
						add("element %s needs an English and an Indonesian description", e.Code)
					}
					if e.Weight <= 0 {
						add("element %s weight must be positive", e.Code)
					}
					sumE += milli(e.Weight)
				}
				if sumE != milli(p.Weight) {
					add("PoF %s: element weights add up to %v, PoF weight is %v", p.Code, fromMilli(sumE), p.Weight)
				}
				sumF += milli(p.Weight)
			}
			if sumF != milli(prin.Weight) {
				add("principle %s: PoF weights add up to %v, principle weight is %v", prin.Code, fromMilli(sumF), prin.Weight)
			}
			sumP += milli(prin.Weight)
		}
		if sumP != milli(comp.Weight) {
			add("component %s: principle weights add up to %v, component weight is %v", comp.Code, fromMilli(sumP), comp.Weight)
		}
		total += milli(comp.Weight)
	}
	if principles != 17 {
		add("expected 17 principles, got %d", principles)
	}
	if total != 100000 {
		add("component weights add up to %v, expected 100", fromMilli(total))
	}
	return issues
}

// CountPofElements returns the number of points of focus and elements.
func CountPofElements(cat *PofCatalogue) (pofs, elements int) {
	for _, c := range cat.Components {
		for _, p := range c.Principles {
			pofs += len(p.Pofs)
			for _, f := range p.Pofs {
				elements += len(f.Elements)
			}
		}
	}
	return
}

// SeedCosoPofCatalogue inserts the built-in catalogue version when missing.
// Insert-if-missing like SeedCosoCatalogue: a row is looked up by its unique
// key and created only when absent; an existing row (edited by hand or from an
// older release) is never updated or deleted. The catalogue is validated first
// and is refused when an invariant fails, so a bad edit of the data file can
// never reach a database. The version becomes current only when the framework
// has no current version yet.
func SeedCosoPofCatalogue(db *gorm.DB) error {
	cat := BuildCosoPofCatalogue()
	if issues := ValidateCosoPofCatalogue(cat); len(issues) > 0 {
		return fmt.Errorf("built-in COSO PoF catalogue is invalid: %s", strings.Join(issues, "; "))
	}
	return db.Transaction(func(tx *gorm.DB) error {
		fw, ver := cat.Framework, cat.Version
		inserted := 0

		var v models.CosoCatalogVersion
		err := tx.Where("framework = ? AND version = ?", fw, ver).Take(&v).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			var current int64
			if err := tx.Model(&models.CosoCatalogVersion{}).Where("framework = ? AND is_current = ?", fw, true).Count(&current).Error; err != nil {
				return err
			}
			res := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.CosoCatalogVersion{
				Framework: fw, Version: ver, Name: cat.Name, IsCurrent: current == 0})
			if res.Error != nil {
				return fmt.Errorf("seed coso catalogue version: %w", res.Error)
			}
			inserted += int(res.RowsAffected)
		} else if err != nil {
			return err
		}

		existingW := map[string]bool{}
		var ws []models.CosoCatalogWeight
		if err := tx.Where("framework = ? AND version = ?", fw, ver).Find(&ws).Error; err != nil {
			return err
		}
		for _, w := range ws {
			existingW[w.Level+"/"+w.Code] = true
		}
		var newWeights []models.CosoCatalogWeight
		var newPofs []models.CosoPointOfFocus
		var newEls []models.CosoPofElement
		existingP := map[string]bool{}
		var ps []models.CosoPointOfFocus
		if err := tx.Select("code").Where("framework = ? AND catalog_version = ?", fw, ver).Find(&ps).Error; err != nil {
			return err
		}
		for _, p := range ps {
			existingP[p.Code] = true
		}
		existingE := map[string]bool{}
		var es []models.CosoPofElement
		if err := tx.Select("code").Where("framework = ? AND catalog_version = ?", fw, ver).Find(&es).Error; err != nil {
			return err
		}
		for _, e := range es {
			existingE[e.Code] = true
		}
		order := 0
		for _, comp := range cat.Components {
			if k := models.CosoWeightLevelComponent + "/" + comp.Code; !existingW[k] {
				newWeights = append(newWeights, models.CosoCatalogWeight{Framework: fw, Version: ver, Level: models.CosoWeightLevelComponent, Code: comp.Code, Weight: comp.Weight})
			}
			for _, prin := range comp.Principles {
				if k := models.CosoWeightLevelPrinciple + "/" + prin.Code; !existingW[k] {
					newWeights = append(newWeights, models.CosoCatalogWeight{Framework: fw, Version: ver, Level: models.CosoWeightLevelPrinciple, Code: prin.Code, Weight: prin.Weight})
				}
				for _, p := range prin.Pofs {
					order++
					if !existingP[p.Code] {
						newPofs = append(newPofs, models.CosoPointOfFocus{Framework: fw, CatalogVersion: ver, Code: p.Code,
							PrincipleCode: prin.Code, ComponentCode: comp.Code, NameEn: p.NameEN, NameID: p.NameID, Weight: p.Weight,
							SortOrder: order, DefaultOwnerRole: strings.Join(p.DefaultOwnerRoles, ","), IsActive: true})
					}
					for _, e := range p.Elements {
						if existingE[e.Code] {
							continue
						}
						newEls = append(newEls, models.CosoPofElement{Framework: fw, CatalogVersion: ver, Code: e.Code, PofCode: p.Code,
							Type: e.Type, DescriptionEn: e.DescriptionEN, DescriptionID: e.DescriptionID, Weight: e.Weight,
							DefaultMethods: strings.Join(e.DefaultMethods, ","), ScaleType: e.ScaleType,
							RequiresIndependentAssessor: e.RequiresIndependentAssessor, SortOrder: order*100 + e.Index})
					}
				}
			}
		}
		for _, batch := range []interface{}{&newWeights, &newPofs, &newEls} {
			var n int
			switch b := batch.(type) {
			case *[]models.CosoCatalogWeight:
				n = len(*b)
			case *[]models.CosoPointOfFocus:
				n = len(*b)
			case *[]models.CosoPofElement:
				n = len(*b)
			}
			if n == 0 {
				continue
			}
			res := tx.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(batch, 100)
			if res.Error != nil {
				return fmt.Errorf("seed coso pof catalogue: %w", res.Error)
			}
			inserted += int(res.RowsAffected)
		}
		if inserted > 0 {
			log.Printf("seeded %d missing COSO PoF catalogue rows (version %s)", inserted, ver)
		}
		return nil
	})
}

// ErrCatalogNotFound is returned by LoadCosoPofCatalogue for an unknown version.
var ErrCatalogNotFound = errors.New("COSO PoF catalogue version not found")

// LoadCosoPofCatalogue reads one catalogue version from the database
// (version "" = the current one). It reports the stored invariants through
// ValidateCosoPofCatalogue's caller if needed; it does not validate itself.
func LoadCosoPofCatalogue(db *gorm.DB, framework, version string) (*PofCatalogue, error) {
	var v models.CosoCatalogVersion
	q := db.Where("framework = ?", framework)
	if strings.TrimSpace(version) == "" {
		q = q.Where("is_current = ?", true)
	} else {
		q = q.Where("version = ?", version)
	}
	if err := q.Order("created_at desc").Take(&v).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCatalogNotFound
		}
		return nil, err
	}
	var weights []models.CosoCatalogWeight
	if err := db.Where("framework = ? AND version = ?", v.Framework, v.Version).Find(&weights).Error; err != nil {
		return nil, err
	}
	wmap := map[string]float64{}
	for _, w := range weights {
		wmap[w.Level+"/"+w.Code] = w.Weight
	}
	var pofs []models.CosoPointOfFocus
	if err := db.Where("framework = ? AND catalog_version = ?", v.Framework, v.Version).Order("sort_order, code").Find(&pofs).Error; err != nil {
		return nil, err
	}
	var els []models.CosoPofElement
	if err := db.Where("framework = ? AND catalog_version = ?", v.Framework, v.Version).Order("sort_order, code").Find(&els).Error; err != nil {
		return nil, err
	}
	elsByPof := map[string][]PofElement{}
	for _, e := range els {
		elsByPof[e.PofCode] = append(elsByPof[e.PofCode], PofElement{
			Code: e.Code, Index: elementIndex(e.Code), Type: e.Type, DescriptionEN: e.DescriptionEn, DescriptionID: e.DescriptionID,
			Weight: e.Weight, DefaultMethods: splitMethods(e.DefaultMethods), ScaleType: e.ScaleType,
			RequiresIndependentAssessor: e.RequiresIndependentAssessor,
		})
	}
	// Principle / component names and numbering come from the base catalogue
	var comps []models.CosoComponent
	_ = db.Where("framework = ?", v.Framework).Find(&comps).Error
	var prins []models.CosoPrinciple
	_ = db.Where("framework = ?", v.Framework).Find(&prins).Error
	prinName := map[string]models.CosoPrinciple{}
	for _, p := range prins {
		prinName[p.Code] = p
	}
	idDefs := indonesianNames()

	cat := &PofCatalogue{Framework: v.Framework, Version: v.Version, Name: v.Name, IsCurrent: v.IsCurrent}
	compIdx := map[string]int{}
	prinIdx := map[string][2]int{}
	order := []string{"CE", "RA", "CA", "IC", "MA"}
	for i, code := range order {
		nameEN := ""
		for _, c := range comps {
			if c.Code == code {
				nameEN = c.NameEn
			}
		}
		if nameEN == "" {
			for _, c := range models.Coso2013Catalogue {
				if c.Code == code {
					nameEN = c.NameEn
				}
			}
		}
		compIdx[code] = len(cat.Components)
		cat.Components = append(cat.Components, PofComponent{Code: code, Number: i + 1, NameEN: nameEN, NameID: idDefs.comp[code], Weight: wmap["COMPONENT/"+code]})
	}
	// principles in numeric order
	var codes []string
	for code := range wmap {
		if strings.HasPrefix(code, "PRINCIPLE/") {
			codes = append(codes, strings.TrimPrefix(code, "PRINCIPLE/"))
		}
	}
	sort.Strings(codes)
	for _, code := range codes {
		base := prinName[code]
		comp := base.ComponentCode
		if comp == "" {
			for _, cd := range models.Coso2013Catalogue {
				for _, p := range cd.Principles {
					if p.Code == code {
						comp = cd.Code
						base = models.CosoPrinciple{Code: code, Number: p.Number, NameEn: p.NameEn}
					}
				}
			}
		}
		ci, ok := compIdx[comp]
		if !ok {
			continue
		}
		c := &cat.Components[ci]
		prinIdx[code] = [2]int{ci, len(c.Principles)}
		c.Principles = append(c.Principles, PofPrinciple{Code: code, Number: base.Number, Index: len(c.Principles) + 1,
			NameEN: base.NameEn, NameID: idDefs.prin[code], Weight: wmap["PRINCIPLE/"+code]})
	}
	for _, p := range pofs {
		at, ok := prinIdx[p.PrincipleCode]
		if !ok {
			continue
		}
		prin := &cat.Components[at[0]].Principles[at[1]]
		owners := splitOwners(p.DefaultOwnerRole)
		if len(prin.DefaultOwnerRoles) == 0 {
			prin.DefaultOwnerRoles = owners
		}
		prin.Pofs = append(prin.Pofs, Pof{Code: p.Code, Index: pofIndex(p.Code), NameEN: p.NameEn, NameID: p.NameID, Weight: p.Weight,
			DefaultOwnerRoles: owners, Elements: elsByPof[p.Code]})
	}
	return cat, nil
}

func elementIndex(code string) int {
	if m := elementCodeRe.FindStringSubmatch(code); m != nil {
		var n int
		fmt.Sscan(m[4], &n)
		return n
	}
	return 0
}

func pofIndex(code string) int {
	if m := pofCodeRe.FindStringSubmatch(code); m != nil {
		var n int
		fmt.Sscan(m[3], &n)
		return n
	}
	return 0
}

type idNames struct {
	comp map[string]string
	prin map[string]string
}

func indonesianNames() idNames {
	n := idNames{comp: map[string]string{}, prin: map[string]string{}}
	for _, c := range cosoPofComponentDefs {
		n.comp[c.Code] = c.NameID
		for _, p := range c.Principles {
			n.prin[p.Code] = p.NameID
		}
	}
	return n
}
