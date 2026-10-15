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
	BusinessUnitID        *uuid.UUID     `gorm:"type:uuid;index"`
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

	if err := db.AutoMigrate(
		&models.Employee{},
		&models.Department{},
		&models.RiskRegister{},
		&models.AnnualAuditPlan{},
		&models.AnnualAuditPlanActivity{},
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
		&models.QAReport{},
		&models.ConsultingService{},
		&models.VisionMissionGoals{},
		&models.VmgGoal{},
		&models.DataSourceConnection{},
		&models.DataSourceActivityLog{},
	); err != nil {
		return err
	}

	return normalizeAnnualPlanStatuses(db)
}

// normalizeAnnualPlanStatuses maps legacy annual_audit_plans.status values to the
// canonical vocabulary (models/annual_audit_plan.go). Idempotent; covers
// soft-deleted rows. Unknown values are left unchanged and logged.
func normalizeAnnualPlanStatuses(db *gorm.DB) error {
	groups := map[models.AuditPlanStatus][]string{
		models.AuditPlanStatusDraft:           {"", "Not Available", "DRAFT"},
		models.AuditPlanStatusPendingReview:   {"Work In Progress", "SUBMITTED"},
		models.AuditPlanStatusPendingApproval: {"PENDING_APPROVAL"},
		models.AuditPlanStatusApproved:        {"Done", "APPROVED"},
		models.AuditPlanStatusRevised:         {"REJECTED", "REVISED"},
	}
	var rewritten int64
	for target, legacy := range groups {
		// The condition strings are constants; only the values are bound.
		cond := "status IN ?"
		if target == models.AuditPlanStatusDraft {
			cond = "status IN ? OR status IS NULL"
		}
		res := db.Unscoped().Model(&models.AnnualAuditPlan{}).
			Where("("+cond+")", legacy).
			Update("status", target)
		if res.Error != nil {
			return fmt.Errorf("normalize annual plan statuses: %w", res.Error)
		}
		rewritten += res.RowsAffected
	}
	if rewritten > 0 {
		fmt.Printf("Annual plan statuses normalized: %d row(s)\n", rewritten)
	}

	canonical := []models.AuditPlanStatus{
		models.AuditPlanStatusDraft, models.AuditPlanStatusPendingReview,
		models.AuditPlanStatusPendingApproval, models.AuditPlanStatusApproved,
		models.AuditPlanStatusRevised,
	}
	var unknown []string
	if err := db.Unscoped().Model(&models.AnnualAuditPlan{}).
		Where("status NOT IN ?", canonical).
		Distinct().Pluck("status", &unknown).Error; err != nil {
		return fmt.Errorf("check annual plan statuses: %w", err)
	}
	if len(unknown) > 0 {
		fmt.Printf("WARNING: annual plans with status outside the vocabulary: %v\n", unknown)
	}
	return nil
}

func dropAllTables(db *gorm.DB) error {
	return db.Migrator().DropTable(
		&models.DataSourceActivityLog{},
		&models.DataSourceConnection{},
		&models.VmgGoal{},
		&models.VisionMissionGoals{},
		&models.QAReport{},
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
