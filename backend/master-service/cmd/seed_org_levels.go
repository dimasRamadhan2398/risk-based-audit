package main

import (
	"fmt"

	"master-service/pkg/database"
	"master-service/pkg/logger"
	"master-service/pkg/orgmode"
	levelRepo "master-service/repositories/org_level"
	orgLevelSvc "master-service/services/org_level"

	"github.com/spf13/cobra"
)

var (
	orgPreset string
	orgLevels string
)

// seedOrgLevelsCmd seeds the org levels of a NEW flexible tenant
// (docs/org-structure-model-plan.md section 7, step 4). It runs even with
// --empty-data, because levels are configuration. It is a no-op when live
// levels exist, so admin edits are never overwritten. It never runs DDL:
// `master migrate` with ORG_STRUCTURE_MODE=flexible must have run first.
var seedOrgLevelsCmd = &cobra.Command{
	Use:   "seed-org-levels",
	Short: "Seed the organisation levels of a flexible tenant (idempotent)",
	Long: `Seed org_levels from a preset: simple (default), corporate, bank or custom.
--levels "A,B,C" overrides the stored names (required for custom).
Requires ORG_STRUCTURE_MODE=flexible and a prior "master migrate".`,
	RunE: runSeedOrgLevels,
}

func init() {
	seedOrgLevelsCmd.Flags().StringVar(&orgPreset, "preset", orgLevelSvc.PresetSimple, "simple | corporate | bank | custom")
	seedOrgLevelsCmd.Flags().StringVar(&orgLevels, "levels", "", `level names in order, e.g. "Direktorat,Divisi,Departemen"`)
	rootCmd.AddCommand(seedOrgLevelsCmd)
}

func runSeedOrgLevels(cmd *cobra.Command, args []string) error {
	if err := initConfig(); err != nil {
		return err
	}
	if err := initLogger(); err != nil {
		return err
	}
	defer logger.Sync()

	if !cfg.OrgStructure.FlexibleRequested() {
		return fmt.Errorf("refusing to seed org levels: ORG_STRUCTURE_MODE is not flexible (this is a legacy tenant)")
	}
	// Validate the preset before touching the database.
	names := orgLevelSvc.ParseLevelNames(orgLevels)
	if _, err := orgLevelSvc.BuildPreset(orgPreset, names); err != nil {
		return err
	}

	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	if !db.Migrator().HasTable(orgmode.LevelsTable) {
		return fmt.Errorf("table %s does not exist: run `master migrate` with ORG_STRUCTURE_MODE=flexible first", orgmode.LevelsTable)
	}
	res, err := orgLevelSvc.SeedPreset(levelRepo.New(db), orgPreset, names)
	if err != nil {
		return fmt.Errorf("seed org levels: %w", err)
	}
	if res.Existing > 0 {
		fmt.Printf("org_levels already has %d live level(s): nothing changed\n", res.Existing)
		return nil
	}
	fmt.Printf("Seeded %d org level(s) from preset %q\n", res.Created, orgPreset)
	return nil
}
