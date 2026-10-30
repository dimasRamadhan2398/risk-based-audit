package annual_plan_sync

import (
	"context"
	"fmt"
	"strings"

	"audit-service/models"
	"audit-service/pkg/masterclient"
	"audit-service/pkg/riskclient"

	"gorm.io/gorm"
)

// This file reconciles the lookup (models.AnnualSnapshotRef) with the source
// services directly, for the data the events never delivered: departments and
// Audit Universe nodes that master-service and risk-service seed straight into
// their databases, and anything written while the broker or the consumer was
// down. It writes the lookup with the same rules as the event handlers in
// sync.go (upsertRef, a blank value never replaces a name) and then runs
// BackfillSnapshots, which only fills empty plan fields.

// SourceState is the current state of the sources, in the shapes of the event
// data the handlers in sync.go apply.
type SourceState struct {
	Departments []DepartmentData
	Entities    []EntityData
	YearScores  []YearScoreData
}

// DepartmentSource lists master-service departments (*masterclient.MasterClient).
type DepartmentSource interface {
	ListDepartments(ctx context.Context, authorization string) ([]masterclient.Department, error)
}

// UniverseSource lists the corporate Audit Universe and its yearly rows
// (*riskclient.RiskClient).
type UniverseSource interface {
	ListCorporateEntities(ctx context.Context, authorization string) ([]riskclient.Entity, error)
	ListYearUniverse(ctx context.Context, authorization string, year int) ([]riskclient.YearRow, error)
}

// PlanYears returns the distinct years of the live annual plans, ascending.
// These are the only years whose risk levels a plan snapshot can use.
func PlanYears(ctx context.Context, db *gorm.DB) ([]int, error) {
	var years []int
	err := db.WithContext(ctx).Model(&models.AuditAnnual{}).
		Where("year > ?", 0).
		Distinct("year").
		Order("year").
		Pluck("year", &years).Error
	return years, err
}

// LoadSourceState reads every department, every corporate Audit Universe node
// and the yearly rows of years, forwarding authorization to both services. Any
// failed call fails the whole load, so a partial view is never written.
//
// Entities come from the corporate tree and, for nodes deeper than the tree
// endpoint preloads, from the entity embedded in each yearly row.
func LoadSourceState(ctx context.Context, depts DepartmentSource, universe UniverseSource, authorization string, years []int) (SourceState, error) {
	var st SourceState

	dl, err := depts.ListDepartments(ctx, authorization)
	if err != nil {
		return SourceState{}, fmt.Errorf("list departments from master-service: %w", err)
	}
	for _, d := range dl {
		st.Departments = append(st.Departments, DepartmentData{
			ID:             d.ID,
			DepartmentCode: d.DepartmentCode,
			DepartmentName: d.DepartmentName,
			CompanyID:      d.CompanyID,
		})
	}

	seen := make(map[string]int) // normalised id -> index in st.Entities
	addEntity := func(e riskclient.Entity) {
		id := normRefID(e.ID)
		if id == "" {
			return
		}
		if i, ok := seen[id]; ok {
			if strings.TrimSpace(st.Entities[i].Name) == "" {
				st.Entities[i].Name = e.Name
			}
			return
		}
		seen[id] = len(st.Entities)
		st.Entities = append(st.Entities, EntityData{ID: e.ID, Name: e.Name, ParentID: e.ParentID})
	}

	tree, err := universe.ListCorporateEntities(ctx, authorization)
	if err != nil {
		return SourceState{}, fmt.Errorf("list corporate audit universe from risk-service: %w", err)
	}
	for _, e := range tree {
		addEntity(e)
	}

	for _, year := range years {
		rows, err := universe.ListYearUniverse(ctx, authorization, year)
		if err != nil {
			return SourceState{}, fmt.Errorf("list %d audit universe from risk-service: %w", year, err)
		}
		for _, r := range rows {
			if r.CorporateAuditUniverse != nil {
				e := *r.CorporateAuditUniverse
				if e.ID == "" {
					e.ID = r.CorporateAuditUniverseID
				}
				addEntity(e)
			}
			y := r.Year
			if y == 0 {
				y = year
			}
			st.YearScores = append(st.YearScores, YearScoreData{
				CorporateAuditUniverseID: r.CorporateAuditUniverseID,
				Year:                     y,
				RiskIndex:                r.RiskIndex,
				RiskLevel:                r.RiskLevel,
				AuditPriority:            r.AuditPriority,
			})
		}
	}
	return st, nil
}

