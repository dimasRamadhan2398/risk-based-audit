package main

import (
	"context"
	"fmt"
	"time"

	ctrlAnnualPlan "audit-service/controllers/annual_plan"
	"audit-service/models"
	"audit-service/pkg/database"
	"audit-service/pkg/logger"
	atrsvc "audit-service/services/action_taken_report"
	"audit-service/services/annual_plan_sync"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var migrateDryRun bool

func init() {
	// migrate up command - Create/update tables using GORM AutoMigrate
	upCmd := &cobra.Command{
		Use:   "up",
		Short: "Run database migrations (GORM AutoMigrate)",
		Long: `Runs GORM AutoMigrate to create or update audit-service database tables.
Creates the following tables:
- audit_charters
- audit_mandates
- audit_assignments
- audit_annuals
- activity_plans

Usage: audit migrate up`,
		RunE: runMigrateUp,
	}
	upCmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "Show what would be migrated")

	// migrate down command - Drop tables
	downCmd := &cobra.Command{
		Use:   "down",
		Short: "Rollback migrations (drop tables)",
		Long: `Drops all audit-service tables.
WARNING: This will delete all data!

Usage: audit migrate down`,
		RunE: runMigrateDown,
	}

	// migrate status command
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		Long:  `Shows which audit tables exist in the database.`,
		RunE:  runMigrateStatus,
	}

	// Parent migrate command
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration commands (GORM)",
		Long: `Audit Service database migration commands using GORM AutoMigrate.

Examples:
  audit migrate up           # Run all migrations
  audit migrate up --dry-run # Show what would be migrated
  audit migrate down         # Drop all tables
  audit migrate status       # Show current status`,
	}
	// migrate backfill-atr - create ATRs for LHAs approved before ATRs were
	// created automatically
	backfillCmd := &cobra.Command{
		Use:   "backfill-atr",
		Short: "Create the missing action taken reports of approved LHAs",
		Long: `Creates one action taken report (ATR) per finding of every approved
(status Final / Approved / Published, any casing), non-deleted audit result
report that does not have one yet. Also gives LHA findings ids if needed.

"migrate up" (and "serve" on startup) already run this step; the command is
for checking with --dry-run, which reports the counts and rolls back.
Idempotent: a second run creates nothing.

Usage: audit migrate backfill-atr [--dry-run]`,
		RunE: runBackfillATR,
	}
	backfillCmd.Flags().BoolVar(&backfillDryRun, "dry-run", false, "Report what would be created and roll back")

	// backfill-annual-snapshots - fill empty PKAT department, risk name and risk
	// level snapshots from the local lookup
	backfillSnapshotsCmd := &cobra.Command{
		Use:   "backfill-annual-snapshots",
		Short: "Fill the empty department, risk name and risk level of annual plan activities",
		Long: `Fills the empty display snapshots (department, riskName, riskLevel) of the
activities of every live annual plan (PKAT) from the local lookup that the
Department and Audit Universe events fill (table annual_snapshot_refs).

A field that already has a value is never changed. Idempotent: a second run
changes nothing.

The lookup only knows what the events delivered. Its rows are seeded by the
resync endpoints of master-service (POST /api/v1/departments/events/resync) and
risk-service (POST /api/v1/audit-universe/events/resync), which re-publish the
current state. Run those first, wait for the sync consumer to catch up, then run
this command. Use --dry-run to report the counts and roll back.

Usage: audit migrate backfill-annual-snapshots [--dry-run]`,
		RunE: runBackfillAnnualSnapshots,
	}
	backfillSnapshotsCmd.Flags().BoolVar(&backfillSnapshotsDryRun, "dry-run", false, "Report what would be filled and roll back")

	migrateCmd.AddCommand(upCmd, downCmd, statusCmd, backfillCmd, backfillSnapshotsCmd)

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

	// Define audit models to migrate
	auditModels := []interface{}{
		&models.AuditCharter{},
		&models.AuditGuideline{},
		&models.AuditSop{},
		&models.AuditMandate{},
		&models.AuditAssignment{},
		&models.AuditAnnual{},
		&models.AnnualSnapshotRef{},
		&models.ActivityPlan{},
		&models.AuditActivity{},
		&models.ActivityCodeSequence{},
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
		&models.ActionTakenReportEvidence{},
		&models.ImportedWorkingPaper{},
		&models.UploadedStrategicPlanDocument{},
		&models.UploadedPlanDocument{},
		&models.UploadedAnnualPlan{},
		&models.UploadedAssignmentLetter{},
		&models.UploadedAuditResultReport{},
		&models.UploadedExecutiveSummary{},
		&models.UploadedExecutiveSummaryReport{},
		&models.UploadedConsultingDocument{},
		&models.UploadedPerformanceReport{},
		&models.KPIAchievement{},
		&models.WorkPlanRealization{},
		&models.AuditeeSurvey{},
		&models.AuditCompletionSnapshot{},
	}

	if migrateDryRun {
		logger.Info("DRY RUN - Tables that would be created/updated:")
		for _, m := range auditModels {
			logger.Info(fmt.Sprintf("  ✓ %T", m))
		}
		return nil
	}

	// Run migrations
	logger.Info("Running GORM AutoMigrate...")

	for _, m := range auditModels {
		logger.Info(fmt.Sprintf("  Migrating model: %T", m))
		if err := db.AutoMigrate(m); err != nil {
			logger.Fatal(fmt.Sprintf("Failed to migrate model %T", m), logger.LogField("error", err))
			return err
		}
	}

	// Ensure assignment_letters letter_number unique index is a partial index on active records only
	db.Exec(`
		DROP INDEX IF EXISTS idx_assignment_letters_letter_number;
		CREATE UNIQUE INDEX IF NOT EXISTS idx_assignment_letters_letter_number ON assignment_letters (letter_number) WHERE deleted_at IS NULL;
	`)

	if err := normalizeAnnualPlanStatuses(db); err != nil {
		logger.Fatal("Failed to normalize annual plan statuses", logger.LogField("error", err))
		return err
	}

	if err := runATRDataMigrations(db); err != nil {
		logger.Fatal("Failed to run action taken report data migrations", logger.LogField("error", err))
		return err
	}

	normalizeExecutiveSummaryRiskCounts(db)

	logger.Info("Database migration completed successfully!")
	return nil
}

