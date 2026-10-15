package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"risk-service/models"
	"risk-service/pkg/masterclient"

	"github.com/google/uuid"
)

// Branches come from the Location master (master-service). A risk is linked to
// a location by risk_profile.location_id; location_name is a denormalised copy
// for the backfill and for humans reading the table. The API always reports the
// master's current name, and null when a risk is not linked to a registered
// location — there is no default branch.

// legacySentinelBranch is how risks were tagged with a branch before
// risk_profile.location_id existed: the branch name was encoded as a sentinel
// UUID in DepartmentID. Production rows still carry only that, so reads map
// sentinel -> legacy name -> master location *by name*. The names here are
// lookup keys only; a name that is not registered in the Location master
// resolves to nothing. Nothing writes these sentinels any more (the demo seeder
// in cmd/seed.go still does; reads handle it the same way).
var legacySentinelBranch = map[uuid.UUID]string{
	uuid.MustParse("00000000-0000-0000-0000-000000000001"): "Head Office",
	uuid.MustParse("00000000-0000-0000-0000-000000000002"): "Jakarta Branch",
	uuid.MustParse("00000000-0000-0000-0000-000000000003"): "Surabaya Branch",
	uuid.MustParse("00000000-0000-0000-0000-000000000004"): "Bandung Branch",
	uuid.MustParse("00000000-0000-0000-0000-000000000005"): "Bali Branch",
}

// ErrLocationsUnavailable is returned by writes that need the Location master
// to validate a branch while master-service is unreachable.
var ErrLocationsUnavailable = errors.New("location master data is unavailable")

// LocationError is a client error: the request named a location that is not a
// registered location (or an unparseable location_id).
type LocationError struct {
	Code    string
	Message string
}

func (e *LocationError) Error() string { return e.Message }

// locationIndex is a lookup over one snapshot of the Location master.
type locationIndex struct {
	byID   map[uuid.UUID]masterclient.Location
	byName map[string]masterclient.Location
}

func newLocationIndex(locs []masterclient.Location) *locationIndex {
	idx := &locationIndex{
		byID:   make(map[uuid.UUID]masterclient.Location, len(locs)),
		byName: make(map[string]masterclient.Location, len(locs)),
	}
	for _, l := range locs {
		idx.byID[l.ID] = l
		key := normalizeLocationName(l.Name)
		// Duplicate names: prefer an active location over an inactive one,
		// otherwise keep the first seen.
		if existing, ok := idx.byName[key]; !ok || (!existing.IsActive && l.IsActive) {
			idx.byName[key] = l
		}
	}
	return idx
}

func (idx *locationIndex) findByName(name string) (masterclient.Location, bool) {
	l, ok := idx.byName[normalizeLocationName(name)]
	return l, ok && name != ""
}

// normalizeLocationName makes name matching case- and whitespace-insensitive.
func normalizeLocationName(s string) string {
	return strings.ToLower(strings.Join(strings.Fields(s), " "))
}

// resolveBranch reports the master location a stored risk profile belongs to.
//
// idx == nil means the Location master could not be read. Then nothing is
// guessed: a stored location_id is still reported (it is a real stored link),
// but the branch name is null.
//
// Order:
//  1. location_id set: that location, or unlinked if it is no longer registered.
//  2. otherwise the stored location_name, then the legacy sentinel DepartmentID,
//     matched to a master location by name.
//  3. otherwise unlinked (null, null).
func resolveBranch(p models.RiskProfile, idx *locationIndex) (locationID, branch *string) {
	if p.LocationID != nil && *p.LocationID != uuid.Nil {
		if idx == nil {
			return strPtr(p.LocationID.String()), nil
		}
		if l, ok := idx.byID[*p.LocationID]; ok {
			return strPtr(l.ID.String()), strPtr(l.Name)
		}
		return nil, nil
	}

	if idx == nil {
		return nil, nil
	}
	if l, ok := matchLegacyLocation(p, idx); ok {
		return strPtr(l.ID.String()), strPtr(l.Name)
	}
	return nil, nil
}

// matchLegacyLocation finds the master location for a profile that has no
// location_id, using the free-text location_name or the legacy sentinel.
func matchLegacyLocation(p models.RiskProfile, idx *locationIndex) (masterclient.Location, bool) {
	if l, ok := idx.findByName(p.LocationName); ok {
		return l, true
	}
	if name, ok := legacySentinelBranch[p.DepartmentID]; ok {
		return idx.findByName(name)
	}
	return masterclient.Location{}, false
}

