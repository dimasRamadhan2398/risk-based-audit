package main

import (
	"audit-service/models"
	"audit-service/pkg/database"
	"audit-service/pkg/logger"
	"time"

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

	logger.Info("Seeding strategic plans...")
	if err := seedStrategicPlans(db); err != nil {
		logger.Fatal("Failed to seed strategic plans", logger.LogField("error", err))
		return err
	}

	logger.Info("Seeding assignment letters...")
	if err := seedAssignmentLetters(db); err != nil {
		logger.Fatal("Failed to seed assignment letters", logger.LogField("error", err))
		return err
	}

	logger.Info("Seeding fieldwork data...")
	if err := seedFieldwork(db); err != nil {
		logger.Fatal("Failed to seed fieldwork data", logger.LogField("error", err))
		return err
	}

	logger.Info("Seeding working papers...")
	if err := seedWorkingPapers(db); err != nil {
		logger.Fatal("Failed to seed working papers", logger.LogField("error", err))
		return err
	}

	logger.Info("Seeding audit executions...")
	if err := seedExecutions(db); err != nil {
		logger.Fatal("Failed to seed audit executions", logger.LogField("error", err))
		return err
	}

	logger.Info("Seeding audit result reports...")
	if err := seedResultReports(db); err != nil {
		logger.Fatal("Failed to seed audit result reports", logger.LogField("error", err))
		return err
	}

	logger.Info("Seeding action taken reports...")
	if err := seedActionTakenReports(db); err != nil {
		logger.Fatal("Failed to seed action taken reports", logger.LogField("error", err))
		return err
	}

	logger.Info("Seeding executive summaries...")
	if err := seedExecutiveSummaries(db); err != nil {
		logger.Fatal("Failed to seed executive summaries", logger.LogField("error", err))
		return err
	}

	logger.Info("Seeding annual plans...")
	if err := seedAnnualPlans(db); err != nil {
		logger.Fatal("Failed to seed annual plans", logger.LogField("error", err))
		return err
	}

	logger.Info("Seeding activity plans...")
	if err := seedActivityPlans(db); err != nil {
		logger.Fatal("Failed to seed activity plans", logger.LogField("error", err))
		return err
	}

	logger.Info("Seeding uploaded plan documents...")
	if err := seedUploadedPlanDocuments(db); err != nil {
		logger.Fatal("Failed to seed uploaded plan documents", logger.LogField("error", err))
		return err
	}

	logger.Info("Seeding imported working papers...")
	if err := seedImportedWorkingPapers(db); err != nil {
		logger.Fatal("Failed to seed imported working papers", logger.LogField("error", err))
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
		&models.ExecutiveSummary{},
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
	var count int64
	db.Model(&models.AuditCharter{}).Count(&count)
	if count > 0 {
		return nil
	}

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

func seedStrategicPlans(db *gorm.DB) error {
	seeds := []models.StrategicPlan{
		{
			Code:               "SO-IA01",
			GoalID:             "G-001",
			StrategicObjective: "Improve Audit Efficiency",
			KPI:                "Revenue Operational Cost",
			Unit:               "%",
			HibHig:             "HIG",
			PeriodType:         "Quartal",
			SelectedPeriod:     "Q1",
			KPITargets:         map[int]string{2024: "300", 2025: "350", 2026: "400", 2027: "450", 2028: "500"},
			InternalAuditSO:    "Optimize resource allocation",
			Actual:             "100",
			Target:             "300",
			Calculation:        "33.33%",
			Status:             "Poor",
		},
		{
			Code:               "SO-IA02",
			GoalID:             "G-002",
			StrategicObjective: "Strengthen Internal Control",
			KPI:                "Customer Satisfaction Index",
			Unit:               "Score",
			HibHig:             "HIG",
			PeriodType:         "Yearly",
			SelectedPeriod:     "2025",
			YearStart:          2022,
			YearEnd:            2026,
			KPITargets:         map[int]string{2025: "90"},
			InternalAuditSO:    "Perform control assessment",
			Actual:             "85",
			Target:             "90",
			Calculation:        "94.44%",
			Status:             "Good",
		},
		{
			Code:               "SO-IA03",
			GoalID:             "G-003",
			StrategicObjective: "Improve Compliance",
			KPI:                "Audit Response Time",
			Unit:               "Hour",
			HibHig:             "HIB",
			PeriodType:         "Quartal",
			SelectedPeriod:     "Q2",
			KPITargets:         map[int]string{2026: "48"},
			Actual:             "24",
			Target:             "48",
			Calculation:        "50.00%",
			Status:             "Good",
		},
	}
	for i := range seeds {
		var existing models.StrategicPlan
		err := db.Where("code = ?", seeds[i].Code).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&seeds[i]).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedAssignmentLetters(db *gorm.DB) error {
	letterDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	letterDate2 := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	letterDate3 := time.Date(2026, 1, 25, 0, 0, 0, 0, time.UTC)
	seeds := []models.AssignmentLetter{
		{
			LetterNumber:    "ST-001/SKAI/2026",
			Status:          "Published",
			AuditTitle:      "Audit Pengendalian Keuangan & Pengeluaran Kas Q1 2026",
			Leader:          "Zeta Ramadhani",
			Category:        "ASSURANCE",
			AuditYear:       "2026",
			AuditTeam:       "SKAI",
			StartPeriod:     "2026-03-01",
			FinishPeriod:    "2026-03-31",
			WorkingUnit:     "Finance",
			ExecutionPeriod: "2026-03-01 to 2026-03-31",
			AuditPurpose:    "Annual Audit",
			LetterDate:      &letterDate,
			CAESignature:    "System",
			MembersList: []models.LetterMember{
				{Name: "Zeta Ramadhani", Role: "Chairperson"},
				{Name: "Budi Santoso", Role: "Supervisor"},
				{Name: "Rina Wulandari", Role: "Member"},
				{Name: "Andi Firmansyah", Role: "Member"},
				{Name: "Dewi Kusumawati", Role: "Person in Charge"},
			},
			PurposeList: []string{
				"Menilai efektivitas pengendalian internal pada divisi keuangan",
				"Memastikan kepatuhan terhadap kebijakan dan prosedur perusahaan",
				"Mengidentifikasi potensi risiko yang dapat mempengaruhi operasional keuangan",
			},
			ScopeList: []string{
				"Proses pencatatan dan pelaporan keuangan periode Januari - Desember 2025",
				"Pengelolaan kas dan setara kas",
				"Rekonsiliasi bank dan laporan arus kas",
				"Kepatuhan terhadap standar akuntansi yang berlaku (PSAK)",
			},
			CcList: []string{
				"President Director",
				"Chief Financial Officer",
				"Head of Internal Audit",
				"Board of Commissioners",
			},
		},
		{
			LetterNumber:    "ST-002/SKAI/2026",
			Status:          "Published",
			AuditTitle:      "Audit Keamanan Sistem Informasi & Infrastruktur ERP 2026",
			Leader:          "Andi Firmansyah",
			Category:        "ASSURANCE",
			AuditYear:       "2026",
			AuditTeam:       "SKAI",
			StartPeriod:     "2026-04-01",
			FinishPeriod:    "2026-04-30",
			WorkingUnit:     "IT",
			ExecutionPeriod: "2026-04-01 to 2026-04-30",
			AuditPurpose:    "IT Security Audit",
			LetterDate:      &letterDate2,
			CAESignature:    "System",
			MembersList: []models.LetterMember{
				{Name: "Andi Firmansyah", Role: "Chairperson"},
				{Name: "Budi Santoso", Role: "Supervisor"},
				{Name: "Dedi Prasetyo", Role: "Member"},
			},
			PurposeList: []string{
				"Evaluasi tata kelola akses pengguna dan keamanan database ERP",
				"Pengujian prosedur disaster recovery dan backup data",
			},
			ScopeList: []string{
				"Hak akses user ERP dan database pusat",
				"Log aktivitas sistem dan perbaikan kerentanan keamanan",
			},
			CcList: []string{
				"Chief Information Officer",
				"Head of Internal Audit",
			},
		},
		{
			LetterNumber:    "ST-003/SKAI/2026",
			Status:          "Published",
			AuditTitle:      "Audit Operasional Branch, Gudang & Logistik 2026",
			Leader:          "Rina Wulandari",
			Category:        "ASSURANCE",
			AuditYear:       "2026",
			AuditTeam:       "SKAI",
			StartPeriod:     "2026-07-01",
			FinishPeriod:    "2026-07-31",
			WorkingUnit:     "Operations",
			ExecutionPeriod: "2026-07-01 to 2026-07-31",
			AuditPurpose:    "Operational Audit",
			LetterDate:      &letterDate3,
			CAESignature:    "System",
			MembersList: []models.LetterMember{
				{Name: "Rina Wulandari", Role: "Chairperson"},
				{Name: "Budi Santoso", Role: "Supervisor"},
			},
			PurposeList: []string{
				"Memastikan pengelolaan persediaan dan stok gudang sesuai SOP",
			},
			ScopeList: []string{
				"Opname stok persediaan gudang pusat dan operasional distribusi branch",
			},
			CcList: []string{
				"Chief Operating Officer",
				"Head of Internal Audit",
			},
		},
		{
			LetterNumber:    "ST-004/SKAI/2026",
			Status:          "Published",
			AuditTitle:      "Audit Kepatuhan Manajemen Risiko & Pengadaan Barang/Jasa 2026",
			Leader:          "Budi Santoso",
			Category:        "ASSURANCE",
			AuditYear:       "2026",
			AuditTeam:       "SKAI",
			StartPeriod:     "2026-08-01",
			FinishPeriod:    "2026-08-31",
			WorkingUnit:     "Procurement",
			ExecutionPeriod: "2026-08-01 to 2026-08-31",
			AuditPurpose:    "Compliance Audit",
			LetterDate:      &letterDate,
			CAESignature:    "System",
			MembersList: []models.LetterMember{
				{Name: "Budi Santoso", Role: "Chairperson"},
				{Name: "Zeta Ramadhani", Role: "Supervisor"},
			},
			PurposeList: []string{"Evaluasi HPS dan kriteria pengadaan barang/jasa"},
			ScopeList:   []string{"Kontrak vendor SCM"},
			CcList:      []string{"Head of Procurement", "Head of Internal Audit"},
		},
		{
			LetterNumber:    "ST-005/SKAI/2026",
			Status:          "Published",
			AuditTitle:      "Audit Pengendalian K3LH & Manufaktur Pembangkit 2026",
			Leader:          "Dewi Kusumawati",
			Category:        "ASSURANCE",
			AuditYear:       "2026",
			AuditTeam:       "SKAI",
			StartPeriod:     "2026-09-01",
			FinishPeriod:    "2026-09-30",
			WorkingUnit:     "Maintenance",
			ExecutionPeriod: "2026-09-01 to 2026-09-30",
			AuditPurpose:    "HSE Audit",
			LetterDate:      &letterDate,
			CAESignature:    "System",
			MembersList: []models.LetterMember{
				{Name: "Dewi Kusumawati", Role: "Chairperson"},
				{Name: "Rina Wulandari", Role: "Member"},
			},
			PurposeList: []string{"Evaluasi implementasi K3LH"},
			ScopeList:   []string{"Fasilitas pemadam kebakaran hidran"},
			CcList:      []string{"Chief Safety Officer", "Head of Internal Audit"},
		},
		{
			LetterNumber:    "020/ST/01/KSIAD/2023",
			Status:          "Published",
			AuditTitle:      "Audit Operasional Pengelolaan Pembangkitan UPDK Kepulauan Riau",
			Leader:          "Tomy Afrilianto",
			Category:        "ASSURANCE",
			AuditYear:       "2023",
			AuditTeam:       "SKAI",
			StartPeriod:     "2023-01-01",
			FinishPeriod:    "2023-08-31",
			WorkingUnit:     "Operasi & Pemeliharaan",
			ExecutionPeriod: "Januari 2023 s.d Agustus 2023",
			AuditPurpose:    "Operational Audit",
			LetterDate:      &letterDate,
			CAESignature:    "System",
			MembersList: []models.LetterMember{
				{Name: "Tomy Afrilianto", Role: "Chairperson"},
				{Name: "Robert Sunarijanto", Role: "Supervisor"},
			},
			PurposeList: []string{"Evaluasi ketersediaan pembangkit UPDK Kepulauan Riau"},
			ScopeList:   []string{"Operasional O&M PLTU"},
			CcList:      []string{"Head of SKAI", "Board of Directors"},
		},
	}
	for i := range seeds {
		var existing models.AssignmentLetter
		err := db.Where("letter_number = ?", seeds[i].LetterNumber).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&seeds[i]).Error; err != nil {
				return err
			}
		} else {
			// Update audit title and other fields if existing
			db.Model(&existing).Updates(map[string]interface{}{
				"audit_title":      seeds[i].AuditTitle,
				"working_unit":     seeds[i].WorkingUnit,
				"execution_period": seeds[i].ExecutionPeriod,
			})
		}
	}
	return nil
}

func seedFieldwork(db *gorm.DB) error {
	// Interviews
	interviews := []models.FieldworkInterview{
		{
			AssignmentLetterID:  "ST-001/SKAI/2026",
			Interviewee:         "Ahmad Yani",
			IntervieweePosition: "Head of Finance",
			Interviewer:         "Zeta Ramadhani",
			InterviewerPosition: "Chairperson",
			Date:                "2026-03-05",
			Topic:               "Internal Control over Financial Reporting (ICOFR) implementation, segregation of duties in payment approvals, and monthly bank reconciliation process.",
			FileName:            "Meeting_Minutes_ICOFR.pdf",
		},
		{
			AssignmentLetterID:  "ST-001/SKAI/2026",
			Interviewee:         "Rudi Hermawan",
			IntervieweePosition: "IT Manager",
			Interviewer:         "Andi Firmansyah",
			InterviewerPosition: "Member",
			Date:                "2026-03-08",
			Topic:               "Access controls to the ERP system, user privilege review procedures, and backup recovery tests.",
			FileName:            "IT_Access_Controls_Interview.pdf",
		},
		{
			AssignmentLetterID:  "ST-002/SKAI/2026",
			Interviewee:         "Bambang Susilo",
			IntervieweePosition: "Head of IT",
			Interviewer:         "Andi Firmansyah",
			InterviewerPosition: "Chairperson",
			Date:                "2026-04-05",
			Topic:               "Keamanan Database ERP & Multi-Factor Authentication enforcement.",
			FileName:            "Interview_IT_Security_ST002.pdf",
		},
		{
			AssignmentLetterID:  "ST-003/SKAI/2026",
			Interviewee:         "Hendra Wijaya",
			IntervieweePosition: "Warehouse Manager",
			Interviewer:         "Rina Wulandari",
			InterviewerPosition: "Chairperson",
			Date:                "2026-07-08",
			Topic:               "Manajemen Stok Persediaan Gudang & Prosedur Opname Fisik.",
			FileName:            "Interview_Warehouse_ST003.pdf",
		},
	}
	for i := range interviews {
		var existing models.FieldworkInterview
		err := db.Where("assignment_letter_id = ? AND interviewee = ?", interviews[i].AssignmentLetterID, interviews[i].Interviewee).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&interviews[i]).Error; err != nil {
				return err
			}
		}
	}

	// Observations
	observations := []models.FieldworkObservation{
		{
			AssignmentLetterID: "ST-001/SKAI/2026",
			Activity:           "Observe cash count process in main vault and verify safety box lock code authorization.",
			Location:           "Headquarters - Vault Room",
			Date:               "2026-03-10",
			Observer:           "Rina Wulandari",
			FileName:           "Cash_Count_Observation.pdf",
		},
		{
			AssignmentLetterID: "ST-001/SKAI/2026",
			Activity:           "Walkthrough observation of transaction recording inside ERP and checking automated system logs.",
			Location:           "IT Server Room & Finance Desk",
			Date:               "2026-03-12",
			Observer:           "Andi Firmansyah",
		},
		{
			AssignmentLetterID: "ST-002/SKAI/2026",
			Activity:           "Pengujian Disaster Recovery & Backup Restore Data ERP.",
			Location:           "Data Center Utama",
			Date:               "2026-04-12",
			Observer:           "Andi Firmansyah",
		},
		{
			AssignmentLetterID: "ST-003/SKAI/2026",
			Activity:           "Inspeksi Fisik dan Pengukuran Suhu Gudang Bahan Kimia.",
			Location:           "Gudang Logistik Branch A",
			Date:               "2026-07-15",
			Observer:           "Rina Wulandari",
		},
	}
	for i := range observations {
		var existing models.FieldworkObservation
		err := db.Where("assignment_letter_id = ? AND activity = ?", observations[i].AssignmentLetterID, observations[i].Activity).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&observations[i]).Error; err != nil {
				return err
			}
		}
	}

	// Documents
	documents := []models.FieldworkDocument{
		{
			AssignmentLetterID: "ST-001/SKAI/2026",
			DocumentName:       "General Ledger 2025",
			Description:        "Full year accounting ledger containing all transactions for balance sheet validation.",
			RequiredDate:       "2026-03-02",
			FileName:           "General_Ledger_2025_Final.xlsx",
		},
		{
			AssignmentLetterID: "ST-001/SKAI/2026",
			DocumentName:       "ERP System Access Logs",
			Description:        "Active user list and permission matrix for segregation of duties audit.",
			RequiredDate:       "2026-03-05",
			FileName:           "ERP_Access_Logs_Q4_2025.csv",
		},
		{
			AssignmentLetterID: "ST-002/SKAI/2026",
			DocumentName:       "ERP Security Matrix Log",
			Description:        "Log autentikasi dan daftar hak akses pengguna ERP Q1 2026.",
			RequiredDate:       "2026-04-02",
			FileName:           "ERP_Security_Matrix.xlsx",
		},
		{
			AssignmentLetterID: "ST-003/SKAI/2026",
			DocumentName:       "Kartu Stok & Laporan Opname",
			Description:        "Berita acara stok opname barang semester 1 2026.",
			RequiredDate:       "2026-07-03",
			FileName:           "Stok_Opname_S1_2026.pdf",
		},
	}
	for i := range documents {
		var existing models.FieldworkDocument
		err := db.Where("assignment_letter_id = ? AND document_name = ?", documents[i].AssignmentLetterID, documents[i].DocumentName).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&documents[i]).Error; err != nil {
				return err
			}
		}
	}

	// Samples
	samples := []models.FieldworkSample{
		{
			AssignmentLetterID: "ST-001/SKAI/2026",
			DocumentName:       "Procurement Invoice",
			DocumentNumber:     "INV-2025-0988",
			Date:               "2025-11-12",
			Description:        "Sample transaction for validation of purchase order matching and invoice approval.",
		},
		{
			AssignmentLetterID: "ST-001/SKAI/2026",
			DocumentName:       "Bank Statement Reconciliation",
			DocumentNumber:     "BR-2025-12",
			Date:               "2025-12-31",
			Description:        "Reconciliation report for main bank account verifying outstanding checks and deposits.",
		},
		{
			AssignmentLetterID: "ST-002/SKAI/2026",
			DocumentName:       "Log Akses User Admin",
			DocumentNumber:     "LOG-ERP-2026-04",
			Date:               "2026-04-10",
			Description:        "Sample log akses superadmin database ERP.",
		},
		{
			AssignmentLetterID: "ST-003/SKAI/2026",
			DocumentName:       "Work Order Pengeluaran Barang",
			DocumentNumber:     "WO-LOG-2026-89",
			Date:               "2026-07-20",
			Description:        "Sample otorisasi pengeluaran barang persediaan.",
		},
	}
	for i := range samples {
		var existing models.FieldworkSample
		err := db.Where("assignment_letter_id = ? AND document_number = ?", samples[i].AssignmentLetterID, samples[i].DocumentNumber).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&samples[i]).Error; err != nil {
				return err
			}
		}
	}

	// Test Controls
	testControls := []models.FieldworkTestControl{
		{
			AssignmentLetterID: "ST-001/SKAI/2026",
			ControlName:        "Payment Authorization Limit",
			ControlDescription: "All payments above Rp 50,000,000 must be approved by the Finance Director.",
			ControlType:        "Preventive",
			TestProcedure:      "Select a sample of payments > Rp 50m and verify presence of Director signature or digital approval.",
			TestResult:         "Effective",
			Finding:            "No deviations found. All sample vouchers had the required dual approval.",
			Recommendation:     "Maintain current process and ensure limits are updated in system configuration.",
			MitigationPlan:     "Routine system configuration audit.",
			PIC:                "Ahmad Yani (Head of Finance)",
			DueDate:            "2026-06-30",
		},
		{
			AssignmentLetterID: "ST-001/SKAI/2026",
			ControlName:        "IT Backup Daily Execution",
			ControlDescription: "ERP database backups must run automatically every night at 23:00.",
			ControlType:        "Automated",
			TestProcedure:      "Review backup logs for the month of December 2025 and verify successful backup statuses.",
			TestResult:         "Ineffective",
			Finding:            "On 3 days (Dec 12, 13, and 18), backups failed due to disk space issues. No alert was sent.",
			Recommendation:     "Implement automated notification alerting IT team when backup status is failed.",
			MitigationPlan:     "Setup automated SMTP notification inside backup script.",
			PIC:                "Rudi Hermawan (IT Manager)",
			DueDate:            "2026-04-15",
		},
	}
	for i := range testControls {
		var existing models.FieldworkTestControl
		err := db.Where("assignment_letter_id = ? AND control_name = ?", testControls[i].AssignmentLetterID, testControls[i].ControlName).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&testControls[i]).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func seedWorkingPapers(db *gorm.DB) error {
	// 1. Header
	headers := []models.WorkingPaperHeader{
		{
			AssignmentLetterID: "ST-001/SKAI/2026",
			AuditPurpose:       "Assess the reliability and integrity of financial recording processes.",
			BusinessProcess:    "Procurement-to-Pay",
			Period:             "2026-03-01 s/d 2026-03-31",
			Location:           "Jakarta Headquarters",
			TeamMembers: []models.TeamMember{
				{ID: 1, Name: "Zeta Ramadhani", Role: "Chairperson"},
				{ID: 2, Name: "Rina Wulandari", Role: "Member"},
			},
		},
		{
			AssignmentLetterID: "ST-002/SKAI/2026",
			AuditPurpose:       "Evaluasi keamanan infrastruktur ERP & kontrol akses database.",
			BusinessProcess:    "IT Governance & ERP Security",
			Period:             "2026-04-01 to 2026-04-30",
			Location:           "Data Center & IT Office",
			TeamMembers: []models.TeamMember{
				{ID: 1, Name: "Andi Firmansyah", Role: "Chairperson"},
				{ID: 2, Name: "Dedi Prasetyo", Role: "Member"},
			},
		},
		{
			AssignmentLetterID: "ST-003/SKAI/2026",
			AuditPurpose:       "Audit akurasi stok opname dan distribusi logistik.",
			BusinessProcess:    "Inventory & Logistics Operations",
			Period:             "2026-07-01 to 2026-07-31",
			Location:           "Central Warehouse",
			TeamMembers: []models.TeamMember{
				{ID: 1, Name: "Rina Wulandari", Role: "Chairperson"},
				{ID: 2, Name: "Budi Santoso", Role: "Supervisor"},
			},
		},
	}
	for i := range headers {
		var existing models.WorkingPaperHeader
		err := db.Where("assignment_letter_id = ? AND business_process = ?", headers[i].AssignmentLetterID, headers[i].BusinessProcess).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&headers[i]).Error; err != nil {
				return err
			}
		}
	}

	// 2. Risks
	risks := []models.WorkingPaperRisk{
		{
			WorkingPaperID:     "ST-001/SKAI/2026",
			Risk:               "Unauthorized payments are processed leading to financial loss.",
			Taxonomy:           "Financial",
			RiskLevel:          "High",
			ControlDescription: "Dual authorization matrix is set up in ERP system.",
		},
		{
			WorkingPaperID:     "ST-001/SKAI/2026",
			Risk:               "IT backup failure causes data loss in transaction history.",
			Taxonomy:           "Operational",
			RiskLevel:          "Medium",
			ControlDescription: "Scheduled nightly automated database backups.",
		},
		{
			WorkingPaperID:     "ST-002/SKAI/2026",
			Risk:               "Akses ilegal ke database ERP akibat ketiadaan MFA dan patch tertunda.",
			Taxonomy:           "IT & Cyber Security",
			RiskLevel:          "High",
			ControlDescription: "User access role matrix and quarterly patch schedule",
		},
		{
			WorkingPaperID:     "ST-003/SKAI/2026",
			Risk:               "Kerusakan persediaan barang akibat kendala lingkungan gudang.",
			Taxonomy:           "Operational",
			RiskLevel:          "Medium",
			ControlDescription: "Daily temperature monitoring log and security gate pass",
		},
	}
	for i := range risks {
		var existing models.WorkingPaperRisk
		err := db.Where("working_paper_id = ? AND risk = ?", risks[i].WorkingPaperID, risks[i].Risk).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&risks[i]).Error; err != nil {
				return err
			}
		}
	}

	// 3. Samples
	pop := 150
	sz := 15
	pop2 := 40
	sz2 := 10
	pop3 := 200
	sz3 := 20
	tVal := true
	fVal := false
	samples := []models.WorkingPaperSample{
		{
			WorkingPaperID: "ST-001/SKAI/2026",
			Population:     &pop,
			SampleSize:     &sz,
			Samples: []models.SampleDoc{
				{ID: 1, Document: "JV-2025-0012", L1: &tVal, L2: &tVal, L3: &tVal},
				{ID: 2, Document: "JV-2025-0045", L1: &tVal, L2: &fVal, L3: &tVal},
			},
			Conclusion: "Out of 15 sample documents, 1 sample failed to show secondary approval. The control has moderate deviation.",
		},
		{
			WorkingPaperID: "ST-002/SKAI/2026",
			Population:     &pop2,
			SampleSize:     &sz2,
			Samples: []models.SampleDoc{
				{ID: 1, Document: "ACC-ERP-01", L1: &tVal, L2: &tVal, L3: &tVal},
				{ID: 2, Document: "ACC-ERP-02", L1: &tVal, L2: &fVal, L3: &fVal},
			},
			Conclusion: "Terdeteksi 2 user superadmin belum mengaktifkan MFA.",
		},
		{
			WorkingPaperID: "ST-003/SKAI/2026",
			Population:     &pop3,
			SampleSize:     &sz3,
			Samples: []models.SampleDoc{
				{ID: 1, Document: "STK-GDG-101", L1: &tVal, L2: &tVal, L3: &tVal},
			},
			Conclusion: "Terdapat selisih 3 unit barang persediaan pada sampel gudang.",
		},
	}
	for i := range samples {
		var existing models.WorkingPaperSample
		err := db.Where("working_paper_id = ?", samples[i].WorkingPaperID).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&samples[i]).Error; err != nil {
				return err
			}
		}
	}

	// 4. Cause / AOI & RCA
	causes := []models.WorkingPaperCause{
		{
			WorkingPaperID: "ST-001/SKAI/2026",
			Condition:      "Payment authorization override occurred twice in Q4 2025 without proper manual override forms.",
			Criteria:       "Financial Authority Matrix SOP Section 4.2 requires override forms signed by CFO.",
			Impact:         "Potential unauthorized payment processing risk and lack of accountability.",
			EvidenceFile:   "ERP_Override_Log_Dec.pdf",
			RootCause: []models.RootCauseItem{
				{ID: 1, Method: "People", W1: "Auditee rushed the payment due to vendor pressure", W2: "CFO was traveling out of country", W3: "Temporary password sharing occurred"},
			},
		},
		{
			WorkingPaperID: "ST-002/SKAI/2026",
			Condition:      "Pengujian DR belum rutin dilakukan setiap semester",
			Criteria:       "SOP IT Security Section 3",
			Impact:         "Risiko downtime sistem ERP saat insiden",
			EvidenceFile:   "DRC_Test_Log.pdf",
			RootCause: []models.RootCauseItem{
				{ID: 1, Method: "System", W1: "Jadwal simulasi bentrok dengan maintenance harian", W2: "Keterbatasan environment staging", W3: "Belum ada otomasisasi failover"},
			},
		},
		{
			WorkingPaperID: "ST-003/SKAI/2026",
			Condition:      "Sensor suhu ruangan gudang belum otomatis terhubung ke dashboard IoT",
			Criteria:       "SOP Pengelolaan Warehouse 2025",
			Impact:         "Potensi penurunan kualitas material",
			EvidenceFile:   "Warehouse_Temp_Report.pdf",
			RootCause: []models.RootCauseItem{
				{ID: 1, Method: "Equipment", W1: "Sensor suhu versi lama belum memiliki koneksi IP", W2: "Pencatatan masih manual di logbook paper", W3: "Keterlambatan update hardware"},
			},
		},
	}
	for i := range causes {
		var existing models.WorkingPaperCause
		err := db.Where("working_paper_id = ? AND condition = ?", causes[i].WorkingPaperID, causes[i].Condition).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&causes[i]).Error; err != nil {
				return err
			}
		}
	}

	// 5. Plan
	plans := []models.WorkingPaperPlan{
		{
			WorkingPaperID:    "ST-001/SKAI/2026",
			Recommendation:    "Enforce strict system block on ERP override and disable temporary password sharing mechanisms.",
			Response:          "Agreed. IT will implement strict AD group policy and remove override privilege from staff level.",
			ActionDescription: "Deactivate user account sharing and enforce AD single sign-on constraints.",
			PIC:               "Rudi Hermawan (IT Manager)",
			PeriodAction:      "2026-05-31",
		},
		{
			WorkingPaperID:    "ST-002/SKAI/2026",
			Recommendation:    "Wajibkan MFA seluruh akun admin ERP dan jadwalkan DRC drill semesteran",
			Response:          "Disetujui. IT akan menerapkan MFA per 15 Mei 2026.",
			ActionDescription: "Implementasi TOTP MFA pada portal SSO admin.",
			PIC:               "Bambang Susilo (Head of IT)",
			PeriodAction:      "2026-05-31",
		},
		{
			WorkingPaperID:    "ST-003/SKAI/2026",
			Recommendation:    "Pasang IoT sensor suhu digital 24/7 di seluruh area gudang bahan kimia",
			Response:          "Siap dilaksanakan pada pengadaan Q3 2026.",
			ActionDescription: "Pengadaan dan konfigurasi sensor IoT suhu dengan alert Telegram/Email.",
			PIC:               "Hendra Wijaya (Warehouse Manager)",
			PeriodAction:      "2026-08-31",
		},
	}
	for i := range plans {
		var existing models.WorkingPaperPlan
		err := db.Where("working_paper_id = ? AND recommendation = ?", plans[i].WorkingPaperID, plans[i].Recommendation).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&plans[i]).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func seedExecutions(db *gorm.DB) error {
	seeds := []models.AuditExecution{
		{
			Ref:                    "ST-001/SKAI/2026",
			Name:                   "Audit Operasional Keuangan",
			Category:               "Assurance",
			Progress:               80,
			LeadAuditor:            "Zeta Ramadhani",
			Status:                 "in_progress",
			StatusDetail:           "Fieldwork Testing",
			SampleDataTestControls: &models.TestControlsSub{Progress: 90, Description: "Controls over invoice payments checked."},
			WorkingPapers:          &models.WorkingPapersSub{Condition: "Good", Criteria: "SOP complied"},
			ActionPlanImprovements: &models.ImprovementsSub{Recommendation: "Add validation check", Deadline: "2026-06-30", PIC: "Finance Team"},
			LatestUpdateProgress:   &models.LatestUpdateSub{Attachment: "BA_Fieldwork.pdf", Description: "Drafting working papers and reviewing findings."},
		},
		{
			Ref:          "ST-002/SKAI/2026",
			Name:         "Audit IT Security & Compliance",
			Category:     "Assurance",
			Progress:     40,
			LeadAuditor:  "Andi Firmansyah",
			Status:       "in_progress",
			StatusDetail: "Document Review",
		},
		{
			Ref:          "ST-003/SKAI/2026",
			Name:         "Audit Unit Procurement",
			Category:     "Assurance",
			Progress:     0,
			LeadAuditor:  "Rina Wulandari",
			Status:       "planned",
			StatusDetail: "Preparation of Work Plan",
		},
	}
	for i := range seeds {
		var existing models.AuditExecution
		err := db.Where("ref = ?", seeds[i].Ref).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&seeds[i]).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedResultReports(db *gorm.DB) error {
	reportDate1 := time.Date(2023, 9, 22, 0, 0, 0, 0, time.UTC)
	reportDate2 := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
	reportDate3 := time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC)
	reportDate4 := time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)
	reportDate5 := time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC)
	seeds := []models.AuditResultReport{
		{
			AssignmentLetterID: "020/ST/01/KSIAD/2023",
			ReportTitle:        "Audit Operasional Pengelolaan Pembangkitan UPDK Kepulauan Riau",
			FindingsCount:      8,
			ReportNumber:       "020/LHA/01/KS IAD/2023",
			Title:              "Laporan Hasil Audit Operasional Tahun 2023",
			AuditObject:        "Unit Pelaksana Pengendalian Pembangkitan Kepulauan Riau",
			Department:         "Operasi & Pemeliharaan",
			AuditPeriod:        "Januari 2023 s.d Agustus 2023",
			ExecutiveSummary:   "Audit Operasional Tahun 2023 di Unit Pelaksana Pengendalian Pembangkitan Kepulauan Riau meliputi ketersediaan pembangkit, K3LH, manajemen risiko, dan pengadaan barang/jasa.",
			Scope:              "Ketersediaan dan Keandalan Pembangkit (EAF/EFOR), Efisiensi Pembangkit & Energi Primer, Kepatuhan K3LH, Manajemen Risiko, Pengadaan Barang/Jasa.",
			Methodology:        "Pendekatan Risiko, Pengujian Terbatas Dokumen, dan Uji Petik (Non-statistical sampling).",
			FindingSummary:     "Ditemukan 8 Area of Improvement meliputi Pengelolaan Manajemen Risiko, Overhaul ME+ PLTU TBK #1, Peralatan Lab Batubara, dan Implementasi Maximo WPC.",
			Recommendation:     "Manajemen UPDK KEPRI perlu melakukan refreshment manajemen risiko, penyusunan DMR, dan sertifikasi personil pemeliharaan.",
			Conclusion:         "Tata kelola dan pengendalian internal berjalan baik dengan 1 Risk Management AoI dan 8 Internal Control AoI.",
			PreparedBy:         "Tomy Afrilianto",
			ReviewedBy:         "Robert Sunarijanto",
			ApprovedBy:         "Head of SKAI",
			ReportDate:         &reportDate1,
			Status:             "APPROVED",
			Attachment:         "LHA_UPDK_Kepri_2023.pdf",
			Findings: []models.AuditReportFinding{
				{Title: "Pengelolaan Manajemen Risiko Belum Sepenuhnya Sesuai Kebijakan Masa Transisi HSH", Category: "Very Significant", Action: "Perbaikan SOP"},
				{Title: "Pelaksanaan Overhaul UPDK KEPRI Belum Optimal Terjadi PE 6 Hari pada ME+ PLTU TBK #1", Category: "Very Significant", Action: "Evaluasi jadwal"},
				{Title: "Peralatan Lab Milik Perusahaan Belum Digunakan Secara Optimal Sebagai Pembanding Surveyor", Category: "Significant", Action: "Kalibrasi ulang"},
				{Title: "Data Maturity Level Manajemen Aset Belum Lengkap Terbatalnya Fitur Maximo WPC", Category: "Significant", Action: "Update Maximo"},
				{Title: "Terdapat Penyusunan HPS dan Pemanfaatan ERP Tidak Sesuai Ketentuan SCM", Category: "Significant", Action: "Review HPS"},
				{Title: "Program Pemeliharaan Aset Tetap Belum Diakui Kepemilikannya Menggunakan Anggaran Operasi", Category: "Quite Significant", Action: "Inventarisasi aset"},
				{Title: "Terdapat Kontrak Pekerjaan Sejenis Yang Tidak Digabungkan (Strategi Squeezing)", Category: "Quite Significant", Action: "Review kontrak"},
				{Title: "Pengelolaan K3 dan Keamanan di UPDK KEPRI Belum Optimal (Fire Fighting & Lightning)", Category: "Significant", Action: "Audit K3"},
			},
		},
		{
			AssignmentLetterID: "ST-001/SKAI/2026",
			ReportTitle:        "Laporan Hasil Audit Operasional Keuangan",
			FindingsCount:      5,
			ReportNumber:       "021/LHA/01/KS IAD/2026",
			Title:              "Laporan Hasil Audit Operasional Keuangan 2025",
			AuditObject:        "PT AIFL Indonesia - Departemen Keuangan",
			Department:         "Finance",
			AuditPeriod:        "Tahun Buku 2025",
			ExecutiveSummary:   "Audit dilakukan untuk mengevaluasi efektivitas ICOFR dan kepatuhan terhadap SOP pembayaran.",
			Scope:              "Pengujian kas, bank, rekonsiliasi, invoice matching, dan otorisasi limit.",
			Methodology:        "Walkthrough, wawancara, sample testing, dan verifikasi dokumen pendukung.",
			FindingSummary:     "Ditemukan kelemahan minor pada backup database ERP yang sempat gagal.",
			Recommendation:     "Disarankan membuat SMTP alert untuk kegagalan backup data.",
			Conclusion:         "Secara umum, pengendalian internal memadai dengan beberapa area peningkatan.",
			PreparedBy:         "Zeta Ramadhani",
			ReviewedBy:         "Budi Santoso",
			ApprovedBy:         "System",
			ReportDate:         &reportDate2,
			Status:             "Draft",
			Attachment:         "LHA_Keuangan_2025_Final.pdf",
			Findings: []models.AuditReportFinding{
				{Title: "Keterlambatan rekonsiliasi kas harian cabang utama", Category: "Very Significant", Action: "Perbaikan jadwal harian"},
				{Title: "Kelemahan kontrol otorisasi transaksi di atas Rp 500jt", Category: "Very Significant", Action: "Review limit otorisasi"},
				{Title: "Selisih pencatatan inventaris fisik vs buku besar", Category: "Very Significant", Action: "Stok opname ulang"},
				{Title: "Dokumentasi bukti transfer eksternal tidak lengkap", Category: "Significant", Action: "Lengkapi berkas transfer"},
				{Title: "Akses user kasir tidak di-nonaktifkan setelah mutasi", Category: "Significant", Action: "Nonaktifkan akun user"},
			},
		},
		{
			AssignmentLetterID: "ST-002/SKAI/2026",
			ReportTitle:        "Laporan Hasil Audit Keamanan Sistem Informasi & ERP 2026",
			FindingsCount:      4,
			ReportNumber:       "022/LHA/01/KS IAD/2026",
			Title:              "Laporan Hasil Audit Keamanan Sistem Informasi 2026",
			AuditObject:        "PT AIFL Indonesia - Divisi Teknologi Informasi",
			Department:         "IT",
			AuditPeriod:        "Q1 2026",
			ExecutiveSummary:   "Audit mengevaluasi tata kelola akses pengguna dan keamanan database ERP serta backup data.",
			Scope:              "Pengujian keamanan database, MFA, simulasi DR, dan log audit akses.",
			Methodology:        "Vulnerability scanning, log review, and interview.",
			FindingSummary:     "Kebutuhan MFA mandatory untuk akun superadmin ERP.",
			Recommendation:     "Implementasi MFA dan jalankan DRC drill semesteran.",
			Conclusion:         "Tata kelola IT terstruktur dengan perhatian khusus pada penguatan autentikasi.",
			PreparedBy:         "Andi Firmansyah",
			ReviewedBy:         "Budi Santoso",
			ApprovedBy:         "Head of IT",
			ReportDate:         &reportDate3,
			Status:             "Final",
			Attachment:         "LHA_IT_Security_2026.pdf",
			Findings: []models.AuditReportFinding{
				{Title: "Keterlambatan patch keamanan server database ERP", Category: "Very Significant", Action: "Update patch rutin"},
				{Title: "Akses Superadmin ERP belum menggunakan Multi-Factor Authentication", Category: "Very Significant", Action: "Implementasi MFA mandatory"},
				{Title: "Prosedur Backup Data belum diuji pemulihannya secara berkala", Category: "Significant", Action: "Jadwalkan DRC drill"},
				{Title: "Log audit aktivitas sistem informasi belum di-review mingguan", Category: "Quite Significant", Action: "Setup SOC log alert"},
			},
		},
		{
			AssignmentLetterID: "ST-003/SKAI/2026",
			ReportTitle:        "Laporan Hasil Audit Operasional Gudang & Persediaan Logistik 2026",
			FindingsCount:      3,
			ReportNumber:       "023/LHA/01/KS IAD/2026",
			Title:              "Laporan Hasil Audit Gudang & Logistik 2026",
			AuditObject:        "PT AIFL Indonesia - Unit Gudang Logistik",
			Department:         "Operations",
			AuditPeriod:        "Semester 1 2026",
			ExecutiveSummary:   "Audit mengevaluasi akurasi pencatatan stok gudang persediaan dan pengelolaan distribusi.",
			Scope:              "Stock opname, verifikasi gate pass, dan kontrol suhu gudang.",
			Methodology:        "Physical counting, sample testing, observation.",
			FindingSummary:     "Terdapat selisih stok minor dan sensor suhu ruangan belum terintegrasi IoT.",
			Recommendation:     "Pasang sensor IoT suhu digital 24/7 dan revisi SOP pengeluaran barang.",
			Conclusion:         "Operasional gudang berjalan dengan baik dengan perbaikan sistem pemantauan lingkungan.",
			PreparedBy:         "Rina Wulandari",
			ReviewedBy:         "Budi Santoso",
			ApprovedBy:         "Head of Operations",
			ReportDate:         &reportDate4,
			Status:             "Draft",
			Attachment:         "LHA_Gudang_Logistik_2026.pdf",
			Findings: []models.AuditReportFinding{
				{Title: "Selisih fisik barang material persediaan gudang cabang", Category: "Significant", Action: "Investigasi selisih stok"},
				{Title: "Suhu penyimpanan gudang bahan kimia belum terpantau 24/7", Category: "Quite Significant", Action: "Pasang IoT sensor suhu"},
				{Title: "Pengeluaran material proyek tanpa Work Order yang disetujui", Category: "Significant", Action: "Kunci sistem release barang"},
			},
		},
		{
			AssignmentLetterID: "ST-004/SKAI/2026",
			ReportTitle:        "Laporan Hasil Audit Kepatuhan Procurement & SCM 2026",
			FindingsCount:      2,
			ReportNumber:       "024/LHA/01/KS IAD/2026",
			Title:              "Laporan Hasil Audit Kepatuhan Procurement",
			AuditObject:        "PT AIFL Indonesia - Pengadaan",
			Department:         "Procurement",
			AuditPeriod:        "Q2 2026",
			ExecutiveSummary:   "Audit evaluasi pelaksanaan rekomendasi audit internal dan kepatuhan pengadaan SCM.",
			Scope:              "Monitoring tindak lanjut LHA SCM.",
			Methodology:        "Verification & compliance testing.",
			FindingSummary:     "Sebagian besar rekomendasi telah ditindaklanjuti.",
			Recommendation:     "Lampirkan kertas kerja survei HPS vendor.",
			Conclusion:         "Tindak lanjut berjalan efektif.",
			PreparedBy:         "Budi Santoso",
			ReviewedBy:         "Zeta Ramadhani",
			ApprovedBy:         "System",
			ReportDate:         &reportDate5,
			Status:             "Final",
			Attachment:         "LHA_Procurement_2026.pdf",
			Findings: []models.AuditReportFinding{
				{Title: "Penyusunan HPS pengadaan komponen turbin belum melampirkan kertas kerja survei harga", Category: "Quite Significant", Action: "Lampirkan bukti survei HPS"},
				{Title: "Monitoring pencairan jaminan bank vendor belum terintegrasi ERP", Category: "Not Significant", Action: "Fitur reminder otomatis ERP"},
			},
		},
		{
			AssignmentLetterID: "ST-005/SKAI/2026",
			ReportTitle:        "Laporan Hasil Audit K3LH & Pemeliharaan Aset Pembangkit 2026",
			FindingsCount:      3,
			ReportNumber:       "025/LHA/01/KS IAD/2026",
			Title:              "Laporan Hasil Audit K3LH Pembangkit 2026",
			AuditObject:        "PT AIFL Indonesia - Unit Pemeliharaan",
			Department:         "Maintenance",
			AuditPeriod:        "Q3 2026",
			ExecutiveSummary:   "Executive Summary Individual DOC-EXSUM-Q1-2026 untuk Laporan Hasil Audit K3LH & Pemeliharaan Aset Pembangkit.",
			Scope:              "Inspeksi fasilitas K3LH dan hidran.",
			Methodology:        "Physical inspection & certification review.",
			FindingSummary:     "Diperlukan penjadwalan pemeliharaan hidran rutin.",
			Recommendation:     "Segera jadwalkan pemeliharaan hidran dan pembaruan sertifikat K3.",
			Conclusion:         "K3LH beroperasi aman.",
			PreparedBy:         "Dewi Kusumawati",
			ReviewedBy:         "Rina Wulandari",
			ApprovedBy:         "Chief Safety Officer",
			ReportDate:         &reportDate5,
			Status:             "Final",
			Attachment:         "LHA_K3LH_2026.pdf",
			Findings: []models.AuditReportFinding{
				{Title: "Inspeksi berkala sistem pemadam kebakaran hidran belum 100% terlaksana", Category: "Very Significant", Action: "Jadwalkan pemeliharaan hidran"},
				{Title: "Sertifikasi K3LH teknisi pemeliharaan pembangkit belum di-renew", Category: "Significant", Action: "Daftarkan pelatihan sertifikasi"},
				{Title: "APBD K3 belum memadai untuk instalasi area berisiko tinggi", Category: "Quite Significant", Action: "Pengadaan APD tambahan"},
			},
		},
	}
	for i := range seeds {
		var existing models.AuditResultReport
		err := db.Where("report_number = ?", seeds[i].ReportNumber).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&seeds[i]).Error; err != nil {
				return err
			}
		} else {
			db.Model(&existing).Updates(map[string]interface{}{
				"report_date":    seeds[i].ReportDate,
				"findings":       seeds[i].Findings,
				"findings_count": len(seeds[i].Findings),
				"report_title":   seeds[i].ReportTitle,
				"audit_object":   seeds[i].AuditObject,
				"department":     seeds[i].Department,
				"audit_period":   seeds[i].AuditPeriod,
			})
		}
	}
	return nil
}

