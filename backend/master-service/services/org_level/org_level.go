// Package org_level implements the org level editor of flexible tenants:
// list, dry-run preview, password-confirmed apply and the change log
// (docs/org-structure-model-plan.md section 9).
package org_level

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/pkg/kafka"
	"master-service/pkg/orgmode"
	deptRepo "master-service/repositories/department"
	levelRepo "master-service/repositories/org_level"
	deptSvc "master-service/services/department"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LevelInput is one level of the desired set. ID is empty for a new level.
// NameKey is honoured only for a new level; renaming an existing level always
// clears its key (the name is then shown as typed).
type LevelInput struct {
	ID      *uuid.UUID `json:"id"`
	Level   int        `json:"level"`
	Name    string     `json:"name"`
	NameKey *string    `json:"name_key"`
}

// ChangeRequest is the full desired level set plus the levels to delete. Every
// current live level must appear exactly once, either in Levels or DeleteIDs.
type ChangeRequest struct {
	Levels    []LevelInput `json:"levels"`
	DeleteIDs []uuid.UUID  `json:"delete_ids"`
}

// ApplyRequest is ChangeRequest plus the base_version returned by preview.
type ApplyRequest struct {
	ChangeRequest
	BaseVersion string `json:"base_version"`
}

// LevelRef identifies a level in a diff.
type LevelRef struct {
	ID      *uuid.UUID `json:"id,omitempty"`
	Level   int        `json:"level"`
	Name    string     `json:"name"`
	NameKey *string    `json:"name_key"`
}

// LevelChange is a rename or renumber of an existing level.
type LevelChange struct {
	ID        uuid.UUID `json:"id"`
	FromLevel int       `json:"from_level"`
	ToLevel   int       `json:"to_level"`
	FromName  string    `json:"from_name"`
	ToName    string    `json:"to_name"`
}

// Diff is the difference between the current and the desired level set.
type Diff struct {
	Added      []LevelRef    `json:"added"`
	Renamed    []LevelChange `json:"renamed"`
	Renumbered []LevelChange `json:"renumbered"`
	Deleted    []LevelRef    `json:"deleted"`
}

// MaxImpactIDs caps department_ids in the impact summary.
const MaxImpactIDs = 50

// Impact is what the change does to existing data.
type Impact struct {
	// Departments whose level number or level name changes.
	DepartmentCount int         `json:"department_count"`
	DepartmentIDs   []uuid.UUID `json:"department_ids"` // first MaxImpactIDs
	EmployeeCount   int64       `json:"employee_count"`
	// UnitUsageIDs: every affected department id, for audit-service
	// POST /assignment-letters/unit-usage (the letter count).
	UnitUsageIDs []uuid.UUID `json:"unit_usage_ids"`
	// Letters are never modified: they keep their snapshot level name.
	LettersModified bool `json:"letters_modified"`
}

// Violation is a reason the change cannot be applied.
type Violation struct {
	Code          string      `json:"code"`
	Message       string      `json:"message"`
	LevelID       *uuid.UUID  `json:"level_id,omitempty"`
	Level         *int        `json:"level,omitempty"`
	DepartmentIDs []uuid.UUID `json:"department_ids,omitempty"`
}

// Violation codes.
const (
	VLevelsEmpty        = "LEVELS_EMPTY"
	VTooManyLevels      = "TOO_MANY_LEVELS"
	VLevelOutOfRange    = "LEVEL_OUT_OF_RANGE"
	VNameRequired       = "LEVEL_NAME_REQUIRED"
	VNameTooLong        = "LEVEL_NAME_TOO_LONG"
	VInvalidNameKey     = "LEVEL_NAME_KEY_INVALID"
	VDuplicateLevel     = "DUPLICATE_LEVEL"
	VDuplicateName      = "DUPLICATE_LEVEL_NAME"
	VUnknownLevelID     = "UNKNOWN_LEVEL_ID"
	VDuplicateLevelID   = "DUPLICATE_LEVEL_ID"
	VDeleteConflict     = "LEVEL_BOTH_KEPT_AND_DELETED"
	VLevelMissing       = "LEVEL_MISSING_FROM_REQUEST"
	VLevelInUse         = "LEVEL_IN_USE"
	VLevelOrderConflict = "LEVEL_ORDER_CONFLICT"
)