// locationForRequest validates the location a create/update request asks for.
//
// It returns (nil, nil) when the request does not touch the location: neither
// location_id nor branch was sent, or location_id equals the one already stored
// on current. location_id wins over branch; a branch name alone is matched to
// the master by name. Anything that is not a registered location is a
// *LocationError; an unreachable master is ErrLocationsUnavailable.
func (s *riskService) locationForRequest(ctx context.Context, req *RiskRequest, current *models.RiskProfile) (*masterclient.Location, error) {
	idStr := strings.TrimSpace(req.LocationID)
	branch := strings.TrimSpace(req.Branch)
	if idStr == "" && branch == "" {
		return nil, nil
	}

	var locID uuid.UUID
	if idStr != "" {
		parsed, err := uuid.Parse(idStr)
		if err != nil || parsed == uuid.Nil {
			return nil, &LocationError{Code: "INVALID_LOCATION_ID", Message: "location_id must be the UUID of a registered location"}
		}
		if current != nil && current.LocationID != nil && *current.LocationID == parsed {
			return nil, nil
		}
		locID = parsed
	}

	locs, err := s.locations.ListLocations(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrLocationsUnavailable, err)
	}
	idx := newLocationIndex(locs)

	if idStr != "" {
		if l, ok := idx.byID[locID]; ok {
			return &l, nil
		}
		return nil, &LocationError{Code: "UNKNOWN_LOCATION", Message: fmt.Sprintf("location_id %s is not a registered location", locID)}
	}
	if l, ok := idx.findByName(branch); ok {
		return &l, nil
	}
	return nil, &LocationError{Code: "UNKNOWN_LOCATION", Message: fmt.Sprintf("branch %q is not a registered location", branch)}
}

// applyLocation links a profile to a master location.
func applyLocation(p *models.RiskProfile, loc *masterclient.Location) {
	if loc == nil {
		return
	}
	id := loc.ID
	p.LocationID = &id
	p.LocationName = loc.Name
}

// locationIndexOrNil reads the Location master for display. Failure is not an
// error for reads: branches are reported as null instead.
func (s *riskService) locationIndexOrNil(ctx context.Context) *locationIndex {
	locs, err := s.locations.ListLocations(ctx)
	if err != nil {
		log.Printf("risk-service: location master unavailable, branches will be null: %v", err)
		return nil
	}
	return newLocationIndex(locs)
}

// LocationBackfill is one planned risk_profile update for the backfill command.
type LocationBackfill struct {
	ProfileID    uuid.UUID
	Source       string // "location_name" or "legacy_department_id"
	FromValue    string // the stored value the match was made from
	LocationID   uuid.UUID
	LocationName string
}

// LocationBackfillMiss is a profile that could not be linked.
type LocationBackfillMiss struct {
	ProfileID uuid.UUID
	Reason    string
}

// PlanLocationBackfill decides which profiles without a location_id can be
// linked to a master location, using the same rules as reads. Profiles that
// already have a location_id are skipped, so re-running is a no-op.
func PlanLocationBackfill(profiles []models.RiskProfile, locs []masterclient.Location) ([]LocationBackfill, []LocationBackfillMiss) {
	idx := newLocationIndex(locs)
	var plan []LocationBackfill
	var misses []LocationBackfillMiss

	for _, p := range profiles {
		if p.LocationID != nil && *p.LocationID != uuid.Nil {
			continue
		}
		legacyName, hasSentinel := legacySentinelBranch[p.DepartmentID]
		if strings.TrimSpace(p.LocationName) == "" && !hasSentinel {
			continue // never had a branch; stays unlinked
		}

		l, ok := matchLegacyLocation(p, idx)
		if !ok {
			from := p.LocationName
			if from == "" {
				from = legacyName
			}
			misses = append(misses, LocationBackfillMiss{
				ProfileID: p.ID,
				Reason:    fmt.Sprintf("branch %q is not a registered location", from),
			})
			continue
		}

		source, from := "legacy_department_id", legacyName
		if _, byName := idx.findByName(p.LocationName); byName {
			source, from = "location_name", p.LocationName
		}
		plan = append(plan, LocationBackfill{
			ProfileID:    p.ID,
			Source:       source,
			FromValue:    from,
			LocationID:   l.ID,
			LocationName: l.Name,
		})
	}
	return plan, misses
}

func strPtr(s string) *string { return &s }
