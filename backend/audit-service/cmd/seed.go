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

	logger.Info("Seeding audit guidelines and SOPs...")
	if err := seedGuidelinesAndSops(db); err != nil {
		logger.Fatal("Failed to seed audit guidelines and SOPs", logger.LogField("error", err))
		return err
	}

	logger.Info("Database seeding completed successfully")

	return nil
}

func dropTables(db *gorm.DB) error {
	return db.Migrator().DropTable(
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
		"audit_sops",
		"audit_guidelines",
		"audit_charters",
	)
}

func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.AuditCharter{},
		&models.AuditGuideline{},
		&models.AuditSop{},
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
		&models.UploadedPlanDocument{},
		&models.KPIAchievement{},
		&models.WorkPlanRealization{},
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

func seedGuidelinesAndSops(db *gorm.DB) error {
	guidelines := []models.AuditGuideline{
		{
			Name:          "Pedoman pengelolaan Satuan Pengawasan Intern menjadi Kebijakan Satuan Pengelolaan Intern",
			Status:        "Aktif",
			EffectiveDate: "2026-06",
			FileName:      "Pedoman_SPI_v1.0.pdf",
			FileUrl:       "/uploads/dummy_guideline.pdf",
			FileSize:      102400,
		},
		{
			Name:          "Draft Pedoman Pelaksanaan Quality Assurance and Improvement Program (QAIP)",
			Status:        "Sedang Diperbarui",
			EffectiveDate: "2026-06",
			FileName:      "Draft_QAIP_v1.0.pdf",
			FileUrl:       "/uploads/dummy_guideline.pdf",
			FileSize:      102400,
		},
		{
			Name:          "Pedoman Pelaksanaan Audit Operasional",
			Status:        "Sedang Diperbarui",
			EffectiveDate: "2026-06",
			FileName:      "Pedoman_Audit_Operasional.pdf",
			FileUrl:       "/uploads/dummy_guideline.pdf",
			FileSize:      102400,
		},
		{
			Name:          "Pedoman Pelaksanaan Probity Audit",
			Status:        "Sedang Diperbarui",
			EffectiveDate: "2026-06",
			FileName:      "Pedoman_Probity_Audit.pdf",
			FileUrl:       "/uploads/dummy_guideline.pdf",
			FileSize:      102400,
		},
		{
			Name:          "Pedoman Pelaksanaan Assurance Non Audit (Evaluasi dan reviu)",
			Status:        "Sedang Diperbarui",
			EffectiveDate: "2026-06",
			FileName:      "Pedoman_Assurance_Non_Audit.pdf",
			FileUrl:       "/uploads/dummy_guideline.pdf",
			FileSize:      102400,
		},
		{
			Name:          "Pedoman Penulisan dan Pelaporan Laporan Hasil Audit, Laporan Hasil Investigasi serta Laporan Hasil Evaluasi dan Laporan Hasil Reviu",
			Status:        "Sedang Diperbarui",
			EffectiveDate: "2026-06",
			FileName:      "Pedoman_Pelaporan_Audit.pdf",
			FileUrl:       "/uploads/dummy_guideline.pdf",
			FileSize:      102400,
		},
		{
			Name:          "Pedoman Pelaksanaan Monitoring Tindak Lanjut Rekomendasi Audit Internal dan Eksternal",
			Status:        "Aktif",
			EffectiveDate: "2026-06",
			FileName:      "Pedoman_Monitoring_Tindak_Lanjut.pdf",
			FileUrl:       "/uploads/dummy_guideline.pdf",
			FileSize:      102400,
		},
		{
			Name:          "Kerangka Kerja Penerapan Continous Auditing and Countinous Monitoring (CACM)",
			Status:        "Aktif",
			EffectiveDate: "2026-06",
			FileName:      "Kerangka_Kerja_CACM.pdf",
			FileUrl:       "/uploads/dummy_guideline.pdf",
			FileSize:      102400,
		},
		{
			Name:          "Pedoman Kendali mutu Audit",
			Status:        "Sedang Diperbarui",
			EffectiveDate: "2026-06",
			FileName:      "Pedoman_Kendali_Mutu.pdf",
			FileUrl:       "/uploads/dummy_guideline.pdf",
			FileSize:      102400,
		},
		{
			Name:          "Pedoman Penyusunan Program Kerja Audit Tahunan",
			Status:        "Sedang Diperbarui",
			EffectiveDate: "2026-06",
			FileName:      "Pedoman_Penyusunan_PKAT.pdf",
			FileUrl:       "/uploads/dummy_guideline.pdf",
			FileSize:      102400,
		},
	}

	for i := range guidelines {
		if err := db.Create(&guidelines[i]).Error; err != nil {
			return err
		}
	}

	var guideline6 models.AuditGuideline
	if err := db.Where("name LIKE ?", "%Penulisan dan Pelaporan Laporan Hasil Audit%").First(&guideline6).Error; err != nil {
		return err
	}

	sops := []models.AuditSop{
		{
			Name:          "SOP/Working Instruction Penyusunan dan pelaporan Laporan Kompilasi Hasil Audit",
			GuidelineID:   guideline6.ID,
			Status:        "Sedang Diperbarui",
			EffectiveDate: "2026-06",
			FileName:      "SOP_Kompilasi_Hasil_Audit.pdf",
			FileUrl:       "/uploads/dummy_sop.pdf",
			FileSize:      51200,
		},
		{
			Name:          "SOP/Working Instruction Penyusunan dan Pelaporan Laporan Kinerja Triwulan",
			GuidelineID:   guideline6.ID,
			Status:        "Sedang Diperbarui",
			EffectiveDate: "2026-06",
			FileName:      "SOP_Laporan_Kinerja_Triwulan.pdf",
			FileUrl:       "/uploads/dummy_sop.pdf",
			FileSize:      51200,
		},
	}

	for i := range sops {
		if err := db.Create(&sops[i]).Error; err != nil {
			return err
		}
	}

	logger.Info("Sample audit guidelines and SOPs created")
	return nil
}