// Preview is the dry-run result.
type Preview struct {
	BaseVersion string      `json:"base_version"`
	Diff        Diff        `json:"diff"`
	Impact      Impact      `json:"impact"`
	Violations  []Violation `json:"violations"`
	CanApply    bool        `json:"can_apply"`
}

// ApplyResult is returned by Apply.
type ApplyResult struct {
	Levels       []models.OrgLevel `json:"levels"`
	BaseVersion  string            `json:"base_version"`
	ChangeLogID  uuid.UUID         `json:"change_log_id"`
	EventsQueued int               `json:"events_queued"`
	Preview      Preview           `json:"preview"`
}

// Actor is who applies a change (from the access token and the request).
type Actor struct {
	UserID, Username string
	Roles            []string
	IP, UserAgent    string
}

// ServiceInterface is the org level service.
type ServiceInterface interface {
	List(ctx context.Context) ([]models.OrgLevel, error)
	Preview(ctx context.Context, req ChangeRequest) (*Preview, error)
	Apply(ctx context.Context, req ApplyRequest, actor Actor) (*ApplyResult, error)
	ChangeLog(ctx context.Context, offset, limit int) ([]models.OrgChangeLog, int64, error)
}

// Service implements ServiceInterface on a flexible tenant's database.
type Service struct {
	db     *gorm.DB
	events deptSvc.OrgEventPublisher
	now    func() time.Time
}

// NewService builds the service. Construct it only in flexible mode.
func NewService(db *gorm.DB, events deptSvc.OrgEventPublisher) *Service {
	return &Service{db: db, events: events, now: time.Now}
}

// ViolationsError is the 422 returned by Apply when the change is invalid.
type ViolationsError struct {
	*apperrors.AppError
	Violations []Violation
}

// Unwrap exposes the AppError so generic error handling still answers 422.
func (e *ViolationsError) Unwrap() error { return e.AppError }

func (s *Service) List(ctx context.Context) ([]models.OrgLevel, error) {
	levels, err := levelRepo.New(s.db.WithContext(ctx)).ListLive()
	if err != nil {
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to load organisation levels.")
	}
	if levels == nil {
		levels = []models.OrgLevel{}
	}
	return levels, nil
}

func (s *Service) Preview(ctx context.Context, req ChangeRequest) (*Preview, error) {
	db := s.db.WithContext(ctx)
	current, err := levelRepo.New(db).ListLive()
	if err != nil {
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to load organisation levels.")
	}
	depts, err := deptRepo.NewOrgRepository(db).ListLive(nil)
	if err != nil {
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to load departments.")
	}
	p := computePlan(current, depts, req)
	n, err := levelRepo.New(db).CountEmployees(p.affected)
	if err != nil {
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to count employees.")
	}
	pv := p.preview(Version(current), n)
	return &pv, nil
}

var errVersionConflict = apperrors.New(apperrors.CodeOrgLevelsVersionConflict,
	"The organisation levels were changed by someone else. Review the preview again.", http.StatusConflict)

