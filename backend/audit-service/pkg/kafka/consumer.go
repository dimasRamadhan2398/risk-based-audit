package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// MessageHandler defines the interface for handling consumed messages
type MessageHandler interface {
	Handle(ctx context.Context, message []byte) error
}

// Consumer wraps Sarama consumer group for message consumption
type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	handlers      map[string]MessageHandler
	topics        []string
	groupID       string
	enabled       bool
	logger        *zap.Logger
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
}

// ConsumerConfig holds Kafka consumer configuration
type ConsumerConfig struct {
	// InitialOffset is used only when the group has no committed offset:
	// "oldest" replays the retained topic, anything else starts at the newest
	// message. Business-event consumers should use "oldest" so events published
	// before the first start are not skipped.
	InitialOffset string
	Brokers       []string
	GroupID       string
	Topics        []string
	Enabled       bool
	Version       string // Kafka version, e.g., "3.6.0"
	Assignor      string // partition assignor: "range", "roundrobin", "sticky"
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg ConsumerConfig, logger *zap.Logger) (*Consumer, error) {
	if !cfg.Enabled {
		return &Consumer{
			enabled: false,
			logger:  logger,
		}, nil
	}

	kafkaVersion, err := sarama.ParseKafkaVersion(cfg.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Kafka version: %w", err)
	}

	config := sarama.NewConfig()
	config.Version = kafkaVersion
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategyRoundRobin(),
	}

	// Session config
	config.Consumer.Group.Session.Timeout = 10 * 1000 * 1000 * 1000   // 10s (nanoseconds for older versions)
	config.Consumer.Group.Heartbeat.Interval = 3 * 1000 * 1000 * 1000 // 3s
	config.Consumer.MaxProcessingTime = 5 * 1000 * 1000 * 1000        // 5s

	// Fetch config
	config.Consumer.Fetch.Min = 1
	config.Consumer.Fetch.Default = 1024 * 1024     // 1MB
	config.Consumer.MaxWaitTime = 500 * 1000 * 1000 // 500ms

	// Offset reset policy
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	if cfg.InitialOffset == "oldest" {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 1 * 1000 * 1000 * 1000 // 1s

	// Set assignor based on config
	switch cfg.Assignor {
	case "range":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
			sarama.NewBalanceStrategyRange(),
		}
	case "sticky":
		config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
			sarama.NewBalanceStrategySticky(),
		}
	default:
		// Already set to roundrobin
	}

	consumerGroup, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	logger.Info("Kafka consumer initialized",
		zap.Strings("brokers", cfg.Brokers),
		zap.String("group_id", cfg.GroupID),
		zap.Strings("topics", cfg.Topics),
	)

	return &Consumer{
		consumerGroup: consumerGroup,
		handlers:      make(map[string]MessageHandler),
		topics:        cfg.Topics,
		groupID:       cfg.GroupID,
		enabled:       true,
		logger:        logger,
		ctx:           ctx,
		cancel:        cancel,
	}, nil
}

// Topics returns the topics the consumer subscribes to (nil when disabled).
// Handlers must be registered under these exact names: a message whose topic
// has no handler is acknowledged and skipped.
func (c *Consumer) Topics() []string {
	return append([]string(nil), c.topics...)
}

// RegisterHandler registers a message handler for a specific topic
func (c *Consumer) RegisterHandler(topic string, handler MessageHandler) {
	c.handlers[topic] = handler
	c.logger.Info("Handler registered for topic", zap.String("topic", topic))
}

