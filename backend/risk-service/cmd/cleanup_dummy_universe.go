package main

import (
	"fmt"
	"log"

	"risk-service/models"
	"risk-service/pkg/database"
	"risk-service/pkg/logger"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

// legacyDummyNames are the sample audit-universe nodes written by the early demo
// seeders. Removing them is a one-time, operator-run clean-up. It must never run
// on container start: the boot path has no legitimate reason to delete rows.
var legacyDummyNames = []string{
	"PT A", "PT B", "PT C", "Jakarta", "Surabaya",
	"Product/Service A", "Product/Service B", "Product/Service C",
}

// Usage:
//
//	./risk cleanup-dummy-universe           # dry run: counts and logs, writes nothing
//	./risk cleanup-dummy-universe --apply   # deletes, in one transaction
var cleanupDummyUniverseCmd = &cobra.Command{
	Use:   "cleanup-dummy-universe",
	Short: "Remove legacy dummy audit-universe nodes (dry run unless --apply)",
	Long: `Removes the legacy demo nodes (PT A/B/C, Jakarta, Surabaya, Product/Service A-C)
and their direct children from standard_audit_universe, plus the matching
corporate_audit_universe rows and their year/score rows. Master-synced and
user-added nodes are not on the legacy list and are left alone.

Without --apply the command only counts and logs what it would delete.`,
	RunE: runCleanupDummyUniverse,
}

var cleanupDummyUniverseApply bool

func init() {
	cleanupDummyUniverseCmd.Flags().BoolVar(&cleanupDummyUniverseApply, "apply", false, "delete the rows (default is a dry run)")
	rootCmd.AddCommand(cleanupDummyUniverseCmd)
}

func runCleanupDummyUniverse(cmd *cobra.Command, args []string) error {
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

	return CleanupDummyAuditUniverseSubEntities(db, cleanupDummyUniverseApply)
}

// universeNode is the shape both audit-universe tables share for planning.
type universeNode struct {
	ID       uuid.UUID
	Name     string
	ParentID *uuid.UUID
}

// dummyPlan lists the rows a clean-up removes from one audit-universe table.
type dummyPlan struct {
	Named    []universeNode // nodes whose name is on the legacy list
	Children []universeNode // direct children of a named node (one level only)
}

// planLegacyDummyCleanup decides which nodes the clean-up removes. It is a pure
// function so the decision can be tested without a database.
func planLegacyDummyCleanup(nodes []universeNode, names []string) dummyPlan {
	legacy := make(map[string]bool, len(names))
	for _, n := range names {
		legacy[n] = true
	}

	named := make(map[uuid.UUID]bool)
	var plan dummyPlan
	for _, n := range nodes {
		if legacy[n.Name] {
			named[n.ID] = true
			plan.Named = append(plan.Named, n)
		}
	}
	for _, n := range nodes {
		if n.ParentID != nil && named[*n.ParentID] && !named[n.ID] {
			plan.Children = append(plan.Children, n)
		}
	}
	return plan
}

func idStrings(nodes []universeNode) []string {
	out := make([]string, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, n.ID.String())
	}
	return out
}

// CleanupDummyAuditUniverseSubEntities removes legacy dummy nodes. With apply
// false it only counts and logs; nothing is written. It is not called on boot.
func CleanupDummyAuditUniverseSubEntities(db *gorm.DB, apply bool) error {
	var stdRows []models.StandardAuditUniverse
	if err := db.Select("id", "name", "parent_id").Find(&stdRows).Error; err != nil {
		return fmt.Errorf("load standard audit universe: %w", err)
	}
	stdNodes := make([]universeNode, 0, len(stdRows))
	stdNames := make(map[uuid.UUID]string, len(stdRows))
	for _, r := range stdRows {
		stdNodes = append(stdNodes, universeNode{ID: r.ID, Name: r.Name, ParentID: r.ParentID})
		stdNames[r.ID] = r.Name
	}
	stdPlan := planLegacyDummyCleanup(stdNodes, legacyDummyNames)

	var corpRows []models.CorporateAuditUniverse
	if err := db.Select("id", "name", "parent_id").Find(&corpRows).Error; err != nil {
		return fmt.Errorf("load corporate audit universe: %w", err)
	}
	corpNodes := make([]universeNode, 0, len(corpRows))
	for _, r := range corpRows {
		corpNodes = append(corpNodes, universeNode{ID: r.ID, Name: r.Name, ParentID: r.ParentID})
	}
	corpPlan := planLegacyDummyCleanup(corpNodes, legacyDummyNames)

	mode := "DRY RUN (no changes)"
	if apply {
		mode = "APPLY"
	}
	log.Printf("cleanup-dummy-universe: %s", mode)

	for _, n := range stdPlan.Named {
		log.Printf("  standard_audit_universe  named    %s  id=%s", n.Name, n.ID)
	}
	for _, n := range stdPlan.Children {
		log.Printf("  standard_audit_universe  child    %s (parent %s)  id=%s", n.Name, stdNames[*n.ParentID], n.ID)
	}
	for _, n := range corpPlan.Named {
		log.Printf("  corporate_audit_universe named    %s  id=%s", n.Name, n.ID)
	}
	for _, n := range corpPlan.Children {
		log.Printf("  corporate_audit_universe child    %s  id=%s", n.Name, n.ID)
	}

	namedCorpIDs := idStrings(corpPlan.Named)
	var yearCount, scoreCount int64
	if len(namedCorpIDs) > 0 {
		if err := db.Model(&models.AuditUniverseYear{}).
			Where("corporate_audit_universe_id IN ?", namedCorpIDs).Count(&yearCount).Error; err != nil {
			return err
		}
		if err := db.Model(&models.AuditUniverseRiskScore{}).
			Where("audit_universe_year_id IN (?)",
				db.Model(&models.AuditUniverseYear{}).Select("id").Where("corporate_audit_universe_id IN ?", namedCorpIDs),
			).Count(&scoreCount).Error; err != nil {
			return err
		}
	}
	log.Printf("cleanup-dummy-universe: standard_audit_universe=%d corporate_audit_universe=%d audit_universe_years=%d audit_universe_risk_scores=%d",
		len(stdPlan.Named)+len(stdPlan.Children),
		len(corpPlan.Named)+len(corpPlan.Children),
		yearCount, scoreCount)

	if !apply {
		log.Printf("cleanup-dummy-universe: dry run, re-run with --apply to delete")
		return nil
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if len(namedCorpIDs) > 0 {
			if err := tx.Exec(`DELETE FROM audit_universe_risk_scores WHERE audit_universe_year_id IN (
				SELECT id FROM audit_universe_years WHERE corporate_audit_universe_id IN ?)`, namedCorpIDs).Error; err != nil {
				return err
			}
			if err := tx.Exec(`DELETE FROM audit_universe_years WHERE corporate_audit_universe_id IN ?`, namedCorpIDs).Error; err != nil {
				return err
			}
		}
		if err := deleteByID(tx, "corporate_audit_universe", idStrings(corpPlan.Children)); err != nil {
			return err
		}
		if err := deleteByID(tx, "corporate_audit_universe", namedCorpIDs); err != nil {
			return err
		}
		if err := deleteByID(tx, "standard_audit_universe", idStrings(stdPlan.Children)); err != nil {
			return err
		}
		return deleteByID(tx, "standard_audit_universe", idStrings(stdPlan.Named))
	})
}

func deleteByID(tx *gorm.DB, table string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE id IN ?", table), ids).Error
}
