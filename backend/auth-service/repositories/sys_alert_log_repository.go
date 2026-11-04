package repositories

import (
	"context"
	"strings"
	"time"

	"auth-service/models"
	"auth-service/pkg/activitylog"
	apperrors "auth-service/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ISysAlertLogRepository is an alias for the sys alert log repository interface
type ISysAlertLogRepository = SysAlertLogRepositoryInterface

// SysAlertLogRepositoryInterface defines the sys alert log (activity log) repository
type SysAlertLogRepositoryInterface interface {
	activitylog.Store
	Create(log *models.SysAlertLog) error
	FindByID(id uuid.UUID) (*models.SysAlertLog, error)
	List(ctx context.Context, f SysAlertLogFilter) ([]models.SysAlertLog, error)
	DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error)
	DeleteOldLogs(days int) error
}

// LogCursor is a keyset position: a timestamp and the row id that breaks ties.
type LogCursor struct {
	Time time.Time
	ID   uuid.UUID
}

// SysAlertLogFilter is the whitelisted query of GET /system-logs. Every field
// maps to one fixed column; nothing from the request names a column.
type SysAlertLogFilter struct {
	Limit int
	// Before pages to older rows: (occurred_at, id) < Before, newest first.
	Before *LogCursor
	// After polls for rows stored after a point: (created_at, id) > After,
	// oldest first (the caller re-sorts). Mutually exclusive with Before.
	After *LogCursor

	Actions     []string
	ActorUserID *uuid.UUID
	Result      string
	From        *time.Time
	To          *time.Time
	// Q matches actor_username or summary, case-insensitively, as a substring.
	Q string
}

// Column names used by the query builder (the whitelist).
var (
	colOccurredAt    = clause.Column{Name: "occurred_at"}
	colCreatedAt     = clause.Column{Name: "created_at"}
	colID            = clause.Column{Name: "id"}
	colAction        = clause.Column{Name: "action"}
	colActorUserID   = clause.Column{Name: "actor_user_id"}
	colActorUsername = clause.Column{Name: "actor_username"}
	colResult        = clause.Column{Name: "result"}
	colSummary       = clause.Column{Name: "summary"}
	colEventID       = clause.Column{Name: "event_id"}
)

// SysAlertLogRepository handles system alert log data operations
type SysAlertLogRepository struct {
	*BaseRepository
}

