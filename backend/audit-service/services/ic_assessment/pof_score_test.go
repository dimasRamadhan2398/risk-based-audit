package ic_assessment

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"audit-service/models"
)

func setAllPoF(pofs []models.ICAssessmentPoF, v float64) {
	for i := range pofs {
		pofs[i].Achievement = fp(v)
	}
}

func TestTestCatalogueIsValid(t *testing.T) {
	if issues := ValidateCatalogue(testCatalogue()); len(issues) > 0 {
		t.Fatalf("test catalogue invalid: %v", issues)
	}
}

func TestEmptyAssessmentScoresAreNullNotZero(t *testing.T) {
	pofs, els := testRows()
	tree := Calculate(pofs, els, DefaultThresholds())
	if tree.Total.Score != nil || tree.Total.Achievement != nil {
		t.Fatalf("total score must be nil when nothing is assessed, got %v / %v", tree.Total.Score, tree.Total.Achievement)
	}
	if tree.Total.Status != NodeNotAssessed {
		t.Errorf("status = %s", tree.Total.Status)
	}
	for _, c := range tree.Components {
		if c.Score != nil || c.Status != NodeNotAssessed {
			t.Errorf("component %s score=%v status=%s", c.Code, c.Score, c.Status)
		}
	}
	if tree.Total.Weight != 100 {
		t.Errorf("weight = %v", tree.Total.Weight)
	}
	if tree.Total.Progress.PofsTotal != 81 && tree.Total.Progress.PofsTotal != 80 {
		t.Logf("test catalogue has %d PoFs", tree.Total.Progress.PofsTotal)
	}
}

func TestPartialAssessmentKeepsNodeIncomplete(t *testing.T) {
	pofs, els := testRows()
	pofs[0].Achievement = fp(1) // 1.1.1 only
	tree := Calculate(pofs, els, DefaultThresholds())
	p01 := tree.Principle("P01")
	if p01.Score != nil || p01.Status != NodeInProgress {
		t.Fatalf("P01 should be in progress with no score: %+v", p01.ScoreBlock)
	}
	near(t, "P01 assessed score", &p01.AssessedScore, 1.375)
	if tree.Total.Score != nil || tree.Total.AssessedScore == 0 {
		t.Errorf("total must stay null with assessed score > 0: %+v", tree.Total)
	}
}

// Hand-computed: P01 has weight 5.5 over 4 PoFs of 1.375. With the PoF values
// 0.96, 0.95, 0.83 and 0.50 (the magnitudes of the Lampiran figures) its score
// is 1.375 x 3.24 = 4.455 and its achievement 0.81.
func TestPoFLevelScoreByHand(t *testing.T) {
	pofs, els := testRows()
	for i, v := range []float64{0.96, 0.95, 0.83, 0.50} {
		pofs[i].Achievement = fp(v)
	}
	tree := Calculate(pofs, els, DefaultThresholds())
	p01 := tree.Principle("P01")
	near(t, "P01 score", p01.Score, 4.455)
	near(t, "P01 achievement", p01.Achievement, 0.81)
	if p01.Status != PrincipleYes {
		t.Errorf("0.81 should be YES at the default threshold, got %s", p01.Status)
	}
	for i, want := range []string{PrincipleYes, PrincipleYes, PrincipleYes, PrinciplePartial} {
		if got := tree.PoF(pofs[i].PofCode).Status; got != want {
			t.Errorf("PoF %s status %s, want %s", pofs[i].PofCode, got, want)
		}
	}
}

func TestElementsDecideThePoF(t *testing.T) {
	pofs, els := testRows()
	pofs[0].Achievement = fp(0.1) // ignored once elements have values
	for i := range els {
		if els[i].PofCode == "1.1.1" {
			if els[i].ElementCode == "1.1.1.1" {
				els[i].Achievement = fp(1)
			} else {
				els[i].Achievement = fp(0.5)
			}
		}
	}
	tree := Calculate(pofs, els, DefaultThresholds())
	pc := tree.PoF("1.1.1")
	if pc.Basis != BasisElements || !pc.Complete {
		t.Fatalf("basis=%s complete=%v", pc.Basis, pc.Complete)
	}
	near(t, "PoF score", pc.Score, 1.03125) // 0.6875 x 1 + 0.6875 x 0.5
	near(t, "PoF achievement", pc.Achievement, 0.75)
	if pc.Status != PrinciplePartial {
		t.Errorf("0.75 is PARTIAL at 0.80, got %s", pc.Status)
	}
}

