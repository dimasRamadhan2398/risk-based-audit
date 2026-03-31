package repositories

import (
	"auth-service/models"
	apperrors "auth-service/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IPermissionRepository is an alias for the permission repository interface
type IPermissionRepository = PermissionRepositoryInterface

// PermissionRepositoryInterface defines the permission repository interface
type PermissionRepositoryInterface interface {
	Create(permission *models.Permission) error
	Update(permission *models.Permission) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.Permission, error)
	FindByName(name string) (*models.Permission, error)
	FindByResourceAndAction(resource, action string) (*models.Permission, error)
	FindAll() ([]*models.Permission, error)
	FindMany(offset, limit int, search string) ([]*models.Permission, error)
	Count(search string) (int64, error)
}

// PermissionRepository handles permission data operations
type PermissionRepository struct {
	*BaseRepository
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(db *gorm.DB) IPermissionRepository {
	return &PermissionRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create creates a new permission
func (r *PermissionRepository) Create(permission *models.Permission) error {
	return r.BaseRepository.Create(permission)
}

// Update updates a permission
func (r *PermissionRepository) Update(permission *models.Permission) error {
	return r.BaseRepository.Update(permission)
}

// Delete deletes a permission
func (r *PermissionRepository) Delete(id uuid.UUID) error {
	return r.BaseRepository.Delete(&models.Permission{ID: id})
}

// FindByID finds a permission by ID
func (r *PermissionRepository) FindByID(id uuid.UUID) (*models.Permission, error) {
	var permission models.Permission
	if err := r.GetDB().First(&permission, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &permission, nil
}

// FindByName finds a permission by name
func (r *PermissionRepository) FindByName(name string) (*models.Permission, error) {
	var permission models.Permission
	if err := r.GetDB().Where("name = ?", name).First(&permission).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &permission, nil
}

// FindByResourceAndAction finds a permission by resource and action
func (r *PermissionRepository) FindByResourceAndAction(resource, action string) (*models.Permission, error) {
	var permission models.Permission
	if err := r.GetDB().Where("resource = ? AND action = ?", resource, action).First(&permission).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &permission, nil
}

// FindAll finds all permissions
func (r *PermissionRepository) FindAll() ([]*models.Permission, error) {
	var permissions []*models.Permission
	if err := r.GetDB().Find(&permissions).Error; err != nil {
		return nil, err
	}
	return permissions, nil
}

// FindMany finds multiple permissions with pagination
func (r *PermissionRepository) FindMany(offset, limit int, search string) ([]*models.Permission, error) {
	var permissions []*models.Permission
	query := r.GetDB().Model(&models.Permission{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name LIKE ? OR resource LIKE ? OR action LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if err := query.Offset(offset).Limit(limit).Find(&permissions).Error; err != nil {
		return nil, err
	}

	return permissions, nil
}

// Count counts permissions with filters
func (r *PermissionRepository) Count(search string) (int64, error) {
	var count int64
	query := r.GetDB().Model(&models.Permission{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name LIKE ? OR resource LIKE ? OR action LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}
