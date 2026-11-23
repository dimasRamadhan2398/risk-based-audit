package services

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"risk-service/models"
	"risk-service/pkg/masterclient"
	"risk-service/pkg/sqlitetest"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func date(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }

func TestDueQuarters(t *testing.T) {
	cases := []struct {
		name    string
		year, q int
		asOf    time.Time
		want    []int
	}{
		{"annual, mid Q3", 2026, 0, date(2026, 8, 15), []int{1, 2}},
		{"annual, last day of Q2 counts", 2026, 0, date(2026, 6, 30), []int{1, 2}},
		{"annual, year closed", 2026, 0, date(2026, 12, 31), []int{1, 2, 3, 4}},
		{"annual, none ended -> current", 2026, 0, date(2026, 2, 10), []int{1}},
		{"quarterly Q2, after Q2", 2026, 2, date(2026, 9, 1), []int{1, 2}},
		{"quarterly Q3, inside Q3", 2026, 3, date(2026, 8, 1), []int{1, 2}},
		{"quarterly Q1, inside Q1 -> Q1", 2026, 1, date(2026, 2, 1), []int{1}},
		{"as_of before the year -> Q1", 2026, 0, date(2025, 12, 1), []int{1}},
	}
	for _, c := range cases {
		if got := DueQuarters(c.year, c.q, c.asOf); !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: DueQuarters = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestPeriodBounds(t *testing.T) {
	s, e := PeriodBounds(2026, 0)
	if s != date(2026, 1, 1) || e != date(2026, 12, 31) {
		t.Errorf("annual = %v..%v", s, e)
	}
	s, e = PeriodBounds(2026, 2)
	if s != date(2026, 4, 1) || e != date(2026, 6, 30) {
		t.Errorf("Q2 = %v..%v", s, e)
	}
	s, e = PeriodBounds(2024, 1)
	if e != date(2024, 3, 31) || s != date(2024, 1, 1) {
		t.Errorf("Q1 = %v..%v", s, e)
	}
}

// Mirrors frontend/tests/unit/rcmBranch.test.ts (resolveRcmItemLocationId).
func TestRCMBranchResolverMatchesFrontendRule(t *testing.T) {
	jkt, sby := uuid.New(), uuid.New()
	locs := []masterclient.Location{{ID: jkt, Name: "Jakarta Branch", IsActive: true}, {ID: sby, Name: "Surabaya Branch", IsActive: true}}
	r := NewRCMBranchResolver(locs, true)

	rJkt := &models.RiskRegister{Profile: models.RiskProfile{LocationID: &jkt}}
	rNone := &models.RiskRegister{}
	rLegacy := &models.RiskRegister{Profile: models.RiskProfile{LocationName: "Surabaya Branch"}}
	rSentinel := &models.RiskRegister{Profile: models.RiskProfile{DepartmentID: uuid.MustParse("00000000-0000-0000-0000-000000000003")}}
	rMoved := &models.RiskRegister{
		Profile:     models.RiskProfile{LocationID: &jkt},
		Assessments: []models.RiskAssessment{{Year: 2025, LocationID: &sby}, {Year: 2026}},
	}
	eq := func(name string, got *uuid.UUID, want *uuid.UUID) {
		t.Helper()
		if (got == nil) != (want == nil) || (got != nil && *got != *want) {
			t.Errorf("%s: got %v, want %v", name, got, want)
		}
	}
	eq("linked risk with a location", r.LocationID(2026, rJkt), &jkt)
	eq("legacy branch name", r.LocationID(0, rLegacy), &sby)
	eq("legacy sentinel department", r.LocationID(2026, rSentinel), &sby)
	eq("no risk / risk not found", r.LocationID(2026, nil), nil)
	eq("risk without a location", r.LocationID(2026, rNone), nil)
	eq("year assessment location wins", r.LocationID(2025, rMoved), &sby)
	eq("year without assessment location", r.LocationID(2026, rMoved), &jkt)

	// Location master unavailable: stored ids still count, names do not.
	down := NewRCMBranchResolver(nil, false)
	eq("down: stored id", down.LocationID(2026, rJkt), &jkt)
	eq("down: legacy name unmatched", down.LocationID(2026, rLegacy), nil)
}

type signalLocations struct {
	locs []masterclient.Location
	err  error
}

func (f signalLocations) ListLocations(context.Context) ([]masterclient.Location, error) {
	return f.locs, f.err
}

func openSignalsDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := sqlitetest.Open(t,
		&models.RiskControlMatrix{}, &models.CosoComponent{}, &models.CosoPrinciple{}, &models.RcmCosoPrinciple{},
		&models.RiskProfile{}, &models.RiskRegister{}, &models.RiskAssessment{}, &models.RiskMitigation{},
	)
	if err := SeedCosoCatalogue(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func addRisk(t *testing.T, db *gorm.DB, lik, imp int, loc *uuid.UUID) uuid.UUID {
	t.Helper()
	p := models.RiskProfile{ID: uuid.New(), Category: "Ops", LocationID: loc}
	if err := db.Create(&p).Error; err != nil {
		t.Fatal(err)
	}
	r := models.RiskRegister{ID: uuid.New(), ProfileID: p.ID, RiskSource: models.RiskSourceDirect, RiskEvent: "r",
		InherentLikelihood: lik, InherentImpact: imp, FinalRiskLevel: models.RiskFinalLevelLow}
	if err := db.Omit("Profile", "Assessments").Create(&r).Error; err != nil {
		t.Fatal(err)
	}
	return r.ID
}

func addRCM(t *testing.T, db *gorm.DB, dept *uuid.UUID, year int, risk *uuid.UUID, score float64, codes ...string) uuid.UUID {
	t.Helper()
	row := models.RiskControlMatrix{ID: uuid.New(), RiskCode: "R", RiskEvent: "e", ControlCode: "C", ControlDescription: "d",
		ControlType: "Preventive", ControlOwner: "o", Department: "D", Year: year, DepartmentID: dept, RiskID: risk,
		DesignEffectivenessRating: 4, OperatingEffectivenessRating: 4, CoverageCompletenessRating: 4, TimelinessRating: 4, AutomationMonitoringRating: 4,
		TotalWeightedScore: score}
	if err := db.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	for _, c := range codes {
		var p models.CosoPrinciple
		if err := db.Where("code = ?", c).First(&p).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Omit("Rcm", "Principle").Create(&models.RcmCosoPrinciple{RcmID: row.ID, PrincipleID: p.ID}).Error; err != nil {
			t.Fatal(err)
		}
	}
	return row.ID
}

func signal(t *testing.T, res *CosoSignalsResult, key string) CosoSignal {
	t.Helper()
	for _, s := range res.Signals {
		if s.SignalKey == key {
			return s
		}
	}
	t.Fatalf("signal %s missing", key)
	return CosoSignal{}
}

func TestCosoSignalsNoData(t *testing.T) {
	db := openSignalsDB(t)
	other := uuid.New()
	addRCM(t, db, &other, 2026, nil, 80)
	addRCM(t, db, nil, 2026, nil, 80)

	res, err := ComputeCosoSignals(context.Background(), db, nil, CosoSignalsParams{DepartmentID: uuid.New(), Year: 2026, AsOf: date(2026, 8, 1)})
	if err != nil {
		t.Fatal(err)
	}
	if res.RCMCount != 0 || res.RiskCount != 0 || res.UnassignedRCMCount != 1 {
		t.Errorf("counts = %d/%d/%d", res.RCMCount, res.RiskCount, res.UnassignedRCMCount)
	}
	for _, key := range []string{SignalRAQuarterlyCoverage, SignalRAAboveAppetiteMitigated, SignalCARCMEffectiveness} {
		s := signal(t, res, key)
		if s.State != SignalStateNoData || s.ValuePct != nil {
			t.Errorf("%s = %s %v, want NO_DATA and no value", key, s.State, s.ValuePct)
		}
	}
	ca := signal(t, res, SignalCARCMEffectiveness)
	if stats := ca.Detail["per_principle"].([]CosoPrincipleStat); len(stats) != 17 {
		t.Errorf("per_principle has %d rows, want 17", len(stats))
	}
}

func TestCosoSignalsValues(t *testing.T) {
	db := openSignalsDB(t)
	dept := uuid.New()
	high := addRisk(t, db, 4, 5, nil) // HIGH inherent
	low := addRisk(t, db, 1, 1, nil)  // LOW
	mod := addRisk(t, db, 3, 3, nil)  // MODERATE

	// high: Q1 and Q2 filled, latest filled Q2 = (lik 4, imp 4) MODERATE_HIGH -> above
	// low: Q1 filled only (LOW)
	// mod: no assessment row -> inherent MODERATE -> above
	for _, a := range []models.RiskAssessment{
		{ID: uuid.New(), RiskRegisterID: high, Year: 2026, ImpactQ1: 5, LikelihoodQ1: 4, ImpactQ2: 4, LikelihoodQ2: 4, ImpactQ3: 1, LikelihoodQ3: 1},
		{ID: uuid.New(), RiskRegisterID: low, Year: 2026, ImpactQ1: 1, LikelihoodQ1: 1},
		{ID: uuid.New(), RiskRegisterID: low, Year: 2025, ImpactQ2: 1, LikelihoodQ2: 1},
	} {
		if err := db.Create(&a).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Create(&models.RiskMitigation{ID: uuid.New(), RiskID: high, RiskEvent: "e", MitigationPlan: "m"}).Error; err != nil {
		t.Fatal(err)
	}
	r1 := addRCM(t, db, &dept, 2026, &high, 80, "P10", "P11")
	addRCM(t, db, &dept, 2026, &low, 60, "P10")
	addRCM(t, db, &dept, 2026, &mod, 70)
	addRCM(t, db, &dept, 2026, &high, 90) // same risk twice: R stays distinct
	addRCM(t, db, &dept, 2025, &high, 10) // other year: not in C
	deleted := addRCM(t, db, &dept, 2026, &low, 0)
	if err := db.Delete(&models.RiskControlMatrix{}, "id = ?", deleted).Error; err != nil {
		t.Fatal(err)
	}

	res, err := ComputeCosoSignals(context.Background(), db, nil, CosoSignalsParams{DepartmentID: dept, Year: 2026, AsOf: date(2026, 8, 1)})
	if err != nil {
		t.Fatal(err)
	}
	if res.RCMCount != 4 || res.RiskCount != 3 {
		t.Fatalf("C=%d R=%d, want 4 and 3", res.RCMCount, res.RiskCount)
	}

	cov := signal(t, res, SignalRAQuarterlyCoverage)
	// due Q1,Q2 x 3 risks = 6; filled: high Q1,Q2 + low Q1 = 3
	if cov.State != SignalStateOK || *cov.Numerator != 3 || *cov.Denominator != 6 || *cov.ValuePct != 50 {
		t.Errorf("coverage = %s %v/%v %v", cov.State, *cov.Numerator, *cov.Denominator, *cov.ValuePct)
	}

	app := signal(t, res, SignalRAAboveAppetiteMitigated)
	// above: high (Q2 MODERATE_HIGH; Q3 is not due yet), mod (inherent MODERATE); mitigated: high
	if app.State != SignalStateOK || *app.Numerator != 1 || *app.Denominator != 2 || *app.ValuePct != 50 {
		t.Errorf("appetite = %s %v/%v %v", app.State, *app.Numerator, *app.Denominator, *app.ValuePct)
	}

	ca := signal(t, res, SignalCARCMEffectiveness)
	if ca.State != SignalStateOK || *ca.ValuePct != 75 || *ca.Denominator != 4 || ca.Numerator != nil {
		t.Errorf("ca = %s %v", ca.State, *ca.ValuePct)
	}
	stats := map[string]CosoPrincipleStat{}
	for _, s := range ca.Detail["per_principle"].([]CosoPrincipleStat) {
		stats[s.PrincipleCode] = s
	}
	if s := stats["P10"]; s.TaggedControlCount != 2 || *s.TaggedControlAvgPct != 70 {
		t.Errorf("P10 = %+v", s)
	}
	if s := stats["P11"]; s.TaggedControlCount != 1 || *s.TaggedControlAvgPct != 80 {
		t.Errorf("P11 = %+v", s)
	}
	if s := stats["P01"]; s.TaggedControlCount != 0 || s.TaggedControlAvgPct != nil {
		t.Errorf("P01 = %+v", s)
	}
	if ca.Detail["untagged_count"] != 2 {
		t.Errorf("untagged_count = %v", ca.Detail["untagged_count"])
	}
	_ = r1
}

func TestCosoSignalsAllWithinAppetiteIsNotApplicable(t *testing.T) {
	db := openSignalsDB(t)
	dept := uuid.New()
	low := addRisk(t, db, 1, 2, nil)
	addRCM(t, db, &dept, 2026, &low, 80)
	res, err := ComputeCosoSignals(context.Background(), db, nil, CosoSignalsParams{DepartmentID: dept, Year: 2026, AsOf: date(2026, 12, 31)})
	if err != nil {
		t.Fatal(err)
	}
	if s := signal(t, res, SignalRAAboveAppetiteMitigated); s.State != SignalStateNotApplicable || s.ValuePct != nil {
		t.Errorf("appetite = %s", s.State)
	}
	if s := signal(t, res, SignalRAQuarterlyCoverage); *s.Denominator != 4 || *s.Numerator != 0 {
		t.Errorf("coverage = %v/%v", *s.Numerator, *s.Denominator)
	}
}

func TestCosoSignalsBranchScope(t *testing.T) {
	db := openSignalsDB(t)
	dept := uuid.New()
	jkt, sby := uuid.New(), uuid.New()
	inJkt := addRisk(t, db, 3, 3, &jkt)
	inSby := addRisk(t, db, 3, 3, &sby)
	addRCM(t, db, &dept, 2026, &inJkt, 90)
	addRCM(t, db, &dept, 2026, &inSby, 50)
	addRCM(t, db, &dept, 2026, nil, 10) // unassigned: never in a branch scope

	locs := signalLocations{locs: []masterclient.Location{{ID: jkt, Name: "Jakarta Branch", IsActive: true}, {ID: sby, Name: "Surabaya Branch", IsActive: true}}}
	res, err := ComputeCosoSignals(context.Background(), db, locs, CosoSignalsParams{DepartmentID: dept, Year: 2026, LocationID: &jkt, AsOf: date(2026, 8, 1)})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Scope.BranchScoped || !res.Scope.LocationMasterAvailable || res.RCMCount != 1 || res.RiskCount != 1 {
		t.Fatalf("scope = %+v C=%d R=%d", res.Scope, res.RCMCount, res.RiskCount)
	}
	if s := signal(t, res, SignalCARCMEffectiveness); *s.ValuePct != 90 {
		t.Errorf("ca = %v", *s.ValuePct)
	}

	// Location master down: stored location ids still narrow, flagged.
	res, err = ComputeCosoSignals(context.Background(), db, signalLocations{err: errors.New("down")}, CosoSignalsParams{DepartmentID: dept, Year: 2026, LocationID: &sby, AsOf: date(2026, 8, 1)})
	if err != nil {
		t.Fatal(err)
	}
	if res.Scope.LocationMasterAvailable || res.RCMCount != 1 {
		t.Errorf("down: available=%v C=%d", res.Scope.LocationMasterAvailable, res.RCMCount)
	}
}

func TestSeedCosoCatalogueInsertIfMissing(t *testing.T) {
	db := sqlitetest.Open(t, &models.CosoComponent{}, &models.CosoPrinciple{})
	count := func() (int64, int64) {
		var c, p int64
		db.Model(&models.CosoComponent{}).Count(&c)
		db.Model(&models.CosoPrinciple{}).Count(&p)
		return c, p
	}
	for run := 1; run <= 2; run++ {
		if err := SeedCosoCatalogue(db); err != nil {
			t.Fatal(err)
		}
		if c, p := count(); c != 5 || p != 17 {
			t.Fatalf("run %d: %d components, %d principles; want 5 and 17", run, c, p)
		}
	}
	// A deleted row comes back; an edited row keeps the edit.
	if err := db.Where("code = ?", "P08").Delete(&models.CosoPrinciple{}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&models.CosoPrinciple{}).Where("code = ?", "P01").
		Updates(map[string]any{"name_en": "Edited", "is_active": false}).Error; err != nil {
		t.Fatal(err)
	}
	if err := SeedCosoCatalogue(db); err != nil {
		t.Fatal(err)
	}
	if _, p := count(); p != 17 {
		t.Errorf("principles = %d after re-seed, want 17", p)
	}
	var p01 models.CosoPrinciple
	db.Where("code = ?", "P01").First(&p01)
	if p01.NameEn != "Edited" || p01.IsActive {
		t.Errorf("P01 was overwritten: %+v", p01)
	}
	var p08 models.CosoPrinciple
	db.Where("code = ?", "P08").First(&p08)
	if p08.ComponentCode != "RA" || p08.Number != 8 || !p08.IsActive || p08.Framework != models.CosoFramework2013 {
		t.Errorf("P08 re-seeded as %+v", p08)
	}
}

// The catalogue is the table in plan 4.1.1: 5 components, P01..P17 in order.
func TestCoso2013CatalogueShape(t *testing.T) {
	want := map[string][]string{
		"CE": {"P01", "P02", "P03", "P04", "P05"},
		"RA": {"P06", "P07", "P08", "P09"},
		"CA": {"P10", "P11", "P12"},
		"IC": {"P13", "P14", "P15"},
		"MA": {"P16", "P17"},
	}
	order := []string{"CE", "RA", "CA", "IC", "MA"}
	next := 1
	for i, c := range models.Coso2013Catalogue {
		if c.Code != order[i] || c.SortOrder != i+1 {
			t.Errorf("component %d = %s/%d", i, c.Code, c.SortOrder)
		}
		var codes []string
		for _, p := range c.Principles {
			codes = append(codes, p.Code)
			if p.Number != next {
				t.Errorf("%s number %d, want %d", p.Code, p.Number, next)
			}
			next++
		}
		if !reflect.DeepEqual(codes, want[c.Code]) {
			t.Errorf("%s principles = %v, want %v", c.Code, codes, want[c.Code])
		}
	}
	if next != 18 {
		t.Errorf("%d principles, want 17", next-1)
	}
}

func TestCosoSignalsRowOwnBranch(t *testing.T) {
	db := openSignalsDB(t)
	dept := uuid.New()
	jkt, sby := uuid.New(), uuid.New()
	inSby := addRisk(t, db, 3, 3, &sby)
	noRisk := addRCM(t, db, &dept, 2026, nil, 90)   // no linked risk, own branch Jakarta
	moved := addRCM(t, db, &dept, 2026, &inSby, 70) // risk in Surabaya, own branch Jakarta
	addRCM(t, db, &dept, 2026, &inSby, 50)          // follows its risk: Surabaya
	addRCM(t, db, &dept, 2026, nil, 10)             // no branch at all
	for _, id := range []uuid.UUID{noRisk, moved} {
		if err := db.Model(&models.RiskControlMatrix{}).Where("id = ?", id).Update("location_id", jkt).Error; err != nil {
			t.Fatal(err)
		}
	}
	locs := signalLocations{locs: []masterclient.Location{{ID: jkt, Name: "Jakarta", IsActive: true}, {ID: sby, Name: "Surabaya", IsActive: true}}}

	res, err := ComputeCosoSignals(context.Background(), db, locs, CosoSignalsParams{DepartmentID: dept, Year: 2026, LocationID: &jkt, AsOf: date(2026, 8, 1)})
	if err != nil {
		t.Fatal(err)
	}
	if res.RCMCount != 2 || res.RiskCount != 1 {
		t.Fatalf("jakarta C=%d R=%d, want 2 and 1", res.RCMCount, res.RiskCount)
	}
	if s := signal(t, res, SignalCARCMEffectiveness); *s.ValuePct != 80 {
		t.Errorf("jakarta ca = %v, want 80", *s.ValuePct)
	}
	res, err = ComputeCosoSignals(context.Background(), db, locs, CosoSignalsParams{DepartmentID: dept, Year: 2026, LocationID: &sby, AsOf: date(2026, 8, 1)})
	if err != nil {
		t.Fatal(err)
	}
	if res.RCMCount != 1 {
		t.Errorf("surabaya C=%d, want 1 (own branch overrides the risk's)", res.RCMCount)
	}
}

func TestCosoSignalsLegacyDepartmentName(t *testing.T) {
	db := openSignalsDB(t)
	dept := uuid.New()
	addRCM(t, db, &dept, 2026, nil, 80)
	legacy := addRCM(t, db, nil, 2026, nil, 60)
	other := addRCM(t, db, nil, 2026, nil, 10)
	lastYear := addRCM(t, db, nil, 2025, nil, 10)
	set := func(id uuid.UUID, name string) {
		if err := db.Model(&models.RiskControlMatrix{}).Where("id = ?", id).Update("department", name).Error; err != nil {
			t.Fatal(err)
		}
	}
	set(legacy, "  finance   & Accounting ")
	set(other, "Operations")
	set(lastYear, "Finance & Accounting")

	res, err := ComputeCosoSignals(context.Background(), db, nil, CosoSignalsParams{DepartmentID: dept, Year: 2026, AsOf: date(2026, 8, 1), LegacyDepartmentName: "Finance & Accounting"})
	if err != nil {
		t.Fatal(err)
	}
	if res.RCMCount != 2 || res.LegacyNameMatchedRCMCount != 1 || res.UnassignedRCMCount != 2 {
		t.Fatalf("C=%d legacy=%d unassigned=%d, want 2/1/2", res.RCMCount, res.LegacyNameMatchedRCMCount, res.UnassignedRCMCount)
	}
	if s := signal(t, res, SignalCARCMEffectiveness); *s.ValuePct != 70 {
		t.Errorf("ca = %v, want 70", *s.ValuePct)
	}

	// No unique master name: legacy rows stay out.
	res, err = ComputeCosoSignals(context.Background(), db, nil, CosoSignalsParams{DepartmentID: dept, Year: 2026, AsOf: date(2026, 8, 1)})
	if err != nil {
		t.Fatal(err)
	}
	if res.RCMCount != 1 || res.LegacyNameMatchedRCMCount != 0 {
		t.Errorf("without name C=%d legacy=%d, want 1/0", res.RCMCount, res.LegacyNameMatchedRCMCount)
	}
}
