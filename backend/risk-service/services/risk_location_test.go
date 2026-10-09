package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"risk-service/models"
	"risk-service/pkg/masterclient"

	"github.com/google/uuid"
)

// Production Location master rows (master-service, 2026-10-08).
var (
	headOffice = masterclient.Location{ID: uuid.MustParse("7d1000c0-0b66-468b-978d-09e32c8d9df6"), Name: "Head Office", IsActive: true}
	surabaya   = masterclient.Location{ID: uuid.MustParse("3430b58b-2377-4992-a32d-48914a7eb470"), Name: "Surabaya Branch", IsActive: true}
	bandung    = masterclient.Location{ID: uuid.MustParse("47a2489b-48ed-4e3d-aa4d-9ec661cadae6"), Name: "Bandung Branch", IsActive: true}
	jakarta    = masterclient.Location{ID: uuid.MustParse("02643243-5bf1-4f02-9c43-67c200708068"), Name: "Jakarta Branch", IsActive: true}
	bali       = masterclient.Location{ID: uuid.MustParse("4c875bfc-3cfb-4066-ada2-757af0ef46a2"), Name: "Bali Branch", IsActive: true}

	prodLocations = []masterclient.Location{headOffice, surabaya, bandung, jakarta, bali}
)

func sentinel(n int) uuid.UUID {
	return uuid.MustParse([]string{
		"",
		"00000000-0000-0000-0000-000000000001",
		"00000000-0000-0000-0000-000000000002",
		"00000000-0000-0000-0000-000000000003",
		"00000000-0000-0000-0000-000000000004",
		"00000000-0000-0000-0000-000000000005",
	}[n])
}

// fakeLocations is a LocationSource; err simulates master-service being down.
type fakeLocations struct {
	locs  []masterclient.Location
	err   error
	calls int
}

func (f *fakeLocations) ListLocations(context.Context) ([]masterclient.Location, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.locs, nil
}

var errDown = errors.New("dial tcp: connection refused")

func deref(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func assertBranch(t *testing.T, gotID, gotBranch *string, wantID, wantBranch string) {
	t.Helper()
	if deref(gotID) != wantID || deref(gotBranch) != wantBranch {
		t.Errorf("got (location_id=%s, branch=%s), want (%s, %s)", deref(gotID), deref(gotBranch), wantID, wantBranch)
	}
}

func TestResolveBranch(t *testing.T) {
	idx := newLocationIndex(prodLocations)
	jakartaID := jakarta.ID
	dangling := uuid.New()

	tests := []struct {
		name       string
		profile    models.RiskProfile
		idx        *locationIndex
		wantID     string
		wantBranch string
	}{
		{
			name: "location_id path uses the master name, not the stored copy",
			profile: models.RiskProfile{
				LocationID:   &jakartaID,
				LocationName: "Old Jakarta Name",
				DepartmentID: sentinel(1), // legacy says Head Office; the link wins
			},
			idx: idx, wantID: jakarta.ID.String(), wantBranch: "Jakarta Branch",
		},
		{
			name:    "legacy sentinel path resolves through the master by name",
			profile: models.RiskProfile{DepartmentID: sentinel(3)},
			idx:     idx, wantID: surabaya.ID.String(), wantBranch: "Surabaya Branch",
		},
		{
			name:    "every production sentinel maps to its master row",
			profile: models.RiskProfile{DepartmentID: sentinel(5)},
			idx:     idx, wantID: bali.ID.String(), wantBranch: "Bali Branch",
		},
		{
			name:    "free-text location_name matching a master row (case/space-insensitive)",
			profile: models.RiskProfile{LocationName: "  bandung   BRANCH "},
			idx:     idx, wantID: bandung.ID.String(), wantBranch: "Bandung Branch",
		},
		{
			name:    "no location and a real department is unlinked, not Head Office",
			profile: models.RiskProfile{DepartmentID: uuid.New()},
			idx:     idx, wantID: "<nil>", wantBranch: "<nil>",
		},
		{
			name:    "nil department is unlinked",
			profile: models.RiskProfile{},
			idx:     idx, wantID: "<nil>", wantBranch: "<nil>",
		},
		{
			name:    "free-text name that is not a registered location is unlinked",
			profile: models.RiskProfile{LocationName: "Medan Branch"},
			idx:     idx, wantID: "<nil>", wantBranch: "<nil>",
		},
		{
			name:    "sentinel whose name is not in the master is unlinked",
			profile: models.RiskProfile{DepartmentID: sentinel(5)},
			idx:     newLocationIndex([]masterclient.Location{headOffice}),
			wantID:  "<nil>", wantBranch: "<nil>",
		},
		{
			name:    "location_id no longer registered is unlinked",
			profile: models.RiskProfile{LocationID: &dangling, LocationName: "Deleted Branch"},
			idx:     idx, wantID: "<nil>", wantBranch: "<nil>",
		},
		{
			name:    "master unreachable: legacy row gets null, nothing invented",
			profile: models.RiskProfile{DepartmentID: sentinel(2), LocationName: "Jakarta Branch"},
			idx:     nil, wantID: "<nil>", wantBranch: "<nil>",
		},
		{
			name:    "master unreachable: stored location_id kept, branch null",
			profile: models.RiskProfile{LocationID: &jakartaID, LocationName: "Jakarta Branch"},
			idx:     nil, wantID: jakarta.ID.String(), wantBranch: "<nil>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, branch := resolveBranch(tt.profile, tt.idx)
			assertBranch(t, id, branch, tt.wantID, tt.wantBranch)
		})
	}
}

