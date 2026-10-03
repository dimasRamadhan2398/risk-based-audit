package seeders

import "testing"

// crpBranches are the branch labels the Corporate Risk Profile and the Risk
// Control Matrix group risks by. Every one of them must exist as location master
// data, otherwise the branch filter offers values the data never contains (and
// vice versa).
var crpBranches = []string{
	"Head Office",
	"Jakarta Branch",
	"Surabaya Branch",
	"Bandung Branch",
	"Bali Branch",
}

func TestLocationSeedsCoverCRPBranches(t *testing.T) {
	seeded := make(map[string]bool, len(LocationSeeds))
	for _, loc := range LocationSeeds {
		seeded[loc.Name] = true
	}

	for _, branch := range crpBranches {
		if !seeded[branch] {
			t.Errorf("branch %q is used by the risk profile but has no location seed", branch)
		}
	}
}

func TestLocationSeedNamesAreUnique(t *testing.T) {
	// SeedLocations upserts on name, so duplicates would silently collapse.
	seen := make(map[string]bool, len(LocationSeeds))
	for _, loc := range LocationSeeds {
		if seen[loc.Name] {
			t.Errorf("duplicate location seed name %q", loc.Name)
		}
		seen[loc.Name] = true
	}
}

func TestLocationSeedsHaveRequiredFields(t *testing.T) {
	// Name, Address and City are NOT NULL on models.Location.
	for _, loc := range LocationSeeds {
		if loc.Name == "" {
			t.Error("location seed with empty name")
		}
		if loc.Address == "" {
			t.Errorf("location seed %q has no address", loc.Name)
		}
		if loc.City == "" {
			t.Errorf("location seed %q has no city", loc.Name)
		}
		if !loc.IsActive {
			t.Errorf("location seed %q is inactive; it would not appear as a branch option", loc.Name)
		}
	}
}

func TestRenamedLocationsPointAtSeededNames(t *testing.T) {
	// Every rename target must be a current seed name, or the rename would leave
	// databases pointing at a location that is never seeded.
	seeded := make(map[string]bool, len(LocationSeeds))
	for _, loc := range LocationSeeds {
		seeded[loc.Name] = true
	}

	for oldName, newName := range renamedLocations {
		if !seeded[newName] {
			t.Errorf("rename %q -> %q targets a name that is not seeded", oldName, newName)
		}
		if seeded[oldName] {
			t.Errorf("%q is both a rename source and a current seed name", oldName)
		}
	}
}
