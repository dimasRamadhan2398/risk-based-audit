package annual_plan_sync

import (
	"context"
	"errors"
	"strings"
	"time"

	"audit-service/models"

	"gorm.io/gorm"
)

// maxWriteAttempts bounds the optimistic-lock retries for one plan.
const maxWriteAttempts = 3

// Syncer applies events to annual plans. It is safe for concurrent use.
type Syncer struct {
	db *gorm.DB
}

func New(db *gorm.DB) *Syncer {
	return &Syncer{db: db}
}

// ApplyEntity handles audit_universe.entity.created and .updated. It refreshes
// the name snapshot on activities linked by id, and links legacy activities that
// have a matching name but no id yet. Deleted entities are not applied: a plan
// keeps the name it was built with.
func (s *Syncer) ApplyEntity(ctx context.Context, data EntityData) error {
	if data.ID == "" {
		return nil
	}
	name := strings.TrimSpace(data.Name)
	if name != "" {
		if err := upsertRef(s.db.WithContext(ctx), models.AnnualSnapshotRef{Kind: models.SnapshotRefEntity, RefID: data.ID, Name: name}); err != nil {
			return err
		}
	}
	return s.mutatePlans(ctx, func(plan *models.AuditAnnual) bool {
		changed := false
		for i := range plan.Activities {
			act := &plan.Activities[i]
			switch {
			case act.AuditUniverseID == data.ID:
				if act.RiskName != name {
					act.RiskName = name
					changed = true
				}
			case act.AuditUniverseID == "" && name != "" && strings.EqualFold(strings.TrimSpace(act.RiskName), name):
				act.AuditUniverseID = data.ID
				changed = true
			}
		}
		return changed
	})
}

// ApplyYearScore handles audit_universe.year.scored: the risk level of an
// activity is the entity's level for the plan's year.
func (s *Syncer) ApplyYearScore(ctx context.Context, data YearScoreData) error {
	if data.CorporateAuditUniverseID == "" || data.Year == 0 {
		return nil
	}
	if data.RiskLevel != "" {
		ref := models.AnnualSnapshotRef{Kind: models.SnapshotRefYearScore, RefID: data.CorporateAuditUniverseID, Year: data.Year, RiskLevel: data.RiskLevel}
		if err := upsertRef(s.db.WithContext(ctx), ref); err != nil {
			return err
		}
	}
	return s.mutatePlansForYear(ctx, data.Year, func(plan *models.AuditAnnual) bool {
		changed := false
		for i := range plan.Activities {
			act := &plan.Activities[i]
			if act.AuditUniverseID == data.CorporateAuditUniverseID && act.RiskLevel != data.RiskLevel {
				act.RiskLevel = data.RiskLevel
				changed = true
			}
		}
		return changed
	})
}

// ApplyDepartment handles department.created and department.updated. The same
// link-by-name rule as ApplyEntity applies to legacy activities.
func (s *Syncer) ApplyDepartment(ctx context.Context, data DepartmentData) error {
	if data.ID == "" {
		return nil
	}
	name := strings.TrimSpace(data.DepartmentName)
	if name != "" {
		if err := upsertRef(s.db.WithContext(ctx), models.AnnualSnapshotRef{Kind: models.SnapshotRefDepartment, RefID: data.ID, Name: name}); err != nil {
			return err
		}
	}
	return s.mutatePlans(ctx, func(plan *models.AuditAnnual) bool {
		changed := false
		for i := range plan.Activities {
			act := &plan.Activities[i]
			switch {
			case act.DepartmentID == data.ID:
				if act.Department != name {
					act.Department = name
					changed = true
				}
			case act.DepartmentID == "" && name != "" && strings.EqualFold(strings.TrimSpace(act.Department), name):
				act.DepartmentID = data.ID
				changed = true
			}
		}
		return changed
	})
}

func (s *Syncer) mutatePlans(ctx context.Context, mutate func(*models.AuditAnnual) bool) error {
	return s.mutateMatching(ctx, nil, mutate)
}

func (s *Syncer) mutatePlansForYear(ctx context.Context, year int, mutate func(*models.AuditAnnual) bool) error {
	return s.mutateMatching(ctx, &year, mutate)
}

// mutateMatching applies mutate to every live plan (optionally only the plans of
// one year). A plan is written only when mutate reports a change. The write is
// guarded by the updated_at value that was read; on a conflict the plan is
// re-read and the mutation is applied again, up to maxWriteAttempts.
func (s *Syncer) mutateMatching(ctx context.Context, year *int, mutate func(*models.AuditAnnual) bool) error {
	q := s.db.WithContext(ctx).Model(&models.AuditAnnual{})
	if year != nil {
		q = q.Where("year = ?", *year)
	}
	var ids []string
	if err := q.Pluck("id", &ids).Error; err != nil {
		return err
	}

	for _, id := range ids {
		if err := s.mutateOne(ctx, id, mutate); err != nil {
			return err
		}
	}
	return nil
}

func (s *Syncer) mutateOne(ctx context.Context, id string, mutate func(*models.AuditAnnual) bool) error {
	for attempt := 0; attempt < maxWriteAttempts; attempt++ {
		var plan models.AuditAnnual
		if err := s.db.WithContext(ctx).First(&plan, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil // deleted between the list and this read
			}
			return err
		}
		if !mutate(&plan) {
			return nil
		}

		readAt := plan.UpdatedAt
		now := time.Now().UTC()
		// A struct update (not a map) so GORM applies the json serializer to
		// activities. Select limits the write to these two columns.
		res := s.db.WithContext(ctx).
			Model(&models.AuditAnnual{}).
			Where("id = ? AND updated_at = ?", plan.ID, readAt).
			Select("activities", "updated_at").
			Updates(models.AuditAnnual{Activities: plan.Activities, UpdatedAt: now})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 1 {
			return nil
		}
		// Another writer changed the plan after we read it. Retry on the new state.
	}
	return ErrPlanContended
}

// ErrPlanContended means a plan kept changing under the sync for maxWriteAttempts
// reads. The handler reports it so the message is redelivered.
var ErrPlanContended = errors.New("annual_plan_sync: plan was modified concurrently; retry")
