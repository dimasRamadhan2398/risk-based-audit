package annual_plan_sync

import (
	"context"
	"errors"
	"testing"

	"audit-service/models"
	"audit-service/pkg/masterclient"
	"audit-service/pkg/riskclient"

	"gorm.io/gorm"
)

type fakeDepts struct {
	list []masterclient.Department
	err  error
	auth string
}

func (f *fakeDepts) ListDepartments(_ context.Context, authorization string) ([]masterclient.Department, error) {
	f.auth = authorization
	return f.list, f.err
}

type fakeUniverse struct {
	tree     []riskclient.Entity
	treeErr  error
	years    map[int][]riskclient.YearRow
	yearErr  error
	asked    []int
	authSeen []string
}

func (f *fakeUniverse) ListCorporateEntities(_ context.Context, authorization string) ([]riskclient.Entity, error) {
	f.authSeen = append(f.authSeen, authorization)
	return f.tree, f.treeErr
}

func (f *fakeUniverse) ListYearUniverse(_ context.Context, authorization string, year int) ([]riskclient.YearRow, error) {
	f.authSeen = append(f.authSeen, authorization)
	f.asked = append(f.asked, year)
	if f.yearErr != nil {
		return nil, f.yearErr
	}
	return f.years[year], nil
}

func refRow(t *testing.T, db *gorm.DB, kind, id string, year int) (models.AnnualSnapshotRef, bool) {
	t.Helper()
	var r models.AnnualSnapshotRef
	err := db.Where("kind = ? AND ref_id = ? AND year = ?", kind, normRefID(id), year).First(&r).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r, false
	}
	if err != nil {
		t.Fatal(err)
	}
	return r, true
}