func TestPartiallyFilledElementsGiveNoPoFScore(t *testing.T) {
	pofs, els := testRows()
	for i := range els {
		if els[i].ElementCode == "1.1.1.1" {
			els[i].Achievement = fp(1)
		}
	}
	pc := Calculate(pofs, els, DefaultThresholds()).PoF("1.1.1")
	if pc.Complete || pc.Score != nil || pc.Status != NodeInProgress || pc.ElementsDone != 1 {
		t.Fatalf("%+v", pc)
	}
}

func TestEveryPoFAtOneGivesFullMarks(t *testing.T) {
	pofs, els := testRows()
	setAllPoF(pofs, 1)
	tree := Calculate(pofs, els, DefaultThresholds())
	near(t, "total", tree.Total.Score, 100)
	want := map[string]float64{"CE": 25, "RA": 25, "CA": 20, "IC": 17.5, "MA": 12.5}
	for _, c := range tree.Components {
		near(t, c.Code, c.Score, want[c.Code])
		if c.Status != PrincipleYes {
			t.Errorf("%s status %s", c.Code, c.Status)
		}
	}
	if tree.Total.Status != PrincipleYes {
		t.Errorf("total status %s", tree.Total.Status)
	}
}

// Synthetic scenario with the magnitudes quoted from the Lampiran 7-29/30
// (CE component 21.64 of 25, total 83.26): the engine adds weight x achievement
// at every level. This checks the arithmetic of the roll-up, NOT the user's
// sheet (see TestLampiranFixture).
func TestComponentRollUpAddsUp(t *testing.T) {
	pofs, els := testRows()
	setAllPoF(pofs, 0.8656) // 0.8656 x 25 = 21.64
	tree := Calculate(pofs, els, DefaultThresholds())
	ce := tree.Component("CE")
	near(t, "CE score", ce.Score, 21.64)
	near(t, "CE achievement", ce.Achievement, 0.8656)
	near(t, "total", tree.Total.Score, 86.56)
	var sum float64
	for _, c := range tree.Components {
		sum += *c.Score
	}
	near(t, "sum of components", &sum, *tree.Total.Score)
}

// TestLampiranFixture reproduces the figures of the user's sheet (Lampiran
// 7-29/30: component CE 21.64 of 25, total 83.26) from the PoF figures and
// weights of that sheet. The sheet is not in the repository (it is third
// party material), so the fixture is read from testdata/lampiran_sample.json
// when someone supplies it; the test skips otherwise. Format:
//
//	{"pofs": [{"code": "1.1.1", "principle": "P01", "weight": 1.375, "achievement": 0.96}, ...],
//	 "expect": {"components": {"CE": 21.64, "RA": ..}, "total": 83.26}}
func TestLampiranFixture(t *testing.T) {
	raw, err := os.ReadFile("testdata/lampiran_sample.json")
	if err != nil {
		t.Skip("testdata/lampiran_sample.json is not present: the sheet figures were not available to the implementation")
	}
	var fx struct {
		Pofs []struct {
			Code        string   `json:"code"`
			Principle   string   `json:"principle"`
			Weight      float64  `json:"weight"`
			Achievement *float64 `json:"achievement"`
		} `json:"pofs"`
		Expect struct {
			Components map[string]float64 `json:"components"`
			Total      float64            `json:"total"`
		} `json:"expect"`
	}
	if err := json.Unmarshal(raw, &fx); err != nil {
		t.Fatal(err)
	}
	var pofs []models.ICAssessmentPoF
	for i, p := range fx.Pofs {
		meta, ok := PrincipleByCode(p.Principle)
		if !ok {
			t.Fatalf("unknown principle %s", p.Principle)
		}
		pofs = append(pofs, models.ICAssessmentPoF{PofCode: p.Code, PrincipleCode: p.Principle, ComponentCode: meta.ComponentCode,
			SortOrder: i + 1, Weight: p.Weight, Achievement: p.Achievement})
	}
	tree := Calculate(pofs, nil, DefaultThresholds())
	for code, want := range fx.Expect.Components {
		near(t, code, tree.Component(code).Score, want)
	}
	near(t, "total", tree.Total.Score, fx.Expect.Total)
}

