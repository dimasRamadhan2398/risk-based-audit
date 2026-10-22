package kafka

import "context"

// EventYearRemoved is recorded when an entity is taken out of a year's audit
// universe (its audit_universe_years row and its scores are deleted).
//
//	audit_universe.year.removed
//	    {"corporate_audit_universe_id","year"}
//
// Adding an event type does not bump EnvelopeVersion: consumers must skip
// types they do not know (audit-service's annual_plan_sync handler logs and
// acknowledges them).
const EventYearRemoved = "audit_universe.year.removed"

// YearRemovedData is the data of the year.removed event.
type YearRemovedData struct {
	CorporateAuditUniverseID string `json:"corporate_audit_universe_id"`
	Year                     int    `json:"year"`
}

// PublishYearRemoved records audit_universe.year.removed. The key is the
// entity id, like every other Audit Universe event.
func (p *Publisher) PublishYearRemoved(ctx context.Context, data YearRemovedData) error {
	return p.publish(ctx, EventYearRemoved, data.CorporateAuditUniverseID, data)
}