// The situation found locally: plans carry valid ids, the lookup has no
// department / entity rows, so the plain backfill fills nothing. Reconciling
// from the sources fills the plans.
func TestReconcileFillsPlansWhenLookupWasEmpty(t *testing.T) {
	db := testDB(t)
	plan := seedPlan(t, db, 2026,
		models.AnnualAuditActivity{DepartmentID: "DEP-1", AuditUniverseID: "uni-sub"},
		models.AnnualAuditActivity{DepartmentID: "dep-2", Department: "Kept", AuditUniverseID: "uni-root", RiskName: "Kept too", RiskLevel: "Low"},
	)

	if res, err := BackfillSnapshots(context.Background(), db); err != nil || res.FieldsFilled != 0 {
		t.Fatalf("precondition: plain backfill = %+v, %v; want nothing filled", res, err)
	}

	st := SourceState{
		Departments: []DepartmentData{{ID: "dep-1", DepartmentName: " Finance "}, {ID: "dep-2", DepartmentName: "IT"}},
		Entities:    []EntityData{{ID: "uni-root", Name: "Treasury"}, {ID: "uni-sub", Name: "Cash Management"}},
		YearScores:  []YearScoreData{{CorporateAuditUniverseID: "uni-sub", Year: 2026, RiskLevel: "High"}},
	}
	res, err := Reconcile(context.Background(), db, st)
	if err != nil {
		t.Fatal(err)
	}
	if res.Refs.Departments != 2 || res.Refs.Entities != 2 || res.Refs.YearScores != 1 || res.Refs.Skipped != 0 {
		t.Fatalf("refs = %+v", res.Refs)
	}
	if res.Backfill.PlansChanged != 1 || res.Backfill.FieldsFilled != 3 {
		t.Fatalf("backfill = %+v, want 1 plan / 3 fields", res.Backfill)
	}

	got := reload(t, db, plan.ID.String())
	first, second := got.Activities[0], got.Activities[1]
	if first.Department != "Finance" || first.RiskName != "Cash Management" || first.RiskLevel != "High" {
		t.Fatalf("first row not filled: %+v", first)
	}
	if second.Department != "Kept" || second.RiskName != "Kept too" || second.RiskLevel != "Low" {
		t.Fatalf("filled fields were overwritten: %+v", second)
	}

	// Idempotent: a second run writes the same refs and fills nothing.
	again, err := Reconcile(context.Background(), db, st)
	if err != nil {
		t.Fatal(err)
	}
	if again.Backfill.PlansChanged != 0 || again.Backfill.FieldsFilled != 0 {
		t.Fatalf("second run changed plans: %+v", again.Backfill)
	}
	counts, err := RefCounts(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if counts[models.SnapshotRefDepartment] != 2 || counts[models.SnapshotRefEntity] != 2 || counts[models.SnapshotRefYearScore] != 1 {
		t.Fatalf("ref counts after two runs = %v (duplicates?)", counts)
	}
}

// Refs follow the event rules: the source's current name replaces an old one
// (as a department.updated would), a blank name or level never replaces a
// value, and records without an id are skipped.
func TestRefreshRefsFollowsEventSemantics(t *testing.T) {
	db := testDB(t)
	seedRefs(t, db,
		dept("dep-1", "Old Finance"),
		entity("uni-1", "Keep me"),
		yearScore("uni-1", 2026, "Medium"),
	)

	res, err := RefreshRefs(context.Background(), db, SourceState{
		Departments: []DepartmentData{{ID: "DEP-1", DepartmentName: "Finance"}, {ID: "", DepartmentName: "No id"}},
		Entities:    []EntityData{{ID: "uni-1", Name: "  "}},
		YearScores: []YearScoreData{
			{CorporateAuditUniverseID: "uni-1", Year: 2026, RiskLevel: ""}, // unscored
			{CorporateAuditUniverseID: "uni-1", Year: 0, RiskLevel: "High"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Departments != 1 || res.Entities != 0 || res.YearScores != 0 || res.Skipped != 4 {
		t.Fatalf("result = %+v", res)
	}
	if r, _ := refRow(t, db, models.SnapshotRefDepartment, "dep-1", 0); r.Name != "Finance" {
		t.Fatalf("department name = %q, want the source's current name", r.Name)
	}
	if r, _ := refRow(t, db, models.SnapshotRefEntity, "uni-1", 0); r.Name != "Keep me" {
		t.Fatalf("blank source name replaced the entity name: %q", r.Name)
	}
	if r, _ := refRow(t, db, models.SnapshotRefYearScore, "uni-1", 2026); r.RiskLevel != "Medium" {
		t.Fatalf("unscored source row replaced the level: %q", r.RiskLevel)
	}
}

// A dry run (rolled-back transaction) writes neither refs nor plans.
func TestReconcileDryRunWritesNothing(t *testing.T) {
	db := testDB(t)
	plan := seedPlan(t, db, 2026, models.AnnualAuditActivity{DepartmentID: "dep-1"})

	errDryRun := errors.New("dry run")
	var res ReconcileResult
	err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		res, err = Reconcile(context.Background(), tx, SourceState{Departments: []DepartmentData{{ID: "dep-1", DepartmentName: "Finance"}}})
		if err != nil {
			return err
		}
		return errDryRun
	})
	if !errors.Is(err, errDryRun) {
		t.Fatalf("transaction error = %v", err)
	}
	if res.Refs.Departments != 1 || res.Backfill.FieldsFilled != 1 {
		t.Fatalf("dry run counts = %+v", res)
	}
	if _, ok := refRow(t, db, models.SnapshotRefDepartment, "dep-1", 0); ok {
		t.Fatal("dry run wrote a ref")
	}
	if got := reload(t, db, plan.ID.String()).Activities[0].Department; got != "" {
		t.Fatalf("dry run wrote Department = %q", got)
	}
}

func TestPlanYearsDistinctLiveAndPositive(t *testing.T) {
	db := testDB(t)
	seedPlan(t, db, 2027)
	seedPlan(t, db, 2026)
	seedPlan(t, db, 2026)
	seedPlan(t, db, 0)
	deleted := seedPlan(t, db, 2030)
	if err := db.Delete(&deleted).Error; err != nil {
		t.Fatal(err)
	}

	years, err := PlanYears(context.Background(), db)
	if err != nil {
		t.Fatal(err)
	}
	if len(years) != 2 || years[0] != 2026 || years[1] != 2027 {
		t.Fatalf("years = %v, want [2026 2027]", years)
	}
}

func TestLoadSourceStateMergesTreeAndYearRows(t *testing.T) {
	parent := "uni-root"
	depts := &fakeDepts{list: []masterclient.Department{{ID: "dep-1", DepartmentCode: "D1", DepartmentName: "Finance", CompanyID: "co-1"}}}
	universe := &fakeUniverse{
		tree: []riskclient.Entity{
			{ID: "uni-root", Name: "Treasury"},
			{ID: "uni-sub", Name: "", ParentID: &parent},
		},
		years: map[int][]riskclient.YearRow{
			2026: {
				// Fills the blank tree name of uni-sub.
				{CorporateAuditUniverseID: "uni-sub", Year: 2026, RiskLevel: "High", CorporateAuditUniverse: &riskclient.Entity{ID: "uni-sub", Name: "Cash"}},
				// Deeper than the tree preload: only known from the year row.
				{CorporateAuditUniverseID: "uni-deep", RiskLevel: "", CorporateAuditUniverse: &riskclient.Entity{Name: "Petty Cash"}},
			},
		},
	}

	st, err := LoadSourceState(context.Background(), depts, universe, "Bearer t", []int{2026})
	if err != nil {
		t.Fatal(err)
	}
	if depts.auth != "Bearer t" {
		t.Fatalf("master auth = %q", depts.auth)
	}
	for _, a := range universe.authSeen {
		if a != "Bearer t" {
			t.Fatalf("risk auth = %q", a)
		}
	}
	if len(st.Departments) != 1 || st.Departments[0].DepartmentName != "Finance" || st.Departments[0].ID != "dep-1" {
		t.Fatalf("departments = %+v", st.Departments)
	}
	names := map[string]string{}
	for _, e := range st.Entities {
		names[e.ID] = e.Name
	}
	if len(st.Entities) != 3 || names["uni-root"] != "Treasury" || names["uni-sub"] != "Cash" || names["uni-deep"] != "Petty Cash" {
		t.Fatalf("entities = %+v", st.Entities)
	}
	if len(st.YearScores) != 2 || st.YearScores[1].Year != 2026 {
		t.Fatalf("year scores = %+v (a row without year must take the requested year)", st.YearScores)
	}
}

func TestLoadSourceStateFailsWholeLoadOnAnyError(t *testing.T) {
	boom := errors.New("401")
	cases := map[string]struct {
		d *fakeDepts
		u *fakeUniverse
	}{
		"master": {&fakeDepts{err: boom}, &fakeUniverse{}},
		"tree":   {&fakeDepts{}, &fakeUniverse{treeErr: boom}},
		"year":   {&fakeDepts{}, &fakeUniverse{yearErr: boom}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			st, err := LoadSourceState(context.Background(), tc.d, tc.u, "Bearer t", []int{2026})
			if !errors.Is(err, boom) {
				t.Fatalf("err = %v, want wrapped source error", err)
			}
			if len(st.Departments)+len(st.Entities)+len(st.YearScores) != 0 {
				t.Fatalf("partial state returned: %+v", st)
			}
		})
	}
}
