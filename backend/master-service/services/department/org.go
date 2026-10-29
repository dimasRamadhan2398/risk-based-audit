package department

import (
	"context"
	"net/http"
	"strings"

	"master-service/models"
	"master-service/pkg/base"
	apperrors "master-service/pkg/errors"
	"master-service/pkg/kafka"
	"master-service/pkg/orgmode"
	repo "master-service/repositories/department"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrgEventPublisher records flexible-mode department events inside the
// caller's transaction. *kafka.DomainPublisher satisfies it; nil disables events.
type OrgEventPublisher interface {
	PublishDepartmentOrgTx(ctx context.Context, tx *gorm.DB, eventType string, dept *models.Department) error
}

// OrgDepartmentServiceInterface is the flexible-mode department service. It
// also satisfies DepartmentServiceInterface, so it is what the registry hands
// out on flexible tenants; controllers detect it by type assertion.
type OrgDepartmentServiceInterface interface {
	DepartmentServiceInterface
	List(ctx *base.BaseService, f repo.ListFilter) ([]models.Department, int64, error)
	Tree(ctx *base.BaseService, companyID *uuid.UUID) ([]*models.DepartmentTreeNode, error)
	// Resync queues department.updated (with org fields) for every live department.
	Resync(ctx context.Context) (int, error)
}

// OrgDepartmentService implements the rules of docs/org-structure-model-plan.md
// 4.2 and 4.3. Every write runs in one transaction together with its outbox
// events.
type OrgDepartmentService struct {
	legacy *DepartmentService // reference checks (company, business unit, PIC)
	org    repo.OrgRepository
	events OrgEventPublisher
}

// NewOrgDepartmentService builds the flexible service. Construct it only when
// the effective mode is flexible.
func NewOrgDepartmentService(departmentRepo repo.IDepartmentRepository, org repo.OrgRepository, events OrgEventPublisher) OrgDepartmentServiceInterface {
	return &OrgDepartmentService{
		legacy: &DepartmentService{departmentRepo: departmentRepo},
		org:    org,
		events: events,
	}
}

func orgErr(code, message string, status int, field, fieldCode string) *apperrors.AppError {
	return fieldErr(code, message, status, field, fieldCode)
}

var (
	errLevelInvalid = orgErr(apperrors.CodeOrgLevelInvalid,
		"The selected level is not an active organisation level.", http.StatusUnprocessableEntity, "level", apperrors.FieldInvalidOption)
	errParentNotFound = orgErr(apperrors.CodeParentDepartmentNotFound,
		"The selected parent unit does not exist.", http.StatusUnprocessableEntity, "parent_department_id", apperrors.FieldNotFound)
	errParentCompany = orgErr(apperrors.CodeParentDepartmentCompanyMismatch,
		"The parent unit belongs to another company.", http.StatusUnprocessableEntity, "parent_department_id", apperrors.FieldMismatch)
	errLevelNotBelowParent = orgErr(apperrors.CodeDepartmentLevelNotBelowParent,
		"A unit must be at a lower level than its parent.", http.StatusUnprocessableEntity, "level", apperrors.FieldInvalid)
	errLevelChildConflict = orgErr(apperrors.CodeDepartmentLevelChildConflict,
		"Every sub-unit must stay at a lower level than this unit.", http.StatusUnprocessableEntity, "level", apperrors.FieldInvalid)
	errParentCycle = orgErr(apperrors.CodeDepartmentParentCycle,
		"A unit cannot be placed under itself or one of its sub-units.", http.StatusUnprocessableEntity, "parent_department_id", apperrors.FieldInvalid)
	errDepth = orgErr(apperrors.CodeOrgDepthExceeded,
		"The organisation tree cannot be deeper than 6 levels.", http.StatusUnprocessableEntity, "parent_department_id", apperrors.FieldOutOfRange)
	errChildrenCompany = orgErr(apperrors.CodeDepartmentChildrenCompanyMismatch,
		"A unit with sub-units cannot move to another company.", http.StatusUnprocessableEntity, "company_id", apperrors.FieldMismatch)
	errHasChildren = apperrors.New(apperrors.CodeDepartmentHasChildren,
		"This unit still has sub-units. Move or delete them first.", http.StatusConflict)
	errScopedNameExists = apperrors.New(apperrors.CodeDepartmentNameAlreadyExists,
		"A unit with this name already exists under the same parent.", http.StatusConflict).
		WithFields(map[string]string{"department_name": apperrors.FieldAlreadyExists})
)

func (s *OrgDepartmentService) publish(tx repo.OrgRepository, eventType string, depts ...*models.Department) error {
	if s.events == nil {
		return nil
	}
	for _, d := range depts {
		if d == nil {
			continue
		}
		if err := s.events.PublishDepartmentOrgTx(context.Background(), tx.DB(), eventType, d); err != nil {
			return apperrors.Internal("Failed to record the department event.", err)
		}
	}
	return nil
}

func (s *OrgDepartmentService) index(r repo.OrgRepository, companyID *uuid.UUID) (*Index, error) {
	depts, err := r.ListLive(companyID)
	if err != nil {
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to load departments.")
	}
	levels, err := r.ActiveLevels()
	if err != nil {
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to load organisation levels.")
	}
	return NewIndex(depts, levels), nil
}

func readErr(err error) error {
	if err == apperrors.ErrNotFound {
		return apperrors.ErrDepartmentNotFound
	}
	return apperrors.DB(err, apperrors.OpRead, "Failed to load the department.")
}

// checkPlacement validates level, parent and depth for d (id may be in ix as
// the unit being updated).
func checkPlacement(ix *Index, d *models.Department, parent *models.Department, updating bool) error {
	if _, ok := ix.Level(d.Level); !ok {
		return errLevelInvalid
	}
	height := 1
	if updating {
		for _, c := range ix.Children(d.ID) {
			if child, ok := ix.Get(c); ok && child.Level <= d.Level {
				return errLevelChildConflict
			}
		}
		height = ix.Height(d.ID)
	}
	if parent == nil {
		if height > orgmode.MaxLevels {
			return errDepth
		}
		return nil
	}
	if updating && (parent.ID == d.ID || contains(ix.Descendants(d.ID), parent.ID)) {
		return errParentCycle
	}
	if parent.CompanyID != d.CompanyID {
		return errParentCompany
	}
	if parent.Level >= d.Level {
		return errLevelNotBelowParent
	}
	if ix.Depth(parent.ID)+height > orgmode.MaxLevels {
		return errDepth
	}
	return nil
}

func contains(ids []uuid.UUID, id uuid.UUID) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}