// NewSysAlertLogRepository creates a new sys alert log repository
func NewSysAlertLogRepository(db *gorm.DB) ISysAlertLogRepository {
	return &SysAlertLogRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new system alert log
func (r *SysAlertLogRepository) Create(log *models.SysAlertLog) error {
	return r.BaseRepository.Create(log)
}

// InsertEntry stores an activity entry, doing nothing when a row with the same
// event_id exists (Kafka redelivery, or the fallback writing an event the
// consumer also stored).
func (r *SysAlertLogRepository) InsertEntry(ctx context.Context, e *activitylog.Entry) (bool, error) {
	row := EntryToRow(e)
	res := r.GetDB().WithContext(ctx).
		Clauses(clause.OnConflict{Columns: []clause.Column{colEventID}, DoNothing: true}).
		Create(&row)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// EntryToRow converts a normalised entry to a row.
func EntryToRow(e *activitylog.Entry) models.SysAlertLog {
	row := models.SysAlertLog{
		ID:            uuid.New(),
		EventID:       e.EventID,
		OccurredAt:    e.OccurredAt,
		ActorUsername: e.ActorUsername,
		Action:        models.SysAlertAction(e.Action),
		EntityType:    e.EntityType,
		EntityID:      e.EntityID,
		Result:        e.Result,
		IPAddress:     e.IPAddress,
		UserAgent:     e.UserAgent,
		Summary:       e.Summary,
		CreatedAt:     time.Now().UTC().Truncate(time.Microsecond),
	}
	row.UpdatedAt = row.CreatedAt
	if id, err := uuid.Parse(e.ActorUserID); err == nil {
		row.ActorUserID = &id
	}
	if len(e.Details) > 0 {
		row.Details = models.JSONMap(e.Details)
	}
	return row
}

// FindByID finds a system alert log by ID
func (r *SysAlertLogRepository) FindByID(id uuid.UUID) (*models.SysAlertLog, error) {
	var log models.SysAlertLog
	if err := r.GetDB().Where(clause.Eq{Column: colID, Value: id}).First(&log).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &log, nil
}

// List returns one page of the activity log for f.
func (r *SysAlertLogRepository) List(ctx context.Context, f SysAlertLogFilter) ([]models.SysAlertLog, error) {
	q := r.listQuery(r.GetDB().WithContext(ctx), f)
	var rows []models.SysAlertLog
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// listQuery builds the SELECT of List (exposed for DryRun tests).
func (r *SysAlertLogRepository) listQuery(db *gorm.DB, f SysAlertLogFilter) *gorm.DB {
	q := db.Model(&models.SysAlertLog{})

	if len(f.Actions) == 1 {
		q = q.Where(clause.Eq{Column: colAction, Value: f.Actions[0]})
	} else if len(f.Actions) > 1 {
		vals := make([]interface{}, len(f.Actions))
		for i, a := range f.Actions {
			vals[i] = a
		}
		q = q.Where(clause.IN{Column: colAction, Values: vals})
	}
	if f.ActorUserID != nil {
		q = q.Where(clause.Eq{Column: colActorUserID, Value: *f.ActorUserID})
	}
	if f.Result != "" {
		q = q.Where(clause.Eq{Column: colResult, Value: f.Result})
	}
	if f.From != nil {
		q = q.Where(clause.Gte{Column: colOccurredAt, Value: f.From.UTC()})
	}
	if f.To != nil {
		q = q.Where(clause.Lte{Column: colOccurredAt, Value: f.To.UTC()})
	}
	if s := strings.TrimSpace(f.Q); s != "" {
		pattern := "%" + escapeLike(strings.ToLower(s)) + "%"
		q = q.Where(clause.Or(
			clause.Expr{SQL: "LOWER(?) LIKE ? ESCAPE '\\'", Vars: []interface{}{colActorUsername, pattern}},
			clause.Expr{SQL: "LOWER(?) LIKE ? ESCAPE '\\'", Vars: []interface{}{colSummary, pattern}},
		))
	}

	switch {
	case f.After != nil:
		q = q.Where(clause.Or(
			clause.Gt{Column: colCreatedAt, Value: f.After.Time.UTC()},
			clause.And(
				clause.Eq{Column: colCreatedAt, Value: f.After.Time.UTC()},
				clause.Gt{Column: colID, Value: f.After.ID},
			),
		)).
			Order(clause.OrderByColumn{Column: colCreatedAt}).
			Order(clause.OrderByColumn{Column: colID})
	default:
		if f.Before != nil {
			q = q.Where(clause.Or(
				clause.Lt{Column: colOccurredAt, Value: f.Before.Time.UTC()},
				clause.And(
					clause.Eq{Column: colOccurredAt, Value: f.Before.Time.UTC()},
					clause.Lt{Column: colID, Value: f.Before.ID},
				),
			))
		}
		q = q.Order(clause.OrderByColumn{Column: colOccurredAt, Desc: true}).
			Order(clause.OrderByColumn{Column: colID, Desc: true})
	}

	if f.Limit > 0 {
		q = q.Limit(f.Limit)
	}
	return q
}

// escapeLike escapes the LIKE wildcards of a user search term.
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// DeleteOlderThan deletes rows that occurred before cutoff and returns how many.
func (r *SysAlertLogRepository) DeleteOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	res := deleteOlderThan(r.GetDB().WithContext(ctx), cutoff)
	return res.RowsAffected, res.Error
}

// deleteOlderThan runs the retention DELETE on db (exposed for DryRun tests).
func deleteOlderThan(db *gorm.DB, cutoff time.Time) *gorm.DB {
	return db.Where(clause.Lt{Column: colOccurredAt, Value: cutoff.UTC()}).Delete(&models.SysAlertLog{})
}

// DeleteOldLogs deletes logs older than the given number of days. The cutoff is
// computed in Go: the previous `INTERVAL '? days'` put the placeholder inside a
// string literal, so it was never bound.
func (r *SysAlertLogRepository) DeleteOldLogs(days int) error {
	if days <= 0 {
		return apperrors.Wrap("INVALID_RETENTION", "Retention days must be positive", 400, nil)
	}
	_, err := r.DeleteOlderThan(context.Background(), RetentionCutoff(time.Now(), days))
	return err
}

// RetentionCutoff is now minus days.
func RetentionCutoff(now time.Time, days int) time.Time {
	return now.UTC().AddDate(0, 0, -days)
}
