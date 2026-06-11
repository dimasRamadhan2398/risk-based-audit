package main

import (
	"audit-service/models"
	"audit-service/pkg/database"
	"audit-service/pkg/logger"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var seedDrop bool

func init() {
	seedCmd := &cobra.Command{
		Use:   "seed",
		Short: "Seed the database with initial data",
		Long: `Seeds the database with initial audit data including:
- Sample audit mandates
- Sample audit charters`,
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

	// Seed data
	logger.Info("Seeding audit mandates...")
	if err := seedAuditMandates(db); err != nil {
		logger.Fatal("Failed to seed audit mandates", logger.LogField("error", err))
		return err
	}

	logger.Info("Seeding audit charters...")
	if err := seedAuditCharters(db); err != nil {
		logger.Fatal("Failed to seed audit charters", logger.LogField("error", err))
		return err
	}

	logger.Info("Database seeding completed successfully")

	return nil
}

func dropTables(db *gorm.DB) error {
	return db.Migrator().DropTable(
		"activity_plans",
		"audit_annuals",
		"audit_charters",
		"audit_mandates",
		"audit_assignments",
	)
}

func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.AuditCharter{},
		&models.AuditMandate{},
		&models.AuditAssignment{},
		&models.AuditAnnual{},
		&models.ActivityPlan{},
	)
}

func seedAuditMandates(db *gorm.DB) error {
	// Create sample audit mandate
	mandate := &models.AuditMandate{
		Title:           "Internal Audit Mandate 2024",
		ReferenceNumber: "MANDATE-2024-001",
		MandateSource:   "Board of Directors",
		LegalBasis:      "Regulation No. 123/2024",
		IsActive:        true,
	}

	result := db.Create(mandate)
	if result.Error != nil {
		return result.Error
	}

	logger.Info("Sample audit mandate created")
	return nil
}

func seedAuditCharters(db *gorm.DB) error {
	// Create sample audit charter
	charter := &models.AuditCharter{
		Filename: "Audit_Charter_v1.0.pdf",
		Version:  "1.0",
		Title:    "Internal Audit Charter",
		Content:  "This charter establishes the authority and responsibilities of the internal audit function.",
		IsActive: true,
	}

	result := db.Create(charter)
	if result.Error != nil {
		return result.Error
	}

	logger.Info("Sample audit charter created")
	return nil
}
