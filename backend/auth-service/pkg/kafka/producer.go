package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// LogMessage represents a log entry sent to Kafka
type LogMessage struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       string                 `json:"level"`
	Service     string                 `json:"service"`
	Message     string                 `json:"message"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	SpanID      string                 `json:"span_id,omitempty"`
	Environment string                 `json:"environment,omitempty"`
}

// Producer wraps Sarama producer with logging-specific functionality
type Producer struct {
	producer    sarama.SyncProducer
	asyncProducer sarama.AsyncProducer
	serviceName string
	topic       string
	enabled     bool
	logger      *zap.Logger
	mu          sync.RWMutex
	useAsync    bool
}

// Config holds Kafka producer configuration
type Config struct {
	Brokers     []string
	Topic       string
	ServiceName string
	Enabled     bool
	UseAsync    bool
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg Config, logger *zap.Logger) (*Producer, error) {
	if !cfg.Enabled {
		return &Producer{
			serviceName: cfg.ServiceName,
			topic:       cfg.Topic,
			enabled:     false,
			logger:      logger,
		}, nil
	}

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3
	config.Producer.Return.Successes = true
	config.Producer.Compression = sarama.CompressionSnappy

	// Use Flush settings instead of deprecated Batch
	config.Producer.Flush.Messages = 10
	config.Producer.Flush.Bytes = 1024 * 1024 // 1MB
	config.Producer.Flush.Frequency = 500 * time.Millisecond

	var producer sarama.SyncProducer
	var asyncProducer sarama.AsyncProducer
	var err error

	if cfg.UseAsync {
		asyncProducer, err = sarama.NewAsyncProducer(cfg.Brokers, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create Kafka async producer: %w", err)
		}
	} else {
		producer, err = sarama.NewSyncProducer(cfg.Brokers, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create Kafka sync producer: %w", err)
		}
	}

	logger.Info("Kafka producer initialized",
		zap.String("service", cfg.ServiceName),
		zap.Strings("brokers", cfg.Brokers),
		zap.String("topic", cfg.Topic),
		zap.Bool("async", cfg.UseAsync),
	)

	return &Producer{
		producer:      producer,
		asyncProducer: asyncProducer,
		serviceName:   cfg.ServiceName,
		topic:         cfg.Topic,
		enabled:       true,
		logger:        logger,
		useAsync:      cfg.UseAsync,
	}, nil
}

// SendLog sends a log message to Kafka (async)
func (p *Producer) SendLog(ctx context.Context, level, message string, fields map[string]interface{}) {
	if !p.enabled {
		return
	}

	// Get trace info from context if available
	traceID := ""
	spanID := ""
	if span := extractSpanFromContext(ctx); span != nil {
		traceID = span.TraceID
		spanID = span.SpanID
	}

	logMsg := LogMessage{
		Timestamp:   time.Now().UTC(),
		Level:       level,
		Service:     p.serviceName,
		Message:     message,
		Fields:      fields,
		TraceID:     traceID,
		SpanID:      spanID,
		Environment: "development", // TODO: make this configurable
	}

	data, err := json.Marshal(logMsg)
	if err != nil {
		p.logger.Error("Failed to marshal log message",
			zap.Error(err),
			zap.String("message", message),
		)
		return
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(traceID), // Use traceID as key for ordering
		Value: sarama.ByteEncoder(data),
	}

	// Add headers for metadata
	msg.Headers = []sarama.RecordHeader{
		{Key: []byte("service"), Value: []byte(p.serviceName)},
		{Key: []byte("level"), Value: []byte(level)},
		{Key: []byte("timestamp"), Value: []byte(logMsg.Timestamp.Format(time.RFC3339))},
	}

	if p.useAsync && p.asyncProducer != nil {
		select {
		case p.asyncProducer.Input() <- msg:
			// Message queued successfully
		case <-time.After(5 * time.Second):
			p.logger.Warn("Failed to send log: timeout")
		case <-ctx.Done():
			p.logger.Warn("Failed to send log: context cancelled")
		}
		return
	}

	// Fallback to sync
	p.sendLogSync(msg)
}

// sendLogSync sends a message synchronously
func (p *Producer) sendLogSync(msg *sarama.ProducerMessage) {
	if p.producer == nil {
		return
	}

	_, _, err := p.producer.SendMessage(msg)
	if err != nil {
		p.logger.Error("Failed to send log message",
			zap.Error(err),
		)
	}
}

// SendLogSync sends a log message and waits for acknowledgment
func (p *Producer) SendLogSync(ctx context.Context, level, message string, fields map[string]interface{}) error {
	if !p.enabled {
		return nil
	}

	logMsg := LogMessage{
		Timestamp:   time.Now().UTC(),
		Level:       level,
		Service:     p.serviceName,
		Message:     message,
		Fields:      fields,
		Environment: "development",
	}

	data, err := json.Marshal(logMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal log message: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.ByteEncoder(data),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	p.logger.Debug("Log message sent",
		zap.String("topic", p.topic),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset),
	)

	return nil
}

// Info sends an info level log
func (p *Producer) Info(ctx context.Context, message string, fields map[string]interface{}) {
	p.SendLog(ctx, "info", message, fields)
}

// Error sends an error level log
func (p *Producer) Error(ctx context.Context, message string, fields map[string]interface{}) {
	p.SendLog(ctx, "error", message, fields)
}

// Warn sends a warning level log
func (p *Producer) Warn(ctx context.Context, message string, fields map[string]interface{}) {
	p.SendLog(ctx, "warn", message, fields)
}

// Debug sends a debug level log
func (p *Producer) Debug(ctx context.Context, message string, fields map[string]interface{}) {
	p.SendLog(ctx, "debug", message, fields)
}

// Close closes the Kafka producer
func (p *Producer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.enabled {
		return nil
	}

	var err error
	if p.useAsync && p.asyncProducer != nil {
		p.asyncProducer.AsyncClose()
	} else if p.producer != nil {
		err = p.producer.Close()
	}

	if err != nil {
		return fmt.Errorf("failed to close Kafka producer: %w", err)
	}

	p.logger.Info("Kafka producer closed")
	return nil
}

// IsEnabled returns whether the producer is enabled
func (p *Producer) IsEnabled() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.enabled
}

// spanInfo holds trace information
type spanInfo struct {
	TraceID string
	SpanID  string
}

// extractSpanFromContext extracts trace information from context
// This is a placeholder - in production, you'd use OpenTelemetry or similar
func extractSpanFromContext(ctx context.Context) *spanInfo {
	// TODO: Implement proper trace context extraction
	// For now, check if there's trace info in context
	if traceID := ctx.Value("trace_id"); traceID != nil {
		return &spanInfo{
			TraceID: traceID.(string),
		}
	}
	return nil
}
