package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"audit-service/pkg/database"
	"audit-service/pkg/logger"
	"audit-service/pkg/masterclient"
	"audit-service/pkg/riskclient"
	"audit-service/pkg/servicetoken"
	"audit-service/pkg/serviceurl"
	"audit-service/services/annual_plan_sync"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// sourceCallTimeout bounds each call to master-service / risk-service. The
// yearly universe list preloads every score row, so it is generous.
const sourceCallTimeout = 60 * time.Second

var (
	backfillSnapshotsDryRun     bool
	backfillSnapshotsFromSource bool
)

// runBackfillAnnualSnapshots is "audit migrate backfill-annual-snapshots".
// Without --from-source it fills plans from the lookup as it is; with it, the
// lookup is first refreshed from master-service and risk-service. Either way
// only annual_snapshot_refs and plan snapshot fields are written, in one
// transaction under the reconcile lock, and --dry-run rolls it back.
func runBackfillAnnualSnapshots(cmd *cobra.Command, args []string) error {
	initLogger()
	defer logger.Sync()

	db, err := database.NewPostgresConnection(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", logger.LogField("error", err))
		return err
	}
	ctx := context.Background()

	before, err := annual_plan_sync.RefCounts(ctx, db)
	if err != nil {
		logger.Error("Failed to count annual_snapshot_refs", logger.LogField("error", err))
		return err
	}
	logger.Info("annual_snapshot_refs before", refCountFields(before)...)

	// Read the sources before opening the transaction, so no transaction is
	// held open across HTTP calls, and a failed read writes nothing.
	var state *annual_plan_sync.SourceState
	if backfillSnapshotsFromSource {
		// SNAPSHOT_SYNC_TOKEN (a user JWT) wins when set; otherwise a
		// short-lived service token is signed with this stack's jwt.secret.
		authorization, source, err := servicetoken.Authorization(os.Getenv(servicetoken.OverrideEnv), cfg.JWT.Secret, time.Now())
		if err != nil {
			logger.Error("Annual snapshot backfill not started: no token for master-service / risk-service", logger.LogField("error", err.Error()))
			return err
		}
		logger.Info("Calling snapshot sources", logger.LogField("auth_source", source))
		st, err := loadSnapshotSources(ctx, db, authorization)
		if err != nil {
			logger.Error("Annual snapshot backfill failed reading the sources; nothing written", logger.LogField("error", err.Error()))
			return err
		}
		state = &st
	}

	errDryRun := errors.New("dry run")
	var res annual_plan_sync.ReconcileResult
	var after map[string]int64
	err = db.Transaction(func(tx *gorm.DB) error {
		got, err := annual_plan_sync.TryReconcileLock(tx)
		if err != nil {
			return fmt.Errorf("take reconcile lock: %w", err)
		}
		if !got {
			return fmt.Errorf("%w (a starting audit-service instance?); retry in a minute", annual_plan_sync.ErrReconcileBusy)
		}
		if state != nil {
			res, err = annual_plan_sync.Reconcile(ctx, tx, *state)
		} else {
			res.Backfill, err = annual_plan_sync.BackfillSnapshots(ctx, tx)
		}
		if err != nil {
			return err
		}
		// Counted inside the transaction so a dry run shows what it would leave.
		if after, err = annual_plan_sync.RefCounts(ctx, tx); err != nil {
			return err
		}
		if backfillSnapshotsDryRun {
			return errDryRun
		}
		return nil
	})
	if err != nil && !errors.Is(err, errDryRun) {
		logger.Error("Annual snapshot backfill failed; nothing written", logger.LogField("error", err.Error()))
		return err
	}

	msg := "Annual snapshot backfill completed"
	if backfillSnapshotsDryRun {
		msg = "Annual snapshot backfill DRY RUN (rolled back, nothing written)"
	}
	fields := []zap.Field{
		logger.LogField("from_source", backfillSnapshotsFromSource),
		logger.LogField("refs_departments_upserted", res.Refs.Departments),
		logger.LogField("refs_entities_upserted", res.Refs.Entities),
		logger.LogField("refs_year_scores_upserted", res.Refs.YearScores),
		logger.LogField("refs_source_records_skipped", res.Refs.Skipped),
		logger.LogField("plans_changed", res.Backfill.PlansChanged),
		logger.LogField("fields_filled", res.Backfill.FieldsFilled),
	}
	logger.Info(msg, fields...)
	logger.Info("annual_snapshot_refs after", refCountFields(after)...)
	return nil
}

// loadSnapshotSources reads departments and the Audit Universe of this stack
// (URLs from pkg/serviceurl, so a tenant container calls its own
// master-service-<slug> / risk-service-<slug>). authorization is never logged.
func loadSnapshotSources(ctx context.Context, db *gorm.DB, authorization string) (annual_plan_sync.SourceState, error) {
	years, err := annual_plan_sync.PlanYears(ctx, db)
	if err != nil {
		return annual_plan_sync.SourceState{}, fmt.Errorf("list plan years: %w", err)
	}
	hc := &http.Client{Timeout: sourceCallTimeout}
	master := masterclient.New(serviceurl.Master(), hc)
	risk := riskclient.NewFromEnv(hc)
	logger.Info("Reading snapshot sources",
		logger.LogField("master_service_url", master.BaseURL()),
		logger.LogField("risk_service_url", risk.BaseURL()),
		logger.LogField("plan_years", years))

	st, err := annual_plan_sync.LoadSourceState(ctx, master, risk, authorization, years)
	if err != nil {
		return annual_plan_sync.SourceState{}, err
	}
	logger.Info("Read snapshot sources",
		logger.LogField("departments", len(st.Departments)),
		logger.LogField("entities", len(st.Entities)),
		logger.LogField("year_rows", len(st.YearScores)))
	return st, nil
}

// startupSnapshotLoader is the LoadFunc of the start-up reconcile: every
// attempt signs a fresh service token. The SNAPSHOT_SYNC_TOKEN override is
// deliberately not used here: a user token left in a container's environment
// would expire and make every boot fail.
func startupSnapshotLoader(db *gorm.DB, secret string) annual_plan_sync.LoadFunc {
	return func(ctx context.Context) (annual_plan_sync.SourceState, error) {
		authorization, _, err := servicetoken.Authorization("", secret, time.Now())
		if err != nil {
			return annual_plan_sync.SourceState{}, err
		}
		return loadSnapshotSources(ctx, db, authorization)
	}
}

func refCountFields(counts map[string]int64) []zap.Field {
	var out []zap.Field
	for _, kind := range []string{"department", "entity", "year_score"} {
		out = append(out, logger.LogField(kind, counts[kind]))
	}
	return out
}
