package main

import (
	"fmt"

	"auth-service/models"
	"auth-service/pkg/database"
	"auth-service/pkg/logger"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var (
	migrateDryRun bool
)

func init() {
	// migrate up command - Create/update tables using GORM AutoMigrate
	upCmd := &cobra.Command{
		Use:   "up",
		Short: "Run database migrations (GORM AutoMigrate)",
		Long: `Runs GORM AutoMigrate to create or update auth-service database tables.
Creates the following tables:
- permissions, roles, role_permissions
- users, user_roles
- refresh_tokens, mfa_setups, trusted_devices
- confidentiality_agreements, sys_alert_logs

Usage: auth migrate up`,
		RunE:  runMigrateUp,
	}
	upCmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "Show what would be migrated")

	// migrate down command - Drop tables
	downCmd := &cobra.Command{
		Use:   "down",
		Short: "Rollback migrations (drop tables)",
		Long: `Drops all auth-service tables.
WARNING: This will delete all data!

Usage: auth migrate down`,
		RunE:  runMigrateDown,
	}
	downCmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "Show what would be dropped")

	// migrate status command
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		Long:  `Shows which auth tables exist in the database.`,
		RunE:  runMigrateStatus,
	}

	// Parent migrate command
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration commands (GORM)",
		Long: `Auth Service database migration commands using GORM AutoMigrate.

Examples:
  auth migrate up           # Run all migrations
  auth migrate up --dry-run # Show what would be migrated
  auth migrate down         # Drop all tables
  auth migrate status       # Show current status`,
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

	// Define auth models to migrate (order matters for foreign keys)
	// First: Independent tables (no FK dependencies)
	authModels := []interface{}{
		&models.Permission{},
		&models.Role{},
	}

	if migrateDryRun {
		logger.Info("DRY RUN - Tables that would be created/updated:")
		for _, m := range authModels {
			logger.Info(fmt.Sprintf("  ✓ %T", m))
		}
		logger.Info("  + users (with user_roles many2many)")
		logger.Info("  + refresh_tokens")
		logger.Info("  + mfa_setups")
		logger.Info("  + trusted_devices")
		logger.Info("  + confidentiality_agreements")
		logger.Info("  + sys_alert_logs")
		return nil
	}

	// Run migrations in order
	logger.Info("Running GORM AutoMigrate...")

	// Step 1: Migrate Permission and Role (no dependencies)
	logger.Info("  Migrating: permissions")
	if err := db.AutoMigrate(&models.Permission{}); err != nil {
		logger.Fatal("Failed to migrate permissions", logger.LogField("error", err))
		return err
	}

	logger.Info("  Migrating: roles")
	if err := db.AutoMigrate(&models.Role{}); err != nil {
		logger.Fatal("Failed to migrate roles", logger.LogField("error", err))
		return err
	}

	// Step 2: Migrate User (depends on Role)
	logger.Info("  Migrating: users")
	if err := db.AutoMigrate(&models.User{}); err != nil {
		logger.Fatal("Failed to migrate users", logger.LogField("error", err))
		return err
	}

	// Step 3: Migrate dependent tables
	logger.Info("  Migrating: refresh_tokens")
	if err := db.AutoMigrate(&models.RefreshToken{}); err != nil {
		logger.Fatal("Failed to migrate refresh_tokens", logger.LogField("error", err))
		return err
	}

	logger.Info("  Migrating: mfa_setups")
	if err := db.AutoMigrate(&models.MFASetup{}); err != nil {
		logger.Fatal("Failed to migrate mfa_setups", logger.LogField("error", err))
		return err
	}

	logger.Info("  Migrating: trusted_devices")
	if err := db.AutoMigrate(&models.TrustedDevice{}); err != nil {
		logger.Fatal("Failed to migrate trusted_devices", logger.LogField("error", err))
		return err
	}

	logger.Info("  Migrating: confidentiality_agreements")
	if err := db.AutoMigrate(&models.ConfidentialityAgreement{}); err != nil {
		logger.Fatal("Failed to migrate confidentiality_agreements", logger.LogField("error", err))
		return err
	}

	logger.Info("  Migrating: sys_alert_logs")
	if err := db.AutoMigrate(&models.SysAlertLog{}); err != nil {
		logger.Fatal("Failed to migrate sys_alert_logs", logger.LogField("error", err))
		return err
	}

	logger.Info("Database migration completed successfully!")
	return nil
}

func runMigrateDown(cmd *cobra.Command, args []string) error {
	initLogger()
	defer logger.Sync()

	logger.Warn("Starting database migration (DOWN)")
	logger.Warn("WARNING: This will drop ALL auth tables and delete ALL data!")

	// Connect to database
	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", logger.LogField("error", err))
		return err
	}

	if migrateDryRun {
		logger.Info("DRY RUN - Tables that would be dropped:")
		authTables := []string{
			"sys_alert_logs",
			"confidentiality_agreements",
			"trusted_devices",
			"mfa_setups",
			"refresh_tokens",
			"user_roles",         // many2many junction table
			"users",
			"role_permissions",    // many2many junction table
			"roles",
			"permissions",
		}
		for _, t := range authTables {
			logger.Info(fmt.Sprintf("  ✗ %s", t))
		}
		return nil
	}

	// Drop tables in reverse order of dependencies
	tablesToDrop := []string{
		"sys_alert_logs",
		"confidentiality_agreements",
		"trusted_devices",
		"mfa_setups",
		"refresh_tokens",
		"user_roles",
		"users",
		"role_permissions",
		"roles",
		"permissions",
	}

	logger.Warn("Dropping tables...")
	for _, table := range tablesToDrop {
		if db.Migrator().HasTable(table) {
			logger.Info(fmt.Sprintf("  Dropping: %s", table))
			if err := db.Migrator().DropTable(table); err != nil {
				logger.Error(fmt.Sprintf("Failed to drop %s", table), logger.LogField("error", err))
			}
		} else {
			logger.Info(fmt.Sprintf("  Skipping (not exists): %s", table))
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

	// Expected auth tables
	expectedTables := []string{
		"permissions",
		"roles",
		"role_permissions",
		"users",
		"user_roles",
		"refresh_tokens",
		"mfa_setups",
		"trusted_devices",
		"confidentiality_agreements",
		"sys_alert_logs",
	}

	logger.Info(fmt.Sprintf("\nDatabase: %s", cfg.Database.Name))
	logger.Info("-------------------------------------------")
	logger.Info("Auth Tables                  | Status")
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
