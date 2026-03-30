package repositories

import (
	"auth-service/models"
	apperrors "auth-service/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IRoleRepository is an alias for the role repository interface
type IRoleRepository = RoleRepositoryInterface

// RoleRepositoryInterface defines the role repository interface
type RoleRepositoryInterface interface {
	Create(role *models.Role) error
	Update(role *models.Role) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.Role, error)
	FindByName(name string) (*models.Role, error)
	FindAll() ([]*models.Role, error)
	FindMany(offset, limit int, search string) ([]*models.Role, error)
	Count(search string) (int64, error)
	AssignPermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error
	RemovePermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error
}

// RoleRepository handles role data operations
type RoleRepository struct {
	*BaseRepository
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *gorm.DB) IRoleRepository {
	return &RoleRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new role
func (r *RoleRepository) Create(role *models.Role) error {
	return r.BaseRepository.Create(role)
}

// Update updates a role
func (r *RoleRepository) Update(role *models.Role) error {
	return r.BaseRepository.Update(role)
}

// Delete deletes a role
func (r *RoleRepository) Delete(id uuid.UUID) error {
	return r.BaseRepository.Delete(&models.Role{ID: id})
}

// FindByID finds a role by ID
func (r *RoleRepository) FindByID(id uuid.UUID) (*models.Role, error) {
	var role models.Role
	if err := r.GetDB().Preload("Permissions").First(&role, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

// FindByName finds a role by name
func (r *RoleRepository) FindByName(name string) (*models.Role, error) {
	var role models.Role
	if err := r.GetDB().Preload("Permissions").Where("name = ?", name).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

// FindAll finds all roles
func (r *RoleRepository) FindAll() ([]*models.Role, error) {
	var roles []*models.Role
	if err := r.GetDB().Preload("Permissions").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// FindMany finds multiple roles with pagination
func (r *RoleRepository) FindMany(offset, limit int, search string) ([]*models.Role, error) {
	var roles []*models.Role
	query := r.GetDB().Model(&models.Role{}).Preload("Permissions")

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", searchPattern, searchPattern)
	}

	if err := query.Offset(offset).Limit(limit).Find(&roles).Error; err != nil {
		return nil, err
	}

	return roles, nil
}

// Count counts roles with filters
func (r *RoleRepository) Count(search string) (int64, error) {
	var count int64
	query := r.GetDB().Model(&models.Role{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", searchPattern, searchPattern)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

// AssignPermissions assigns permissions to a role
func (r *RoleRepository) AssignPermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	var permissions []models.Permission
	if err := r.GetDB().Find(&permissions, permissionIDs).Error; err != nil {
		return err
	}

	var role models.Role
	if err := r.GetDB().First(&role, roleID).Error; err != nil {
		return err
	}

	return r.GetDB().Model(&role).Association("Permissions").Append(permissions)
}

// RemovePermissions removes permissions from a role
func (r *RoleRepository) RemovePermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	var permissions []models.Permission
	if err := r.GetDB().Find(&permissions, permissionIDs).Error; err != nil {
		return err
	}

	var role models.Role
	if err := r.GetDB().First(&role, roleID).Error; err != nil {
		return err
	}

	return r.GetDB().Model(&role).Association("Permissions").Delete(permissions)
}
