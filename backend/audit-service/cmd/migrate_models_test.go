package main

import (
	"reflect"
	"testing"

	"audit-service/models"
)

func TestAllAuditModelsIncludesCriticalTablesWithoutDuplicates(t *testing.T) {
	list := allAuditModels()
	seen := make(map[reflect.Type]bool, len(list))
	for _, m := range list {
		typ := reflect.TypeOf(m)
		if seen[typ] {
			t.Fatalf("duplicate model in allAuditModels(): %v", typ)
		}
		seen[typ] = true
	}

	required := []interface{}{
		&models.AssignmentLetter{},
		&models.StrategicPlan{},
		&models.AuditAnnual{},
		&models.ActivityPlan{},
		&models.AuditActivity{},
		&models.AuditExecution{},
		&models.AuditResultReport{},
		&models.ExecutiveSummary{},
		&models.ActionTakenReport{},
		&models.ActionTakenReportEvidence{},
		&models.WorkingPaper{},
		&models.FieldworkSample{},
		&models.FieldworkTestControl{},
	}
	for _, req := range required {
		typ := reflect.TypeOf(req)
		if !seen[typ] {
			t.Errorf("allAuditModels() is missing %v — startup AutoMigrate in 'audit serve' would skip its schema updates", typ)
		}
	}
}
