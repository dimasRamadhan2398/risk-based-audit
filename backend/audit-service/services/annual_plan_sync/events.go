// Package annual_plan_sync applies Audit Universe and Department events to the
// activities of annual audit plans.
//
// Producers: risk-service (topic audit-universe.events) and master-service
// (topic department.events). Envelope and data shapes are documented in
// backend/risk-service/pkg/kafka/events.go and
// backend/master-service/pkg/kafka/domain_publisher.go.
//
// Consistency model:
//   - Each event is applied as a convergent update: a field is written only when
//     it differs from the event, so redelivery and resync are no-ops.
//   - Writes use an optimistic check on audit_annuals.updated_at, so a concurrent
//     plan save from a user is never silently overwritten by a stale read.
//   - Events for one entity share a partition key (the entity id), so they are
//     applied in the order the producer emitted them.
package annual_plan_sync

import (
	"encoding/json"
	"time"
)

const (
	TopicAuditUniverse = "audit-universe.events"
	TopicDepartments   = "department.events"

	// SupportedEnvelopeVersion is the only envelope version this consumer
	// understands. Unknown versions are skipped, not guessed at.
	SupportedEnvelopeVersion = 1

	EventEntityCreated = "audit_universe.entity.created"
	EventEntityUpdated = "audit_universe.entity.updated"
	EventEntityDeleted = "audit_universe.entity.deleted"
	EventYearScored    = "audit_universe.year.scored"

	EventDepartmentCreated = "department.created"
	EventDepartmentUpdated = "department.updated"
	EventDepartmentDeleted = "department.deleted"
)

// Envelope is the common wrapper of every business event.
type Envelope struct {
	Version    int             `json:"version"`
	EventID    string          `json:"event_id"`
	EventType  string          `json:"event_type"`
	OccurredAt time.Time       `json:"occurred_at"`
	Producer   string          `json:"producer"`
	Data       json.RawMessage `json:"data"`
}

// EntityData is the data of audit_universe.entity.{created,updated,deleted}.
type EntityData struct {
	ID                      string  `json:"id"`
	Name                    string  `json:"name"`
	ParentID                *string `json:"parent_id"`
	StandardAuditUniverseID *string `json:"standard_audit_universe_id,omitempty"`
}

// YearScoreData is the data of audit_universe.year.scored.
type YearScoreData struct {
	CorporateAuditUniverseID string  `json:"corporate_audit_universe_id"`
	Year                     int     `json:"year"`
	RiskIndex                float64 `json:"risk_index"`
	RiskLevel                string  `json:"risk_level"`
	AuditPriority            bool    `json:"audit_priority"`
}

// DepartmentData is the data of department.{created,updated,deleted}.
type DepartmentData struct {
	ID             string  `json:"id"`
	DepartmentCode string  `json:"department_code"`
	DepartmentName string  `json:"department_name"`
	IsActive       bool    `json:"is_active"`
	CompanyID      string  `json:"company_id"`
	BusinessUnitID *string `json:"business_unit_id"`
}
