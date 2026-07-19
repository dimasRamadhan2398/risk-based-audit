package main

import (
	"fmt"

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

	return db.AutoMigrate(
		&models.RiskLevel{},
		&models.RiskAppetite{},
		&models.RiskProfile{},
		&models.RiskRegister{},
		&models.RiskAssessment{},
		&models.RiskControl{},
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
	)
}

func dropAllTables(db *gorm.DB) error {
	return db.Migrator().DropTable(
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
