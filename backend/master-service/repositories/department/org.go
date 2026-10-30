package department

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"master-service/models"
	"master-service/pkg/database/orgstructure"
	apperrors "master-service/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// OrgRepository is the flexible-mode department access
// (ORG_STRUCTURE_MODE=flexible). It is separate from
// DepartmentRepositoryInterface so the legacy path is untouched, and it is the
// only code that reads or writes departments.parent_department_id, which
// exists on flexible tenants only.
type OrgRepository interface {
	// Transaction runs fn with a repository bound to one transaction.
	Transaction(fn func(tx OrgRepository) error) error
	// DB is the handle (the transaction inside Transaction), for the outbox.
	DB() *gorm.DB

	FindByID(id uuid.UUID) (*models.Department, error)
	// LockByID loads a live department FOR UPDATE (no-op lock on sqlite).
	LockByID(id uuid.UUID) (*models.Department, error)
	// ListLive returns every live department with ParentDepartmentID set,
	// optionally limited to one company.
	ListLive(companyID *uuid.UUID) ([]models.Department, error)
	// ListPage returns one filtered, sorted page and the total count.
	ListPage(f ListFilter) ([]models.Department, int64, error)

	NameTaken(companyID uuid.UUID, parentID *uuid.UUID, name string, excludeID *uuid.UUID) (bool, error)
	CodeTaken(code string, companyID uuid.UUID, excludeID *uuid.UUID) (bool, error)
	HasLiveChildren(id uuid.UUID) (bool, error)

	// Create and Update persist the department and its parent.
	Create(d *models.Department) error
	Update(d *models.Department) error
	Delete(id uuid.UUID) error

	// ActiveLevels returns live, active org levels ordered by level.
	ActiveLevels() ([]models.OrgLevel, error)
}

// ListFilter is the flexible GET /departments query. Order columns must be
// validated with ParseOrder.
type ListFilter struct {
	Offset, Limit int
	Search        string
	CompanyID     *uuid.UUID
	Level         *int
	ParentID      *uuid.UUID
	RootsOnly     bool // parent_id=root
	IsActive      *bool
	Order         clause.OrderBy
}

// SortableColumns is the sort whitelist of the flexible list (spec section 6).
var SortableColumns = []string{"department_name", "department_code", "level", "created_at"}

var departmentColumns = func() map[string]bool {
	cols := map[string]bool{}
	s, err := schema.Parse(&models.Department{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		return cols
	}
	for _, f := range s.Fields {
		if f.DBName != "" {
			cols[f.DBName] = true
		}
	}
	return cols
}()

// ParseOrder turns "column [ASC|DESC][, ...]" into an ORDER BY. A column must
// be in SortableColumns AND in the GORM schema of models.Department; it is
// passed as a quoted clause.Column, never interpolated. Empty means
// department_name ASC. id is appended as a tiebreaker.
func ParseOrder(raw string) (clause.OrderBy, bool) {
	var ob clause.OrderBy
	if strings.TrimSpace(raw) == "" {
		raw = "department_name ASC"
	}
	allowed := map[string]bool{}
	for _, c := range SortableColumns {
		allowed[c] = departmentColumns[c]
	}
	for _, part := range strings.Split(raw, ",") {
		fields := strings.Fields(part)
		if len(fields) == 0 || len(fields) > 2 {
			return ob, false
		}
		col := strings.ToLower(fields[0])
		if !allowed[col] {
			return ob, false
		}
		desc := false
		if len(fields) == 2 {
			switch strings.ToUpper(fields[1]) {
			case "ASC":
			case "DESC":
				desc = true
			default:
				return ob, false
			}
		}
		ob.Columns = append(ob.Columns, clause.OrderByColumn{Column: clause.Column{Name: col}, Desc: desc})
	}
	ob.Columns = append(ob.Columns, clause.OrderByColumn{Column: clause.Column{Name: "id"}})
	return ob, true
}

type orgRepository struct {
	db *gorm.DB
}

// NewOrgRepository returns the flexible-mode repository. Only construct it when
// the effective mode is flexible.
func NewOrgRepository(db *gorm.DB) OrgRepository {
	return &orgRepository{db: db}
}

func (r *orgRepository) DB() *gorm.DB { return r.db }

func (r *orgRepository) Transaction(fn func(tx OrgRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fn(&orgRepository{db: tx})
	})
}

func notFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperrors.ErrNotFound
	}
	return err
}

func (r *orgRepository) hydrate(depts []models.Department) error {
	if len(depts) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, len(depts))
	for i := range depts {
		ids[i] = depts[i].ID
	}
	parents := map[uuid.UUID]*uuid.UUID{}
	for start := 0; start < len(ids); start += 500 {
		end := min(start+500, len(ids))
		var rows []models.DepartmentOrg
		if err := r.db.Model(&models.DepartmentOrg{}).Select("id", "parent_department_id").
			Where("id IN ?", ids[start:end]).Find(&rows).Error; err != nil {
			return err
		}
		for _, row := range rows {
			parents[row.ID] = row.ParentDepartmentID
		}
	}
	for i := range depts {
		depts[i].ParentDepartmentID = parents[depts[i].ID]
	}
	return nil
}

