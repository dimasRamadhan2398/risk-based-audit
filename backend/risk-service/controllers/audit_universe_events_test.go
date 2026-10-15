package controllers

import (
	"net/http"
	"net/http/httptest"
	"risk-service/models"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestResyncEventsWithoutKafkaReturns503(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := NewAuditUniverseController(nil, nil) // db is never reached when publishing is disabled

	r := gin.New()
	r.POST("/audit-universe/events/resync", ctrl.ResyncEvents)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/audit-universe/events/resync", nil))

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503 when Kafka is disabled", w.Code)
	}
}

func TestEntityDataCarriesParentAndStandardIDs(t *testing.T) {
	parent := uuid.New()
	std := uuid.New()
	node := models.CorporateAuditUniverse{ID: uuid.New(), Name: "Treasury", ParentID: &parent, StandardAuditUniverseID: &std}

	d := entityData(node)
	if d.ID != node.ID.String() || d.Name != "Treasury" {
		t.Fatalf("unexpected entity data: %+v", d)
	}
	if d.ParentID == nil || *d.ParentID != parent.String() {
		t.Fatalf("parent_id not mapped: %+v", d.ParentID)
	}
	if d.StandardAuditUniverseID == nil || *d.StandardAuditUniverseID != std.String() {
		t.Fatalf("standard_audit_universe_id not mapped")
	}

	root := entityData(models.CorporateAuditUniverse{ID: uuid.New(), Name: "Root"})
	if root.ParentID != nil || root.StandardAuditUniverseID != nil {
		t.Fatalf("root custom node should have null parent and standard id")
	}
}

func TestYearScoreDataMapping(t *testing.T) {
	corp := uuid.New()
	d := yearScoreData(models.AuditUniverseYear{CorporateAuditUniverseID: corp, Year: 2026, RiskIndex: 83.6, RiskLevel: "High", AuditPriority: true})
	if d.CorporateAuditUniverseID != corp.String() || d.Year != 2026 || d.RiskLevel != "High" || !d.AuditPriority || d.RiskIndex != 83.6 {
		t.Fatalf("unexpected year score data: %+v", d)
	}
}
