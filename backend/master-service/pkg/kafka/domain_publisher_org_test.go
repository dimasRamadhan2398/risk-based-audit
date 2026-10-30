package kafka

import (
	"encoding/json"
	"testing"
	"time"

	"master-service/models"

	"github.com/google/uuid"
)

func dataKeys(t *testing.T, env DomainEnvelope) map[string]json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Data map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	return got.Data
}

// The legacy payload must keep exactly its six keys, even when the department
// carries response-only org fields.
func TestLegacyDepartmentPayloadUnchanged(t *testing.T) {
	p := &DomainPublisher{serviceName: "m", now: time.Now}
	parent := uuid.New()
	dept := &models.Department{ID: uuid.New(), CompanyID: uuid.New(), ParentDepartmentID: &parent, LevelName: "Divisi",
		Path: []models.DepartmentPathSeg{{ID: parent, Name: "X"}}}
	env, _ := p.departmentEnvelope(DomainEventDepartmentUpdated, dept)
	keys := dataKeys(t, env)
	want := []string{"id", "department_code", "department_name", "is_active", "company_id", "business_unit_id"}
	if len(keys) != len(want) {
		t.Fatalf("legacy keys changed: %v", keys)
	}
	for _, k := range want {
		if _, ok := keys[k]; !ok {
			t.Fatalf("missing %s", k)
		}
	}
}

func TestFlexibleDepartmentPayloadIsAdditive(t *testing.T) {
	p := &DomainPublisher{serviceName: "m", now: time.Now}
	key := models.OrgLevelKeyDivision
	dept := &models.Department{ID: uuid.New(), CompanyID: uuid.New(), Level: 20, LevelName: "Divisi", NameKey: &key}
	env, _ := p.departmentEnvelopeOrg(DomainEventDepartmentUpdated, dept)
	keys := dataKeys(t, env)
	for _, k := range []string{"id", "department_code", "department_name", "is_active", "company_id", "business_unit_id",
		"parent_department_id", "level", "level_name", "name_key", "path"} {
		if _, ok := keys[k]; !ok {
			t.Fatalf("missing %s in %v", k, keys)
		}
	}
	if string(keys["parent_department_id"]) != "null" || string(keys["path"]) != "[]" || string(keys["level"]) != "20" {
		t.Fatalf("unexpected values: %s %s %s", keys["parent_department_id"], keys["path"], keys["level"])
	}
}