func (s *Service) Apply(ctx context.Context, req ApplyRequest, actor Actor) (*ApplyResult, error) {
	var result ApplyResult
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		lv := levelRepo.New(tx)
		dr := deptRepo.NewOrgRepository(tx)

		current, err := lv.LockLive()
		if err != nil {
			return apperrors.DB(err, apperrors.OpRead, "Failed to load organisation levels.")
		}
		base := Version(current)
		if strings.TrimSpace(req.BaseVersion) == "" || req.BaseVersion != base {
			return errVersionConflict
		}
		depts, err := dr.ListLive(nil)
		if err != nil {
			return apperrors.DB(err, apperrors.OpRead, "Failed to load departments.")
		}
		// Recomputed server-side; the client's preview is never trusted.
		p := computePlan(current, depts, req.ChangeRequest)
		employees, err := lv.CountEmployees(p.affected)
		if err != nil {
			return apperrors.DB(err, apperrors.OpRead, "Failed to count employees.")
		}
		pv := p.preview(base, employees)
		if !pv.CanApply {
			return &ViolationsError{
				AppError: apperrors.New(apperrors.CodeOrgLevelChangeInvalid,
					"The level change is not valid. See the violations.", http.StatusUnprocessableEntity),
				Violations: pv.Violations,
			}
		}

		at := s.now().UTC().Truncate(time.Microsecond)
		if err := lv.SoftDelete(p.deleteIDs, at); err != nil {
			return apperrors.DB(err, apperrors.OpUpdate, "Failed to delete levels.")
		}
		// Two-phase: park every renamed/renumbered row outside the live unique
		// indexes, rewrite it, then bring it back. Unique indexes cannot be
		// deferred, and a +offset renumber would break the 1..999 CHECK.
		changedIDs := make([]uuid.UUID, 0, len(p.changed))
		for id := range p.changed {
			changedIDs = append(changedIDs, id)
		}
		sort.Slice(changedIDs, func(i, j int) bool { return changedIDs[i].String() < changedIDs[j].String() })
		if err := lv.Park(changedIDs, at); err != nil {
			return apperrors.DB(err, apperrors.OpUpdate, "Failed to update levels.")
		}
		for _, id := range changedIDs {
			c := p.changed[id]
			if err := lv.SetParked(id, c.Level, c.Name, c.NameKey, at); err != nil {
				return apperrors.DB(err, apperrors.OpUpdate, "Failed to update levels.")
			}
		}
		if err := lv.Unpark(changedIDs, at); err != nil {
			return apperrors.DB(err, apperrors.OpUpdate, "Failed to update levels.")
		}
		if _, err := lv.RenumberDepartments(p.mapping); err != nil {
			return apperrors.DB(err, apperrors.OpUpdate, "Failed to move departments to the new level numbers.")
		}
		for i := range p.added {
			l := p.added[i]
			l.CreatedAt, l.UpdatedAt = at, at
			if err := lv.Create(&l); err != nil {
				return apperrors.DB(err, apperrors.OpCreate, "Failed to create a level.")
			}
		}

		after, err := lv.ListLive()
		if err != nil {
			return apperrors.DB(err, apperrors.OpRead, "Failed to load organisation levels.")
		}
		row, err := changeLogRow(actor, base, current, after, pv)
		if err != nil {
			return apperrors.Internal("Failed to record the change.", err)
		}
		row.CreatedAt = at
		if err := lv.InsertChangeLog(row); err != nil {
			return apperrors.DB(err, apperrors.OpCreate, "Failed to record the change.")
		}

		queued, err := s.republish(ctx, tx, dr, p.affected)
		if err != nil {
			return err
		}
		result = ApplyResult{Levels: after, BaseVersion: Version(after), ChangeLogID: row.ID, EventsQueued: queued, Preview: pv}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// republish enqueues department.updated for every affected department and its
// descendants (their paths carry the ancestors' level names).
func (s *Service) republish(ctx context.Context, tx *gorm.DB, dr deptRepo.OrgRepository, affected []uuid.UUID) (int, error) {
	if s.events == nil || len(affected) == 0 {
		return 0, nil
	}
	depts, err := dr.ListLive(nil)
	if err != nil {
		return 0, apperrors.DB(err, apperrors.OpRead, "Failed to load departments.")
	}
	levels, err := dr.ActiveLevels()
	if err != nil {
		return 0, apperrors.DB(err, apperrors.OpRead, "Failed to load organisation levels.")
	}
	ix := deptSvc.NewIndex(depts, levels)
	seen := map[uuid.UUID]bool{}
	n := 0
	for _, id := range affected {
		for _, d := range append([]uuid.UUID{id}, ix.Descendants(id)...) {
			if seen[d] {
				continue
			}
			seen[d] = true
			dec := ix.Decorated(d)
			if dec == nil {
				continue
			}
			if err := s.events.PublishDepartmentOrgTx(ctx, tx, kafka.DomainEventDepartmentUpdated, dec); err != nil {
				return n, apperrors.Internal("Failed to record the department event.", err)
			}
			n++
		}
	}
	return n, nil
}

func (s *Service) ChangeLog(ctx context.Context, offset, limit int) ([]models.OrgChangeLog, int64, error) {
	rows, total, err := levelRepo.New(s.db.WithContext(ctx)).ListChangeLog(offset, limit)
	if err != nil {
		return nil, 0, apperrors.DB(err, apperrors.OpRead, "Failed to load the change history.")
	}
	if rows == nil {
		rows = []models.OrgChangeLog{}
	}
	return rows, total, nil
}

func changeLogRow(a Actor, base string, before, after []models.OrgLevel, pv Preview) (*models.OrgChangeLog, error) {
	roles := a.Roles
	if roles == nil {
		roles = []string{}
	}
	r, err := models.ToJSONText(roles)
	if err != nil {
		return nil, err
	}
	b, err := models.ToJSONText(before)
	if err != nil {
		return nil, err
	}
	af, err := models.ToJSONText(after)
	if err != nil {
		return nil, err
	}
	im, err := models.ToJSONText(map[string]any{"diff": pv.Diff, "impact": pv.Impact})
	if err != nil {
		return nil, err
	}
	return &models.OrgChangeLog{
		ID:            uuid.New(),
		Action:        models.OrgChangeLogActionApply,
		ActorUserID:   truncate(a.UserID, 100),
		ActorUsername: truncate(a.Username, 255),
		ActorRoles:    r,
		IP:            truncate(a.IP, 64),
		UserAgent:     truncate(a.UserAgent, 500),
		BaseVersion:   base,
		Before:        b,
		After:         af,
		Impact:        im,
	}, nil
}

func truncate(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n])
}