func seedActionTakenReports(db *gorm.DB) error {
	seeds := []models.ActionTakenReport{
		{
			AuditRef:            "ST-001/SKAI/2026",
			Title:               "Pemasangan SMTP Alert Backup ERP",
			Department:          "IT",
			AuditObject:         "IT Infrastructure",
			FindingCategory:     "Technology & Access",
			Condition:           "Kegagalan backup database tidak memicu notifikasi otomatis.",
			Criteria:            "SOP IT Recovery mensyaratkan notifikasi insiden langsung terkirim ke tim sysadmin.",
			Recommendation:      "Konfigurasi SMTP server untuk mengirim log gagal.",
			PIC:                 "Rudi Hermawan",
			Deadline:            "2026-04-15",
			Status:              "COMPLETED",
			Attachment:          "Config_Alert_SMTP.pdf",
			ProgressDescription: "SMTP alert dikonfigurasi dan sukses diuji coba pada backup simulasi.",
		},
	}
	for i := range seeds {
		var existing models.ActionTakenReport
		err := db.Where("audit_ref = ? AND title = ?", seeds[i].AuditRef, seeds[i].Title).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&seeds[i]).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedExecutiveSummaries(db *gorm.DB) error {
	seeds := []models.ExecutiveSummary{
		{
			Quarter:             2,
			PeriodeBulan:        "April 2026",
			Tahun:               2026,
			NomorDokumen:        "021/LHA/01/KS IAD/2026",
			DokumenPath:         "Executive_Summary_021_LHA_2026.pdf",
			Status:              "Approved",
			Narrative:           "Executive Summary Individual untuk Laporan Hasil Audit Operasional Keuangan (021/LHA/01/KS IAD/2026). Audit dilakukan untuk mengevaluasi efektivitas ICOFR dan kepatuhan terhadap SOP pembayaran.",
			JumlahLaporan:       1,
			RisikoTinggi:        3,
			RisikoSedang:        2,
			RisikoRendah:        0,
			JumlahRekomendasi:   5,
			FollowUpTable:       `[{"unit":"Departemen Keuangan","open":2,"closed":3}]`,
			TopFindings:         `[{"finding":"Selisih pencatatan inventaris fisik vs buku besar","severity":"High"},{"finding":"Keterlambatan rekonsiliasi kas harian cabang utama","severity":"High"}]`,
			MatriksKompilasi:    `[{"nomor":"021/LHA/01","division":"Finance","unitKerja":"Departemen Keuangan","prosesBisnis":"ICOFR","judulTemuan":"Selisih pencatatan inventaris fisik vs buku besar","nilaiRisiko":"Tinggi","rekomendasi":"Lakukan rekonsiliasi harian dan alert SMTP","dueDate":"2026-05-15","picUnit":"Manager Keuangan","progres":60,"status":"In Progress","buktiTL":"BA_Rekonsiliasi.pdf"}]`,
			AkarMasalah:         "Kurangnya otomatisasi alarm kegagalan backup data dan kelalaian non-aktifkan akses user kasir.",
			Kesimpulan:          "Secara umum pengendalian internal departemen keuangan memadai dengan beberapa area peningkatan yang perlu segera ditindaklanjuti.",
			SignatureTempat:     "Jakarta",
			SignatureTanggal:    "2026-04-15",
			SignatureNamaKepala: "Zeta Ramadhani",
			SignatureNIK:        "NIK-100240",
		},
		{
			Quarter:             3,
			PeriodeBulan:        "September 2023",
			Tahun:               2023,
			NomorDokumen:        "020/LHA/01/KS IAD/2023",
			DokumenPath:         "Executive_Summary_020_LHA_2023.pdf",
			Status:              "Approved",
			Narrative:           "Executive Summary Individual untuk Audit Operasional Pengelolaan Pembangkitan UPDK Kepulauan Riau (020/LHA/01/KS IAD/2023). Audit mengevaluasi ketersediaan pembangkit (EAF/EFOR), K3LH, manajemen risiko, dan SCM.",
			JumlahLaporan:       1,
			RisikoTinggi:        2,
			RisikoSedang:        4,
			RisikoRendah:        2,
			JumlahRekomendasi:   8,
			FollowUpTable:       `[{"unit":"UPDK Kepulauan Riau","open":3,"closed":5}]`,
			TopFindings:         `[{"finding":"Pelaksanaan Overhaul ME+ PLTU TBK #1 Terjadi PE 6 Hari","severity":"High"},{"finding":"Pengelolaan Manajemen Risiko Belum Sepenuhnya Sesuai Masa Transisi HSH","severity":"High"}]`,
			MatriksKompilasi:    `[{"nomor":"020/LHA/01","division":"Operasi","unitKerja":"UPDK Kepulauan Riau","prosesBisnis":"O&M Pembangkit","judulTemuan":"Pelaksanaan Overhaul ME+ PLTU TBK #1 Terjadi PE 6 Hari","nilaiRisiko":"Tinggi","rekomendasi":"Penyusunan DMR & Sertifikasi Pemeliharaan","dueDate":"2023-11-30","picUnit":"Manager UPDK","progres":80,"status":"In Progress","buktiTL":"Laporan_Overhaul.pdf"}]`,
			AkarMasalah:         "Masa transisi holding sub-holding, keterbatasan sarana lab batubara, dan belum lengkapnya fitur Maximo WPC.",
			Kesimpulan:          "Tata kelola dan pengendalian internal berjalan baik dengan 1 Risk Management AoI dan 8 Internal Control AoI.",
			SignatureTempat:     "Tanjung Pinang",
			SignatureTanggal:    "2023-09-22",
			SignatureNamaKepala: "Tomy Afrilianto",
			SignatureNIK:        "NIK-100188",
		},
		{
			Quarter:             1,
			PeriodeBulan:        "Januari - Maret",
			Tahun:               2026,
			NomorDokumen:        "DOC-EXSUM-Q1-2026",
			DokumenPath:         "ExSum_Q1_2026.pdf",
			Status:              "Approved",
			Narrative:           "Ringkasan eksekutif ini menjelaskan pencapaian program kerja audit Triwulan I 2026.",
			JumlahLaporan:       3,
			RisikoTinggi:        1,
			RisikoSedang:        2,
			RisikoRendah:        5,
			JumlahRekomendasi:   12,
			FollowUpTable:       `[{"unit":"Finance","open":1,"closed":5}]`,
			TopFindings:         `[{"finding":"ERP Backup issue","severity":"Medium"}]`,
			MatriksKompilasi:    `[]`,
			AkarMasalah:         "Kurangnya monitoring sistem secara real-time.",
			Kesimpulan:          "Secara umum, tata kelola perusahaan berjalan dengan baik.",
			SignatureTempat:     "Jakarta",
			SignatureTanggal:    "2026-04-01",
			SignatureNamaKepala: "Head of SKAI",
			SignatureNIK:        "NIK-100239",
		},
	}
	for i := range seeds {
		var existing models.ExecutiveSummary
		err := db.Where("nomor_dokumen = ?", seeds[i].NomorDokumen).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&seeds[i]).Error; err != nil {
				return err
			}
		} else {
			db.Model(&existing).Updates(map[string]interface{}{
				"narrative":         seeds[i].Narrative,
				"periode_bulan":     seeds[i].PeriodeBulan,
				"tahun":             seeds[i].Tahun,
				"quarter":           seeds[i].Quarter,
				"jumlah_laporan":    seeds[i].JumlahLaporan,
				"risiko_tinggi":     seeds[i].RisikoTinggi,
				"risiko_sedang":     seeds[i].RisikoSedang,
				"risiko_rendah":     seeds[i].RisikoRendah,
				"jumlah_rekomendasi": seeds[i].JumlahRekomendasi,
			})
		}
	}
	return nil
}

