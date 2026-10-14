package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"master-service/models"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

// Department business events. Consumed by audit-service (services/annual_plan_sync
// there). These go to their own topic, never to the service-logs stream that
// the Producer in producer.go writes to.
//
//	topic      department.events
//	key        department.id
//	envelope   {"version":1,"event_id","event_type","occurred_at","producer","data"}
//	event_type department.created | department.updated | department.deleted
//	data       {"id","department_code","department_name","is_active","company_id","business_unit_id"|null}
const (
	TopicDepartments = "department.events"

	DomainEnvelopeVersion = 1

	DomainEventDepartmentCreated = "department.created"
	DomainEventDepartmentUpdated = "department.updated"
	DomainEventDepartmentDeleted = "department.deleted"
)

// DomainEnvelope wraps every business event published by this service.
type DomainEnvelope struct {
	Version    int       `json:"version"`
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	OccurredAt time.Time `json:"occurred_at"`
	Producer   string    `json:"producer"`
	Data       any       `json:"data"`
}

// DepartmentData is the data of every department event.
type DepartmentData struct {
	ID             string  `json:"id"`
	DepartmentCode string  `json:"department_code"`
	DepartmentName string  `json:"department_name"`
	IsActive       bool    `json:"is_active"`
	CompanyID      string  `json:"company_id"`
	BusinessUnitID *string `json:"business_unit_id"`
}

// DomainPublisher sends business events. A nil *DomainPublisher is valid and
// publishes nothing.
type DomainPublisher struct {
	producer    sarama.SyncProducer
	serviceName string
	now         func() time.Time
}

// NewDomainPublisher connects a synchronous producer with WaitForAll acks.
func NewDomainPublisher(brokers []string, serviceName string) (*DomainPublisher, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Retry.Max = 3
	cfg.Producer.Return.Successes = true
	cfg.Producer.Compression = sarama.CompressionSnappy

	producer, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		return nil, fmt.Errorf("kafka: create domain producer: %w", err)
	}
	return &DomainPublisher{producer: producer, serviceName: serviceName, now: time.Now}, nil
}

// Close releases the broker connection.
func (p *DomainPublisher) Close() error {
	if p == nil || p.producer == nil {
		return nil
	}
	return p.producer.Close()
}

// PublishDepartment emits one department event for dept.
func (p *DomainPublisher) PublishDepartment(ctx context.Context, eventType string, dept *models.Department) error {
	if p == nil || p.producer == nil || dept == nil {
		return nil
	}
	_ = ctx // sarama's SyncProducer has no context hook
	env, key := p.departmentEnvelope(eventType, dept)
	body, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("kafka: marshal %s: %w", eventType, err)
	}
	_, _, err = p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: TopicDepartments,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(body),
	})
	if err != nil {
		return fmt.Errorf("kafka: send %s: %w", eventType, err)
	}
	return nil
}

// departmentEnvelope builds the wire envelope and the partition key (the
// department id) for one event.
func (p *DomainPublisher) departmentEnvelope(eventType string, dept *models.Department) (DomainEnvelope, string) {
	data := DepartmentData{
		ID:             dept.ID.String(),
		DepartmentCode: dept.DepartmentCode,
		DepartmentName: dept.DepartmentName,
		IsActive:       dept.IsActive,
		CompanyID:      dept.CompanyID.String(),
		BusinessUnitID: uuidPtrString(dept.BusinessUnitID),
	}
	return DomainEnvelope{
		Version:    DomainEnvelopeVersion,
		EventID:    uuid.NewString(),
		EventType:  eventType,
		OccurredAt: p.now().UTC(),
		Producer:   p.serviceName,
		Data:       data,
	}, data.ID
}

func uuidPtrString(id *uuid.UUID) *string {
	if id == nil || *id == uuid.Nil {
		return nil
	}
	s := id.String()
	return &s
}
