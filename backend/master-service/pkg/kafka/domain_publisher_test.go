package kafka

import (
	"encoding/json"
	"testing"
	"time"

	"master-service/models"

	"github.com/google/uuid"
)

func TestDepartmentEnvelopeShape(t *testing.T) {
	bu := uuid.New()
	dept := &models.Department{
		ID:             uuid.New(),
		DepartmentCode: "FIN",
		DepartmentName: "Finance",
		IsActive:       true,
		CompanyID:      uuid.New(),
		BusinessUnitID: &bu,
	}
	p := &DomainPublisher{serviceName: "master-service", now: func() time.Time { return time.Date(2026, 10, 14, 1, 2, 3, 0, time.UTC) }}

	env, key := p.departmentEnvelope(DomainEventDepartmentUpdated, dept)
	if key != dept.ID.String() {
		t.Fatalf("key = %s, want department id", key)
	}
	raw, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Version    int    `json:"version"`
		EventType  string `json:"event_type"`
		Producer   string `json:"producer"`
		OccurredAt string `json:"occurred_at"`
		Data       struct {
			ID             string  `json:"id"`
			DepartmentName string  `json:"department_name"`
			BusinessUnitID *string `json:"business_unit_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.Version != 1 || got.EventType != "department.updated" || got.Producer != "master-service" {
		t.Fatalf("unexpected envelope: %+v", got)
	}
	if got.Data.ID != dept.ID.String() || got.Data.DepartmentName != "Finance" {
		t.Fatalf("unexpected data: %+v", got.Data)
	}
	if got.Data.BusinessUnitID == nil || *got.Data.BusinessUnitID != bu.String() {
		t.Fatalf("business_unit_id not carried through")
	}
}

func TestNilDomainPublisherIsNoop(t *testing.T) {
	var p *DomainPublisher
	if err := p.PublishDepartment(nil, DomainEventDepartmentDeleted, &models.Department{}); err != nil {
		t.Fatalf("nil publisher returned error: %v", err)
	}
}
