package main

import (
	"context"
	"fmt"
	"time"

	"risk-service/models"
	"risk-service/pkg/database"
	"risk-service/pkg/logger"
	"risk-service/pkg/masterclient"
	"risk-service/services"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

// backfill-locations links risk_profile rows to the Location master.
//
// Rows written before risk_profile.location_id existed only carry their branch
// as a legacy sentinel UUID in department_id (or as free text in
// location_name). Reads already resolve those at request time, so this is
// optional: it makes the link explicit in the data.
//
// Safe by construction:
//   - dry run unless --apply is given; the dry run only SELECTs;
//   - only rows with location_id IS NULL are touched, and the UPDATE re-checks
//     that condition, so re-running is a no-op;
//   - only location_id and location_name are written (UpdateColumns: no hooks,
//     updated_at untouched); department_id is left as is;
//   - names that match no registered location are reported and left alone.
//
// Usage inside the risk-service container of the stack to backfill:
//
//	./risk backfill-locations           # dry run, prints the plan
//	./risk backfill-locations --apply   # writes it, in one transaction
var backfillLocationsCmd = &cobra.Command{
	Use:   "backfill-locations",
	Short: "Link legacy risk profiles to Location master rows (dry run unless --apply)",
	RunE:  runBackfillLocations,
}

var backfillApply bool

func init() {
	backfillLocationsCmd.Flags().BoolVar(&backfillApply, "apply", false, "write the changes (default is a dry run)")
	rootCmd.AddCommand(backfillLocationsCmd)
}

func runBackfillLocations(cmd *cobra.Command, args []string) error {
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

	client := masterclient.NewClient(
		cfg.MasterService.BaseURL(),
		time.Duration(cfg.MasterService.TimeoutSeconds)*time.Second,
		time.Duration(cfg.MasterService.CacheTTLSeconds)*time.Second,
	)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	locs, err := client.ListLocations(ctx)
	if err != nil {
		// Never fall back to built-in names: abort instead.
		return fmt.Errorf("read Location master from %s: %w", client.BaseURL(), err)
	}

	var profiles []models.RiskProfile
	if err := db.Where("location_id IS NULL").Find(&profiles).Error; err != nil {
		return fmt.Errorf("load risk profiles: %w", err)
	}

	plan, misses := services.PlanLocationBackfill(profiles, locs)

	fmt.Printf("Database: %s  Location master: %s (%d locations)\n", cfg.Database.Name, client.BaseURL(), len(locs))
	fmt.Printf("Profiles without location_id: %d  to link: %d  unmatched: %d\n", len(profiles), len(plan), len(misses))
	for _, p := range plan {
		fmt.Printf("  link  %s  %s=%q -> %s (%s)\n", p.ProfileID, p.Source, p.FromValue, p.LocationName, p.LocationID)
	}
	for _, m := range misses {
		fmt.Printf("  skip  %s  %s\n", m.ProfileID, m.Reason)
	}

	if !backfillApply {
		fmt.Println("Dry run: nothing written. Re-run with --apply to write.")
		return nil
	}
	if len(plan) == 0 {
		fmt.Println("Nothing to write.")
		return nil
	}

	var updated int64
	err = db.Transaction(func(tx *gorm.DB) error {
		for _, p := range plan {
			res := tx.Model(&models.RiskProfile{}).
				Where("id = ? AND location_id IS NULL", p.ProfileID).
				UpdateColumns(map[string]interface{}{
					"location_id":   p.LocationID,
					"location_name": p.LocationName,
				})
			if res.Error != nil {
				return fmt.Errorf("update profile %s: %w", p.ProfileID, res.Error)
			}
			updated += res.RowsAffected
		}
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("Updated %d risk_profile rows.\n", updated)
	return nil
}