// normalizeAnnualPlanStatuses maps legacy audit_annuals.status values to the
// canonical vocabulary (models/annual_plan_status.go). Idempotent. serve runs it
// on startup as well, because tenant stacks start with "audit serve" only.
func normalizeAnnualPlanStatuses(db *gorm.DB) error {
	n, unknown, err := ctrlAnnualPlan.NormalizeStoredStatuses(db)
	if err != nil {
		return fmt.Errorf("normalize annual plan statuses: %w", err)
	}
	if n > 0 {
		logger.Info("Annual plan statuses normalized", logger.LogField("rows", n))
	}
	if len(unknown) > 0 {
		logger.Warn("Annual plans with status outside the vocabulary; they cannot change status until fixed",
			logger.LogField("statuses", unknown))
	}
	return nil
}

// normalizeExecutiveSummaryRiskCounts copies the legacy executive summary risk
// counts into the five-level columns where those are all 0
// (models.NormalizeExecutiveSummaryRiskCounts). Idempotent; serve runs it on
// startup as well. A failure is logged, not fatal: the rows keep their legacy
// counts and the next start retries.
func normalizeExecutiveSummaryRiskCounts(db *gorm.DB) {
	n, err := models.NormalizeExecutiveSummaryRiskCounts(db)
	if err != nil {
		logger.Error("Executive summary risk count normalization failed", logger.LogField("error", err))
		return
	}
	if n > 0 {
		logger.Info("Executive summary risk counts copied to the five levels", logger.LogField("rows", n))
	}
}

// runATRDataMigrations runs the data steps of the LHA-bound ATR model after
// AutoMigrate. All are idempotent.
//  1. Every LHA finding gets a stable id (audit_result_reports.findings[].id).
//  2. The old assignment-letter based action_taken_reports columns are
//     converted (title -> finding_title, pic -> pic_name, deadline -> due_date,
//     progress_description -> action_plan; rows linked to an LHA finding when
//     their letter's LHA has a finding with the same title) and dropped.
//  3. Every approved LHA gets its missing ATRs (one per finding), as if it
//     had been approved now. "audit migrate backfill-atr --dry-run" shows
//     what this step would create without writing.
func runATRDataMigrations(db *gorm.DB) error {
	n, err := atrsvc.BackfillFindingIDs(db)
	if err != nil {
		return fmt.Errorf("backfill LHA finding ids: %w", err)
	}
	logger.Info("LHA finding ids backfilled", logger.LogField("reports_updated", n))

	res, err := atrsvc.MigrateLegacyColumns(db)
	if err != nil {
		return fmt.Errorf("migrate legacy action_taken_reports columns: %w", err)
	}
	if len(res.DroppedColumns) > 0 {
		logger.Info("Legacy action_taken_reports columns migrated",
			logger.LogField("rows", res.Rows),
			logger.LogField("linked_to_lha_finding", res.Linked),
			logger.LogField("unlinked", res.Unlinked),
			logger.LogField("dropped_columns", res.DroppedColumns))
	}

	bf, err := atrsvc.BackfillApprovedReports(db, time.Now())
	if err != nil {
		return fmt.Errorf("create action taken reports of approved LHAs: %w", err)
	}
	logger.Info("Action taken reports of approved LHAs backfilled",
		logger.LogField("approved_reports", bf.ApprovedReports),
		logger.LogField("atrs_created", bf.CreatedATRs),
		logger.LogField("atrs_cancelled_finding_removed", bf.CancelledATRs))
	return nil
}

