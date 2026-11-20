package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"risk-service/middleware"
	"risk-service/models"
	"risk-service/pkg/database"
	"risk-service/pkg/logger"
	"risk-service/pkg/masterclient"
	"risk-service/services"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

// backfill-rcm-department links legacy RCM rows to master departments
// (docs/coso-2013-component-assessment-plan.md 6, decision Q8: prod only,
// dry run reviewed first).
//
// Safe by construction:
//   - dry run unless --apply; the dry run only SELECTs;
//   - only live rows with department_id IS NULL are considered, and the
//     UPDATE re-checks that, so re-running is a no-op;
//   - a row is linked only when its department text matches exactly one
//     active master department name (trim, whitespace-collapsed,
//     case-insensitive); ambiguous, inactive-only and unmatched values are
//     reported and left alone;
//   - only department_id is written (UpdateColumns: no hooks, updated_at and
//     the department text untouched), all in one transaction;
//   - it never runs as part of `risk up`.
//
// master-service requires a JWT. The command signs a short-lived token with
// this stack's jwt.secret that carries no roles (the department list needs
// only authentication), or uses RCM_BACKFILL_TOKEN when set.
//
//	./risk backfill-rcm-department           # dry run, prints the plan
//	./risk backfill-rcm-department --apply   # writes it
var backfillRCMDepartmentCmd = &cobra.Command{
	Use:   "backfill-rcm-department",
	Short: "Link legacy RCM rows to master departments by name (dry run unless --apply)",
	RunE:  runBackfillRCMDepartment,
}

var backfillRCMDepartmentApply bool

// rcmBackfillTokenEnv may hold a user JWT to present instead of the
// self-signed one ("Bearer " prefix optional).
const rcmBackfillTokenEnv = "RCM_BACKFILL_TOKEN"

func init() {
	backfillRCMDepartmentCmd.Flags().BoolVar(&backfillRCMDepartmentApply, "apply", false, "write the changes (default is a dry run)")
	rootCmd.AddCommand(backfillRCMDepartmentCmd)
}

// rcmDepartmentLink is one planned update.
type rcmDepartmentLink struct {
	RCMID          uuid.UUID
	ControlCode    string
	FromValue      string
	DepartmentID   uuid.UUID
	DepartmentName string
}

// rcmDepartmentMiss is a row that stays unassigned.
type rcmDepartmentMiss struct {
	RCMID       uuid.UUID
	ControlCode string
	FromValue   string
	Reason      string
}

func normalizeDepartmentName(s string) string { return services.NormalizeDepartmentName(s) }

// planRCMDepartmentBackfill decides which rows can be linked. Rows that
// already have a department_id are skipped.
func planRCMDepartmentBackfill(rows []models.RiskControlMatrix, depts []masterclient.Department) ([]rcmDepartmentLink, []rcmDepartmentMiss) {
	active := map[string][]masterclient.Department{}
	inactive := map[string]bool{}
	for _, d := range depts {
		key := normalizeDepartmentName(d.Name)
		if key == "" {
			continue
		}
		if d.IsActive {
			active[key] = append(active[key], d)
		} else {
			inactive[key] = true
		}
	}

	var plan []rcmDepartmentLink
	var misses []rcmDepartmentMiss
	for _, r := range rows {
		if r.DepartmentID != nil {
			continue
		}
		key := normalizeDepartmentName(r.Department)
		miss := rcmDepartmentMiss{RCMID: r.ID, ControlCode: r.ControlCode, FromValue: r.Department}
		matches := active[key]
		switch {
		case key == "":
			miss.Reason = "no department text"
		case len(matches) == 1:
			plan = append(plan, rcmDepartmentLink{
				RCMID: r.ID, ControlCode: r.ControlCode, FromValue: r.Department,
				DepartmentID: matches[0].ID, DepartmentName: matches[0].Name,
			})
			continue
		case len(matches) > 1:
			miss.Reason = fmt.Sprintf("matches %d active departments", len(matches))
		case inactive[key]:
			miss.Reason = "matches an inactive department only"
		default:
			miss.Reason = "no master department with this name"
		}
		misses = append(misses, miss)
	}
	sort.SliceStable(plan, func(i, j int) bool { return plan[i].RCMID.String() < plan[j].RCMID.String() })
	sort.SliceStable(misses, func(i, j int) bool { return misses[i].RCMID.String() < misses[j].RCMID.String() })
	return plan, misses
}

