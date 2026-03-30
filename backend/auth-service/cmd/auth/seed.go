package main

import (
	"auth-service/models"
	"auth-service/pkg/database"
	"auth-service/pkg/database/seeders"
	"auth-service/pkg/logger"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var (
	seedDrop bool
)

func init() {
	seedCmd := &cobra.Command{
		Use:   "seed",
		Short: "Seed the database with initial data",
		Long: `Seeds the database with initial data including:
- Permissions
- Roles
- Default users (admin, auditor, risk_manager)

Default password for all seeded users: password123`,
		RunE: runSeed,
	}

	seedCmd.Flags().BoolVar(&seedDrop, "drop", false, "Drop all tables before seeding (WARNING: destructive)")

	rootCmd.AddCommand(seedCmd)
}

func runSeed(cmd *cobra.Command, args []string) error {
	// Initialize logger
	initLogger()
	defer logger.Sync()

	logger.Info("Starting database seeder")

	// Connect to database
	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", logger.LogField("error", err))
		return err
	}

	// Run migrations if drop flag is set
	if seedDrop {
		logger.Warn("Dropping all tables...")
		if err := dropTables(db); err != nil {
			logger.Fatal("Failed to drop tables", logger.LogField("error", err))
			return err
		}
	}

	// Run migrations
	logger.Info("Running database migrations...")
	if err := runMigrations(db); err != nil {
		logger.Fatal("Failed to run migrations", logger.LogField("error", err))
		return err
	}

	// Create seeder
	seeder := seeders.NewSeeder(db)

	// Run all seeders
	logger.Info("Seeding database...")

	if err := seeder.SeedPermissions(); err != nil {
		logger.Fatal("Failed to seed permissions", logger.LogField("error", err))
		return err
	}
	logger.Info("Permissions seeded successfully")

	if err := seeder.SeedRoles(); err != nil {
		logger.Fatal("Failed to seed roles", logger.LogField("error", err))
		return err
	}
	logger.Info("Roles seeded successfully")

	if err := seeder.SeedUsers(); err != nil {
		logger.Fatal("Failed to seed users", logger.LogField("error", err))
		return err
	}
	logger.Info("Users seeded successfully")

	logger.Info("Database seeding completed successfully")
	logger.Info("Default credentials:")
	logger.Info("  Username: admin    | Password: password123 | Role: ADMIN")
	logger.Info("  Username: auditor  | Password: password123 | Role: AUDITOR")
	logger.Info("  Username: risk_manager | Password: password123 | Role: RISK_MANAGER")

	return nil
}

func dropTables(db *gorm.DB) error {
	return db.Migrator().DropTable(
		"users",
		"roles",
		"permissions",
		"user_roles",
		"role_permissions",
		"refresh_tokens",
		"mfa_setups",
		"trusted_devices",
		"confidentiality_agreements",
		"sys_alert_logs",
	)
}

func runMigrations(db *gorm.DB) error {
	// Auto migrate models - this will create/update tables
	return db.AutoMigrate(
		&models.Permission{},
		&models.Role{},
		&models.User{},
		&models.RefreshToken{},
		&models.MFASetup{},
		&models.TrustedDevice{},
		&models.ConfidentialityAgreement{},
		&models.SysAlertLog{},
	)
}