// ---- thresholds ----

func TestDefaultThresholdsAreValid(t *testing.T) {
	if err := DefaultThresholds().Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestThresholdValidation(t *testing.T) {
	bad := []Thresholds{{0.8, 0.8}, {0.4, 0.8}, {1.2, 0.4}, {0.8, 0}, {0.8, -0.1}, {0.8, 0.4005}, {0, 0}}
	for _, th := range bad {
		if th.Validate() == nil {
			t.Errorf("%+v should be rejected", th)
		}
	}
	for _, th := range []Thresholds{{0.8, 0.4}, {1, 0.5}, {0.9, 0.6}, {0.75, 0.25}} {
		if err := th.Validate(); err != nil {
			t.Errorf("%+v: %v", th, err)
		}
	}
}

// The default boundaries sit in the gaps of the discrete scale (0, .25, .5,
// .75, 1): 0.40 is between 0.25 and 0.50, 0.80 between 0.75 and 1. So no
// discrete value ever lands on a boundary, and the questionnaire value 0.82
// from the sheet is YES.
func TestThresholdClassificationOfSheetFigures(t *testing.T) {
	th := DefaultThresholds()
	cases := map[float64]string{
		1: PrincipleYes, 0.96: PrincipleYes, 0.95: PrincipleYes, 0.83: PrincipleYes, 0.82: PrincipleYes, 0.80: PrincipleYes,
		0.79: PrinciplePartial, 0.75: PrinciplePartial, 0.50: PrinciplePartial, 0.40: PrinciplePartial,
		0.39: PrincipleNo, 0.33: PrincipleNo, 0.25: PrincipleNo, 0.06: PrincipleNo, 0: PrincipleNo,
	}
	for a, want := range cases {
		if got := th.StatusFor(a); got != want {
			t.Errorf("StatusFor(%v) = %s, want %s", a, got, want)
		}
	}
	for _, v := range []float64{0, 0.25, 0.5, 0.75, 1} {
		if v == th.Yes || v == th.Partial {
			t.Errorf("discrete value %v sits on a threshold", v)
		}
	}
}

func TestStatusForToleratesFloatingPointSums(t *testing.T) {
	th := DefaultThresholds()
	if th.StatusFor(0.7999999999) != PrincipleYes || th.StatusFor(0.3999999999) != PrinciplePartial {
		t.Error("floating point dust must not flip the status")
	}
}

// ---- value validation ----

func TestValidateAchievement(t *testing.T) {
	ok := []struct {
		v     float64
		scale string
		pof   bool
	}{{0, ScaleDiscrete, false}, {0.25, ScaleDiscrete, false}, {0.5, ScaleDiscrete, false}, {0.75, ScaleDiscrete, false}, {1, ScaleDiscrete, false},
		{0.82, ScaleQuestionnaire, false}, {0.333, ScaleQuestionnaire, false}, {0.96, ScaleDiscrete, true}, {0.05, "", true}}
	for _, c := range ok {
		if err := ValidateAchievement(c.v, c.scale, c.pof); err != nil {
			t.Errorf("%+v rejected: %v", c, err)
		}
	}
	bad := []struct {
		v     float64
		scale string
		pof   bool
	}{{0.8, ScaleDiscrete, false}, {0.3, ScaleDiscrete, false}, {-0.1, ScaleDiscrete, false}, {1.01, ScaleQuestionnaire, false},
		{0.12345, ScaleQuestionnaire, false}, {0.1234, "", true}, {2, "", true}}
	for _, c := range bad {
		if ValidateAchievement(c.v, c.scale, c.pof) == nil {
			t.Errorf("%+v should be rejected", c)
		}
	}
}

func TestNormalizeMethods(t *testing.T) {
	got, err := NormalizeMethods([]string{"w", "D"})
	if err != nil || got != "D,W" {
		t.Fatalf("%q %v", got, err)
	}
	for _, in := range [][]string{{"X"}, {"D", "d"}} {
		if _, err := NormalizeMethods(in); err == nil {
			t.Errorf("%v should be rejected", in)
		}
	}
	if got, _ := NormalizeMethods(nil); got != "" {
		t.Errorf("empty = %q", got)
	}
}

// ---- derived rows and conclusion ----

func newQuantAssessment() (*models.ICAssessment, []models.ICAssessmentPoF, []models.ICAssessmentElement) {
	pofs, els := testRows()
	a := &models.ICAssessment{Method: MethodQuantitativePoF}
	for _, c := range ComponentCodes {
		a.Components = append(a.Components, models.ICAssessmentComponent{ComponentCode: c})
	}
	for _, p := range Principles {
		a.Principles = append(a.Principles, models.ICAssessmentPrinciple{PrincipleCode: p.Code, ComponentCode: p.ComponentCode, Status: PrincipleNotAssessed, Severity: SeverityNone})
	}
	return a, pofs, els
}

func TestSyncDerivedRowsAllGood(t *testing.T) {
	a, pofs, els := newQuantAssessment()
	setAllPoF(pofs, 0.9)
	SyncDerivedRows(a, Calculate(pofs, els, DefaultThresholds()))
	for _, p := range a.Principles {
		if p.Status != PrincipleYes || p.Severity != SeverityNone {
			t.Errorf("%s %s/%s", p.PrincipleCode, p.Status, p.Severity)
		}
		if !strings.HasPrefix(p.Rationale, AutoRationalePrefix) && p.Rationale != "" {
			t.Errorf("rationale %q", p.Rationale)
		}
		if err := ValidatePrinciple(&p); err != nil {
			t.Errorf("derived principle fails the qualitative validator: %v", err)
		}
	}
	for _, c := range a.Components {
		if c.FinalStatus != PrincipleYes || c.FinalRating == nil || *c.FinalRating != 5 || c.IsOverride {
			t.Errorf("%s %+v", c.ComponentCode, c)
		}
	}
	if a.DerivedConclusion != ConclusionEffective {
		t.Errorf("conclusion %s", a.DerivedConclusion)
	}
}

func TestSyncDerivedRowsLeavesIncompleteEmpty(t *testing.T) {
	a, pofs, els := newQuantAssessment()
	for i := range pofs {
		if pofs[i].ComponentCode == "CE" {
			pofs[i].Achievement = fp(1)
		}
	}
	SyncDerivedRows(a, Calculate(pofs, els, DefaultThresholds()))
	for _, c := range a.Components {
		if c.ComponentCode == "CE" {
			if c.FinalStatus != PrincipleYes {
				t.Errorf("CE should be derived: %+v", c)
			}
			continue
		}
		if c.FinalRating != nil || c.FinalStatus != "" || c.FinalSeverity != "" {
			t.Errorf("%s must stay empty while unassessed: %+v", c.ComponentCode, c)
		}
	}
	if a.DerivedConclusion != ConclusionIncomplete {
		t.Errorf("conclusion %s", a.DerivedConclusion)
	}
}

func TestOneWeakPrincipleMakesTheComponentNo(t *testing.T) {
	a, pofs, els := newQuantAssessment()
	setAllPoF(pofs, 1)
	for i := range pofs {
		if pofs[i].PrincipleCode == "P03" {
			pofs[i].Achievement = fp(0.2)
		}
	}
	tree := Calculate(pofs, els, DefaultThresholds())
	SyncDerivedRows(a, tree)
	var ce models.ICAssessmentComponent
	for _, c := range a.Components {
		if c.ComponentCode == "CE" {
			ce = c
		}
	}
	// CE achievement is (25 - 4.5 x 0.8) / 25 = 0.856 (YES by score) but P03 is
	// NO, so the component is NO with a MAJOR deficiency and a rating of at most 2.
	if tree.Component("CE").Status != PrincipleYes {
		t.Fatalf("by score alone CE is YES, got %s", tree.Component("CE").Status)
	}
	if ce.FinalStatus != PrincipleNo || ce.FinalSeverity != SeverityMajor || ce.FinalRating == nil || *ce.FinalRating > 2 {
		t.Errorf("CE: %+v", ce)
	}
	if a.DerivedConclusion != ConclusionNotEffective {
		t.Errorf("conclusion %s", a.DerivedConclusion)
	}
}

func TestPartialPrincipleSeverityFollowsTheMidpoint(t *testing.T) {
	th := DefaultThresholds() // midpoint 0.60
	if principleSeverity(PrinciplePartial, 0.70, th) != SeverityDeficiency || principleSeverity(PrinciplePartial, 0.50, th) != SeveritySignificant {
		t.Error("PARTIAL severity should split at the midpoint of the thresholds")
	}
	if principleSeverity(PrincipleNo, 0.1, th) != SeverityMajor || principleSeverity(PrincipleYes, 0.9, th) != SeverityNone {
		t.Error("YES is NONE and NO is MAJOR")
	}
}

func TestRatingBands(t *testing.T) {
	for a, want := range map[float64]int{1: 5, 0.9: 5, 0.89: 4, 0.8: 4, 0.7: 3, 0.6: 2, 0.59: 1, 0: 1} {
		if got := RatingFromAchievement(a); got != want {
			t.Errorf("rating(%v) = %d, want %d", a, got, want)
		}
	}
}

func TestQuantSubmitIssues(t *testing.T) {
	a, pofs, els := newQuantAssessment()
	tree := Calculate(pofs, els, DefaultThresholds())
	SyncDerivedRows(a, tree)
	issues := ValidateQuantForSubmit(a, tree)
	if len(issues) < len(pofs) {
		t.Fatalf("every PoF should be reported as not assessed, got %d issues", len(issues))
	}

	// all assessed at 1, conclusion missing
	setAllPoF(pofs, 1)
	tree = Calculate(pofs, els, DefaultThresholds())
	SyncDerivedRows(a, tree)
	issues = ValidateQuantForSubmit(a, tree)
	if len(issues) != 1 || !strings.Contains(issues[0], "final_conclusion") {
		t.Fatalf("only the conclusion should block: %v", issues)
	}
	a.FinalConclusion = ConclusionEffective
	if issues = ValidateQuantForSubmit(a, tree); len(issues) != 0 {
		t.Fatalf("should pass: %v", issues)
	}

	// below 1 without an explanation blocks
	pofs[0].Achievement = fp(0.9)
	tree = Calculate(pofs, els, DefaultThresholds())
	SyncDerivedRows(a, tree)
	issues = ValidateQuantForSubmit(a, tree)
	if len(issues) != 1 || !strings.Contains(issues[0], "1.1.1") || !strings.Contains(issues[0], "explanation") {
		t.Fatalf("%v", issues)
	}
	pofs[0].Explanation = "Two of four sampled documents were not signed."
	tree = Calculate(pofs, els, DefaultThresholds())
	SyncDerivedRows(a, tree)
	if issues = ValidateQuantForSubmit(a, tree); len(issues) != 0 {
		t.Fatalf("%v", issues)
	}
}

func TestQuantSubmitBlocksPartialElements(t *testing.T) {
	a, pofs, els := newQuantAssessment()
	setAllPoF(pofs, 1)
	pofs[0].Achievement = nil
	for i := range els {
		if els[i].ElementCode == "1.1.1.1" {
			els[i].Achievement = fp(1)
		}
	}
	a.FinalConclusion = ConclusionEffective
	tree := Calculate(pofs, els, DefaultThresholds())
	SyncDerivedRows(a, tree)
	issues := ValidateQuantForSubmit(a, tree)
	found := false
	for _, i := range issues {
		if strings.HasPrefix(i, "1.1.1: 1 of 2 elements are not assessed") {
			found = true
		}
	}
	if !found {
		t.Fatalf("%v", issues)
	}
}

func TestEffectiveConclusionRefusedWithANoComponent(t *testing.T) {
	a, pofs, els := newQuantAssessment()
	setAllPoF(pofs, 0.2)
	a.FinalConclusion = ConclusionEffective
	tree := Calculate(pofs, els, DefaultThresholds())
	SyncDerivedRows(a, tree)
	if len(ValidateQuantForSubmit(a, tree)) == 0 {
		t.Fatal("EFFECTIVE with all components NO must be refused")
	}
}

func TestSafeCellAndUnsafeCell(t *testing.T) {
	for _, in := range []string{"=1+1", "+SUM(A1)", "-2", "@cmd", "\tx", "\rx"} {
		out := SafeCell(in)
		if out == in || out[0] != '\'' {
			t.Errorf("%q not neutralised: %q", in, out)
		}
		if UnsafeCell(out) != in {
			t.Errorf("round trip of %q gave %q", in, UnsafeCell(out))
		}
	}
	for _, in := range []string{"plain", "", "a=b", "'quoted"} {
		if SafeCell(in) != in {
			t.Errorf("%q must stay as it is", in)
		}
	}
}
