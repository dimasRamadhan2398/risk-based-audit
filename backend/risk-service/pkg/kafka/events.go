// Package kafka publishes Audit Universe domain events.
//
// Contract (consumed by audit-service, see services/annual_plan_sync there):
//
//	topic      [<KAFKA_TOPIC_PREFIX>.]audit-universe.events
//	key        corporate_audit_universe.id (keeps one entity's events ordered)
//	envelope   {"version":1,"event_id","event_type","occurred_at","producer","data"}
//
// Event types and data:
//
//	audit_universe.entity.created | audit_universe.entity.updated
//	    {"id","name","parent_id"|null,"standard_audit_universe_id"|null}
//	audit_universe.entity.deleted
//	    {"id","name","parent_id"|null}
//	audit_universe.year.scored
//	    {"corporate_audit_universe_id","year","risk_index","risk_level","audit_priority"}
//
// This is a business-event stream, not the service-logs stream that
// pkg/kafka in the other services uses for log lines.
package kafka

import (
	"os"
	"strings"
	"time"
)

// TopicPrefixEnv names the env var that isolates one stack's business events
// from every other stack on the shared broker. Tenant stacks set it to the
// tenant slug (scripts/templates/docker-compose.tenant.yml.tpl); audit-service
// reads the same variable, so producer and consumer always agree. Unset or
// empty keeps the historical, unprefixed topic used by the shared prod stack.
const TopicPrefixEnv = "KAFKA_TOPIC_PREFIX"

// BaseTopicAuditUniverse is the unprefixed Audit Universe topic name.
const BaseTopicAuditUniverse = "audit-universe.events"

// TopicAuditUniverse carries Audit Universe entity and yearly score events. It
// is BaseTopicAuditUniverse, prefixed with KAFKA_TOPIC_PREFIX when that is set,
// and is resolved once at start-up.
var TopicAuditUniverse = AuditUniverseTopicFromEnv()

// AuditUniverseTopicFromEnv returns the Audit Universe topic for this process's
// KAFKA_TOPIC_PREFIX.
func AuditUniverseTopicFromEnv() string {
	return TopicName(os.Getenv(TopicPrefixEnv), BaseTopicAuditUniverse)
}

// TopicName returns base, or "<prefix>.<base>" when prefix is non-empty.
// Surrounding whitespace and dots on the prefix are ignored.
func TopicName(prefix, base string) string {
	prefix = strings.Trim(strings.TrimSpace(prefix), ".")
	if prefix == "" {
		return base
	}
	return prefix + "." + base
}

const (
	// EnvelopeVersion is bumped only for breaking changes to the envelope or to
	// an existing event's data. Adding optional fields does not bump it.
	EnvelopeVersion = 1

	EventEntityCreated = "audit_universe.entity.created"
	EventEntityUpdated = "audit_universe.entity.updated"
	EventEntityDeleted = "audit_universe.entity.deleted"
	EventYearScored    = "audit_universe.year.scored"
)

// Envelope wraps every event published to Kafka.
type Envelope struct {
	Version    int       `json:"version"`
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	OccurredAt time.Time `json:"occurred_at"`
	Producer   string    `json:"producer"`
	Data       any       `json:"data"`
}

// EntityData is the data of the entity.created, entity.updated and
// entity.deleted events.
type EntityData struct {
	ID                      string  `json:"id"`
	Name                    string  `json:"name"`
	ParentID                *string `json:"parent_id"`
	StandardAuditUniverseID *string `json:"standard_audit_universe_id,omitempty"`
}

// YearScoreData is the data of the year.scored event.
type YearScoreData struct {
	CorporateAuditUniverseID string  `json:"corporate_audit_universe_id"`
	Year                     int     `json:"year"`
	RiskIndex                float64 `json:"risk_index"`
	RiskLevel                string  `json:"risk_level"`
	AuditPriority            bool    `json:"audit_priority"`
}