func TestLocationIndexPrefersActiveOnDuplicateNames(t *testing.T) {
	inactive := masterclient.Location{ID: uuid.New(), Name: "Bali Branch", IsActive: false}
	idx := newLocationIndex([]masterclient.Location{inactive, bali})

	l, ok := idx.findByName("Bali Branch")
	if !ok || l.ID != bali.ID {
		t.Fatalf("findByName = %v, %v; want the active Bali row", l.ID, ok)
	}
	// The inactive row is still resolvable by ID for risks already linked to it.
	if _, ok := idx.byID[inactive.ID]; !ok {
		t.Error("inactive location must stay resolvable by ID")
	}
}

func TestGetAllResolvesProductionShape(t *testing.T) {
	repo := newFakeRepo()
	repo.addRisk("Fraud", models.RiskProfile{DepartmentID: sentinel(1)})
	repo.addRisk("Vendor", models.RiskProfile{DepartmentID: sentinel(2)})
	repo.addRisk("Unlinked", models.RiskProfile{DepartmentID: uuid.Nil})

	svc := NewRiskService(repo, &fakeLocations{locs: prodLocations})
	data, err := svc.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	got := map[string]RiskResponse{}
	for _, r := range data {
		got[r.Name] = r
	}
	assertBranch(t, got["Fraud"].LocationID, got["Fraud"].Branch, headOffice.ID.String(), "Head Office")
	assertBranch(t, got["Vendor"].LocationID, got["Vendor"].Branch, jakarta.ID.String(), "Jakarta Branch")
	assertBranch(t, got["Unlinked"].LocationID, got["Unlinked"].Branch, "<nil>", "<nil>")

	// Read-time resolution must not modify data.
	if repo.profileSaves != 0 {
		t.Errorf("GetAll saved %d profiles; reads must not write", repo.profileSaves)
	}
}

func TestGetAllMasterUnreachableReturnsNullBranches(t *testing.T) {
	repo := newFakeRepo()
	repo.addRisk("Fraud", models.RiskProfile{DepartmentID: sentinel(1)})
	repo.addRisk("Vendor", models.RiskProfile{DepartmentID: sentinel(2)})

	svc := NewRiskService(repo, &fakeLocations{err: errDown})
	data, err := svc.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll must not fail when master-service is down: %v", err)
	}
	if len(data) != 2 {
		t.Fatalf("got %d risks, want 2", len(data))
	}
	for _, r := range data {
		assertBranch(t, r.LocationID, r.Branch, "<nil>", "<nil>")
	}
}

func TestRiskResponseSerializesUnlinkedAsNull(t *testing.T) {
	b, err := json.Marshal(RiskResponse{ID: "x", Assessments: []RiskAssessmentRes{}})
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"location_id", "branch"} {
		v, present := m[key]
		if !present || v != nil {
			t.Errorf("%s = %v (present=%v), want explicit null", key, v, present)
		}
	}
}

