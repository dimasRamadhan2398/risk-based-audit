package controllers

import (
	"testing"

	"risk-service/models"
	"risk-service/pkg/sqlitetest"

	"gorm.io/gorm"
)

// openTestDB returns a sqlite database with the Audit Universe, risk factor
// and outbox tables (see pkg/sqlitetest for what the tables do not enforce).
func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	return sqlitetest.Open(t,
		&models.StandardRiskFactor{},
		&models.CorporateRiskFactor{},
		&models.StandardAuditUniverse{},
		&models.CorporateAuditUniverse{},
		&models.AuditUniverseYear{},
		&models.AuditUniverseRiskScore{},
		&models.KafkaOutboxEvent{},
	)
}
