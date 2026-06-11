package main

import (
	"fmt"
	"os"
	"time"

	"master-service/models"
	"master-service/pkg/config"
	"master-service/pkg/database"
	"master-service/pkg/database/seeders"
	"master-service/pkg/logger"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var (
	configPath string
	cfg        *config.Config
)

// departmentMigration breaks Department<->Employee FK cycle during initial table creation.
type departmentMigration struct {
	ID                    uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DepartmentCode        string     `gorm:"type:varchar(50);uniqueIndex;not null"`
	DepartmentName        string     `gorm:"type:varchar(100);uniqueIndex;not null"`
	DepartmentDescription string     `gorm:"type:text"`
	PicID                 *uuid.UUID `gorm:"type:uuid;index"`
	Level                 int        `gorm:"type:int;not null"`
	IsActive              bool       `gorm:"default:true"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
	DeletedAt             gorm.DeletedAt `gorm:"index"`
	CompanyID             uuid.UUID      `gorm:"type:uuid;not null;index"`
	BusinessUnitID        uuid.UUID      `gorm:"type:uuid;not null;index"`
}

func (departmentMigration) TableName() string { return "departments" }

var rootCmd = &cobra.Command{
	Use:   "master",
	Short: "Master-service management commands",
	Long:  `Commands for running the master-service HTTP server and database tasks`,
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `Run auto-migrate to create or update database tables`,
	RunE:  runMigrate,
}

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed database with initial data",
	Long:  `Seed the database with initial master data`,
	RunE:  runSeed,
}

var freshCmd = &cobra.Command{
	Use:     "fresh",
	Aliases: []string{"root"},
	Short:   "Drop all tables and re-migrate",
	Long:    `Drop all tables and run migrations fresh`,
	RunE:    runFresh,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(seedCmd)
	rootCmd.AddCommand(freshCmd)

	rootCmd.PersistentFlags().StringVar(&configPath, "config", "pkg/config/config.yaml", "Path to config file")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func initConfig() error {
	var err error
	cfg, err = config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	return nil
}

func initLogger() error {
	if cfg == nil {
		return fmt.Errorf("config not loaded")
	}
	if err := logger.Init(&cfg.Log); err != nil {
		return fmt.Errorf("failed to init logger: %w", err)
	}
	return nil
}

func runMigrate(cmd *cobra.Command, args []string) error {
	if err := initConfig(); err != nil {
		return err
	}
	if err := initLogger(); err != nil {
		return err
	}
	defer logger.Sync()

	// Initialize database connection
	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	fmt.Println("Running database migrations...")

	// Auto-migrate all models
	if err := autoMigrate(db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	fmt.Println("Migrations completed successfully!")
	return nil
}

func runSeed(cmd *cobra.Command, args []string) error {
	if err := initConfig(); err != nil {
		return err
	}
	if err := initLogger(); err != nil {
		return err
	}
	defer logger.Sync()

	// Initialize database connection
	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// First run migrations
	fmt.Println("Running migrations before seeding...")
	if err := autoMigrate(db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Run seeders
	fmt.Println("\nSeeding database...")
	seeder := seeders.NewSeeder(db)
	if err := seeder.RunAll(); err != nil {
		return fmt.Errorf("failed to run seeders: %w", err)
	}

	fmt.Println("\nSeeding completed successfully!")
	return nil
}

func runFresh(cmd *cobra.Command, args []string) error {
	if err := initConfig(); err != nil {
		return err
	}
	if err := initLogger(); err != nil {
		return err
	}
	defer logger.Sync()

	// Initialize database connection
	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	fmt.Println("Dropping all tables...")
	if err := dropAllTables(db); err != nil {
		return fmt.Errorf("failed to drop tables: %w", err)
	}

	// Run migrations
	fmt.Println("\nRunning migrations...")
	if err := autoMigrate(db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Run seeders
	fmt.Println("\nSeeding database...")
	seeder := seeders.NewSeeder(db)
	if err := seeder.RunAll(); err != nil {
		return fmt.Errorf("failed to run seeders: %w", err)
	}

	fmt.Println("\nFresh migration and seeding completed successfully!")
	return nil
}

func autoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&models.Location{},
		&models.Company{},
		&models.JobRole{},
		&models.Likelihood{},
		&models.Impact{},
		&models.RiskLevel{},
		&models.RiskCategory{},
		&models.RiskMatrixCell{},
		&models.BusinessUnit{},
		&models.AuditPeriod{},
	); err != nil {
		return err
	}

	if err := db.AutoMigrate(&departmentMigration{}); err != nil {
		return err
	}

	return db.AutoMigrate(
		&models.Employee{},
		&models.Department{},
		&models.RiskRegister{},
		&models.AnnualAuditPlan{},
		&models.AuditScope{},
		&models.Control{},
		&models.ControlAssessment{},
		&models.AuditFinding{},
		&models.AuditIssue{},
		&models.AuditRecommendation{},
		&models.MitigationAction{},
		&models.RiskCause{},
		&models.RiskEffect{},
		&models.RiskIndicator{},
		&models.RiskIndicatorLog{},
		&models.AuditUniverse{},
		&models.AuditWorkpaper{},
	)
}

func dropAllTables(db *gorm.DB) error {
	return db.Migrator().DropTable(
		&models.AuditWorkpaper{},
		&models.RiskIndicatorLog{},
		&models.RiskIndicator{},
		&models.RiskEffect{},
		&models.RiskCause{},
		&models.MitigationAction{},
		&models.AuditRecommendation{},
		&models.AuditIssue{},
		&models.AuditFinding{},
		&models.ControlAssessment{},
		&models.Control{},
		&models.AuditScope{},
		&models.AnnualAuditPlan{},
		&models.AuditUniverse{},
		&models.AuditPeriod{},
		&models.RiskRegister{},
		&models.Employee{},
		&models.Department{},
		&models.BusinessUnit{},
		&models.RiskMatrixCell{},
		&models.RiskCategory{},
		&models.RiskLevel{},
		&models.Impact{},
		&models.Likelihood{},
		&models.JobRole{},
		&models.Company{},
		&models.Location{},
	)
}
