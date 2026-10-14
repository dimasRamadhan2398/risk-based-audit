package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"risk-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Publisher records Audit Universe events in the transactional outbox
// (models.KafkaOutboxEvent). It does not talk to the broker: the Relay publishes
// the rows, so a request never waits on Kafka and an event survives a broker
// outage or a restart. A nil *Publisher is valid and records nothing, so
// callers do not need to guard every call site.
type Publisher struct {
	db          *gorm.DB
	serviceName string
	now         func() time.Time
}

// NewPublisher returns a publisher that writes to the outbox of db.
func NewPublisher(db *gorm.DB, serviceName string) *Publisher {
	return &Publisher{db: db, serviceName: serviceName, now: time.Now}
}

// PublishEntity records entity.created / entity.updated / entity.deleted.
func (p *Publisher) PublishEntity(ctx context.Context, eventType string, data EntityData) error {
	return p.publish(ctx, eventType, data.ID, data)
}

// PublishYearScore records audit_universe.year.scored.
func (p *Publisher) PublishYearScore(ctx context.Context, data YearScoreData) error {
	return p.publish(ctx, EventYearScored, data.CorporateAuditUniverseID, data)
}

func (p *Publisher) publish(ctx context.Context, eventType, key string, data any) error {
	if p == nil || p.db == nil {
		return nil
	}
	env := Envelope{
		Version:    EnvelopeVersion,
		EventID:    uuid.NewString(),
		EventType:  eventType,
		OccurredAt: p.now().UTC(),
		Producer:   p.serviceName,
		Data:       data,
	}
	body, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("kafka: marshal %s: %w", eventType, err)
	}
	return Enqueue(p.db.WithContext(ctx), TopicAuditUniverse, key, eventType, body)
}

// Close is kept for callers that closed the broker connection. The outbox holds
// no connection, so it does nothing.
func (p *Publisher) Close() error { return nil }

// Enqueue appends one message to the outbox. Use it with the handle that
// performed the write, after that write has committed.
func Enqueue(db *gorm.DB, topic, key, eventType string, body []byte) error {
	row := models.KafkaOutboxEvent{
		Topic:      topic,
		MessageKey: key,
		EventType:  eventType,
		Payload:    string(body),
	}
	if err := db.Create(&row).Error; err != nil {
		return fmt.Errorf("kafka: outbox insert %s: %w", eventType, err)
	}
	return nil
}
