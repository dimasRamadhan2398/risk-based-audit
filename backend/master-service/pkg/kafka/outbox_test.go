package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"master-service/models"

	"github.com/IBM/sarama"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func outboxDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&models.KafkaOutboxEvent{}); err != nil {
		t.Fatal(err)
	}
	return db
}

// fakeSender records sent messages and fails on the Nth send when failAt > 0.
type fakeSender struct {
	sent   []*sarama.ProducerMessage
	failAt int
}

func (f *fakeSender) SendMessage(msg *sarama.ProducerMessage) (int32, int64, error) {
	if f.failAt > 0 && len(f.sent)+1 == f.failAt {
		return 0, 0, errors.New("broker unavailable")
	}
	f.sent = append(f.sent, msg)
	return 0, int64(len(f.sent)), nil
}

func TestDepartmentEventIsQueuedNotSent(t *testing.T) {
	db := outboxDB(t)
	deptID := uuid.New()
	p := NewDomainPublisher(db, "master-service")
	dept := &models.Department{ID: deptID, DepartmentCode: "FIN", DepartmentName: "Finance", IsActive: true, CompanyID: uuid.New()}

	if err := p.PublishDepartment(context.Background(), DomainEventDepartmentCreated, dept); err != nil {
		t.Fatalf("publish: %v", err)
	}

	var rows []models.KafkaOutboxEvent
	if err := db.Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 outbox row, got %d", len(rows))
	}
	row := rows[0]
	if row.Topic != TopicDepartments || row.MessageKey != deptID.String() || row.EventType != DomainEventDepartmentCreated {
		t.Fatalf("unexpected row: %+v", row)
	}
	var env DomainEnvelope
	if err := json.Unmarshal([]byte(row.Payload), &env); err != nil {
		t.Fatalf("payload is not an envelope: %v", err)
	}
	if env.Version != DomainEnvelopeVersion || env.EventType != DomainEventDepartmentCreated || env.Producer != "master-service" {
		t.Fatalf("bad envelope: %+v", env)
	}
}

func TestNilPublisherQueuesNothing(t *testing.T) {
	db := outboxDB(t)
	var p *DomainPublisher
	if err := p.PublishDepartment(context.Background(), DomainEventDepartmentUpdated, &models.Department{}); err != nil {
		t.Fatal(err)
	}
	var n int64
	db.Model(&models.KafkaOutboxEvent{}).Count(&n)
	if n != 0 {
		t.Fatalf("nil publisher wrote %d rows", n)
	}
}

func TestResyncQueuesEveryLiveDepartment(t *testing.T) {
	db := outboxDB(t)
	// AutoMigrate cannot build departments on sqlite (uuid defaults), so the
	// table is created with the columns the query uses.
	if err := db.Exec(`CREATE TABLE departments (id TEXT PRIMARY KEY, department_code TEXT, department_name TEXT,
		is_active NUMERIC, company_id TEXT, business_unit_id TEXT, created_at DATETIME, deleted_at DATETIME)`).Error; err != nil {
		t.Fatal(err)
	}
	for _, code := range []string{"A", "B", "C"} {
		if err := db.Exec(`INSERT INTO departments (id, department_code, department_name, is_active, company_id, created_at)
			VALUES (?, ?, ?, 1, ?, CURRENT_TIMESTAMP)`, uuid.NewString(), code, "Dept "+code, uuid.NewString()).Error; err != nil {
			t.Fatal(err)
		}
	}
	p := NewDomainPublisher(db, "master-service")
	n, err := ResyncDepartments(context.Background(), db, p)
	if err != nil || n != 3 {
		t.Fatalf("resync: n=%d err=%v", n, err)
	}
	var rows []models.KafkaOutboxEvent
	db.Order("id ASC").Find(&rows)
	for _, r := range rows {
		if r.EventType != DomainEventDepartmentUpdated {
			t.Fatalf("resync event type %q", r.EventType)
		}
	}
}

func TestRelayPublishesInOrderAndRemovesRows(t *testing.T) {
	db := outboxDB(t)
	for i := 0; i < 3; i++ {
		if err := Enqueue(db, TopicDepartments, "k", "department.updated", []byte(`{"n":`+string(rune('0'+i))+`}`)); err != nil {
			t.Fatal(err)
		}
	}
	sender := &fakeSender{}
	r := NewRelay(db, nil, TopicDepartments)
	n, err := r.publishBatch(context.Background(), sender)
	if err != nil || n != 3 {
		t.Fatalf("publishBatch: n=%d err=%v", n, err)
	}
	for i, msg := range sender.sent {
		if got := string(msg.Value.(sarama.ByteEncoder)); got != `{"n":`+string(rune('0'+i))+`}` {
			t.Fatalf("message %d out of order: %s", i, got)
		}
	}
	var left int64
	db.Model(&models.KafkaOutboxEvent{}).Count(&left)
	if left != 0 {
		t.Fatalf("%d rows left after publish", left)
	}
}

// A failed send must stop the batch: the rows before it are removed, the
// failed row keeps its error, and the rows after it wait so order is kept.
func TestRelayStopsAtFailedSend(t *testing.T) {
	db := outboxDB(t)
	for i := 0; i < 3; i++ {
		if err := Enqueue(db, TopicDepartments, "k", "department.updated", []byte(`{}`)); err != nil {
			t.Fatal(err)
		}
	}
	sender := &fakeSender{failAt: 2}
	r := NewRelay(db, nil, TopicDepartments)
	n, err := r.publishBatch(context.Background(), sender)
	if err == nil {
		t.Fatal("publishBatch must report the failed send")
	}
	if n != 1 {
		t.Fatalf("want 1 published before the failure, got %d", n)
	}
	var rows []models.KafkaOutboxEvent
	db.Order("id ASC").Find(&rows)
	if len(rows) != 2 {
		t.Fatalf("want 2 rows left, got %d", len(rows))
	}
	if rows[0].Attempts != 1 || rows[0].LastError == "" {
		t.Fatalf("failed row not recorded: %+v", rows[0])
	}
	if rows[1].Attempts != 0 {
		t.Fatalf("row after the failure was touched: %+v", rows[1])
	}
}

func TestBackoffIsCapped(t *testing.T) {
	d := relayRetryMin
	for i := 0; i < 20; i++ {
		d = nextRelayBackoff(d)
	}
	if d != relayRetryMax {
		t.Fatalf("backoff %s, want cap %s", d, relayRetryMax)
	}
	if !sleepCtx(context.Background(), time.Millisecond) {
		t.Fatal("sleepCtx should complete")
	}
}
