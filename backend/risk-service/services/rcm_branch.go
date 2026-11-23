package services

import (
	"risk-service/models"
	"risk-service/pkg/masterclient"

	"github.com/google/uuid"
)

// RCM branch rule, ported from resolveRcmItemLocationId in
// frontend/utils/rcmBranch.ts so the server narrows RCM rows by branch exactly
// as the RCM page does:
//
//  0. the row's own location_id (picked from the Location master) -> that id;
//  1. no risk_id, or the risk is not found (deleted)  -> unassigned (nil);
//  2. the risk's assessment for the row's year has a location_id -> that id;
//  3. otherwise the risk's own location as the risk API reports it
//     (resolveBranch: location_id, else location_name / legacy sentinel
//     matched to the Location master by name);
//  4. otherwise unassigned.

// RCMBranchResolver applies the rule against one Location master snapshot.
type RCMBranchResolver struct {
	idx *locationIndex // nil: Location master unavailable
}

// NewRCMBranchResolver builds a resolver. Pass available=false when the
// Location master could not be read: stored location ids are still used, but
// legacy names cannot be matched (same degradation as the risk API).
func NewRCMBranchResolver(locs []masterclient.Location, available bool) *RCMBranchResolver {
	if !available {
		return &RCMBranchResolver{}
	}
	return &RCMBranchResolver{idx: newLocationIndex(locs)}
}

// RowLocationID is the branch of one RCM row (rule 0-4). reg is the row's
// linked live risk (nil when none), with Profile and Assessments loaded.
func (r *RCMBranchResolver) RowLocationID(row models.RiskControlMatrix, reg *models.RiskRegister) *uuid.UUID {
	if row.LocationID != nil && *row.LocationID != uuid.Nil {
		id := *row.LocationID
		return &id
	}
	return r.LocationID(row.Year, reg)
}

// LocationID is the branch of an RCM row of the given year whose linked risk
// is reg (nil when the row has no linked live risk). reg.Profile and
// reg.Assessments must be loaded.
func (r *RCMBranchResolver) LocationID(year int, reg *models.RiskRegister) *uuid.UUID {
	if reg == nil {
		return nil
	}
	if year != 0 {
		for _, a := range reg.Assessments {
			if a.Year == year && a.LocationID != nil && *a.LocationID != uuid.Nil {
				id := *a.LocationID
				return &id
			}
		}
	}
	locID, _ := resolveBranch(reg.Profile, r.idx)
	if locID == nil {
		return nil
	}
	id, err := uuid.Parse(*locID)
	if err != nil || id == uuid.Nil {
		return nil
	}
	return &id
}
