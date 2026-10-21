package activityplanform

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"audit-service/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PKAT is a loaded annual plan with its rows indexed by row id.
type PKAT struct {
	Plan  *models.AuditAnnual
	byID  map[string]int
	Order []string // row ids in PKAT order
}

// NewPKAT indexes plan.Activities by id, after EnsureRowIDs.
func NewPKAT(plan *models.AuditAnnual) *PKAT {
	EnsureRowIDs(plan)
	p := &PKAT{Plan: plan, byID: make(map[string]int, len(plan.Activities))}
	for i, row := range plan.Activities {
		p.byID[row.ID] = i
		p.Order = append(p.Order, row.ID)
	}
	return p
}

// EnsureRowIDs gives every PKAT row of plan a canonical UUID in memory. Row ids
// are only assigned by the server since 2026-10-14, so older plans have rows
// without one. Such a row (and a row whose id is not a UUID or repeats an
// earlier row's id) gets a deterministic id derived from the plan id and the
// row position, so every read of the same stored plan yields the same ids
// without rewriting stored data. GET /annual-audit-plans returns these ids, the
// PKAT form sends them back, and the next PKAT save stores them (the update
// hook keeps a valid client id), after which they are permanent.
func EnsureRowIDs(plan *models.AuditAnnual) {
	seen := make(map[string]bool, len(plan.Activities))
	for i := range plan.Activities {
		id := ""
		if u, err := uuid.Parse(strings.TrimSpace(plan.Activities[i].ID)); err == nil && u != uuid.Nil {
			id = u.String()
		}
		if id == "" || seen[id] {
			id = uuid.NewSHA1(plan.ID, []byte("pkat-row:"+strconv.Itoa(i))).String()
		}
		seen[id] = true
		plan.Activities[i].ID = id
	}
}

// RowKey is the canonical form of a PKAT row id: a lower-case UUID when it
// parses as one, else the trimmed value.
func RowKey(id string) string {
	id = strings.TrimSpace(id)
	if u, err := uuid.Parse(id); err == nil {
		return u.String()
	}
	return id
}

// Row returns the PKAT row with the given id and its 0-based position.
func (p *PKAT) Row(id string) (*models.AnnualAuditActivity, int, bool) {
	if p == nil {
		return nil, -1, false
	}
	i, ok := p.byID[RowKey(id)]
	if !ok {
		return nil, -1, false
	}
	return &p.Plan.Activities[i], i, true
}

// Missing returns the ids that are not rows of this PKAT, in the given order.
func (p *PKAT) Missing(ids []string) []string {
	var out []string
	for _, id := range ids {
		if _, _, ok := p.Row(id); !ok {
			out = append(out, id)
		}
	}
	return out
}

// LoadPKAT reads a live annual plan. With lock the row is locked FOR UPDATE
// until the transaction ends, which serialises RKA saves against PKAT updates
// and deletes (no-op on sqlite). A missing or soft-deleted plan gives 404
// ANNUAL_PLAN_NOT_FOUND.
func LoadPKAT(tx *gorm.DB, id uuid.UUID, lock bool) (*PKAT, error) {
	q := tx
	if lock {
		q = q.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var plan models.AuditAnnual
	if err := q.First(&plan, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, Fail(CodeAnnualPlanNotFound, http.StatusNotFound, "annual audit plan (PKAT) not found", Details{Field: "annualPlanId"})
		}
		return nil, err
	}
	return NewPKAT(&plan), nil
}

// loadPKATs reads annual plans by id, soft-deleted ones included (deleted is
// true for those), for read views.
func loadPKATs(db *gorm.DB, ids []uuid.UUID) (map[uuid.UUID]*PKAT, map[uuid.UUID]bool, error) {
	out := map[uuid.UUID]*PKAT{}
	deleted := map[uuid.UUID]bool{}
	if len(ids) == 0 {
		return out, deleted, nil
	}
	var plans []models.AuditAnnual
	if err := db.Unscoped().Where("id IN ?", ids).Find(&plans).Error; err != nil {
		return nil, nil, err
	}
	for i := range plans {
		out[plans[i].ID] = NewPKAT(&plans[i])
		if plans[i].DeletedAt.Valid {
			deleted[plans[i].ID] = true
		}
	}
	return out, deleted, nil
}
