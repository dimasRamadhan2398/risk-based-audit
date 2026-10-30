package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"master-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
//	           flexible tenants only, additive: "parent_department_id"|null,"level",
//	           "level_name","name_key"|null,"path":[{"id","name","level","level_name","name_key"}]
//
// Events are written to the transactional outbox (models.KafkaOutboxEvent) and
// published by the Relay, so a broker outage delays events but does not lose them.
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

	// Set only on flexible tenants (ORG_STRUCTURE_MODE=flexible). A nil
	// pointer adds no keys, so the legacy payload is byte-for-byte unchanged.
	*DepartmentOrgData
}

// DepartmentOrgData is the additive part of a department event on flexible
// tenants (docs/org-structure-model-plan.md section 6).
type DepartmentOrgData struct {
	ParentDepartmentID *string                    `json:"parent_department_id"`
	Level              int                        `json:"level"`
	LevelName          string                     `json:"level_name"`
	NameKey            *string                    `json:"name_key"`
	Path               []models.DepartmentPathSeg `json:"path"`
}

// DomainPublisher records department events in the outbox. A nil
// *DomainPublisher is valid and records nothing.
type DomainPublisher struct {
	db          *gorm.DB
	serviceName string
	now         func() time.Time
}

// NewDomainPublisher returns a publisher that writes to the outbox of db.
func NewDomainPublisher(db *gorm.DB, serviceName string) *DomainPublisher {
	return &DomainPublisher{db: db, serviceName: serviceName, now: time.Now}
}

// Close is kept for callers that closed the broker connection. The outbox holds
// no connection, so it does nothing.
func (p *DomainPublisher) Close() error { return nil }

// PublishDepartment records one department event for dept.
func (p *DomainPublisher) PublishDepartment(ctx context.Context, eventType string, dept *models.Department) error {
	if p == nil || p.db == nil || dept == nil {
		return nil
	}
	env, key := p.departmentEnvelope(eventType, dept)
	body, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("kafka: marshal %s: %w", eventType, err)
	}
	return Enqueue(p.db.WithContext(ctx), TopicDepartments, key, eventType, body)
}

// PublishDepartmentOrgTx records one flexible-mode department event (with the
// additive org fields taken from dept's ParentDepartmentID, LevelName, NameKey
// and Path) in the outbox through tx, so it commits or rolls back with the
// write that caused it. A nil publisher records nothing.
func (p *DomainPublisher) PublishDepartmentOrgTx(ctx context.Context, tx *gorm.DB, eventType string, dept *models.Department) error {
	if p == nil || dept == nil {
		return nil
	}
	if tx == nil {
		tx = p.db
	}
	if tx == nil {
		return nil
	}
	env, key := p.departmentEnvelopeOrg(eventType, dept)
	body, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("kafka: marshal %s: %w", eventType, err)
	}
	return Enqueue(tx.WithContext(ctx), TopicDepartments, key, eventType, body)
}

func (p *DomainPublisher) departmentEnvelopeOrg(eventType string, dept *models.Department) (DomainEnvelope, string) {
	env, key := p.departmentEnvelope(eventType, dept)
	data := env.Data.(DepartmentData)
	data.DepartmentOrgData = &DepartmentOrgData{
		ParentDepartmentID: uuidPtrString(dept.ParentDepartmentID),
		Level:              dept.Level,
		LevelName:          dept.LevelName,
		NameKey:            dept.NameKey,
		Path:               dept.Path,
	}
	if data.Path == nil {
		data.Path = []models.DepartmentPathSeg{}
	}
	env.Data = data
	return env, key
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

// Enqueue appends one message to the outbox. Use it with the handle that
// performed the write, after that write has committed.
func Enqueue(db *gorm.DB, topic, key, eventType string, body []byte) error {
	row := models.KafkaOutboxEvent{
		Topic:      topic,
		MessageKey: key,
		EventType:  eventType,
		Payload:    string(body),
	}
	if err := db.Create(&row).Error; err != nil {
		return fmt.Errorf("kafka: outbox insert %s: %w", eventType, err)
	}
	return nil
}

// ResyncDepartments queues a department.updated event for every live
// department and returns how many were queued. Consumers apply events
// idempotently, so this repairs links that were never published.
func ResyncDepartments(ctx context.Context, db *gorm.DB, p *DomainPublisher) (int, error) {
	var depts []models.Department
	if err := db.WithContext(ctx).Order("created_at ASC").Find(&depts).Error; err != nil {
		return 0, fmt.Errorf("kafka: load departments: %w", err)
	}
	for i := range depts {
		if err := p.PublishDepartment(ctx, DomainEventDepartmentUpdated, &depts[i]); err != nil {
			return i, err
		}
	}
	return len(depts), nil
}

func uuidPtrString(id *uuid.UUID) *string {
	if id == nil || *id == uuid.Nil {
		return nil
	}
	s := id.String()
	return &s
}
