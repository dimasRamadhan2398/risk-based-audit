package activitylog

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"auth-service/pkg/kafka"

	"github.com/google/uuid"
)

// memStore is an idempotent in-memory Store.
type memStore struct {
	mu   sync.Mutex
	rows map[string]Entry
	fail error
}

func newMemStore() *memStore { return &memStore{rows: map[string]Entry{}} }

func (m *memStore) InsertEntry(_ context.Context, e *Entry) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.fail != nil {
		return false, m.fail
	}
	if _, ok := m.rows[e.EventID]; ok {
		return false, nil
	}
	m.rows[e.EventID] = *e
	return true, nil
}

func (m *memStore) all() []Entry {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Entry, 0, len(m.rows))
	for _, e := range m.rows {
		out = append(out, e)
	}
	return out
}

type fakePublisher struct {
	mu      sync.Mutex
	enabled bool
	err     error
	sent    [][]byte
	topics  []string
}

func (p *fakePublisher) SendRaw(topic, key string, value []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.err != nil {
		return p.err
	}
	p.topics = append(p.topics, topic)
	p.sent = append(p.sent, value)
	return nil
}
func (p *fakePublisher) IsEnabled() bool { return p.enabled }

func TestRedactRemovesSecretsRecursively(t *testing.T) {
	in := map[string]interface{}{
		"password":           "hunter2",
		"new_password":       "x",
		"temporary_password": "Temp#123",
		"step_up_token":      "eyJ...",
		"mfa_secret":         "JBSWY3DP",
		"code":               "123456",
		"jti":                "abc",
		"role":               "ADMIN",
		"nested": map[string]interface{}{
			"Authorization": "Bearer x",
			"device_name":   "Laptop",
		},
		"list": []interface{}{map[string]interface{}{"api_key": "k", "ok": 1}},
	}
	out := Redact(in)
	for _, k := range []string{"password", "new_password", "temporary_password", "step_up_token", "mfa_secret", "code", "jti"} {
		if out[k] != Redacted {
			t.Errorf("%s = %v, want redacted", k, out[k])
		}
	}
	if out["role"] != "ADMIN" {
		t.Errorf("role was redacted")
	}
	nested := out["nested"].(map[string]interface{})
	if nested["Authorization"] != Redacted || nested["device_name"] != "Laptop" {
		t.Errorf("nested = %v", nested)
	}
	item := out["list"].([]interface{})[0].(map[string]interface{})
	if item["api_key"] != Redacted || item["ok"] != 1 {
		t.Errorf("list item = %v", item)
	}
	if in["password"] != "hunter2" {
		t.Error("Redact mutated its input")
	}
	b, _ := json.Marshal(out)
	for _, secret := range []string{"hunter2", "Temp#123", "JBSWY3DP", "123456", "Bearer x"} {
		if strings.Contains(string(b), secret) {
			t.Errorf("secret %q survived redaction", secret)
		}
	}
}

