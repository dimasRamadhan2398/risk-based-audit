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

func init() {
	rootCmd.AddCommand(seedCmd)
}

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed initial data for risk-service",
	RunE:  runSeed,
}

func runSeed(cmd *cobra.Command, args []string) error {
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

	return seedRiskLevels(db)
}

func seedRiskLevels(db *gorm.DB) error {
	riskLevels := []models.RiskLevel{
		{
			RiskCode:        "LOW",
			RiskName:        "Low Risk",
			RiskDescription: "Risiko rendah, operasional berjalan sesuai tupoksi. Cukup dipantau berkala.",
			MinScore:        1,
			MaxScore:        5,
			Color:           "#E8FAEF",
			IsActive:        true,
		},
		{
			RiskCode:        "LOW_MODERATE",
			RiskName:        "Low to Moderate Risk",
			RiskDescription: "Risiko rendah-sedang, perlu perhatian untuk menjaga kinerja.",
			MinScore:        6,
			MaxScore:        11,
			Color:           "#FFF9E6",
			IsActive:        true,
		},
		{
			RiskCode:        "MODERATE",
			RiskName:        "Moderate Risk",
			RiskDescription: "Risiko sedang, memerlukan tindakan mitigasi dalam waktu dekat.",
			MinScore:        12,
			MaxScore:        15,
			Color:           "#FFF3CD",
			IsActive:        true,
		},
		{
			RiskCode:        "MODERATE_HIGH",
			RiskName:        "Moderate to High Risk",
			RiskDescription: "Risiko sedang-tinggi, memerlukan perhatian serius dan rencana mitigasi.",
			MinScore:        16,
			MaxScore:        19,
			Color:           "#FFE5CC",
			IsActive:        true,
		},
		{
			RiskCode:        "HIGH",
			RiskName:        "High Risk",
			RiskDescription: "Risiko tinggi, memerlukan tindakan segera dan eskalasi ke manajemen.",
			MinScore:        20,
			MaxScore:        25,
			Color:           "#FFE5E5",
			IsActive:        true,
		},
	}

	for _, level := range riskLevels {
		// Check if already exists
		var existing models.RiskLevel
		result := db.Where("risk_code = ?", level.RiskCode).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&level).Error; err != nil {
				return fmt.Errorf("failed to seed risk level %s: %w", level.RiskCode, err)
			}
			fmt.Printf("✅ Seeded: %s (score: %d-%d)\n", level.RiskName, level.MinScore, level.MaxScore)
		} else if result.Error != nil {
			return fmt.Errorf("failed to check risk level %s: %w", level.RiskCode, result.Error)
		} else {
			fmt.Printf("⏭️  Already exists: %s\n", level.RiskCode)
		}
	}

	fmt.Println("\n✅ Risk level seeding completed!")
	return nil
}