func (r *orgRepository) FindByID(id uuid.UUID) (*models.Department, error) {
	var d models.Department
	if err := r.db.Where("id = ?", id).First(&d).Error; err != nil {
		return nil, notFound(err)
	}
	one := []models.Department{d}
	if err := r.hydrate(one); err != nil {
		return nil, err
	}
	return &one[0], nil
}

func (r *orgRepository) LockByID(id uuid.UUID) (*models.Department, error) {
	var d models.Department
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&d).Error; err != nil {
		return nil, notFound(err)
	}
	one := []models.Department{d}
	if err := r.hydrate(one); err != nil {
		return nil, err
	}
	return &one[0], nil
}

func (r *orgRepository) ListLive(companyID *uuid.UUID) ([]models.Department, error) {
	var depts []models.Department
	q := r.db.Model(&models.Department{})
	if companyID != nil {
		q = q.Where("company_id = ?", *companyID)
	}
	if err := q.Order("department_name ASC").Order("id ASC").Find(&depts).Error; err != nil {
		return nil, err
	}
	if err := r.hydrate(depts); err != nil {
		return nil, err
	}
	return depts, nil
}

func (r *orgRepository) ListPage(f ListFilter) ([]models.Department, int64, error) {
	q := r.db.Model(&models.Department{})
	if f.Search != "" {
		p := "%" + strings.ToLower(f.Search) + "%"
		q = q.Where("lower(department_code) LIKE ? OR lower(department_name) LIKE ? OR lower(department_description) LIKE ?", p, p, p)
	}
	if f.CompanyID != nil {
		q = q.Where("company_id = ?", *f.CompanyID)
	}
	if f.Level != nil {
		q = q.Where("level = ?", *f.Level)
	}
	if f.IsActive != nil {
		q = q.Where("is_active = ?", *f.IsActive)
	}
	if f.RootsOnly {
		q = q.Where("parent_department_id IS NULL")
	} else if f.ParentID != nil {
		q = q.Where("parent_department_id = ?", *f.ParentID)
	}
	var total int64
	if err := q.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var depts []models.Department
	if len(f.Order.Columns) > 0 {
		q = q.Clauses(f.Order)
	}
	if err := q.Offset(f.Offset).Limit(f.Limit).Find(&depts).Error; err != nil {
		return nil, 0, err
	}
	if err := r.hydrate(depts); err != nil {
		return nil, 0, err
	}
	return depts, total, nil
}

func (r *orgRepository) NameTaken(companyID uuid.UUID, parentID *uuid.UUID, name string, excludeID *uuid.UUID) (bool, error) {
	q := r.db.Model(&models.Department{}).
		Where("company_id = ? AND lower(department_name) = ?", companyID, strings.ToLower(strings.TrimSpace(name)))
	if parentID == nil {
		q = q.Where("parent_department_id IS NULL")
	} else {
		q = q.Where("parent_department_id = ?", *parentID)
	}
	if excludeID != nil {
		q = q.Where("id <> ?", *excludeID)
	}
	var n int64
	err := q.Count(&n).Error
	return n > 0, err
}

// CodeTaken applies orgstructure.DepartmentCodeScope (decision D10), the same
// policy as the gated unique index.
func (r *orgRepository) CodeTaken(code string, companyID uuid.UUID, excludeID *uuid.UUID) (bool, error) {
	q := r.db.Model(&models.Department{}).Where("department_code = ?", code)
	if orgstructure.DepartmentCodeScope == orgstructure.CodeScopePerCompany {
		q = q.Where("company_id = ?", companyID)
	}
	if excludeID != nil {
		q = q.Where("id <> ?", *excludeID)
	}
	var n int64
	err := q.Count(&n).Error
	return n > 0, err
}

func (r *orgRepository) HasLiveChildren(id uuid.UUID) (bool, error) {
	var n int64
	err := r.db.Model(&models.Department{}).Where("parent_department_id = ?", id).Count(&n).Error
	return n > 0, err
}

func (r *orgRepository) writeParent(d *models.Department) error {
	return r.db.Model(&models.DepartmentOrg{}).Where("id = ?", d.ID).
		Update("parent_department_id", d.ParentDepartmentID).Error
}

func (r *orgRepository) Create(d *models.Department) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	if err := r.db.Omit(clause.Associations).Create(d).Error; err != nil {
		return dbError(err)
	}
	if d.ParentDepartmentID != nil {
		if err := r.writeParent(d); err != nil {
			return dbError(err)
		}
	}
	return nil
}

func (r *orgRepository) Update(d *models.Department) error {
	if err := r.db.Omit(clause.Associations).Save(d).Error; err != nil {
		return dbError(err)
	}
	if err := r.writeParent(d); err != nil {
		return dbError(err)
	}
	return nil
}

func (r *orgRepository) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&models.Department{ID: id}).Error; err != nil {
		return dbError(err)
	}
	return nil
}

func (r *orgRepository) ActiveLevels() ([]models.OrgLevel, error) {
	var levels []models.OrgLevel
	err := r.db.Where("is_active = ?", true).Order("level ASC").Find(&levels).Error
	return levels, err
}

// dbError keeps the driver error in the chain (so constraint names map to API
// codes) under ErrDatabase, like repositories.BaseRepository.
func dbError(err error) error {
	return fmt.Errorf("%w: %w", apperrors.ErrDatabase, err)
}
