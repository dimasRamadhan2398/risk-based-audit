package kafka

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"master-service/models"

	"github.com/IBM/sarama"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Relay publishes outbox rows to Kafka in id order and deletes each row after
// the broker acknowledges it. It runs for the life of the process and retries
// with backoff while the broker or the topic is unavailable.
//
// Ordering: rows are published one at a time, and a failed row stops the batch,
// so later events for the same key (the message key is the entity id) never
// overtake an earlier one. Delivery is at least once: a crash between the
// broker ack and the delete publishes the row again, and consumers are
// idempotent.
//
// Several replicas may run a relay. Postgres rows are claimed with
// FOR UPDATE SKIP LOCKED, so two replicas never publish the same row at once,
// but events for one key can be split between replicas and interleave.
type Relay struct {
	db      *gorm.DB
	brokers []string
	topics  []string
}

const (
	relayBatchSize = 100
	relayIdleWait  = time.Second
	relayRetryMin  = time.Second
	relayRetryMax  = 30 * time.Second
)

// NewRelay returns a relay that publishes the outbox to brokers. topics are
// created on the broker, with its default partitions and replication factor, if
// they do not exist.
func NewRelay(db *gorm.DB, brokers []string, topics ...string) *Relay {
	return &Relay{db: db, brokers: brokers, topics: topics}
}

// Run blocks until ctx is cancelled.
func (r *Relay) Run(ctx context.Context) {
	backoff := relayRetryMin
	for ctx.Err() == nil {
		producer, err := r.connect()
		if err != nil {
			log.Printf("kafka relay: broker not ready, retrying in %s: %v", backoff, err)
			if !sleepCtx(ctx, backoff) {
				return
			}
			backoff = nextRelayBackoff(backoff)
			continue
		}
		backoff = relayRetryMin

		err = r.drain(ctx, producer)
		_ = producer.Close()
		if ctx.Err() != nil {
			return
		}
		log.Printf("kafka relay: publish failed, retrying in %s: %v", backoff, err)
		if !sleepCtx(ctx, backoff) {
			return
		}
		backoff = nextRelayBackoff(backoff)
	}
}

// drain publishes batches until an error occurs or ctx is cancelled.
func (r *Relay) drain(ctx context.Context, producer messageSender) error {
	for ctx.Err() == nil {
		n, err := r.publishBatch(ctx, producer)
		if err != nil {
			return err
		}
		if n == 0 && !sleepCtx(ctx, relayIdleWait) {
			return nil
		}
	}
	return nil
}

// publishBatch sends up to relayBatchSize rows in one transaction. The
// transaction commits even when a send fails, so the rows already published are
// removed and the failed row records its error.
func (r *Relay) publishBatch(ctx context.Context, producer messageSender) (int, error) {
	published := 0
	var sendErr error
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := tx.Order("id ASC").Limit(relayBatchSize)
		if tx.Dialector.Name() == "postgres" {
			q = q.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"})
		}
		var rows []models.KafkaOutboxEvent
		if err := q.Find(&rows).Error; err != nil {
			return err
		}
		for _, row := range rows {
			_, _, err := producer.SendMessage(&sarama.ProducerMessage{
				Topic: row.Topic,
				Key:   sarama.StringEncoder(row.MessageKey),
				Value: sarama.ByteEncoder([]byte(row.Payload)),
			})
			if err != nil {
				sendErr = fmt.Errorf("outbox %d (%s): %w", row.ID, row.EventType, err)
				if dbErr := tx.Model(&models.KafkaOutboxEvent{}).Where("id = ?", row.ID).
					Updates(map[string]interface{}{"attempts": gorm.Expr("attempts + 1"), "last_error": err.Error()}).Error; dbErr != nil {
					return dbErr
				}
				return nil
			}
			if err := tx.Delete(&models.KafkaOutboxEvent{}, row.ID).Error; err != nil {
				return err
			}
			published++
		}
		return nil
	})
	if err != nil {
		return published, err
	}
	return published, sendErr
}

// messageSender is the part of sarama.SyncProducer the relay uses. Tests
// supply a fake.
type messageSender interface {
	SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error)
}

// connect creates the producer and makes sure the topics exist.
func (r *Relay) connect() (sarama.SyncProducer, error) {
	cfg := producerConfig()
	if err := ensureTopics(r.brokers, cfg, r.topics); err != nil {
		return nil, err
	}
	return sarama.NewSyncProducer(r.brokers, cfg)
}

// producerConfig is the producer setting used for business events: all
// in-sync replicas acknowledge, the producer is idempotent and sends one request
// at a time, so retries cannot reorder events for a key.
func producerConfig() *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_6_0_0
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Idempotent = true
	cfg.Net.MaxOpenRequests = 1
	cfg.Producer.Retry.Max = 5
	cfg.Producer.Return.Successes = true
	cfg.Producer.Compression = sarama.CompressionSnappy
	return cfg
}

// ensureTopics creates each topic with the broker's default partition count and
// replication factor. An existing topic is not an error.
func ensureTopics(brokers []string, cfg *sarama.Config, topics []string) error {
	admin, err := sarama.NewClusterAdmin(brokers, cfg)
	if err != nil {
		return fmt.Errorf("kafka: cluster admin: %w", err)
	}
	defer admin.Close()
	for _, topic := range topics {
		err := admin.CreateTopic(topic, &sarama.TopicDetail{NumPartitions: -1, ReplicationFactor: -1}, false)
		if err != nil && !errors.Is(err, sarama.ErrTopicAlreadyExists) {
			return fmt.Errorf("kafka: create topic %s: %w", topic, err)
		}
	}
	return nil
}

func nextRelayBackoff(current time.Duration) time.Duration {
	if next := current * 2; next < relayRetryMax {
		return next
	}
	return relayRetryMax
}

// sleepCtx waits for d or until ctx is cancelled. It reports whether the wait
// completed.
func sleepCtx(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
