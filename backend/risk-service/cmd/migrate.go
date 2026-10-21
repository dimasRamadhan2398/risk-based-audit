package main

import (
	"fmt"
	"log"

	"risk-service/models"
	"risk-service/pkg/database"
	"risk-service/pkg/logger"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var migrateCmd = &cobra.Command{
	Use:   "up",
	Short: "Run auto migration",
	RunE:  runMigrate,
}

var freshCmd = &cobra.Command{
	Use:   "fresh",
	Short: "Drop and recreate all risk-service tables",
	RunE:  runFresh,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(freshCmd)
}

func runMigrate(cmd *cobra.Command, args []string) error {
	if err := initConfig(); err != nil {
		return err
	}
	if err := initLogger(); err != nil {
		return err
	}
	defer logger.Sync()

	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	return autoMigrate(db)
}

func runFresh(cmd *cobra.Command, args []string) error {
	if err := initConfig(); err != nil {
		return err
	}
	if err := initLogger(); err != nil {
		return err
	}
	defer logger.Sync()

	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := dropAllTables(db); err != nil {
		return err
	}
	return autoMigrate(db)
}

func autoMigrate(db *gorm.DB) error {
	if db.Migrator().HasTable(&models.RiskProfile{}) {
		db.Exec(`ALTER TABLE "risk_profile" ADD COLUMN IF NOT EXISTS "impact" bigint DEFAULT 0 NOT NULL`)
		db.Exec(`ALTER TABLE "risk_profile" ADD COLUMN IF NOT EXISTS "likelihood" bigint DEFAULT 0 NOT NULL`)
		db.Exec(`ALTER TABLE "risk_profile" ADD COLUMN IF NOT EXISTS "risk_score" bigint DEFAULT 0 NOT NULL`)
		db.Exec(`ALTER TABLE "risk_profile" ADD COLUMN IF NOT EXISTS "severity_weight" numeric(5,4) DEFAULT 0`)
	}

	if err := addAuditUniverseScoredAt(db); err != nil {
		return err
	}

	if err := db.AutoMigrate(
		&models.KafkaOutboxEvent{},
		&models.RiskLevel{},
		&models.RiskAppetite{},
		&models.RiskProfile{},
		&models.RiskRegister{},
		&models.RiskAssessment{},
		&models.RiskControl{},
		&models.RiskControlMatrix{},
		&models.SOPDocument{},
		&models.StrategicObjective{},
		&models.RiskSOPMap{},
		&models.RiskStrategyMap{},
		&models.RiskApprovalLog{},
		&models.ActionPlan{},
		&models.RiskMitigation{},
		&models.StandardRiskFactor{},
		&models.CorporateRiskFactor{},
		&models.StandardAuditUniverse{},
		&models.CorporateAuditUniverse{},
		&models.AuditUniverseYear{},
		&models.AuditUniverseRiskScore{},
	); err != nil {
		return err
	}

	// Migrations never delete data. The legacy dummy audit-universe cleanup is
	// an explicit operator command (cleanup-dummy-universe), not a boot step.
	return backfillAssessmentSnapshots(db)
}

// addAuditUniverseScoredAt adds audit_universe_years.scored_at and, in the same
// transaction, marks every row that existed before the column as scored
// (scored_at = updated_at, or created_at). Those rows were all created with an
// index and level (the old default of 60 / "Medium to High" or a real score),
// so they keep counting as scored.
//
// The backfill runs only in the run that adds the column. It must not run
// again: after this deploy a NULL scored_at means "never scored" and a later
// migrate would otherwise turn newly established rows into scored ones. The
// ALTER and the UPDATE share one transaction (Postgres DDL is transactional),
// so a failed backfill also leaves the column absent and the next run retries.
// AutoMigrate then sees the column already present.
func addAuditUniverseScoredAt(db *gorm.DB) error {
	m := db.Migrator()
	if !m.HasTable(&models.AuditUniverseYear{}) || m.HasColumn(&models.AuditUniverseYear{}, "scored_at") {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`ALTER TABLE "audit_universe_years" ADD COLUMN "scored_at" timestamptz`).Error; err != nil {
			return fmt.Errorf("add audit_universe_years.scored_at: %w", err)
		}
		res := tx.Exec(`UPDATE "audit_universe_years" SET "scored_at" = COALESCE("updated_at", "created_at") WHERE "scored_at" IS NULL`)
		if res.Error != nil {
			return fmt.Errorf("backfill audit_universe_years.scored_at: %w", res.Error)
		}
		log.Printf("added audit_universe_years.scored_at; backfilled %d existing rows as scored", res.RowsAffected)
		return nil
	})
}

// backfillAssessmentSnapshots gives assessment rows written before the
// per-year snapshot columns existed the risk as it reads now — the only
// version still on record. Rows that already have a snapshot are left alone,
// so re-running is a no-op.
func backfillAssessmentSnapshots(db *gorm.DB) error {
	res := db.Exec(`
		UPDATE risk_assessment a
		SET risk_event    = r.risk_event,
		    category      = p.category,
		    description   = p.description,
		    location_id   = p.location_id,
		    location_name = p.location_name,
		    snapshot_at   = NOW()
		FROM risk_register r
		JOIN risk_profile p ON p.id = r.profile_id
		WHERE a.risk_register_id = r.id AND a.snapshot_at IS NULL`)
	if res.Error != nil {
		return fmt.Errorf("backfill risk_assessment snapshots: %w", res.Error)
	}
	if res.RowsAffected > 0 {
		log.Printf("backfilled %d risk_assessment snapshots", res.RowsAffected)
	}
	return nil
}

func dropAllTables(db *gorm.DB) error {
	return db.Migrator().DropTable(
		&models.RiskControlMatrix{},
		&models.AuditUniverseRiskScore{},
		&models.AuditUniverseYear{},
		&models.CorporateAuditUniverse{},
		&models.StandardAuditUniverse{},
		&models.CorporateRiskFactor{},
		&models.StandardRiskFactor{},
		&models.RiskMitigation{},
		&models.ActionPlan{},
		&models.RiskApprovalLog{},
		&models.RiskStrategyMap{},
		&models.RiskSOPMap{},
		&models.StrategicObjective{},
		&models.SOPDocument{},
		&models.RiskControl{},
		&models.RiskAssessment{},
		&models.RiskRegister{},
		&models.RiskProfile{},
		&models.RiskAppetite{},
		&models.RiskLevel{},
	)
}
