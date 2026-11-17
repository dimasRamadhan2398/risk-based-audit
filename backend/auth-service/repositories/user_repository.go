package repositories

import (
	"strings"

	"auth-service/models"
	apperrors "auth-service/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// IUserRepository is an alias for the user repository interface
type IUserRepository = UserRepositoryInterface

// UserRepositoryInterface defines the user repository interface
type UserRepositoryInterface interface {
	Create(user *models.User) error
	Update(user *models.User) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindByEmployeeID(employeeID string) (*models.User, error)
	FindMany(offset, limit int, search, department string, isActive *bool) ([]*models.User, error)
	Count(search, department string, isActive *bool) (int64, error)
	AssignRoles(userID uuid.UUID, roleNames []string) error
	FindAssignable(offset, limit int, search string, id *uuid.UUID) ([]*models.User, int64, error)
	FindContacts(ids []uuid.UUID, departmentHeadOf string) ([]*models.User, error)

	// WithSeatLock runs fn in one transaction holding the seat lock
	// (pg_advisory_xact_lock(SeatLockKey) on Postgres, nothing elsewhere). fn
	// gets a repository bound to that transaction; returning an error rolls
	// everything back. Every write that can change the seat count or the
	// number of active admins goes through here.
	WithSeatLock(fn func(tx UserRepositoryInterface) error) error
	// CountSeats counts active, non-deleted users holding at least one role
	// other than VIEWER (each user once).
	CountSeats() (int64, error)
	// CountActiveAdmins counts active, non-deleted users holding ADMIN.
	CountActiveAdmins() (int64, error)
	// SetActive sets users.is_active of one user.
	SetActive(id uuid.UUID, active bool) error
	// HardDelete permanently removes a user (soft-deleted or not) and the rows
	// that reference it: role links, MFA setups, trusted devices, refresh
	// tokens and confidentiality agreements. Activity log rows are kept;
	// their legacy sys_alert_logs.user_id is cleared. Returns ErrNotFound when
	// there is no such user.
	HardDelete(id uuid.UUID) error
}

// SeatLockKey is the fixed pg_advisory_xact_lock key of the seat lock. Each
// tenant has its own database, so one key per database is enough.
const SeatLockKey int64 = 7_412_993_050_001

// seatRoleSubquery selects the users holding a role other than VIEWER.
const seatRoleSubquery = "id IN (SELECT ur.user_id FROM user_roles ur JOIN roles ro ON ro.id = ur.role_id WHERE UPPER(TRIM(ro.name)) <> ?)"

// adminRoleSubquery selects the users holding ADMIN.
const adminRoleSubquery = "id IN (SELECT ur.user_id FROM user_roles ur JOIN roles ro ON ro.id = ur.role_id WHERE UPPER(TRIM(ro.name)) = ?)"

// WithSeatLock implements UserRepositoryInterface.
func (r *UserRepository) WithSeatLock(fn func(tx UserRepositoryInterface) error) error {
	return r.GetDB().Transaction(func(tx *gorm.DB) error {
		// Transaction-level lock: released at commit or rollback. SQLite (tests)
		// has no advisory locks; there the lock is a no-op.
		if tx.Dialector.Name() == "postgres" {
			if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", SeatLockKey).Error; err != nil {
				return err
			}
		}
		return fn(&UserRepository{BaseRepository: NewBaseRepository(tx)})
	})
}

// CountSeats implements UserRepositoryInterface.
func (r *UserRepository) CountSeats() (int64, error) {
	var n int64
	err := r.GetDB().Model(&models.User{}).
		Where("is_active = ?", true).
		Where(seatRoleSubquery, "VIEWER").
		Count(&n).Error
	return n, err
}

// CountActiveAdmins implements UserRepositoryInterface.
func (r *UserRepository) CountActiveAdmins() (int64, error) {
	var n int64
	err := r.GetDB().Model(&models.User{}).
		Where("is_active = ?", true).
		Where(adminRoleSubquery, "ADMIN").
		Count(&n).Error
	return n, err
}

