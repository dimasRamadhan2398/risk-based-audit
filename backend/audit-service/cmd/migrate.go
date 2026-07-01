package main

import (
	"fmt"

	"audit-service/models"
	"audit-service/pkg/database"
	"audit-service/pkg/logger"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var migrateDryRun bool

func init() {
	// migrate up command - Create/update tables using GORM AutoMigrate
	upCmd := &cobra.Command{
		Use:   "up",
		Short: "Run database migrations (GORM AutoMigrate)",
		Long: `Runs GORM AutoMigrate to create or update audit-service database tables.
Creates the following tables:
- audit_charters
- audit_mandates
- audit_assignments
- audit_annuals
- activity_plans

Usage: audit migrate up`,
		RunE: runMigrateUp,
	}
	upCmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "Show what would be migrated")

	// migrate down command - Drop tables
	downCmd := &cobra.Command{
		Use:   "down",
		Short: "Rollback migrations (drop tables)",
		Long: `Drops all audit-service tables.
WARNING: This will delete all data!

Usage: audit migrate down`,
		RunE: runMigrateDown,
	}

	// migrate status command
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		Long:  `Shows which audit tables exist in the database.`,
		RunE:  runMigrateStatus,
	}

	// Parent migrate command
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration commands (GORM)",
		Long: `Audit Service database migration commands using GORM AutoMigrate.

Examples:
  audit migrate up           # Run all migrations
  audit migrate up --dry-run # Show what would be migrated
  audit migrate down         # Drop all tables
  audit migrate status       # Show current status`,
	}
	migrateCmd.AddCommand(upCmd, downCmd, statusCmd)

	rootCmd.AddCommand(migrateCmd)
}

func runMigrateUp(cmd *cobra.Command, args []string) error {
	initLogger()
	defer logger.Sync()

	logger.Info("Starting database migration (UP)")

	// Connect to database
	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", logger.LogField("error", err))
		return err
	}

	// Enable UUID extension
	if err := enableUUIDExtension(db); err != nil {
		logger.Warn("UUID extension may already exist or failed to create", logger.LogField("error", err))
	}

	// Define audit models to migrate
	auditModels := []interface{}{
		&models.AuditCharter{},
		&models.AuditMandate{},
		&models.AuditAssignment{},
		&models.AuditAnnual{},
		&models.ActivityPlan{},
		&models.AuditActivity{},
		&models.StrategicPlan{},
		&models.AssignmentLetter{},
		&models.AuditExecution{},
		&models.FieldworkInterview{},
		&models.FieldworkObservation{},
		&models.FieldworkDocument{},
		&models.FieldworkSample{},
		&models.FieldworkTestControl{},
		&models.WorkingPaperHeader{},
		&models.WorkingPaperRisk{},
		&models.WorkingPaperSample{},
		&models.WorkingPaperCause{},
		&models.WorkingPaperPlan{},
		&models.WorkingPaper{},
		&models.AuditResultReport{},
		&models.ActionTakenReport{},
		&models.ImportedWorkingPaper{},
		&models.KPIAchievement{},
		&models.WorkPlanRealization{},
		&models.ExecutiveSummary{},
	}

	if migrateDryRun {
		logger.Info("DRY RUN - Tables that would be created/updated:")
		for _, m := range auditModels {
			logger.Info(fmt.Sprintf("  ✓ %T", m))
		}
		return nil
	}

	// Run migrations
	logger.Info("Running GORM AutoMigrate...")

	for _, m := range auditModels {
		logger.Info(fmt.Sprintf("  Migrating model: %T", m))
		if err := db.AutoMigrate(m); err != nil {
			logger.Fatal(fmt.Sprintf("Failed to migrate model %T", m), logger.LogField("error", err))
			return err
		}
	}

	logger.Info("Database migration completed successfully!")
	return nil
}

func runMigrateDown(cmd *cobra.Command, args []string) error {
	initLogger()
	defer logger.Sync()

	logger.Warn("Starting database migration (DOWN)")
	logger.Warn("WARNING: This will drop ALL audit tables and delete ALL data!")

	// Connect to database
	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", logger.LogField("error", err))
		return err
	}

	// Drop tables in reverse order of dependencies
	tablesToDrop := []string{
		"imported_working_papers",
		"action_taken_reports",
		"work_plan_realizations",
		"kpi_achievements",
		"audit_result_reports",
		"working_paper_plans",
		"working_paper_causes",
		"working_paper_samples",
		"working_paper_risks",
		"working_paper_headers",
		"working_papers",
		"fieldwork_test_controls",
		"fieldwork_samples",
		"fieldwork_documents",
		"fieldwork_observations",
		"fieldwork_interviews",
		"audit_executions",
		"assignment_letters",
		"strategic_plans",
		"activity_plans",
		"audit_activities",
		"audit_annuals",
		"audit_assignments",
		"audit_mandates",
		"audit_charters",
	}

	logger.Warn("Dropping tables...")
	for _, table := range tablesToDrop {
		if db.Migrator().HasTable(table) {
			logger.Info(fmt.Sprintf("  Dropping: %s", table))
			if err := db.Migrator().DropTable(table); err != nil {
				logger.Error(fmt.Sprintf("Failed to drop %s", table), logger.LogField("error", err))
			}
		} else {
			logger.Info(fmt.Sprintf("Skipping (not exists): %s", table))
		}
	}

	logger.Warn("Database migration (DOWN) completed!")
	return nil
}

func runMigrateStatus(cmd *cobra.Command, args []string) error {
	initLogger()
	defer logger.Sync()

	logger.Info("Checking database migration status...")

	// Connect to database
	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", logger.LogField("error", err))
		return err
	}

	// Expected audit tables
	expectedTables := []string{
		"imported_working_papers",
		"audit_charters",
		"audit_mandates",
		"audit_assignments",
		"audit_annuals",
		"activity_plans",
		"audit_activities",
		"strategic_plans",
		"assignment_letters",
		"audit_executions",
		"fieldwork_interviews",
		"fieldwork_observations",
		"fieldwork_documents",
		"fieldwork_samples",
		"fieldwork_test_controls",
		"working_papers",
		"working_paper_headers",
		"working_paper_risks",
		"working_paper_samples",
		"working_paper_causes",
		"working_paper_plans",
		"audit_result_reports",
		"action_taken_reports",
		"kpi_achievements",
		"work_plan_realizations",
	}

	logger.Info(fmt.Sprintf("\nDatabase: %s", cfg.Database.Name))
	logger.Info("-------------------------------------------")
	logger.Info("Audit Tables                  | Status")
	logger.Info("-------------------------------------------")

	for _, table := range expectedTables {
		exists := db.Migrator().HasTable(table)
		status := "✓ EXISTS"
		if !exists {
			status = "✗ MISSING"
		}
		logger.Info(fmt.Sprintf("%-30s| %s", table, status))
	}
	logger.Info("-------------------------------------------")

	return nil
}

// enableUUIDExtension enables the PostgreSQL UUID extension
func enableUUIDExtension(db *gorm.DB) error {
	result := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	if result.Error != nil {
		// Try alternative
		result = db.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\"")
	}
	return result.Error
}
