package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"risk-service/models"
	"risk-service/pkg/database"
	"risk-service/pkg/logger"

	"github.com/google/uuid"
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

	if err := seedInitialRisks(db); err != nil {
		return err
	}

	if err := seedRiskLevels(db); err != nil {
		return err
	}

	if err := seedRiskFactorsAndUniverse(db); err != nil {
		return err
	}

	return nil
}

func seedInitialRisks(db *gorm.DB) error {
	var count int64
	db.Model(&models.RiskRegister{}).Count(&count)
	if count > 0 {
		return nil
	}

	log.Println("Seeding initial risk data...")

	type SeedRisk struct {
		Name         string `json:"name"`
		Impact       int    `json:"impact"`
		Likelihood   int    `json:"likelihood"`
		Severity     int    `json:"severity"`
		Category     string `json:"category"`
		Branch       string `json:"branch"`
		Description  string `json:"description"`
		ImpactQ1     int
		ImpactQ2     int
		ImpactQ3     int
		ImpactQ4     int
		LikelihoodQ1 int
		LikelihoodQ2 int
		LikelihoodQ3 int
		LikelihoodQ4 int
	}

	rawSeeds := []SeedRisk{
		{
			Name: "Target pendapatan dan laba tidak tercapai", Impact: 5, Likelihood: 4, Severity: 98, Category: "Financial", Branch: "Head Office", Description: "Terget pendapatan tidak tercapai karena kinerja tim marketing yang kurang maksimal and strategi marketing yang tidak efektif.",
			ImpactQ1: 5, ImpactQ2: 4, ImpactQ3: 3, ImpactQ4: 2,
			LikelihoodQ1: 4, LikelihoodQ2: 3, LikelihoodQ3: 2, LikelihoodQ4: 2,
		},
		{
			Name: "Target efisiensi biaya operasional dan umum tidak tercapai", Impact: 5, Likelihood: 4, Severity: 95, Category: "Financial", Branch: "Head Office", Description: "Target efisiensi biaya operasional dan umum tidak tercapai karena kinerja tim keuangan yang kurang maksimal dan strategi keuangan yang tidak efektif.",
			ImpactQ1: 5, ImpactQ2: 4, ImpactQ3: 3, ImpactQ4: 2,
			LikelihoodQ1: 4, LikelihoodQ2: 3, LikelihoodQ3: 3, LikelihoodQ4: 2,
		},
		{
			Name: "Ancaman terhadap Cyber Security dan perlindungan data pribadi", Impact: 5, Likelihood: 4, Severity: 88, Category: "Technology", Branch: "Head Office", Description: "Ancaman terhadap cyber security dan kebocoran data pelanggan/karyawan.",
			ImpactQ1: 5, ImpactQ2: 4, ImpactQ3: 3, ImpactQ4: 2,
			LikelihoodQ1: 4, LikelihoodQ2: 3, LikelihoodQ3: 2, LikelihoodQ4: 2,
		},
		{
			Name: "Terjadinya fraud", Impact: 4, Likelihood: 4, Severity: 92, Category: "Financial", Branch: "Head Office", Description: "Penyalahgunaan wewenang atau kecurangan keuangan di lingkungan internal.",
			ImpactQ1: 4, ImpactQ2: 3, ImpactQ3: 2, ImpactQ4: 1,
			LikelihoodQ1: 4, LikelihoodQ2: 3, LikelihoodQ3: 2, LikelihoodQ4: 1,
		},
		{
			Name: "Implementasi teknologi dan digitalisasi tidak berhasil", Impact: 4, Likelihood: 3, Severity: 72, Category: "Technology", Branch: "Bali Branch", Description: "Kegagalan implementasi sistem baru yang menghambat operasional.",
			ImpactQ1: 4, ImpactQ2: 3, ImpactQ3: 2, ImpactQ4: 2,
			LikelihoodQ1: 4, LikelihoodQ2: 3, LikelihoodQ3: 2, LikelihoodQ4: 2,
		},
		{
			Name: "Pengembangan kompetensi karyawan tidak terlaksana sesuai rencana", Impact: 4, Likelihood: 3, Severity: 58, Category: "Human Resources", Branch: "Head Office", Description: "Kesenjangan keahlian karyawan akibat program training tidak berjalan.",
			ImpactQ1: 4, ImpactQ2: 3, ImpactQ3: 2, ImpactQ4: 2,
			LikelihoodQ1: 3, LikelihoodQ2: 3, LikelihoodQ3: 2, LikelihoodQ4: 2,
		},
		{Name: "Talent Attrition / Brain Drain", Impact: 3, Likelihood: 3, Severity: 50, Category: "Human Resources", Branch: "Head Office", Description: "Loss of key employees and institutional knowledge affecting operational continuity."},
		{Name: "Reputational Damage", Impact: 4, Likelihood: 2, Severity: 75, Category: "Strategic", Branch: "Surabaya Branch", Description: "Significant brand damage due to public scandals, social media crises, or product failures."},
		{Name: "Environmental Compliance Failure", Impact: 3, Likelihood: 2, Severity: 55, Category: "Compliance", Branch: "Bandung Branch", Description: "Violations of environmental regulations leading to fines, shutdowns, or cleanup obligations."},
		{Name: "Operational System Failure", Impact: 4, Likelihood: 3, Severity: 70, Category: "Technology", Branch: "Bali Branch", Description: "Critical failure in core business systems causing operational downtime and revenue loss."},
		{Name: "Insider Trading", Impact: 5, Likelihood: 2, Severity: 90, Category: "Financial", Branch: "Head Office", Description: "Illegal trading of securities based on material, non-public information by employees."},
		{Name: "Workplace Safety Incident", Impact: 3, Likelihood: 2, Severity: 58, Category: "Human Resources", Branch: "Surabaya Branch", Description: "Accidents or hazardous conditions leading to employee injury or regulatory action."},
		{Name: "Third-Party Vendor Risk", Impact: 2, Likelihood: 3, Severity: 40, Category: "Operations", Branch: "Jakarta Branch", Description: "Risks arising from outsourced vendors failing to meet service, security, or compliance standards."},
		{Name: "Intellectual Property Theft", Impact: 4, Likelihood: 2, Severity: 78, Category: "Strategic", Branch: "Bandung Branch", Description: "Unauthorized copying, use, or distribution of company trade secrets and proprietary technology."},
		{Name: "Natural Disaster Impact", Impact: 5, Likelihood: 1, Severity: 60, Category: "Operations", Branch: "Bali Branch", Description: "Disruption from earthquakes, floods, hurricanes, or other catastrophic natural events."},
		{Name: "Interest Rate Fluctuation", Impact: 2, Likelihood: 4, Severity: 35, Category: "Financial", Branch: "Head Office", Description: "Exposure to changing interest rates affecting debt servicing and investment returns."},
		{Name: "Political / Geopolitical Risk", Impact: 3, Likelihood: 3, Severity: 52, Category: "Strategic", Branch: "Jakarta Branch", Description: "Business disruption from political instability, sanctions, trade wars, or regime changes."},
		{Name: "Product Liability", Impact: 4, Likelihood: 1, Severity: 68, Category: "Operations", Branch: "Surabaya Branch", Description: "Legal liability from defective products causing harm to consumers or businesses."},
		{Name: "Pandemic / Health Crisis", Impact: 5, Likelihood: 2, Severity: 82, Category: "Operations", Branch: "Head Office", Description: "Widespread health emergencies causing workforce disruption and operational shutdowns."},
	}

	var risk1ID uuid.UUID
	var risk2ID uuid.UUID

	for _, s := range rawSeeds {
		pID := uuid.New()
		profile := models.RiskProfile{
			ID:           pID,
			DepartmentID: uuid.MustParse(branchToUUID[s.Branch]),
			OwnerID:      uuid.Nil,
			Category:     s.Category,
			Description:  s.Description,
		}
		if err := db.Create(&profile).Error; err != nil {
			return err
		}

		register := models.RiskRegister{
			ID:                   uuid.New(),
			ProfileID:            pID,
			RiskSource:           models.RiskSourceDirect,
			RiskEvent:            s.Name,
			InherentLikelihood:   s.Likelihood,
			InherentImpact:       s.Impact,
			InherentScore:        s.Severity,
			ControlEffectiveness: 0,
			ResidualScore:        s.Severity,
			FinalRiskLevel:       models.RiskFinalLevelHigh,
			Status:               models.RiskRegisterStatusApproved,
		}
		if err := db.Create(&register).Error; err != nil {
			return err
		}

		if s.Name == "Target pendapatan dan laba tidak tercapai" {
			risk1ID = register.ID
		} else if s.Name == "Target efisiensi biaya operasional dan umum tidak tercapai" {
			risk2ID = register.ID
		}

		ast := models.RiskAssessment{
			ID:             uuid.New(),
			RiskRegisterID: register.ID,
			Year:           2026,
			ImpactQ1:       s.ImpactQ1,
			ImpactQ2:       s.ImpactQ2,
			ImpactQ3:       s.ImpactQ3,
			ImpactQ4:       s.ImpactQ4,
			LikelihoodQ1:   s.LikelihoodQ1,
			LikelihoodQ2:   s.LikelihoodQ2,
			LikelihoodQ3:   s.LikelihoodQ3,
			LikelihoodQ4:   s.LikelihoodQ4,
		}
		if ast.ImpactQ1 == 0 {
			ast.ImpactQ1 = s.Impact
			ast.ImpactQ2 = s.Impact
			ast.ImpactQ3 = s.Impact
			ast.ImpactQ4 = s.Impact
			ast.LikelihoodQ1 = s.Likelihood
			ast.LikelihoodQ2 = s.Likelihood
			ast.LikelihoodQ3 = s.Likelihood
			ast.LikelihoodQ4 = s.Likelihood
		}
		if err := db.Create(&ast).Error; err != nil {
			return err
		}
	}

	// Seed mitigations for Risk 1
	if risk1ID != uuid.Nil {
		mit1 := models.RiskMitigation{
			ID:             uuid.New(),
			RiskID:         risk1ID,
			RiskEvent:      "Target pendapatan dan laba usaha tidak tercapai",
			MitigationPlan: "Melakukan kampanye promosi penjualan secara konsisten pada seluruh saluran promosi",
			Supervisor:     "Budi Hartanto",
			PIC:            "Carolina Wijaya",
			UnitInCharge:   "Sales & Marketing",
			StartDate:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:        time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC),
			Notes:          "",
		}
		mit1.Monitoring = models.GenerateMonitoringChecks(mit1.StartDate, mit1.EndDate)
		if data, err := json.Marshal(mit1.Monitoring); err == nil {
			mit1.MonitoringData = string(data)
		}
		db.Create(&mit1)

		mit2 := models.RiskMitigation{
			ID:             uuid.New(),
			RiskID:         risk1ID,
			RiskEvent:      "Target pendapatan dan laba usaha tidak tercapai",
			MitigationPlan: "Meningkatkan jumlah customer dan nilai pembelian customer",
			Supervisor:     "Budi Hartanto",
			PIC:            "Carolina Wijaya",
			UnitInCharge:   "Sales & Marketing",
			StartDate:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:        time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC),
			Notes:          "",
		}
		mit2.Monitoring = models.GenerateMonitoringChecks(mit2.StartDate, mit2.EndDate)
		if data, err := json.Marshal(mit2.Monitoring); err == nil {
			mit2.MonitoringData = string(data)
		}
		db.Create(&mit2)

		mit3 := models.RiskMitigation{
			ID:             uuid.New(),
			RiskID:         risk1ID,
			RiskEvent:      "Target pendapatan dan laba usaha tidak tercapai",
			MitigationPlan: "Memaksimalkan jumlah customer visit",
			Supervisor:     "Budi Hartanto",
			PIC:            "Carolina Wijaya",
			UnitInCharge:   "Sales & Marketing",
			StartDate:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:        time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC),
			Notes:          "",
		}
		mit3.Monitoring = models.GenerateMonitoringChecks(mit3.StartDate, mit3.EndDate)
		if data, err := json.Marshal(mit3.Monitoring); err == nil {
			mit3.MonitoringData = string(data)
		}
		db.Create(&mit3)

		mit4 := models.RiskMitigation{
			ID:             uuid.New(),
			RiskID:         risk1ID,
			RiskEvent:      "Target pendapatan dan laba usaha tidak tercapai",
			MitigationPlan: "Peluncuran produk/jasa baru ke pasar",
			Supervisor:     "Anton Hermawan",
			PIC:            "Budi Hartanto & Carolina Wijaya",
			UnitInCharge:   "Product Development",
			StartDate:      time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
			EndDate:        time.Date(2026, 5, 31, 23, 59, 59, 0, time.UTC),
			Notes:          "",
		}
		mit4.Monitoring = models.GenerateMonitoringChecks(mit4.StartDate, mit4.EndDate)
		if data, err := json.Marshal(mit4.Monitoring); err == nil {
			mit4.MonitoringData = string(data)
		}
		db.Create(&mit4)
	}

	// Seed mitigations for Risk 2
	if risk2ID != uuid.Nil {
		mit5 := models.RiskMitigation{
			ID:             uuid.New(),
			RiskID:         risk2ID,
			RiskEvent:      "Target efisiensi biaya operasional dan umum tidak tercapai",
			MitigationPlan: "Mengontrol biaya operasional dan umum agar efisien namun tetap efektif",
			Supervisor:     "Indarto",
			PIC:            "Wahyu Hidayat & Arief Kuncoro",
			UnitInCharge:   "Operasional",
			StartDate:      time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:        time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC),
			Notes:          "",
		}
		mit5.Monitoring = models.GenerateMonitoringChecks(mit5.StartDate, mit5.EndDate)
		if data, err := json.Marshal(mit5.Monitoring); err == nil {
			mit5.MonitoringData = string(data)
		}
		db.Create(&mit5)
	}

	log.Println("Seeding finished successfully.")
	return nil
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