// RefreshResult counts the lookup rows RefreshRefs wrote and the source
// records it skipped (no id, or no name / no risk level yet).
type RefreshResult struct {
	Departments int
	Entities    int
	YearScores  int
	Skipped     int
}

// RefreshRefs upserts the lookup from st with the rules of the event handlers:
// the name (or, for a year score, the risk level) of an id is replaced by the
// source's current value, and a blank value is skipped, never written. Rows
// are never deleted. Running it again with the same state changes nothing but
// updated_at.
func RefreshRefs(ctx context.Context, db *gorm.DB, st SourceState) (RefreshResult, error) {
	tx := db.WithContext(ctx)
	var res RefreshResult

	for _, d := range st.Departments {
		name := strings.TrimSpace(d.DepartmentName)
		if normRefID(d.ID) == "" || name == "" {
			res.Skipped++
			continue
		}
		if err := upsertRef(tx, models.AnnualSnapshotRef{Kind: models.SnapshotRefDepartment, RefID: d.ID, Name: name}); err != nil {
			return res, err
		}
		res.Departments++
	}

	for _, e := range st.Entities {
		name := strings.TrimSpace(e.Name)
		if normRefID(e.ID) == "" || name == "" {
			res.Skipped++
			continue
		}
		if err := upsertRef(tx, models.AnnualSnapshotRef{Kind: models.SnapshotRefEntity, RefID: e.ID, Name: name}); err != nil {
			return res, err
		}
		res.Entities++
	}

	for _, y := range st.YearScores {
		level := strings.TrimSpace(y.RiskLevel)
		if normRefID(y.CorporateAuditUniverseID) == "" || y.Year == 0 || level == "" {
			res.Skipped++
			continue
		}
		ref := models.AnnualSnapshotRef{Kind: models.SnapshotRefYearScore, RefID: y.CorporateAuditUniverseID, Year: y.Year, RiskLevel: level}
		if err := upsertRef(tx, ref); err != nil {
			return res, err
		}
		res.YearScores++
	}
	return res, nil
}

// ReconcileResult is what Reconcile wrote.
type ReconcileResult struct {
	Refs     RefreshResult
	Backfill BackfillResult
}

// Reconcile refreshes the lookup from st and then fills the empty snapshots of
// every live plan from it (BackfillSnapshots: a field with a value is never
// changed). It touches only annual_snapshot_refs and the Department, RiskName
// and RiskLevel of plan activities. Pass a transaction so a failure (or a dry
// run) leaves nothing half-written.
func Reconcile(ctx context.Context, db *gorm.DB, st SourceState) (ReconcileResult, error) {
	refs, err := RefreshRefs(ctx, db, st)
	if err != nil {
		return ReconcileResult{Refs: refs}, fmt.Errorf("refresh annual snapshot refs: %w", err)
	}
	fill, err := BackfillSnapshots(ctx, db)
	if err != nil {
		return ReconcileResult{Refs: refs}, fmt.Errorf("backfill annual plan snapshots: %w", err)
	}
	return ReconcileResult{Refs: refs, Backfill: fill}, nil
}

// RefCounts returns the number of lookup rows per kind (for operator logs).
func RefCounts(ctx context.Context, db *gorm.DB) (map[string]int64, error) {
	type row struct {
		Kind  string
		Count int64
	}
	var rows []row
	if err := db.WithContext(ctx).Model(&models.AnnualSnapshotRef{}).
		Select("kind, COUNT(*) AS count").Group("kind").Scan(&rows).Error; err != nil {
		return nil, err
	}
	out := map[string]int64{
		models.SnapshotRefDepartment: 0,
		models.SnapshotRefEntity:     0,
		models.SnapshotRefYearScore:  0,
	}
	for _, r := range rows {
		out[r.Kind] = r.Count
	}
	return out, nil
}
