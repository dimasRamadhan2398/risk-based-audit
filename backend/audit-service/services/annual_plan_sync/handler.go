package annual_plan_sync

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"audit-service/pkg/kafka"
	"audit-service/pkg/logger"
)

// handlerAttempts and handlerBackoff bound the in-process retries for a
// transient failure. After that the error is returned and the consumer leaves
// the offset uncommitted, so the message is retried after a restart.
const (
	handlerAttempts = 3
	handlerBackoff  = 200 * time.Millisecond
)

// Handler decodes one Kafka message and routes it by event type. It implements
// kafka.MessageHandler and can be registered on both topics.
type Handler struct {
	syncer *Syncer
}

func NewHandler(syncer *Syncer) *Handler {
	return &Handler{syncer: syncer}
}

var _ kafka.MessageHandler = (*Handler)(nil)

// Handle applies one event. Malformed messages, unknown versions and unknown
// event types are logged and acknowledged (return nil): redelivering them cannot
// succeed, and holding the partition would block every later event.
func (h *Handler) Handle(ctx context.Context, message []byte) error {
	var env Envelope
	if err := json.Unmarshal(message, &env); err != nil {
		logger.Warn("annual plan sync: skipping malformed event", logger.LogField("error", err))
		return nil
	}
	if env.Version != SupportedEnvelopeVersion {
		logger.Warn("annual plan sync: skipping event with unsupported version",
			logger.LogField("event_type", env.EventType), logger.LogField("version", env.Version))
		return nil
	}

	apply, ok, err := h.route(env)
	if err != nil {
		logger.Warn("annual plan sync: skipping event with undecodable data",
			logger.LogField("event_type", env.EventType), logger.LogField("error", err))
		return nil
	}
	if !ok {
		return nil
	}
	return withRetry(ctx, env, apply)
}

// route returns the apply function for the event type. ok is false for event
// types that need no plan change (deletions keep the name snapshot).
func (h *Handler) route(env Envelope) (func(context.Context) error, bool, error) {
	switch env.EventType {
	case EventEntityCreated, EventEntityUpdated:
		var d EntityData
		if err := json.Unmarshal(env.Data, &d); err != nil {
			return nil, false, err
		}
		return func(ctx context.Context) error { return h.syncer.ApplyEntity(ctx, d) }, true, nil
	case EventYearScored:
		var d YearScoreData
		if err := json.Unmarshal(env.Data, &d); err != nil {
			return nil, false, err
		}
		return func(ctx context.Context) error { return h.syncer.ApplyYearScore(ctx, d) }, true, nil
	case EventDepartmentCreated, EventDepartmentUpdated:
		var d DepartmentData
		if err := json.Unmarshal(env.Data, &d); err != nil {
			return nil, false, err
		}
		return func(ctx context.Context) error { return h.syncer.ApplyDepartment(ctx, d) }, true, nil
	case EventEntityDeleted, EventDepartmentDeleted:
		logger.Info("annual plan sync: keeping plan snapshot for deleted source",
			logger.LogField("event_type", env.EventType), logger.LogField("event_id", env.EventID))
		return nil, false, nil
	default:
		logger.Warn("annual plan sync: skipping unknown event type", logger.LogField("event_type", env.EventType))
		return nil, false, nil
	}
}

func withRetry(ctx context.Context, env Envelope, apply func(context.Context) error) error {
	var err error
	for attempt := 1; attempt <= handlerAttempts; attempt++ {
		if err = apply(ctx); err == nil {
			return nil
		}
		logger.Warn("annual plan sync: apply failed, will retry",
			logger.LogField("event_type", env.EventType),
			logger.LogField("event_id", env.EventID),
			logger.LogField("attempt", attempt),
			logger.LogField("error", err))
		if attempt < handlerAttempts {
			time.Sleep(handlerBackoff * time.Duration(attempt))
		}
	}
	return fmt.Errorf("annual plan sync: %s %s: %w", env.EventType, env.EventID, err)
}

// Register attaches the handler to the consumer for every topic it subscribes
// to. The subscribed names come from config (kafka_consumer.topics, renamed by
// KAFKA_TOPIC_PREFIX), so they may be "<prefix>.audit-universe.events" rather
// than TopicAuditUniverse; registering the subscribed names keeps the two from
// drifting apart, which would make the consumer acknowledge and drop every
// event. Handler routes by event type, so one handler serves every topic.
//
// The consumer must not have started yet (kafka.Consumer.RegisterHandler is not
// synchronised with the consume loop).
func Register(consumer *kafka.Consumer, h *Handler) {
	for _, topic := range consumer.Topics() {
		consumer.RegisterHandler(topic, h)
	}
}