// applyRCMDepartmentBackfill writes plan in one transaction and returns the
// number of rows updated.
func applyRCMDepartmentBackfill(db *gorm.DB, plan []rcmDepartmentLink) (int64, error) {
	var updated int64
	err := db.Transaction(func(tx *gorm.DB) error {
		for _, p := range plan {
			res := tx.Model(&models.RiskControlMatrix{}).
				Where("id = ? AND department_id IS NULL", p.RCMID).
				UpdateColumn("department_id", p.DepartmentID)
			if res.Error != nil {
				return fmt.Errorf("update rcm %s: %w", p.RCMID, res.Error)
			}
			updated += res.RowsAffected
		}
		return nil
	})
	return updated, err
}

// masterAuthorization returns the Authorization header for CLI calls to
// master-service: RCM_BACKFILL_TOKEN, else a 5-minute self-signed token with
// no roles. Never log the result.
func masterAuthorization(secret string, now time.Time) (string, string, error) {
	if raw := strings.TrimSpace(os.Getenv(rcmBackfillTokenEnv)); raw != "" {
		fields := strings.Fields(raw)
		if len(fields) > 0 && strings.EqualFold(fields[0], "bearer") {
			fields = fields[1:]
		}
		if len(fields) != 1 {
			return "", "", fmt.Errorf("%s is set but does not hold a token", rcmBackfillTokenEnv)
		}
		return "Bearer " + fields[0], "override", nil
	}
	if err := middleware.ValidateJWTSecret(secret); err != nil {
		return "", "", fmt.Errorf("cannot sign a token for master-service: %w", err)
	}
	const subject = "system:risk-service"
	now = now.UTC()
	claims := middleware.Claims{
		UserID:   subject,
		Username: subject,
		Roles:    []string{},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			Issuer:    "risk-service",
			Audience:  []string{"master-service"},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
			ID:        uuid.NewString(),
		},
	}
	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		return "", "", err
	}
	return "Bearer " + tok, "self-signed", nil
}

func runBackfillRCMDepartment(cmd *cobra.Command, args []string) error {
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
	if !db.Migrator().HasColumn(&models.RiskControlMatrix{}, "department_id") {
		return fmt.Errorf("risk_control_matrix.department_id does not exist: run `./risk up` first")
	}

	client := masterclient.NewDepartmentClient(
		cfg.MasterService.BaseURL(),
		time.Duration(cfg.MasterService.TimeoutSeconds)*time.Second,
		time.Duration(cfg.MasterService.CacheTTLSeconds)*time.Second,
	)
	auth, source, err := masterAuthorization(cfg.JWT.Secret, time.Now())
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	depts, err := client.ListDepartments(masterclient.WithAuthorization(ctx, auth))
	if err != nil {
		// Never guess: abort instead.
		return fmt.Errorf("read departments from %s (token: %s): %w", client.BaseURL(), source, err)
	}

	var rows []models.RiskControlMatrix
	if err := db.Where("department_id IS NULL").Order("year, control_code, id").Find(&rows).Error; err != nil {
		return fmt.Errorf("load rcm rows: %w", err)
	}

	plan, misses := planRCMDepartmentBackfill(rows, depts)

	fmt.Printf("Database: %s  Department master: %s (%d departments, token: %s)\n", cfg.Database.Name, client.BaseURL(), len(depts), source)
	fmt.Printf("RCM rows without department_id: %d  to link: %d  unmatched: %d\n", len(rows), len(plan), len(misses))
	for _, p := range plan {
		fmt.Printf("  link  %s  %-14s department=%q -> %s (%s)\n", p.RCMID, p.ControlCode, p.FromValue, p.DepartmentName, p.DepartmentID)
	}
	for _, m := range misses {
		fmt.Printf("  skip  %s  %-14s department=%q: %s\n", m.RCMID, m.ControlCode, m.FromValue, m.Reason)
	}

	if !backfillRCMDepartmentApply {
		fmt.Println("Dry run: nothing written. Re-run with --apply to write.")
		return nil
	}
	if len(plan) == 0 {
		fmt.Println("Nothing to write.")
		return nil
	}
	updated, err := applyRCMDepartmentBackfill(db, plan)
	if err != nil {
		return err
	}
	fmt.Printf("Updated %d risk_control_matrix rows.\n", updated)
	return nil
}
