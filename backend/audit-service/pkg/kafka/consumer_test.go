package kafka

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// fakeSession implements only what processMessage uses. Calling any other
// method panics (the embedded interface is nil).
type fakeSession struct {
	sarama.ConsumerGroupSession
	ctx    context.Context
	marked []*sarama.ConsumerMessage
}

func (f *fakeSession) Context() context.Context { return f.ctx }

func (f *fakeSession) MarkMessage(m *sarama.ConsumerMessage, _ string) {
	f.marked = append(f.marked, m)
}

func fastRetries(t *testing.T) {
	t.Helper()
	min, max := consumerRetryMin, consumerRetryMax
	consumerRetryMin, consumerRetryMax = time.Millisecond, 4*time.Millisecond
	t.Cleanup(func() { consumerRetryMin, consumerRetryMax = min, max })
}

// A failed handler is retried in place. The offset is marked only after the
// handler succeeds, so the failed message is not skipped.
func TestProcessMessageRetriesUntilSuccess(t *testing.T) {
	fastRetries(t)
	var calls atomic.Int32
	h := GenericMessageHandler(func(ctx context.Context, msg []byte) error {
		if calls.Add(1) < 3 {
			return errors.New("db down")
		}
		return nil
	})
	c := &Consumer{handlers: map[string]MessageHandler{"t": h}, logger: zap.NewNop(), ctx: context.Background()}
	sess := &fakeSession{ctx: context.Background()}
	msg := &sarama.ConsumerMessage{Topic: "t", Offset: 7, Value: []byte("x")}

	c.processMessage(sess, msg)

	if calls.Load() != 3 {
		t.Fatalf("handler called %d times, want 3", calls.Load())
	}
	if len(sess.marked) != 1 || sess.marked[0] != msg {
		t.Fatalf("message not marked exactly once after success: %v", sess.marked)
	}
}

// When the session ends while a message is failing, the message is left
// unmarked so it is redelivered to the next session.
func TestProcessMessageLeavesMessageWhenSessionEnds(t *testing.T) {
	fastRetries(t)
	h := GenericMessageHandler(func(ctx context.Context, msg []byte) error { return errors.New("db down") })
	c := &Consumer{handlers: map[string]MessageHandler{"t": h}, logger: zap.NewNop(), ctx: context.Background()}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	sess := &fakeSession{ctx: ctx}

	c.processMessage(sess, &sarama.ConsumerMessage{Topic: "t", Offset: 1})

	if len(sess.marked) != 0 {
		t.Fatalf("failed message was marked: %v", sess.marked)
	}
}

func TestNextBackoffIsCapped(t *testing.T) {
	d := time.Second
	for i := 0; i < 10; i++ {
		d = nextBackoff(d, 30*time.Second)
	}
	if d != 30*time.Second {
		t.Fatalf("backoff %s, want 30s", d)
	}
}