func TestCreateWithLocationID(t *testing.T) {
	repo := newFakeRepo()
	svc := NewRiskService(repo, &fakeLocations{locs: prodLocations})

	res, err := svc.Create(context.Background(), &RiskRequest{Name: "R", LocationID: bali.ID.String(), Branch: "Head Office"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	// location_id wins over a conflicting branch name.
	assertBranch(t, res.LocationID, res.Branch, bali.ID.String(), "Bali Branch")

	p := repo.lastProfile
	if p.LocationID == nil || *p.LocationID != bali.ID || p.LocationName != "Bali Branch" {
		t.Errorf("stored profile location = (%v, %q), want Bali", p.LocationID, p.LocationName)
	}
	if p.DepartmentID != uuid.Nil {
		t.Errorf("DepartmentID = %v; new writes must not store legacy sentinels", p.DepartmentID)
	}
}

func TestCreateWithBranchNameOnly(t *testing.T) {
	repo := newFakeRepo()
	svc := NewRiskService(repo, &fakeLocations{locs: prodLocations})

	res, err := svc.Create(context.Background(), &RiskRequest{Name: "R", Branch: "surabaya branch"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	assertBranch(t, res.LocationID, res.Branch, surabaya.ID.String(), "Surabaya Branch")
	if repo.lastProfile.LocationName != "Surabaya Branch" {
		t.Errorf("stored location_name = %q, want the master name", repo.lastProfile.LocationName)
	}
}

func TestCreateWithoutLocationIsUnlinked(t *testing.T) {
	locs := &fakeLocations{locs: prodLocations}
	svc := NewRiskService(newFakeRepo(), locs)

	res, err := svc.Create(context.Background(), &RiskRequest{Name: "R"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	assertBranch(t, res.LocationID, res.Branch, "<nil>", "<nil>")
	if locs.calls != 0 {
		t.Errorf("master-service called %d times for a risk with no location", locs.calls)
	}
}

func TestCreateRejectsUnregisteredLocations(t *testing.T) {
	tests := []struct {
		name     string
		req      RiskRequest
		wantCode string
	}{
		{"unknown branch name", RiskRequest{Name: "R", Branch: "Medan Branch"}, "UNKNOWN_LOCATION"},
		{"unknown location_id", RiskRequest{Name: "R", LocationID: uuid.NewString()}, "UNKNOWN_LOCATION"},
		{"malformed location_id", RiskRequest{Name: "R", LocationID: "not-a-uuid"}, "INVALID_LOCATION_ID"},
		{"nil location_id", RiskRequest{Name: "R", LocationID: uuid.Nil.String()}, "INVALID_LOCATION_ID"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeRepo()
			svc := NewRiskService(repo, &fakeLocations{locs: prodLocations})

			_, err := svc.Create(context.Background(), &tt.req)
			var locErr *LocationError
			if !errors.As(err, &locErr) || locErr.Code != tt.wantCode {
				t.Fatalf("err = %v, want LocationError %s", err, tt.wantCode)
			}
			if repo.writes() != 0 {
				t.Errorf("%d rows written for a rejected request", repo.writes())
			}
		})
	}
}

func TestCreateWithBranchWhenMasterUnreachable(t *testing.T) {
	repo := newFakeRepo()
	svc := NewRiskService(repo, &fakeLocations{err: errDown})

	_, err := svc.Create(context.Background(), &RiskRequest{Name: "R", Branch: "Bali Branch"})
	if !errors.Is(err, ErrLocationsUnavailable) {
		t.Fatalf("err = %v, want ErrLocationsUnavailable", err)
	}
	if repo.writes() != 0 {
		t.Errorf("%d rows written while the branch could not be validated", repo.writes())
	}
}

func TestUpdateLinksLegacyRow(t *testing.T) {
	repo := newFakeRepo()
	id := repo.addRisk("Legacy", models.RiskProfile{DepartmentID: sentinel(4)})
	svc := NewRiskService(repo, &fakeLocations{locs: prodLocations})

	res, err := svc.Update(context.Background(), id, &RiskRequest{Name: "Legacy", LocationID: jakarta.ID.String()})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	assertBranch(t, res.LocationID, res.Branch, jakarta.ID.String(), "Jakarta Branch")
	p := repo.registers[id].Profile
	if p.LocationID == nil || *p.LocationID != jakarta.ID {
		t.Errorf("stored location_id = %v, want Jakarta", p.LocationID)
	}
}

func TestUpdateWithoutLocationKeepsLegacyResolution(t *testing.T) {
	repo := newFakeRepo()
	id := repo.addRisk("Legacy", models.RiskProfile{DepartmentID: sentinel(4)})
	svc := NewRiskService(repo, &fakeLocations{locs: prodLocations})

	res, err := svc.Update(context.Background(), id, &RiskRequest{Name: "Legacy renamed"})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	assertBranch(t, res.LocationID, res.Branch, bandung.ID.String(), "Bandung Branch")
	if repo.registers[id].Profile.LocationID != nil {
		t.Error("an update without location fields must not link the row")
	}
}

func TestUpdateWithUnchangedLocationIDWorksWhileMasterDown(t *testing.T) {
	repo := newFakeRepo()
	baliID := bali.ID
	id := repo.addRisk("Linked", models.RiskProfile{LocationID: &baliID, LocationName: "Bali Branch"})
	svc := NewRiskService(repo, &fakeLocations{err: errDown})

	res, err := svc.Update(context.Background(), id, &RiskRequest{Name: "Linked", LocationID: bali.ID.String(), Severity: 9})
	if err != nil {
		t.Fatalf("editing a risk without changing its location must not need master-service: %v", err)
	}
	assertBranch(t, res.LocationID, res.Branch, bali.ID.String(), "<nil>")
}

func TestUpdateRejectsUnknownBranchBeforeWriting(t *testing.T) {
	repo := newFakeRepo()
	id := repo.addRisk("Legacy", models.RiskProfile{DepartmentID: sentinel(1)})
	svc := NewRiskService(repo, &fakeLocations{locs: prodLocations})
	before := repo.writes()

	_, err := svc.Update(context.Background(), id, &RiskRequest{Name: "X", Branch: "Medan Branch"})
	var locErr *LocationError
	if !errors.As(err, &locErr) {
		t.Fatalf("err = %v, want LocationError", err)
	}
	if repo.writes() != before {
		t.Error("rejected update must not write the register or profile")
	}
	if repo.registers[id].RiskEvent != "Legacy" {
		t.Error("register was modified by a rejected update")
	}
}

func TestPlanLocationBackfill(t *testing.T) {
	linked := surabaya.ID
	profiles := []models.RiskProfile{
		{ID: uuid.New(), DepartmentID: sentinel(1)},                            // -> Head Office
		{ID: uuid.New(), DepartmentID: sentinel(5)},                            // -> Bali
		{ID: uuid.New(), LocationName: "jakarta branch"},                       // free text -> Jakarta
		{ID: uuid.New(), LocationID: &linked, LocationName: "Surabaya Branch"}, // already linked: skip
		{ID: uuid.New(), DepartmentID: uuid.New()},                             // never had a branch: skip
		{ID: uuid.New(), LocationName: "Medan Branch"},                         // unregistered: miss
	}

	plan, misses := PlanLocationBackfill(profiles, prodLocations)
	if len(plan) != 3 {
		t.Fatalf("plan has %d entries, want 3: %+v", len(plan), plan)
	}
	want := []struct {
		loc    masterclient.Location
		source string
	}{
		{headOffice, "legacy_department_id"},
		{bali, "legacy_department_id"},
		{jakarta, "location_name"},
	}
	for i, w := range want {
		if plan[i].ProfileID != profiles[i].ID || plan[i].LocationID != w.loc.ID || plan[i].LocationName != w.loc.Name || plan[i].Source != w.source {
			t.Errorf("plan[%d] = %+v, want %s via %s", i, plan[i], w.loc.Name, w.source)
		}
	}
	if len(misses) != 1 || misses[0].ProfileID != profiles[5].ID {
		t.Errorf("misses = %+v, want only the Medan profile", misses)
	}

	// Idempotent: once applied, nothing is left to do.
	for i := range plan {
		id := plan[i].LocationID
		profiles[i].LocationID = &id
		profiles[i].LocationName = plan[i].LocationName
	}
	again, _ := PlanLocationBackfill(profiles, prodLocations)
	if len(again) != 0 {
		t.Errorf("second run planned %d updates, want 0", len(again))
	}
}
