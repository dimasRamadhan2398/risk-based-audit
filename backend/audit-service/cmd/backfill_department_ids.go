package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"audit-service/models"
	"audit-service/pkg/database"
	"audit-service/pkg/logger"
	"audit-service/pkg/masterclient"
	"audit-service/pkg/servicetoken"
	"audit-service/pkg/serviceurl"
	"audit-service/services/department_link"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var backfillDeptIDsApply bool

// runBackfillDepartmentIDs is "audit migrate backfill-department-ids": a dry
// run unless --apply. It reads the master departments first (nothing is
// written if that fails), then writes only working_unit_department_id /
// department_id with UpdateColumn (no hooks, updated_at and the text
// untouched), re-checking IS NULL, in one transaction.
func runBackfillDepartmentIDs(_ *cobra.Command, _ []string) error {
	initLogger()
	defer logger.Sync()

	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", logger.LogField("error", err))
		return err
	}
	authorization, source, err := servicetoken.Authorization(os.Getenv(servicetoken.OverrideEnv), cfg.JWT.Secret, time.Now())
	if err != nil {
		logger.Error("Department id backfill not started: no token for master-service", logger.LogField("error", err.Error()))
		return err
	}
	master := masterclient.New(serviceurl.Master(), &http.Client{Timeout: sourceCallTimeout})
	logger.Info("Reading master departments", logger.LogField("master_service_url", master.BaseURL()), logger.LogField("auth_source", source))
	depts, err := master.ListDepartments(context.Background(), authorization)
	if err != nil {
		logger.Error("Department id backfill failed reading master departments; nothing written", logger.LogField("error", err.Error()))
		return err
	}

	var letters []models.AssignmentLetter
	if err := db.Select("id", "letter_number", "working_unit", "working_unit_department_id").Find(&letters).Error; err != nil {
		return fmt.Errorf("load assignment letters: %w", err)
	}
	var reports []models.AuditResultReport
	if err := db.Select("id", "report_number", "assignment_letter_id", "department", "department_id").
		Where("department_id IS NULL").Find(&reports).Error; err != nil {
		return fmt.Errorf("load audit result reports: %w", err)
	}

	links, misses := department_link.Plan(letters, reports, depts)
	for _, l := range links {
		fmt.Printf("LINK  %-22s %-36s %-24q -> %s (via %s)\n", l.Table, l.ID, l.Ref+" "+l.FromValue, l.DepartmentID, l.Via)
	}
	for _, m := range misses {
		fmt.Printf("SKIP  %-22s %-36s %-24q %s\n", m.Table, m.ID, m.Ref+" "+m.FromValue, m.Reason)
	}
	fmt.Printf("\n%d master departments, %d to link, %d left unlinked\n", len(depts), len(links), len(misses))

	if !backfillDeptIDsApply {
		fmt.Println("DRY RUN: nothing written. Re-run with --apply to write the links.")
		return nil
	}
	var updated int64
	err = db.Transaction(func(tx *gorm.DB) error {
		for _, l := range links {
			var res *gorm.DB
			switch l.Table {
			case "assignment_letters":
				res = tx.Model(&models.AssignmentLetter{}).
					Where("id = ? AND working_unit_department_id IS NULL", l.ID).
					UpdateColumn("working_unit_department_id", l.DepartmentID)
			default:
				res = tx.Model(&models.AuditResultReport{}).
					Where("id = ? AND department_id IS NULL", l.ID).
					UpdateColumn("department_id", l.DepartmentID)
			}
			if res.Error != nil {
				return fmt.Errorf("update %s %s: %w", l.Table, l.ID, res.Error)
			}
			updated += res.RowsAffected
		}
		return nil
	})
	if err != nil {
		logger.Error("Department id backfill failed; nothing written", logger.LogField("error", err.Error()))
		return err
	}
	fmt.Printf("Applied: %d rows updated\n", updated)
	return nil
}