// SetActive implements UserRepositoryInterface.
func (r *UserRepository) SetActive(id uuid.UUID, active bool) error {
	res := r.GetDB().Model(&models.User{}).Where("id = ?", id).Update("is_active", active)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// HardDelete implements UserRepositoryInterface. Callers run it inside
// WithSeatLock so all deletes commit or roll back together.
func (r *UserRepository) HardDelete(id uuid.UUID) error {
	db := r.GetDB()
	var n int64
	if err := db.Unscoped().Model(&models.User{}).Where("id = ?", id).Count(&n).Error; err != nil {
		return err
	}
	if n == 0 {
		return apperrors.ErrNotFound
	}
	steps := []func() error{
		func() error { return db.Exec("DELETE FROM user_roles WHERE user_id = ?", id).Error },
		func() error { return db.Where("user_id = ?", id).Delete(&models.MFASetup{}).Error },
		func() error { return db.Where("user_id = ?", id).Delete(&models.TrustedDevice{}).Error },
		func() error { return db.Where("user_id = ?", id).Delete(&models.RefreshToken{}).Error },
		func() error {
			return db.Unscoped().Where("user_id = ?", id).Delete(&models.ConfidentialityAgreement{}).Error
		},
		func() error {
			return db.Model(&models.SysAlertLog{}).Where("user_id = ?", id).UpdateColumn("user_id", nil).Error
		},
		func() error { return db.Unscoped().Where("id = ?", id).Delete(&models.User{}).Error },
	}
	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}

// UserRepository handles user data operations
type UserRepository struct {
	*BaseRepository
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) IUserRepository {
	return &UserRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	return r.BaseRepository.Create(user)
}

// Update updates a user
func (r *UserRepository) Update(user *models.User) error {
	return r.BaseRepository.Update(user)
}

// Delete deletes a user
func (r *UserRepository) Delete(id uuid.UUID) error {
	return r.BaseRepository.Delete(&models.User{ID: id})
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.GetDB().Preload("Roles").First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// FindByUsername finds a user by username
func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.GetDB().Preload("Roles").Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.GetDB().Preload("Roles").Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmployeeID finds a user by the employee code it is linked to
func (r *UserRepository) FindByEmployeeID(employeeID string) (*models.User, error) {
	var user models.User
	if err := r.GetDB().Preload("Roles").Where("employee_id = ?", employeeID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// FindMany finds multiple users with filters
func (r *UserRepository) FindMany(offset, limit int, search, department string, isActive *bool) ([]*models.User, error) {
	var users []*models.User
	query := r.GetDB().Model(&models.User{}).Preload("Roles")

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("username LIKE ? OR email LIKE ? OR full_name LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if department != "" {
		query = query.Where("department = ?", department)
	}

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

// Count counts users with filters
func (r *UserRepository) Count(search, department string, isActive *bool) (int64, error) {
	var count int64
	query := r.GetDB().Model(&models.User{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("username LIKE ? OR email LIKE ? OR full_name LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if department != "" {
		query = query.Where("department = ?", department)
	}

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// FindAssignable returns one page of the active users that can be picked as
// the PIC of a follow-up item, and the total. search matches full_name,
// department or position case-insensitively (LIKE wildcards in it are taken
// literally); id, when set, restricts the result to that user. Only the
// columns the picker needs are read. Ordered by full_name, then id, so pages
// are stable.
func (r *UserRepository) FindAssignable(offset, limit int, search string, id *uuid.UUID) ([]*models.User, int64, error) {
	query := r.GetDB().Model(&models.User{}).Where("is_active = ?", true)
	if id != nil {
		query = query.Where("id = ?", *id)
	}
	if s := strings.TrimSpace(search); s != "" {
		pattern := "%" + likeEscaper.Replace(strings.ToLower(s)) + "%"
		query = query.Where(`(LOWER(full_name) LIKE ? ESCAPE '\' OR LOWER(department) LIKE ? ESCAPE '\' OR LOWER(position) LIKE ? ESCAPE '\')`,
			pattern, pattern, pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []*models.User
	err := query.Select("id", "full_name", "department", "position").
		Order(clause.OrderBy{Columns: []clause.OrderByColumn{
			{Column: clause.Column{Name: "full_name"}},
			{Column: clause.Column{Name: "id"}},
		}}).
		Offset(offset).Limit(limit).
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// FindContacts returns the active users matching the filters, with only the
// contact columns. ids restricts to those users; departmentHeadOf restricts to
// users holding the DEPARTMENT_HEAD role in that department (case-insensitive,
// trimmed). Both filters apply when both are set; the caller must set one.
// Ordered by full_name, then id.
func (r *UserRepository) FindContacts(ids []uuid.UUID, departmentHeadOf string) ([]*models.User, error) {
	query := r.GetDB().Model(&models.User{}).Where("is_active = ?", true)
	if len(ids) > 0 {
		query = query.Where("id IN ?", ids)
	}
	if dept := strings.TrimSpace(departmentHeadOf); dept != "" {
		query = query.
			Where("LOWER(TRIM(department)) = ?", strings.ToLower(dept)).
			Where("id IN (SELECT ur.user_id FROM user_roles ur JOIN roles ro ON ro.id = ur.role_id WHERE UPPER(ro.name) = ?)", "DEPARTMENT_HEAD")
	}

	var users []*models.User
	err := query.Select("id", "full_name", "email", "department").
		Order(clause.OrderBy{Columns: []clause.OrderByColumn{
			{Column: clause.Column{Name: "full_name"}},
			{Column: clause.Column{Name: "id"}},
		}}).
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// AssignRoles assigns roles to a user by role names
func (r *UserRepository) AssignRoles(userID uuid.UUID, roleNames []string) error {
	var user models.User
	if err := r.GetDB().First(&user, userID).Error; err != nil {
		return err
	}

	var roles []models.Role
	if err := r.GetDB().Where("name IN ?", roleNames).Find(&roles).Error; err != nil {
		return err
	}

	return r.GetDB().Model(&user).Association("Roles").Replace(roles)
}