func TestMaskIP(t *testing.T) {
	cases := map[string]string{
		"192.168.10.25":         "192.168.x.x",
		"10.0.0.1":              "10.0.x.x",
		"::ffff:172.16.5.4":     "172.16.x.x",
		"2001:db8:85a3::8a2e:1": "2001:db8:x:x:x:x:x:x",
		"fe80::1":               "fe80:0:x:x:x:x:x:x",
		"":                      "",
		"not-an-ip":             "x.x.x.x",
	}
	for in, want := range cases {
		if got := MaskIP(in); got != want {
			t.Errorf("MaskIP(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeBoundsAndDefaults(t *testing.T) {
	now := time.Date(2026, 11, 4, 1, 2, 3, 456789123, time.FixedZone("WIB", 7*3600))
	e := Entry{
		Action:      " User.Login ",
		Result:      "weird",
		IPAddress:   "203.0.113.9, 10.0.0.1",
		ActorUserID: "not-a-uuid",
		UserAgent:   strings.Repeat("a", 2000),
		Summary:     strings.Repeat("é", 400),
		Details:     map[string]interface{}{"password": "p"},
	}
	e.Normalize(now)
	if e.EventID == "" || e.Version != EnvelopeVersion {
		t.Fatalf("defaults not set: %+v", e)
	}
	if e.Action != "user.login" || e.Result != ResultSuccess || e.IPAddress != "203.0.113.9" || e.ActorUserID != "" {
		t.Fatalf("normalised = %+v", e)
	}
	if !e.OccurredAt.Equal(now.Truncate(time.Microsecond)) || e.OccurredAt.Location() != time.UTC {
		t.Fatalf("occurred_at = %v", e.OccurredAt)
	}
	if len(e.UserAgent) != maxUserAgent || len(e.Summary) > maxSummary || !strings.HasSuffix(e.Summary, "é") {
		t.Fatalf("bounds: ua %d summary %d", len(e.UserAgent), len(e.Summary))
	}
	if e.Details["password"] != Redacted {
		t.Fatal("details not redacted")
	}
}

func TestResolveTopicIsTenantScoped(t *testing.T) {
	cases := []struct{ explicit, prefix, service, want string }{
		{"", "", "auth-service", "activity-logs"},
		{"", "", "", "activity-logs"},
		{"", "bai", "auth-service-bai", "bai.activity-logs"},
		{"", "", "auth-service-bai", "bai.activity-logs"}, // tenant stacks do not set the prefix today
		{"", ".acme.", "auth-service-bai", "acme.activity-logs"},
		{"custom", "bai", "auth-service-bai", "custom"},
	}
	for _, tc := range cases {
		if got := ResolveTopic(tc.explicit, tc.prefix, tc.service); got != tc.want {
			t.Errorf("ResolveTopic(%q,%q,%q) = %q, want %q", tc.explicit, tc.prefix, tc.service, got, tc.want)
		}
	}
	if g := ResolveGroupID("", "auth-service-bai"); g != "auth-service-bai-activity-log" {
		t.Errorf("group = %q", g)
	}
}

func TestProcessorIsIdempotentAndIgnoresOtherStacks(t *testing.T) {
	store := newMemStore()
	p := NewMessageProcessor(store, "auth-service-bai", nil)
	msg := func(source, id string) []byte {
		b, _ := json.Marshal(Entry{Version: 1, EventID: id, Source: source, Action: ActionLogin, Result: ResultSuccess,
			OccurredAt: time.Now(), Details: map[string]interface{}{"token": "t0k"}})
		return b
	}
	for i := 0; i < 3; i++ { // redelivery
		if err := p.Process(context.Background(), msg("auth-service-bai", "e1")); err != nil {
			t.Fatal(err)
		}
	}
	if err := p.Process(context.Background(), msg("auth-service-acme", "e2")); err != nil {
		t.Fatal(err)
	}
	if err := p.Process(context.Background(), []byte("{not json")); err != nil {
		t.Fatalf("malformed message must be dropped, got %v", err)
	}
	rows := store.all()
	if len(rows) != 1 || rows[0].EventID != "e1" {
		t.Fatalf("stored %v, want only e1", rows)
	}
	if rows[0].Details["token"] != Redacted {
		t.Fatal("consumer did not redact")
	}

	store.fail = errors.New("db down")
	if err := p.Process(context.Background(), msg("auth-service-bai", "e3")); err == nil {
		t.Fatal("store failure must be returned so the message is retried")
	}
}

func TestRecorderPublishesToTopic(t *testing.T) {
	store := newMemStore()
	pub := &fakePublisher{enabled: true}
	r := NewKafkaRecorder(RecorderConfig{Topic: "bai.activity-logs", Source: "auth-service-bai"}, pub, store, nil)
	r.Record(context.Background(), Entry{Action: ActionLogin, ActorUsername: "alice"})
	r.Close(time.Second)

	if len(pub.sent) != 1 || pub.topics[0] != "bai.activity-logs" {
		t.Fatalf("published %d to %v", len(pub.sent), pub.topics)
	}
	if len(store.all()) != 0 {
		t.Fatal("published event must not also be written directly")
	}
	var e Entry
	_ = json.Unmarshal(pub.sent[0], &e)
	if e.Source != "auth-service-bai" || e.EventID == "" || e.Action != ActionLogin {
		t.Fatalf("message = %+v", e)
	}
	// The consumer of the same stack stores it.
	if err := NewMessageProcessor(store, "auth-service-bai", nil).Process(context.Background(), pub.sent[0]); err != nil {
		t.Fatal(err)
	}
	if len(store.all()) != 1 {
		t.Fatal("consumer did not store the published event")
	}
}

func TestRecorderFallsBackToDirectWrite(t *testing.T) {
	cases := map[string]Publisher{
		"no producer": nil,
		"disabled":    &fakePublisher{enabled: false},
		"send error":  &fakePublisher{enabled: true, err: errors.New("broker down")},
	}
	for name, pub := range cases {
		store := newMemStore()
		r := NewKafkaRecorder(RecorderConfig{Topic: "activity-logs"}, pub, store, nil)
		r.Record(context.Background(), Entry{Action: ActionLoginFailed, ActorUsername: "bob", Result: ResultFailure})
		r.Record(context.Background(), Entry{Action: ActionLogout, ActorUsername: "bob"})
		r.Close(time.Second)
		if n := len(store.all()); n != 2 {
			t.Errorf("%s: %d rows written directly, want 2", name, n)
		}
	}
}

func TestRecorderAfterCloseWritesDirectly(t *testing.T) {
	store := newMemStore()
	r := NewKafkaRecorder(RecorderConfig{}, &fakePublisher{enabled: true}, store, nil)
	r.Close(time.Second)
	r.Record(context.Background(), Entry{Action: ActionLogin})
	if len(store.all()) != 1 {
		t.Fatal("event recorded after Close was lost")
	}
}

func TestEnrichUsesContextActorAndMeta(t *testing.T) {
	ctx := WithActor(context.Background(), Actor{UserID: uuid.NewString(), Username: "admin"})
	ctx = WithRequestMeta(ctx, RequestMeta{IPAddress: "198.51.100.7", UserAgent: "UA"})

	e := Enrich(ctx, Entry{Action: ActionRoleCreated})
	if e.ActorUsername != "admin" || e.IPAddress != "198.51.100.7" || e.UserAgent != "UA" {
		t.Fatalf("enriched = %+v", e)
	}
	// An explicit actor (login) is kept.
	e = Enrich(ctx, Entry{ActorUsername: "someone-else"})
	if e.ActorUsername != "someone-else" || e.ActorUserID != "" {
		t.Fatalf("explicit actor overwritten: %+v", e)
	}
}

func TestFromKafkaEventMapping(t *testing.T) {
	admin := Actor{UserID: uuid.NewString(), Username: "admin"}
	ctx := WithActor(context.Background(), admin)

	e, ok := FromKafkaEvent(ctx, &kafka.Event{Type: kafka.EventRoleAssigned, Success: true,
		Data: map[string]interface{}{"role": "AUDITOR", OperationKey: ActionRoleCreated}})
	if !ok || e.Action != ActionRoleCreated || e.EntityType != "role" || e.EntityID != "AUDITOR" || e.ActorUsername != "admin" {
		t.Fatalf("role event = %+v", e)
	}
	if _, has := e.Details[OperationKey]; has {
		t.Fatal("operation key leaked into details")
	}

	user := uuid.NewString()
	e, ok = FromKafkaEvent(context.Background(), &kafka.Event{Type: kafka.EventStepUp, UserID: user, Username: "carol",
		Success: false, ErrorReason: "invalid_password", Data: map[string]interface{}{"purpose": "org_structure", "jti": "abc"}})
	if !ok || e.Action != ActionStepUp || e.Result != ResultFailure || e.ActorUserID != user || e.Details["reason"] != "invalid_password" {
		t.Fatalf("step-up = %+v", e)
	}
	e.Normalize(time.Now())
	if e.Details["jti"] != Redacted {
		t.Fatal("step-up jti not redacted")
	}

	e, ok = FromKafkaEvent(context.Background(), &kafka.Event{Type: kafka.EventTrustedDeviceAdded, UserID: user, Username: "carol",
		Success: true, Data: map[string]interface{}{"device_id": "d1"}})
	if !ok || e.Action != ActionDeviceTrustedAdded || e.EntityID != "d1" {
		t.Fatalf("device = %+v", e)
	}
	e, ok = FromKafkaEvent(context.Background(), &kafka.Event{Type: kafka.EventMFAFailed, UserID: user, Success: false})
	if !ok || e.Action != ActionMFAFailed || e.Result != ResultFailure {
		t.Fatalf("mfa = %+v", e)
	}
	if _, ok := FromKafkaEvent(context.Background(), &kafka.Event{Type: kafka.EventTokenRefresh}); ok {
		t.Fatal("unmapped type must be skipped")
	}
}

// The publisher tap records events even when Kafka is disabled.
func TestTapRecordsWhenKafkaDisabled(t *testing.T) {
	store := newMemStore()
	rec := NewKafkaRecorder(RecorderConfig{}, nil, store, nil)
	pub := kafka.NewEventPublisher(kafka.EventPublisherConfig{Enabled: false, Tap: Tap(rec)})
	if err := pub.PublishMFAEvent(context.Background(), uuid.NewString(), "dave", kafka.EventMFAEnabled, true, "10.1.2.3"); err != nil {
		t.Fatal(err)
	}
	rec.Close(time.Second)
	rows := store.all()
	if len(rows) != 1 || rows[0].Action != ActionMFAEnabled || rows[0].IPAddress != "10.1.2.3" {
		t.Fatalf("rows = %+v", rows)
	}
}

type fakePurger struct{ cutoff time.Time }

func (f *fakePurger) DeleteOlderThan(_ context.Context, cutoff time.Time) (int64, error) {
	f.cutoff = cutoff
	return 3, nil
}

func TestPurgeOnceUsesRetentionDays(t *testing.T) {
	p := &fakePurger{}
	now := time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)
	n, err := PurgeOnce(context.Background(), p, 90, now, nil)
	if err == nil && n != 3 {
		t.Fatalf("n = %d", n)
	}
	if !p.cutoff.Equal(now.AddDate(0, 0, -90)) {
		t.Fatalf("cutoff = %v", p.cutoff)
	}
}
