package kafka

import (
	"context"
	"encoding/json"
	"testing"
)

func TestNilPublisherIsNoop(t *testing.T) {
	var p *Publisher
	if err := p.PublishEntity(context.Background(), EventEntityCreated, EntityData{ID: "x"}); err != nil {
		t.Fatalf("nil publisher returned error: %v", err)
	}
	if err := p.PublishYearScore(context.Background(), YearScoreData{}); err != nil {
		t.Fatalf("nil publisher returned error: %v", err)
	}
	if err := p.Close(); err != nil {
		t.Fatalf("nil publisher Close returned error: %v", err)
	}
}

func TestEnvelopeShape(t *testing.T) {
	parent := "parent-id"
	env := Envelope{
		Version:   EnvelopeVersion,
		EventID:   "evt-1",
		EventType: EventEntityUpdated,
		Producer:  "risk-service",
		Data: EntityData{
			ID:       "node-1",
			Name:     "Treasury",
			ParentID: &parent,
		},
	}
	raw, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Version   int    `json:"version"`
		EventID   string `json:"event_id"`
		EventType string `json:"event_type"`
		Producer  string `json:"producer"`
		Data      struct {
			ID                      string  `json:"id"`
			Name                    string  `json:"name"`
			ParentID                *string `json:"parent_id"`
			StandardAuditUniverseID *string `json:"standard_audit_universe_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.Version != 1 || got.EventType != "audit_universe.entity.updated" || got.Producer != "risk-service" {
		t.Fatalf("unexpected envelope: %+v", got)
	}
	if got.Data.ID != "node-1" || got.Data.Name != "Treasury" || got.Data.ParentID == nil || *got.Data.ParentID != "parent-id" {
		t.Fatalf("unexpected data: %+v", got.Data)
	}
	if got.Data.StandardAuditUniverseID != nil {
		t.Fatalf("expected null standard_audit_universe_id for a custom entity")
	}
}