// Version is a hash of the live level rows, so apply can detect a concurrent change.
func Version(levels []models.OrgLevel) string {
	rows := make([]string, 0, len(levels))
	for _, l := range levels {
		key := ""
		if l.NameKey != nil {
			key = *l.NameKey
		}
		rows = append(rows, fmt.Sprintf("%s|%d|%s|%s|%t", l.ID, l.Level, l.Name, key, l.IsActive))
	}
	sort.Strings(rows)
	sum := sha256.Sum256([]byte(strings.Join(rows, "\n")))
	return hex.EncodeToString(sum[:])
}

// plan is the computed effect of a ChangeRequest.
type plan struct {
	diff       Diff
	violations []Violation
	changed    map[uuid.UUID]models.OrgLevel // renamed and/or renumbered, final values
	added      []models.OrgLevel
	deleteIDs  []uuid.UUID
	mapping    map[int]int // old level number -> new, renumbered levels only
	affected   []uuid.UUID // departments whose level number or level name changes
}

func (p *plan) violate(v Violation) { p.violations = append(p.violations, v) }

func intPtr(i int) *int { return &i }

// computePlan is pure: current live levels + live departments (with parents)
// + the request -> diff, violations and the write plan.
func computePlan(current []models.OrgLevel, depts []models.Department, req ChangeRequest) *plan {
	p := &plan{changed: map[uuid.UUID]models.OrgLevel{}, mapping: map[int]int{}}
	p.diff = Diff{Added: []LevelRef{}, Renamed: []LevelChange{}, Renumbered: []LevelChange{}, Deleted: []LevelRef{}}

	byID := map[uuid.UUID]models.OrgLevel{}
	for _, l := range current {
		byID[l.ID] = l
	}

	if len(req.Levels) == 0 {
		p.violate(Violation{Code: VLevelsEmpty, Message: "At least one level is required."})
	}
	if len(req.Levels) > orgmode.MaxLevels {
		p.violate(Violation{Code: VTooManyLevels, Message: fmt.Sprintf("At most %d levels are allowed.", orgmode.MaxLevels)})
	}

	deleted := map[uuid.UUID]bool{}
	for _, id := range req.DeleteIDs {
		id := id
		if _, ok := byID[id]; !ok {
			p.violate(Violation{Code: VUnknownLevelID, Message: "A level to delete does not exist.", LevelID: &id})
			continue
		}
		if !deleted[id] {
			deleted[id] = true
			p.deleteIDs = append(p.deleteIDs, id)
		}
	}

	kept := map[uuid.UUID]bool{}
	numbers := map[int]int{}
	names := map[string]int{}
	finalByOld := map[int]int{} // old number -> final number for kept levels
	for _, in := range req.Levels {
		name := strings.TrimSpace(in.Name)
		if in.Level < 1 || in.Level > 999 {
			p.violate(Violation{Code: VLevelOutOfRange, Message: "Level numbers must be between 1 and 999.", Level: intPtr(in.Level), LevelID: in.ID})
		}
		if name == "" {
			p.violate(Violation{Code: VNameRequired, Message: "Every level needs a name.", Level: intPtr(in.Level), LevelID: in.ID})
		} else if utf8.RuneCountInString(name) > MaxNameLength {
			p.violate(Violation{Code: VNameTooLong, Message: fmt.Sprintf("Level names are at most %d characters.", MaxNameLength), Level: intPtr(in.Level), LevelID: in.ID})
		}
		numbers[in.Level]++
		names[strings.ToLower(name)]++

		if in.ID == nil {
			var key *string
			if in.NameKey != nil && *in.NameKey != "" {
				if !models.IsOrgLevelNameKey(*in.NameKey) {
					p.violate(Violation{Code: VInvalidNameKey, Message: "Unknown level name key.", Level: intPtr(in.Level)})
				} else {
					k := *in.NameKey
					key = &k
				}
			}
			l := models.OrgLevel{ID: uuid.New(), Level: in.Level, Name: name, NameKey: key, IsActive: true}
			p.added = append(p.added, l)
			p.diff.Added = append(p.diff.Added, LevelRef{Level: l.Level, Name: l.Name, NameKey: l.NameKey})
			continue
		}

		id := *in.ID
		cur, ok := byID[id]
		if !ok {
			p.violate(Violation{Code: VUnknownLevelID, Message: "A level in the request does not exist.", LevelID: &id})
			continue
		}
		if kept[id] {
			p.violate(Violation{Code: VDuplicateLevelID, Message: "A level appears more than once.", LevelID: &id})
			continue
		}
		kept[id] = true
		if deleted[id] {
			p.violate(Violation{Code: VDeleteConflict, Message: "A level cannot be kept and deleted at once.", LevelID: &id})
		}
		finalByOld[cur.Level] = in.Level
		final := cur
		final.Level = in.Level
		renamed := name != cur.Name
		renumbered := in.Level != cur.Level
		if renamed {
			final.Name = name
			final.NameKey = nil // a user rename clears the preset key (D3)
			p.diff.Renamed = append(p.diff.Renamed, LevelChange{ID: id, FromLevel: cur.Level, ToLevel: in.Level, FromName: cur.Name, ToName: name})
		}
		if renumbered {
			p.mapping[cur.Level] = in.Level
			p.diff.Renumbered = append(p.diff.Renumbered, LevelChange{ID: id, FromLevel: cur.Level, ToLevel: in.Level, FromName: cur.Name, ToName: final.Name})
		}
		if renamed || renumbered {
			p.changed[id] = final
		}
	}
	for n, c := range numbers {
		if c > 1 {
			p.violate(Violation{Code: VDuplicateLevel, Message: "Two levels have the same number.", Level: intPtr(n)})
		}
	}
	for n, c := range names {
		if c > 1 && n != "" {
			p.violate(Violation{Code: VDuplicateName, Message: fmt.Sprintf("Two levels are named %q.", n)})
		}
	}
	for _, l := range current {
		l := l
		if !kept[l.ID] && !deleted[l.ID] {
			p.violate(Violation{Code: VLevelMissing, Message: "Every current level must be kept or deleted.", LevelID: &l.ID, Level: intPtr(l.Level)})
		}
		if deleted[l.ID] {
			p.diff.Deleted = append(p.diff.Deleted, LevelRef{ID: &l.ID, Level: l.Level, Name: l.Name, NameKey: l.NameKey})
		}
	}

	// Data checks against live departments.
	byNumber := map[int]models.OrgLevel{}
	for _, l := range current {
		byNumber[l.Level] = l
	}
	newLevel := func(old int) int {
		if n, ok := finalByOld[old]; ok {
			return n
		}
		return old
	}
	deptByID := map[uuid.UUID]*models.Department{}
	for i := range depts {
		deptByID[depts[i].ID] = &depts[i]
	}
	inUse := map[uuid.UUID][]uuid.UUID{}
	var order []string
	conflicts := map[string][]uuid.UUID{}
	for i := range depts {
		d := &depts[i]
		l, ok := byNumber[d.Level]
		if ok {
			if deleted[l.ID] {
				inUse[l.ID] = append(inUse[l.ID], d.ID)
			} else if _, changed := p.changed[l.ID]; changed {
				p.affected = append(p.affected, d.ID)
			}
		}
		if d.ParentDepartmentID == nil {
			continue
		}
		parent, ok := deptByID[*d.ParentDepartmentID]
		if !ok {
			continue
		}
		if newLevel(d.Level) <= newLevel(parent.Level) {
			key := fmt.Sprintf("%d>%d", parent.Level, d.Level)
			if _, seen := conflicts[key]; !seen {
				order = append(order, key)
			}
			conflicts[key] = append(conflicts[key], d.ID)
		}
	}
	for _, l := range current {
		ids := inUse[l.ID]
		if len(ids) == 0 {
			continue
		}
		l := l
		p.violate(Violation{Code: VLevelInUse, Message: fmt.Sprintf("Level %q is still used by %d unit(s).", l.Name, len(ids)),
			LevelID: &l.ID, Level: intPtr(l.Level), DepartmentIDs: capIDs(ids)})
	}
	for _, key := range order {
		var pl, cl int
		fmt.Sscanf(key, "%d>%d", &pl, &cl)
		p.violate(Violation{Code: VLevelOrderConflict,
			Message: fmt.Sprintf("Units at level %d would no longer be below their parents at level %d.", cl, pl),
			Level:   intPtr(cl), DepartmentIDs: capIDs(conflicts[key])})
	}
	return p
}

func capIDs(ids []uuid.UUID) []uuid.UUID {
	if len(ids) > MaxImpactIDs {
		return append([]uuid.UUID(nil), ids[:MaxImpactIDs]...)
	}
	return ids
}

func (p *plan) preview(base string, employees int64) Preview {
	affected := p.affected
	if affected == nil {
		affected = []uuid.UUID{}
	}
	v := p.violations
	if v == nil {
		v = []Violation{}
	}
	return Preview{
		BaseVersion: base,
		Diff:        p.diff,
		Impact: Impact{
			DepartmentCount: len(affected),
			DepartmentIDs:   capIDs(affected),
			EmployeeCount:   employees,
			UnitUsageIDs:    affected,
		},
		Violations: v,
		CanApply:   len(v) == 0,
	}
}
