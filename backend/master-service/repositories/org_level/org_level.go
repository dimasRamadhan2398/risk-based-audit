// Package org_level is the data access for org_levels and org_change_log.
// Both tables exist only on flexible tenants (gated DDL), so construct this
// repository only when the effective mode is flexible.
package org_level

import (
	"fmt"
	"strings"
	"time"

	"master-service/models"
	apperrors "master-service/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository is the org level data access.
type Repository interface {
	Transaction(fn func(tx Repository) error) error
	DB() *gorm.DB

	// ListLive returns every live level (active or not) ordered by level.
	ListLive() ([]models.OrgLevel, error)
	// LockLive is ListLive with FOR UPDATE (no-op lock on sqlite).
	LockLive() ([]models.OrgLevel, error)
	CountLive() (int64, error)
	Create(l *models.OrgLevel) error

	// Park takes rows out of the live unique indexes for the rest of the
	// transaction (deleted_at = marker), so levels and names can be swapped
	// without tripping the non-deferrable partial unique indexes.
	Park(ids []uuid.UUID, at time.Time) error
	// SetParked updates a parked row.
	SetParked(id uuid.UUID, level int, name string, nameKey *string, at time.Time) error
	Unpark(ids []uuid.UUID, at time.Time) error
	SoftDelete(ids []uuid.UUID, at time.Time) error

	// RenumberDepartments rewrites departments.level of live departments
	// old -> new in one statement (so chains like 10->20, 20->30 are safe).
	RenumberDepartments(mapping map[int]int) (int64, error)
	// DepartmentIDsAtLevels returns live departments at any of levels.
	DepartmentIDsAtLevels(levels []int) ([]uuid.UUID, error)
	CountEmployees(departmentIDs []uuid.UUID) (int64, error)

	InsertChangeLog(row *models.OrgChangeLog) error
	ListChangeLog(offset, limit int) ([]models.OrgChangeLog, int64, error)
}

type repository struct{ db *gorm.DB }

// New returns the repository.
func New(db *gorm.DB) Repository { return &repository{db: db} }

func wrap(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %w", apperrors.ErrDatabase, err)
}

func (r *repository) DB() *gorm.DB { return r.db }

func (r *repository) Transaction(fn func(tx Repository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error { return fn(&repository{db: tx}) })
}

func (r *repository) ListLive() ([]models.OrgLevel, error) {
	var out []models.OrgLevel
	err := r.db.Order("level ASC").Find(&out).Error
	return out, wrap(err)
}

func (r *repository) LockLive() ([]models.OrgLevel, error) {
	var out []models.OrgLevel
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).Order("level ASC").Find(&out).Error
	return out, wrap(err)
}

func (r *repository) CountLive() (int64, error) {
	var n int64
	err := r.db.Model(&models.OrgLevel{}).Count(&n).Error
	return n, wrap(err)
}

func (r *repository) Create(l *models.OrgLevel) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return wrap(r.db.Create(l).Error)
}

func (r *repository) Park(ids []uuid.UUID, at time.Time) error {
	if len(ids) == 0 {
		return nil
	}
	return wrap(r.db.Model(&models.OrgLevel{}).Where("id IN ?", ids).
		UpdateColumn("deleted_at", at).Error)
}

func (r *repository) SetParked(id uuid.UUID, level int, name string, nameKey *string, at time.Time) error {
	return wrap(r.db.Unscoped().Model(&models.OrgLevel{}).Where("id = ?", id).
		UpdateColumns(map[string]any{"level": level, "name": name, "name_key": nameKey, "updated_at": at}).Error)
}

func (r *repository) Unpark(ids []uuid.UUID, at time.Time) error {
	if len(ids) == 0 {
		return nil
	}
	return wrap(r.db.Unscoped().Model(&models.OrgLevel{}).Where("id IN ?", ids).
		UpdateColumn("deleted_at", nil).Error)
}

func (r *repository) SoftDelete(ids []uuid.UUID, at time.Time) error {
	if len(ids) == 0 {
		return nil
	}
	return wrap(r.db.Model(&models.OrgLevel{}).Where("id IN ?", ids).
		UpdateColumns(map[string]any{"deleted_at": at, "updated_at": at}).Error)
}

func (r *repository) RenumberDepartments(mapping map[int]int) (int64, error) {
	if len(mapping) == 0 {
		return 0, nil
	}
	// CASE level WHEN ? THEN ? ... END: the SQL text is built from constant
	// fragments only; every number is a bound parameter.
	var sb strings.Builder
	sb.WriteString("CASE level")
	args := make([]any, 0, 2*len(mapping))
	olds := make([]int, 0, len(mapping))
	for oldLevel, newLevel := range mapping {
		sb.WriteString(" WHEN ? THEN ?")
		args = append(args, oldLevel, newLevel)
		olds = append(olds, oldLevel)
	}
	sb.WriteString(" ELSE level END")
	res := r.db.Model(&models.Department{}).Where("level IN ?", olds).
		Update("level", gorm.Expr(sb.String(), args...))
	return res.RowsAffected, wrap(res.Error)
}

func (r *repository) DepartmentIDsAtLevels(levels []int) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	if len(levels) == 0 {
		return ids, nil
	}
	err := r.db.Model(&models.Department{}).Where("level IN ?", levels).
		Order("department_name ASC").Order("id ASC").Pluck("id", &ids).Error
	return ids, wrap(err)
}

func (r *repository) CountEmployees(departmentIDs []uuid.UUID) (int64, error) {
	if len(departmentIDs) == 0 {
		return 0, nil
	}
	var n int64
	err := r.db.Model(&models.Employee{}).Where("department_id IN ?", departmentIDs).Count(&n).Error
	return n, wrap(err)
}

func (r *repository) InsertChangeLog(row *models.OrgChangeLog) error {
	if row.ID == uuid.Nil {
		row.ID = uuid.New()
	}
	return wrap(r.db.Create(row).Error)
}

func (r *repository) ListChangeLog(offset, limit int) ([]models.OrgChangeLog, int64, error) {
	var total int64
	if err := r.db.Model(&models.OrgChangeLog{}).Count(&total).Error; err != nil {
		return nil, 0, wrap(err)
	}
	var rows []models.OrgChangeLog
	err := r.db.Order("created_at DESC").Order("id DESC").Offset(offset).Limit(limit).Find(&rows).Error
	return rows, total, wrap(err)
}
