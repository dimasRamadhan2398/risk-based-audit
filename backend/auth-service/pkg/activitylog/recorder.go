package activitylog

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// Recorder is what write points call. Record never fails the caller: an event
// that cannot be published is written directly, and a write that fails is
// logged.
type Recorder interface {
	Record(ctx context.Context, e Entry)
}

// Store persists entries idempotently on EventID (repositories.SysAlertLogRepository).
type Store interface {
	// InsertEntry stores e unless a row with the same EventID exists.
	// inserted is false for a duplicate.
	InsertEntry(ctx context.Context, e *Entry) (inserted bool, err error)
}

// Publisher sends one message to a topic and waits for the broker ack
// (pkg/kafka.Producer.SendRaw).
type Publisher interface {
	SendRaw(topic, key string, value []byte) error
	IsEnabled() bool
}

// NopRecorder drops every event (tests, tools).
type NopRecorder struct{}

// Record implements Recorder.
func (NopRecorder) Record(context.Context, Entry) {}

// RecorderConfig configures NewKafkaRecorder.
type RecorderConfig struct {
	Topic  string
	Source string
	// QueueSize bounds the in-memory hand-off to the publish worker. When it is
	// full Record writes directly instead of blocking the request.
	QueueSize int
	// Cooldown is how long publishing is skipped (direct writes only) after a
	// publish error, so a dead broker costs one timeout, not one per event.
	Cooldown time.Duration
}

// KafkaRecorder publishes to the activity topic from a background worker and
// falls back to a direct DB write when publishing is not possible.
type KafkaRecorder struct {
	publisher Publisher
	store     Store
	topic     string
	source    string
	cooldown  time.Duration
	log       *zap.Logger
	now       func() time.Time

	queue  chan Entry
	wg     sync.WaitGroup
	mu     sync.RWMutex // guards closed and sends on queue
	closed bool
	// pausedUntil (unix nanos) skips publishing after a publish error.
	pausedUntil atomic.Int64
}

// NewKafkaRecorder starts the publish worker. publisher may be nil (Kafka
// disabled or not reachable at start-up): every event is then written directly.
func NewKafkaRecorder(cfg RecorderConfig, publisher Publisher, store Store, log *zap.Logger) *KafkaRecorder {
	if log == nil {
		log = zap.NewNop()
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 1024
	}
	if cfg.Cooldown <= 0 {
		cfg.Cooldown = 30 * time.Second
	}
	r := &KafkaRecorder{
		publisher: publisher,
		store:     store,
		topic:     cfg.Topic,
		source:    ResolveSource(cfg.Source),
		cooldown:  cfg.Cooldown,
		log:       log,
		now:       time.Now,
		queue:     make(chan Entry, cfg.QueueSize),
	}
	r.wg.Add(1)
	go r.run()
	return r
}

// Record implements Recorder. Actor and request metadata missing from e are
// taken from ctx (activitylog.WithActor / ContextMiddleware).
func (r *KafkaRecorder) Record(ctx context.Context, e Entry) {
	e = Enrich(ctx, e)
	e.Source = r.source
	e.Normalize(r.now())

	r.mu.RLock()
	queued := false
	if !r.closed {
		select {
		case r.queue <- e:
			queued = true
		default:
		}
	}
	r.mu.RUnlock()
	if !queued {
		r.writeDirect(e, "publish queue full or recorder closed")
	}
}

// Close stops accepting events and waits (up to timeout) for queued ones to
// be published or written.
func (r *KafkaRecorder) Close(timeout time.Duration) {
	r.mu.Lock()
	if !r.closed {
		r.closed = true
		close(r.queue)
	}
	r.mu.Unlock()
	done := make(chan struct{})
	go func() { r.wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(timeout):
		r.log.Warn("activity log: close timed out with events still queued")
	}
}

func (r *KafkaRecorder) run() {
	defer r.wg.Done()
	for e := range r.queue {
		r.deliver(e)
	}
}

func (r *KafkaRecorder) deliver(e Entry) {
	if !r.canPublish() {
		r.writeDirect(e, "kafka unavailable")
		return
	}
	payload, err := json.Marshal(e)
	if err != nil {
		r.writeDirect(e, "marshal failed")
		return
	}
	if err := r.publisher.SendRaw(r.topic, e.EventID, payload); err != nil {
		r.pausedUntil.Store(r.now().Add(r.cooldown).UnixNano())
		r.log.Warn("activity log: publish failed, writing directly",
			zap.String("topic", r.topic), zap.String("event_id", e.EventID), zap.Error(err))
		r.writeDirect(e, "publish failed")
		return
	}
}

func (r *KafkaRecorder) canPublish() bool {
	if r.publisher == nil || !r.publisher.IsEnabled() {
		return false
	}
	return r.now().UnixNano() >= r.pausedUntil.Load()
}

func (r *KafkaRecorder) writeDirect(e Entry, reason string) {
	if r.store == nil {
		r.log.Error("activity log: event dropped, no store", zap.String("event_id", e.EventID), zap.String("action", e.Action))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := r.store.InsertEntry(ctx, &e); err != nil {
		r.log.Error("activity log: direct write failed, event lost",
			zap.String("reason", reason), zap.String("event_id", e.EventID), zap.String("action", e.Action), zap.Error(err))
	}
}

// Enrich fills the actor, IP and user agent of e from ctx when e leaves them
// empty. An explicit actor (e.g. the user who just logged in) is kept.
func Enrich(ctx context.Context, e Entry) Entry {
	if a, ok := ActorFrom(ctx); ok {
		switch {
		case e.ActorUserID == "" && e.ActorUsername == "":
			e.ActorUserID, e.ActorUsername = a.UserID, a.Username
		case e.ActorUsername == "" && e.ActorUserID == a.UserID:
			e.ActorUsername = a.Username
		}
	}
	meta := RequestMetaFrom(ctx)
	if e.IPAddress == "" {
		e.IPAddress = meta.IPAddress
	}
	if e.UserAgent == "" {
		e.UserAgent = meta.UserAgent
	}
	return e
}
