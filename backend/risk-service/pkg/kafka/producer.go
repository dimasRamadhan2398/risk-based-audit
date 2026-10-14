package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

// Publisher sends Audit Universe events to Kafka. A nil *Publisher is valid and
// publishes nothing, so callers do not need to guard every call site.
type Publisher struct {
	producer    sarama.SyncProducer
	serviceName string
	now         func() time.Time
}

// NewPublisher connects a synchronous producer. Acks are WaitForAll so an
// acknowledged event survives a broker failover.
func NewPublisher(brokers []string, serviceName string) (*Publisher, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Retry.Max = 3
	cfg.Producer.Return.Successes = true
	cfg.Producer.Compression = sarama.CompressionSnappy

	producer, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		return nil, fmt.Errorf("kafka: create sync producer: %w", err)
	}
	return &Publisher{producer: producer, serviceName: serviceName, now: time.Now}, nil
}

// Close releases the broker connection.
func (p *Publisher) Close() error {
	if p == nil || p.producer == nil {
		return nil
	}
	return p.producer.Close()
}

// PublishEntity emits entity.created / entity.updated / entity.deleted.
func (p *Publisher) PublishEntity(ctx context.Context, eventType string, data EntityData) error {
	return p.publish(ctx, eventType, data.ID, data)
}

// PublishYearScore emits audit_universe.year.scored.
func (p *Publisher) PublishYearScore(ctx context.Context, data YearScoreData) error {
	return p.publish(ctx, EventYearScored, data.CorporateAuditUniverseID, data)
}

func (p *Publisher) publish(ctx context.Context, eventType, key string, data any) error {
	if p == nil || p.producer == nil {
		return nil
	}
	_ = ctx // sarama's SyncProducer has no context hook; kept for the call shape
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
	_, _, err = p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: TopicAuditUniverse,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(body),
	})
	if err != nil {
		return fmt.Errorf("kafka: send %s: %w", eventType, err)
	}
	return nil
}
