package location

import (
	"master-service/models"
	apperrors "master-service/pkg/errors"
	"master-service/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ILocationRepository = LocationRepositoryInterface

// LocationRepositoryInterface defines the location repository interface
type LocationRepositoryInterface interface {
	Create(location *models.Location) error
	Update(location *models.Location) error
	Delete(id uuid.UUID) error
	FindByID(id uuid.UUID) (*models.Location, error)
	FindByName(name string) (*models.Location, error)
	FindByResourceAndAction(resource, action string) (*models.Location, error)
	FindAll() ([]*models.Location, error)
	FindMany(offset, limit int, search string) ([]*models.Location, error)
	Count(search string) (int64, error)
}

// LocationRepository handles location data operations
type LocationRepository struct {
	*repositories.BaseRepository
}

// NewLocationRepository creates a new location repository
func NewLocationRepository(db *gorm.DB) ILocationRepository {
	return &LocationRepository{
		BaseRepository: repositories.NewBaseRepository(db),
	}
}

// Create creates a new location
func (r *LocationRepository) Create(location *models.Location) error {
	return r.BaseRepository.Create(location)
}

// Update updates a location
func (r *LocationRepository) Update(location *models.Location) error {
	return r.BaseRepository.Update(location)
}

// Delete deletes a location
func (r *LocationRepository) Delete(id uuid.UUID) error {
	return r.BaseRepository.Delete(&models.Location{ID: id})
}

// FindByID finds a location by ID
func (r *LocationRepository) FindByID(id uuid.UUID) (*models.Location, error) {
	var location models.Location
	if err := r.GetDB().First(&location, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &location, nil
}

// FindByName finds a location by name
func (r *LocationRepository) FindByName(name string) (*models.Location, error) {
	var location models.Location
	if err := r.GetDB().Where("name = ?", name).First(&location).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &location, nil
}

// FindByResourceAndAction finds a location by resource and action
func (r *LocationRepository) FindByResourceAndAction(resource, action string) (*models.Location, error) {
	var location models.Location
	if err := r.GetDB().Where("resource = ? AND action = ?", resource, action).First(&location).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &location, nil
}

// FindAll finds all locations
func (r *LocationRepository) FindAll() ([]*models.Location, error) {
	var locations []*models.Location
	if err := r.GetDB().Find(&locations).Error; err != nil {
		return nil, err
	}
	return locations, nil
}

// FindMany finds multiple locations with pagination
func (r *LocationRepository) FindMany(offset, limit int, search string) ([]*models.Location, error) {
	var locations []*models.Location
	query := r.GetDB().Model(&models.Location{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name LIKE ? OR resource LIKE ? OR action LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if err := query.Offset(offset).Limit(limit).Find(&locations).Error; err != nil {
		return nil, err
	}

	return locations, nil
}

// Count counts locations with filters
func (r *LocationRepository) Count(search string) (int64, error) {
	var count int64
	query := r.GetDB().Model(&models.Location{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name LIKE ? OR resource LIKE ? OR action LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}