func seedAnnualPlans(db *gorm.DB) error {
	seeds := []models.AuditAnnual{
		{
			Code:                 "PKAT-2026-001",
			Version:              "v1.0",
			Status:               "APPROVED",
			SelectedMonths:       []int{1, 2, 3},
			Quarters:             []string{"Q1"},
			AuditorCount:         3,
			DaysPerAuditor:       10,
			TotalMandays:         30,
			SupervisorID:         "SUP-101",
			SupervisorName:       "Budi Santoso",
			Year:                 2026,
			Notes:                "Annual plan focused on Core Financial and IT compliance controls.",
			AttachmentCategory:   "Plan",
			AttachmentUploadedBy: "Zeta Ramadhani",
			AttachmentUploadDate: "2026-01-05",
			IsActive:             true,
			Activities: []models.AnnualAuditActivity{
				{Name: "Financial Audit Q1", Category: "Assurance", Department: "Finance", RiskName: "Payment Overrides", RiskLevel: "High"},
			},
			Attachments: []models.AnnualAuditAttachment{
				{Name: "PKAT_2026_V1.pdf", Size: "2.4 MB", URL: "/uploads/pkat_2026.pdf"},
			},
		},
	}
	for i := range seeds {
		var existing models.AuditAnnual
		err := db.Where("code = ?", seeds[i].Code).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&seeds[i]).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedActivityPlans(db *gorm.DB) error {
	seeds := []models.ActivityPlan{
		{
			PlanTitle:       "Rencana Kegiatan Audit Keuangan Q1 2026",
			PlanYear:        "2026",
			PlanPeriodStart: "2026-01-01",
			PlanPeriodEnd:   "2026-03-31",
			Department:      "Finance",
			CreatedBy:       "Zeta Ramadhani",
			CreationDate:    "2025-12-15",
			Status:          "completed",
			Budget: models.PlanBudget{
				TotalEstimatedCost:   15000000,
				TotalAllocatedBudget: 15000000,
				BudgetNotes:          "Include transport and allowances",
			},
			Review: models.PlanReview{
				CreatorName:      "Zeta Ramadhani",
				CreatorPosition:  "Lead Auditor",
				ApproverName:     "Budi Santoso",
				ApproverPosition: "CAE",
				ApprovalDate:     "2025-12-20",
			},
			PlannedActivities: []models.PlannedActivity{
				{ID: "act-1", AuditName: "Audit Pengeluaran Kas", Auditee: "Kasir & Keuangan", Category: "ASSURANCE", RiskName: "Fraud Kas", RiskLevel: "High", Duration: 15, Priority: "P1", NumberOfAuditors: 2, EstimatedSchedule: "2026-01-15", BudgetEstimation: "10,000,000"},
				{ID: "act-2", AuditName: "Audit Rekonsiliasi Bank", Auditee: "Departemen Akuntansi", Category: "ASSURANCE", RiskName: "Selisih Rekonsiliasi", RiskLevel: "Medium", Duration: 10, Priority: "P2", NumberOfAuditors: 1, EstimatedSchedule: "2026-02-10", BudgetEstimation: "5,000,000"},
			},
			ResourceAuditors: []models.ResourceAuditor{
				{ID: "aud-1", Name: "Zeta Ramadhani", Position: "Lead Auditor", Competence: "Finance & Accounting", Availability: "100%"},
				{ID: "aud-2", Name: "Andi Firmansyah", Position: "Senior Auditor", Competence: "Treasury & Tax", Availability: "100%"},
			},
		},
		{
			PlanTitle:       "Rencana Kegiatan Audit Keamanan Sistem Informasi 2026",
			PlanYear:        "2026",
			PlanPeriodStart: "2026-04-01",
			PlanPeriodEnd:   "2026-06-30",
			Department:      "IT",
			CreatedBy:       "Andi Firmansyah",
			CreationDate:    "2026-01-10",
			Status:          "in_progress",
			Budget: models.PlanBudget{
				TotalEstimatedCost:   25000000,
				TotalAllocatedBudget: 25000000,
				BudgetNotes:          "External security tool licensing included",
			},
			Review: models.PlanReview{
				CreatorName:      "Andi Firmansyah",
				CreatorPosition:  "IT Auditor",
				ApproverName:     "Budi Santoso",
				ApproverPosition: "CAE",
				ApprovalDate:     "2026-01-15",
			},
			PlannedActivities: []models.PlannedActivity{
				{ID: "act-3", AuditName: "Audit Akses ERP & Database", Auditee: "IT Infrastructure", Category: "ASSURANCE", RiskName: "Unqualified Access", RiskLevel: "High", Duration: 20, Priority: "P1", NumberOfAuditors: 2, EstimatedSchedule: "2026-04-10", BudgetEstimation: "15,000,000"},
				{ID: "act-4", AuditName: "Audit Penetrasi & Vulnerability", Auditee: "IT Security Desk", Category: "CONSULTING", RiskName: "Cyber Vulnerability", RiskLevel: "High", Duration: 15, Priority: "P1", NumberOfAuditors: 2, EstimatedSchedule: "2026-05-15", BudgetEstimation: "10,000,000"},
			},
			ResourceAuditors: []models.ResourceAuditor{
				{ID: "aud-3", Name: "Andi Firmansyah", Position: "IT Auditor Specialist", Competence: "CISA / Cybersecurity", Availability: "100%"},
			},
		},
		{
			PlanTitle:       "Rencana Kegiatan Audit Operasional Branch & Logistik 2026",
			PlanYear:        "2026",
			PlanPeriodStart: "2026-07-01",
			PlanPeriodEnd:   "2026-09-30",
			Department:      "Operations",
			CreatedBy:       "Rina Wulandari",
			CreationDate:    "2026-02-01",
			Status:          "planned",
			Budget: models.PlanBudget{
				TotalEstimatedCost:   20000000,
				TotalAllocatedBudget: 20000000,
				BudgetNotes:          "Operational field visits",
			},
			Review: models.PlanReview{
				CreatorName:      "Rina Wulandari",
				CreatorPosition:  "Auditor",
				ApproverName:     "Budi Santoso",
				ApproverPosition: "CAE",
				ApprovalDate:     "2026-02-05",
			},
			PlannedActivities: []models.PlannedActivity{
				{ID: "act-5", AuditName: "Audit Inventaris Gudang & Logistik", Auditee: "Gudang Pusat", Category: "ASSURANCE", RiskName: "Loss Stock", RiskLevel: "Medium", Duration: 15, Priority: "P2", NumberOfAuditors: 2, EstimatedSchedule: "2026-07-15", BudgetEstimation: "20,000,000"},
			},
			ResourceAuditors: []models.ResourceAuditor{
				{ID: "aud-4", Name: "Rina Wulandari", Position: "Operational Auditor", Competence: "Supply Chain & Operations", Availability: "100%"},
			},
		},
	}
	for i := range seeds {
		var existing models.ActivityPlan
		err := db.Where("plan_title = ?", seeds[i].PlanTitle).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&seeds[i]).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedUploadedPlanDocuments(db *gorm.DB) error {
	seeds := []models.UploadedPlanDocument{
		{
			Title:       "Rencana Kerja Audit Tahunan (PKAT) 2026 - Final Approved",
			Description: "Dokumen resmi PKAT 2026 yang telah disetujui Komite Audit dan Direksi.",
			FileName:    "PKAT_2026_Final_Approved.pdf",
			FilePath:    "uploads/PKAT_2026_Final_Approved.pdf",
			FileSize:    2450000,
			FileType:    "application/pdf",
		},
		{
			Title:       "Pedoman Operasional Rencana Kegiatan Audit 2026",
			Description: "Petunjuk teknis dan metodologi pelaksanaan rencana kegiatan audit per divisi.",
			FileName:    "Pedoman_Operasional_Audit_2026.docx",
			FilePath:    "uploads/Pedoman_Operasional_Audit_2026.docx",
			FileSize:    1850000,
			FileType:    "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		},
		{
			Title:       "Matriks Alokasi Sumber Daya & Anggaran Audit Q1-Q4 2026",
			Description: "Rincian alokasi auditor, durasi mandays, dan rencana anggaran kegiatan audit tahun 2026.",
			FileName:    "Matriks_Alokasi_Anggaran_Audit_2026.pdf",
			FilePath:    "uploads/Matriks_Alokasi_Anggaran_Audit_2026.pdf",
			FileSize:    1200000,
			FileType:    "application/pdf",
		},
	}
	for i := range seeds {
		var existing models.UploadedPlanDocument
		err := db.Where("title = ?", seeds[i].Title).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&seeds[i]).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func seedImportedWorkingPapers(db *gorm.DB) error {
	seeds := []models.ImportedWorkingPaper{
		{
			Title:       "Kertas Kerja Audit Operasional & Logistik Q1 2026",
			Description: "Pengujian siklus pengadaan persediaan dan verifikasi stok gudang pusat.",
			FileName:    "KK_Audit_Operasional_Logistik_Q1_2026.xlsx",
			FilePath:    "uploads/KK_Audit_Operasional_Logistik_Q1_2026.xlsx",
			FileSize:    3150000,
			FileType:    "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		},
		{
			Title:       "Working Paper Evaluation IT General Controls (ITGC) 2026",
			Description: "Dokumen pengujian kontrol akses user, change management, dan sistem backup ERP.",
			FileName:    "WP_ITGC_Evaluation_2026.docx",
			FilePath:    "uploads/WP_ITGC_Evaluation_2026.docx",
			FileSize:    1920000,
			FileType:    "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		},
		{
			Title:       "Kertas Kerja Verifikasi Laporan Keuangan & Pengeluaran Kas 2025",
			Description: "Hasil pengujian sampel voucher pengeluaran di atas Rp 50 juta dan pembuktian mutasi bank.",
			FileName:    "KK_Financial_Statement_Review_2025.pdf",
			FilePath:    "uploads/KK_Financial_Statement_Review_2025.pdf",
			FileSize:    4200000,
			FileType:    "application/pdf",
		},
	}
	for i := range seeds {
		var existing models.ImportedWorkingPaper
		err := db.Where("title = ?", seeds[i].Title).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.Create(&seeds[i]).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
