package activitylog

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// MessageProcessor turns one topic message into a stored row.
type MessageProcessor struct {
	store  Store
	source string
	log    *zap.Logger
	now    func() time.Time
}

// NewMessageProcessor stores messages stamped with source and ignores the rest.
func NewMessageProcessor(store Store, source string, log *zap.Logger) *MessageProcessor {
	if log == nil {
		log = zap.NewNop()
	}
	return &MessageProcessor{store: store, source: ResolveSource(source), log: log, now: time.Now}
}

// Process handles one message. A returned error is transient (the store
// failed) and the message must be retried; malformed messages and messages
// from another stack are logged and dropped (nil).
//
// The source check is the tenant-isolation backstop: even if two stacks were
// pointed at the same topic, each stores only its own events in its own DB.
func (p *MessageProcessor) Process(ctx context.Context, value []byte) error {
	var e Entry
	if err := json.Unmarshal(value, &e); err != nil {
		p.log.Warn("activity log: dropping malformed message", zap.Error(err))
		return nil
	}
	if e.Source != p.source {
		p.log.Warn("activity log: dropping message from another stack",
			zap.String("source", e.Source), zap.String("want", p.source), zap.String("event_id", e.EventID))
		return nil
	}
	if e.EventID == "" || e.Action == "" {
		p.log.Warn("activity log: dropping message without event_id or action")
		return nil
	}
	e.Normalize(p.now())
	if _, err := p.store.InsertEntry(ctx, &e); err != nil {
		return err
	}
	return nil
}

// ConsumerConfig configures RunConsumer.
type ConsumerConfig struct {
	Brokers []string
	Topic   string
	GroupID string
	Version string
	// MaxBackoff caps the reconnect/retry delay (default 60s).
	MaxBackoff time.Duration
}

// Consumer runs the activity-log consumer group in the background. It never
// exits on a broker error: connecting, consuming and storing all retry with
// exponential backoff until Stop.
type Consumer struct {
	cfg       ConsumerConfig
	processor *MessageProcessor
	log       *zap.Logger
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// StartConsumer starts the background loop and returns at once, so a broker
// that is down at boot does not delay or crash the HTTP server.
func StartConsumer(cfg ConsumerConfig, processor *MessageProcessor, log *zap.Logger) *Consumer {
	if log == nil {
		log = zap.NewNop()
	}
	if cfg.MaxBackoff <= 0 {
		cfg.MaxBackoff = 60 * time.Second
	}
	ctx, cancel := context.WithCancel(context.Background())
	c := &Consumer{cfg: cfg, processor: processor, log: log, cancel: cancel}
	c.wg.Add(1)
	go c.loop(ctx)
	return c
}

// Stop cancels the loop and waits for it to finish.
func (c *Consumer) Stop() {
	if c == nil {
		return
	}
	c.cancel()
	c.wg.Wait()
}

func (c *Consumer) loop(ctx context.Context) {
	defer c.wg.Done()
	backoff := newBackoff(time.Second, c.cfg.MaxBackoff)
	for ctx.Err() == nil {
		group, err := c.newGroup()
		if err != nil {
			d := backoff.next()
			c.log.Warn("activity log consumer: cannot connect, retrying",
				zap.Strings("brokers", c.cfg.Brokers), zap.Duration("in", d), zap.Error(err))
			if !sleep(ctx, d) {
				return
			}
			continue
		}
		c.log.Info("activity log consumer started",
			zap.String("topic", c.cfg.Topic), zap.String("group_id", c.cfg.GroupID))
		backoff.reset()
		for ctx.Err() == nil {
			started := time.Now()
			err := group.Consume(ctx, []string{c.cfg.Topic}, &groupHandler{c: c})
			if err == nil && time.Since(started) < time.Second {
				// A session that ends at once (e.g. the topic does not exist
				// yet) must not spin.
				if !sleep(ctx, time.Second) {
					break
				}
			}
			if err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					break
				}
				d := backoff.next()
				c.log.Warn("activity log consumer: consume error, retrying", zap.Duration("in", d), zap.Error(err))
				if !sleep(ctx, d) {
					break
				}
				continue
			}
			backoff.reset()
		}
		_ = group.Close()
	}
}

func (c *Consumer) newGroup() (sarama.ConsumerGroup, error) {
	sc := sarama.NewConfig()
	if c.cfg.Version != "" {
		if v, err := sarama.ParseKafkaVersion(c.cfg.Version); err == nil {
			sc.Version = v
		}
	}
	// A new group starts at the oldest retained message so events published
	// before the first start are not skipped; replays are idempotent.
	sc.Consumer.Offsets.Initial = sarama.OffsetOldest
	sc.Consumer.Offsets.AutoCommit.Enable = true
	sc.Consumer.Return.Errors = false
	sc.Net.DialTimeout = 10 * time.Second
	sc.Metadata.Retry.Max = 1
	return sarama.NewConsumerGroup(c.cfg.Brokers, c.cfg.GroupID, sc)
}

type groupHandler struct{ c *Consumer }

func (h *groupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *groupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim marks a message only after it is stored. A store failure is
// retried in place with backoff (offsets must not move past an unsaved event).
func (h *groupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := session.Context()
	for {
		select {
		case <-ctx.Done():
			return nil
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			backoff := newBackoff(500*time.Millisecond, h.c.cfg.MaxBackoff)
			for {
				err := h.c.processor.Process(ctx, msg.Value)
				if err == nil {
					break
				}
				d := backoff.next()
				h.c.log.Error("activity log consumer: store failed, retrying",
					zap.Int64("offset", msg.Offset), zap.Duration("in", d), zap.Error(err))
				if !sleep(ctx, d) {
					return nil
				}
			}
			session.MarkMessage(msg, "")
		}
	}
}

type backoff struct {
	min, max, cur time.Duration
}

func newBackoff(min, max time.Duration) *backoff {
	if max < min {
		max = min
	}
	return &backoff{min: min, max: max}
}

func (b *backoff) next() time.Duration {
	if b.cur == 0 {
		b.cur = b.min
	} else {
		b.cur *= 2
		if b.cur > b.max {
			b.cur = b.max
		}
	}
	return b.cur
}

func (b *backoff) reset() { b.cur = 0 }

// sleep waits d or until ctx is done; false means ctx is done.
func sleep(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
