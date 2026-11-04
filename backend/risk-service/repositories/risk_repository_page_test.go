package repositories

import (
	"fmt"
	"testing"

	"risk-service/models"
	"risk-service/pkg/sqlitetest"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type seedRisk struct {
	name, category, desc string
	score                int
	loc                  *uuid.UUID
	years                []int
	q1Impact, q1Like     int
	snapCategory         string
}

func seed(t *testing.T, db *gorm.DB, risks []seedRisk) map[string]uuid.UUID {
	t.Helper()
	ids := map[string]uuid.UUID{}
	for _, r := range risks {
		p := models.RiskProfile{ID: uuid.New(), Category: r.category, Description: r.desc, LocationID: r.loc}
		reg := models.RiskRegister{ID: uuid.New(), ProfileID: p.ID, RiskEvent: r.name, InherentScore: r.score}
		if err := db.Create(&p).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Omit("Profile", "Assessments").Create(&reg).Error; err != nil {
			t.Fatal(err)
		}
		for _, y := range r.years {
			a := models.RiskAssessment{ID: uuid.New(), RiskRegisterID: reg.ID, Year: y,
				ImpactQ1: r.q1Impact, LikelihoodQ1: r.q1Like, RiskEvent: r.name, Category: r.snapCategory}
			if err := db.Create(&a).Error; err != nil {
				t.Fatal(err)
			}
		}
		ids[r.name] = reg.ID
	}
	return ids
}

func newRepo(t *testing.T) (*riskRepository, *gorm.DB) {
	db := sqlitetest.Open(t, &models.RiskProfile{}, &models.RiskRegister{}, &models.RiskAssessment{})
	return &riskRepository{db: db}, db
}

func names(regs []models.RiskRegister) []string {
	out := make([]string, len(regs))
	for i, r := range regs {
		out[i] = r.RiskEvent
	}
	return out
}

func TestFindPageBoundariesAndStableOrder(t *testing.T) {
	repo, db := newRepo(t)
	var rs []seedRisk
	for i := 1; i <= 25; i++ {
		rs = append(rs, seedRisk{name: fmt.Sprintf("R%02d", i), category: "Ops", score: i, years: []int{2026}})
	}
	seed(t, db, rs)

	page := func(p, size int) ([]models.RiskRegister, int64) {
		got, total, err := repo.FindPage(RiskFilter{Paged: true, Page: p, PageSize: size})
		if err != nil {
			t.Fatal(err)
		}
		return got, total
	}

	got, total := page(1, 10)
	if total != 25 || len(got) != 10 || got[0].RiskEvent != "R25" || got[9].RiskEvent != "R16" {
		t.Fatalf("page1: total=%d names=%v", total, names(got))
	}
	if len(got[0].Assessments) != 1 || got[0].Profile.Category != "Ops" {
		t.Errorf("preloads missing: %+v", got[0])
	}
	got, _ = page(3, 10)
	if len(got) != 5 || got[4].RiskEvent != "R01" {
		t.Errorf("last page: %v", names(got))
	}
	got, total = page(4, 10)
	if len(got) != 0 || total != 25 {
		t.Errorf("past end: len=%d total=%d", len(got), total)
	}

	// No overlap or gaps across pages.
	seen := map[string]bool{}
	for p := 1; p <= 3; p++ {
		g, _ := page(p, 10)
		for _, n := range names(g) {
			if seen[n] {
				t.Fatalf("%s repeated", n)
			}
			seen[n] = true
		}
	}
	if len(seen) != 25 {
		t.Errorf("covered %d", len(seen))
	}
}

func TestFindPageTieBreakIsStable(t *testing.T) {
	repo, db := newRepo(t)
	seed(t, db, []seedRisk{
		{name: "B", category: "c", score: 5}, {name: "A", category: "c", score: 5}, {name: "C", category: "c", score: 5},
	})
	got, _, err := repo.FindPage(RiskFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(names(got)) != "[A B C]" {
		t.Errorf("order = %v", names(got))
	}
}

func TestFindPageFilters(t *testing.T) {
	repo, db := newRepo(t)
	locA, locB := uuid.New(), uuid.New()
	seed(t, db, []seedRisk{
		{name: "Fraud", category: "Operational", desc: "100% loss_event", score: 9, loc: &locA, years: []int{2025, 2026}},
		{name: "Cyber", category: " operational ", desc: "phishing", score: 8, loc: &locB, years: []int{2026}},
		{name: "Credit", category: "Financial", desc: "default", score: 7, years: []int{2025}},
		// Category re-labelled in 2026: latest profile says Compliance, 2025 snapshot says Financial.
		{name: "Legacy", category: "Compliance", desc: "x", score: 6, years: []int{2025}, snapCategory: "Financial"},
	})
	count := func(f RiskFilter) (int64, []string) {
		got, total, err := repo.FindPage(f)
		if err != nil {
			t.Fatal(err)
		}
		return total, names(got)
	}
	page := func(f RiskFilter) RiskFilter { f.Paged, f.Page, f.PageSize = true, 1, 10; return f }

	if n, got := count(page(RiskFilter{Category: "  OPERATIONAL"})); n != 2 || fmt.Sprint(got) != "[Fraud Cyber]" {
		t.Errorf("category: %d %v", n, got)
	}
	if n, _ := count(page(RiskFilter{FiscalYear: 2025})); n != 3 {
		t.Errorf("year 2025: %d", n)
	}
	if n, got := count(page(RiskFilter{FiscalYear: 2025, Category: "financial"})); n != 2 || fmt.Sprint(got) != "[Credit Legacy]" {
		t.Errorf("year+snapshot category: %d %v", n, got)
	}
	if n, _ := count(page(RiskFilter{Category: "financial"})); n != 1 {
		t.Errorf("latest category: %d", n)
	}
	if n, got := count(page(RiskFilter{LocationID: locA.String()})); n != 1 || got[0] != "Fraud" {
		t.Errorf("branch: %d %v", n, got)
	}
	if n, _ := count(page(RiskFilter{Unassigned: true})); n != 2 {
		t.Errorf("unassigned: %d", n)
	}
	if n, got := count(page(RiskFilter{Search: "PHISH"})); n != 1 || got[0] != "Cyber" {
		t.Errorf("search: %d %v", n, got)
	}
	// Wildcards are literal.
	if n, _ := count(page(RiskFilter{Search: "%"})); n != 1 {
		t.Errorf("search %%: %d", n)
	}
	if n, _ := count(page(RiskFilter{Search: "_"})); n != 1 {
		t.Errorf("search _: %d", n)
	}
	if n, _ := count(page(RiskFilter{Search: "100% loss_"})); n != 1 {
		t.Errorf("search combined: %d", n)
	}
	if n, _ := count(page(RiskFilter{Search: "'; DROP TABLE risk_register;--"})); n != 0 {
		t.Errorf("injection: %d", n)
	}
	// Combined, and total counts the filter before the page is cut.
	f := RiskFilter{Category: "operational", FiscalYear: 2026, Paged: true, Page: 2, PageSize: 1}
	if got, total, _ := repo.FindPage(f); total != 2 || len(got) != 1 || got[0].RiskEvent != "Cyber" {
		t.Errorf("combined: total=%d %v", total, names(got))
	}
	// Unpaged returns everything matching.
	if n, got := count(RiskFilter{Category: "operational"}); n != 2 || len(got) != 2 {
		t.Errorf("unpaged: %d", n)
	}
}

func TestFindPageSortAllowlist(t *testing.T) {
	repo, db := newRepo(t)
	seed(t, db, []seedRisk{
		{name: "Low", category: "b", score: 1, years: []int{2026}, q1Impact: 1, q1Like: 1},
		{name: "High", category: "a", score: 2, years: []int{2026}, q1Impact: 5, q1Like: 5},
		{name: "Mid", category: "c", score: 3, years: []int{2026}, q1Impact: 3, q1Like: 2},
	})
	asc := false
	cases := []struct {
		f    RiskFilter
		want string
	}{
		{RiskFilter{SortBy: SortName}, "[High Low Mid]"},
		{RiskFilter{SortBy: SortName, SortDesc: boolp(true)}, "[Mid Low High]"},
		{RiskFilter{SortBy: SortCategory}, "[High Low Mid]"},
		{RiskFilter{SortBy: SortSeverity, SortDesc: &asc}, "[Low High Mid]"},
		{RiskFilter{SortBy: SortScore, FiscalYear: 2026, Period: "q1"}, "[High Mid Low]"},
		{RiskFilter{SortBy: SortScore}, "[Mid High Low]"}, // no year: falls back to severity
	}
	for _, c := range cases {
		got, _, err := repo.FindPage(c.f)
		if err != nil {
			t.Fatalf("%+v: %v", c.f, err)
		}
		if fmt.Sprint(names(got)) != c.want {
			t.Errorf("%+v: got %v want %s", c.f, names(got), c.want)
		}
	}
}

func boolp(b bool) *bool { return &b }

func TestDistinctCategories(t *testing.T) {
	repo, db := newRepo(t)
	seed(t, db, []seedRisk{
		{name: "1", category: "Operational", years: []int{2026}},
		{name: "2", category: " operational", years: []int{2025}},
		{name: "3", category: "Financial", years: []int{2025}, snapCategory: "Strategic"},
		{name: "4", category: "  "},
	})
	got, err := repo.DistinctCategories(0)
	if err != nil || fmt.Sprint(got) != "[Financial Operational]" {
		t.Errorf("all: %v %v", got, err)
	}
	got, err = repo.DistinctCategories(2025)
	if err != nil || fmt.Sprint(got) != "[operational Strategic]" {
		t.Errorf("2025: %v %v", got, err)
	}
}
