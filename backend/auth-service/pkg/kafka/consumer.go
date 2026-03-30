package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

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
	ready         chan bool
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
}

// ConsumerConfig holds Kafka consumer configuration
type ConsumerConfig struct {
	Brokers  []string
	GroupID  string
	Topics   []string
	Enabled  bool
	Version  string // Kafka version, e.g., "3.6.0"
	Assignor string // partition assignor: "range", "roundrobin", "sticky"
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
	config.Consumer.Group.Session.Timeout = 10 * 1000 * 1000 * 1000 // 10s (nanoseconds for older versions)
	config.Consumer.Group.Heartbeat.Interval = 3 * 1000 * 1000 * 1000 // 3s
	config.Consumer.MaxProcessingTime = 5 * 1000 * 1000 * 1000 // 5s

	// Fetch config
	config.Consumer.Fetch.Min = 1
	config.Consumer.Fetch.Default = 1024 * 1024 // 1MB
	config.Consumer.MaxWaitTime = 500 * 1000 * 1000 // 500ms

	// Offset reset policy
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
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
		ready:         make(chan bool),
		ctx:           ctx,
		cancel:        cancel,
	}, nil
}

// RegisterHandler registers a message handler for a specific topic
func (c *Consumer) RegisterHandler(topic string, handler MessageHandler) {
	c.handlers[topic] = handler
	c.logger.Info("Handler registered for topic", zap.String("topic", topic))
}

// Start begins consuming messages
func (c *Consumer) Start() error {
	if !c.enabled {
		c.logger.Info("Kafka consumer is disabled, not starting")
		return nil
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			if err := c.consumerGroup.Consume(c.ctx, c.topics, c); err != nil {
				if err == sarama.ErrClosedConsumerGroup {
					return
				}
				c.logger.Error("Error from consumer", zap.Error(err))
			}

			// Check if context was cancelled
			if c.ctx.Err() != nil {
				return
			}

			c.ready = make(chan bool)
		}
	}()

	// Wait for consumer to be ready
	<-c.ready
	c.logger.Info("Kafka consumer is ready")

	return nil
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
	close(c.ready)
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

// processMessage handles a single message
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

	msgCtx := context.Background()

	if err := handler.Handle(msgCtx, message.Value); err != nil {
		c.logger.Error("Failed to handle message",
			zap.Error(err),
			zap.String("topic", topic),
			zap.Int64("offset", message.Offset),
		)
		// Don't mark message as processed on error to retry
		// In production, you might want dead letter queue or retry logic
		return
	}

	session.MarkMessage(message, "")
	session.Context()
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