// migrateATR brings the action taken report tables and data up to date:
// AutoMigrate of the two ATR tables, then runATRDataMigrations. serve runs it
// on startup because tenant stacks start with "audit serve" only; without it
// the old table layout would make the ATR list fail and, since approving an
// LHA creates ATRs in the same transaction, LHA saves too.
func migrateATR(db *gorm.DB) error {
	if err := db.AutoMigrate(&models.ActionTakenReport{}, &models.ActionTakenReportEvidence{}); err != nil {
		return fmt.Errorf("auto-migrate action taken report tables: %w", err)
	}
	return runATRDataMigrations(db)
}

var backfillDryRun bool

var backfillSnapshotsDryRun bool

// runBackfillATR creates the missing ATRs of every approved LHA
func runBackfillATR(cmd *cobra.Command, args []string) error {
	initLogger()
	defer logger.Sync()

	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", logger.LogField("error", err))
		return err
	}

	errDryRun := fmt.Errorf("dry run")
	var res atrsvc.BackfillResult
	err = db.Transaction(func(tx *gorm.DB) error {
		var err error
		res, err = atrsvc.BackfillApprovedReports(tx, time.Now())
		if err != nil {
			return err
		}
		if backfillDryRun {
			return errDryRun
		}
		return nil
	})
	if err != nil && err != errDryRun {
		logger.Error("ATR backfill failed", logger.LogField("error", err))
		return err
	}
	msg := "ATR backfill completed"
	if backfillDryRun {
		msg = "ATR backfill DRY RUN (rolled back, nothing written)"
	}
	logger.Info(msg,
		logger.LogField("reports_given_finding_ids", res.FindingIDReports),
		logger.LogField("approved_reports", res.ApprovedReports),
		logger.LogField("atrs_created", res.CreatedATRs),
		logger.LogField("atrs_cancelled_finding_removed", res.CancelledATRs))
	return nil
}

func runBackfillAnnualSnapshots(cmd *cobra.Command, args []string) error {
	initLogger()
	defer logger.Sync()

	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", logger.LogField("error", err))
		return err
	}

	errDryRun := fmt.Errorf("dry run")
	var res annual_plan_sync.BackfillResult
	err = db.Transaction(func(tx *gorm.DB) error {
		var err error
		res, err = annual_plan_sync.BackfillSnapshots(context.Background(), tx)
		if err != nil {
			return err
		}
		if backfillSnapshotsDryRun {
			return errDryRun
		}
		return nil
	})
	if err != nil && err != errDryRun {
		logger.Error("Annual snapshot backfill failed", logger.LogField("error", err))
		return err
	}
	msg := "Annual snapshot backfill completed"
	if backfillSnapshotsDryRun {
		msg = "Annual snapshot backfill DRY RUN (rolled back, nothing written)"
	}
	logger.Info(msg,
		logger.LogField("plans_changed", res.PlansChanged),
		logger.LogField("fields_filled", res.FieldsFilled))
	return nil
}

func runMigrateDown(cmd *cobra.Command, args []string) error {
	initLogger()
	defer logger.Sync()

	logger.Warn("Starting database migration (DOWN)")
	logger.Warn("WARNING: This will drop ALL audit tables and delete ALL data!")

	// Connect to database
	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", logger.LogField("error", err))
		return err
	}

	// Drop tables in reverse order of dependencies
	tablesToDrop := []string{
		"imported_working_papers",
		"action_taken_report_evidences",
		"action_taken_reports",
		"auditee_surveys",
		"audit_completion_snapshots",
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
		"annual_snapshot_refs",
		"audit_annuals",
		"audit_assignments",
		"audit_mandates",
		"audit_sops",
		"audit_guidelines",
		"audit_charters",
	}

	logger.Warn("Dropping tables...")
	for _, table := range tablesToDrop {
		if db.Migrator().HasTable(table) {
			logger.Info(fmt.Sprintf("  Dropping: %s", table))
			if err := db.Migrator().DropTable(table); err != nil {
				logger.Error(fmt.Sprintf("Failed to drop %s", table), logger.LogField("error", err))
			}
		} else {
			logger.Info(fmt.Sprintf("Skipping (not exists): %s", table))
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

	// Expected audit tables
	expectedTables := []string{
		"imported_working_papers",
		"audit_guidelines",
		"audit_sops",
		"audit_charters",
		"audit_mandates",
		"audit_assignments",
		"audit_annuals",
		"activity_plans",
		"audit_activities",
		"strategic_plans",
		"assignment_letters",
		"audit_executions",
		"fieldwork_interviews",
		"fieldwork_observations",
		"fieldwork_documents",
		"fieldwork_samples",
		"fieldwork_test_controls",
		"working_papers",
		"working_paper_headers",
		"working_paper_risks",
		"working_paper_samples",
		"working_paper_causes",
		"working_paper_plans",
		"audit_result_reports",
		"action_taken_reports",
		"action_taken_report_evidences",
		"kpi_achievements",
		"work_plan_realizations",
		"auditee_surveys",
		"audit_completion_snapshots",
	}

	logger.Info(fmt.Sprintf("\nDatabase: %s", cfg.Database.Name))
	logger.Info("-------------------------------------------")
	logger.Info("Audit Tables                  | Status")
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