// Start begins consuming messages in the background and returns at once. A
// failed session (broker down, topic not created yet, rebalance error) is
// retried with exponential backoff until Stop is called, so the consumer does
// not need a restart to recover.
func (c *Consumer) Start() error {
	if !c.enabled {
		c.logger.Info("Kafka consumer is disabled, not starting")
		return nil
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		backoff := consumerRetryMin
		for {
			err := c.consumerGroup.Consume(c.ctx, c.topics, c)
			if c.ctx.Err() != nil || errors.Is(err, sarama.ErrClosedConsumerGroup) {
				return
			}
			wait := backoff
			if err != nil {
				c.logger.Error("Kafka consumer session failed, retrying",
					zap.Error(err), zap.Duration("retry_in", wait))
				backoff = nextBackoff(backoff, consumerRetryMax)
			} else {
				// A session that ended without error (for example a rebalance)
				// restarts quickly.
				backoff = consumerRetryMin
				wait = consumerRetryMin
			}
			select {
			case <-c.ctx.Done():
				return
			case <-time.After(wait):
			}
		}
	}()

	return nil
}

// Retry bounds for consumer sessions and for a message whose handler failed.
// Variables so tests can shorten them.
var (
	consumerRetryMin = 1 * time.Second
	consumerRetryMax = 30 * time.Second
)

func nextBackoff(current, limit time.Duration) time.Duration {
	next := current * 2
	if next > limit {
		return limit
	}
	return next
}

// Stop gracefully shuts down the consumer
func (c *Consumer) Stop() error {
	if !c.enabled {
		return nil
	}

	c.cancel()
	c.wg.Wait()

	if err := c.consumerGroup.Close(); err != nil {
		return fmt.Errorf("failed to close consumer group: %w", err)
	}

	c.logger.Info("Kafka consumer stopped")
	return nil
}

// IsEnabled returns whether the consumer is enabled
func (c *Consumer) IsEnabled() bool {
	return c.enabled
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	c.logger.Info("Kafka consumer session started", zap.String("group_id", c.groupID))
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim processes messages from a partition
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				c.logger.Info("Message channel closed")
				return nil
			}

			c.processMessage(session, message)

		case <-session.Context().Done():
			return nil
		}
	}
}

// processMessage handles a single message. A handler error is retried in place
// with backoff until it succeeds or the session ends. The offset is marked only
// after success, and the partition is not read past a failed message, so a
// later message is never committed over an earlier one that was not applied.
// If the session ends first, the message is not marked and is redelivered.
func (c *Consumer) processMessage(session sarama.ConsumerGroupSession, message *sarama.ConsumerMessage) {
	topic := message.Topic
	handler, exists := c.handlers[topic]

	if !exists {
		c.logger.Warn("No handler registered for topic",
			zap.String("topic", topic),
			zap.Int64("offset", message.Offset),
		)
		session.MarkMessage(message, "")
		return
	}

	wait := consumerRetryMin
	for {
		err := handler.Handle(c.ctx, message.Value)
		if err == nil {
			break
		}
		c.logger.Error("Failed to handle message, retrying",
			zap.Error(err),
			zap.String("topic", topic),
			zap.Int64("offset", message.Offset),
			zap.Duration("retry_in", wait),
		)
		select {
		case <-session.Context().Done():
			return
		case <-c.ctx.Done():
			return
		case <-time.After(wait):
		}
		wait = nextBackoff(wait, consumerRetryMax)
	}

	session.MarkMessage(message, "")
}

// GenericMessageHandler is a function type for simple message handling
type GenericMessageHandler func(ctx context.Context, message []byte) error

// Handle implements MessageHandler for GenericMessageHandler
func (h GenericMessageHandler) Handle(ctx context.Context, message []byte) error {
	return h(ctx, message)
}

// JSONMessageHandler decodes JSON messages and calls the handler function
type JSONMessageHandler[T any] struct {
	Handler func(ctx context.Context, msg T) error
}

func (h *JSONMessageHandler[T]) Handle(ctx context.Context, message []byte) error {
	var msg T
	if err := json.Unmarshal(message, &msg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}
	return h.Handler(ctx, msg)
}

// NewJSONMessageHandler creates a typed JSON message handler
func NewJSONMessageHandler[T any](handler func(ctx context.Context, msg T) error) *JSONMessageHandler[T] {
	return &JSONMessageHandler[T]{Handler: handler}
}
