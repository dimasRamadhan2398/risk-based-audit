package repositories

import (
	"auth-service/models"
	apperrors "auth-service/pkg/errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
	FindMany(offset, limit int, search, department string, isActive *bool) ([]*models.User, error)
	Count(search, department string, isActive *bool) (int64, error)
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
