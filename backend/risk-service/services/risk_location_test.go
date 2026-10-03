package services

import (
	"testing"

	"github.com/google/uuid"

	"risk-service/models"
)

func TestResolveBranchNamePrefersLocationName(t *testing.T) {
	profile := models.RiskProfile{
		LocationName: "Medan Branch",
		// Legacy sentinel says "Head Office"; the location must win.
		DepartmentID: uuid.MustParse("00000000-0000-0000-0000-000000000001"),
	}

	if got := resolveBranchName(profile); got != "Medan Branch" {
		t.Errorf("resolveBranchName() = %q, want %q", got, "Medan Branch")
	}
}

func TestResolveBranchNameFallsBackToLegacySentinel(t *testing.T) {
	// Rows written before location_id existed carry the branch in DepartmentID.
	profile := models.RiskProfile{
		DepartmentID: uuid.MustParse("00000000-0000-0000-0000-000000000003"),
	}

	if got := resolveBranchName(profile); got != "Surabaya Branch" {
		t.Errorf("resolveBranchName() = %q, want %q", got, "Surabaya Branch")
	}
}

func TestResolveBranchNameFallsBackToDefault(t *testing.T) {
	// A real department UUID is not a branch; without a location we can only
	// report the default.
	profile := models.RiskProfile{DepartmentID: uuid.New()}

	if got := resolveBranchName(profile); got != defaultBranchName {
		t.Errorf("resolveBranchName() = %q, want %q", got, defaultBranchName)
	}
}

func TestApplyLocationAcceptsAnyBranchName(t *testing.T) {
	// The old code silently relabelled unknown branches as "Head Office";
	// any location from the master data must now round-trip.
	profile := models.RiskProfile{}
	applyLocation(&profile, &RiskRequest{Branch: "Medan Branch"})

	if profile.LocationName != "Medan Branch" {
		t.Errorf("LocationName = %q, want %q", profile.LocationName, "Medan Branch")
	}
	if got := resolveBranchName(profile); got != "Medan Branch" {
		t.Errorf("resolveBranchName() = %q, want %q", got, "Medan Branch")
	}
}

func TestApplyLocationStoresLocationID(t *testing.T) {
	locID := uuid.New()
	profile := models.RiskProfile{}
	applyLocation(&profile, &RiskRequest{Branch: "Bali Branch", LocationID: locID.String()})

	if profile.LocationID == nil {
		t.Fatal("LocationID was not set")
	}
	if *profile.LocationID != locID {
		t.Errorf("LocationID = %v, want %v", *profile.LocationID, locID)
	}
	if got := locationIDString(profile); got != locID.String() {
		t.Errorf("locationIDString() = %q, want %q", got, locID.String())
	}
}

func TestApplyLocationIgnoresMalformedLocationID(t *testing.T) {
	profile := models.RiskProfile{}
	applyLocation(&profile, &RiskRequest{Branch: "Bali Branch", LocationID: "not-a-uuid"})

	if profile.LocationID != nil {
		t.Errorf("LocationID = %v, want nil", *profile.LocationID)
	}
	// The branch name still applies.
	if profile.LocationName != "Bali Branch" {
		t.Errorf("LocationName = %q, want %q", profile.LocationName, "Bali Branch")
	}
}

func TestApplyLocationKeepsLegacySentinelForKnownBranches(t *testing.T) {
	// Older readers still derive the branch from DepartmentID, so known names
	// must keep writing their sentinel.
	profile := models.RiskProfile{}
	applyLocation(&profile, &RiskRequest{Branch: "Bandung Branch"})

	want := "00000000-0000-0000-0000-000000000004"
	if profile.DepartmentID.String() != want {
		t.Errorf("DepartmentID = %q, want %q", profile.DepartmentID.String(), want)
	}
}

func TestApplyLocationLeavesDepartmentAloneForNewBranches(t *testing.T) {
	// An unknown branch has no sentinel; it must not clobber a real department.
	deptID := uuid.New()
	profile := models.RiskProfile{DepartmentID: deptID}
	applyLocation(&profile, &RiskRequest{Branch: "Medan Branch"})

	if profile.DepartmentID != deptID {
		t.Errorf("DepartmentID = %v, want it unchanged (%v)", profile.DepartmentID, deptID)
	}
}

func TestApplyLocationWithEmptyRequestKeepsExistingLocation(t *testing.T) {
	locID := uuid.New()
	profile := models.RiskProfile{LocationID: &locID, LocationName: "Bali Branch"}
	applyLocation(&profile, &RiskRequest{})

	if profile.LocationName != "Bali Branch" {
		t.Errorf("LocationName = %q, want it unchanged", profile.LocationName)
	}
	if profile.LocationID == nil || *profile.LocationID != locID {
		t.Error("LocationID should not be cleared by a request that omits it")
	}
}

func TestLegacyMapsAreConsistent(t *testing.T) {
	for name, id := range legacyBranchToUUID {
		if back, ok := legacyUUIDToBranch[id]; !ok || back != name {
			t.Errorf("legacy maps disagree for %q: %q -> %q", name, id, back)
		}
	}
	if len(legacyBranchToUUID) != len(legacyUUIDToBranch) {
		t.Errorf("legacy maps differ in size: %d vs %d", len(legacyBranchToUUID), len(legacyUUIDToBranch))
	}
}