func sameParent(a, b *uuid.UUID) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

// loadParent locks the parent row (so it cannot be deleted concurrently) and
// returns it from the index.
func loadParent(tx repo.OrgRepository, ix *Index, parentID *uuid.UUID) (*models.Department, error) {
	if parentID == nil {
		return nil, nil
	}
	p, err := tx.LockByID(*parentID)
	if err != nil {
		if err == apperrors.ErrNotFound {
			return nil, errParentNotFound
		}
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to load the parent unit.")
	}
	if indexed, ok := ix.Get(p.ID); ok {
		return indexed, nil
	}
	return p, nil
}

// Create implements DepartmentServiceInterface.
func (s *OrgDepartmentService) Create(ctx *base.BaseService, d *models.Department) (*models.Department, error) {
	if err := s.legacy.validateReferences(d); err != nil {
		return nil, err
	}
	d.DepartmentName = strings.TrimSpace(d.DepartmentName)
	var out *models.Department
	err := s.org.Transaction(func(tx repo.OrgRepository) error {
		ix, err := s.index(tx, &d.CompanyID)
		if err != nil {
			return err
		}
		parent, err := loadParent(tx, ix, d.ParentDepartmentID)
		if err != nil {
			return err
		}
		if err := checkPlacement(ix, d, parent, false); err != nil {
			return err
		}
		if taken, err := tx.CodeTaken(d.DepartmentCode, d.CompanyID, nil); err != nil {
			return apperrors.DB(err, apperrors.OpRead, "Failed to validate the department code.")
		} else if taken {
			return errDepartmentCodeExists
		}
		if taken, err := tx.NameTaken(d.CompanyID, d.ParentDepartmentID, d.DepartmentName, nil); err != nil {
			return apperrors.DB(err, apperrors.OpRead, "Failed to validate the department name.")
		} else if taken {
			return errScopedNameExists
		}
		if err := tx.Create(d); err != nil {
			return apperrors.DB(err, apperrors.OpCreate, "Failed to create the department.")
		}
		ix.Put(*d)
		out = ix.Decorated(d.ID)
		return s.publish(tx, kafka.DomainEventDepartmentCreated, out)
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Update implements DepartmentServiceInterface. d carries the full desired
// state, including ParentDepartmentID.
func (s *OrgDepartmentService) Update(ctx *base.BaseService, id string, d *models.Department) (*models.Department, error) {
	departmentID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.InvalidID(err)
	}
	if err := s.legacy.validateReferences(d); err != nil {
		return nil, err
	}
	d.DepartmentName = strings.TrimSpace(d.DepartmentName)
	var out *models.Department
	err = s.org.Transaction(func(tx repo.OrgRepository) error {
		existing, err := tx.LockByID(departmentID)
		if err != nil {
			return readErr(err)
		}
		d.ID = existing.ID
		ix, err := s.index(tx, &existing.CompanyID)
		if err != nil {
			return err
		}
		if d.CompanyID != existing.CompanyID {
			if len(ix.Children(existing.ID)) > 0 {
				return errChildrenCompany
			}
			if ix, err = s.index(tx, &d.CompanyID); err != nil {
				return err
			}
			ix.Put(*existing)
		}
		parent, err := loadParent(tx, ix, d.ParentDepartmentID)
		if err != nil {
			return err
		}
		if err := checkPlacement(ix, d, parent, true); err != nil {
			return err
		}
		if d.DepartmentCode != existing.DepartmentCode {
			if taken, err := tx.CodeTaken(d.DepartmentCode, d.CompanyID, &existing.ID); err != nil {
				return apperrors.DB(err, apperrors.OpRead, "Failed to validate the department code.")
			} else if taken {
				return errDepartmentCodeExists
			}
		}
		moved := !sameParent(existing.ParentDepartmentID, d.ParentDepartmentID) || d.CompanyID != existing.CompanyID
		renamed := existing.DepartmentName != d.DepartmentName
		relevelled := existing.Level != d.Level
		if moved || renamed {
			if taken, err := tx.NameTaken(d.CompanyID, d.ParentDepartmentID, d.DepartmentName, &existing.ID); err != nil {
				return apperrors.DB(err, apperrors.OpRead, "Failed to validate the department name.")
			} else if taken {
				return errScopedNameExists
			}
		}

		row := *existing
		row.DepartmentCode = d.DepartmentCode
		row.DepartmentName = d.DepartmentName
		row.DepartmentDescription = d.DepartmentDescription
		row.PicID = d.PicID
		row.Level = d.Level
		row.IsActive = d.IsActive
		row.CompanyID = d.CompanyID
		row.BusinessUnitID = d.BusinessUnitID
		row.ParentDepartmentID = d.ParentDepartmentID
		row.Path, row.LevelName, row.NameKey = nil, "", nil
		if err := tx.Update(&row); err != nil {
			return apperrors.DB(err, apperrors.OpUpdate, "Failed to update the department.")
		}
		ix.Put(row)
		out = ix.Decorated(row.ID)

		// D5: a rename, move or level change alters the path of every
		// descendant, so the whole subtree is republished (depth <= 6).
		events := []*models.Department{out}
		if moved || renamed || relevelled {
			for _, c := range ix.Descendants(row.ID) {
				events = append(events, ix.Decorated(c))
			}
		}
		return s.publish(tx, kafka.DomainEventDepartmentUpdated, events...)
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Delete implements DepartmentServiceInterface. Refused while the unit has
// live sub-units.
func (s *OrgDepartmentService) Delete(ctx *base.BaseService, id string) error {
	departmentID, err := uuid.Parse(id)
	if err != nil {
		return apperrors.InvalidID(err)
	}
	return s.org.Transaction(func(tx repo.OrgRepository) error {
		existing, err := tx.LockByID(departmentID)
		if err != nil {
			return readErr(err)
		}
		has, err := tx.HasLiveChildren(existing.ID)
		if err != nil {
			return apperrors.DB(err, apperrors.OpRead, "Failed to check sub-units.")
		}
		if has {
			return errHasChildren
		}
		ix, err := s.index(tx, &existing.CompanyID)
		if err != nil {
			return err
		}
		dec := ix.Decorated(existing.ID)
		if err := tx.Delete(existing.ID); err != nil {
			return apperrors.DB(err, apperrors.OpDelete, "Failed to delete the department.")
		}
		return s.publish(tx, kafka.DomainEventDepartmentDeleted, dec)
	})
}

// FindById implements DepartmentServiceInterface.
func (s *OrgDepartmentService) FindById(ctx *base.BaseService, id string) (*models.Department, error) {
	departmentID, err := uuid.Parse(id)
	if err != nil {
		return nil, apperrors.InvalidID(err)
	}
	d, err := s.org.FindByID(departmentID)
	if err != nil {
		return nil, readErr(err)
	}
	ix, err := s.index(s.org, &d.CompanyID)
	if err != nil {
		return nil, err
	}
	if dec := ix.Decorated(d.ID); dec != nil {
		return dec, nil
	}
	return d, nil
}

// FindAll implements DepartmentServiceInterface.
func (s *OrgDepartmentService) FindAll(ctx *base.BaseService) (*[]models.Department, error) {
	ix, err := s.index(s.org, nil)
	if err != nil {
		return nil, err
	}
	depts, err := s.org.ListLive(nil)
	if err != nil {
		return nil, apperrors.DB(err, apperrors.OpRead, "Failed to load departments.")
	}
	for i := range depts {
		ix.Decorate(&depts[i])
	}
	return &depts, nil
}

// FindMany implements DepartmentServiceInterface (unfiltered page).
func (s *OrgDepartmentService) FindMany(ctx *base.BaseService, offset, limit int, search string) (*[]models.Department, int64, error) {
	order, _ := repo.ParseOrder("")
	depts, total, err := s.List(ctx, repo.ListFilter{Offset: offset, Limit: limit, Search: search, Order: order})
	if err != nil {
		return nil, 0, err
	}
	return &depts, total, nil
}

// List returns one filtered page with level names and paths.
func (s *OrgDepartmentService) List(ctx *base.BaseService, f repo.ListFilter) ([]models.Department, int64, error) {
	depts, total, err := s.org.ListPage(f)
	if err != nil {
		return nil, 0, apperrors.DB(err, apperrors.OpRead, "Failed to load departments.")
	}
	ix, err := s.index(s.org, f.CompanyID)
	if err != nil {
		return nil, 0, err
	}
	for i := range depts {
		ix.Decorate(&depts[i])
	}
	if depts == nil {
		depts = []models.Department{}
	}
	return depts, total, nil
}

// Tree returns the nested tree, optionally for one company.
func (s *OrgDepartmentService) Tree(ctx *base.BaseService, companyID *uuid.UUID) ([]*models.DepartmentTreeNode, error) {
	ix, err := s.index(s.org, companyID)
	if err != nil {
		return nil, err
	}
	return ix.Tree(), nil
}

// Resync queues department.updated with org fields for every live department.
func (s *OrgDepartmentService) Resync(ctx context.Context) (int, error) {
	n := 0
	err := s.org.Transaction(func(tx repo.OrgRepository) error {
		ix, err := s.index(tx, nil)
		if err != nil {
			return err
		}
		depts, err := tx.ListLive(nil)
		if err != nil {
			return err
		}
		for i := range depts {
			if err := s.publish(tx, kafka.DomainEventDepartmentUpdated, ix.Decorated(depts[i].ID)); err != nil {
				return err
			}
			n++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return n, nil
}
