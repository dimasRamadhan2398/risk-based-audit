package main

import (
	"fmt"
	"os"

	"risk-service/models"
	"risk-service/pkg/config"
	"risk-service/pkg/database"
	"risk-service/pkg/logger"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var (
	configPath string
	cfg        *config.Config
)

// var rootCmd = &cobra.Command{
// 	Use:   "migrate",
// 	Short: "Database migration commands for risk-service",
// }

var rootCmd = &cobra.Command{
	Use:   "risk",
	Short: "Database migration commands for risk-service",
}

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
	if err := logger.Init(nil); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
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
	)
}

func dropAllTables(db *gorm.DB) error {
	return db.Migrator().DropTable(
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
